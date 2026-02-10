package tools

import (
	"context"
	"fmt"

	"github.com/kjbreil/dlc-sidecar/internal/debugger"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerState(s *server.MCPServer, pool *debugger.Pool) {
	// get_state
	s.AddTool(mcp.NewTool("get_state",
		mcp.WithDescription("Get the current debugger state including position, goroutine, and thread info"),
	), makeGetState(pool))

	// stacktrace
	s.AddTool(mcp.NewTool("stacktrace",
		mcp.WithDescription("Get a stacktrace of the current goroutine"),
		mcp.WithNumber("goroutineID",
			mcp.Description("Goroutine ID (default: -1 for current)"),
		),
		mcp.WithNumber("depth",
			mcp.Description("Maximum stack depth to return (default: 50)"),
		),
		mcp.WithBoolean("full",
			mcp.Description("Include local variables and arguments in each frame (default: false)"),
		),
	), makeStacktrace(pool))

	// list_goroutines
	s.AddTool(mcp.NewTool("list_goroutines",
		mcp.WithDescription("List all goroutines in the debugged process"),
		mcp.WithNumber("count",
			mcp.Description("Maximum number of goroutines to return (default: 100)"),
		),
	), makeListGoroutines(pool))
}

func makeGetState(pool *debugger.Pool) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		req := debugger.StateIn{NonBlocking: true}
		var resp debugger.StateOut
		if err := pool.Call("State", req, &resp); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("RPC call failed: %v", err)), nil
		}

		return jsonResult(map[string]interface{}{
			"state": resp.State,
		})
	}
}

func makeStacktrace(pool *debugger.Pool) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		goroutineID := int64(-1)
		depth := 50
		full := false
		if v, err := request.RequireInt("goroutineID"); err == nil {
			goroutineID = int64(v)
		}
		if v, err := request.RequireInt("depth"); err == nil {
			depth = v
		}
		if v, err := request.RequireBool("full"); err == nil {
			full = v
		}

		req := debugger.StacktraceIn{
			Id:    goroutineID,
			Depth: depth,
			Full:  full,
		}
		if full {
			cfg := debugger.DefaultLoadConfig()
			req.Cfg = &cfg
		}
		var resp debugger.StacktraceOut
		if err := pool.Call("Stacktrace", req, &resp); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("RPC call failed: %v", err)), nil
		}

		return jsonResult(map[string]interface{}{
			"frames": resp.Locations,
		})
	}
}

func makeListGoroutines(pool *debugger.Pool) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		count := 100
		if v, err := request.RequireInt("count"); err == nil {
			count = v
		}

		req := debugger.ListGoroutinesIn{
			Start: 0,
			Count: count,
		}
		var resp debugger.ListGoroutinesOut
		if err := pool.Call("ListGoroutines", req, &resp); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("RPC call failed: %v", err)), nil
		}

		return jsonResult(map[string]interface{}{
			"goroutines": resp.Goroutines,
		})
	}
}
