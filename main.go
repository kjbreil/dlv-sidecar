package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc/jsonrpc"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	delveAddr = "localhost:2345"
	toolName  = "set_breakpoint"
)

// CreateBreakpointRequest matches Delve's RPC request structure
type CreateBreakpointRequest struct {
	Breakpoint Breakpoint
}

// Breakpoint represents a Delve breakpoint
type Breakpoint struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

// CreateBreakpointResponse matches Delve's RPC response structure
type CreateBreakpointResponse struct {
	Breakpoint BreakpointInfo `json:"Breakpoint"`
}

// BreakpointInfo contains information about a created breakpoint
type BreakpointInfo struct {
	ID   int    `json:"id"`
	File string `json:"file"`
	Line int    `json:"line"`
}

// setBreakpoint handles the set_breakpoint tool call
func setBreakpoint(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract and validate arguments
	file, err := request.RequireString("file")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("file parameter error: %v", err)), nil
	}

	line, err := request.RequireInt("line")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("line parameter error: %v", err)), nil
	}

	// Connect to Delve debugger
	conn, err := net.Dial("tcp", delveAddr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to connect to Delve at %s: %v", delveAddr, err)), nil
	}
	defer conn.Close()

	// Create JSON-RPC client
	client := jsonrpc.NewClient(conn)
	defer client.Close()

	// Prepare the request
	req := CreateBreakpointRequest{
		Breakpoint: Breakpoint{
			File: file,
			Line: line,
		},
	}

	// Call the RPC method
	var resp CreateBreakpointResponse
	err = client.Call("RPCServer.CreateBreakpoint", req, &resp)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("RPC call failed: %v", err)), nil
	}

	// Return success result
	resultJSON, err := json.Marshal(map[string]interface{}{
		"success":     true,
		"breakpoint":  resp.Breakpoint,
		"message":     fmt.Sprintf("Breakpoint set at %s:%d (ID: %d)", file, line, resp.Breakpoint.ID),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"dlc-sidecar",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Define the set_breakpoint tool
	tool := mcp.NewTool(toolName,
		mcp.WithDescription("Set a breakpoint in the Delve debugger at the specified file and line"),
		mcp.WithString("file",
			mcp.Required(),
			mcp.Description("Absolute path to the source file"),
		),
		mcp.WithNumber("line",
			mcp.Required(),
			mcp.Description("Line number where the breakpoint should be set"),
		),
	)

	// Register the tool
	s.AddTool(tool, setBreakpoint)

	// Start the server with stdio transport
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
