package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kjbreil/dlc-sidecar/internal/debugger"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerBreakpoints(s *server.MCPServer, addr string) {
	// set_breakpoint
	s.AddTool(mcp.NewTool("set_breakpoint",
		mcp.WithDescription("Set a breakpoint in the Delve debugger at the specified file and line"),
		mcp.WithString("file",
			mcp.Required(),
			mcp.Description("Absolute path to the source file"),
		),
		mcp.WithNumber("line",
			mcp.Required(),
			mcp.Description("Line number where the breakpoint should be set"),
		),
	), makeSetBreakpoint(addr))

	// clear_breakpoint
	s.AddTool(mcp.NewTool("clear_breakpoint",
		mcp.WithDescription("Clear a breakpoint by its ID"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("ID of the breakpoint to clear"),
		),
	), makeClearBreakpoint(addr))

	// list_breakpoints
	s.AddTool(mcp.NewTool("list_breakpoints",
		mcp.WithDescription("List all breakpoints currently set in the debugger"),
	), makeListBreakpoints(addr))
}

func makeSetBreakpoint(addr string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file, err := request.RequireString("file")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("file parameter error: %v", err)), nil
		}
		line, err := request.RequireInt("line")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("line parameter error: %v", err)), nil
		}

		c, err := dial(addr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to connect to Delve at %s: %v", addr, err)), nil
		}
		defer c.Close()

		req := debugger.CreateBreakpointIn{
			Breakpoint: debugger.Breakpoint{
				File: file,
				Line: line,
			},
		}
		var resp debugger.CreateBreakpointOut
		if err := c.Call("CreateBreakpoint", req, &resp); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("RPC call failed: %v", err)), nil
		}

		return jsonResult(map[string]interface{}{
			"success":    true,
			"breakpoint": resp.Breakpoint,
			"message":    fmt.Sprintf("Breakpoint set at %s:%d (ID: %d)", file, line, resp.Breakpoint.ID),
		})
	}
}

func makeClearBreakpoint(addr string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireInt("id")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("id parameter error: %v", err)), nil
		}

		c, err := dial(addr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to connect to Delve at %s: %v", addr, err)), nil
		}
		defer c.Close()

		req := debugger.ClearBreakpointIn{Id: id}
		var resp debugger.ClearBreakpointOut
		if err := c.Call("ClearBreakpoint", req, &resp); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("RPC call failed: %v", err)), nil
		}

		return jsonResult(map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Breakpoint %d cleared", id),
		})
	}
}

func makeListBreakpoints(addr string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, err := dial(addr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to connect to Delve at %s: %v", addr, err)), nil
		}
		defer c.Close()

		req := debugger.ListBreakpointsIn{All: true}
		var resp debugger.ListBreakpointsOut
		if err := c.Call("ListBreakpoints", req, &resp); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("RPC call failed: %v", err)), nil
		}

		return jsonResult(map[string]interface{}{
			"breakpoints": resp.Breakpoints,
		})
	}
}

// jsonResult marshals v to JSON and returns it as a tool result.
func jsonResult(v interface{}) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}
