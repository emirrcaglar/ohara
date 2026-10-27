[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [string]$Server,
    [switch]$BuildOnly
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$configFile = Join-Path $PSScriptRoot 'deploy.conf'
if (!(Test-Path -LiteralPath $configFile)) {
    $configFile = Join-Path $PSScriptRoot '.deploy.conf'
}

$settings = @{
    DEPLOY_SERVER = ''
    DEPLOY_DIR = '/opt/ohara'
    DEPLOY_DATA_DIR = '/var/lib/ohara'
    DEPLOY_CONFIG_DIR = '/etc/ohara'
    DEPLOY_CACHE_DIR = '/var/cache/ohara'
    DEPLOY_BINARY_NAME = 'ohara'
    DEPLOY_SERVICE_NAME = 'ohara'
}

# Read literal assignments without executing the Bash config or loading passwords.
if (Test-Path -LiteralPath $configFile) {
    foreach ($line in Get-Content -LiteralPath $configFile) {
        if ($line -notmatch '^\s*(DEPLOY_[A-Z_]+)=(.*?)\s*$') { continue }
        $key = $Matches[1]
        $value = $Matches[2]
        if (!$settings.ContainsKey($key)) { continue }
        if ($value -match '^"([^"]*)"\s*(?:#.*)?$' -or $value -match "^'([^']*)'\s*(?:#.*)?$") {
            $value = $Matches[1]
        } elseif ($value -match '^([^\s#]*)\s*(?:#.*)?$') {
            $value = $Matches[1]
        } else {
            throw "Use a literal value for $key in $configFile."
        }
        if ($value) { $settings[$key] = $value }
    }
}
if (!$Server) { $Server = $settings.DEPLOY_SERVER }
if (!$BuildOnly -and $Server -notmatch '^[A-Za-z0-9_.-]+@(?:[A-Za-z0-9_.-]+|\[[0-9A-Fa-f:]+\])$') {
    throw 'Specify a target as user@host: .\deploy\deploy.ps1 user@host, or set DEPLOY_SERVER in deploy.conf.'
}
foreach ($key in @('DEPLOY_DIR', 'DEPLOY_DATA_DIR', 'DEPLOY_CONFIG_DIR', 'DEPLOY_CACHE_DIR')) {
    if ($settings[$key] -notmatch '^/[A-Za-z0-9_./-]+$' -or $settings[$key] -match '%') {
        throw "$key must be an absolute Linux path without spaces or shell expressions."
    }
}
foreach ($key in @('DEPLOY_BINARY_NAME', 'DEPLOY_SERVICE_NAME')) {
    if ($settings[$key] -notmatch '^[A-Za-z0-9_][A-Za-z0-9_.-]*$') {
        throw "$key must contain only letters, digits, underscores, dots, or hyphens."
    }
}

function Find-Tool {
    param([string[]]$Names)
    foreach ($name in $Names) {
        $command = Get-Command $name -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($command) { return $command.Source }
    }
    throw "Install '$($Names[-1])' on Windows and reopen PowerShell before deploying."
}

function Invoke-Checked {
    param([string]$Command, [string[]]$ArgumentList)
    & $Command @ArgumentList
    if ($LASTEXITCODE -ne 0) {
        throw "$(Split-Path -Leaf $Command) failed with exit code $LASTEXITCODE."
    }
}

function Write-UnixFile {
    param([string]$Path, [string]$Content)
    [IO.File]::WriteAllText($Path, ($Content.Replace("`r`n", "`n") + "`n"), [Text.UTF8Encoding]::new($false))
}

$npm = Find-Tool @('npm.cmd', 'npm')
$go = Find-Tool @('go')
if (!$BuildOnly) {
    $ssh = Find-Tool @('ssh')
    $scp = Find-Tool @('scp')
}
$timer = [Diagnostics.Stopwatch]::StartNew()
$tempDir = $null
$deployPassword = $null
$previousSshEnv = @{}
Push-Location $repoRoot
try {
    Write-Host 'Building frontend...'
    if (!(Test-Path -LiteralPath 'src/frontend/node_modules/.bin/vite')) {
        Invoke-Checked $npm @('--prefix', 'src/frontend', 'install')
    }
    Invoke-Checked $npm @('--prefix', 'src/frontend', 'run', 'build:embed')

    Write-Host 'Building Linux amd64 backend...'
    New-Item -ItemType Directory -Path 'dist' -Force | Out-Null
    $binaryPath = Join-Path $repoRoot "dist/$($settings.DEPLOY_BINARY_NAME)"
    $previousGoEnv = @{}
    foreach ($name in @('GOOS', 'GOARCH', 'CGO_ENABLED')) {
        $previousGoEnv[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
    }
    try {
        $env:GOOS = 'linux'
        $env:GOARCH = 'amd64'
        $env:CGO_ENABLED = '0'
        Invoke-Checked $go @('-C', 'src/backend', 'build', '-o', $binaryPath, './cmd')
    } finally {
        foreach ($name in $previousGoEnv.Keys) {
            if ($null -eq $previousGoEnv[$name]) {
                Remove-Item -LiteralPath "Env:$name" -ErrorAction SilentlyContinue
            } else {
                [Environment]::SetEnvironmentVariable($name, $previousGoEnv[$name], 'Process')
            }
        }
    }
    if ($BuildOnly) {
        Write-Host "Build complete: $binaryPath"
        return
    }

    $tempDir = Join-Path ([IO.Path]::GetTempPath()) "ohara-deploy-$([Guid]::NewGuid().ToString('N'))"
    New-Item -ItemType Directory -Path $tempDir | Out-Null
    $securePassword = Read-Host 'SSH password (also used for sudo; leave empty for SSH keys)' -AsSecureString
    $passwordPointer = [IntPtr]::Zero
    try {
        $passwordPointer = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($securePassword)
        $deployPassword = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($passwordPointer)
    } finally {
        if ($passwordPointer -ne [IntPtr]::Zero) {
            [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($passwordPointer)
        }
        $securePassword.Dispose()
    }
    if ($deployPassword) {
        # The helper contains no credentials; SSH and SCP read the same in-memory password.
        $askpassFile = Join-Path $tempDir 'askpass.cmd'
        $askpassShell = Join-Path $PSHOME 'powershell.exe'
        if (!(Test-Path -LiteralPath $askpassShell)) { $askpassShell = Join-Path $PSHOME 'pwsh.exe' }
        [IO.File]::WriteAllText($askpassFile, "@echo off`r`n`"$askpassShell`" -NoLogo -NoProfile -NonInteractive -Command `"[Console]::OutputEncoding = [Text.UTF8Encoding]::new(`$false); [Console]::WriteLine(`$env:OHARA_DEPLOY_PASSWORD)`"`r`n")
        foreach ($name in @('SSH_ASKPASS', 'SSH_ASKPASS_REQUIRE', 'OHARA_DEPLOY_PASSWORD')) {
            $previousSshEnv[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
        }
        $env:SSH_ASKPASS = $askpassFile
        $env:SSH_ASKPASS_REQUIRE = 'force'
        $env:OHARA_DEPLOY_PASSWORD = $deployPassword
    }

    $sshOptions = @('-o', 'ConnectTimeout=15', '-o', 'ServerAliveInterval=15', '-o', 'ServerAliveCountMax=3', '-o', 'StrictHostKeyChecking=accept-new')
    if ($deployPassword) { $sshOptions += @('-o', 'NumberOfPasswordPrompts=1') }
    Write-Host 'Preparing VPS staging directory...'
    $remoteInfo = @(& $ssh @sshOptions $Server 'set -e; id -un; id -gn; mktemp -d /tmp/ohara-deploy.XXXXXXXX')
    if ($LASTEXITCODE -ne 0) { throw 'Failed to prepare the VPS staging directory.' }
    if ($remoteInfo.Count -ne 3 -or $remoteInfo[0] -notmatch '^[A-Za-z_][A-Za-z0-9_.-]*\$?$' -or $remoteInfo[1] -notmatch '^[A-Za-z_][A-Za-z0-9_.-]*\$?$' -or $remoteInfo[2] -notmatch '^/tmp/ohara-deploy\.[A-Za-z0-9]+$') {
        throw 'Unexpected SSH response while identifying the service user and staging directory.'
    }
    $serviceUser = $remoteInfo[0]
    $serviceGroup = $remoteInfo[1]
    $remoteTemp = $remoteInfo[2]
    $serviceName = $settings.DEPLOY_SERVICE_NAME
    $remoteDir = $settings.DEPLOY_DIR
    $dataDir = $settings.DEPLOY_DATA_DIR
    $configDir = $settings.DEPLOY_CONFIG_DIR
    $cacheDir = $settings.DEPLOY_CACHE_DIR
    $binaryName = $settings.DEPLOY_BINARY_NAME
    $deployedAt = [DateTime]::UtcNow.ToString("yyyy-MM-ddTHH:mm:ssZ")

    $serviceFile = Join-Path $tempDir 'ohara.service'
    $installFile = Join-Path $tempDir 'install.sh'
    $stagedBinary = Join-Path $tempDir 'ohara.bin'
    Copy-Item -LiteralPath $binaryPath -Destination $stagedBinary
    Write-UnixFile $serviceFile @"
[Unit]
Description=Ohara Backend Service
After=network.target

[Service]
Type=simple
User=$serviceUser
WorkingDirectory=$dataDir
ExecStart=$remoteDir/$binaryName
Environment="OHARA_DEPLOYED_AT=$deployedAt"
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
"@
    Write-UnixFile $installFile @"
#!/usr/bin/env bash
set -euo pipefail
deploy_tmp='$remoteTemp'
cleanup() {
    rm -f -- "`$deploy_tmp/ohara.bin" "`$deploy_tmp/ohara.service" "`$deploy_tmp/install.sh"
    rmdir -- "`$deploy_tmp"
}
trap cleanup EXIT
mkdir -p '$remoteDir' '$dataDir' '$configDir' '$cacheDir'
install -m 644 -- "`$deploy_tmp/ohara.service" '/etc/systemd/system/$serviceName.service'
systemctl daemon-reload
systemctl enable '$serviceName'
systemctl stop '$serviceName' || true
install -m 755 -- "`$deploy_tmp/ohara.bin" '$remoteDir/$binaryName'
chown -R '${serviceUser}:${serviceGroup}' '$remoteDir' '$dataDir' '$configDir' '$cacheDir'
systemctl reset-failed '$serviceName' || true
systemctl start '$serviceName'
sleep 2
if ! systemctl is-active --quiet '$serviceName'; then
    systemctl status --no-pager '$serviceName' || true
    journalctl -u '$serviceName' --no-pager -n 50 || true
    exit 1
fi
"@

    Write-Host 'Uploading binary and service files...'
    Invoke-Checked $scp ($sshOptions + @($stagedBinary, $serviceFile, $installFile, "${Server}:${remoteTemp}/"))
    Write-Host 'Installing and restarting the service...'
    if ($deployPassword) {
        # Base64 keeps native PowerShell pipe encoding from altering the password.
        # Remove transport line endings before decoding, and never put it in a remote command.
        $encodedPassword = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($deployPassword + "`n"))
        $encodedPassword | & $ssh @sshOptions -T $Server "tr -d '\r\n' | base64 -d | sudo -S -p '' bash '$remoteTemp/install.sh'"
        if ($LASTEXITCODE -ne 0) { throw "SSH installation failed with exit code $LASTEXITCODE." }
    } else {
        Invoke-Checked $ssh ($sshOptions + @('-t', $Server, "sudo bash '$remoteTemp/install.sh'"))
    }
    Write-Host ('Deployment complete in {0:m\:ss}. Service: {1}' -f $timer.Elapsed, $serviceName)
} finally {
    foreach ($name in $previousSshEnv.Keys) {
        if ($null -eq $previousSshEnv[$name]) {
            Remove-Item -LiteralPath "Env:$name" -ErrorAction SilentlyContinue
        } else {
            [Environment]::SetEnvironmentVariable($name, $previousSshEnv[$name], 'Process')
        }
    }
    $deployPassword = $null
    $encodedPassword = $null
    if ($tempDir -and (Test-Path -LiteralPath $tempDir)) {
        foreach ($fileName in @('ohara.bin', 'ohara.service', 'install.sh', 'askpass.cmd')) {
            Remove-Item -LiteralPath (Join-Path $tempDir $fileName) -Force -ErrorAction SilentlyContinue
        }
        Remove-Item -LiteralPath $tempDir -ErrorAction SilentlyContinue
    }
    Pop-Location
}
