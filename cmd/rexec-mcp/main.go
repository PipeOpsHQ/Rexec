package main

import (
	"fmt"
	"os"

	"github.com/rexec/rexec/internal/mcp"
)

// Version is set with -ldflags "-X main.Version=...".
var Version = "dev"

func main() {
	if err := mcp.Run(Version, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "rexec-mcp: %v\n", err)
		os.Exit(1)
	}
}
