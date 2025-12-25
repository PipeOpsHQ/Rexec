package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
)

const (
	DefaultHost = "https://rexec.pipeops.io"
	ConfigDir   = ".rexec"
	ConfigFile  = "config.json"
)

// Styles
var (
	// Colors
	primaryColor   = lipgloss.Color("#00D4FF")
	secondaryColor = lipgloss.Color("#7C3AED")
	successColor   = lipgloss.Color("#10B981")
	warningColor   = lipgloss.Color("#F59E0B")
	errorColor     = lipgloss.Color("#EF4444")
	dimColor       = lipgloss.Color("#6B7280")
	bgColor        = lipgloss.Color("#0F172A")
	cardBgColor    = lipgloss.Color("#1E293B")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1E293B")).
			Foreground(primaryColor).
			Bold(true).
			Padding(1, 2).
			Width(80)

	cardStyle = lipgloss.NewStyle().
			Background(cardBgColor).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimColor).
			Padding(1, 2)

	selectedCardStyle = lipgloss.NewStyle().
				Background(cardBgColor).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(1, 2)

	statusRunningStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	statusStoppedStyle = lipgloss.NewStyle().
				Foreground(warningColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			Padding(1, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(dimColor)
)

// Config
type Config struct {
	Host     string `json:"host"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Tier     string `json:"tier"`
}

// Models
type Container struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Status    string `json:"status"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	MFALocked bool   `json:"mfa_locked"`
}

type Snippet struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Content  string `json:"content,omitempty"`
	Command  string `json:"command,omitempty"` // legacy
	Desc     string `json:"description"`
	Language string `json:"language,omitempty"`
	Category string `json:"category,omitempty"`
}

// List item implementations
func (c Container) Title() string { return c.Name }
func (c Container) Description() string {
	lockIcon := ""
	if c.MFALocked {
		lockIcon = "🔒 "
	}
	return fmt.Sprintf("%s%s • %s • %s", lockIcon, c.Image, c.Status, c.Role)
}
func (c Container) FilterValue() string { return c.Name }

func (s Snippet) Title() string       { return s.Name }
func (s Snippet) Description() string { return s.Desc }
func (s Snippet) FilterValue() string { return s.Name }

// Views
type View int

const (
	ViewDashboard View = iota
	ViewTerminals
	ViewSnippets
	ViewTerminal
	ViewCreate
	ViewMFAPrompt
)

// Messages
type containersMsg []Container
type snippetsMsg []Snippet
type errorMsg error
type terminalOutputMsg string
type tickMsg time.Time
type mfaVerifiedMsg struct{ containerID string }
type mfaFailedMsg struct{ err string }

// Model
type model struct {
	config *Config
	view   View
	width  int
	height int

	// Lists
	terminalList list.Model
	snippetList  list.Model

	// Data
	containers []Container
	snippets   []Snippet

	// UI State
	selectedIdx int
	loading     bool
	spinner     spinner.Model
	err         error

	// Terminal
	terminalConn   *websocket.Conn
	terminalOutput string
	viewport       viewport.Model

	// MFA
	mfaInput       string
	mfaContainerID string
	mfaError       string
}

