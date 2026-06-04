# Ohara

Minimal personal media server for manga, audio, and video.

## Installation

### Quick install

```bash
curl -sSL https://raw.githubusercontent.com/emirrcaglar/ohara/main/install.sh | bash
```

The installer downloads the latest pre-built binary for Linux or macOS and installs it to `/usr/local/bin` by default.

On Linux systems with systemd, it also creates and starts an `ohara` service using these paths:

- Binary: `/usr/local/bin/ohara`
- Data: `/var/lib/ohara`
- Config: `/etc/ohara`
- Cache: `/var/cache/ohara`

### Manual download

Download the appropriate archive from [GitHub Releases](https://github.com/emirrcaglar/ohara/releases). Release assets are named with the version, OS, and architecture, for example:

```text
ohara_<version>_linux_amd64.tar.gz
ohara_<version>_darwin_arm64.tar.gz
ohara_<version>_windows_amd64.zip
```

Then extract the archive and place the `ohara` binary somewhere on your `PATH`.

### Docker

```bash
docker compose up -d
```

This builds the local image and serves Ohara on `http://localhost:3000`, storing data in the `ohara-data` Docker volume.

## Running

```bash
ohara
```

Ohara stores data in its working directory and listens on port `3000`:

```text
http://localhost:3000
```

The installed systemd service and Docker image both use `/var/lib/ohara` as the working directory.

## First login

On first database setup, Ohara creates a default admin account:

- Username: `admin`
- Password: `admin`

Change this password immediately after logging in.

## Development

### Requirements

- Go `1.24.4` or newer compatible `1.24.x`
- Node.js `20.19+` or `22.12+`
- npm

### Build from source

```bash
npm --prefix src/frontend install
npm --prefix src/frontend run build:embed
go -C src/backend build -o ../../dist/ohara ./cmd
```

Run the built binary:

```bash
./dist/ohara
```

### Deploy from source

Create a local deploy config:

```bash
cp deploy/deploy.conf.example deploy/deploy.conf
```

Edit `deploy/deploy.conf`:

```bash
DEPLOY_SERVER="user@your-vps"
DEPLOY_PASSWORD="yourpassword"
```

Then deploy:

```bash
./deploy/deploy.sh
```

The deploy script builds the frontend and Linux `amd64` backend locally, uploads the binary to the target server, installs it under `/opt/ohara`, and registers/restarts the systemd service.

`deploy/deploy.conf` is gitignored. See `deploy/deploy.conf.example` for optional path and service-name overrides.

## Documentation

Additional notes live in [docs/](docs/).
