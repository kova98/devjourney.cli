#!/bin/sh
set -e

# Configuration
GITHUB_REPO="kova98/devjourney.cli"
BINARY_NAME="devjourney.cli"
INSTALL_NAME="devjourney"
INSTALL_DIR="/usr/local/bin"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Detect OS and architecture
detect_os() {
    case "$(uname -s)" in
        Linux*)     os=linux;;
        Darwin*)    os=darwin;;
        *)         echo "Unsupported operating system: $(uname -s)"; exit 1;;
    esac
    echo $os
}

detect_arch() {
    case "$(uname -m)" in
        x86_64*)    arch=amd64;;
        i386*)      arch=386;;
        i686*)      arch=386;;
        arm64*)     arch=arm64;;
        aarch64*)   arch=arm64;;
        *)          echo "Unsupported architecture: $(uname -m)"; exit 1;;
    esac
    echo $arch
}

# Get the latest release version from GitHub
get_latest_version() {
    curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | 
    grep '"tag_name":' | 
    sed -E 's/.*"([^"]+)".*/\1/'
}

# Main installation
main() {
    echo "🔍 Detecting system information..."
    OS=$(detect_os)
    ARCH=$(detect_arch)
    VERSION=$(get_latest_version)
    
    if [ -z "$VERSION" ]; then
        echo "${RED}❌ Failed to get latest version${NC}"
        exit 1
    fi
    
    echo "📦 Downloading ${BINARY_NAME} ${VERSION} for ${OS}/${ARCH}..."
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${BINARY_NAME}_${VERSION#v}_${OS}_${ARCH}.tar.gz"
    
    # Create a temporary directory
    TMP_DIR=$(mktemp -d)
    
    # Download and extract the binary
    if ! curl -sL "$DOWNLOAD_URL" | tar xz -C "${TMP_DIR}"; then
        echo "${RED}❌ Failed to download or extract binary${NC}"
        rm -rf "${TMP_DIR}"
        exit 1
    fi
    
    # Make it executable
    chmod +x "${TMP_DIR}/${BINARY_NAME}"
    
    # Move to installation directory with new name (requires sudo)
    echo "📥 Installing to ${INSTALL_DIR}..."
    if ! sudo mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${INSTALL_NAME}"; then
        echo "${RED}❌ Failed to install binary${NC}"
        rm -rf "${TMP_DIR}"
        exit 1
    fi
    
    # Clean up
    rm -rf "${TMP_DIR}"
    
    # Verify installation
    if command -v ${INSTALL_NAME} >/dev/null 2>&1; then
        echo "${GREEN}✅ Successfully installed ${INSTALL_NAME} ${VERSION}${NC}"
        echo "Run '${INSTALL_NAME} --help' to get started"
    else
        echo "${RED}❌ Installation failed${NC}"
        exit 1
    fi
}

main 