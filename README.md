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

These helpers can be integrated into a custom MCP server or used directly with the `mcp-go` SDK.
