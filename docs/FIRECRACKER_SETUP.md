# Firecracker Setup Guide

This guide explains how to set up Firecracker support for Rexec.

## Prerequisites

### 1. System Requirements

- **Linux kernel 4.14+** with KVM support
- **KVM enabled** and accessible
- **Root/sudo access** for network setup
- **Firecracker binary** (v1.2.0+)

### 2. Install Firecracker

```bash
# Download latest release
VERSION=1.2.0
ARCH=$(uname -m)
wget https://github.com/firecracker-microvm/firecracker/releases/download/v${VERSION}/firecracker-v${VERSION}-${ARCH}.tgz
tar -xzf firecracker-v${VERSION}-${ARCH}.tgz
sudo mv release-${VERSION}-${ARCH}/firecracker-${VERSION}-${ARCH} /usr/local/bin/firecracker
sudo chmod +x /usr/local/bin/firecracker

# Or set FIRECRACKER_BINARY_PATH environment variable
export FIRECRACKER_BINARY_PATH=/path/to/firecracker
```

### 3. Enable KVM Access

```bash
# Check if KVM is available
ls -l /dev/kvm

# Add user to kvm group
sudo usermod -aG kvm $USER

# Log out and back in for group changes to take effect
```

## Kernel Preparation

Firecracker requires a custom-compiled Linux kernel. You have two options:

### Option 1: Build from Source (Recommended)

