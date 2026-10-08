package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	maxReadBytes  = 256 * 1024
	maxWriteBytes = 512 * 1024
	maxJSONBytes  = 2 << 20
)

// Client calls the Rexec REST API with a user token.
type Client struct {
	BaseURL   string
	Token     string
	UserAgent string
	// ForwardedFor is the caller address to send as X-Forwarded-For so a
	// loopback call from the API server is rate-limited as that caller.
	ForwardedFor string
	HTTP         *http.Client
}

// Sandbox is the subset of a container record agents need.
type Sandbox struct {
	ID        string `json:"id"`
	DBID      string `json:"db_id,omitempty"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Role      string `json:"role,omitempty"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at,omitempty"`
}

// ExecResult is the JSON body of POST /api/containers/:id/exec.
type ExecResult struct {
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	Output     string `json:"output"`
	ExitCode   int    `json:"exit_code"`
	DurationMS int64  `json:"duration_ms"`
	Truncated  bool   `json:"truncated"`
	Command    string `json:"command,omitempty"`
}

// FileInfo is one entry from the file list endpoint.
type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	IsDir   bool   `json:"is_dir"`
	ModTime string `json:"mod_time,omitempty"`
}

// Image is a creatable sandbox image alias.
type Image struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

// Profile is the signed-in account, limited to fields an agent should see.
type Profile struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
	Tier     string `json:"tier"`
}

// ExecRequest is the body of a non-interactive exec.
type ExecRequest struct {
	Command        string   `json:"command,omitempty"`
	Cmd            []string `json:"cmd,omitempty"`
	WorkDir        string   `json:"workdir,omitempty"`
	Env            []string `json:"env,omitempty"`
	User           string   `json:"user,omitempty"`
	TimeoutSeconds int      `json:"timeout_seconds,omitempty"`
}

// NewClient returns a client that refuses cross-host redirects so the bearer
// token cannot follow a redirect off the configured Rexec host.
func NewClient(baseURL, token, userAgent string) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	u, err := url.Parse(baseURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("invalid rexec url %q", baseURL)
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("missing API token")
	}
	if userAgent == "" {
		userAgent = "rexec-mcp"
	}
	c := &Client{
		BaseURL:   baseURL,
		Token:     token,
		UserAgent: userAgent,
		HTTP:      &http.Client{},
	}
	c.HTTP.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("stopped after 5 redirects")
		}
		if req.URL.Host != via[0].URL.Host {
			return errors.New("refusing cross-host redirect")
		}
		return nil
	}
	return c, nil
}

func (c *Client) call(ctx context.Context, method, path string, body any, out any, limit int64) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, rdr)
	if err != nil {
		return err
	}
	c.authorize(req)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req, out, limit)
}

func (c *Client) authorize(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("User-Agent", c.UserAgent)
	if ip := strings.TrimSpace(c.ForwardedFor); ip != "" {
		req.Header.Set("X-Forwarded-For", ip)
		req.Header.Set("X-Real-IP", ip)
	}
}

func (c *Client) do(req *http.Request, out any, limit int64) error {
	if limit <= 0 {
		limit = maxJSONBytes
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return err
	}
	if int64(len(data)) > limit {
		return fmt.Errorf("rexec api response exceeded %d bytes", limit)
	}
	if resp.StatusCode >= 300 {
		return apiStatusError(resp.StatusCode, data)
	}
	if out == nil || len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if raw, ok := out.(*[]byte); ok {
		*raw = append([]byte(nil), data...)
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode rexec response: %w", err)
	}
	return nil
}

func apiStatusError(status int, body []byte) error {
	var wrapped struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &wrapped)
	msg := strings.TrimSpace(wrapped.Error)
	if msg == "" {
		msg = strings.TrimSpace(wrapped.Message)
	}
	if msg == "" {
		msg = strings.TrimSpace(string(body))
	}
	if len(msg) > 500 {
		msg = msg[:500]
	}
	if status == http.StatusUnauthorized {
		if msg == "" {
			msg = "token rejected"
		}
		return fmt.Errorf("unauthorized (%s). Run rexec login or set REXEC_TOKEN", msg)
	}
	if msg == "" {
		return fmt.Errorf("rexec api returned %d", status)
	}
	return fmt.Errorf("rexec api returned %d: %s", status, msg)
}

