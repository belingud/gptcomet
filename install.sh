#!/bin/bash

# Exit script if any command fails
set -e

# Detect operating system
detect_os() {
    case "$(uname -s)" in
    Darwin*)
        echo "darwin"
        ;;
    Linux*)
        echo "linux"
        ;;
    *)
        echo "Unknown"
        ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
    x86_64 | amd64)
        echo "amd64"
        ;;
    arm64 | aarch64)
        echo "arm64"
        ;;
    *)
        echo "Unknown"
        ;;
    esac
}

# Get OS and architecture
OS=$(detect_os)
ARCH=$(detect_arch)

if [ "$OS" = "Unknown" ] || [ "$ARCH" = "Unknown" ]; then
    echo "Unsupported system or architecture"
    exit 1
fi

# Create temporary directory
TMP_DIR=$(mktemp -d)
cd $TMP_DIR

# Get latest version
LATEST_VERSION=$(curl -s https://api.github.com/repos/belingud/gptcomet/releases/latest | grep tag_name | cut -d'"' -f4)
VERSION=${LATEST_VERSION#v} # Remove 'v' prefix

# Build download URL based on OS and architecture
echo "Detected: $OS $ARCH"
DOWNLOAD_URL="https://github.com/belingud/gptcomet/releases/download/${LATEST_VERSION}/gptcomet_${VERSION}_${OS}_${ARCH}.tar.gz"

# Download latest release
echo "Downloading gptcomet version ${VERSION}..."
curl -sL "$DOWNLOAD_URL" -o "gptcomet_archive" || {
    echo "Failed to download from: $DOWNLOAD_URL"
    exit 1
}

# Extract files
tar xzf "gptcomet_archive"

# Create target directory if not exists
if [ ! -d ~/.local/bin ]; then
    mkdir -p ~/.local/bin
fi

# Copy executable file
cp gptcomet ~/.local/bin/gptcomet
ln -s ~/.local/bin/gptcomet ~/.local/bin/gmsg
chmod +x ~/.local/bin/gptcomet

# Clean up temporary files
cd
rm -rf $TMP_DIR

echo "Installation completed! gptcomet has been installed to ~/.local/bin/"
echo "  + gmsg"
echo "  + gptcomet"
echo "Please ensure ~/.local/bin is in your PATH"
