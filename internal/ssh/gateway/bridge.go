package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// TerminalBridge bridges an SSH session to a container terminal via WebSocket
type TerminalBridge struct {
	apiURL    string
	token     string
	conn      *websocket.Conn
	stdin     io.Reader
	stdout    io.Writer
	done      chan struct{}
	closeOnce sync.Once
	cols      int
	rows      int
}

// TerminalMessage represents a WebSocket message for terminal communication
type TerminalMessage struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

// BridgeConfig holds configuration for creating a terminal bridge
type BridgeConfig struct {
	APIURL      string
	Token       string
	ContainerID string
	Cols        int
	Rows        int
	Stdin       io.Reader
	Stdout      io.Writer
}

// NewTerminalBridge creates a new terminal bridge
func NewTerminalBridge(cfg BridgeConfig) (*TerminalBridge, error) {
	if cfg.Cols == 0 {
		cfg.Cols = 80
	}
	if cfg.Rows == 0 {
		cfg.Rows = 24
	}

	bridge := &TerminalBridge{
		apiURL: cfg.APIURL,
		token:  cfg.Token,
		stdin:  cfg.Stdin,
		stdout: cfg.Stdout,
		done:   make(chan struct{}),
		cols:   cfg.Cols,
		rows:   cfg.Rows,
	}

	// Connect to the terminal WebSocket
	if err := bridge.connect(cfg.ContainerID); err != nil {
		return nil, err
	}

	return bridge, nil
}

// connect establishes the WebSocket connection to the terminal
func (b *TerminalBridge) connect(containerID string) error {
	// Convert HTTP URL to WebSocket URL
	wsURL := strings.Replace(b.apiURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)

	// Build WebSocket URL with query params
	u, err := url.Parse(fmt.Sprintf("%s/ws/terminal/%s", wsURL, containerID))
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Add query parameters
	q := u.Query()
	q.Set("cols", fmt.Sprintf("%d", b.cols))
	q.Set("rows", fmt.Sprintf("%d", b.rows))
	u.RawQuery = q.Encode()

	// Set up headers with auth token
	headers := http.Header{}
	if b.token != "" {
		headers.Set("Authorization", "Bearer "+b.token)
	}

	// Connect to WebSocket
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, resp, err := dialer.Dial(u.String(), headers)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("WebSocket connection failed (status %d): %w", resp.StatusCode, err)
		}
		return fmt.Errorf("WebSocket connection failed: %w", err)
	}

	b.conn = conn
	log.Printf("[Bridge] Connected to terminal %s", containerID)
	return nil
}

// Start begins bidirectional communication between SSH and WebSocket
func (b *TerminalBridge) Start(ctx context.Context) error {
	// Send initial resize
	if err := b.sendResize(b.cols, b.rows); err != nil {
		log.Printf("[Bridge] Failed to send initial resize: %v", err)
	}

	errChan := make(chan error, 2)

	// Read from WebSocket, write to SSH stdout
	go func() {
		errChan <- b.readLoop()
	}()

	// Read from SSH stdin, write to WebSocket
	go func() {
		errChan <- b.writeLoop()
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		b.Close()
		return ctx.Err()
	case err := <-errChan:
		b.Close()
		return err
	case <-b.done:
		return nil
	}
}

// readLoop reads from WebSocket and writes to SSH stdout
func (b *TerminalBridge) readLoop() error {
	for {
		select {
		case <-b.done:
			return nil
		default:
		}

		_, message, err := b.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return nil
			}
			return fmt.Errorf("WebSocket read error: %w", err)
		}

		// Parse the message
		var msg TerminalMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			// Try as raw output
			if _, err := b.stdout.Write(message); err != nil {
				return fmt.Errorf("stdout write error: %w", err)
			}
			continue
		}

		switch msg.Type {
		case "output":
			if _, err := b.stdout.Write([]byte(msg.Data)); err != nil {
				return fmt.Errorf("stdout write error: %w", err)
			}
		case "error":
			log.Printf("[Bridge] Terminal error: %s", msg.Data)
		case "exit":
			log.Printf("[Bridge] Terminal exited")
			return nil
		}
	}
}

// writeLoop reads from SSH stdin and writes to WebSocket
func (b *TerminalBridge) writeLoop() error {
	buf := make([]byte, 4096)

	for {
		select {
		case <-b.done:
			return nil
		default:
		}

		n, err := b.stdin.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("stdin read error: %w", err)
		}

		if n > 0 {
			// Check for escape sequence (Ctrl+])
			if n == 1 && buf[0] == 0x1d { // Ctrl+]
				log.Printf("[Bridge] Escape sequence detected, returning to dashboard")
				return nil
			}

			msg := TerminalMessage{
				Type: "input",
				Data: string(buf[:n]),
			}

			data, err := json.Marshal(msg)
			if err != nil {
				return fmt.Errorf("JSON marshal error: %w", err)
			}

			if err := b.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return fmt.Errorf("WebSocket write error: %w", err)
			}
		}
	}
}

