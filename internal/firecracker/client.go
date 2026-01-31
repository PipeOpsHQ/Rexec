package firecracker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// FirecrackerClient handles communication with Firecracker API
type FirecrackerClient struct {
	socketPath string
	httpClient *http.Client
	process    *exec.Cmd
}

// NewFirecrackerClient creates a new Firecracker client
func NewFirecrackerClient(socketPath string) *FirecrackerClient {
	// Create HTTP client with Unix socket transport
	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}

	return &FirecrackerClient{
		socketPath: socketPath,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
}

// StartFirecrackerProcess starts the Firecracker process
func StartFirecrackerProcess(ctx context.Context, socketPath, kernelPath, rootfsPath string, config *VMConfig) (*FirecrackerClient, error) {
	// Find Firecracker binary
	firecrackerPath := os.Getenv("FIRECRACKER_BINARY_PATH")
	if firecrackerPath == "" {
		firecrackerPath = "firecracker"
	}

	// Check if binary exists
	if _, err := exec.LookPath(firecrackerPath); err != nil {
		return nil, fmt.Errorf("firecracker binary not found: %w", err)
	}

	// Ensure socket directory exists
	if err := os.MkdirAll(filepath.Dir(socketPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Remove existing socket if it exists
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			return nil, fmt.Errorf("failed to remove existing socket: %w", err)
		}
	}

	// Start Firecracker process
	cmd := exec.CommandContext(ctx, firecrackerPath,
		"--api-sock", socketPath,
		"--id", config.VMName,
	)

	// Set up logging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start firecracker: %w", err)
	}

	// Wait for socket to be created (with timeout)
	timeout := 5 * time.Second
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(socketPath); err == nil {
			// Socket exists, try to connect
			conn, err := net.Dial("unix", socketPath)
			if err == nil {
				conn.Close()
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Verify socket is accessible
	if _, err := os.Stat(socketPath); err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("firecracker socket not created after %v: %w", timeout, err)
	}

	client := NewFirecrackerClient(socketPath)
	client.process = cmd

	// Configure the VM
	if err := client.ConfigureVM(ctx, kernelPath, rootfsPath, config); err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to configure VM: %w", err)
	}

	return client, nil
}

// VMConfig holds configuration for a Firecracker VM
type VMConfig struct {
	VMName     string
	VCPUs      int64
	MemoryMB   int64
	KernelArgs string
	RootfsPath string
	KernelPath string
	Network    NetworkConfig
}

// NetworkConfig holds network configuration
type NetworkConfig struct {
	IfaceID      string
	GuestMAC     string
	HostDevName  string
	AllowMMDS    bool
	RxRateLimiter *RateLimiter
	TxRateLimiter *RateLimiter
}

// RateLimiter configuration
type RateLimiter struct {
	Bandwidth Bandwidth `json:"bandwidth"`
	Ops       Ops       `json:"ops"`
}

// Bandwidth configuration
type Bandwidth struct {
	Size  int64 `json:"size"`  // bytes per second
	OneTimeBurst int64 `json:"one_time_burst,omitempty"`
	RefillTime int64 `json:"refill_time"` // milliseconds
}

// Ops configuration
type Ops struct {
	Size  int64 `json:"size"`  // operations per second
	OneTimeBurst int64 `json:"one_time_burst,omitempty"`
	RefillTime int64 `json:"refill_time"` // milliseconds
}

