#!/usr/bin/env bash
set -e

REPO="0DayMonxrch/vaultify"
INSTALL_DIR="/usr/local/bin"

# Detect OS and Arch
OS="$(uname -s)"
ARCH="$(uname -m)"

if [ "$OS" = "Linux" ]; then
    OS_NAME="Linux"
elif [ "$OS" = "Darwin" ]; then
    OS_NAME="Darwin"
else
    echo "Unsupported OS: $OS"
    exit 1
fi

if [ "$ARCH" = "x86_64" ]; then
    ARCH_NAME="x86_64"
elif [ "$ARCH" = "arm64" ] || [ "$ARCH" = "aarch64" ]; then
    ARCH_NAME="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

echo "Fetching latest version of Vaultify..."
LATEST_URL=$(curl -w "%{url_effective}" -I -L -s -S https://github.com/$REPO/releases/latest -o /dev/null)
VERSION=$(basename $LATEST_URL)

if [ -z "$VERSION" ]; then
    echo "Failed to fetch latest version."
    exit 1
fi

FILENAME="vaultify_${OS_NAME}_${ARCH_NAME}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$FILENAME"

echo "Downloading $DOWNLOAD_URL..."
TMP_DIR=$(mktemp -d)
curl -sSL "$DOWNLOAD_URL" | tar -xz -C "$TMP_DIR" vaultify

echo "Installing to $INSTALL_DIR (might require sudo password)..."
sudo mv "$TMP_DIR/vaultify" "$INSTALL_DIR/"
sudo chmod +x "$INSTALL_DIR/vaultify"

rm -rf "$TMP_DIR"
echo "Vaultify successfully installed! Run 'vaultify' to get started."
