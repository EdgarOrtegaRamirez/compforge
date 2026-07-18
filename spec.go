package compforge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// CompletionFormat represents supported shell completion formats.
type CompletionFormat string

const (
	FormatBash       CompletionFormat = "bash"
	FormatZsh        CompletionFormat = "zsh"
	FormatFish       CompletionFormat = "fish"
	FormatPowerShell CompletionFormat = "powershell"
	FormatAll        CompletionFormat = "all"
)

// ParseCompletionFormat parses a string into a CompletionFormat.
func ParseCompletionFormat(s string) (CompletionFormat, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "bash":
		return FormatBash, nil
	case "zsh":
		return FormatZsh, nil
	case "fish":
		return FormatFish, nil
	case "powershell", "pwsh":
		return FormatPowerShell, nil
	case "all":
		return FormatAll, nil
	default:
		return "", fmt.Errorf("unknown completion format %q: must be one of bash, zsh, fish, powershell, or all", s)
	}
}

// CLICommand represents a command in the CLI specification.
type CLICommand struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Aliases     []string     `json:"aliases,omitempty"`
	Options     []CLIOption  `json:"options,omitempty"`
	SubCommands []CLICommand `json:"subcommands,omitempty"`
	Args        []CLIArg     `json:"args,omitempty"`
	Requires    []string     `json:"requires,omitempty"`
}

// CLIOption represents an option/flag for a command.
type CLIOption struct {
	Name        string   `json:"name"`
	Shorthand   string   `json:"shorthand,omitempty"`
	Type        string   `json:"type,omitempty"`
	Description string   `json:"description,omitempty"`
	Default     string   `json:"default,omitempty"`
	Choices     []string `json:"choices,omitempty"`
	Required    bool     `json:"required,omitempty"`
}

// CLIArg represents a positional argument.
type CLIArg struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type,omitempty"`
	Choices     []string `json:"choices,omitempty"`
	Optional    bool     `json:"optional,omitempty"`
}

// Spec represents the full CLI specification.
type Spec struct {
	Name        string       `json:"name"`
	Version     string       `json:"version,omitempty"`
	Description string       `json:"description,omitempty"`
	Author      string       `json:"author,omitempty"`
	Commands    []CLICommand `json:"commands,omitempty"`
	Options     []CLIOption  `json:"options,omitempty"`
}

// ParseSpecFile reads and parses a YAML/JSON spec file.
func ParseSpecFile(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file %s: %w", path, err)
	}
	spec, err := ParseSpec(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec: %w", err)
	}
	return spec, nil
}

// ParseSpec parses JSON spec data.
func ParseSpec(data []byte) (*Spec, error) {
	spec := &Spec{}
	if err := json.Unmarshal(data, spec); err != nil {
		return nil, fmt.Errorf("failed to parse spec as JSON: %w", err)
	}
	if spec.Name == "" {
		return nil, fmt.Errorf("spec must have a 'name' field")
	}
	return spec, nil
}

