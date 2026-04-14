#!/bin/sh

set -e

REPO="largeoliu/redmine-cli"
BINARY_NAME="redmine"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() {
    echo "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo "${RED}[ERROR]${NC} $1"
    exit 1
}

detect_os() {
    case "$(uname -s)" in
        Darwin*)    echo "darwin" ;;
        Linux*)     echo "linux" ;;
        CYGWIN*|MINGW*|MSYS*)    echo "windows" ;;
        *)          error "Unsupported OS: $(uname -s)" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)    echo "amd64" ;;
        arm64|aarch64)   echo "arm64" ;;
        *)               error "Unsupported architecture: $(uname -m)" ;;
    esac
}

get_latest_version() {
    latest_url="https://github.com/${REPO}/releases/latest"
    version=$(curl -sI "$latest_url" | grep -i "location:" | sed 's/.*\/tag\/\(.*\)/\1/' | tr -d '\r\n')
    if [ -z "$version" ]; then
        error "Failed to get latest version"
    fi
    echo "$version"
}

download_binary() {
    version="$1"
    os="$2"
    arch="$3"
    
    if [ "$os" = "windows" ]; then
        archive_name="${BINARY_NAME}_${version#v}_${os}_${arch}.zip"
    else
        archive_name="${BINARY_NAME}_${version#v}_${os}_${arch}.tar.gz"
    fi
    
    download_url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"
    
    info "Downloading ${archive_name}..."
    
    tmp_dir=$(mktemp -d)
    archive_path="${tmp_dir}/${archive_name}"
    
    if ! curl -fsSL "$download_url" -o "$archive_path"; then
        error "Failed to download ${archive_name}"
    fi
    
    info "Extracting..."
    
    if [ "$os" = "windows" ]; then
        if ! unzip -q "$archive_path" -d "$tmp_dir"; then
            error "Failed to extract archive"
        fi
    else
        if ! tar -xzf "$archive_path" -C "$tmp_dir"; then
            error "Failed to extract archive"
        fi
    fi
    
    echo "$tmp_dir"
}

install_binary() {
    tmp_dir="$1"
    
    if [ ! -d "$INSTALL_DIR" ]; then
        info "Creating install directory: $INSTALL_DIR"
        mkdir -p "$INSTALL_DIR"
    fi
    
    binary_path="${tmp_dir}/${BINARY_NAME}"
    
    if [ ! -f "$binary_path" ]; then
        error "Binary not found in archive"
    fi
    
    chmod +x "$binary_path"
    mv "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
    
    rm -rf "$tmp_dir"
}

check_path() {
    case ":$PATH:" in
        *":$INSTALL_DIR:"*)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

main() {
    info "Installing ${BINARY_NAME}..."
    
    os=$(detect_os)
    arch=$(detect_arch)
    version=$(get_latest_version)
    
    info "OS: ${os}, Arch: ${arch}, Version: ${version}"
    
    tmp_dir=$(download_binary "$version" "$os" "$arch")
    install_binary "$tmp_dir"
    
    info "Successfully installed ${BINARY_NAME} to ${INSTALL_DIR}"
    
    if ! check_path; then
        echo ""
        warn "${INSTALL_DIR} is not in your PATH"
        echo ""
        echo "Add the following to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo ""
        echo "    export PATH=\"\$PATH:${INSTALL_DIR}\""
        echo ""
        echo "Then restart your shell or run: source ~/.bashrc (or ~/.zshrc)"
    fi
    
    echo ""
    info "Run '${BINARY_NAME} --help' to get started"
}

main "$@"
