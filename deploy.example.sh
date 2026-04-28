#!/bin/bash
set -e

SERVER=
REMOTE_DIR=
BINARY_NAME=
SERVICE_NAME=

echo "🚀 Building the binary..."
npm --prefix frontend install
npm --prefix frontend run build:embed

cd backend
GOOS=linux GOARCH=amd64 go build -o ../$BINARY_NAME ./cmd
cd ..

echo "🛑 Stopping the service on the VPS..."
ssh $SERVER "systemctl stop $SERVICE_NAME"

echo "🔄 Syncing new binary to VPS..."
scp ./$BINARY_NAME $SERVER:$REMOTE_DIR/$BINARY_NAME

echo "🔧 Setting executable permissions..."
ssh $SERVER "chmod +x $REMOTE_DIR/$BINARY_NAME"

echo "🟢 Restarting the service..."
ssh $SERVER "systemctl start $SERVICE_NAME" 

echo "✅ Deployment complete!"