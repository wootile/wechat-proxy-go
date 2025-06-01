#!/bin/bash

# WeChat API Proxy Server - Simple Deployment Script
set -e

echo "🚀 Deploying WeChat API Proxy Server..."

# Configuration
DEPLOY_DIR="/opt/wechat-proxy"
SERVICE_NAME="wechat-proxy"
PROXY_PORT=${PROXY_PORT:-8080}

echo "📋 Configuration:"
echo "  Port: $PROXY_PORT"
echo "  Deploy dir: $DEPLOY_DIR"

# Check root permissions
if [ "$EUID" -ne 0 ]; then
    echo "❌ Root permissions required"
    echo "💡 Run with: sudo $0"
    exit 1
fi

# Find binary
BINARY_PATH=""
if [ -f "wechat-proxy" ]; then
    BINARY_PATH="wechat-proxy"
elif [ -f "./wechat-proxy" ]; then
    BINARY_PATH="./wechat-proxy"
else
    echo "❌ Binary 'wechat-proxy' not found in current directory"
    echo "💡 Please ensure 'wechat-proxy' executable is in the current directory"
    exit 1
fi

echo "✅ Found binary: $BINARY_PATH"

# Check port availability
if command -v netstat > /dev/null && netstat -tln 2>/dev/null | grep -q ":$PROXY_PORT "; then
    echo "⚠️  Port $PROXY_PORT is in use"
    read -p "Continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Stop existing service
if systemctl is-active --quiet $SERVICE_NAME 2>/dev/null; then
    echo "🛑 Stopping existing service..."
    systemctl stop $SERVICE_NAME
fi

# Create deployment directory and copy files
echo "📁 Setting up deployment..."
mkdir -p $DEPLOY_DIR
cp "$BINARY_PATH" "$DEPLOY_DIR/wechat-proxy"
chmod +x "$DEPLOY_DIR/wechat-proxy"

# Create systemd service
echo "🔧 Creating systemd service..."
cat > /etc/systemd/system/$SERVICE_NAME.service <<EOF
[Unit]
Description=WeChat API Proxy Server
After=network.target

[Service]
Type=simple
User=nobody
ExecStart=$DEPLOY_DIR/wechat-proxy
Restart=always
Environment=PROXY_PORT=$PROXY_PORT

[Install]
WantedBy=multi-user.target
EOF

# Start service
echo "🚀 Starting service..."
systemctl daemon-reload
systemctl enable $SERVICE_NAME
systemctl start $SERVICE_NAME

# Wait and check status
sleep 2
if systemctl is-active --quiet $SERVICE_NAME; then
    echo "✅ Deployment successful!"
    echo ""
    echo "🌐 Service info:"
    echo "  Status: $(systemctl is-active $SERVICE_NAME)"
    echo "  Port: $PROXY_PORT"
    echo ""
    echo "📝 Commands:"
    echo "  Status: systemctl status $SERVICE_NAME"
    echo "  Logs: journalctl -f -u $SERVICE_NAME"
    echo "  Restart: systemctl restart $SERVICE_NAME"
    echo ""
    echo "🔗 Test:"
    echo "  curl -x http://localhost:$PROXY_PORT https://api.weixin.qq.com"
else
    echo "❌ Deployment failed!"
    systemctl status $SERVICE_NAME --no-pager
    journalctl -u $SERVICE_NAME --no-pager --lines=10
    exit 1
fi 