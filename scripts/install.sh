#!/bin/bash
set -euo pipefail

BINARY_NAME="yartime-client"
INSTALL_PATH="/usr/local/bin/${BINARY_NAME}"
SERVICE_NAME="yartime-client"
SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}.service"

usage() {
    echo "Usage: $0 --server-url=<url> [--client-id=<id>]"
    echo ""
    echo "Options:"
    echo "  --server-url   Yartime server URL (required)"
    echo "  --client-id    Client identifier (default: default-client)"
    exit 1
}

SERVER_URL=""
CLIENT_ID="default-client"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --server-url=*)
            SERVER_URL="${1#*=}"
            shift
            ;;
        --client-id=*)
            CLIENT_ID="${1#*=}"
            shift
            ;;
        *)
            usage
            ;;
    esac
done

if [[ -z "$SERVER_URL" ]]; then
    echo "error: --server-url is required"
    usage
fi

if [[ $EUID -ne 0 ]]; then
    echo "error: this script must be run as root (use sudo)"
    exit 1
fi

echo "Downloading yartime client from ${SERVER_URL}..."
curl -fsSL -o "${INSTALL_PATH}" "${SERVER_URL}/client/binary?os=linux&arch=amd64"
chmod +x "${INSTALL_PATH}"

echo "Creating systemd service..."
cat > "${SERVICE_PATH}" <<EOF
[Unit]
Description=Yartime Client
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=${INSTALL_PATH} --server-url=${SERVER_URL} --client-id=${CLIENT_ID}
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

chmod 644 "${SERVICE_PATH}"

echo "Enabling and starting service..."
systemctl daemon-reload
systemctl enable "${SERVICE_NAME}"
systemctl start "${SERVICE_NAME}"

echo ""
echo "Installation complete!"
echo "  Binary: ${INSTALL_PATH}"
echo "  Service: ${SERVICE_NAME}"
echo "  Server: ${SERVER_URL}"
echo "  Client ID: ${CLIENT_ID}"
echo ""
echo "Useful commands:"
echo "  systemctl status ${SERVICE_NAME}    # Check status"
echo "  journalctl -u ${SERVICE_NAME} -f    # View logs"
echo "  systemctl stop ${SERVICE_NAME}      # Stop"
echo "  systemctl restart ${SERVICE_NAME}   # Restart"
