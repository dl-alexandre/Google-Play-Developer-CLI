#!/bin/bash
set -e

REPO="dl-alexandre/gpd"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH="x86_64" ;;
  amd64) ARCH="x86_64" ;;
  arm64) ARCH="arm64" ;;
  aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  darwin) OS="darwin" ;;
  linux) OS="linux" ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Get latest release
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
VERSION="${LATEST#v}"

if [ -z "$VERSION" ]; then
  echo "Failed to get latest version"
  exit 1
fi

echo "Installing gpd ${LATEST}..."

# Download and extract
FILENAME="gpd_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${FILENAME}"

TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

curl -fsSL "$URL" -o "$TMP_DIR/$FILENAME"
tar -xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"

# Install
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_DIR/gpd" "$INSTALL_DIR/gpd"
else
  echo "Installing to $INSTALL_DIR (requires sudo)..."
  sudo mv "$TMP_DIR/gpd" "$INSTALL_DIR/gpd"
fi

chmod +x "$INSTALL_DIR/gpd"

echo "gpd installed to $INSTALL_DIR/gpd"
"$INSTALL_DIR/gpd" version
