#!/bin/bash

# Exit script if any command fails
set -e

# Show usage instructions
show_usage() {
    echo "Usage: $0 [-v VERSION]"
    echo "  -v VERSION  Install specific version (e.g. 0.4.2)"
    echo
    echo "Examples:"
    echo "  $0              # Install latest version"
    echo "  $0 -v 0.4.2     # Install version 0.4.2"
    exit 1
}

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

# Parse arguments
VERSION=""
while getopts ":v:" opt; do
  case $opt in
    v)
      VERSION="$OPTARG"
      ;;
    \?)
      echo "Error: Invalid option -$OPTARG" >&2
      show_usage
      ;;
    :)
      echo "Error: Option -$OPTARG requires an argument." >&2
      show_usage
      ;;
  esac
done

# Validate no extra arguments
shift $((OPTIND-1))
if [ $# -ne 0 ]; then
    echo "Error: Unexpected arguments: $@" >&2
    show_usage
fi

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

# Get version
if [ -z "$VERSION" ]; then
    # Get latest version if no version specified
    LATEST_VERSION=$(curl -s https://api.github.com/repos/belingud/gptcomet/releases/latest | grep tag_name | cut -d'"' -f4)
    VERSION=${LATEST_VERSION#v} # Remove 'v' prefix
else
    # Validate specified version exists
    VERSION=${VERSION#v} # Remove 'v' prefix if present
    if ! curl -s -o /dev/null -I -w "%{http_code}" https://api.github.com/repos/belingud/gptcomet/releases/tags/v$VERSION | grep -q 200; then
        echo "Error: Version v$VERSION not found"
        exit 1
    fi
    LATEST_VERSION="v$VERSION"
fi

# Build download URL based on OS and architecture
echo "Detected: $OS $ARCH"
DOWNLOAD_URL="https://github.com/belingud/gptcomet/releases/download/${LATEST_VERSION}/gptcomet_${VERSION}_${OS}_${ARCH}.tar.gz"

# Download specified release
echo "Downloading gptcomet version ${VERSION}..."
curl -sL "$DOWNLOAD_URL" -o "gptcomet_archive" || {
    echo "Error: Failed to download from: $DOWNLOAD_URL"
    exit 1
}

# Extract files
tar xzf "gptcomet_archive"

# Create target directory if not exists
if [ ! -d ~/.local/bin ]; then
    mkdir -p ~/.local/bin
fi

# Copy executable file
cp gptcomet ~/.local/bin/gmsg
ln -s ~/.local/bin/gmsg ~/.local/bin/gptcomet
chmod +x ~/.local/bin/gmsg

# Clean up temporary files
cd
rm -rf $TMP_DIR

echo "Installation completed! gptcomet has been installed to ~/.local/bin/"
echo "  + gmsg"
echo "  + gptcomet"
echo "Please ensure ~/.local/bin is in your PATH"
