# Ralph Loop

This is the configuration file for the autonomous sweeping capabilities of the Ralph loop.

## Verification Gates
- **Shell Lint**: `make lint` ensures `bash -n`, `shellcheck`, and `jq` parsing pass.
- **Go Tests**: Tests across `mcp/` directories pass (`go test ./...` in the respective dirs).

## Iteration Configuration
- Max Iterations: 10
- Autonomy Level: 1 (safe mode)
- Tranche Mode: ON

## Operational Directives
When participating in a controlled autonomous sweep, the agent must adhere to the core instruction set located in `AGENTS.md`. No new configuration patterns or non-standard tools should be introduced without explicit approval.
