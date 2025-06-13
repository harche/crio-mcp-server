package sdkserver

import (
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// New creates a new MCP server using the mcp-go SDK.
// configPath is currently unused but will be wired into
// tool and resource handlers in future implementations.
func New(configPath string) *mcpserver.MCPServer {
	srv := mcpserver.NewMCPServer(
		"CRI-O MCP Server",
		"0.1.0",
		mcpserver.WithToolCapabilities(false),
	)

	// Register tools
	srv.AddTools(
		mcpserver.ServerTool{Tool: debugNodeTool, Handler: handleDebugNode},
		mcpserver.ServerTool{Tool: nodeLogsTool, Handler: handleNodeLogs},
	)

	_ = configPath

	return srv
}

// StartStdio starts the provided MCP server using the
// stdio transport. It blocks until the server exits.
func StartStdio(s *mcpserver.MCPServer) error {
	return mcpserver.ServeStdio(s)
}
