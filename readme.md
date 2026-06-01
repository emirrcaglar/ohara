# Ohara

Minimal personal media server for manga, audio, and video.

## Dependency model

Ohara aims for self-contained deployment.

- Local/source builds use the tools available on your machine.
- Production deploys should not pollute the VPS with global media packages.
- Runtime helpers, when needed, should be managed inside Ohara's own install directory.

## Deploy from source

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

The script builds locally, uploads the release to the server, and registers/restarts the systemd service — no further prompts.

The VPS receives Ohara, not your local build toolchain.

`deploy.conf` is gitignored. See `deploy/deploy.conf.example` for all available config keys.

---

For development setup, architecture, API reference, and project structure see [docs/](docs/).
