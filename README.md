# CRI-O MCP Server

This project provides a basic Model Context Protocol (MCP) server for CRI-O built using the [mcp-go](https://github.com/mark3labs/mcp-go) SDK. The goal is to expose CRI-O information to AI agents through the MCP interface.

## Structure
- `cmd/mcp-server`: entry point for running the server
- `pkg/sdkserver`: MCP server scaffolding based on the SDK
- `pkg/proto`: protobuf definitions from the original gRPC implementation
- `pkg/server`: legacy gRPC server (will be replaced)
- `pkg/api`: placeholder APIs
- `pkg/cri`: utilities for interacting with the CRI implementation

## Getting Started
Initialize dependencies and build the binary:

```bash
go build ./cmd/mcp-server
```

Run the server using the stdio transport:

```bash
./mcp-server --config /etc/crio/crio.conf
```

The current implementation only sets up the server scaffolding. Additional tools and resources will be added in future iterations.