func (c *Client) Profile(ctx context.Context) (Profile, error) {
	var wrap struct {
		User Profile `json:"user"`
	}
	if err := c.call(ctx, http.MethodGet, "/api/profile", nil, &wrap, 0); err != nil {
		return Profile{}, err
	}
	return wrap.User, nil
}

func (c *Client) ListSandboxes(ctx context.Context) ([]Sandbox, error) {
	var raw []byte
	if err := c.call(ctx, http.MethodGet, "/api/containers", nil, &raw, 0); err != nil {
		return nil, err
	}
	list, err := decodeList[Sandbox](raw, "containers")
	return list, err
}

func (c *Client) GetSandbox(ctx context.Context, id string) (Sandbox, error) {
	var sb Sandbox
	err := c.call(ctx, http.MethodGet, "/api/containers/"+url.PathEscape(id), nil, &sb, 0)
	return sb, err
}

func (c *Client) CreateSandbox(ctx context.Context, body map[string]any) (Sandbox, error) {
	var sb Sandbox
	err := c.call(ctx, http.MethodPost, "/api/containers", body, &sb, 0)
	return sb, err
}

func (c *Client) DeleteSandbox(ctx context.Context, id string) error {
	return c.call(ctx, http.MethodDelete, "/api/containers/"+url.PathEscape(id), nil, nil, 0)
}

func (c *Client) StartSandbox(ctx context.Context, id string) (Sandbox, error) {
	var sb Sandbox
	err := c.call(ctx, http.MethodPost, "/api/containers/"+url.PathEscape(id)+"/start", map[string]any{}, &sb, 0)
	return sb, err
}

func (c *Client) StopSandbox(ctx context.Context, id string) (Sandbox, error) {
	var sb Sandbox
	err := c.call(ctx, http.MethodPost, "/api/containers/"+url.PathEscape(id)+"/stop", map[string]any{}, &sb, 0)
	return sb, err
}

func (c *Client) Exec(ctx context.Context, id string, req ExecRequest) (ExecResult, error) {
	var out ExecResult
	err := c.call(ctx, http.MethodPost, "/api/containers/"+url.PathEscape(id)+"/exec", req, &out, 0)
	return out, err
}

func (c *Client) ListFiles(ctx context.Context, id, dir string) ([]FileInfo, error) {
	q := url.Values{}
	q.Set("path", dir)
	var raw []byte
	err := c.call(ctx, http.MethodGet, "/api/containers/"+url.PathEscape(id)+"/files/list?"+q.Encode(), nil, &raw, 0)
	if err != nil {
		return nil, err
	}
	return decodeList[FileInfo](raw, "files")
}

// ReadFile downloads a file. truncated is true when the body was longer than maxReadBytes.
func (c *Client) ReadFile(ctx context.Context, id, filePath string) (data []byte, truncated bool, err error) {
	q := url.Values{}
	q.Set("path", filePath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/containers/"+url.PathEscape(id)+"/files?"+q.Encode(), nil)
	if err != nil {
		return nil, false, err
	}
	c.authorize(req)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxReadBytes+1))
	if err != nil {
		return nil, false, err
	}
	if resp.StatusCode >= 300 {
		return nil, false, apiStatusError(resp.StatusCode, body)
	}
	if int64(len(body)) > maxReadBytes {
		return body[:maxReadBytes], true, nil
	}
	return body, false, nil
}

