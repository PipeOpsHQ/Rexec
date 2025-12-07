# Rexec Security Hardening

This document outlines the security measures implemented in Rexec containers.

## Container Security Configuration

### Capability Management

Containers run with minimal Linux capabilities:

```
CapDrop: ALL
CapAdd:
  - CHOWN           # Change file ownership
  - DAC_OVERRIDE    # Bypass file permission checks (needed for sudo)
  - FOWNER          # Bypass permission checks on file owner
  - SETGID          # Set group ID
  - SETUID          # Set user ID (needed for su/sudo)
  - KILL            # Send signals
  - NET_BIND_SERVICE # Bind to ports < 1024
```

### Privilege Escalation Prevention

- `no-new-privileges:true` - Prevents processes from gaining additional privileges
- `Privileged: false` - Container cannot access host devices
- Default Docker seccomp profile blocks dangerous syscalls

### Filesystem Security

**Read-only Root Filesystem:**
- `ReadonlyRootfs: true` - Root filesystem is mounted read-only
- Writable directories use tmpfs mounts with size limits

**Tmpfs Mounts:**
| Path | Options | Size Limit |
|------|---------|------------|
| /tmp | rw,noexec,nosuid | 100MB |
| /var/tmp | rw,noexec,nosuid | 50MB |
| /run | rw,noexec,nosuid | 50MB |
| /var/run | rw,noexec,nosuid | 50MB |
| /root | rw,nosuid | 50MB |

**User Data:**
- `/home/user` uses a persistent Docker volume (not tmpfs)
- Volume is isolated per-container

### Masked Paths

Sensitive host information is hidden from containers:

- `/proc/acpi` - ACPI information
- `/proc/asound` - Sound devices
- `/proc/kcore` - Kernel memory
- `/proc/keys` - Kernel keyrings
- `/proc/latency_stats` - Latency statistics
- `/proc/timer_list` - Timer information
- `/proc/timer_stats` - Timer statistics
- `/proc/sched_debug` - Scheduler debug info
- `/proc/scsi` - SCSI devices
- `/sys/firmware` - Firmware information
- `/sys/devices/virtual/powercap` - Power capping
- `/sys/kernel` - Kernel parameters

### Read-only Paths

Critical paths are mounted read-only:

- `/proc/bus` - Bus information
- `/proc/fs` - Filesystem info
- `/proc/irq` - Interrupt info
- `/proc/sys` - Kernel parameters
- `/proc/sysrq-trigger` - Magic SysRq

### Resource Limits

- **Memory**: Configurable per-tier (512MB - 4GB)
- **CPU**: Quota-based limiting
- **PIDs**: Maximum 256 processes (fork bomb protection)
- **Disk**: Quota support on XFS with pquota

### Network Isolation

- Containers run in isolated Docker network (`rexec-isolated`)
- Inter-container communication is restricted
- Controlled sysctl: `net.ipv4.ip_unprivileged_port_start=0`

## Security Verification

Run these commands inside a container to verify security settings:

```bash
# Check capabilities
cat /proc/self/status | grep Cap

# Check user context
id
whoami

# Check mount restrictions
mount | grep -E "(proc|sys|dev)"

# Check network isolation
ip addr show

# Check resource limits
ulimit -a
```

## Security Scoring

| Configuration | Score | Description |
|---------------|-------|-------------|
| Basic Docker | 6/10 | Default Docker isolation |
| Current Rexec | 8/10 | Enhanced container security |
| With gVisor | 9/10 | Kernel-level isolation |

## Future Improvements

1. **User Namespace Remapping** - Map container root to unprivileged host user
2. **Custom Seccomp Profile** - More restrictive syscall filtering
3. **AppArmor/SELinux** - Mandatory access control profiles
4. **Network Policies** - Fine-grained egress controls

## References

- [Docker Security Best Practices](https://docs.docker.com/engine/security/)
- [CIS Docker Benchmark](https://www.cisecurity.org/benchmark/docker)
- [OWASP Docker Security](https://owasp.org/www-project-docker-top-10/)
- [Container Security Guide](https://snyk.io/blog/10-docker-image-security-best-practices/)
