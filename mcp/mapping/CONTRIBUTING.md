# Contributing to mapping

Thank you for your interest in contributing to mapping, the shared Go package
for the mapitall controller-mapping ecosystem. This library provides the unified
type system, rule-resolution engine, TOML profile parsing, value-transform LUTs,
and legacy-format conversion used across all mapping tools.

**Before you start:**

- Browse [GitHub Issues](https://github.com/hairglasses-studio/mapping/issues) for
  existing work and discussion.
- Read the project [CLAUDE.md](CLAUDE.md) for architecture details and key types.

## Development Setup

### Prerequisites

- **Go 1.26.2** or later
- **git**

### Clone and Build

```bash
git clone https://github.com/hairglasses-studio/mapping.git
cd mapping
go build ./...
```

### Run Tests

```bash
go test ./... -count=1
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

## Pull Request Workflow

1. **Fork** the repository and clone your fork.
2. **Create a branch** from `main` using the naming convention below.
3. **Make your changes** with tests.
4. **Run the full pipeline locally:**
   ```bash
   go build ./... && go vet ./... && go test ./... -count=1 -race
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

Examples: `feat/midi-output-action`, `fix/lut-boundary-clamp`, `test/engine-modifier-matrix`

### Commit Messages

Use conventional-style prefixes:

```
feat: add threshold gating to value transforms
fix: handle empty modifier list in rule resolution
docs: add examples for legacy profile conversion
test: add table-driven tests for LUT compilation
refactor: simplify condition evaluation in engine
chore: update go.mod dependencies
```

## Domain-Specific Guidance

### Mapping Profiles

Profiles are TOML files describing how device inputs map to outputs. If you are
adding or changing profile schema fields:

- Update the types in `types.go` with proper TOML and JSON struct tags.
- Ensure both unified (`schema_version >= 1`) and legacy makima formats are handled.
- Update `validate.go` to catch new invalid states.
- Add conversion logic in `convert.go` if the field existed in legacy format.

### Rule Resolution Engine

The `RuleIndex` in `engine.go` provides O(1) best-match resolution. Changes to
resolution logic must:

- Preserve the priority order: app-override > layer > condition > modifier > priority.
- Maintain concurrency safety on `EngineState` (use `sync.RWMutex`).
- Include benchmarks if the change affects the hot path.

### Value Transforms and LUTs

`lut.go` pre-compiles 256-entry lookup tables for fast per-event transformation.
If you add a new curve type:

- Add the curve constant to `types.go`.
- Implement the computation in `lut.go`.
- Add test coverage for boundary values (0, 127, 255) and midpoints.

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

Return descriptive errors with lowercase messages and no trailing punctuation:

```go
return nil, fmt.Errorf("unknown curve type: %s", curve)
```

Never use naked panics. If something is genuinely unrecoverable, return an error.

### Thread Safety

`EngineState` and any shared runtime state must use `sync.RWMutex`. Use `RLock`
for reads and `Lock` for writes:

```go
func (s *EngineState) GetModifiers() []string {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.modifiers
}
```

### General Style

- Format with `gofmt` (or `goimports`).
- Pass `go vet ./...` with no warnings.
- Prefer returning concrete types over interfaces.
- Error messages: lowercase, no trailing punctuation.
- Context: always accept and propagate `context.Context` where applicable.
- Imports: stdlib first, then external, then internal (blank line separated).

## Testing Requirements

### All PRs Must

- Include tests for new functionality.
- Pass `go test ./... -count=1` with zero failures.
- Pass `go test ./... -count=1 -race` with zero race conditions.

### Table-Driven Tests (Required Pattern)

Use table-driven tests with `t.Run` subtests:

```go
func TestResolve(t *testing.T) {
    tests := []struct {
        name     string
        source   string
        state    EngineState
        wantRule string
    }{
        {name: "exact match", source: "BTN_SOUTH", wantRule: "jump"},
        {name: "with modifier", source: "BTN_SOUTH", state: EngineState{modifiers: []string{"BTN_SELECT"}}, wantRule: "special_jump"},
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

## Reporting Bugs

Open a [GitHub Issue](https://github.com/hairglasses-studio/mapping/issues) with:

- Go version and OS
- Minimal reproduction steps (include a sample TOML profile if relevant)
- Expected vs. actual behavior
- Relevant error output

## License

mapping is MIT licensed (Copyright 2024-2026 hairglasses-studio). By
contributing, you agree that your contributions will be licensed under the same
[MIT License](LICENSE).
