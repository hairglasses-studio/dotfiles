# Contributing to mapitall

Thank you for your interest in contributing to mapitall, the cross-platform Go
event daemon for real-time controller mapping. Mapitall captures device events
(gamepads, MIDI controllers, mice) and dispatches them through a rule-resolution
pipeline to 11 output types including key presses, commands, OSC, MIDI, D-Bus,
and WebSocket.

**Before you start:**

- Browse [GitHub Issues](https://github.com/hairglasses-studio/mapitall/issues)
  for existing work and discussion.
- Read the project [CLAUDE.md](CLAUDE.md) for architecture details and the full
  output/input provider matrix.

## Development Setup

### Prerequisites

- **Go 1.26.1** or later
- **git**
- **make** (for Makefile targets)
- **Linux** recommended for full testing (evdev, uinput, ALSA MIDI, D-Bus).
  macOS and Windows builds compile but require platform-specific hardware for
  integration testing.

### Clone and Build

```bash
git clone https://github.com/hairglasses-studio/mapitall.git
cd mapitall
make build      # or: go build ./cmd/mapitall
```

### Run Tests

```bash
make test       # or: go test ./... -count=1
```

### Run Tests with Race Detection

```bash
go test ./... -count=1 -race
```

### Lint

```bash
go vet ./...
golangci-lint run ./...   # if installed
```

### Full Pipeline

```bash
make check      # build + vet + test
```

## Pull Request Workflow

1. **Fork** the repository and clone your fork.
2. **Create a branch** from `main` using the naming convention below.
3. **Make your changes** with tests.
4. **Run the full pipeline locally:**
   ```bash
   make build && go vet ./... && go test ./... -count=1 -race
   ```
5. **Commit** with conventional commit messages (see below).
6. **Push** to your fork and open a PR against `main`.
7. **Fill out the PR description** covering what changed, why, and how you tested it.
8. **Wait for CI** and maintainer review.

Keep PRs focused. One logical change per PR is easier to review than a combined
refactor-plus-feature.

### Branch Naming

Use `type/short-description`:

| Prefix | Use for |
|--------|---------|
| `feat/` | New features |
| `fix/` | Bug fixes |
| `docs/` | Documentation changes |
| `test/` | Test additions or improvements |
| `refactor/` | Code restructuring without behavior changes |
| `chore/` | Dependency updates, CI, tooling |

Examples: `feat/websocket-output`, `fix/hotplug-race-condition`, `test/pipeline-integration`

### Commit Messages

Use conventional-style prefixes:

```
feat: add WebSocket output target
fix: prevent race in hot-plug device reconnect
docs: document IPC protocol and JSON-RPC methods
test: add pipeline integration tests for modifier interception
refactor: extract output dispatch into interface
chore: update mapping dependency to latest
```

## Domain-Specific Guidance

### Adding a New Output Type

Output targets live in `internal/output/`. To add a new one:

1. Create `internal/output/{type}.go` implementing the `OutputTarget` interface.
2. Add platform-specific files if needed (`{type}_{linux,darwin,windows}.go`).
3. Register the output type in the pipeline dispatch table.
4. Add tests covering both success and error paths.
5. Update the output type matrix in `CLAUDE.md`.

### Adding a New Input Provider

Input providers live in mcpkit's `device` package. If you need a new provider
for mapitall:

1. Implement the provider in mcpkit/device following its conventions.
2. Wire it into `internal/daemon/` initialization.
3. Add platform build tags as needed.

### Pipeline and Hot Path

The event pipeline in `internal/pipeline/` is the performance-critical path.
Changes here must:

- Avoid allocations in the per-event loop where possible.
- Maintain the modifier interception logic (custom modifiers update EngineState,
  not dispatch).
- Include benchmarks if modifying the hot path.

### IPC Protocol

The Unix socket IPC in `internal/ipc/` uses JSON-RPC 2.0. When adding new
methods:

- Add both the server handler and client method.
- Document the method signature and parameters.
- Add round-trip tests.

### Service Files

Platform service files (`mapitall.service`, `50-mapitall.rules`,
`com.hairglasses.mapitall.plist`) live in the repo root. If you modify daemon
behavior that affects service lifecycle:

- Update the relevant service file.
- Test with `systemd` (Linux) or `launchd` (macOS) where possible.

## Code Conventions

### Formatting and Linting

All code must pass these checks before submission:

```bash
gofmt -l .              # Must produce no output (all files formatted)
go vet ./...            # Must produce no warnings
golangci-lint run ./... # If installed
```

Use `gofmt` or `goimports` to format code.

### Error Handling

Return descriptive errors with lowercase messages and no trailing punctuation.
Never use naked panics. The daemon must recover gracefully from transient
failures (device disconnects, IPC errors).

### Thread Safety

The daemon runs concurrent goroutines for each device pipeline, hot-plug polling,
active-app tracking, and IPC. Protect shared state with `sync.RWMutex`. Use
`RLock` for reads and `Lock` for writes.

### CGO

macOS outputs use CGO for CoreGraphics, Carbon, and CoreMIDI. Linux and Windows
outputs are pure Go. When adding CGO code:

- Gate it with `//go:build darwin` build tags.
- Provide a stub file for other platforms (`_stub.go`).
- Document any framework dependencies.

### General Style

- Format with `gofmt` (or `goimports`).
- Pass `go vet ./...` with no warnings.
- Prefer returning concrete types over interfaces.
- Error messages: lowercase, no trailing punctuation.
- Context: always accept and propagate `context.Context`.
- Imports: stdlib first, then external, then internal (blank line separated).

## Testing Requirements

### All PRs Must

- Include tests for new functionality.
- Pass `go test ./... -count=1` with zero failures.
- Pass `go test ./... -count=1 -race` with zero race conditions.

### Table-Driven Tests (Required Pattern)

Use table-driven tests with `t.Run` subtests:

```go
func TestDispatch(t *testing.T) {
    tests := []struct {
        name       string
        outputType string
        action     OutputAction
        wantErr    bool
    }{
        {name: "key press", outputType: "key", action: OutputAction{Key: "BTN_A"}},
        {name: "command exec", outputType: "command", action: OutputAction{Command: "echo test"}},
        {name: "unknown type", outputType: "invalid", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

### Unit Tests

- Use stdlib `testing` package (no testify, no gomega).
- Name test files `*_test.go` in the same package.
- Test error paths, not just happy paths.
- Mock hardware dependencies for unit tests (no real device access required).

## Reporting Bugs

Open a [GitHub Issue](https://github.com/hairglasses-studio/mapitall/issues)
with:

- Go version, OS, and kernel version
- Controller/device model and connection method (USB, Bluetooth)
- Minimal TOML profile that reproduces the issue
- Expected vs. actual behavior
- Relevant log output (`mapitall run -v`)

## License

mapitall is MIT licensed (Copyright 2024-2026 hairglasses-studio). By
contributing, you agree that your contributions will be licensed under the same
[MIT License](LICENSE).
