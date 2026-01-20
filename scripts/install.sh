#!/bin/sh

# muxbee Installer
# Downloads the correct binary for your platform to the current directory
#
# Usage: curl -fsSL https://raw.githubusercontent.com/tobocop2/muxbee/main/scripts/install.sh | sh

REPO="tobocop2/muxbee"
BINARY_NAME="muxbee"

info() { echo "==> $1"; }
error() { echo "Error: $1" >&2; exit 1; }

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac

    case "$OS" in
        linux|darwin) ;;
        *) error "Unsupported OS: $OS" ;;
    esac

    PLATFORM="${OS}-${ARCH}"
}

# Get latest release version
get_latest_version() {
    VERSION=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        error "No releases found"
    fi
}

# Download binary
download() {
    URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}-${PLATFORM}"
    info "Downloading muxbee ${VERSION} for ${PLATFORM}..."

    if ! curl -fSL "$URL" -o muxbee; then
        error "Download failed"
    fi
    chmod +x muxbee

    info "Downloaded: ./muxbee"
    echo ""
    echo "Move it to your PATH, e.g.:"
    echo "  sudo mv muxbee /usr/local/bin/    # system-wide"
    echo "  mv muxbee ~/.local/bin/           # user only"
    echo ""
    ./muxbee --version
}

detect_platform
get_latest_version
download
