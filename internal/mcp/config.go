package mcp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const defaultHost = "https://rexec.pipeops.io"

// Settings is the API endpoint and bearer token the MCP server will use.
type Settings struct {
	BaseURL string
	Token   string
}

type fileConfig struct {
	Host  string `json:"host"`
	Token string `json:"token"`
}

// ResolveSettings loads credentials the same way the CLI does.
// A config file is applied first, then environment variables, then explicit
// host and token arguments. An empty configPath uses REXEC_CONFIG or
// ~/.rexec/config.json.
func ResolveSettings(configPath, hostFlag, tokenFlag string) (Settings, error) {
	s := Settings{BaseURL: defaultHost}
	path := strings.TrimSpace(configPath)
	if path == "" {
		path = strings.TrimSpace(os.Getenv("REXEC_CONFIG"))
	}
	if path == "" {
		home, err := os.UserHomeDir()
		if err == nil && home != "" {
			path = filepath.Join(home, ".rexec", "config.json")
		}
	}
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return Settings{}, fmt.Errorf("read config: %w", err)
		}
		if err == nil && len(bytes.TrimSpace(data)) > 0 {
			var f fileConfig
			if err := json.Unmarshal(data, &f); err != nil {
				return Settings{}, fmt.Errorf("parse config %s: %w", path, err)
			}
			if strings.TrimSpace(f.Host) != "" {
				s.BaseURL = strings.TrimSpace(f.Host)
			}
			if strings.TrimSpace(f.Token) != "" {
				s.Token = strings.TrimSpace(f.Token)
			}
		}
	}
	if v := firstEnv("REXEC_URL", "REXEC_HOST", "REXEC_API"); v != "" {
		s.BaseURL = v
	}
	if v := strings.TrimSpace(os.Getenv("REXEC_TOKEN")); v != "" {
		s.Token = v
	}
	if strings.TrimSpace(hostFlag) != "" {
		s.BaseURL = strings.TrimSpace(hostFlag)
	}
	if strings.TrimSpace(tokenFlag) != "" {
		s.Token = strings.TrimSpace(tokenFlag)
	}
	s.BaseURL = strings.TrimRight(s.BaseURL, "/")
	u, err := url.Parse(s.BaseURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return Settings{}, fmt.Errorf("invalid rexec url %q", s.BaseURL)
	}
	if s.Token == "" {
		return Settings{}, errors.New("missing API token. Run rexec login or set REXEC_TOKEN")
	}
	return s, nil
}

func firstEnv(names ...string) string {
	for _, name := range names {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			return v
		}
	}
	return ""
}
