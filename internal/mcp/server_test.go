package mcp

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestResolveSettingsOrder(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.json"
	if err := writeFile(path, `{"host":"https://from-file.example","token":"file-token"}`); err != nil {
		t.Fatal(err)
	}
	clearCredEnv(t)

	s, err := ResolveSettings(path, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if s.BaseURL != "https://from-file.example" || s.Token != "file-token" {
		t.Fatalf("file settings: %+v", s)
	}

	t.Setenv("REXEC_HOST", "https://from-env.example")
	t.Setenv("REXEC_TOKEN", "env-token")
	s, err = ResolveSettings(path, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if s.BaseURL != "https://from-env.example" || s.Token != "env-token" {
		t.Fatalf("env should override file: %+v", s)
	}

	t.Setenv("REXEC_URL", "https://from-url.example/")
	s, err = ResolveSettings(path, "https://from-flag.example", "flag-token")
	if err != nil {
		t.Fatal(err)
	}
	if s.BaseURL != "https://from-flag.example" || s.Token != "flag-token" {
		t.Fatalf("flags should override env: %+v", s)
	}
}

func TestResolveSettingsMissingToken(t *testing.T) {
	clearCredEnv(t)
	_, err := ResolveSettings(t.TempDir()+"/missing.json", "https://rexec.example", "")
	if err == nil || !strings.Contains(err.Error(), "REXEC_TOKEN") {
		t.Fatalf("expected missing token error, got %v", err)
	}
}

func TestToolsAgainstFakeAPI(t *testing.T) {
	pollInterval = 5 * time.Millisecond
	t.Cleanup(func() { pollInterval = 1500 * time.Millisecond })

	fake := newFakeAPI()
	srv := httptest.NewServer(fake.handler())
	t.Cleanup(srv.Close)

	api, err := NewClient(srv.URL, "test-token", "rexec-mcp/test")
	if err != nil {
		t.Fatal(err)
	}
	session := connect(t, NewServer(api, "test"))

	assertText := func(name string, args map[string]any, want string) {
		t.Helper()
		res := call(t, session, name, args)
		if res.IsError {
			t.Fatalf("%s error: %s", name, textOf(res))
		}
		got := textOf(res)
		if !strings.Contains(got, want) {
			t.Fatalf("%s result %q does not contain %q", name, got, want)
		}
	}

	assertText("whoami", nil, `"username": "ada"`)
	assertText("list_sandboxes", nil, `"id": "sb-1"`)
	assertText("list_images", nil, `"name": "ubuntu"`)
	assertText("create_sandbox", map[string]any{"name": "agent"}, `"status": "running"`)
	if fake.creates[0]["image"] != "ubuntu" {
		t.Fatalf("default image: %#v", fake.creates[0])
	}
	assertText("exec", map[string]any{"sandbox_id": "sb-1", "command": "echo hi"}, "exit_code: 0")
	assertText("exec", map[string]any{"sandbox_id": "sb-1", "command": "false"}, "exit_code: 1")
	if !strings.Contains(textOf(call(t, session, "exec", map[string]any{"sandbox_id": "sb-1", "command": "false"})), "--- stderr ---") {
		t.Fatal("expected stderr section")
	}

	assertText("write_file", map[string]any{
		"sandbox_id": "sb-1",
		"path":       "/home/user/src/main.py",
		"content":    "print(1)\n",
	}, `"ok": true`)
	if fake.lastWriteDir != "/home/user/src" || fake.lastWriteName != "main.py" || string(fake.lastWriteBody) != "print(1)\n" {
		t.Fatalf("upload dir=%q name=%q body=%q", fake.lastWriteDir, fake.lastWriteName, fake.lastWriteBody)
	}
	assertText("read_file", map[string]any{"sandbox_id": "sb-1", "path": "/home/user/src/main.py"}, "print(1)")
	assertText("list_files", map[string]any{"sandbox_id": "sb-1", "path": "/home/user"}, `"name": "src"`)
	assertText("mkdir", map[string]any{"sandbox_id": "sb-1", "path": "/home/user/work"}, `"ok": true`)
	assertText("start_sandbox", map[string]any{"sandbox_id": "sb-1"}, `"status": "running"`)
	assertText("delete_sandbox", map[string]any{"sandbox_id": "sb-1"}, `"deleted": true`)

	bad := call(t, session, "read_file", map[string]any{"sandbox_id": "sb-1", "path": "relative.txt"})
	if !bad.IsError {
		t.Fatal("relative path should be a tool error")
	}
	bad = call(t, session, "delete_file", map[string]any{"sandbox_id": "sb-1", "path": "/etc"})
	if !bad.IsError || !strings.Contains(textOf(bad), "refusing") {
		t.Fatalf("system path: %s", textOf(bad))
	}
	bad = call(t, session, "exec", map[string]any{"sandbox_id": "../x", "command": "id"})
	if !bad.IsError {
		t.Fatal("bad id should be a tool error")
	}

	unauthAPI, err := NewClient(srv.URL, "nope", "rexec-mcp/test")
	if err != nil {
		t.Fatal(err)
	}
	unauth := connect(t, NewServer(unauthAPI, "test"))
	denied := call(t, unauth, "whoami", nil)
	if !denied.IsError || !strings.Contains(textOf(denied), "unauthorized") {
		t.Fatalf("unauthorized: %s", textOf(denied))
	}
}

func TestCreateWaitsThroughCreating(t *testing.T) {
	pollInterval = time.Millisecond
	t.Cleanup(func() { pollInterval = 1500 * time.Millisecond })

	fake := newFakeAPI()
	fake.creatingPolls = 2
	srv := httptest.NewServer(fake.handler())
	t.Cleanup(srv.Close)
	api, err := NewClient(srv.URL, "test-token", "test")
	if err != nil {
		t.Fatal(err)
	}
	session := connect(t, NewServer(api, "test"))
	res := call(t, session, "create_sandbox", map[string]any{
		"image":           "debian",
		"network_mode":    "none",
		"timeout_seconds": 5,
	})
	if res.IsError {
		t.Fatal(textOf(res))
	}
	if fake.gets < 2 {
		t.Fatalf("expected polls, gets=%d", fake.gets)
	}
	if fake.creates[0]["image"] != "debian" || fake.creates[0]["network_mode"] != "none" {
		t.Fatalf("create body: %#v", fake.creates[0])
	}
	if !strings.Contains(textOf(res), `"status": "running"`) {
		t.Fatal(textOf(res))
	}
}

func TestExecTruncatesForContext(t *testing.T) {
	long := strings.Repeat("a", maxToolText+100)
	got := clip(formatExec(ExecResult{Stdout: long, ExitCode: 0}))
	if len(got) > maxToolText+32 {
		t.Fatalf("clipped length %d", len(got))
	}
	if !strings.Contains(got, "[truncated]") {
		t.Fatal("expected truncation marker")
	}
}

func TestInstructionsPresent(t *testing.T) {
	if !strings.Contains(Instructions, "list_sandboxes") || !strings.Contains(Instructions, "delete_sandbox") {
		t.Fatal("instructions should describe the workflow")
	}
}

type fakeAPI struct {
	mu               sync.Mutex
	gets             int
	creatingPolls    int
	creates          []map[string]any
	files            map[string][]byte
	lastWriteDir     string
	lastWriteName    string
	lastWriteBody    []byte
	lastAuth         string
	lastForwardedFor string
}

func newFakeAPI() *fakeAPI {
	return &fakeAPI{files: map[string][]byte{}}
}

func (f *fakeAPI) handler() http.Handler {
	mux := http.NewServeMux()
	auth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			f.mu.Lock()
			f.lastAuth = r.Header.Get("Authorization")
			f.lastForwardedFor = r.Header.Get("X-Forwarded-For")
			f.mu.Unlock()
			if r.Header.Get("Authorization") != "Bearer test-token" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
				return
			}
			next(w, r)
		}
	}
	mux.HandleFunc("GET /api/profile", auth(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"user": map[string]any{
			"id": "u1", "username": "ada", "email": "ada@example.com", "tier": "pro",
		}})
	}))
	mux.HandleFunc("GET /api/containers", auth(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"containers": []map[string]any{
			{"id": "sb-1", "name": "dev", "image": "ubuntu", "status": "running"},
		}, "count": 1})
	}))
	mux.HandleFunc("POST /api/containers", auth(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		f.mu.Lock()
		f.creates = append(f.creates, body)
		f.mu.Unlock()
		writeJSON(w, map[string]any{"id": "sb-new", "name": body["name"], "image": body["image"], "status": "creating"})
	}))
	mux.HandleFunc("GET /api/containers/{id}", auth(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.gets++
		n := f.gets
		need := f.creatingPolls
		f.mu.Unlock()
		status := "running"
		if r.PathValue("id") == "sb-new" && n <= need {
			status = "creating"
		}
		writeJSON(w, map[string]any{"id": r.PathValue("id"), "image": "ubuntu", "status": status})
	}))
	mux.HandleFunc("POST /api/containers/{id}/exec", auth(func(w http.ResponseWriter, r *http.Request) {
		var body ExecRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		code := 0
		stderr := ""
		if body.Command == "false" {
			code = 1
			stderr = "failed"
		}
		writeJSON(w, ExecResult{Stdout: body.Command + "\n", Stderr: stderr, ExitCode: code, DurationMS: 3})
	}))
	mux.HandleFunc("POST /api/containers/{id}/files", auth(func(w http.ResponseWriter, r *http.Request) {
		_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		part, err := mr.NextPart()
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		body, _ := io.ReadAll(part)
		dir := r.URL.Query().Get("path")
		f.mu.Lock()
		f.lastWriteDir = dir
		f.lastWriteName = part.FileName()
		f.lastWriteBody = body
		f.files[dir+"/"+part.FileName()] = body
		f.mu.Unlock()
		writeJSON(w, map[string]any{"ok": true})
	}))
	mux.HandleFunc("GET /api/containers/{id}/files/list", auth(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"path": r.URL.Query().Get("path"), "files": []FileInfo{
			{Name: "src", Path: "/home/user/src", IsDir: true},
		}, "count": 1})
	}))
	mux.HandleFunc("GET /api/containers/{id}/files", auth(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Query().Get("path")
		f.mu.Lock()
		b := f.files[p]
		f.mu.Unlock()
		if b == nil {
			http.Error(w, `{"error":"file not found"}`, 404)
			return
		}
		_, _ = w.Write(b)
	}))
	mux.HandleFunc("POST /api/containers/{id}/files/mkdir", auth(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("path") == "" {
			http.Error(w, `{"error":"path required"}`, 400)
			return
		}
		writeJSON(w, map[string]any{"message": "directory created"})
	}))
	mux.HandleFunc("POST /api/containers/{id}/start", auth(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"id": r.PathValue("id"), "status": "running"})
	}))
	mux.HandleFunc("DELETE /api/containers/{id}", auth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.HandleFunc("GET /api/images", auth(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"images": []Image{{Name: "ubuntu", DisplayName: "Ubuntu", Description: "LTS"}}})
	}))
	return mux
}

func connect(t *testing.T, server *sdk.Server) *sdk.ClientSession {
	t.Helper()
	ctx := context.Background()
	client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "v0"}, nil)
	sTransport, cTransport := sdk.NewInMemoryTransports()
	if _, err := server.Connect(ctx, sTransport, nil); err != nil {
		t.Fatal(err)
	}
	session, err := client.Connect(ctx, cTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func call(t *testing.T, session *sdk.ClientSession, name string, args map[string]any) *sdk.CallToolResult {
	t.Helper()
	res, err := session.CallTool(context.Background(), &sdk.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return res
}

func textOf(res *sdk.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if t, ok := c.(*sdk.TextContent); ok {
			b.WriteString(t.Text)
		}
	}
	return b.String()
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeFile(path, contents string) error {
	return os.WriteFile(path, []byte(contents), 0o600)
}

func clearCredEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"REXEC_TOKEN", "REXEC_URL", "REXEC_HOST", "REXEC_API", "REXEC_CONFIG"} {
		t.Setenv(k, "")
	}
}
