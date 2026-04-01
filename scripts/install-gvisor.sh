#!/bin/bash

# Rexec - gVisor (runsc) Sandbox Installer for Linux
# This script installs gVisor and configures Docker to use it as a secure runtime.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 0. Safety Check: Only run on Linux
if [ "$(uname -s)" != "Linux" ]; then
    echo -e "${RED}ERROR: This script is intended for Linux servers only.${NC}"
    echo -e "You are currently on $(uname -s). No changes were made."
    exit 1
fi

echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}      Rexec gVisor Sandbox Installer (Linux)      ${NC}"
echo -e "${BLUE}==================================================${NC}"

# 1. Check Architecture
ARCH=$(uname -m)
URL_ARCH=$ARCH
if [ "$ARCH" == "x86_64" ]; then
    URL_ARCH="x86_64"
elif [ "$ARCH" == "aarch64" ]; then
    URL_ARCH="arm64"
else
    echo -e "${RED}Unsupported architecture: $ARCH${NC}"
    exit 1
fi

# 2. Download gVisor Binaries
echo -e "${BLUE}Downloading gVisor (runsc) for $ARCH...${NC}"
URL="https://storage.googleapis.com/gvisor/releases/release/latest/${URL_ARCH}"

# Create temp dir
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

curl -sSL "${URL}/runsc" -o runsc
curl -sSL "${URL}/runsc.sha256" -o runsc.sha256
curl -sSL "${URL}/containerd-shim-runsc-v1" -o containerd-shim-runsc-v1
curl -sSL "${URL}/containerd-shim-runsc-v1.sha256" -o containerd-shim-runsc-v1.sha256

# 3. Verify Checksums
echo -n "Verifying checksums... "
# Use grep to only check the files we actually have
sha256sum -c runsc.sha256 --status || (echo -e "${RED}runsc checksum failed${NC}" && exit 1)
sha256sum -c containerd-shim-runsc-v1.sha256 --status || (echo -e "${RED}shim checksum failed${NC}" && exit 1)
echo -e "${GREEN}PASS${NC}"

# 4. Install Binaries
echo -e "${BLUE}Installing to /usr/local/bin...${NC}"
chmod +x runsc containerd-shim-runsc-v1
sudo mv runsc containerd-shim-runsc-v1 /usr/local/bin/

# 5. Configure Docker Runtime
echo -e "${BLUE}Configuring Docker to use runsc...${NC}"
# This command updates /etc/docker/daemon.json automatically
sudo /usr/local/bin/runsc install

# 6. Restart Docker
echo -e "${BLUE}Restarting Docker service...${NC}"
if systemctl is-active --quiet docker; then
    sudo systemctl restart docker
    echo -e "${GREEN}Docker restarted.${NC}"
else
    echo -e "${YELLOW}Docker service not active. Please restart manually.${NC}"
fi

# Cleanup
cd - > /dev/null
rm -rf "$TEMP_DIR"

echo -e "${BLUE}--------------------------------------------------${NC}"
echo -e "${GREEN}Success! gVisor is now installed on your Linux node.${NC}"
echo -e "Test it with: ${YELLOW}docker run --rm --runtime=runsc alpine uname -a${NC}"
echo -e "${BLUE}==================================================${NC}"
