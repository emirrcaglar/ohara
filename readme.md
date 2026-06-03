# Ohara

Minimal personal media server for manga, audio, and video.

## Installation

### Quick Install (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/emirrcaglar/ohara/main/install.sh | bash
```

This downloads the pre-built binary for your OS and installs it to `/usr/local/bin`.

### Manual Download

Download the appropriate binary for your system from [GitHub Releases](https://github.com/emirrcaglar/ohara/releases):

```bash
# Linux x64
curl -LO https://github.com/emirrcaglar/ohara/releases/latest/download/ohara_linux_amd64.tar.gz
tar xzf ohara_linux_amd64.tar.gz
sudo mv ohara /usr/local/bin/

# macOS Apple Silicon
curl -LO https://github.com/emirrcaglar/ohara/releases/latest/download/ohara_darwin_arm64.tar.gz
tar xzf ohara_darwin_arm64.tar.gz
sudo mv ohara /usr/local/bin/
```

### Run

```bash
ohara
```

By default, Ohara runs on port 8080. Visit `http://localhost:8080` in your browser.

## Configuration

Ohara can be configured via environment variables:

```bash
export OHARA_ADMIN_USER=admin
export OHARA_ADMIN_PASS=yourpassword
ohara
```

## Dependency model

Ohara aims for self-contained deployment.

- Local/source builds use the tools available on your machine.
- Production deploys should not pollute the VPS with global media packages.
- Runtime helpers, when needed, should be managed inside Ohara's own install directory.

## Development

### Deploy from source (for developers)

#### 1. Create your config

```bash
cp deploy/deploy.conf.example deploy/deploy.conf
```

Edit `deploy/deploy.conf`:

```bash
DEPLOY_SERVER="you@your-server"
DEPLOY_PASSWORD="yourpassword"
```

#### 2. Run

```bash
./deploy/deploy.sh
```

The script builds locally, uploads the release to the server, and registers/restarts the systemd service — no further prompts.

The VPS receives Ohara, not your local build toolchain.

`deploy.conf` is gitignored. See `deploy/deploy.conf.example` for all available config keys.

---

For development setup, architecture, API reference, and project structure see [docs/](docs/).
