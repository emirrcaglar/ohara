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
ASKPASS_SCRIPT=""
INSTALL_SCRIPT_TMP=""

SERVER="${SERVER//$'\r'/}"
REMOTE_DIR="${REMOTE_DIR//$'\r'/}"
BINARY_NAME="${BINARY_NAME//$'\r'/}"
SERVICE_NAME="${SERVICE_NAME//$'\r'/}"
DEPLOY_PASSWORD="${DEPLOY_PASSWORD//$'\r'/}"

cleanup() {
	if [[ -n "$ASKPASS_SCRIPT" && -f "$ASKPASS_SCRIPT" ]]; then
		rm -f "$ASKPASS_SCRIPT"
	fi
	if [[ -n "$INSTALL_SCRIPT_TMP" && -f "$INSTALL_SCRIPT_TMP" ]]; then
		rm -f "$INSTALL_SCRIPT_TMP"
	fi
}
trap cleanup EXIT

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

SSH_CMD=(ssh)
SCP_CMD=(scp)

if [[ -n "$DEPLOY_PASSWORD" ]]; then
	PASSWORD_ESCAPED="${DEPLOY_PASSWORD//\'/\'\"\'\"\'}"
	ASKPASS_SCRIPT="$(mktemp "${TMPDIR:-/tmp}/ohara-askpass.XXXXXX")"
	cat <<EOF > "$ASKPASS_SCRIPT"
#!/usr/bin/env bash
printf '%s\n' '$PASSWORD_ESCAPED'
EOF
	chmod 700 "$ASKPASS_SCRIPT"
	SSH_ASKPASS_ENV=(env SSH_ASKPASS="$ASKPASS_SCRIPT" SSH_ASKPASS_REQUIRE=force DISPLAY=none)
	SSH_OPTS=(-o PreferredAuthentications=password -o PubkeyAuthentication=no -o KbdInteractiveAuthentication=no)
	if command -v setsid >/dev/null 2>&1; then
		SSH_CMD=("${SSH_ASKPASS_ENV[@]}" setsid -w ssh "${SSH_OPTS[@]}")
		SCP_CMD=("${SSH_ASKPASS_ENV[@]}" setsid -w scp "${SSH_OPTS[@]}")
	else
		# SSH_ASKPASS_REQUIRE=force (OpenSSH >= 8.4) is sufficient without setsid.
		SSH_CMD=("${SSH_ASKPASS_ENV[@]}" ssh "${SSH_OPTS[@]}")
		SCP_CMD=("${SSH_ASKPASS_ENV[@]}" scp "${SSH_OPTS[@]}")
	fi
fi

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
"${SCP_CMD[@]}" "./$BINARY_NAME" "$SERVER:/tmp/$BINARY_NAME.tmp"

echo "Step 2: Generating and uploading service file..."
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

echo "Step 3: Running remote installation..."
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

if [[ -n "$DEPLOY_PASSWORD" ]]; then
	ESCAPED_DEPLOY_PASSWORD="${DEPLOY_PASSWORD//\'/\'\"\'\"\'}"
	"${SSH_CMD[@]}" -t "$SERVER" "printf '%s\n' '$ESCAPED_DEPLOY_PASSWORD' | sudo -S -p '' bash '$REMOTE_INSTALL_TMP'"
else
	"${SSH_CMD[@]}" -t "$SERVER" "sudo bash '$REMOTE_INSTALL_TMP'"
fi

echo "Deployment complete! Your service is now fully managed by systemd."
