package container

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// DefaultRestrictedEgressAllow is the baseline host allowlist for restricted mode
// (package mirrors + common package registries). Callers may add more via create.
var DefaultRestrictedEgressAllow = []string{
	"archive.ubuntu.com",
	"security.ubuntu.com",
	"ports.ubuntu.com",
	"*.ubuntu.com",
	"*.archive.ubuntu.com",
	"deb.debian.org",
	"security.debian.org",
	"*.debian.org",
	"dl-cdn.alpinelinux.org",
	"pypi.org",
	"files.pythonhosted.org",
	"pypi.python.org",
	"registry.npmjs.org",
	"registry.yarnpkg.com",
	"github.com",
	"api.github.com",
	"codeload.github.com",
	"objects.githubusercontent.com",
	"raw.githubusercontent.com",
	"ghcr.io",
	"pkg-containers.githubusercontent.com",
	"proxy.golang.org",
	"sum.golang.org",
	"storage.googleapis.com",
	"crates.io",
	"static.crates.io",
	"index.crates.io",
	"rubygems.org",
	"index.rubygems.org",
	"repo1.maven.org",
	"repo.maven.apache.org",
	"cdn.jsdelivr.net",
}

// EgressSession is one sandbox's proxy credentials + allowlist.
type EgressSession struct {
	SandboxKey string // docker id or db id
	User       string
	Pass       string
	Allow      []string
	CreatedAt  time.Time
}

// EgressProxy is an HTTP/HTTPS (CONNECT) forward proxy with Basic auth
// and per-session host allowlists.
type EgressProxy struct {
	addr     string
	listener net.Listener
	mu       sync.RWMutex
	sessions map[string]*EgressSession // key: username
	stopCh   chan struct{}
}

// NewEgressProxy creates a proxy that will listen on listenAddr (e.g. ":13128").
func NewEgressProxy(listenAddr string) *EgressProxy {
	if listenAddr == "" {
		listenAddr = ":13128"
	}
	return &EgressProxy{
		addr:     listenAddr,
		sessions: make(map[string]*EgressSession),
		stopCh:   make(chan struct{}),
	}
}

// Start listens and serves. Non-blocking.
func (p *EgressProxy) Start() error {
	ln, err := net.Listen("tcp", p.addr)
	if err != nil {
		return fmt.Errorf("egress proxy listen %s: %w", p.addr, err)
	}
	p.listener = ln
	log.Printf("[EgressProxy] listening on %s", ln.Addr().String())
	go p.serve()
	return nil
}

// Stop closes the listener.
func (p *EgressProxy) Stop() {
	select {
	case <-p.stopCh:
	default:
		close(p.stopCh)
	}
	if p.listener != nil {
		_ = p.listener.Close()
	}
}

// Addr returns the bound address (host:port).
func (p *EgressProxy) Addr() string {
	if p.listener == nil {
		return p.addr
	}
	return p.listener.Addr().String()
}

// RegisterSession adds or replaces a sandbox egress session.
func (p *EgressProxy) RegisterSession(s *EgressSession) {
	if p == nil || s == nil || s.User == "" {
		return
	}
	s.CreatedAt = time.Now()
	p.mu.Lock()
	p.sessions[s.User] = s
	p.mu.Unlock()
}

// UnregisterSession removes a session by proxy username.
func (p *EgressProxy) UnregisterSession(user string) {
	if p == nil {
		return
	}
	p.mu.Lock()
	delete(p.sessions, user)
	p.mu.Unlock()
}

// UnregisterBySandbox removes sessions tied to a sandbox docker id.
func (p *EgressProxy) UnregisterBySandbox(sandboxKey string) {
	if p == nil || sandboxKey == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for user, s := range p.sessions {
		if s != nil && (s.SandboxKey == sandboxKey || strings.HasPrefix(s.SandboxKey, sandboxKey) || strings.HasPrefix(sandboxKey, s.SandboxKey)) {
			delete(p.sessions, user)
		}
	}
}

func (p *EgressProxy) serve() {
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			select {
			case <-p.stopCh:
				return
			default:
				log.Printf("[EgressProxy] accept: %v", err)
				continue
			}
		}
		go p.handleConn(conn)
	}
}

