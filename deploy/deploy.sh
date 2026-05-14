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
DEPLOY_PASSWORD="${DEPLOY_PASSWORD:-}"
INSTALL_SCRIPT_TMP=""
DEPLOY_START_SECONDS=$SECONDS

SERVER="${SERVER//$'\r'/}"
REMOTE_DIR="${REMOTE_DIR//$'\r'/}"
BINARY_NAME="${BINARY_NAME//$'\r'/}"
SERVICE_NAME="${SERVICE_NAME//$'\r'/}"
DEPLOY_PASSWORD="${DEPLOY_PASSWORD//$'\r'/}"

cleanup() {
	if [[ -n "$INSTALL_SCRIPT_TMP" && -f "$INSTALL_SCRIPT_TMP" ]]; then
		rm -f "$INSTALL_SCRIPT_TMP"
	fi
}
trap cleanup EXIT

format_duration() {
	local total_seconds=$1
	local minutes=$((total_seconds / 60))
	local seconds=$((total_seconds % 60))

	printf '%dm %02ds' "$minutes" "$seconds"
}

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

for cmd in npm go ssh scp sshpass; do
	if ! command -v "$cmd" >/dev/null 2>&1; then
		echo "Error: '$cmd' is required but not installed."
		[[ "$cmd" == "sshpass" ]] && echo "Install it with: sudo apt install sshpass"
		exit 1
	fi
done

if [[ -z "$DEPLOY_PASSWORD" ]]; then
	echo "Error: DEPLOY_PASSWORD is not set in $CONFIG_FILE"
	exit 1
fi

SSH_OPTS=(-o ConnectTimeout=15 -o ServerAliveInterval=15 -o ServerAliveCountMax=3 -o StrictHostKeyChecking=accept-new)
SSH_CMD=(sshpass -p "$DEPLOY_PASSWORD" ssh "${SSH_OPTS[@]}")
SCP_CMD=(sshpass -p "$DEPLOY_PASSWORD" scp "${SSH_OPTS[@]}")

echo "Building the binary..."
if [[ ! -x "frontend/node_modules/.bin/vite" ]]; then
	echo "Installing frontend dependencies..."
	npm --prefix frontend install
fi
npm --prefix frontend run build:embed

cd backend
GOOS=linux GOARCH=amd64 go build -o "../$BINARY_NAME" ./cmd
cd ..

echo "Step 1/3: Uploading binary to temporary storage..."
# We upload to /tmp because the non-root users always have permission there.
"${SCP_CMD[@]}" "./$BINARY_NAME" "$SERVER:/tmp/$BINARY_NAME.tmp"

echo "Step 2/3: Generating and uploading service file..."
SERVICE_FILE_TMP="/tmp/$SERVICE_NAME.service"
SERVICE_USER="$("${SSH_CMD[@]}" "$SERVER" "whoami")"

# Generate the service file locally using a Here-Doc
cat <<EOF > "$SERVICE_FILE_TMP"
[Unit]
Description=Ohara Backend Service
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
WorkingDirectory=$REMOTE_DIR
ExecStart=$REMOTE_DIR/$BINARY_NAME
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

"${SCP_CMD[@]}" "$SERVICE_FILE_TMP" "$SERVER:/tmp/$SERVICE_NAME.service.tmp"
rm "$SERVICE_FILE_TMP"

echo "Step 3/3: Running remote installation..."
REMOTE_DIR_Q=$(printf '%q' "$REMOTE_DIR")
SERVICE_USER_Q=$(printf '%q' "$SERVICE_USER")
SERVICE_REMOTE_TMP_Q=$(printf '%q' "/tmp/$SERVICE_NAME.service.tmp")
SERVICE_REMOTE_PATH_Q=$(printf '%q' "/etc/systemd/system/$SERVICE_NAME.service")
SERVICE_NAME_Q=$(printf '%q' "$SERVICE_NAME")
BINARY_REMOTE_TMP_Q=$(printf '%q' "/tmp/$BINARY_NAME.tmp")
BINARY_REMOTE_PATH_Q=$(printf '%q' "$REMOTE_DIR/$BINARY_NAME")
REMOTE_INSTALL_TMP="/tmp/$SERVICE_NAME.install.sh"
INSTALL_SCRIPT_TMP="$(mktemp "${TMPDIR:-/tmp}/ohara-install.XXXXXX")"

cat <<EOF > "$INSTALL_SCRIPT_TMP"
#!/usr/bin/env bash
set -euo pipefail
trap 'rm -f "$REMOTE_INSTALL_TMP"' EXIT
echo "Finalizing deployment as root..."
mkdir -p $REMOTE_DIR_Q
mv $SERVICE_REMOTE_TMP_Q $SERVICE_REMOTE_PATH_Q
systemctl daemon-reload
systemctl enable $SERVICE_NAME_Q
systemctl stop $SERVICE_NAME_Q || true
mv $BINARY_REMOTE_TMP_Q $BINARY_REMOTE_PATH_Q
chmod +x $BINARY_REMOTE_PATH_Q
chown -R $SERVICE_USER_Q:$SERVICE_USER_Q $REMOTE_DIR_Q
systemctl reset-failed $SERVICE_NAME_Q || true
systemctl start $SERVICE_NAME_Q
if ! systemctl is-active --quiet $SERVICE_NAME_Q; then
	echo "Error: $SERVICE_NAME_Q did not stay running."
	systemctl status --no-pager $SERVICE_NAME_Q || true
	journalctl -u $SERVICE_NAME_Q --no-pager -n 50 || true
	exit 1
fi
EOF

chmod 700 "$INSTALL_SCRIPT_TMP"
"${SCP_CMD[@]}" "$INSTALL_SCRIPT_TMP" "$SERVER:$REMOTE_INSTALL_TMP"
rm -f "$INSTALL_SCRIPT_TMP"
INSTALL_SCRIPT_TMP=""

"${SSH_CMD[@]}" -t "$SERVER" "echo '$DEPLOY_PASSWORD' | sudo -S -p '' bash '$REMOTE_INSTALL_TMP'"

elapsed_seconds=$((SECONDS - DEPLOY_START_SECONDS))
echo "Deployment complete in $(format_duration "$elapsed_seconds")! Your service is now fully managed by systemd."
