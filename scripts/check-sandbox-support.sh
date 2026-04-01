#!/bin/bash

# Rexec - Sandbox Readiness Checker (Multi-platform)
# This script verifies if the host meets the requirements for high-isolation sandboxing.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

OS=$(uname -s)

echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}      Rexec Sandbox Readiness Check ($OS)       ${NC}"
echo -e "${BLUE}==================================================${NC}"

if [ "$OS" == "Darwin" ]; then
    echo -e "${BLUE}Detected macOS environment.${NC}"
    echo -e "Note: macOS uses Apple's Virtualization.framework via Docker Desktop."
    
    # 1. Check Docker Desktop / Sandbox support
    echo -n "Checking Native Docker Sandbox... "
    if docker sandbox version >/dev/null 2>&1; then
        SBX_VER=$(docker sandbox version | grep "sandboxd version" | awk '{print $3}')
        echo -e "${GREEN}PASS ($SBX_VER)${NC}"
    else
        echo -e "${RED}FAIL${NC}"
        echo -e "  ${YELLOW}Suggestion: Ensure Docker Desktop is running and up-to-date (v4.58+).${NC}"
    fi

    # 2. Check sbx CLI (standalone)
    echo -n "Checking sbx CLI (standalone)... "
    if command -v sbx >/dev/null 2>&1; then
        SBX_CLI_VER=$(sbx --version | awk '{print $3}' 2>/dev/null || echo "Installed")
        echo -e "${GREEN}PASS ($SBX_CLI_VER)${NC}"
    else
        echo -e "${YELLOW}NOT INSTALLED (Optional)${NC}"
        echo -e "    ${BLUE}To Install: brew install docker/tap/sbx${NC}"
    fi

else
    # Linux Path
    echo -e "${BLUE}Detected Linux environment.${NC}"

    # 1. Check Hardware Virtualization
    echo -n "Checking CPU Virtualization (VT-x/AMD-V)... "
    if lscpu 2>/dev/null | grep -q "Virtualization:"; then
        VIRT_TYPE=$(lscpu | grep "Virtualization:" | awk '{print $2}')
        echo -e "${GREEN}PASS (${VIRT_TYPE})${NC}"
    else
        if grep -Eoc "vmx|svm" /proc/cpuinfo >/dev/null 2>&1; then
            echo -e "${GREEN}PASS (detected via cpuinfo)${NC}"
        else
            echo -e "${RED}FAIL${NC}"
        fi
    fi

    # 2. Check KVM Kernel Module
    echo -n "Checking KVM Kernel Modules... "
    if lsmod | grep -q "kvm"; then
        echo -e "${GREEN}PASS${NC}"
    else
        echo -e "${RED}FAIL${NC}"
    fi

    # 3. Check High-Isolation Providers (Linux)
    echo -e "${BLUE}Checking Sandbox Providers:${NC}"
    echo -n "  - gVisor (runsc): "
    if command -v runsc >/dev/null 2>&1; then
        echo -e "${GREEN}PASS${NC}"
    else
        echo -e "${YELLOW}NOT INSTALLED${NC}"
    fi
fi

# 4. Check Docker Engine Version
echo -n "Checking Docker Engine... "
if command -v docker >/dev/null 2>&1; then
    DOCKER_VER=$(docker version --format '{{.Server.Version}}' 2>/dev/null || echo "0.0.0")
    echo -e "${GREEN}PASS ($DOCKER_VER)${NC}"
else
    echo -e "${RED}NOT INSTALLED${NC}"
fi

echo -e "${BLUE}--------------------------------------------------${NC}"
echo -e "${BLUE}Conclusion:${NC}"
if [ "$OS" == "Darwin" ] && docker sandbox version >/dev/null 2>&1; then
    echo -e "${GREEN}  Your Mac is READY for Native Docker Sandboxes!${NC}"
    echo -e "  You can use 'docker sandbox' or 'sbx' for isolation."
elif command -v runsc >/dev/null 2>&1; then
    echo -e "${GREEN}  Your node is READY for High-Isolation Sandboxes (gVisor)!${NC}"
else
    echo -e "${YELLOW}  Your environment needs setup before high-isolation${NC}"
    echo -e "${YELLOW}  sandboxes can be fully utilized.${NC}"
fi
echo -e "${BLUE}==================================================${NC}"
