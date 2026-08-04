package handlers

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	dockerclient "github.com/moby/moby/client"
	mgr "github.com/rexec/rexec/internal/container"
)

// maxExecTimeoutSeconds caps how long a single non-interactive exec may run.
const maxExecTimeoutSeconds = 300

// defaultExecTimeoutSeconds is used when the client omits timeout_seconds.
const defaultExecTimeoutSeconds = 60

// maxExecOutputBytes truncates captured stdout/stderr to protect the API.
const maxExecOutputBytes = 1 << 20 // 1 MiB

// ExecHandler runs non-interactive commands inside containers.
type ExecHandler struct {
	containerManager *mgr.Manager
}

// NewExecHandler creates an ExecHandler.
func NewExecHandler(cm *mgr.Manager) *ExecHandler {
	return &ExecHandler{containerManager: cm}
}

// ExecRequest is the body for POST /api/containers/:id/exec.
// Provide either Command (shell string) or Cmd (argv). Command is preferred
// for agent tooling and matches Cortex's expected shape.
type ExecRequest struct {
	// Command is run via ["sh", "-c", command] when Cmd is empty.
	Command string `json:"command"`
	// Cmd is an argv vector; used when non-empty (takes precedence over Command).
	Cmd []string `json:"cmd"`
	// WorkDir is the working directory inside the container (optional).
	WorkDir string `json:"workdir,omitempty"`
	// Env is extra environment variables as KEY=VALUE (optional).
	Env []string `json:"env,omitempty"`
	// User is the user to run as inside the container (optional).
	User string `json:"user,omitempty"`
	// TimeoutSeconds defaults to 60, max 300.
	TimeoutSeconds int `json:"timeout_seconds,omitempty"`
}

// Exec runs a one-shot command in a running container and returns captured output.
// POST /api/containers/:id/exec
//
// Response (Cortex / agent friendly):
//
//	{
//	  "stdout": "...",
//	  "stderr": "...",
//	  "output": "...",   // combined stdout+stderr for simple clients
//	  "exit_code": 0
//	}
func (h *ExecHandler) Exec(c *gin.Context) {
	userID := c.GetString("userID")
	containerID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if strings.TrimSpace(containerID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "container id required"})
		return
	}

	var req ExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body: " + err.Error()})
		return
	}

	cmd := req.Cmd
	if len(cmd) == 0 {
		command := strings.TrimSpace(req.Command)
		if command == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "command or cmd is required"})
			return
		}
		cmd = []string{"sh", "-c", command}
	}

	timeoutSecs := req.TimeoutSeconds
	if timeoutSecs <= 0 {
		timeoutSecs = defaultExecTimeoutSeconds
	}
	if timeoutSecs > maxExecTimeoutSeconds {
		timeoutSecs = maxExecTimeoutSeconds
	}

	// Resolve container: manager is keyed by Docker ID.
	containerInfo, ok := h.containerManager.GetContainer(containerID)
	if !ok {
		// Try partial / alternate match via user map if only short id was given.
		c.JSON(http.StatusNotFound, gin.H{"error": "container not found"})
		return
	}
	if containerInfo.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if !strings.EqualFold(containerInfo.Status, "running") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "container is not running"})
		return
	}

	dockerID := containerInfo.ID
	if dockerID == "" {
		dockerID = containerID
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	client := h.containerManager.GetClient()
	execConfig := dockerclient.ExecCreateOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Env:          req.Env,
		WorkingDir:   strings.TrimSpace(req.WorkDir),
		User:         strings.TrimSpace(req.User),
	}

	execResp, err := client.ExecCreate(ctx, dockerID, execConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create exec: " + err.Error()})
		return
	}

	attachResp, err := client.ExecAttach(ctx, execResp.ID, dockerclient.ExecAttachOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to attach exec: " + err.Error()})
		return
	}
	defer attachResp.Close()

	// Cap total bytes read to avoid OOM on runaway output.
	limited := io.LimitReader(attachResp.Reader, maxExecOutputBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil && err != io.EOF {
		// Context deadline often surfaces here for long-running commands.
		if ctx.Err() != nil {
			partial := demuxDockerStream(raw)
			c.JSON(http.StatusGatewayTimeout, gin.H{
				"error":     "exec timed out",
				"timeout_s": timeoutSecs,
				"exit_code": -1,
				"output":    partial.Combined,
				"stdout":    partial.Stdout,
				"stderr":    partial.Stderr,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read exec output: " + err.Error()})
		return
	}
	truncated := len(raw) > maxExecOutputBytes
	if truncated {
		raw = raw[:maxExecOutputBytes]
	}

	streams := demuxDockerStream(raw)

	exitCode := 0
	inspect, err := client.ExecInspect(ctx, execResp.ID, dockerclient.ExecInspectOptions{})
	if err == nil {
		exitCode = inspect.ExitCode
	}

	h.containerManager.TouchContainer(dockerID)

	resp := gin.H{
		"stdout":    streams.Stdout,
		"stderr":    streams.Stderr,
		"output":    streams.Combined,
		"exit_code": exitCode,
		"command":   req.Command,
	}
	if len(req.Cmd) > 0 {
		resp["cmd"] = req.Cmd
	}
	if truncated {
		resp["truncated"] = true
	}
	c.JSON(http.StatusOK, resp)
}

type demuxedStreams struct {
	Stdout   string
	Stderr   string
	Combined string
}

// demuxDockerStream splits a Docker multiplexed attach stream into stdout/stderr.
// If the payload is not multiplexed (plain text), it is treated as stdout.
func demuxDockerStream(raw []byte) demuxedStreams {
	if len(raw) == 0 {
		return demuxedStreams{}
	}

	// Heuristic: multiplexed frames start with stream type 1 or 2 and have size in big-endian.
	if len(raw) >= 8 && (raw[0] == 1 || raw[0] == 2) && raw[1] == 0 && raw[2] == 0 && raw[3] == 0 {
		var stdout, stderr bytes.Buffer
		for len(raw) >= 8 {
			streamType := raw[0]
			frameLen := int(binary.BigEndian.Uint32(raw[4:8]))
			raw = raw[8:]
			if frameLen < 0 || frameLen > len(raw) {
				// Corrupt / partial frame — append remainder as stdout.
				stdout.Write(raw)
				break
			}
			chunk := raw[:frameLen]
			raw = raw[frameLen:]
			switch streamType {
			case 1:
				stdout.Write(chunk)
			case 2:
				stderr.Write(chunk)
			default:
				stdout.Write(chunk)
			}
		}
		out := sanitizeExecText(stdout.Bytes())
		errOut := sanitizeExecText(stderr.Bytes())
		combined := out
		if errOut != "" {
			if combined != "" && !strings.HasSuffix(combined, "\n") {
				combined += "\n"
			}
			combined += errOut
		}
		return demuxedStreams{Stdout: out, Stderr: errOut, Combined: combined}
	}

	// Non-multiplexed (or already demuxed) plain output.
	text := sanitizeExecText(raw)
	return demuxedStreams{Stdout: text, Combined: text}
}

// sanitizeExecText strips NULs and ensures valid UTF-8 for JSON responses.
func sanitizeExecText(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	// Drop NULs which break some JSON consumers / terminals.
	b = bytes.ReplaceAll(b, []byte{0}, nil)
	if !utf8.Valid(b) {
		return strings.ToValidUTF8(string(b), "�")
	}
	return string(b)
}
