# Ohara

Media server for manga and audio content.

## Deploy

### Prerequisites

**Ubuntu / Debian**
```bash
sudo apt install sshpass nodejs npm golang-go
```

**Arch / Manjaro**
```bash
sudo pacman -S sshpass nodejs npm go
```

**Fedora / RHEL**
```bash
sudo dnf install sshpass nodejs npm golang
```

**macOS**
```bash
brew install sshpass node go
```

### 1. Create your config

```bash
cp deploy/deploy.conf.example deploy/deploy.conf
```

Edit `deploy/deploy.conf`:

```bash
DEPLOY_SERVER="you@your-server"
DEPLOY_PASSWORD="yourpassword"
```

### 2. Run

```bash
./deploy/deploy.sh
```

The script builds the frontend, compiles the Go binary for Linux, uploads both to the server, and registers/restarts the systemd service — no further prompts.

`deploy.conf` is gitignored. See `deploy/deploy.conf.example` for all available config keys.

---

For development setup, architecture, API reference, and project structure see [docs/](docs/).
