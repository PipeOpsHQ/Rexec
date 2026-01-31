#!/bin/bash
# Build Firecracker rootfs images
# Usage: ./build-rootfs.sh <distro> [version]
# Example: ./build-rootfs.sh ubuntu 24.04

set -e

DISTRO="${1:-ubuntu}"
VERSION="${2:-24.04}"
ROOTFS_DIR="/tmp/rexec-rootfs-${DISTRO}-${VERSION}"
IMAGE_NAME="${DISTRO}-${VERSION}.ext4"
IMAGE_SIZE="${3:-2048}" # Size in MB
OUTPUT_DIR="${FIRECRACKER_ROOTFS_PATH:-/var/lib/rexec/firecracker/rootfs/images}"

echo "🔨 Building rootfs image for ${DISTRO} ${VERSION}"

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Clean up any existing rootfs directory
if [ -d "${ROOTFS_DIR}" ]; then
    echo "🧹 Cleaning up existing rootfs directory..."
    sudo rm -rf "${ROOTFS_DIR}"
fi

# Create rootfs directory
mkdir -p "${ROOTFS_DIR}"

# Install debootstrap if not available
if ! command -v debootstrap &> /dev/null; then
    echo "📦 Installing debootstrap..."
    sudo apt-get update
    sudo apt-get install -y debootstrap
fi

# Bootstrap base system
echo "📥 Bootstrapping ${DISTRO} ${VERSION}..."
case "${DISTRO}" in
    ubuntu)
        sudo debootstrap --arch=amd64 --variant=minbase "${VERSION}" "${ROOTFS_DIR}" http://archive.ubuntu.com/ubuntu/
        ;;
    debian)
        sudo debootstrap --arch=amd64 --variant=minbase "${VERSION}" "${ROOTFS_DIR}" http://deb.debian.org/debian/
        ;;
    *)
        echo "❌ Unsupported distro: ${DISTRO}"
        echo "Supported: ubuntu, debian"
        exit 1
        ;;
esac

# Mount required filesystems
echo "🔗 Mounting required filesystems..."
sudo mount --bind /dev "${ROOTFS_DIR}/dev"
sudo mount --bind /proc "${ROOTFS_DIR}/proc"
sudo mount --bind /sys "${ROOTFS_DIR}/sys"

# Configure chroot environment
echo "⚙️  Configuring rootfs..."

# Basic system configuration
sudo chroot "${ROOTFS_DIR}" /bin/bash <<EOF
set -e

# Set hostname
echo "rexec-vm" > /etc/hostname

# Configure networking (DHCP)
cat > /etc/network/interfaces <<NETCONF
auto lo
iface lo inet loopback

auto eth0
iface eth0 inet dhcp
NETCONF

# Install essential packages
apt-get update
apt-get install -y \
    systemd \
    systemd-sysv \
    openssh-server \
    curl \
    wget \
    vim \
    less \
    sudo \
    net-tools \
    iproute2 \
    iputils-ping \
    dnsutils \
    ca-certificates

# Create rexec user
useradd -m -s /bin/bash -G sudo rexec
echo "rexec:rexec" | chpasswd
echo "rexec ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Configure SSH
mkdir -p /home/rexec/.ssh
chmod 700 /home/rexec/.ssh
chown -R rexec:rexec /home/rexec

# Enable SSH service
systemctl enable ssh

# Set root password (for emergency access)
echo "root:root" | chpasswd

# Configure timezone
ln -sf /usr/share/zoneinfo/UTC /etc/localtime

# Clean up
apt-get clean
rm -rf /var/lib/apt/lists/*
EOF

# Install guest agent
echo "📦 Installing guest agent..."
GUEST_AGENT_BINARY="rexec-guest-agent"
GUEST_AGENT_PATH="${ROOTFS_DIR}/usr/local/bin/${GUEST_AGENT_BINARY}"

# Check if guest agent binary exists
if [ ! -f "./bin/${GUEST_AGENT_BINARY}" ]; then
    echo "⚠️  Guest agent binary not found at ./bin/${GUEST_AGENT_BINARY}"
    echo "   Building guest agent..."
    # Build guest agent
    cd cmd/rexec-guest-agent
    go build -o ../../bin/${GUEST_AGENT_BINARY} .
    cd ../..
fi

# Copy guest agent to rootfs
sudo cp "./bin/${GUEST_AGENT_BINARY}" "${GUEST_AGENT_PATH}"
sudo chmod +x "${GUEST_AGENT_PATH}"

# Create systemd service for guest agent
echo "🔧 Creating guest agent systemd service..."
sudo tee "${ROOTFS_DIR}/etc/systemd/system/rexec-guest-agent.service" > /dev/null <<EOF
[Unit]
Description=Rexec Guest Agent
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/rexec-guest-agent -port 1234
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Enable guest agent service
sudo chroot "${ROOTFS_DIR}" /bin/bash <<EOF
systemctl enable rexec-guest-agent.service
EOF

# Unmount filesystems
echo "🔓 Unmounting filesystems..."
sudo umount "${ROOTFS_DIR}/sys" || true
sudo umount "${ROOTFS_DIR}/proc" || true
sudo umount "${ROOTFS_DIR}/dev" || true

# Create ext4 image
echo "💾 Creating ext4 image (${IMAGE_SIZE}MB)..."
IMAGE_PATH="${OUTPUT_DIR}/${IMAGE_NAME}"
sudo rm -f "${IMAGE_PATH}"

# Create empty image file
sudo dd if=/dev/zero of="${IMAGE_PATH}" bs=1M count="${IMAGE_SIZE}" status=progress

# Format as ext4
sudo mkfs.ext4 -F "${IMAGE_PATH}"

# Mount image
TMP_MOUNT="/tmp/rexec-mount-$$"
mkdir -p "${TMP_MOUNT}"
sudo mount -o loop "${IMAGE_PATH}" "${TMP_MOUNT}"

# Copy rootfs to image
echo "📋 Copying rootfs to image..."
sudo cp -a "${ROOTFS_DIR}"/* "${TMP_MOUNT}/"

# Unmount image
sudo umount "${TMP_MOUNT}"
rmdir "${TMP_MOUNT}"

# Set ownership
sudo chown "${USER}:${USER}" "${IMAGE_PATH}"

# Clean up rootfs directory
echo "🧹 Cleaning up..."
sudo rm -rf "${ROOTFS_DIR}"

echo "✅ Rootfs image created: ${IMAGE_PATH}"
echo "   Size: $(du -h "${IMAGE_PATH}" | cut -f1)"
echo ""
echo "📝 Next steps:"
echo "   1. Test the image with Firecracker"
echo "   2. Verify guest agent starts automatically"
echo "   3. Test terminal connection"