// GenerateCompletion generates shell completion scripts for the given format.
func (s *Spec) GenerateCompletion(format CompletionFormat) (string, error) {
	switch format {
	case FormatBash:
		return generateBash(s)
	case FormatZsh:
		return generateZsh(s)
	case FormatFish:
		return generateFish(s)
	case FormatPowerShell:
		return generatePowerShell(s)
	case FormatAll:
		var parts []string
		for _, f := range []CompletionFormat{FormatBash, FormatZsh, FormatFish, FormatPowerShell} {
			script, err := s.GenerateCompletion(f)
			if err != nil {
				return "", fmt.Errorf("failed to generate %s completion: %w", f, err)
			}
			parts = append(parts, fmt.Sprintf("=== %s ===\n\n%s", f, script))
		}
		return strings.Join(parts, "\n\n"), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// GetInstallPaths returns suggested installation paths for each shell.
func (s *Spec) GetInstallPaths() map[CompletionFormat]string {
	paths := map[CompletionFormat]string{
		FormatBash:       fmt.Sprintf("/etc/bash_completion.d/%s", s.Name),
		FormatZsh:        fmt.Sprintf("/usr/share/zsh/site-functions/_%s", s.Name),
		FormatFish:       fmt.Sprintf("~/.config/fish/completions/%s.fish", s.Name),
		FormatPowerShell: "~/.local/share/compforge/completions/" + s.Name + ".ps1",
	}
	return paths
}

// Validate checks the spec for common issues.
func (s *Spec) Validate() []string {
	var issues []string
	if s.Name == "" {
		issues = append(issues, "spec name is required")
	}
	if s.Version == "" {
		issues = append(issues, "spec version is recommended")
	}
	seenCmds := make(map[string]bool)
	for _, cmd := range s.Commands {
		if cmd.Name == "" {
			issues = append(issues, "command name is required")
		}
		if seenCmds[cmd.Name] {
			issues = append(issues, fmt.Sprintf("duplicate command name: %s", cmd.Name))
		}
		seenCmds[cmd.Name] = true
		seenOpts := make(map[string]bool)
		for _, opt := range cmd.Options {
			if opt.Name == "" {
				issues = append(issues, fmt.Sprintf("option in command %q has no name", cmd.Name))
			}
			if seenOpts[opt.Name] {
				issues = append(issues, fmt.Sprintf("duplicate option %q in command %q", opt.Name, cmd.Name))
			}
			seenOpts[opt.Name] = true
		}
	}
	return issues
}

// FormatDate returns a formatted date string for use in generated scripts.
func FormatDate() string {
	return time.Now().UTC().Format("2006-01-02")
}

// EscapeShellWord escapes a string for safe use in shell scripts.
func EscapeShellWord(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// CmdInfo holds parsed command info for shell generation.
type CmdInfo struct {
	FullName    string
	ShortName   string
	Description string
	Aliases     []string
	Options     []CLIOption
	SubCommands []CmdInfo
	Args        []CLIArg
}

// FlattenCommands flattens the command tree into a list of all command paths.
func (s *Spec) FlattenCommands() []CmdInfo {
	var result []CmdInfo
	for _, cmd := range s.Commands {
		result = append(result, toCmdInfo("", cmd)...)
	}
	return result
}

func toCmdInfo(parent string, cmd CLICommand) []CmdInfo {
	name := cmd.Name
	if parent != "" {
		name = parent + " " + cmd.Name
	}
	info := CmdInfo{
		FullName:    name,
		ShortName:   cmd.Name,
		Description: cmd.Description,
		Aliases:     cmd.Aliases,
		Options:     cmd.Options,
		Args:        cmd.Args,
	}
	var result []CmdInfo
	result = append(result, info)
	for _, sub := range cmd.SubCommands {
		result = append(result, toCmdInfo(name, sub)...)
	}
	return result
}

// GetCompletionExtensions returns the file extension for each format.
func (f CompletionFormat) Extension() string {
	switch f {
	case FormatBash:
		return "bash"
	case FormatZsh:
		return "zsh"
	case FormatFish:
		return "fish"
	case FormatPowerShell:
		return "ps1"
	default:
		return "sh"
	}
}

// GetInstallDir returns the recommended install directory for a shell.
func (f CompletionFormat) GetInstallDir() string {
	switch f {
	case FormatBash:
		return "/etc/bash_completion.d"
	case FormatZsh:
		return "/usr/share/zsh/site-functions"
	case FormatFish:
		return os.ExpandEnv("~/.config/fish/completions")
	case FormatPowerShell:
		home := os.ExpandEnv("~")
		if home == "~" {
			home = "/root"
		}
		return filepath.Join(home, ".local", "share", "compforge", "completions")
	default:
		return "."
	}
}

// IsYAML checks if data looks like YAML content.
func IsYAML(data []byte) bool {
	content := strings.TrimSpace(string(data))
	if len(content) == 0 || strings.HasPrefix(content, "{") {
		return false
	}
	return strings.Contains(content, ":") || strings.HasPrefix(content, "- ")
}

// UTF8Len returns the character count of a string.
func UTF8Len(s string) int {
	return utf8.RuneCountInString(s)
}

// SanitizeKey converts a key to a safe JSON key.
func SanitizeKey(key string) string {
	key = strings.TrimSpace(key)
	if len(key) >= 2 && key[0] == '"' && key[len(key)-1] == '"' {
		key = key[1 : len(key)-1]
	}
	return strings.ReplaceAll(key, " ", "_")
}

// ProcessValue converts a YAML value string to a JSON value.
func ProcessValue(val string) string {
	val = strings.TrimSpace(val)
	if val == "" || val == "null" || val == "~" {
		return "null"
	}
	if val == "true" {
		return "true"
	}
	if val == "false" {
		return "false"
	}
	if n, err := strconv.Atoi(val); err == nil {
		return strconv.Itoa(n)
	}
	if f, err := strconv.ParseFloat(val, 64); err == nil {
		return fmt.Sprintf("%g", f)
	}
	if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
		return val
	}
	if len(val) >= 2 && val[0] == '\'' && val[len(val)-1] == '\'' {
		return val[1 : len(val)-1]
	}
	return strconv.Quote(val)
}
