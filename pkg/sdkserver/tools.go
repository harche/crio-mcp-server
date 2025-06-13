package sdkserver

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/harche/crio-mcp-server/pkg/openshift"
	mcp "github.com/mark3labs/mcp-go/mcp"
)

// debugNodeTool defines the debug_node MCP tool.
var debugNodeTool = mcp.NewTool(
	"debug_node",
	mcp.WithTitleAnnotation("Run oc debug on a node"),
	mcp.WithDescription(`Creates an OpenShift debug pod on the specified node, runs arbitrary commands, and returns the output.

Use "oc -h" to discover the full set of commands and flags provided by the OpenShift CLI. When unsure about a subcommand, recursively run "oc <command> -h" to inspect available options.`),
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

// pprofTool defines the analyze_pprof MCP tool.
var pprofTool = mcp.NewTool(
	"analyze_pprof",
	mcp.WithTitleAnnotation("Analyze Go profiles via go tool pprof"),
	mcp.WithDescription(`Runs "go tool pprof" with the provided arguments to analyze CPU or memory profiles. Use "go tool pprof -h" for available options.`),
	mcp.WithArray("args",
		mcp.Description("Arguments passed directly to 'go tool pprof', e.g. ['-top', './bin', 'profile.pb.gz']"),
		mcp.Items(map[string]any{"type": "string"}),
		mcp.Required(),
	),
)

// mustGatherTool defines the collect_must_gather MCP tool.
var mustGatherTool = mcp.NewTool(
	"collect_must_gather",
	mcp.WithTitleAnnotation("Collect cluster data via oc adm must-gather"),
	mcp.WithDescription(`Runs "oc adm must-gather" to capture debugging information.

Create a temporary directory and pass it using the --dest-dir option to store the output in a single place.

oc adm must-gather can scoop up almost every artifact engineers or support need in a single shot: it exports the full YAML for all cluster-scoped and namespaced resources (Deployments, CRDs, Nodes, ClusterOperators, etc.); captures pod and container logs as well as systemd journal slices from each node to trace runtime crashes or OOMs; grabs API-server and OAuth audit logs for security or compliance forensics; collects kernel, cgroup, and other node sysinfo plus tuned and kubelet configs for performance tuning; optionally runs add-on scripts such as gather_network_logs to archive iptables/OVN flows and CNI pod logs, or gather_profiling_node to fetch 30-second CPU and heap pprof dumps from both kubelet and CRI-O for hotspot analysis; and, through plug-in images, can extend to operator-specific data like storage states or virtualization metrics, ensuring one reproducible tarball contains configuration, logs, network traces, performance profiles, and security audits for thorough offline debugging.`),
	mcp.WithString("dest_dir",
		mcp.Description("Directory to write gathered data"),
	),
	mcp.WithArray("extra_args",
		mcp.Description("Additional arguments passed directly to oc adm must-gather"),
		mcp.Items(map[string]any{"type": "string"}),
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

// handlePprof runs "go tool pprof" with the provided arguments and returns the output.
func handlePprof(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsAny, _ := req.GetArguments()["args"].([]any)
	if len(argsAny) == 0 {
		return mcp.NewToolResultError("args is required"), nil
	}
	args := make([]string, len(argsAny))
	for i, a := range argsAny {
		args[i] = fmt.Sprint(a)
	}
	cmd := exec.CommandContext(ctx, "go", append([]string{"tool", "pprof"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return mcp.NewToolResultError(fmt.Errorf("pprof failed: %w: %s", err, out).Error()), nil
	}
	return mcp.NewToolResultText(string(out)), nil
}

// handleMustGather executes oc adm must-gather with the provided arguments.
func handleMustGather(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dest := req.GetString("dest_dir", "")
	extraAny, _ := req.GetArguments()["extra_args"].([]any)
	extras := make([]string, len(extraAny))
	for i, a := range extraAny {
		extras[i] = fmt.Sprint(a)
	}
	out, err := openshift.MustGather(dest, extras)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

// sosReportTool defines the collect_sosreport MCP tool.
var sosReportTool = mcp.NewTool(
	"collect_sosreport",
	mcp.WithTitleAnnotation("Collect sosreport from a node"),
	mcp.WithDescription(`Runs "sosreport" in a debug pod via toolbox to collect node diagnostics.`),
	mcp.WithString("node_name",
		mcp.Description("Node to collect sosreport from"),
		mcp.Required(),
	),
	mcp.WithString("case_id",
		mcp.Description("Optional Red Hat support case ID"),
	),
)

// handleSosReport executes sosreport on the target node using toolbox.
func handleSosReport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := req.RequireString("node_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	caseID := req.GetString("case_id", "")
	out, err := openshift.SosReport(nodeName, caseID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}
