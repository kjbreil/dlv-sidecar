package tools

import (
	"context"
	"fmt"

	"github.com/kjbreil/dlc-sidecar/internal/debugger"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerExecution(s *server.MCPServer, addr string) {
	// continue
	s.AddTool(mcp.NewTool("continue",
		mcp.WithDescription("Continue execution until the next breakpoint or program exit"),
	), makeCommand(addr, debugger.CmdContinue))

	// next (step over)
	s.AddTool(mcp.NewTool("next",
		mcp.WithDescription("Step to the next source line, stepping over function calls"),
	), makeCommand(addr, debugger.CmdNext))

	// step (step into)
	s.AddTool(mcp.NewTool("step",
		mcp.WithDescription("Step to the next source line, stepping into function calls"),
	), makeCommand(addr, debugger.CmdStep))

	// step_out
	s.AddTool(mcp.NewTool("step_out",
		mcp.WithDescription("Step out of the current function, continuing to the return address"),
	), makeCommand(addr, debugger.CmdStepOut))

	// step_instruction
	s.AddTool(mcp.NewTool("step_instruction",
		mcp.WithDescription("Step exactly one CPU instruction"),
	), makeCommand(addr, debugger.CmdStepInstruction))

	// halt
	s.AddTool(mcp.NewTool("halt",
		mcp.WithDescription("Halt the running program"),
	), makeCommand(addr, debugger.CmdHalt))
}

func makeCommand(addr string, cmd string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, err := dial(addr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to connect to Delve at %s: %v", addr, err)), nil
		}
		defer c.Close()

		command := debugger.DebuggerCommand{Name: cmd}
		var resp debugger.CommandOut
		if err := c.Call("Command", command, &resp); err != nil {
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
