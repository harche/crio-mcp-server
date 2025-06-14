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
  The zipped data is returned as a blob resource named `node-logs-<node>.txt.gz`.

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

### `run_crictl`
Runs `crictl` inside a debug pod to interact directly with the node's container runtime. Use the `-h` flag on any subcommand for help.

crictl is the lightweight command-line client from the cri-tools project that
speaks the Kubernetes Container Runtime Interface (CRI) directly.  Because it
talks to the node’s container runtime (CRI-O, containerd, etc.) over the local
/var/run/<runtime>.sock, it works even when kubelet or the API server are
unhealthy.  Common commands include `crictl ps` and `crictl pods` to list
running containers or sandboxes, `crictl inspect`/`inspectp` for JSON-formatted
metadata, `crictl logs` to read container stdout, `crictl exec` for a shell,
`crictl images` and `crictl pull` to manage images, `crictl stats` for live CPU
and memory usage, and `crictl runp|create|start` to launch test sandboxes.
Because it bypasses Kubernetes control-plane layers, crictl is the first tool
engineers reach for when debugging low-level runtime or cgroup issues on an
OpenShift node.

Arguments:
- `node_name` (string, required) – node on which to run the command
- `args` (array of string) – arguments forwarded to `crictl` (defaults to `ps`)

### `traverse_cgroupfs`
Drops a debug pod onto a node and walks its unified cgroup-v2 hierarchy. By default it lists `memory.current` for every pod under `/sys/fs/cgroup/kubepods.slice`, but you can supply custom commands to inspect other files.

Cgroup files are the ground truth for how the Linux kernel enforces every pod’s CPU, memory, I/O and PIDs limits. Reading `cpu.max`, `memory.max`, `io.stat`, `pids.max` or pressure-stall metrics straight from `/sys/fs/cgroup/kubepods.slice/...` lets you verify that the values the kubelet intended actually reached the kernel; spot runaway memory or CPU throttling even when metrics-server is down; correlate CRI-O OOM-kills with mis-configured requests; and confirm that topology-aware features like CPU Manager wrote the right `cpuset.cpus` mask.

Arguments:
- `node_name` (string, required) – node whose cgroupfs should be inspected
- `commands` (array of string) – optional shell commands to run inside the debug pod

### `gather_network_logs`
Runs the `gather_network_logs` must-gather addon to capture iptables and OVN flows along with CNI pod logs.

Arguments:
- `dest_dir` (string) – directory where the network logs are stored

### `gather_profiling_node`
Collects 30-second CPU and heap profiles from kubelet and CRI-O using the `gather_profiling_node` script.

Arguments:
- `dest_dir` (string) – directory where the profiling output is written

### `collect_events`
Fetches recent Kubernetes events from all namespaces via `oc get events -A`.

### `collect_pod_logs`
Retrieves logs from a specific pod similar to `oc logs`.

Arguments:
- `namespace` (string, required) – namespace of the pod
- `pod_name` (string, required) – pod to read logs from
- `container` (string) – optional container within the pod
- `since` (string) – optional duration (e.g. `5m`) to limit logs

### `collect_node_config`
Uses `oc debug` to print kubelet and CRI-O configuration files from the node.

Arguments:
- `node_name` (string, required) – node to inspect

