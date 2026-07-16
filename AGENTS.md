# CompForge — AI Agent Guide

## Project Overview
CompForge generates shell completion scripts (bash, zsh, fish, PowerShell) from a simple JSON/YAML specification file. The core is a spec parser + shell code generator with cobra-based CLI.

## Build & Test
```bash
go build ./...
go test ./...
go vet ./...
golangci-lint run
```

## Architecture
- `spec.go` — Spec data types, JSON parsing, validation, command flattening, shell format helpers
- `bash.go` — Bash completion script generator (case-based completion)
- `shells.go` — Zsh, Fish, and PowerShell completion generators
- `cmd/compforge/main.go` — Cobra CLI with generate, validate, info, sample commands

## Key Design Decisions
- Spec parsed as JSON via `encoding/json`; YAML spec files require manual conversion or external tools
- Completions support: command names, subcommand hierarchies, option flags, and value choices
- No external shell dependencies — generates standalone scripts
- Each shell generator produces self-contained scripts (no runtime library needed)

## Adding Shell Support
1. Add new `CompletionFormat` constant
2. Implement `generateShellName(spec *Spec) (string, error)` in a new file or existing shells.go
3. Add case in `GenerateCompletion()` switch
4. Add tests

## Testing
```bash
# Run all tests
go test ./... -v

# Run with race detector
go test -race ./...
```

## Common Pitfalls
- JSON spec requires valid JSON syntax (YAML needs manual conversion for now)
- Empty aliases arrays produce empty strings in bash case statements — always check for empty slices
- The `complete` command line at the end of bash completions must list all top-level commands