# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in CompForge, please **do not open a public issue**.

Instead, please report it directly via [GitHub Security Advisories](https://github.com/EdgarOrtegaRamirez/compforge/security/advisories/new) or email the maintainers.

## Security Considerations

CompForge is a CLI tool that generates shell completion scripts from specification files. It:

- Does NOT make network calls
- Does NOT execute system commands
- Does NOT modify system files (unless `-i` flag is used with user consent)
- Validates and sanitizes input before generating scripts

## Input Validation

- Spec files are parsed as JSON/YAML with standard libraries
- Command names and option names are validated
- Output is sanitized for safe shell script generation

## Dependencies

All Go dependencies are pinned in `go.mod`. Regular dependency audits are recommended.