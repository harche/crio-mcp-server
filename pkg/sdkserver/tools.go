package sdkserver

import (
	"bytes"
	"context"
	"fmt"

	"github.com/harche/crio-mcp-server/pkg/openshift"
	mcp "github.com/mark3labs/mcp-go/mcp"
)

// debugNodeTool defines the debug_node MCP tool.
var debugNodeTool = mcp.NewTool(
	"debug_node",
	mcp.WithTitleAnnotation("Run oc debug on a node"),
	mcp.WithDescription("Creates an OpenShift debug pod on the specified node, runs arbitrary commands, and returns the output."),
	mcp.WithString("node_name",
		mcp.Description("Name of the node to debug (e.g. ip-10-0-1-5.ec2.internal)"),
		mcp.Required(),
	),
	mcp.WithArray("commands",
		mcp.Description("List of shell commands to execute inside the debug pod (default: ['journalctl --no-pager -u crio'])"),
		mcp.Items(map[string]any{"type": "string"}),
	),
	mcp.WithBoolean("collect_files",
		mcp.Description("If true, copy the specified paths back and return them as resources"),
		mcp.DefaultBool(false),
	),
	mcp.WithArray("paths",
		mcp.Description("File or directory paths on the host to retrieve (requires collect_files=true)"),
		mcp.Items(map[string]any{"type": "string"}),
	),
)

// nodeLogsTool defines the collect_node_logs MCP tool.
var nodeLogsTool = mcp.NewTool(
	"collect_node_logs",
	mcp.WithTitleAnnotation("Collect node logs via oc adm node-logs"),
	mcp.WithDescription("Streams systemd journal and container runtime logs from a given node using oc adm node-logs."),
	mcp.WithString("node_name",
		mcp.Description("Target node"),
		mcp.Required(),
	),
	mcp.WithString("since",
		mcp.Description("RFC3339 timestamp or relative value understood by journalctl/oc (e.g. '2h')"),
	),
	mcp.WithBoolean("compress",
		mcp.Description("If true, return logs as a gzip tarball resource instead of inline text"),
		mcp.DefaultBool(false),
	),
)

// handleDebugNode executes oc debug with the provided arguments.
func handleDebugNode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := req.RequireString("node_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	commands, _ := req.GetArguments()["commands"].([]any)
	if len(commands) == 0 {
		commands = []any{"journalctl --no-pager -u crio"}
	}
	var output bytes.Buffer
	for _, cmd := range commands {
		out, err := openshift.DebugNode(nodeName, fmt.Sprint(cmd))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		output.WriteString(out)
	}
	return mcp.NewToolResultText(output.String()), nil
}

// handleNodeLogs collects logs from a node using oc adm node-logs.
func handleNodeLogs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := req.RequireString("node_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	since := req.GetString("since", "")
	out, err := openshift.NodeLogs(nodeName, since)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}