func (p *EgressProxy) handleConn(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(60 * time.Second))

	br := bufio.NewReader(conn)
	req, err := http.ReadRequest(br)
	if err != nil {
		return
	}

	sess, ok := p.authSession(req)
	if !ok {
		_ = writeProxyResponse(conn, http.StatusProxyAuthRequired, "Proxy Authentication Required")
		return
	}

	hostPort := req.Host
	if hostPort == "" {
		hostPort = req.URL.Host
	}
	host := hostPort
	if h, _, err := net.SplitHostPort(hostPort); err == nil {
		host = h
	}
	host = strings.ToLower(strings.TrimSpace(host))

	if !hostAllowed(host, sess.Allow) {
		log.Printf("[EgressProxy] deny %s for session %s", host, sess.SandboxKey)
		_ = writeProxyResponse(conn, http.StatusForbidden, "egress host not allowlisted: "+host)
		return
	}

	if req.Method == http.MethodConnect {
		p.handleConnect(conn, br, hostPort)
		return
	}

	// Plain HTTP proxy
	p.handleHTTP(conn, req, hostPort)
}

func (p *EgressProxy) authSession(req *http.Request) (*EgressSession, bool) {
	h := req.Header.Get("Proxy-Authorization")
	if h == "" {
		return nil, false
	}
	const prefix = "Basic "
	if !strings.HasPrefix(h, prefix) {
		return nil, false
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(h, prefix))
	if err != nil {
		return nil, false
	}
	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 {
		return nil, false
	}
	user, pass := parts[0], parts[1]
	p.mu.RLock()
	sess, ok := p.sessions[user]
	p.mu.RUnlock()
	if !ok || sess.Pass != pass {
		return nil, false
	}
	return sess, true
}

func (p *EgressProxy) handleConnect(client net.Conn, br *bufio.Reader, hostPort string) {
	if !strings.Contains(hostPort, ":") {
		hostPort += ":443"
	}
	dest, err := net.DialTimeout("tcp", hostPort, 15*time.Second)
	if err != nil {
		_ = writeProxyResponse(client, http.StatusBadGateway, "dial failed")
		return
	}
	defer dest.Close()

	if _, err := client.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n")); err != nil {
		return
	}
	_ = client.SetDeadline(time.Time{})
	_ = dest.SetDeadline(time.Time{})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(dest, br)
		closeWrite(dest)
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(client, dest)
		closeWrite(client)
	}()
	wg.Wait()
}

func closeWrite(c net.Conn) {
	if tc, ok := c.(*net.TCPConn); ok {
		_ = tc.CloseWrite()
	}
}

func (p *EgressProxy) handleHTTP(client net.Conn, req *http.Request, hostPort string) {
	if req.URL.Scheme == "" {
		req.URL.Scheme = "http"
	}
	if req.URL.Host == "" {
		req.URL.Host = hostPort
	}
	// Remove hop-by-hop
	req.RequestURI = ""
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("Proxy-Connection")

	transport := http.Transport{DialContext: (&net.Dialer{Timeout: 15 * time.Second}).DialContext}
	resp, err := transport.RoundTrip(req)
	if err != nil {
		_ = writeProxyResponse(client, http.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()
	if err := resp.Write(client); err != nil {
		return
	}
}

func writeProxyResponse(w io.Writer, code int, msg string) error {
	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\nContent-Type: text/plain\r\nConnection: close\r\nContent-Length: %d\r\n\r\n%s",
		code, http.StatusText(code), len(msg), msg)
	return err
}

// hostAllowed returns true if host matches any allow pattern (exact or *.suffix).
func hostAllowed(host string, allow []string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return false
	}
	for _, pat := range allow {
		pat = strings.ToLower(strings.TrimSpace(pat))
		if pat == "" {
			continue
		}
		if strings.HasPrefix(pat, "*.") {
			suf := pat[1:] // .example.com
			if strings.HasSuffix(host, suf) || host == pat[2:] {
				return true
			}
			continue
		}
		if host == pat {
			return true
		}
	}
	return false
}

// MergeEgressAllow unions defaults with extra hosts, de-duped, lowercased.
func MergeEgressAllow(extra []string) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(list []string) {
		for _, h := range list {
			h = strings.ToLower(strings.TrimSpace(h))
			if h == "" {
				continue
			}
			if _, ok := seen[h]; ok {
				continue
			}
			seen[h] = struct{}{}
			out = append(out, h)
		}
	}
	// Env override for defaults
	if env := strings.TrimSpace(os.Getenv("RESTRICTED_EGRESS_ALLOW")); env != "" {
		add(strings.Split(env, ","))
	} else {
		add(DefaultRestrictedEgressAllow)
	}
	add(extra)
	return out
}

// ParseEgressListen returns listen addr from RESTRICTED_EGRESS_PROXY_ADDR (default :13128).
func ParseEgressListen() string {
	if v := strings.TrimSpace(os.Getenv("RESTRICTED_EGRESS_PROXY_ADDR")); v != "" {
		return v
	}
	return ":13128"
}

// EgressEnabled is true unless RESTRICTED_EGRESS_ENABLED=false.
func EgressEnabled() bool {
	return !strings.EqualFold(os.Getenv("RESTRICTED_EGRESS_ENABLED"), "false")
}