See the [Firecracker documentation](https://github.com/firecracker-microvm/firecracker/blob/main/docs/rootfs-and-kernels-setup.md) for detailed instructions.

Quick steps:
```bash
# Clone Linux kernel
git clone https://github.com/torvalds/linux.git
cd linux

# Use Firecracker's kernel config
wget https://raw.githubusercontent.com/firecracker-microvm/firecracker/main/resources/guest_configs/microvm-kernel-x86_64.config
mv microvm-kernel-x86_64.config .config

# Build kernel
make -j$(nproc) vmlinux

# Copy to Firecracker directory
sudo mkdir -p /opt/firecracker
sudo cp vmlinux /opt/firecracker/vmlinux.bin
```

### Option 2: Extract from System Kernel

```bash
# Install extract-vmlinux script
wget https://raw.githubusercontent.com/torvalds/linux/master/scripts/extract-vmlinux
chmod +x extract-vmlinux

# Extract from system kernel
sudo mkdir -p /opt/firecracker
sudo ./extract-vmlinux /boot/vmlinuz-$(uname -r) > /opt/firecracker/vmlinux.bin
```

### Option 3: Use Helper Script

```bash
./scripts/prepare-kernel.sh
```

Set kernel path:
```bash
export FIRECRACKER_KERNEL_PATH=/opt/firecracker/vmlinux.bin
```

## Rootfs Image Preparation

### Build Rootfs Images

Use the provided script to build rootfs images:

```bash
# Build Ubuntu 24.04 image
./scripts/build-rootfs.sh ubuntu 24.04

# Build Debian 12 image
./scripts/build-rootfs.sh debian 12

# Build with custom size (default: 2048MB)
./scripts/build-rootfs.sh ubuntu 24.04 4096
```

The script will:
1. Bootstrap the base system using `debootstrap`
2. Install essential packages (systemd, SSH, etc.)
3. Create `rexec` user with sudo access
4. Install and configure guest agent
5. Create ext4 image file

### Image Location

Images are stored at:
```
/var/lib/rexec/firecracker/rootfs/images/
├── ubuntu-24.04.ext4
├── ubuntu-22.04.ext4
├── debian-12.ext4
└── debian-11.ext4
```

Set custom path:
```bash
export FIRECRACKER_ROOTFS_PATH=/path/to/rootfs
```

### Manual Rootfs Creation

If you prefer to build manually:

```bash
# Create empty image
dd if=/dev/zero of=ubuntu.ext4 bs=1M count=2048
mkfs.ext4 ubuntu.ext4

# Mount and bootstrap
sudo mount ubuntu.ext4 /mnt
sudo debootstrap --arch=amd64 ubuntu /mnt http://archive.ubuntu.com/ubuntu/

# Configure system (see build-rootfs.sh for details)
# ...

# Unmount
sudo umount /mnt
```

## Guest Agent Setup

The guest agent is automatically installed by `build-rootfs.sh`. If building manually:

1. **Build guest agent binary:**
   ```bash
   cd cmd/rexec-guest-agent
   go build -o rexec-guest-agent .
   ```

2. **Copy to rootfs:**
   ```bash
   sudo cp rexec-guest-agent /mnt/usr/local/bin/
   sudo chmod +x /mnt/usr/local/bin/rexec-guest-agent
   ```

3. **Create systemd service:**
   ```bash
   sudo tee /mnt/etc/systemd/system/rexec-guest-agent.service > /dev/null <<EOF
   [Unit]
   Description=Rexec Guest Agent
   After=network.target

   [Service]
   Type=simple
   ExecStart=/usr/local/bin/rexec-guest-agent -port 1234
   Restart=always

   [Install]
   WantedBy=multi-user.target
   EOF
   ```

4. **Enable service:**
   ```bash
   sudo chroot /mnt systemctl enable rexec-guest-agent.service
   ```

## Network Setup

The Firecracker manager automatically creates:
- Bridge: `rexec-bridge` (configurable via `FIRECRACKER_BRIDGE_NAME`)
- Tap devices: `tap-{vm-id}` for each VM

### Manual Network Setup

If you need to set up networking manually:

```bash
# Create bridge
sudo ip link add name rexec-bridge type bridge
sudo ip link set rexec-bridge up

# Configure DHCP (optional, using dnsmasq)
sudo apt-get install dnsmasq
sudo tee /etc/dnsmasq.d/rexec.conf > /dev/null <<EOF
interface=rexec-bridge
dhcp-range=10.0.0.2,10.0.0.254,255.255.255.0,12h
EOF
sudo systemctl restart dnsmasq
```

## Environment Variables

Configure Firecracker via environment variables:

```bash
# Firecracker binary path
export FIRECRACKER_BINARY_PATH=/usr/local/bin/firecracker

# Kernel image path
export FIRECRACKER_KERNEL_PATH=/opt/firecracker/vmlinux.bin

# Rootfs base path
export FIRECRACKER_ROOTFS_PATH=/var/lib/rexec/firecracker/rootfs

# Socket directory
export FIRECRACKER_SOCKET_DIR=/tmp/firecracker

# Bridge name
export FIRECRACKER_BRIDGE_NAME=rexec-bridge
```

## Testing

### 1. Test Firecracker Availability

```bash
# Check if Firecracker is available
firecracker --version

# Check if KVM is accessible
ls -l /dev/kvm
```

### 2. Test VM Creation

```bash
# Start Rexec server
make run

# Create a VM via API
curl -X POST http://localhost:8080/api/vms \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-vm",
    "image": "ubuntu",
    "provider": "firecracker",
    "memory_mb": 512,
    "cpu_shares": 1000
  }'
```

### 3. Test Terminal Connection

```bash
# Connect via WebSocket
wscat -c "ws://localhost:8080/ws/terminal/vm:vm-USER-test-vm?token=YOUR_TOKEN"
```

## Troubleshooting

### VM Creation Fails

1. **Check KVM access:**
   ```bash
   ls -l /dev/kvm
   groups  # Should include 'kvm'
   ```

2. **Check Firecracker binary:**
   ```bash
   which firecracker
   firecracker --version
   ```

3. **Check kernel and rootfs:**
   ```bash
   ls -lh /opt/firecracker/vmlinux.bin
   ls -lh /var/lib/rexec/firecracker/rootfs/images/
   ```

### Terminal Connection Fails

1. **Check guest agent is running:**
   ```bash
   # Inside VM (if you can access it)
   systemctl status rexec-guest-agent
   ```

2. **Check network connectivity:**
   ```bash
   # Check bridge
   ip link show rexec-bridge
   
   # Check tap device
   ip link show tap-{vm-id}
   ```

3. **Check Firecracker logs:**
   ```bash
   journalctl -u firecracker
   ```

### Network Issues

1. **Bridge not created:**
   ```bash
   sudo ip link add name rexec-bridge type bridge
   sudo ip link set rexec-bridge up
   ```

2. **Tap device issues:**
   ```bash
   # List tap devices
   ip link show type tap
   
   # Remove stuck tap device
   sudo ip link delete tap-{name}
   ```

## Production Considerations

1. **Security:**
   - Run Firecracker as non-root user (requires proper permissions)
   - Isolate network bridge from host network
   - Use firewall rules to prevent inter-VM communication

2. **Performance:**
   - Use local NVMe/SSD for rootfs images
   - Enable KVM hardware acceleration
   - Tune kernel parameters for microVMs

3. **Monitoring:**
   - Monitor Firecracker process health
   - Track VM resource usage
   - Set up alerts for failed VM creations

4. **Backup:**
   - Backup rootfs images regularly
   - Consider ZFS snapshots for instant cloning
   - Document custom rootfs modifications

## Next Steps

- [ ] Build rootfs images for your distros
- [ ] Test VM creation and terminal access
- [ ] Configure production environment variables
- [ ] Set up monitoring and alerts
- [ ] Document custom configurations
