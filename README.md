# CompForge

**Generate shell completions from CLI specifications — bash, zsh, fish, and PowerShell.**

[![CI](https://github.com/EdgarOrtegaRamirez/compforge/actions/workflows/ci.yml/badge.svg)](https://github.com/EdgarOrtegaRamirez/compforge/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/EdgarOrtegaRamirez/compforge)](https://goreportcard.com/report/github.com/EdgarOrtegaRamirez/compforge)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-00ADD8)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## What It Does

CompForge generates shell completion scripts from a simple CLI specification file. Write your CLI commands once, and get completions for all major shells:

- **Bash** — Full completion with command, option, and subcommand support
- **Zsh** — Native zsh completion functions with `_values` and `_arguments`
- **Fish** — Fish shell completion functions with proper argument handling
- **PowerShell** — `Register-ArgumentCompleter` scripts for Windows

## Install

```bash
go install github.com/EdgarOrtegaRamirez/compforge/cmd/compforge@latest
```

Or download a binary from [Releases](https://github.com/EdgarOrtegaRamirez/compforge/releases).

## Quick Start

### 1. Write a Spec File

Create a `spec.json` (or `spec.yaml`) file:

```json
{
  "name": "mycli",
  "version": "1.0.0",
  "description": "A sample CLI tool",
  "commands": [
    {
      "name": "build",
      "description": "Build the project",
      "options": [
        {
          "name": "--target",
          "description": "Build target",
          "choices": ["debug", "release", "profile"]
        }
      ]
    },
    {
      "name": "deploy",
      "description": "Deploy the application",
      "options": [
        { "name": "--env", "choices": ["staging", "production"] },
        { "name": "--force", "shorthand": "f" }
      ]
    }
  ]
}
```

### 2. Generate Completions

```bash
# Bash
compforge generate spec.json -f bash > mycli.bash
source mycli.bash

# Zsh
compforge generate spec.json -f zsh > _mycli
mv _mycli /usr/share/zsh/site-functions/

# Fish
compforge generate spec.json -f fish > mycli.fish
cp mycli.fish ~/.config/fish/completions/

# PowerShell (run in PowerShell)
compforge generate spec.json -f powershell > mycli.ps1
. ./mycli.ps1
```

### 3. Auto-Install All

```bash
compforge generate spec.json -f all -o completions/
```

This creates `completions/mycli.bash`, `completions/_mycli` (zsh), `completions/mycli.fish`, and `completions/mycli.ps1`.

### 4. Install Directly

```bash
compforge generate spec.json -f bash -i
# Installs to /etc/bash_completion.d/mycli
```

## Spec File Format

### JSON Example

```json
{
  "name": "project-cli",
  "version": "2.0.0",
  "description": "Project management CLI",
  "options": [
    { "name": "--verbose", "shorthand": "v", "description": "Verbose output" },
    { "name": "--quiet", "shorthand": "q", "description": "Quiet mode" }
  ],
  "commands": [
    {
      "name": "init",
      "description": "Initialize a new project",
      "options": [
        { "name": "--lang", "description": "Project language", "choices": ["go", "python", "rust", "node"] }
      ]
    },
    {
      "name": "build",
      "description": "Build the project",
      "subcommands": [
        { "name": "all", "description": "Build all targets" },
        { "name": "deps", "description": "Build dependencies only" }
      ]
    },
    {
      "name": "config",
      "description": "Manage configuration",
      "aliases": ["cfg", "conf"],
      "options": [
        { "name": "--format", "choices": ["json", "yaml"] }
      ]
    }
  ]
}
```

### YAML Example

```yaml
name: project-cli
version: 2.0.0
description: Project management CLI
commands:
  - name: init
    description: Initialize a new project
    options:
      - name: --lang
        description: Project language
        choices: [go, python, rust, node]
  - name: build
    description: Build the project
    subcommands:
      - name: all
        description: Build all targets
      - name: deps
        description: Build dependencies only
```

## Spec File Schema

| Field       | Type              | Description                              |
|-------------|-------------------|------------------------------------------|
| `name`      | string (required) | CLI binary name                          |
| `version`   | string            | Version string                           |
| `description`| string           | CLI description                          |
| `commands`  | array             | List of CLI commands                     |
| `options`   | array             | Global CLI options                       |

### Command Fields

| Field        | Type              | Description                          |
|--------------|-------------------|--------------------------------------|
| `name`       | string (required) | Command name                         |
| `description`| string            | Command description                  |
| `aliases`    | array of strings  | Command aliases                      |
| `options`    | array             | Command-specific options             |
| `subcommands`| array             | Nested subcommands                   |
| `args`       | array             | Positional arguments                 |

### Option Fields

| Field       | Type              | Description                              |
|-------------|-------------------|------------------------------------------|
| `name`      | string (required) | Option name (e.g., `--verbose`)          |
| `shorthand` | string            | Short flag (e.g., `v`)                   |
| `choices`   | array of strings  | Allowed values for tab completion        |
| `description`| string           | Option description                       |
| `default`   | string            | Default value                            |
| `type`      | string            | Type hint (bool, string, int, etc.)      |
| `required`  | boolean           | Whether the option is required           |

## CLI Commands

| Command         | Description                                        |
|-----------------|----------------------------------------------------|
| `generate`      | Generate shell completions from a spec file        |
| `validate`      | Validate a spec file and report issues             |
| `info`          | Display parsed information about a spec            |
| `sample`        | Generate a sample spec file template               |

## Validation

CompForge validates spec files and reports common issues:

```bash
compforge validate spec.json
# Spec is valid!

compforge validate spec.json 2>&1
#   - duplicate command name: build
#   - option in command "build" has no name
#
# 2 issue(s) found
```

## License

MIT — see [LICENSE](LICENSE) for details.

## Contributing

Issues and pull requests are welcome!