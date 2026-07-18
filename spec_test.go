package compforge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSpecFile_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.json")

	data := `{
		"name": "testcli",
		"version": "1.0.0",
		"description": "A test CLI",
		"commands": [
			{"name": "build", "description": "Build the project"},
			{"name": "test", "description": "Run tests"}
		]
	}`
	if err := os.WriteFile(specPath, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	spec, err := ParseSpecFile(specPath)
	if err != nil {
		t.Fatalf("ParseSpecFile failed: %v", err)
	}

	if spec.Name != "testcli" {
		t.Errorf("expected name 'testcli', got '%s'", spec.Name)
	}
	if spec.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", spec.Version)
	}
	if len(spec.Commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(spec.Commands))
	}
	if spec.Commands[0].Name != "build" {
		t.Errorf("expected first command 'build', got '%s'", spec.Commands[0].Name)
	}
}

func TestParseSpecFile_MissingName(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.json")

	data := `{"version": "1.0.0"}`
	if err := os.WriteFile(specPath, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ParseSpecFile(specPath)
	if err == nil {
		t.Error("expected error for missing name, got nil")
	}
}

func TestParseSpecFile_BadJSON(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.json")

	data := `not valid json`
	if err := os.WriteFile(specPath, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ParseSpecFile(specPath)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestParseCompletionFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected CompletionFormat
		wantErr  bool
	}{
		{"bash", FormatBash, false},
		{"zsh", FormatZsh, false},
		{"fish", FormatFish, false},
		{"powershell", FormatPowerShell, false},
		{"pwsh", FormatPowerShell, false},
		{"all", FormatAll, false},
		{"Bash", FormatBash, false}, // case-insensitive
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseCompletionFormat(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestSpec_GenerateBash(t *testing.T) {
	spec := &Spec{
		Name:        "test",
		Description: "A test CLI",
		Commands: []CLICommand{
			{Name: "build", Description: "Build the project"},
			{Name: "test", Description: "Run tests"},
		},
	}

	result, err := spec.GenerateCompletion(FormatBash)
	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty bash completion")
	}
	if !contains(result, "_test_completion") {
		t.Error("bash completion should contain function name _test_completion")
	}
	if !contains(result, "test") {
		t.Error("bash completion should mention 'test'")
	}
}

func TestSpec_GenerateZsh(t *testing.T) {
	spec := &Spec{
		Name: "test",
		Commands: []CLICommand{
			{Name: "build"},
			{Name: "test"},
		},
	}

	result, err := spec.GenerateCompletion(FormatZsh)
	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty zsh completion")
	}
	if !contains(result, "_test") {
		t.Error("zsh completion should contain function name _test")
	}
}

func TestSpec_GenerateFish(t *testing.T) {
	spec := &Spec{
		Name: "test",
		Commands: []CLICommand{
			{Name: "build"},
		},
	}

	result, err := spec.GenerateCompletion(FormatFish)
	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty fish completion")
	}
	if !contains(result, "__test_complete") {
		t.Error("fish completion should contain function name __test_complete")
	}
}

func TestSpec_GeneratePowerShell(t *testing.T) {
	spec := &Spec{
		Name: "test",
		Commands: []CLICommand{
			{Name: "build"},
		},
	}

	result, err := spec.GenerateCompletion(FormatPowerShell)
	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty powershell completion")
	}
	if !contains(result, "Register-ArgumentCompleter") {
		t.Error("powershell completion should contain Register-ArgumentCompleter")
	}
}

func TestSpec_GenerateAll(t *testing.T) {
	spec := &Spec{
		Name: "test",
		Commands: []CLICommand{
			{Name: "build"},
		},
	}

	result, err := spec.GenerateCompletion(FormatAll)
	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty all completions")
	}
	if !contains(result, "=== bash ===") {
		t.Error("all completions should include bash section")
	}
	if !contains(result, "=== zsh ===") {
		t.Error("all completions should include zsh section")
	}
	if !contains(result, "=== fish ===") {
		t.Error("all completions should include fish section")
	}
	if !contains(result, "=== powershell ===") {
		t.Error("all completions should include powershell section")
	}
}

func TestSpec_Validate(t *testing.T) {
	tests := []struct {
		name        string
		spec        *Spec
		expectCount int
	}{
		{
			name:        "empty name",
			spec:        &Spec{},
			expectCount: 2, // name is required + version recommended
		},
		{
			name:        "no version",
			spec:        &Spec{Name: "test"},
			expectCount: 1, // version recommended
		},
		{
			name: "empty command name",
			spec: &Spec{
				Name:     "test",
				Commands: []CLICommand{{}},
			},
			expectCount: 2, // name required + version recommended
		},
		{
			name: "duplicate command names",
			spec: &Spec{
				Name: "test",
				Commands: []CLICommand{
					{Name: "build"},
					{Name: "build"},
				},
			},
			expectCount: 2, // version recommended + duplicate detected
		},
		{
			name: "duplicate options",
			spec: &Spec{
				Name: "test",
				Commands: []CLICommand{
					{
						Name: "build",
						Options: []CLIOption{
							{Name: "target"},
							{Name: "target"},
						},
					},
				},
			},
			expectCount: 2, // version recommended + duplicate detected
		},
		{
			name: "valid spec",
			spec: &Spec{
				Name: "test",
				Commands: []CLICommand{
					{Name: "build", Options: []CLIOption{{Name: "target"}}},
				},
			},
			expectCount: 1, // only version recommended
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := tt.spec.Validate()
			if len(issues) != tt.expectCount {
				t.Errorf("expected %d issues, got %d: %v", tt.expectCount, len(issues), issues)
			}
		})
	}
}

