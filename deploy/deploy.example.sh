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

SERVER="${SERVER//$'\r'/}"
REMOTE_DIR="${REMOTE_DIR//$'\r'/}"
BINARY_NAME="${BINARY_NAME//$'\r'/}"
SERVICE_NAME="${SERVICE_NAME//$'\r'/}"
DEPLOY_PASSWORD="${DEPLOY_PASSWORD//$'\r'/}"

cleanup() {
	if [[ -n "$ASKPASS_SCRIPT" && -f "$ASKPASS_SCRIPT" ]]; then
		rm -f "$ASKPASS_SCRIPT"
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
	SSH_CMD=(env SSH_ASKPASS="$ASKPASS_SCRIPT" SSH_ASKPASS_REQUIRE=force DISPLAY=none setsid -w ssh -o PreferredAuthentications=password -o PubkeyAuthentication=no -o KbdInteractiveAuthentication=no)
	SCP_CMD=(env SSH_ASKPASS="$ASKPASS_SCRIPT" SSH_ASKPASS_REQUIRE=force DISPLAY=none setsid -w scp -o PreferredAuthentications=password -o PubkeyAuthentication=no -o KbdInteractiveAuthentication=no)
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

# Generate the service file locally using a Here-Doc
cat <<EOF > "$SERVICE_FILE_TMP"
[Unit]
Description=Ohara Backend Service
After=network.target

[Service]
Type=simple
User=$("${SSH_CMD[@]}" "$SERVER" "whoami")
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
if [[ -n "$DEPLOY_PASSWORD" ]]; then
	ESCAPED_DEPLOY_PASSWORD="${DEPLOY_PASSWORD//\'/\'\"\'\"\'}"
	REMOTE_DIR_Q=$(printf '%q' "$REMOTE_DIR")
	SERVICE_REMOTE_TMP_Q=$(printf '%q' "/tmp/$SERVICE_NAME.service.tmp")
	SERVICE_REMOTE_PATH_Q=$(printf '%q' "/etc/systemd/system/$SERVICE_NAME.service")
	SERVICE_NAME_Q=$(printf '%q' "$SERVICE_NAME")
	BINARY_REMOTE_TMP_Q=$(printf '%q' "/tmp/$BINARY_NAME.tmp")
	BINARY_REMOTE_PATH_Q=$(printf '%q' "$REMOTE_DIR/$BINARY_NAME")

	"${SSH_CMD[@]}" -t "$SERVER" "printf '%s\n' '$ESCAPED_DEPLOY_PASSWORD' | sudo -S -p '' bash -se" <<EOF
set -euo pipefail
echo "Finalizing deployment as root..."
mkdir -p $REMOTE_DIR_Q
mv $SERVICE_REMOTE_TMP_Q $SERVICE_REMOTE_PATH_Q
systemctl daemon-reload
systemctl enable $SERVICE_NAME_Q
systemctl stop $SERVICE_NAME_Q || true
mv $BINARY_REMOTE_TMP_Q $BINARY_REMOTE_PATH_Q
chmod +x $BINARY_REMOTE_PATH_Q
systemctl start $SERVICE_NAME_Q
EOF
else
	"${SSH_CMD[@]}" -t "$SERVER" "
    echo 'Finalizing deployment as root...' && \
    sudo mkdir -p '$REMOTE_DIR' && \

    # 1. Move and setup the service file
    sudo mv '/tmp/$SERVICE_NAME.service.tmp' '/etc/systemd/system/$SERVICE_NAME.service' && \
    sudo systemctl daemon-reload && \
    sudo systemctl enable '$SERVICE_NAME' && \

    # 2. Update the binary
    sudo systemctl stop '$SERVICE_NAME' || true && \
    sudo mv '/tmp/$BINARY_NAME.tmp' '$REMOTE_DIR/$BINARY_NAME' && \
    sudo chmod +x '$REMOTE_DIR/$BINARY_NAME' && \

    # 3. Start it back up
    sudo systemctl start '$SERVICE_NAME'
"
fi

echo "Deployment complete! Your service is now fully managed by systemd."
