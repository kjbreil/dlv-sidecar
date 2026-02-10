package tools

import (
	"github.com/kjbreil/dlc-sidecar/internal/debugger"
	"github.com/mark3labs/mcp-go/server"
)

// Register adds all debugging tools to the MCP server.
func Register(s *server.MCPServer, addr string) {
	registerBreakpoints(s, addr)
	registerExecution(s, addr)
	registerVariables(s, addr)
	registerState(s, addr)
}

// dial creates a new connection to the Delve debugger.
func dial(addr string) (*debugger.Client, error) {
	return debugger.Dial(addr)
}
