# CRI-O MCP Tools

This repository defines a couple of [Model Context Protocol](https://github.com/mark3labs/mcp-go) tools for debugging OpenShift nodes and collecting logs. The tools are implemented in `pkg/sdkserver/tools.go` and can be registered with any MCP server built using the `mcp-go` SDK.

## Tools

### `debug_node`
Runs `oc debug` on a specified node and executes arbitrary shell commands inside the temporary debug pod.

Arguments:
- `node_name` (string, required) – node to debug
- `commands` (array of string) – commands executed in the pod (defaults to `journalctl --no-pager -u crio`)
- `collect_files` (bool) – when true, files listed in `paths` are returned as resources
- `paths` (array of string) – file or directory paths to copy from the host

### `collect_node_logs`
Streams systemd journal and container runtime logs from a node using `oc adm node-logs`.

Arguments:
- `node_name` (string, required) – target node
- `since` (string) – RFC3339 timestamp or relative value accepted by `journalctl`
- `compress` (bool) – if true, return logs as a gzip tarball instead of inline text

### `analyze_pprof`
Runs `go tool pprof` with the supplied arguments to inspect CPU or memory profiles. Refer to `go tool pprof -h` for the full set of options.

Arguments:
- `args` (array of string, required) – command-line arguments passed directly to `go tool pprof`

### `collect_must_gather`
Runs `oc adm must-gather` to capture cluster information. Create a temporary directory and pass it using `dest_dir` to keep all gathered data in one location. Explore `oc adm must-gather -h` for the full set of options.

oc adm must-gather can scoop up almost every artifact engineers or support need in a single shot: it exports the full YAML for all cluster-scoped and namespaced resources (Deployments, CRDs, Nodes, ClusterOperators, etc.); captures pod and container logs as well as systemd journal slices from each node to trace runtime crashes or OOMs; grabs API-server and OAuth audit logs for security or compliance forensics; collects kernel, cgroup, and other node sysinfo plus tuned and kubelet configs for performance tuning; optionally runs add-on scripts such as gather_network_logs to archive iptables/OVN flows and CNI pod logs, or gather_profiling_node to fetch 30-second CPU and heap pprof dumps from both kubelet and CRI-O for hotspot analysis; and, through plug-in images, can extend to operator-specific data like storage states or virtualization metrics, ensuring one reproducible tarball contains configuration, logs, network traces, performance profiles, and security audits for thorough offline debugging.

Arguments:
- `dest_dir` (string) – local directory where the must-gather output is stored
- `extra_args` (array of string) – additional flags forwarded to `oc adm must-gather`

These helpers can be integrated into a custom MCP server or used directly with the `mcp-go` SDK.

### `collect_sosreport`
Runs `sosreport` inside a debug pod using toolbox. This captures detailed diagnostics from a node. Provide a Red Hat case ID if available.

Arguments:
- `node_name` (string, required) – node from which to gather the report
- `case_id` (string) – optional support case identifier passed to `sosreport`

