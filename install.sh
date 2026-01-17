#!/bin/bash
set -e

REPO="https://github.com/gtgrthrst/system-monitor-api.git"
INSTALL_DIR="/opt/sysinfo-api"
SERVICE_NAME="sysinfo-api"

echo "=== System Monitor API Installer ==="

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Installing Go..."
    wget -q https://go.dev/dl/go1.22.0.linux-amd64.tar.gz -O /tmp/go.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf /tmp/go.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    rm /tmp/go.tar.gz
fi

echo "Go version: $(go version)"

# Clone or update repo
if [ -d "$INSTALL_DIR" ]; then
    echo "Updating existing installation..."
    cd "$INSTALL_DIR"
    git pull
else
    echo "Cloning repository..."
    git clone "$REPO" "$INSTALL_DIR"
    cd "$INSTALL_DIR"
fi

# Build
echo "Building..."
go build -o sysinfo-api

# Create systemd service
echo "Creating systemd service..."
cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=System Monitor API
After=network.target

[Service]
Type=simple
ExecStart=${INSTALL_DIR}/sysinfo-api
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
systemctl daemon-reload
systemctl enable ${SERVICE_NAME}
systemctl restart ${SERVICE_NAME}

echo ""
echo "=== Installation Complete ==="
echo "Service status: $(systemctl is-active ${SERVICE_NAME})"
echo "API endpoints:"
echo "  http://localhost:8088/health"
echo "  http://localhost:8088/api/system"
