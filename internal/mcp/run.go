package mcp

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Run starts the MCP server. With no flags it speaks MCP on stdin and stdout,
// which is what Claude Desktop, Cursor, and other local clients launch.
// Pass --http 127.0.0.1:8090 to serve the streamable HTTP transport instead.
// Logs go to stderr so they do not corrupt the stdio protocol.
func Run(version string, args []string) error {
	if version == "" {
		version = "dev"
	}
	fs := flag.NewFlagSet("rexec-mcp", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `rexec-mcp %s — MCP server for Rexec sandboxes

Speaks the Model Context Protocol so an LLM can create sandboxes, run commands,
and read and write files. Credentials come from ~/.rexec/config.json (rexec login),
then REXEC_TOKEN and REXEC_URL / REXEC_HOST / REXEC_API, then flags.

Usage:
  rexec-mcp
  rexec-mcp --http 127.0.0.1:8090
  rexec mcp

Environment:
  REXEC_TOKEN   API bearer token (overrides the config file)
  REXEC_URL     API base URL (also REXEC_HOST, REXEC_API)
  REXEC_CONFIG  config file path (default ~/.rexec/config.json)

`, version)
		fs.PrintDefaults()
	}
	httpAddr := fs.String("http", "", "serve streamable HTTP on this address instead of stdio (example: 127.0.0.1:8090)")
	host := fs.String("host", "", "Rexec API base URL")
	token := fs.String("token", "", "API bearer token (prefer REXEC_TOKEN)")
	configPath := fs.String("config", "", "config file path")
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *showVersion {
		fmt.Fprintf(os.Stderr, "rexec-mcp %s\n", version)
		return nil
	}

	settings, err := ResolveSettings(*configPath, *host, *token)
	if err != nil {
		return err
	}
	client, err := NewClient(settings.BaseURL, settings.Token, "rexec-mcp/"+version)
	if err != nil {
		return err
	}
	server := NewServer(client, version)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if addr := strings.TrimSpace(*httpAddr); addr != "" {
		return serveHTTP(ctx, server, addr)
	}
	err = server.Run(ctx, &sdk.StdioTransport{})
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

func serveHTTP(ctx context.Context, server *sdk.Server, addr string) error {
	handler := sdk.NewStreamableHTTPHandler(func(*http.Request) *sdk.Server {
		return server
	}, nil)
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()
	fmt.Fprintf(os.Stderr, "rexec-mcp listening on http://%s\n", addr)
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
