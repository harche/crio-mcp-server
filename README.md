# CRI-O MCP Server

This project implements a minimal gRPC server that speaks the **Model Context Protocol (MCP)** for CRI-O. It allows AI agents to query CRI-O and retrieve debugging information such as configuration, runtime status and container logs.

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

The server exposes several RPCs, including `GetCrioConfig`, to retrieve CRI-O configuration and other runtime information.
