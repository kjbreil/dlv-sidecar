package tools

import (
	"context"
	"fmt"

	"github.com/kjbreil/dlc-sidecar/internal/debugger"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerExecution(s *server.MCPServer, pool *debugger.Pool) {
	// continue
	s.AddTool(mcp.NewTool("continue",
		mcp.WithDescription("Continue execution until the next breakpoint or program exit"),
	), makeCommand(pool, debugger.CmdContinue))

	// next (step over)
	s.AddTool(mcp.NewTool("next",
		mcp.WithDescription("Step to the next source line, stepping over function calls"),
	), makeCommand(pool, debugger.CmdNext))

	// step (step into)
	s.AddTool(mcp.NewTool("step",
		mcp.WithDescription("Step to the next source line, stepping into function calls"),
	), makeCommand(pool, debugger.CmdStep))

	// step_out
	s.AddTool(mcp.NewTool("step_out",
		mcp.WithDescription("Step out of the current function, continuing to the return address"),
	), makeCommand(pool, debugger.CmdStepOut))

	// step_instruction
	s.AddTool(mcp.NewTool("step_instruction",
		mcp.WithDescription("Step exactly one CPU instruction"),
	), makeCommand(pool, debugger.CmdStepInstruction))

	// halt
	s.AddTool(mcp.NewTool("halt",
		mcp.WithDescription("Halt the running program"),
	), makeCommand(pool, debugger.CmdHalt))
}

func makeCommand(pool *debugger.Pool, cmd string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		command := debugger.DebuggerCommand{Name: cmd}
		var resp debugger.CommandOut
		if err := pool.Call("Command", command, &resp); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("RPC call failed: %v", err)), nil
		}

		result := map[string]interface{}{
			"command": cmd,
		}
		st := resp.State
		if st.Exited {
			result["exited"] = true
			result["exitStatus"] = st.ExitStatus
		}
		if st.CurrentThread != nil {
			result["file"] = st.CurrentThread.File
			result["line"] = st.CurrentThread.Line
			if st.CurrentThread.Function != nil {
				result["function"] = st.CurrentThread.Function.Name
			}
		}
		if st.SelectedGoroutine != nil {
			result["goroutineID"] = st.SelectedGoroutine.ID
		}

		return jsonResult(result)
	}
}
