# CRI-O MCP Server

This project provides a minimal skeleton for building a Machine Config Pool (MCP) server for CRI-O deployments.

## Structure
- `cmd/mcp-server`: entry point for running the server
- `pkg/api`: APIs exposed by the server
- `pkg/cri`: utilities for interacting with the CRI implementation

## Getting Started
Initialize dependencies and build the binary:

```bash
go build ./cmd/mcp-server
```

This repository currently contains placeholder packages and is ready for further development.
