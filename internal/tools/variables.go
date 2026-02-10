package tools

import (
	"context"
	"fmt"

	"github.com/kjbreil/dlc-sidecar/internal/debugger"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerVariables(s *server.MCPServer, pool *debugger.Pool) {
	// list_local_vars
	s.AddTool(mcp.NewTool("list_local_vars",
		mcp.WithDescription("List all local variables in the current scope"),
		mcp.WithNumber("goroutineID",
			mcp.Description("Goroutine ID to scope the request to (default: -1 for current)"),
		),
		mcp.WithNumber("frame",
			mcp.Description("Stack frame index (default: 0 for current frame)"),
		),
	), makeListLocalVars(pool))

	// list_function_args
	s.AddTool(mcp.NewTool("list_function_args",
		mcp.WithDescription("List all arguments of the current function"),
		mcp.WithNumber("goroutineID",
			mcp.Description("Goroutine ID to scope the request to (default: -1 for current)"),
		),
		mcp.WithNumber("frame",
			mcp.Description("Stack frame index (default: 0 for current frame)"),
		),
	), makeListFunctionArgs(pool))

	// eval
	s.AddTool(mcp.NewTool("eval",
		mcp.WithDescription("Evaluate an expression in the current scope and return the result"),
		mcp.WithString("expr",
			mcp.Required(),
			mcp.Description("Expression to evaluate (e.g. variable name, struct field, slice index)"),
		),
		mcp.WithNumber("goroutineID",
			mcp.Description("Goroutine ID to scope the request to (default: -1 for current)"),
		),
		mcp.WithNumber("frame",
			mcp.Description("Stack frame index (default: 0 for current frame)"),
		),
	), makeEval(pool))
}

func makeListLocalVars(pool *debugger.Pool) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		scope := evalScope(request)
		req := debugger.ListLocalVarsIn{
			Scope: scope,
			Cfg:   debugger.DefaultLoadConfig(),
		}
		var resp debugger.ListLocalVarsOut
		if err := pool.Call("ListLocalVars", req, &resp); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("RPC call failed: %v", err)), nil
		}

		return jsonResult(map[string]interface{}{
			"variables": resp.Variables,
		})
	}
}

func makeListFunctionArgs(pool *debugger.Pool) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		scope := evalScope(request)
		req := debugger.ListFunctionArgsIn{
			Scope: scope,
			Cfg:   debugger.DefaultLoadConfig(),
		}
		var resp debugger.ListFunctionArgsOut
		if err := pool.Call("ListFunctionArgs", req, &resp); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("RPC call failed: %v", err)), nil
		}

		return jsonResult(map[string]interface{}{
			"args": resp.Args,
		})
	}
}

func makeEval(pool *debugger.Pool) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		expr, err := request.RequireString("expr")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("expr parameter error: %v", err)), nil
		}

		scope := evalScope(request)
		cfg := debugger.DefaultLoadConfig()
		req := debugger.EvalIn{
			Scope: scope,
			Expr:  expr,
			Cfg:   &cfg,
		}
		var resp debugger.EvalOut
		if err := pool.Call("Eval", req, &resp); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("RPC call failed: %v", err)), nil
		}

		return jsonResult(map[string]interface{}{
			"variable": resp.Variable,
		})
	}
}

// evalScope extracts an EvalScope from optional request parameters.
func evalScope(request mcp.CallToolRequest) debugger.EvalScope {
	goroutineID := int64(-1)
	frame := 0
	if v, err := request.RequireInt("goroutineID"); err == nil {
		goroutineID = int64(v)
	}
	if v, err := request.RequireInt("frame"); err == nil {
		frame = v
	}
	return debugger.EvalScope{
		GoroutineID: goroutineID,
		Frame:       frame,
	}
}
