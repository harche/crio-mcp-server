package sdkserver

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"

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

oc adm must-gather can scoop up almost every artifact engineers or support need in a single shot: it exports the full YAML for all cluster-scoped and namespaced resources (Deployments, CRDs, Nodes, ClusterOperators, etc.); captures pod and container logs as well as systemd journal slices from each node to trace runtime crashes or OOMs; grabs API-server and OAuth audit logs for security or compliance forensics; collects kernel, cgroup, and other node sysinfo plus tuned and kubelet configs for performance tuning; optionally runs add-on scripts such as gather_network_logs to archive iptables/OVN flows and CNI pod logs, or gather_profiling_node to fetch 30-second CPU and heap pprof dumps from both kubelet and CRI-O for hotspot analysis; and, through plug-in images, can extend to operator-specific data like storage states or virtualization metrics, ensuring one reproducible tarball contains configuration, logs, network traces, performance profiles, and security audits for thorough offline debugging. Use "oc adm must-gather -h" for available options.`),
	mcp.WithString("dest_dir",
		mcp.Description("Directory to write gathered data"),
	),
	mcp.WithArray("extra_args",
		mcp.Description("Additional arguments passed directly to oc adm must-gather"),
		mcp.Items(map[string]any{"type": "string"}),
	),
)

// crictlTool defines the run_crictl MCP tool.
var crictlTool = mcp.NewTool(
	"run_crictl",
	mcp.WithTitleAnnotation("Run crictl on a node"),
	mcp.WithDescription(`Executes "crictl" inside an OpenShift debug pod on the specified node.

crictl is the lightweight command-line client from the cri-tools project that speaks the Kubernetes Container Runtime Interface (CRI) directly. Because it talks to the node's container runtime (CRI-O, containerd, etc.) over the local /var/run/<runtime>.sock, it works even when kubelet or the API server are unhealthy. Common commands include "crictl ps" and "crictl pods" to list running containers or sandboxes, "crictl inspect"/"inspectp" for JSON-formatted metadata, "crictl logs" to read container stdout, "crictl exec" for a shell, "crictl images" and "crictl pull" to manage images, "crictl stats" for live CPU and memory usage, and "crictl runp|create|start" to launch test sandboxes. Because it bypasses Kubernetes control-plane layers, crictl is the first tool engineers reach for when debugging low-level runtime or cgroup issues on an OpenShift node.

Use the -h flag to discover available subcommands.`),
	mcp.WithString("node_name",
		mcp.Description("Node on which to run crictl"),
		mcp.Required(),
	),
	mcp.WithArray("args",
		mcp.Description("Arguments passed directly to crictl (default: ['ps'])"),
		mcp.Items(map[string]any{"type": "string"}),
	),
)

// cgroupfsTool defines the traverse_cgroupfs MCP tool.
var cgroupfsTool = mcp.NewTool(
	"traverse_cgroupfs",
	mcp.WithTitleAnnotation("Traverse cgroupfs on a node"),
	mcp.WithDescription(`Drops an oc debug pod onto the node, chroots into the host rootfs and walks the unified cgroup-v2 hierarchy under /sys/fs/cgroup/kubepods.slice.

