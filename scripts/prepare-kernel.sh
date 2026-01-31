#!/bin/bash
# Download and prepare Firecracker kernel
# Usage: ./prepare-kernel.sh

set -e

KERNEL_VERSION="${1:-6.1}"
KERNEL_DIR="${FIRECRACKER_KERNEL_PATH:-/opt/firecracker}"
KERNEL_FILE="vmlinux.bin"

echo "🔨 Preparing Firecracker kernel..."

# Create kernel directory
sudo mkdir -p "${KERNEL_DIR}"

# Check if kernel already exists
if [ -f "${KERNEL_DIR}/${KERNEL_FILE}" ]; then
    echo "✅ Kernel already exists at ${KERNEL_DIR}/${KERNEL_FILE}"
    exit 0
fi

echo "📥 Downloading Firecracker kernel..."
echo ""
echo "⚠️  Note: Firecracker doesn't provide pre-built kernels."
echo "   You need to build the kernel yourself or use a compatible kernel."
echo ""
echo "Options:"
echo "1. Build from source (recommended)"
echo "   See: https://github.com/firecracker-microvm/firecracker/blob/main/docs/rootfs-and-kernels-setup.md"
echo ""
echo "2. Use a compatible kernel from your distribution"
echo "   Extract vmlinux from your kernel package"
echo ""
echo "3. Download from a trusted source"
echo "   (Not recommended for production)"
echo ""

# Try to find existing kernel
if [ -f "/boot/vmlinuz-$(uname -r)" ]; then
    echo "💡 Found system kernel at /boot/vmlinuz-$(uname -r)"
    echo "   You can extract vmlinux using:"
    echo "   extract-vmlinux /boot/vmlinuz-$(uname -r) > ${KERNEL_DIR}/${KERNEL_FILE}"
    echo ""
    echo "   Install extract-vmlinux:"
    echo "   wget https://raw.githubusercontent.com/torvalds/linux/master/scripts/extract-vmlinux"
    echo "   chmod +x extract-vmlinux"
fi

# Check if extract-vmlinux script exists
if command -v extract-vmlinux &> /dev/null; then
    echo "🔧 Extracting vmlinux from system kernel..."
    if [ -f "/boot/vmlinuz-$(uname -r)" ]; then
        sudo extract-vmlinux /boot/vmlinuz-$(uname -r) > "/tmp/vmlinux"
        sudo mv "/tmp/vmlinux" "${KERNEL_DIR}/${KERNEL_FILE}"
        sudo chmod 644 "${KERNEL_DIR}/${KERNEL_FILE}"
        echo "✅ Kernel extracted to ${KERNEL_DIR}/${KERNEL_FILE}"
        exit 0
    fi
fi

echo "❌ Could not automatically prepare kernel."
echo ""
echo "Manual steps:"
echo "1. Download or build a Firecracker-compatible kernel"
echo "2. Place it at: ${KERNEL_DIR}/${KERNEL_FILE}"
echo "3. Ensure it's readable: chmod 644 ${KERNEL_DIR}/${KERNEL_FILE}"
echo ""
echo "For building from source, see:"
echo "https://github.com/firecracker-microvm/firecracker/blob/main/docs/rootfs-and-kernels-setup.md"

exit 1
