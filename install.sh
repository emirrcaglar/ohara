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

# Verify
if command -v ohara >/dev/null 2>&1; then
    echo "✓ Ohara installed successfully!"
    echo ""
    echo "Run 'ohara' to start the server"
    echo "Data will be stored in ./app-data (relative to where you run it)"
    echo ""
    echo "Tip: Use -data flag to specify a custom data directory:"
    echo "  ohara -data /path/to/data"
else
    echo "✓ Binary installed to $INSTALL_DIR/ohara"
    echo ""
    echo "If 'ohara' command not found, add to PATH:"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
fi
