#!/bin/bash
# Rexec CLI Installer
# Usage: curl -fsSL https://rexec.pipeops.io/install-cli.sh | bash
#
# This script installs the rexec CLI tool for managing cloud terminals
# from your local machine.

set -e

# Configuration
REXEC_API="${REXEC_API:-https://rexec.pipeops.io}"
REPO="rexec/rexec"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

print_banner() {
    echo -e "${CYAN}${BOLD}"
    echo "██████╗ ███████╗██╗  ██╗███████╗ ██████╗"
    echo "██╔══██╗██╔════╝╚██╗██╔╝██╔════╝██╔════╝"
    echo "██████╔╝█████╗   ╚███╔╝ █████╗  ██║     "
    echo "██╔══██╗██╔══╝   ██╔██╗ ██╔══╝  ██║     "
    echo "██║  ██║███████╗██╔╝ ██╗███████╗╚██████╗"
    echo "╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚══════╝ ╚═════╝"
    echo -e "${NC}"
    echo -e "${BOLD}Cloud Terminal Environment CLI Installer${NC}"
    echo ""
}

detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            ;;
        *)
            echo -e "${RED}Unsupported OS: $OS${NC}"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    echo -e "${GREEN}Detected platform: ${PLATFORM}${NC}"
}

get_latest_version() {
    echo -e "${CYAN}Fetching latest version...${NC}"
    # Try GitHub releases first
    VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$VERSION" ]; then
        # Try rexec API for version
        VERSION=$(curl -s "${REXEC_API}/api/version" 2>/dev/null | grep -o '"version":"[^"]*"' | cut -d'"' -f4)
    fi
    
    if [ -z "$VERSION" ]; then
        VERSION="v1.0.0"
    fi
    
    echo -e "${GREEN}Latest version: ${VERSION}${NC}"
}

download_binaries() {
    SUFFIX="${PLATFORM}"
    if [ "$OS" = "windows" ]; then
        SUFFIX="${SUFFIX}.exe"
    fi

    TEMP_DIR=$(mktemp -d)
    CLI_PATH="${TEMP_DIR}/rexec"

    if [ "$OS" = "windows" ]; then
        CLI_PATH="${CLI_PATH}.exe"
    fi

    echo -e "${CYAN}Downloading rexec-cli...${NC}" >&2
    
    # Try GitHub releases first (most reliable for releases)
    CLI_URL="https://github.com/${REPO}/releases/download/${VERSION}/rexec-cli-${SUFFIX}"
    if curl -fsSL "$CLI_URL" -o "$CLI_PATH" 2>/dev/null; then
        # Verify it's a binary, not an HTML error page
        if file "$CLI_PATH" | grep -qE 'executable|ELF|Mach-O'; then
            echo -e "${GREEN}Downloaded from GitHub releases${NC}" >&2
            chmod +x "$CLI_PATH"
            echo "$TEMP_DIR"
            return 0
        fi
    fi
    
    # Try rexec.pipeops.io (direct binary hosting)
    CLI_URL="${REXEC_API}/downloads/rexec-cli-${SUFFIX}"
    if curl -fsSL "$CLI_URL" -o "$CLI_PATH" 2>/dev/null; then
        # Verify it's a binary, not an HTML error page
        if file "$CLI_PATH" | grep -qE 'executable|ELF|Mach-O'; then
            echo -e "${GREEN}Downloaded from rexec.pipeops.io${NC}" >&2
            chmod +x "$CLI_PATH"
            echo "$TEMP_DIR"
            return 0
        fi
    fi
    
    # If all else fails, provide build instructions
    echo "" >&2
    echo -e "${RED}═══════════════════════════════════════════════════════════${NC}" >&2
    echo -e "${RED}  CLI binary not available for download${NC}" >&2
    echo -e "${RED}═══════════════════════════════════════════════════════════${NC}" >&2
    echo "" >&2
    echo -e "${YELLOW}The CLI binary hasn't been released yet for ${PLATFORM}.${NC}" >&2
    echo "" >&2
    echo -e "${BOLD}Option 1: Build from source${NC}" >&2
    echo "" >&2
    echo "  # Install Go 1.21+ if not installed" >&2
    echo "  git clone https://github.com/${REPO}.git" >&2
    echo "  cd rexec" >&2
    echo "  go build -o rexec ./cmd/rexec-cli" >&2
    echo "  sudo mv rexec /usr/local/bin/" >&2
    echo "  sudo chmod +x /usr/local/bin/rexec" >&2
    echo "" >&2
    echo -e "${BOLD}Option 2: Wait for release${NC}" >&2
    echo "  Check: https://github.com/${REPO}/releases" >&2
    echo "" >&2
    rm -rf "$TEMP_DIR"
    exit 1
}

install_binaries() {
    TEMP_DIR=$1

    echo -e "${CYAN}Installing to ${INSTALL_DIR}...${NC}"

    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        SUDO=""
    else
        SUDO="sudo"
        echo -e "${YELLOW}Requires sudo access to install to ${INSTALL_DIR}${NC}"
    fi

    if [ "$OS" = "windows" ]; then
        $SUDO mv "${TEMP_DIR}/rexec.exe" "${INSTALL_DIR}/rexec.exe"
    else
        $SUDO mv "${TEMP_DIR}/rexec" "${INSTALL_DIR}/rexec"
    fi

    rm -rf "$TEMP_DIR"
}

verify_installation() {
    echo ""
    if command -v rexec &> /dev/null; then
        echo -e "${GREEN}${BOLD}✓ Installation successful!${NC}"
        echo ""
        rexec version
    else
        echo -e "${RED}Installation may have failed. Please check if ${INSTALL_DIR} is in your PATH.${NC}"
        exit 1
    fi
}

show_next_steps() {
    echo ""
    echo -e "${BOLD}Next steps:${NC}"
    echo ""
    echo -e "  1. Login to rexec:"
    echo -e "     ${CYAN}rexec login${NC}"
    echo ""
    echo -e "  2. List your terminals:"
    echo -e "     ${CYAN}rexec ls${NC}"
    echo ""
    echo -e "  3. Create a terminal:"
    echo -e "     ${CYAN}rexec create --name mydev${NC}"
    echo ""
    echo -e "  4. Connect to a terminal:"
    echo -e "     ${CYAN}rexec connect <terminal-id>${NC}"
    echo ""
    echo -e "  5. Launch interactive TUI:"
    echo -e "     ${CYAN}rexec -i${NC}"
    echo ""
    echo -e "${BOLD}Documentation:${NC} https://rexec.pipeops.io/agents"
    echo ""
}

main() {
    print_banner
    detect_platform
    get_latest_version
    TEMP_DIR=$(download_binaries)
    install_binaries "$TEMP_DIR"
    verify_installation
    show_next_steps
}

main "$@"