func initialModel() model {
	cfg := loadConfig()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)

	// Terminal list
	terminalDelegate := list.NewDefaultDelegate()
	terminalDelegate.Styles.SelectedTitle = terminalDelegate.Styles.SelectedTitle.Foreground(primaryColor)
	terminalDelegate.Styles.SelectedDesc = terminalDelegate.Styles.SelectedDesc.Foreground(dimColor)

	terminalList := list.New([]list.Item{}, terminalDelegate, 0, 0)
	terminalList.Title = "Terminals"
	terminalList.SetShowHelp(false)
	terminalList.SetFilteringEnabled(true)
	terminalList.Styles.Title = titleStyle

	// Snippet list
	snippetDelegate := list.NewDefaultDelegate()
	snippetList := list.New([]list.Item{}, snippetDelegate, 0, 0)
	snippetList.Title = "Snippets"
	snippetList.SetShowHelp(false)
	snippetList.Styles.Title = titleStyle

	return model{
		config:       cfg,
		view:         ViewDashboard,
		terminalList: terminalList,
		snippetList:  snippetList,
		spinner:      s,
		loading:      true,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchContainers(m.config),
		fetchSnippets(m.config),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.view == ViewDashboard {
				return m, tea.Quit
			}
			m.view = ViewDashboard
			return m, nil
		case "esc":
			if m.view != ViewDashboard {
				m.view = ViewDashboard
			}
			return m, nil
		case "r":
			m.loading = true
			return m, tea.Batch(
				fetchContainers(m.config),
				fetchSnippets(m.config),
			)
		case "t":
			m.view = ViewTerminals
			return m, nil
		case "s":
			m.view = ViewSnippets
			return m, nil
		case "c":
			return m, createTerminal(m.config)
		case "enter":
			if m.view == ViewMFAPrompt {
				// Submit MFA code
				if len(m.mfaInput) == 6 {
					return m, verifyMFA(m.config, m.mfaContainerID, m.mfaInput)
				}
				m.mfaError = "Enter a 6-digit code"
				return m, nil
			}
			if m.view == ViewTerminals && len(m.containers) > 0 {
				idx := m.terminalList.Index()
				if idx >= 0 && idx < len(m.containers) {
					container := m.containers[idx]
					if container.MFALocked {
						m.view = ViewMFAPrompt
						m.mfaContainerID = container.ID
						m.mfaInput = ""
						m.mfaError = ""
						return m, nil
					}
					return m, connectToTerminal(m.config, container.ID)
				}
			}
		case "backspace":
			if m.view == ViewMFAPrompt && len(m.mfaInput) > 0 {
				m.mfaInput = m.mfaInput[:len(m.mfaInput)-1]
				return m, nil
			}
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// MFA input takes priority
			if m.view == ViewMFAPrompt && len(m.mfaInput) < 6 {
				m.mfaInput += msg.String()
				return m, nil
			}
			// Quick connect from dashboard (1-9 keys, not 0)
			if m.view == ViewDashboard && msg.String() != "0" {
				idx := int(msg.String()[0] - '1')
				if idx < len(m.containers) {
					container := m.containers[idx]
					if container.MFALocked {
						m.view = ViewMFAPrompt
						m.mfaContainerID = container.ID
						m.mfaInput = ""
						m.mfaError = ""
						return m, nil
					}
					return m, connectToTerminal(m.config, container.ID)
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.terminalList.SetSize(msg.Width-4, msg.Height-10)
		m.snippetList.SetSize(msg.Width-4, msg.Height-10)

	case containersMsg:
		m.loading = false
		m.containers = msg
		items := make([]list.Item, len(msg))
		for i, c := range msg {
			items[i] = c
		}
		m.terminalList.SetItems(items)

	case snippetsMsg:
		m.snippets = msg
		items := make([]list.Item, len(msg))
		for i, s := range msg {
			items[i] = s
		}
		m.snippetList.SetItems(items)

	case errorMsg:
		m.loading = false
		m.err = msg

	case mfaVerifiedMsg:
		m.view = ViewDashboard
		m.mfaInput = ""
		m.mfaError = ""
		return m, connectToTerminal(m.config, msg.containerID)

	case mfaFailedMsg:
		m.mfaError = msg.err
		m.mfaInput = ""

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update sub-models
	switch m.view {
	case ViewTerminals:
		var cmd tea.Cmd
		m.terminalList, cmd = m.terminalList.Update(msg)
		cmds = append(cmds, cmd)
	case ViewSnippets:
		var cmd tea.Cmd
		m.snippetList, cmd = m.snippetList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var s strings.Builder

	// Header
	header := m.renderHeader()
	s.WriteString(header)
	s.WriteString("\n")

	// Content based on view
	switch m.view {
	case ViewDashboard:
		s.WriteString(m.renderDashboard())
	case ViewTerminals:
		s.WriteString(m.terminalList.View())
	case ViewSnippets:
		s.WriteString(m.snippetList.View())
	case ViewMFAPrompt:
		s.WriteString(m.renderMFAPrompt())
	}

	// Footer/Help
	s.WriteString("\n")
	s.WriteString(m.renderHelp())

	return s.String()
}

func (m model) renderHeader() string {
	logo := `
  ██████╗ ███████╗██╗  ██╗███████╗ ██████╗
  ██╔══██╗██╔════╝╚██╗██╔╝██╔════╝██╔════╝
  ██████╔╝█████╗   ╚███╔╝ █████╗  ██║
  ██╔══██╗██╔══╝   ██╔██╗ ██╔══╝  ██║
  ██║  ██║███████╗██╔╝ ██╗███████╗╚██████╗
  ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚══════╝ ╚═════╝`

	logoStyle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)

	userInfo := dimStyle.Render(fmt.Sprintf("  %s • %s", m.config.Username, m.config.Tier))

	return logoStyle.Render(logo) + "\n" + userInfo
}

func (m model) renderDashboard() string {
	var s strings.Builder

	s.WriteString("\n")

	// Stats row
	runningCount := 0
	for _, c := range m.containers {
		if c.Status == "running" {
			runningCount++
		}
	}

	statsStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Margin(0, 1)

	statBox := func(label, value string, color lipgloss.Color) string {
		return statsStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				lipgloss.NewStyle().Foreground(color).Bold(true).Render(value),
				dimStyle.Render(label),
			),
		)
	}

	stats := lipgloss.JoinHorizontal(
		lipgloss.Top,
		statBox("Total", fmt.Sprintf("%d", len(m.containers)), primaryColor),
		statBox("Running", fmt.Sprintf("%d", runningCount), successColor),
		statBox("Stopped", fmt.Sprintf("%d", len(m.containers)-runningCount), warningColor),
		statBox("Snippets", fmt.Sprintf("%d", len(m.snippets)), secondaryColor),
	)

	s.WriteString(stats)
	s.WriteString("\n\n")

	// Recent terminals
	s.WriteString(titleStyle.Render("  Recent Terminals"))
	s.WriteString("\n\n")

	if m.loading {
		s.WriteString(fmt.Sprintf("  %s Loading...\n", m.spinner.View()))
	} else if len(m.containers) == 0 {
		s.WriteString(dimStyle.Render("  No terminals yet. Press 'c' to create one.\n"))
	} else {
		maxShow := 5
		if len(m.containers) < maxShow {
			maxShow = len(m.containers)
		}

		for i := 0; i < maxShow; i++ {
			c := m.containers[i]

			statusIcon := "○"
			statusStyle := statusStoppedStyle
			if c.Status == "running" {
				statusIcon = "●"
				statusStyle = statusRunningStyle
			}

			shortID := c.ID
			if len(shortID) > 8 {
				shortID = shortID[:8]
			}

			line := fmt.Sprintf("  %s %d. %-25s %s %s\n",
				statusStyle.Render(statusIcon),
				i+1,
				c.Name,
				dimStyle.Render(c.Image),
				dimStyle.Render("["+shortID+"]"),
			)
			s.WriteString(line)
		}

		if len(m.containers) > 5 {
			s.WriteString(dimStyle.Render(fmt.Sprintf("\n  ... and %d more (press 't' to see all)\n", len(m.containers)-5)))
		}
	}

	return s.String()
}

