// Package rexec provides a Go SDK for interacting with Rexec - Terminal as a Service.
//
// The SDK allows you to programmatically create, manage, and interact with
// sandboxed Linux environments through the Rexec API.
//
// Basic usage:
//
//	client := rexec.NewClient("https://rexec.sh", "your-api-token")
//
//	// Create a sandbox (preferred)
//	sandbox, err := client.Sandboxes.Create(ctx, &rexec.CreateSandboxRequest{
//	    Image: "ubuntu",
//	    Name:  "my-sandbox",
//	})
//	// Legacy: client.Containers is the same service
//
//	term, err := client.Terminal.Connect(ctx, sandbox.ID)
//	term.Write([]byte("echo hello\n"))
package rexec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Client is the main Rexec API client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client

	// Sandboxes is the preferred accessor for sandbox lifecycle APIs.
	Sandboxes *SandboxService
	// Containers is a deprecated alias for Sandboxes (same pointer). Prefer Sandboxes.
	Containers *SandboxService
	Files      *FileService
	Terminal   *TerminalService
}

// NewClient creates a new Rexec client.
func NewClient(baseURL, token string) *Client {
	baseURL = strings.TrimSuffix(baseURL, "/")

	c := &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	svc := &SandboxService{client: c}
	c.Sandboxes = svc
	c.Containers = svc
	c.Files = &FileService{client: c}
	c.Terminal = &TerminalService{client: c}

	return c
}

// SetHTTPClient sets a custom HTTP client.
func (c *Client) SetHTTPClient(httpClient *http.Client) {
	c.httpClient = httpClient
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    errResp.Error,
		}
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) websocketURL(path string) string {
	u, _ := url.Parse(c.baseURL)
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}
	u.Path = path
	return u.String()
}

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// APIError represents an API error.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// Sandbox represents a Rexec sandbox (isolated Linux environment).
// Wire protocol still uses the /api/containers resource.
type Sandbox struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Status      string            `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// Container is a deprecated alias for Sandbox.
// Deprecated: use Sandbox.
type Container = Sandbox

// CreateSandboxRequest represents a request to create a sandbox.
// Prefer image aliases such as "ubuntu".
type CreateSandboxRequest struct {
	Name        string            `json:"name,omitempty"`
	Image       string            `json:"image"`
	Environment map[string]string `json:"environment,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// CreateContainerRequest is a deprecated alias for CreateSandboxRequest.
// Deprecated: use CreateSandboxRequest.
type CreateContainerRequest = CreateSandboxRequest

// SandboxService handles sandbox lifecycle operations.
// HTTP paths remain /api/containers for backend compatibility.
type SandboxService struct {
	client *Client
}

// ContainerService is a deprecated alias for SandboxService.
// Deprecated: use SandboxService.
type ContainerService = SandboxService

// containerListResponse matches the production API list payload.
type containerListResponse struct {
	Containers []Sandbox `json:"containers"`
	Count      int       `json:"count"`
	Limit      int       `json:"limit"`
}

// List returns all sandboxes for the authenticated user.
func (s *SandboxService) List(ctx context.Context) ([]Sandbox, error) {
	var wrapped containerListResponse
	if err := s.client.doRequest(ctx, http.MethodGet, "/api/containers", nil, &wrapped); err != nil {
		return nil, err
	}
	if wrapped.Containers == nil {
		return []Sandbox{}, nil
	}
	return wrapped.Containers, nil
}

// Get returns a sandbox by ID.
func (s *SandboxService) Get(ctx context.Context, id string) (*Sandbox, error) {
	var sandbox Sandbox
	err := s.client.doRequest(ctx, http.MethodGet, "/api/containers/"+id, nil, &sandbox)
	return &sandbox, err
}

// Create creates a new sandbox.
func (s *SandboxService) Create(ctx context.Context, req *CreateSandboxRequest) (*Sandbox, error) {
	var sandbox Sandbox
	err := s.client.doRequest(ctx, http.MethodPost, "/api/containers", req, &sandbox)
	return &sandbox, err
}

// Delete deletes a sandbox.
func (s *SandboxService) Delete(ctx context.Context, id string) error {
	return s.client.doRequest(ctx, http.MethodDelete, "/api/containers/"+id, nil, nil)
}

// Start starts a stopped sandbox.
func (s *SandboxService) Start(ctx context.Context, id string) error {
	return s.client.doRequest(ctx, http.MethodPost, "/api/containers/"+id+"/start", nil, nil)
}

// Stop stops a running sandbox.
func (s *SandboxService) Stop(ctx context.Context, id string) error {
	return s.client.doRequest(ctx, http.MethodPost, "/api/containers/"+id+"/stop", nil, nil)
}

// FileInfo represents file metadata.
type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	Mode    string    `json:"mode"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
}

// FileService handles file operations.
type FileService struct {
	client *Client
}

// List lists files in a container directory.
func (s *FileService) List(ctx context.Context, containerID, path string) ([]FileInfo, error) {
	var files []FileInfo
	endpoint := fmt.Sprintf("/api/containers/%s/files/list?path=%s", containerID, url.QueryEscape(path))
	err := s.client.doRequest(ctx, http.MethodGet, endpoint, nil, &files)
	return files, err
}

// Download downloads a file from a container.
func (s *FileService) Download(ctx context.Context, containerID, path string) ([]byte, error) {
	endpoint := fmt.Sprintf("/api/containers/%s/files?path=%s", containerID, url.QueryEscape(path))
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.client.token)

	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// Mkdir creates a directory in a container.
func (s *FileService) Mkdir(ctx context.Context, containerID, path string) error {
	return s.client.doRequest(ctx, http.MethodPost, "/api/containers/"+containerID+"/files/mkdir", map[string]string{"path": path}, nil)
}

// Terminal represents a WebSocket terminal connection.
type Terminal struct {
	conn *websocket.Conn
}

// TerminalService handles terminal connections.
type TerminalService struct {
	client *Client
}

// Connect establishes a WebSocket terminal connection to a container.
func (s *TerminalService) Connect(ctx context.Context, containerID string) (*Terminal, error) {
	wsURL := s.client.websocketURL("/ws/terminal/" + containerID)

	header := http.Header{}
	header.Set("Authorization", "Bearer "+s.client.token)

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, header)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to terminal: %w", err)
	}

	return &Terminal{conn: conn}, nil
}

// Write sends data to the terminal.
func (t *Terminal) Write(data []byte) error {
	return t.conn.WriteMessage(websocket.BinaryMessage, data)
}

// Read reads data from the terminal.
func (t *Terminal) Read() ([]byte, error) {
	_, data, err := t.conn.ReadMessage()
	return data, err
}

// Resize resizes the terminal.
func (t *Terminal) Resize(cols, rows int) error {
	msg := map[string]interface{}{
		"type": "resize",
		"cols": cols,
		"rows": rows,
	}
	return t.conn.WriteJSON(msg)
}

// Close closes the terminal connection.
func (t *Terminal) Close() error {
	return t.conn.Close()
}