// Resize sends a resize message to the terminal
func (b *TerminalBridge) Resize(cols, rows int) error {
	b.cols = cols
	b.rows = rows
	return b.sendResize(cols, rows)
}

// sendResize sends a resize message over WebSocket
func (b *TerminalBridge) sendResize(cols, rows int) error {
	if b.conn == nil {
		return nil
	}

	msg := TerminalMessage{
		Type: "resize",
		Cols: cols,
		Rows: rows,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return b.conn.WriteMessage(websocket.TextMessage, data)
}

// Close closes the bridge connection
func (b *TerminalBridge) Close() error {
	b.closeOnce.Do(func() {
		close(b.done)
		if b.conn != nil {
			b.conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			b.conn.Close()
		}
		log.Printf("[Bridge] Connection closed")
	})
	return nil
}

// AgentBridge bridges an SSH session to an agent terminal
type AgentBridge struct {
	*TerminalBridge
}

// NewAgentBridge creates a new agent terminal bridge
func NewAgentBridge(cfg BridgeConfig, agentID string) (*AgentBridge, error) {
	if cfg.Cols == 0 {
		cfg.Cols = 80
	}
	if cfg.Rows == 0 {
		cfg.Rows = 24
	}

	bridge := &TerminalBridge{
		apiURL: cfg.APIURL,
		token:  cfg.Token,
		stdin:  cfg.Stdin,
		stdout: cfg.Stdout,
		done:   make(chan struct{}),
		cols:   cfg.Cols,
		rows:   cfg.Rows,
	}

	// Connect to the agent terminal WebSocket
	if err := bridge.connectAgent(agentID); err != nil {
		return nil, err
	}

	return &AgentBridge{TerminalBridge: bridge}, nil
}

// connectAgent establishes the WebSocket connection to an agent terminal
func (b *TerminalBridge) connectAgent(agentID string) error {
	// Convert HTTP URL to WebSocket URL
	wsURL := strings.Replace(b.apiURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)

	// Build WebSocket URL with query params
	u, err := url.Parse(fmt.Sprintf("%s/ws/agent/%s/terminal", wsURL, agentID))
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Add query parameters
	q := u.Query()
	q.Set("cols", fmt.Sprintf("%d", b.cols))
	q.Set("rows", fmt.Sprintf("%d", b.rows))
	u.RawQuery = q.Encode()

	// Set up headers with auth token
	headers := http.Header{}
	if b.token != "" {
		headers.Set("Authorization", "Bearer "+b.token)
	}

	// Connect to WebSocket
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, resp, err := dialer.Dial(u.String(), headers)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("Agent WebSocket connection failed (status %d): %w", resp.StatusCode, err)
		}
		return fmt.Errorf("Agent WebSocket connection failed: %w", err)
	}

	b.conn = conn
	log.Printf("[Bridge] Connected to agent %s", agentID)
	return nil
}

// BridgeSession represents an active terminal bridge session
type BridgeSession struct {
	ID          string
	ContainerID string
	AgentID     string
	Bridge      *TerminalBridge
	StartedAt   time.Time
}

// BridgeManager manages active terminal bridge sessions
type BridgeManager struct {
	sessions map[string]*BridgeSession
	mu       sync.RWMutex
}

// NewBridgeManager creates a new bridge manager
func NewBridgeManager() *BridgeManager {
	return &BridgeManager{
		sessions: make(map[string]*BridgeSession),
	}
}

// AddSession adds a bridge session
func (m *BridgeManager) AddSession(sessionID string, session *BridgeSession) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[sessionID] = session
}

// RemoveSession removes a bridge session
func (m *BridgeManager) RemoveSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if session, ok := m.sessions[sessionID]; ok {
		if session.Bridge != nil {
			session.Bridge.Close()
		}
		delete(m.sessions, sessionID)
	}
}

// GetSession returns a bridge session by ID
func (m *BridgeManager) GetSession(sessionID string) (*BridgeSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[sessionID]
	return session, ok
}

// CloseAll closes all active bridge sessions
func (m *BridgeManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, session := range m.sessions {
		if session.Bridge != nil {
			session.Bridge.Close()
		}
		delete(m.sessions, id)
	}
}