func TestSpec_FlattenCommands(t *testing.T) {
	spec := &Spec{
		Name: "test",
		Commands: []CLICommand{
			{
				Name: "build",
				SubCommands: []CLICommand{
					{Name: "all"},
					{Name: "deps"},
				},
			},
			{Name: "test"},
		},
	}

	flat := spec.FlattenCommands()
	if len(flat) != 4 {
		t.Errorf("expected 4 flattened commands, got %d", len(flat))
	}

	// Check full name includes parent
	found := false
	for _, cmd := range flat {
		if cmd.ShortName == "deps" {
			if cmd.FullName != "build deps" {
				t.Errorf("expected full name 'build deps', got '%s'", cmd.FullName)
			}
			found = true
		}
	}
	if !found {
		t.Error("expected to find 'deps' in flattened commands")
	}
}

func TestParseSpec_JSON(t *testing.T) {
	data := []byte(`{"name": "test", "version": "1.0.0"}`)
	spec, err := ParseSpec(data)
	if err != nil {
		t.Fatalf("ParseSpec failed: %v", err)
	}
	if spec.Name != "test" {
		t.Errorf("expected 'test', got '%s'", spec.Name)
	}
}

func TestCompletionFormat_Extension(t *testing.T) {
	tests := []struct {
		format CompletionFormat
		ext    string
	}{
		{FormatBash, "bash"},
		{FormatZsh, "zsh"},
		{FormatFish, "fish"},
		{FormatPowerShell, "ps1"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if got := tt.format.Extension(); got != tt.ext {
				t.Errorf("expected extension '%s', got '%s'", tt.ext, got)
			}
		})
	}
}

func TestEscapeShellWord(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "'hello'"},
		{"hello world", "'hello world'"},
		{"it's", "'it'\\''s'"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := EscapeShellWord(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestProcessValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", `"hello"`},
		{"true", "true"},
		{"false", "false"},
		{"null", "null"},
		{"42", "42"},
		{"3.14", "3.14"},
		{`"quoted"`, `"quoted"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ProcessValue(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSpec_GetInstallPaths(t *testing.T) {
	spec := &Spec{Name: "test"}
	paths := spec.GetInstallPaths()

	if paths[FormatBash] != "/etc/bash_completion.d/test" {
		t.Errorf("unexpected bash path: %s", paths[FormatBash])
	}
	if paths[FormatZsh] != "/usr/share/zsh/site-functions/_test" {
		t.Errorf("unexpected zsh path: %s", paths[FormatZsh])
	}
	if paths[FormatFish] != "~/.config/fish/completions/test.fish" {
		t.Errorf("unexpected fish path: %s", paths[FormatFish])
	}
}

func TestSpec_GenerateCompletion_InvalidFormat(t *testing.T) {
	spec := &Spec{Name: "test"}
	_, err := spec.GenerateCompletion(CompletionFormat("invalid"))
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestSpec_GenerateBash_WithOptions(t *testing.T) {
	spec := &Spec{
		Name: "test",
		Commands: []CLICommand{
			{
				Name: "build",
				Options: []CLIOption{
					{Name: "--target", Description: "Build target", Choices: []string{"debug", "release"}},
				},
			},
		},
	}

	result, err := spec.GenerateCompletion(FormatBash)
	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}

	if !contains(result, "--target") {
		t.Error("bash completion should contain option --target")
	}
	if !contains(result, "debug") {
		t.Error("bash completion should contain choice 'debug'")
	}
}

func TestSpec_GenerateZsh_WithChoices(t *testing.T) {
	spec := &Spec{
		Name: "test",
		Commands: []CLICommand{
			{
				Name: "deploy",
				Options: []CLIOption{
					{Name: "--env", Description: "Environment", Choices: []string{"staging", "production"}},
				},
			},
		},
	}

	result, err := spec.GenerateCompletion(FormatZsh)
	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}

	if !contains(result, "--env") {
		t.Error("zsh completion should contain option --env")
	}
}

func TestSpec_GenerateFish_WithChoices(t *testing.T) {
	spec := &Spec{
		Name: "test",
		Commands: []CLICommand{
			{
				Name: "config",
				Options: []CLIOption{
					{Name: "--format", Choices: []string{"json", "yaml"}},
				},
			},
		},
	}

	result, err := spec.GenerateCompletion(FormatFish)
	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}

	if !contains(result, "--format") {
		t.Error("fish completion should contain option --format")
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		(len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
