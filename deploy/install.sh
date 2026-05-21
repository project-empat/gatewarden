#!/bin/bash
set -euo pipefail

REPO="${REPO:-yourorg/gatewarden}"
VERSION="${VERSION:-latest}"
BINDIR="${BINDIR:-/usr/local/bin}"

echo "Installing Gatewarden Agent..."

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *)       echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Download agent binary
URL="https://github.com/${REPO}/releases/download/${VERSION}/gatewarden-agent-linux-${ARCH}"

echo "Downloading from $URL"
curl -fsSL "$URL" -o "${BINDIR}/gatewarden-agent"
chmod +x "${BINDIR}/gatewarden-agent"

# Install systemd service
if command -v systemctl &> /dev/null; then
  echo "Installing systemd service..."
  curl -fsSL "https://raw.githubusercontent.com/${REPO}/${VERSION}/deploy/systemd/gatewarden-agent.service" \
    -o /etc/systemd/system/gatewarden-agent.service
  systemctl daemon-reload
  echo "Service installed. Configure /etc/systemd/system/gatewarden-agent.service and run:"
  echo "  sudo systemctl enable --now gatewarden-agent"
fi

echo "Gatewarden Agent installed successfully!"
echo "Binary: ${BINDIR}/gatewarden-agent"
