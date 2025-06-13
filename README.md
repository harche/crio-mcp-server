# CRI-O MCP Server

This repository contains a minimal [Model Context Protocol](https://github.com/mark3labs/mcp-go) (MCP) server that exposes CRI-O information through a set of command line tools.  The server is intended to be run from VS Code using the MCP extension and communicates over standard input/output.

## Tools

### `debug_node`
Creates a temporary debug pod on a specified OpenShift node and runs arbitrary commands inside it.

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

## Running inside VS Code

1. Open this repository in VS Code with the MCP extension enabled.
2. Build the server:
   ```bash
   go build ./cmd/mcp-server
   ```
3. Start the binary from the integrated terminal:
   ```bash
   ./mcp-server
   ```
   Because it uses the stdio transport, the MCP extension will automatically attach.
4. Use the command palette or MCP view to invoke `debug_node` or `collect_node_logs` and supply the required parameters.

This implementation focuses solely on the MCP server and the tools above. The legacy gRPC server and old README have been removed.
