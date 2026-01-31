# Firecracker MicroVM Integration Plan

## Overview

This document outlines the plan to add **Firecracker microVM support as an additional provider option** to Rexec's terminal-as-a-service platform. This builds on top of the existing infrastructure and adds a new use case path for:

1. **Developers & Homelabs**: Users who want better isolation, faster boot times, and full Linux kernel support for their terminal environments
2. **Datacenter Providers**: Providers who want to offer VM-based terminals on bare-metal infrastructure with API-driven provisioning

**Important**: This is **NOT a migration**. Docker containers remain the default and primary option. Firecracker is an **optional, advanced provider** for specific use cases.

## Goals

1. **Add Firecracker as Provider Option**: Extend Rexec's terminal-as-a-service to support microVMs
   - Sub-second boot times (<1s) for faster environment provisioning
   - Better isolation than containers (full kernel, systemd)
   - ZFS snapshot/clone support for instant environment duplication
   - Full Linux OS support (not just containerized apps)

2. **Datacenter Provider Use Case**: Enable providers to offer VM-based terminals
   - API-driven VM provisioning on bare-metal hosts
   - Cost-effective bare-metal deployment (no cloud overhead)
   - Disposable Kubernetes clusters for testing/support
   - Customer support environments that match production

3. **Developer/Homelab Use Case**: Give developers more control and isolation
   - Full systemd support for services
   - Better security isolation for untrusted code
   - Faster environment cloning for testing
   - Native Linux experience (not containerized)

## Architecture

### Current State

```
[Browser] ←(WebSocket)→ [Rexec API] ←→ [PostgreSQL]
                              │
                              ├── [Container Manager] ──→ [Docker Engine] (Default)
                              │
                              └── [Agent Handler] ←(WebSocket)→ [Remote Agents]
```

### Target State (Additive)

```
[Browser] ←(WebSocket)→ [Rexec API] ←→ [PostgreSQL]
                              │
                              ├── [Container Manager] ──→ [Docker Engine] (Default)
                              │
                              ├── [Firecracker Manager] ──→ [Firecracker API] (Optional)
                              │                              ├── [VM Lifecycle]
                              │                              ├── [Guest Agent]
                              │                              └── [ZFS Storage]
                              │
                              └── [Agent Handler] ←(WebSocket)→ [Remote Agents]
```

**Key Point**: All three providers (Docker, Firecracker, Agents) coexist. Users/providers choose based on their needs.

## Components

### 1. Firecracker Manager (`internal/firecracker/manager.go`)

**Responsibilities:**
- VM lifecycle management (create, start, stop, delete)
- Firecracker API client wrapper
- Resource management (CPU, memory, disk)
- Network configuration (tap devices, bridges)
- Guest agent communication

**Key Interfaces:**
```go
type Manager interface {
    CreateVM(ctx context.Context, cfg VMConfig) (*VMInfo, error)
    StartVM(ctx context.Context, vmID string) error
    StopVM(ctx context.Context, vmID string) error
    DeleteVM(ctx context.Context, vmID string) error
    GetVM(ctx context.Context, vmID string) (*VMInfo, error)
    ListVMs(ctx context.Context, userID string) ([]*VMInfo, error)
    ExecInVM(ctx context.Context, vmID string, cmd []string) ([]byte, error)
    CopyToVM(ctx context.Context, vmID string, src, dst string) error
    CopyFromVM(ctx context.Context, vmID string, src, dst string) error
    GetVMStats(ctx context.Context, vmID string) (*VMStats, error)
    CreateSnapshot(ctx context.Context, vmID string, name string) error
    CloneFromSnapshot(ctx context.Context, snapshotName string, cfg VMConfig) (*VMInfo, error)
}
```

**Dependencies:**
- Firecracker Go SDK: `github.com/firecracker-microvm/firecracker-go-sdk`
- ZFS bindings: `github.com/bicomsystems/go-libzfs` or shell commands
- Network utilities: `github.com/vishvananda/netlink` for tap/bridge management

### 2. Guest Agent (`internal/firecracker/guest_agent.go`)

**Responsibilities:**
- Communication with VM guest via vsock
- Command execution
- File copy operations
- Metrics collection
- Port forwarding setup

