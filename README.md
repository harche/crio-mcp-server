# CRI-O MCP Server

This project implements a minimal gRPC server for serving CRI-O configuration to Machine Config Pool (MCP) clients.

## Structure
- `cmd/mcp-server`: entry point for running the server
- `pkg/proto`: protobuf definitions and generated code
- `pkg/server`: gRPC server implementation
- `pkg/api`: placeholder APIs
- `pkg/cri`: utilities for interacting with the CRI implementation

## Getting Started
Initialize dependencies and build the binary:

```bash
go build ./cmd/mcp-server
```

Run the server (listens on `:50051` by default):

```bash
./mcp-server --config /etc/crio/crio.conf
```

The server exposes a `GetCrioConfig` RPC that returns the contents of the CRI-O configuration file.
