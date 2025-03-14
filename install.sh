#!/bin/bash

# Configuration
REPO="lucasvmiguel/k8run"
BINARY_NAME="k8run"
VERSION="0.0.13"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [[ "$ARCH" == "x86_64" ]]; then
    ARCH="amd64"
elif [[ "$ARCH" == "aarch64" ]]; then
    ARCH="arm64"
elif [[ "$ARCH" == "arm64" ]]; then
    ARCH="arm64"
else
    echo "❌ Unsupported architecture: $ARCH"
    exit 1
fi

# Get release data for the specified version
echo "📡 Fetching release $VERSION for $OS-$ARCH from $REPO..."
RELEASE_DATA=$(curl -s "https://api.github.com/repos/$REPO/releases/tags/v$VERSION")

# Extract download URL for the correct OS and architecture
DOWNLOAD_URL=$(echo "$RELEASE_DATA" | grep "browser_download_url" | grep "${BINARY_NAME}-${OS}-${ARCH}" | cut -d '"' -f 4)

if [[ -z "$DOWNLOAD_URL" ]]; then
    echo "❌ No compatible release found for version $VERSION ($OS-$ARCH)."
    exit 1
fi

# Download the binary
echo "⬇️  Downloading $BINARY_NAME v$VERSION from $DOWNLOAD_URL..."
curl -L -o "$BINARY_NAME" "$DOWNLOAD_URL"

# Make it executable
chmod +x "$BINARY_NAME"

# Move to the install directory
echo "🚀 Installing $BINARY_NAME to $INSTALL_DIR..."
sudo mv "$BINARY_NAME" "$INSTALL_DIR/"

# Confirm installation
if ! command -v "$BINARY_NAME" &> /dev/null; then
    echo "⚠️ Installed, but not found in PATH. Adding it to PATH..."
    echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> ~/.bashrc
    source ~/.bashrc
fi

echo "✅ $BINARY_NAME v$VERSION installed successfully!"
echo "👉 Run it using: '$BINARY_NAME'"

# Verify installation
"$BINARY_NAME" --version 2>/dev/null || echo "⚠️ Warning: Unable to verify installation."