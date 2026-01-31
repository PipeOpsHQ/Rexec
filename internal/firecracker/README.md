# Firecracker Integration

This package provides Firecracker microVM support for Rexec.

## Prerequisites

1. **Firecracker binary**: Download from https://github.com/firecracker-microvm/firecracker/releases
   - Set `FIRECRACKER_BINARY_PATH` environment variable if not in PATH

2. **KVM support**: Requires `/dev/kvm` device
   - Check: `ls -l /dev/kvm`
   - Add user to kvm group: `sudo usermod -aG kvm $USER`

3. **Kernel image**: Linux kernel compiled for Firecracker
   - Set `FIRECRACKER_KERNEL_PATH` environment variable
   - Default: `/opt/firecracker/vmlinux.bin`
   - Download from: https://github.com/firecracker-microvm/firecracker/blob/main/resources/guest_configs/microvm-kernel-x86_64.config

4. **Rootfs images**: Pre-built rootfs images for each distro
   - Set `FIRECRACKER_ROOTFS_PATH` environment variable
   - Default: `/var/lib/rexec/firecracker/rootfs`
   - Images should be named: `{distro}.ext4` (e.g., `ubuntu.ext4`, `debian.ext4`)

## Environment Variables

- `FIRECRACKER_BINARY_PATH`: Path to firecracker binary (default: `firecracker`)
- `FIRECRACKER_KERNEL_PATH`: Path to kernel image (default: `/opt/firecracker/vmlinux.bin`)
- `FIRECRACKER_ROOTFS_PATH`: Base path for rootfs images (default: `/var/lib/rexec/firecracker/rootfs`)
- `FIRECRACKER_SOCKET_DIR`: Directory for Unix sockets (default: `/tmp/firecracker`)
- `FIRECRACKER_BRIDGE_NAME`: Bridge name for networking (default: `rexec-bridge`)

## Architecture

- **Manager**: Handles VM lifecycle and coordinates with network/storage
- **Client**: Communicates with Firecracker API via Unix socket
- **NetworkManager**: Manages tap devices and bridge networking
- **StorageManager**: Manages rootfs images and snapshots

## Usage

```go
// Create manager
mgr, err := firecracker.NewManager()
if err != nil {
    log.Fatal(err)
}

// Check availability
if !mgr.IsAvailable(ctx) {
    log.Fatal("Firecracker not available")
}

// Create VM
cfg := providers.CreateConfig{
    UserID:    "user-123",
    Name:      "my-vm",
    Image:     "ubuntu",
    MemoryMB:  512,
    CPUShares: 1000, // 1 CPU
    DiskMB:    2048,
}

terminal, err := mgr.Create(ctx, cfg)
if err != nil {
    log.Fatal(err)
}

// Start/Stop/Delete
mgr.Start(ctx, terminal.ID)
mgr.Stop(ctx, terminal.ID)
mgr.Delete(ctx, terminal.ID)
```

## Network Setup

The manager automatically:
1. Creates a bridge (`rexec-bridge`) if it doesn't exist
2. Creates a tap device for each VM
3. Connects tap device to bridge
4. Assigns IP via DHCP (requires DHCP server on bridge)

## Rootfs Images

Rootfs images should be:
- Ext4 filesystem images
- Pre-configured with guest agent (optional, for future use)
- Named: `{distro}.ext4` (e.g., `ubuntu.ext4`)

To create a rootfs image:
```bash
# Create empty image
dd if=/dev/zero of=ubuntu.ext4 bs=1M count=2048
mkfs.ext4 ubuntu.ext4

# Mount and install base system
sudo mount ubuntu.ext4 /mnt
sudo debootstrap ubuntu /mnt
# ... configure system ...
sudo umount /mnt
```

## Future Work

- [ ] Guest agent implementation (vsock communication)
- [ ] ZFS snapshot/clone support
- [ ] Terminal connection via guest agent
- [ ] Metrics collection
- [ ] File copy operations
