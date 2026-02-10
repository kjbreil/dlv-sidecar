# dlc-sidecar

MCP server to control a Delve debugger instance over JSON-RPC API.

## Overview

This is a Model Context Protocol (MCP) server that provides a `set_breakpoint` tool for setting breakpoints in a running Delve debugger session. It communicates with Delve via JSON-RPC over TCP at `localhost:2345`.

## Building

```bash
go build -o dlc-sidecar .
```

## Usage

The server uses stdio transport for MCP communication. To use it with an MCP client:

1. Start your Go application with Delve in headless mode:
   ```bash
   dlv debug --headless --listen=localhost:2345 --api-version=2 ./your-app
   ```

2. Run the MCP server:
   ```bash
   ./dlc-sidecar
   ```

3. The MCP server will expose a `set_breakpoint` tool with the following parameters:
   - `file` (string, required): Absolute path to the source file
   - `line` (integer, required): Line number where the breakpoint should be set

## Tool: set_breakpoint

Sets a breakpoint in the Delve debugger at the specified file and line.

**Parameters:**
- `file`: Absolute path to the source file (e.g., `/path/to/your/main.go`)
- `line`: Line number (e.g., `42`)

**Returns:**
```json
{
  "success": true,
  "breakpoint": {
    "id": 1,
    "file": "/path/to/your/main.go",
    "line": 42
  },
  "message": "Breakpoint set at /path/to/your/main.go:42 (ID: 1)"
}
```

## Requirements

- Go 1.24 or later
- Running Delve debugger instance at `localhost:2345`

## Dependencies

- [github.com/mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - MCP server implementation

