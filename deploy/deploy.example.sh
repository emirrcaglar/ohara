#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONFIG_FILE="$SCRIPT_DIR/deploy.conf"
if [[ ! -f "$CONFIG_FILE" && -f "$SCRIPT_DIR/.deploy.conf" ]]; then
	CONFIG_FILE="$SCRIPT_DIR/.deploy.conf"
fi

# Load deploy defaults from local config if present.
# shellcheck source=/dev/null
if [[ -f "$CONFIG_FILE" ]]; then
	source "$CONFIG_FILE"
fi

SERVER="${1:-${DEPLOY_SERVER:-}}"
REMOTE_DIR="${DEPLOY_DIR:-/opt/ohara}"
BINARY_NAME="${DEPLOY_BINARY_NAME:-ohara}"
SERVICE_NAME="${DEPLOY_SERVICE_NAME:-ohara}"

SERVER="${SERVER//$'\r'/}"
REMOTE_DIR="${REMOTE_DIR//$'\r'/}"
BINARY_NAME="${BINARY_NAME//$'\r'/}"
SERVICE_NAME="${SERVICE_NAME//$'\r'/}"

if [[ -z "$SERVER" ]]; then
	echo "Error: No deploy target specified."
	echo "Usage: $0 user@host"
	echo "Or set DEPLOY_SERVER in $CONFIG_FILE"
	exit 1
fi

if [[ "$SERVER" == *"://"* ]]; then
	echo "Error: Invalid deploy target '$SERVER'."
	echo "Use user@host (without http:// or https://)."
	exit 1
fi

cd "$REPO_ROOT"

for cmd in npm go ssh scp; do
	if ! command -v "$cmd" >/dev/null 2>&1; then
		echo "Error: '$cmd' is required but not installed in this shell."
		exit 1
	fi
done

echo "Building the binary..."
if [[ ! -d "frontend/node_modules" ]]; then
	npm --prefix frontend install
fi
npm --prefix frontend run build:embed

cd backend
GOOS=linux GOARCH=amd64 go build -o "../$BINARY_NAME" ./cmd
cd ..

echo "Step 1: Uploading binary to temporary storage..."
# We upload to /tmp because the user 'emirc' always has permission there.
# This will ask for your SSH password once.
scp "./$BINARY_NAME" "$SERVER:/tmp/$BINARY_NAME.tmp"

echo "Step 2: Running remote installation..."
# We use 'ssh -t' to force a terminal so you can see the sudo password prompt.
# All commands are chained so you only enter the sudo password once.
ssh -t "$SERVER" "
    echo 'Finalizing deployment as root...' && \
    sudo mkdir -p '$REMOTE_DIR' && \
    sudo systemctl stop '$SERVICE_NAME' || true && \
    sudo mv '/tmp/$BINARY_NAME.tmp' '$REMOTE_DIR/$BINARY_NAME' && \
    sudo chmod +x '$REMOTE_DIR/$BINARY_NAME' && \
    sudo systemctl start '$SERVICE_NAME'
"

echo "Deployment complete!"