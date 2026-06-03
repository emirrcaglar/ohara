#!/usr/bin/env bash
set -e

# Ohara installer - downloads pre-built binary from GitHub Releases

REPO="emirrcaglar/ohara"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# normalize architecture names for goreleaser
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)
        echo "Error: Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case $OS in
    linux|darwin) ;;
    *)
        echo "Error: Unsupported OS: $OS"  # no windows support yet
        echo "Supported: Linux, macOS"
        exit 1
        ;;
esac

echo "Fetching latest release..."
LATEST=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)

if [ -z "$LATEST" ]; then
    echo "Error: Could not fetch latest release"
    exit 1
fi

echo "Installing Ohara $LATEST for $OS/$ARCH..."

# Download
FILENAME="ohara_${LATEST#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST/$FILENAME"

echo "Downloading from $URL..."
curl -sL "$URL" -o "/tmp/$FILENAME"

echo "Extracting..."
tar -xzf "/tmp/$FILENAME" -C /tmp
rm "/tmp/$FILENAME"

# Install
echo "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    mv /tmp/ohara "$INSTALL_DIR/ohara"
else
    sudo mv /tmp/ohara "$INSTALL_DIR/ohara"
    sudo chmod +x "$INSTALL_DIR/ohara"
fi

# Setup systemd service on Linux
if [ -d /etc/systemd/system ] && [ "$OS" = "linux" ]; then
    echo ""
    echo "Setting up systemd service..."

    # Stop and disable existing service if it exists
    if systemctl is-active --quiet ohara 2>/dev/null; then
        sudo systemctl stop ohara
        echo "Stopped existing ohara service"
    fi
    if systemctl is-enabled --quiet ohara 2>/dev/null; then
        sudo systemctl disable ohara
        echo "Disabled existing ohara service"
    fi

    # Create dedicated system user if it doesn't exist
    if ! id -u ohara >/dev/null 2>&1; then
        sudo useradd --system --no-create-home --shell /bin/false ohara
        echo "Created system user 'ohara'"
    fi

    # FHS-compliant paths
    DATA_DIR="/var/lib/ohara"          # Application state data
    CONFIG_DIR="/etc/ohara"             # Configuration files
    CACHE_DIR="/var/cache/ohara"        # Cache/temporary data

    # Create directories with proper ownership
    sudo mkdir -p "$DATA_DIR" "$CONFIG_DIR" "$CACHE_DIR"
    sudo chown -R ohara:ohara "$DATA_DIR" "$CONFIG_DIR" "$CACHE_DIR"
    sudo chmod 750 "$DATA_DIR" "$CONFIG_DIR"
    sudo chmod 755 "$CACHE_DIR"

    # Create environment file for configuration
    sudo tee "$CONFIG_DIR/environment" > /dev/null <<EOF
# Ohara configuration
# Uncomment and set these for production deployments:
# OHARA_ADMIN_USER=admin
# OHARA_ADMIN_PASS=changeme
# OHARA_DEPLOYED_AT=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
EOF
    sudo chown ohara:ohara "$CONFIG_DIR/environment"
    sudo chmod 640 "$CONFIG_DIR/environment"

    # Create systemd service with security hardening
    sudo tee /etc/systemd/system/ohara.service > /dev/null <<EOF
[Unit]
Description=Ohara Media Server
After=network.target
Documentation=https://github.com/emirrcaglar/ohara

[Service]
Type=simple
User=ohara
Group=ohara
WorkingDirectory=$DATA_DIR

# Paths
ExecStart=$INSTALL_DIR/ohara -data $DATA_DIR
EnvironmentFile=-$CONFIG_DIR/environment

# Restart policy
Restart=on-failure
RestartSec=5s

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$DATA_DIR $CACHE_DIR
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable ohara
    sudo systemctl start ohara

    echo ""
    echo "✓ Systemd service installed with FHS-compliant paths:"
    echo "  Binary:  $INSTALL_DIR/ohara"
    echo "  Data:    $DATA_DIR"
    echo "  Config:  $CONFIG_DIR"
    echo "  Cache:   $CACHE_DIR"
    echo ""

    # Check if service started successfully
    if systemctl is-active --quiet ohara; then
        echo "✓ Service is running"
    else
        echo "⚠ Service failed to start. Check logs:"
        echo "  sudo journalctl -u ohara -n 20"
    fi

    echo ""
    echo "Commands:"
    echo "  Status:  sudo systemctl status ohara"
    echo "  Logs:    sudo journalctl -u ohara -f"
    echo "  Restart: sudo systemctl restart ohara"
    echo "  Config:  sudo nano $CONFIG_DIR/environment"
fi

# Verify
if command -v ohara >/dev/null 2>&1; then
    echo ""
    echo "✓ Ohara installed successfully!"
else
    echo "✓ Binary installed to $INSTALL_DIR/ohara"
    echo ""
    echo "If 'ohara' command not found, add to PATH:"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
fi
