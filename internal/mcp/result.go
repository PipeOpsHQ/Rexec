package mcp

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// maxToolText caps a single tool response so a long command log does not
// consume the model's context window. The API already caps exec output at 1 MiB.
const maxToolText = 24 * 1024

func clip(s string) string {
	if len(s) <= maxToolText {
		return s
	}
	// Cut on a rune boundary.
	cut := maxToolText
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + "\n…[truncated]"
}

func textResult(s string) (*sdk.CallToolResult, any, error) {
	return &sdk.CallToolResult{
		Content: []sdk.Content{&sdk.TextContent{Text: clip(s)}},
	}, nil, nil
}

func jsonResult(v any) (*sdk.CallToolResult, any, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fail(err)
	}
	return textResult(string(b))
}

func fail(err error) (*sdk.CallToolResult, any, error) {
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}
	return &sdk.CallToolResult{
		IsError: true,
		Content: []sdk.Content{&sdk.TextContent{Text: clip(msg)}},
	}, nil, nil
}

func formatExec(r ExecResult) string {
	var b strings.Builder
	b.WriteString("exit_code: ")
	b.WriteString(itoa(r.ExitCode))
	b.WriteByte('\n')
	if r.DurationMS > 0 {
		b.WriteString("duration_ms: ")
		b.WriteString(itoa(int(r.DurationMS)))
		b.WriteByte('\n')
	}
	if r.Truncated {
		b.WriteString("truncated: true\n")
	}
	b.WriteString("--- stdout ---\n")
	b.WriteString(r.Stdout)
	if r.Stdout != "" && !strings.HasSuffix(r.Stdout, "\n") {
		b.WriteByte('\n')
	}
	if r.Stderr != "" {
		b.WriteString("--- stderr ---\n")
		b.WriteString(r.Stderr)
		if !strings.HasSuffix(r.Stderr, "\n") {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [16]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
