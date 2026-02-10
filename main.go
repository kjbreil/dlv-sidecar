package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kjbreil/dlc-sidecar/internal/debugger"
	"github.com/kjbreil/dlc-sidecar/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	port := flag.Int("port", 0, "Delve debugger port (default: 2345, or DLV_PORT env var)")
	flag.Parse()

	addr := resolveAddr(*port)

	pool := debugger.NewPool(addr)
	defer pool.Close()

	s := server.NewMCPServer(
		"dlc-sidecar",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	tools.Register(s, pool)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// resolveAddr determines the Delve debugger address from flag, env var, or default.
func resolveAddr(flagPort int) string {
	if flagPort != 0 {
		return fmt.Sprintf("localhost:%d", flagPort)
	}
	if envPort := os.Getenv("DLV_PORT"); envPort != "" {
		return fmt.Sprintf("localhost:%s", envPort)
	}
	return debugger.DefaultAddr
}
