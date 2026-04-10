# dotfiles-mcp — Gemini CLI Instructions

## Overview
Go MCP server for dotfiles management — config editing, symlink checks, service reloading via stdio transport.

## Build & Test
```bash
GOWORK=off go build -o dotfiles-mcp .
GOWORK=off go test ./...
```

## Key Details
- Single-file server: `main.go`
- SDK: mcp-go (github.com/mark3labs/mcp-go)
- Env: `DOTFILES_DIR` sets dotfiles path

## Shared Research Repository

Cross-project research lives at `~/hairglasses-studio/docs/` (git: hairglasses-studio/docs). When launching research agents, check existing docs first and write reusable research outputs back to the shared repo rather than local docs/.
