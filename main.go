package main

import (
	"log"

	"github.com/kjbreil/dlc-sidecar/internal/debugger"
	"github.com/kjbreil/dlc-sidecar/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer(
		"dlc-sidecar",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	tools.Register(s, debugger.DefaultAddr)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