// ConfigureVM configures the VM via Firecracker API
func (c *FirecrackerClient) ConfigureVM(ctx context.Context, kernelPath, rootfsPath string, config *VMConfig) error {
	// 1. Set boot source (kernel)
	bootSource := map[string]interface{}{
		"kernel_image_path": kernelPath,
		"boot_args":         config.KernelArgs,
	}

	// Only include initrd_path if it's not empty
	if config.KernelArgs != "" {
		bootSource["boot_args"] = config.KernelArgs
	}

	if err := c.put(ctx, "/boot-source", bootSource); err != nil {
		return fmt.Errorf("failed to set boot source: %w", err)
	}

	// 2. Set drives (rootfs) - Firecracker uses PATCH for drive updates
	driveConfig := map[string]interface{}{
		"drive_id":      "rootfs",
		"path_on_host":  rootfsPath,
		"is_root_device": true,
		"is_read_only":   false,
	}

	if err := c.patch(ctx, "/drives/rootfs", driveConfig); err != nil {
		return fmt.Errorf("failed to set rootfs drive: %w", err)
	}

	// 3. Set machine configuration (vCPUs and memory) - Firecracker uses PATCH
	machineConfig := map[string]interface{}{
		"vcpu_count":  config.VCPUs,
		"mem_size_mib": config.MemoryMB,
		"smt":         false, // Disable SMT for security
	}

	if err := c.patch(ctx, "/machine-config", machineConfig); err != nil {
		return fmt.Errorf("failed to set machine config: %w", err)
	}

	// 4. Set network interface (only if host device is specified)
	if config.Network.HostDevName != "" {
		netIface := map[string]interface{}{
			"iface_id":      config.Network.IfaceID,
			"guest_mac":     config.Network.GuestMAC,
			"host_dev_name": config.Network.HostDevName,
			"allow_mmds":    config.Network.AllowMMDS,
		}

		// Add rate limiters if configured
		if config.Network.RxRateLimiter != nil {
			netIface["rx_rate_limiter"] = config.Network.RxRateLimiter
		}
		if config.Network.TxRateLimiter != nil {
			netIface["tx_rate_limiter"] = config.Network.TxRateLimiter
		}

		// Firecracker uses PUT for network interface creation
		if err := c.put(ctx, fmt.Sprintf("/network-interfaces/%s", config.Network.IfaceID), netIface); err != nil {
			return fmt.Errorf("failed to set network interface: %w", err)
		}
	}

	return nil
}

// StartVM starts the VM
func (c *FirecrackerClient) StartVM(ctx context.Context) error {
	return c.put(ctx, "/actions", map[string]interface{}{
		"action_type": "InstanceStart",
	})
}

// StopVM sends a shutdown action
func (c *FirecrackerClient) StopVM(ctx context.Context) error {
	return c.put(ctx, "/actions", map[string]interface{}{
		"action_type": "SendCtrlAltDel",
	})
}

// GetVMInfo retrieves VM information
func (c *FirecrackerClient) GetVMInfo(ctx context.Context) (map[string]interface{}, error) {
	return c.get(ctx, "/vm")
}

// GetMachineConfig retrieves machine configuration
func (c *FirecrackerClient) GetMachineConfig(ctx context.Context) (map[string]interface{}, error) {
	return c.get(ctx, "/machine-config")
}

// GetInstanceInfo retrieves instance information
func (c *FirecrackerClient) GetInstanceInfo(ctx context.Context) (map[string]interface{}, error) {
	return c.get(ctx, "/vm")
}

// Shutdown shuts down the VM gracefully
func (c *FirecrackerClient) Shutdown(ctx context.Context) error {
	return c.put(ctx, "/actions", map[string]interface{}{
		"action_type": "FlushMetrics",
	})
}

// Close stops the Firecracker process
func (c *FirecrackerClient) Close() error {
	if c.process != nil {
		// Send shutdown signal
		if err := c.process.Process.Signal(os.Interrupt); err != nil {
			log.Printf("[Firecracker] Failed to send interrupt signal: %v", err)
		}

		// Wait for process to exit (with timeout)
		done := make(chan error, 1)
		go func() {
			done <- c.process.Wait()
		}()

		select {
		case <-time.After(5 * time.Second):
			// Force kill if still running
			if err := c.process.Process.Kill(); err != nil {
				log.Printf("[Firecracker] Failed to kill process: %v", err)
			}
		case err := <-done:
			if err != nil {
				log.Printf("[Firecracker] Process exited with error: %v", err)
			}
		}

		// Clean up socket
		if _, err := os.Stat(c.socketPath); err == nil {
			os.Remove(c.socketPath)
		}
	}
	return nil
}

// HTTP helper methods

func (c *FirecrackerClient) put(ctx context.Context, path string, data interface{}) error {
	url := "http://localhost" + path
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("firecracker API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// patch sends a PATCH request (used for some Firecracker operations)
func (c *FirecrackerClient) patch(ctx context.Context, path string, data interface{}) error {
	url := "http://localhost" + path
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("firecracker API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *FirecrackerClient) get(ctx context.Context, path string) (map[string]interface{}, error) {
	url := "http://localhost" + path
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("firecracker API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