func (m model) renderMFAPrompt() string {
	var s strings.Builder

	s.WriteString("\n\n")
	s.WriteString(titleStyle.Render("  🔒 MFA Required"))
	s.WriteString("\n\n")
	s.WriteString(dimStyle.Render("  This terminal is protected with MFA."))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  Enter your 6-digit authenticator code:"))
	s.WriteString("\n\n")

	// Display input with masked/visible digits
	codeDisplay := "  "
	for i := 0; i < 6; i++ {
		if i < len(m.mfaInput) {
			codeDisplay += lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(string(m.mfaInput[i])) + " "
		} else {
			codeDisplay += dimStyle.Render("_") + " "
		}
	}
	s.WriteString(codeDisplay)
	s.WriteString("\n\n")

	if m.mfaError != "" {
		s.WriteString(errorStyle.Render("  " + m.mfaError))
		s.WriteString("\n")
	}

	s.WriteString(dimStyle.Render("  Press Enter to verify • Esc to cancel"))

	return s.String()
}

func (m model) renderHelp() string {
	var keys []string

	switch m.view {
	case ViewDashboard:
		keys = []string{
			"1-9 connect",
			"t terminals",
			"s snippets",
			"c create",
			"r refresh",
			"q quit",
		}
	case ViewTerminals, ViewSnippets:
		keys = []string{
			"↑↓ navigate",
			"enter select",
			"/ filter",
			"esc back",
			"q quit",
		}
	case ViewMFAPrompt:
		keys = []string{
			"0-9 enter code",
			"enter verify",
			"esc cancel",
		}
	}

	helpItems := make([]string, len(keys))
	for i, k := range keys {
		parts := strings.SplitN(k, " ", 2)
		if len(parts) == 2 {
			helpItems[i] = lipgloss.NewStyle().Foreground(primaryColor).Render(parts[0]) + " " + dimStyle.Render(parts[1])
		} else {
			helpItems[i] = dimStyle.Render(k)
		}
	}

	return helpStyle.Render(strings.Join(helpItems, "  •  "))
}