**Protocol:**
- JSON-RPC over vsock
- Commands: `exec`, `cp`, `shell`, `metrics`, `port-forward`

**Guest-side Implementation:**
- Systemd service that runs in VM
- Listens on vsock socket
- Executes commands and returns results
- Reports metrics (CPU, memory, disk)

### 3. Storage Manager (`internal/firecracker/storage.go`)

**Responsibilities:**
- ZFS dataset management
- Snapshot creation/management
- Clone creation from snapshots
- Disk image management (rootfs, data volumes)

**ZFS Structure:**
```
zpool/rexec/
  ├── vms/
  │   ├── {userID}/
  │   │   ├── {vmID}/
  │   │   │   ├── rootfs@snapshot-{timestamp}
  │   │   │   └── data@snapshot-{timestamp}
  │   │   └── templates/
  │   │       ├── ubuntu-24.04@base
  │   │       └── debian-12@base
  └── snapshots/
      └── {snapshot-name}/
```

### 4. Network Manager (`internal/firecracker/network.go`)

**Responsibilities:**
- Tap device creation/management
- Bridge configuration
- IP address allocation
- Network isolation (similar to Docker's isolated network)

**Network Topology:**
```
[VM] ←→ [tap-{vmID}] ←→ [rexec-bridge] ←→ [Host Network]
```

### 5. VM Handler (`internal/api/handlers/vm.go`)

**Responsibilities:**
- HTTP API endpoints for VM management
- Integration with existing container handler patterns
- Progress events via WebSocket
- Resource validation

**Endpoints:**
- `POST /api/v1/vms` - Create VM
- `GET /api/v1/vms` - List VMs
- `GET /api/v1/vms/:id` - Get VM details
- `DELETE /api/v1/vms/:id` - Delete VM
- `POST /api/v1/vms/:id/start` - Start VM
- `POST /api/v1/vms/:id/stop` - Stop VM
- `POST /api/v1/vms/:id/snapshot` - Create snapshot
- `POST /api/v1/vms/:id/clone` - Clone from snapshot
- `POST /api/v1/vms/:id/exec` - Execute command
- `POST /api/v1/vms/:id/copy` - Copy file
- `GET /api/v1/vms/:id/stats` - Get VM stats

### 6. Terminal Integration (`internal/api/handlers/terminal.go`)

**Modifications:**
- Support VM-based terminals (in addition to containers)
- Guest agent communication for shell access
- vsock-based terminal streaming

**Terminal ID Format:**
- Containers: `{dockerID}`
- VMs: `vm:{vmID}`
- Agents: `agent:{agentID}`

## Implementation Phases

### Phase 1: Core Infrastructure (Weeks 1-2)

**Tasks:**
1. Set up Firecracker development environment
   - Install Firecracker binary
   - Configure kernel and rootfs images
   - Test basic VM creation

2. Create Firecracker Manager
   - Basic VM lifecycle (create, start, stop, delete)
   - Resource limits (CPU, memory)
   - Network setup (tap devices, bridges)

3. Database schema updates
   - Add `vms` table (similar to containers)
   - Add `vm_snapshots` table
   - Migration scripts

**Deliverables:**
- `internal/firecracker/manager.go` (basic implementation)
- `internal/firecracker/network.go` (tap/bridge management)
- Database migrations
- Unit tests for manager

### Phase 2: Guest Agent & Communication (Weeks 3-4)

**Tasks:**
1. Implement guest agent protocol
   - vsock communication
   - JSON-RPC message format
   - Command execution

2. Build guest agent systemd service
   - Agent binary/service
   - vsock listener
   - Command executor

3. Integrate guest agent with manager
   - Agent discovery
   - Health checks
   - Command execution via agent

**Deliverables:**
- `internal/firecracker/guest_agent.go`
- `cmd/rexec-guest-agent/` (guest agent binary)
- Guest agent systemd service file
- Integration tests

### Phase 3: Storage & Snapshots (Weeks 5-6)

**Tasks:**
1. Implement ZFS storage manager
   - Dataset creation
   - Snapshot operations
   - Clone operations

2. Rootfs image management
   - Pre-built rootfs images (Ubuntu, Debian, etc.)
   - Image preparation scripts
   - Template management

3. Snapshot/clone API
   - Create snapshot endpoint
   - Clone from snapshot
   - Snapshot listing/deletion

**Deliverables:**
- `internal/firecracker/storage.go`
- Rootfs image build scripts
- Snapshot management API
- Integration tests

### Phase 4: API Integration (Weeks 7-8)

**Tasks:**
1. Create VM handler
   - HTTP endpoints
   - Request validation
   - Error handling

2. Integrate with existing systems
   - Container handler patterns
   - WebSocket events
   - Database storage

3. Terminal integration
   - VM terminal support
   - Guest agent terminal streaming
   - Terminal ID resolution

**Deliverables:**
- `internal/api/handlers/vm.go`
- Terminal integration updates
- API documentation
- E2E tests

### Phase 5: UI Integration (Weeks 9-10)

**Tasks:**
1. Frontend VM management
   - VM list view
   - VM creation form
   - VM details page

2. Terminal UI updates
   - VM terminal support
   - Snapshot/clone UI
   - Resource monitoring

3. Provider selection
   - Docker vs Firecracker toggle
   - Provider-specific options

**Deliverables:**
- Frontend VM components
- Terminal UI updates
- Provider selection UI

### Phase 6: Datacenter Provider Support (Weeks 11-12)

**Tasks:**
1. Provider abstraction
   - Provider interface
   - Provider registry
   - Provider selection logic

2. Datacenter provider implementation
   - Bare-metal host management
   - Multi-host support
   - Resource pooling

3. Kubernetes integration
   - K3s cluster creation
   - Node provisioning
   - Cluster management

**Deliverables:**
- `internal/providers/` package
- Datacenter provider implementation
- Kubernetes integration
- Documentation

## Technical Details

### VM Configuration

```go
type VMConfig struct {
    UserID        string
    VMName        string
    ImageType     string // ubuntu, debian, etc.
    RootfsPath    string // Path to rootfs image
    KernelPath    string // Path to kernel image
    MemoryMB      int64
    VCPUs         int64
    DiskSizeGB    int64
    NetworkConfig NetworkConfig
    UserData      string // Cloud-init userdata script
    Labels        map[string]string
}
```

### Guest Agent Protocol

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "exec",
  "params": {
    "command": ["ls", "-la"],
    "timeout": 30
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "exit_code": 0,
    "stdout": "...",
    "stderr": "",
    "duration_ms": 45
  }
}
```

### Rootfs Images

**Requirements:**
- Pre-built rootfs images for each distro
- Guest agent pre-installed
- SSH server configured
- Cloud-init support
- Minimal size (< 100MB compressed)

**Build Process:**
1. Use `debootstrap` or similar for base system
2. Install guest agent systemd service
3. Configure networking (DHCP)
4. Create rootfs tarball
5. Compress and store in ZFS dataset

### Network Configuration

**Isolated Network:**
- Bridge: `rexec-bridge`
- Subnet: `10.0.0.0/16`
- DHCP: `dnsmasq` or `systemd-networkd`
- Firewall: `iptables` rules for isolation

**VM Network:**
- Tap device: `tap-{vmID}`
- MAC address: Auto-generated
- IP: DHCP-assigned from isolated subnet

### Resource Limits

**Default Limits:**
- Memory: 512MB (guest), 2GB (pro), 4GB (enterprise)
- VCPUs: 1 (guest), 2 (pro), 4 (enterprise)
- Disk: 5GB (guest), 20GB (pro), 50GB (enterprise)

**Enforcement:**
- Firecracker API resource limits
- Cgroup v2 for host-level limits
- ZFS quota for disk limits

## Database Schema

### VMs Table

```sql
CREATE TABLE vms (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    image_type VARCHAR(50) NOT NULL,
    rootfs_path TEXT NOT NULL,
    kernel_path TEXT NOT NULL,
    memory_mb INTEGER NOT NULL,
    vcpus INTEGER NOT NULL,
    disk_gb INTEGER NOT NULL,
    ip_address INET,
    status VARCHAR(50) NOT NULL, -- creating, running, stopped, error
    vm_id VARCHAR(255) UNIQUE, -- Firecracker VM ID
    labels JSONB,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_vms_user_id ON vms(user_id);
CREATE INDEX idx_vms_status ON vms(status);
CREATE INDEX idx_vms_vm_id ON vms(vm_id);
```

### VM Snapshots Table

```sql
CREATE TABLE vm_snapshots (
    id UUID PRIMARY KEY,
    vm_id UUID NOT NULL REFERENCES vms(id),
    name VARCHAR(255) NOT NULL,
    snapshot_path TEXT NOT NULL, -- ZFS snapshot path
    memory_snapshot_path TEXT, -- Optional memory snapshot
    created_at TIMESTAMP NOT NULL,
    UNIQUE(vm_id, name)
);

CREATE INDEX idx_vm_snapshots_vm_id ON vm_snapshots(vm_id);
```

## Security Considerations

1. **VM Isolation**
   - Each VM runs in separate cgroup
   - Network isolation via bridge firewall rules
   - No shared filesystem (except read-only rootfs)

2. **Guest Agent Security**
   - vsock communication (no network exposure)
   - Command execution limits (timeout, resource limits)
   - User permissions (non-root execution)

3. **Resource Limits**
   - Firecracker API limits
   - Host-level cgroup limits
   - ZFS disk quotas

4. **Network Security**
   - Isolated bridge network
   - Firewall rules preventing inter-VM communication
   - No direct host network access (unless configured)

## Performance Targets

- **Boot Time**: < 1 second (cold start)
- **Clone Time**: < 100ms (from snapshot)
- **Memory Overhead**: < 5MB per VM (idle)
- **CPU Overhead**: < 1% per VM (idle)

## Provider Selection Strategy

### User Choice Model

Users/providers can choose the provider that fits their use case:

1. **Docker (Default)**: 
   - Fast, lightweight, familiar
   - Best for: General development, quick sandboxes
   - No special requirements

2. **Firecracker (Optional)**:
   - Better isolation, faster boot, full OS
   - Best for: Security-sensitive workloads, homelabs, datacenter providers
   - Requires: KVM support, Firecracker binary

3. **Agent (BYOS)**:
   - Connect existing machines
   - Best for: Accessing existing infrastructure
   - No special requirements

### Provider Selection UI

- **Sandbox Creation**: Dropdown to select provider (Docker/Firecracker)
- **Provider Settings**: Configure default provider per user/org
- **Feature Detection**: Show Firecracker option only if available
- **Clear Labeling**: Explain use cases for each provider

### Datacenter Provider Configuration

Providers can configure which providers to offer:
- Enable/disable Firecracker support
- Set resource limits per provider
- Configure provider-specific features (Kubernetes, etc.)

## Testing Strategy

1. **Unit Tests**
   - Manager operations
   - Guest agent protocol
   - Storage operations

2. **Integration Tests**
   - VM lifecycle
   - Guest agent communication
   - Snapshot/clone operations

3. **E2E Tests**
   - Full VM creation flow
   - Terminal access
   - Resource limits

4. **Performance Tests**
   - Boot time benchmarks
   - Clone performance
   - Resource overhead

## Dependencies

### Go Packages

```go
github.com/firecracker-microvm/firecracker-go-sdk v1.2.0
github.com/vishvananda/netlink v1.2.0
github.com/bicomsystems/go-libzfs v0.3.0 // or use shell commands
```

### System Requirements

- Linux kernel 4.14+ (for KVM)
- KVM enabled
- ZFS support (optional, can use regular filesystem)
- Firecracker binary (v1.2.0+)
- Rootfs images (pre-built)

## Use Case Scenarios

### Scenario 1: Developer Homelab
**User**: Developer running Rexec on home server
**Need**: Better isolation for running untrusted code, faster environment cloning
**Solution**: Choose Firecracker provider when creating sandbox
**Benefits**: 
- Full kernel isolation (safer than containers)
- Instant cloning for testing different scenarios
- Systemd support for running services

### Scenario 2: Datacenter Provider
**User**: Provider offering terminal-as-a-service to customers
**Need**: Cost-effective bare-metal deployment, API-driven provisioning
**Solution**: Configure Firecracker as available provider
**Benefits**:
- Lower costs (bare-metal vs cloud)
- Faster provisioning (<1s boot)
- Better resource utilization
- Kubernetes cluster provisioning for customer support

### Scenario 3: Security-Sensitive Workloads
**User**: Developer working with untrusted code
**Need**: Maximum isolation between environments
**Solution**: Use Firecracker provider for critical sandboxes
**Benefits**:
- Full kernel isolation (not shared with host)
- No container escape vulnerabilities
- Isolated network stack

## Open Questions

1. **Storage Backend**: ZFS vs regular filesystem?
   - ZFS: Better snapshots, but requires ZFS support
   - Regular FS: More compatible, but slower snapshots
   - **Recommendation**: Support both, prefer ZFS if available

2. **Rootfs Management**: Pre-built vs on-demand?
   - Pre-built: Faster, but requires storage
   - On-demand: More flexible, but slower first boot
   - **Recommendation**: Pre-built templates, on-demand customization

3. **Network Model**: Bridge vs macvtap?
   - Bridge: More compatible, easier management
   - Macvtap: Better performance, more complex
   - **Recommendation**: Bridge for compatibility, macvtap as opt-in

4. **Guest Agent**: vsock vs network?
   - vsock: More secure, requires kernel support
   - Network: More compatible, less secure
   - **Recommendation**: vsock preferred, network fallback

## Success Metrics

1. **Functionality**
   - ✅ VM creation/management works alongside Docker containers
   - ✅ Terminal access works identically to containers
   - ✅ Snapshots/clones work for instant environment duplication
   - ✅ Guest agent communication works reliably
   - ✅ Provider selection works seamlessly

2. **Performance**
   - ✅ Boot time < 1s (vs Docker's ~2-5s)
   - ✅ Clone time < 100ms (vs Docker's ~1-2s)
   - ✅ Memory overhead < 5MB/VM (vs Docker's ~10-20MB)

3. **Reliability**
   - ✅ 99.9% VM creation success rate
   - ✅ No memory leaks
   - ✅ Proper cleanup on errors
   - ✅ Graceful fallback to Docker if Firecracker unavailable

4. **Adoption**
   - ✅ Users can choose provider without friction
   - ✅ Clear use case documentation
   - ✅ Datacenter providers can enable/configure easily

## Timeline

- **Total Duration**: 12 weeks
- **Team Size**: 2-3 engineers
- **Milestones**: 6 phases (2 weeks each)

## Integration with Existing Terminal-as-a-Service

### Unified Terminal Interface

All providers (Docker, Firecracker, Agents) expose the same terminal interface:
- Same WebSocket protocol
- Same terminal ID format (with provider prefix: `vm:{id}`, `container:{id}`, `agent:{id}`)
- Same API endpoints (provider abstraction layer)

### Provider Abstraction Layer

```go
type TerminalProvider interface {
    Create(ctx context.Context, cfg CreateConfig) (*TerminalInfo, error)
    Start(ctx context.Context, id string) error
    Stop(ctx context.Context, id string) error
    Delete(ctx context.Context, id string) error
    ConnectTerminal(ctx context.Context, id string) (*TerminalConnection, error)
    Exec(ctx context.Context, id string, cmd []string) ([]byte, error)
    GetStats(ctx context.Context, id string) (*ResourceStats, error)
}

// Registry of providers
type ProviderRegistry struct {
    docker      *container.Manager
    firecracker *firecracker.Manager
    agent       *AgentHandler
}
```

### Database Schema Updates

Extend existing `containers` table to support VMs (or create unified `terminals` table):

```sql
-- Option 1: Extend containers table
ALTER TABLE containers ADD COLUMN provider VARCHAR(50) DEFAULT 'docker';
ALTER TABLE containers ADD COLUMN vm_id VARCHAR(255);
ALTER TABLE containers ADD COLUMN provider_config JSONB;

-- Option 2: Unified terminals table (better long-term)
CREATE TABLE terminals (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL, -- 'docker', 'firecracker', 'agent'
    provider_id VARCHAR(255) NOT NULL, -- Docker ID, VM ID, or Agent ID
    provider_config JSONB,
    -- ... rest of fields
);
```

## Next Steps

1. ✅ Review and approve plan
2. Set up development environment (Firecracker, KVM, ZFS)
3. Create provider abstraction layer
4. Implement Firecracker Manager (Phase 1)
5. Integrate with existing terminal handler
6. Add provider selection UI
7. Document use cases and when to choose each provider