Cgroup files are the ground truth for how the Linux kernel enforces every pod's CPU, memory, I/O and PIDs limits. Reading cpu.max, memory.max, io.stat, pids.max or pressure-stall metrics straight from /sys/fs/cgroup/kubepods.slice/... lets you verify that the values the kubelet intended actually reached the kernel; spot runaway memory or CPU throttling even when metrics-server is down; correlate CRI-O OOM-kills with misconfigured requests; and confirm that topology-aware features like CPU Manager wrote the right cpuset.cpus mask.`),
	mcp.WithString("node_name",
		mcp.Description("Node whose cgroupfs should be inspected"),
		mcp.Required(),
	),
	mcp.WithArray("commands",
		mcp.Description("Shell commands executed inside the debug pod (default: list memory.current for all pods)"),
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
	compressLogs := req.GetBool("compress", false)
	out, err := openshift.NodeLogs(nodeName, since)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !compressLogs {
		return mcp.NewToolResultText(out), nil
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(out)); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := gz.Close(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	res := mcp.BlobResourceContents{
		URI:      fmt.Sprintf("node-logs-%s.txt.gz", nodeName),
		MIMEType: "application/gzip",
		Blob:     encoded,
	}
	return mcp.NewToolResultResource("Node logs compressed", res), nil
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

// handleCrictl runs crictl commands on a node via oc debug.
func handleCrictl(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := req.RequireString("node_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	argsAny, _ := req.GetArguments()["args"].([]any)
	args := make([]string, len(argsAny))
	for i, a := range argsAny {
		args[i] = fmt.Sprint(a)
	}
	if len(args) == 0 {
		args = []string{"ps"}
	}
	out, err := openshift.Crictl(nodeName, args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

// handleTraverseCgroupfs walks the cgroup hierarchy on a node via oc debug.
func handleTraverseCgroupfs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := req.RequireString("node_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	cmdsAny, _ := req.GetArguments()["commands"].([]any)
	var script string
	if len(cmdsAny) == 0 {
		script = "find /sys/fs/cgroup/kubepods.slice -name memory.current | xargs grep -H ."
	} else {
		cmds := make([]string, len(cmdsAny))
		for i, c := range cmdsAny {
			cmds[i] = fmt.Sprint(c)
		}
		script = strings.Join(cmds, " && ")
	}
	out, err := openshift.DebugNode(nodeName, script)
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

// networkLogsTool defines the gather_network_logs MCP tool.
var networkLogsTool = mcp.NewTool(
	"gather_network_logs",
	mcp.WithTitleAnnotation("Collect network diagnostics via gather_network_logs"),
	mcp.WithDescription(`Runs the gather_network_logs must-gather addon to capture iptables, OVN flows and CNI pod logs.`),
	mcp.WithString("dest_dir",
		mcp.Description("Directory to store captured logs"),
	),
)

// profilingTool defines the gather_profiling_node MCP tool.
var profilingTool = mcp.NewTool(
	"gather_profiling_node",
	mcp.WithTitleAnnotation("Collect kubelet and CRI-O profiles"),
	mcp.WithDescription(`Runs gather_profiling_node to grab 30 second CPU and heap profiles from kubelet and CRI-O.`),
	mcp.WithString("dest_dir",
		mcp.Description("Directory to store profiling data"),
	),
)

// eventsTool defines the collect_events MCP tool.
var eventsTool = mcp.NewTool(
	"collect_events",
	mcp.WithTitleAnnotation("Retrieve recent cluster events"),
	mcp.WithDescription("Runs 'oc get events -A' to capture warnings and failures across all namespaces."),
)

// podLogsTool defines the collect_pod_logs MCP tool.
var podLogsTool = mcp.NewTool(
	"collect_pod_logs",
	mcp.WithTitleAnnotation("Collect logs from a specific pod"),
	mcp.WithDescription("Wraps 'oc logs' to fetch logs from a pod/container."),
	mcp.WithString("namespace",
		mcp.Description("Namespace of the pod"),
		mcp.Required(),
	),
	mcp.WithString("pod_name",
		mcp.Description("Name of the pod"),
		mcp.Required(),
	),
	mcp.WithString("container",
		mcp.Description("Container name inside the pod"),
	),
	mcp.WithString("since",
		mcp.Description("Only return logs newer than a relative duration like 5m"),
	),
)

// nodeConfigTool defines the collect_node_config MCP tool.
var nodeConfigTool = mcp.NewTool(
	"collect_node_config",
	mcp.WithTitleAnnotation("Gather kubelet and CRI-O configuration"),
	mcp.WithDescription("Uses oc debug to print kubelet.conf and crio.conf from the node."),
	mcp.WithString("node_name",
		mcp.Description("Node to inspect"),
		mcp.Required(),
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

// handleNetworkLogs runs gather_network_logs via oc adm must-gather.
func handleNetworkLogs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dest := req.GetString("dest_dir", "")
	out, err := openshift.NetworkLogs(dest)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

// handleProfilingNode collects kubelet and CRI-O profiles using gather_profiling_node.
func handleProfilingNode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dest := req.GetString("dest_dir", "")
	out, err := openshift.ProfilingNode(dest)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

// handleEvents fetches recent cluster events.
func handleEvents(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	out, err := openshift.Events()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

// handlePodLogs retrieves logs from the specified pod and container.
func handlePodLogs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns, err := req.RequireString("namespace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	pod, err := req.RequireString("pod_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	container := req.GetString("container", "")
	since := req.GetString("since", "")
	out, err := openshift.PodLogs(ns, pod, container, since)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

// handleNodeConfig collects kubelet and CRI-O configuration from a node.
func handleNodeConfig(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := req.RequireString("node_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	out, err := openshift.NodeConfig(nodeName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}