// WriteFile uploads content to an absolute file path. The API expects the
// destination directory as the path query and the basename as the form filename.
func (c *Client) WriteFile(ctx context.Context, id, dir, filename string, content []byte) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	if _, err := part.Write(content); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("path", dir)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/containers/"+url.PathEscape(id)+"/files?"+q.Encode(), &buf)
	if err != nil {
		return err
	}
	c.authorize(req)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return c.do(req, nil, maxJSONBytes)
}

func (c *Client) Mkdir(ctx context.Context, id, dir string) error {
	q := url.Values{}
	q.Set("path", dir)
	return c.call(ctx, http.MethodPost, "/api/containers/"+url.PathEscape(id)+"/files/mkdir?"+q.Encode(), nil, nil, 0)
}

func (c *Client) DeleteFile(ctx context.Context, id, filePath string) error {
	q := url.Values{}
	q.Set("path", filePath)
	return c.call(ctx, http.MethodDelete, "/api/containers/"+url.PathEscape(id)+"/files?"+q.Encode(), nil, nil, 0)
}

func (c *Client) ListImages(ctx context.Context, all bool) ([]Image, error) {
	path := "/api/images"
	if all {
		path += "?all=true"
	}
	var wrap struct {
		Images []Image `json:"images"`
	}
	if err := c.call(ctx, http.MethodGet, path, nil, &wrap, 0); err != nil {
		return nil, err
	}
	return wrap.Images, nil
}

func (c *Client) ListTemplates(ctx context.Context) ([]map[string]any, error) {
	return c.listKeyed(ctx, "/api/templates", "templates")
}

func (c *Client) CreateTemplate(ctx context.Context, name, sandboxID, description string) (map[string]any, error) {
	body := map[string]any{"name": name, "from_sandbox_id": sandboxID}
	if description != "" {
		body["description"] = description
	}
	var out map[string]any
	err := c.call(ctx, http.MethodPost, "/api/templates", body, &out, 0)
	return out, err
}

func (c *Client) DeleteTemplate(ctx context.Context, id string) error {
	return c.call(ctx, http.MethodDelete, "/api/templates/"+url.PathEscape(id), nil, nil, 0)
}

func (c *Client) ListSnapshots(ctx context.Context) ([]map[string]any, error) {
	return c.listKeyed(ctx, "/api/snapshots", "snapshots")
}

func (c *Client) CreateSnapshot(ctx context.Context, sandboxID, name, description string) (map[string]any, error) {
	body := map[string]any{}
	if name != "" {
		body["name"] = name
	}
	if description != "" {
		body["description"] = description
	}
	var out map[string]any
	err := c.call(ctx, http.MethodPost, "/api/containers/"+url.PathEscape(sandboxID)+"/snapshot", body, &out, 0)
	return out, err
}

func (c *Client) ForkSandbox(ctx context.Context, sandboxID string, body map[string]any) (map[string]any, error) {
	var out map[string]any
	err := c.call(ctx, http.MethodPost, "/api/containers/"+url.PathEscape(sandboxID)+"/fork", body, &out, 0)
	return out, err
}

func (c *Client) listKeyed(ctx context.Context, path, key string) ([]map[string]any, error) {
	var raw []byte
	if err := c.call(ctx, http.MethodGet, path, nil, &raw, 0); err != nil {
		return nil, err
	}
	return decodeList[map[string]any](raw, key)
}

func decodeList[T any](data []byte, key string) ([]T, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return []T{}, nil
	}
	if data[0] == '[' {
		var list []T
		if err := json.Unmarshal(data, &list); err != nil {
			return nil, err
		}
		if list == nil {
			list = []T{}
		}
		return list, nil
	}
	var wrap map[string]json.RawMessage
	if err := json.Unmarshal(data, &wrap); err != nil {
		return nil, err
	}
	raw, ok := wrap[key]
	if !ok || len(bytes.TrimSpace(raw)) == 0 || string(raw) == "null" {
		return []T{}, nil
	}
	var list []T
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []T{}
	}
	return list, nil
}

// withTimeout adds a deadline when the caller did not set one.
func withTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, d)
}
