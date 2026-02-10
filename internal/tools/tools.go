package tools

import (
	"github.com/kjbreil/dlc-sidecar/internal/debugger"
	"github.com/mark3labs/mcp-go/server"
)

// Register adds all debugging tools to the MCP server.
func Register(s *server.MCPServer, pool *debugger.Pool) {
	registerBreakpoints(s, pool)
	registerExecution(s, pool)
	registerVariables(s, pool)
	registerState(s, pool)
}