// Commands
func fetchContainers(cfg *Config) tea.Cmd {
	return func() tea.Msg {
		containers, err := apiGetContainers(cfg)
		if err != nil {
			return errorMsg(err)
		}
		return containersMsg(containers)
	}
}

func fetchSnippets(cfg *Config) tea.Cmd {
	return func() tea.Msg {
		snippets, err := apiGetSnippets(cfg)
		if err != nil {
			return errorMsg(err)
		}
		return snippetsMsg(snippets)
	}
}

func createTerminal(cfg *Config) tea.Cmd {
	return func() tea.Msg {
		// For now, just create with defaults
		body := map[string]interface{}{
			"name":  fmt.Sprintf("terminal-%d", time.Now().Unix()),
			"image": "ubuntu",
			"role":  "default",
		}

		_, err := apiRequest(cfg, "POST", "/api/containers", body)
		if err != nil {
			return errorMsg(err)
		}

		// Refresh list
		containers, err := apiGetContainers(cfg)
		if err != nil {
			return errorMsg(err)
		}
		return containersMsg(containers)
	}
}

func connectToTerminal(cfg *Config, containerID string) tea.Cmd {
	return func() tea.Msg {
		// Launch external terminal connection
		// This exits the TUI and connects via the CLI
		cmd := exec.Command("rexec", "connect", containerID)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return nil
	}
}

func verifyMFA(cfg *Config, containerID, code string) tea.Cmd {
	return func() tea.Msg {
		body := map[string]interface{}{
			"code": code,
		}

		data, err := apiRequest(cfg, "POST", "/api/security/terminal/"+containerID+"/mfa-verify", body)
		if err != nil {
			return mfaFailedMsg{err: "Failed to verify MFA"}
		}

		var resp struct {
			Verified bool   `json:"verified"`
			Error    string `json:"error"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			return mfaFailedMsg{err: "Invalid response"}
		}

		if resp.Error != "" {
			return mfaFailedMsg{err: resp.Error}
		}

		if !resp.Verified {
			return mfaFailedMsg{err: "Invalid MFA code"}
		}

		return mfaVerifiedMsg{containerID: containerID}
	}
}

// API helpers
func loadConfig() *Config {
	cfg := &Config{Host: DefaultHost}

	if host := os.Getenv("REXEC_HOST"); host != "" {
		cfg.Host = host
	}
	if token := os.Getenv("REXEC_TOKEN"); token != "" {
		cfg.Token = token
	}

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ConfigDir, ConfigFile)

	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, cfg)
	}

	if host := os.Getenv("REXEC_HOST"); host != "" {
		cfg.Host = host
	}
	if token := os.Getenv("REXEC_TOKEN"); token != "" {
		cfg.Token = token
	}

	return cfg
}

func apiRequest(cfg *Config, method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewReader(data)
	}

	reqURL := cfg.Host + endpoint
	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.Token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func apiGetContainers(cfg *Config) ([]Container, error) {
	data, err := apiRequest(cfg, "GET", "/api/containers", nil)
	if err != nil {
		return nil, err
	}

	var wrapped struct {
		Containers []Container `json:"containers"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.Containers != nil {
		return wrapped.Containers, nil
	}

	var containers []Container
	err = json.Unmarshal(data, &containers)
	return containers, err
}

func apiGetSnippets(cfg *Config) ([]Snippet, error) {
	data, err := apiRequest(cfg, "GET", "/api/snippets", nil)
	if err != nil {
		return nil, err
	}

	var wrapped struct {
		Snippets []Snippet `json:"snippets"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.Snippets != nil {
		return wrapped.Snippets, nil
	}

	var snippets []Snippet
	err = json.Unmarshal(data, &snippets)
	return snippets, err
}

// Keymap
type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Back    key.Binding
	Quit    key.Binding
	Create  key.Binding
	Refresh key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Create: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "create"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
}

func main() {
	cfg := loadConfig()

	if cfg.Token == "" {
		fmt.Println("Not logged in. Run 'rexec login' first.")
		os.Exit(1)
	}

	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// Unused but needed for websocket import
var _ = url.QueryEscape
var _ = websocket.DefaultDialer
