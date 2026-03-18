package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Theme != "default" {
		t.Errorf("theme: got %q, want %q", cfg.Theme, "default")
	}
	if cfg.Format != "table" {
		t.Errorf("format: got %q, want %q", cfg.Format, "table")
	}
	if !cfg.Modules.Pwd.Enabled {
		t.Error("pwd should be enabled by default")
	}
	if !cfg.Modules.Git.Enabled {
		t.Error("git should be enabled by default")
	}
	if cfg.Modules.Kube.Enabled {
		t.Error("kube should be disabled by default")
	}
	if cfg.Modules.Gcp.Enabled {
		t.Error("gcp should be disabled by default")
	}
}

func TestDefaultGitSymbols(t *testing.T) {
	s := DefaultGitSymbols()
	tests := []struct {
		name, got, want string
	}{
		{"unstaged", s.Unstaged, "*"},
		{"staged", s.Staged, "+"},
		{"stash", s.Stash, "$"},
		{"untracked", s.Untracked, "%"},
		{"ahead", s.Ahead, "↑"},
		{"behind", s.Behind, "↓"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg := Load("/nonexistent/path/config.yaml")
	if cfg.Theme != "default" {
		t.Errorf("missing file should return defaults, got theme=%q", cfg.Theme)
	}
}

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
theme: "dracula"
format: "table"
modules:
  pwd:
    enabled: false
    style: "full"
  git:
    enabled: true
    show_repo: true
    symbols:
      unstaged: "!"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := Load(path)
	if cfg.Theme != "dracula" {
		t.Errorf("theme: got %q, want %q", cfg.Theme, "dracula")
	}
	if cfg.Format != "table" {
		t.Errorf("format: got %q, want %q", cfg.Format, "table")
	}
	if cfg.Modules.Pwd.Enabled {
		t.Error("pwd should be disabled")
	}
	if cfg.Modules.Pwd.Style != "full" {
		t.Errorf("pwd style: got %q, want %q", cfg.Modules.Pwd.Style, "full")
	}
	if !cfg.Modules.Git.ShowRepo {
		t.Error("git show_repo should be true")
	}
	if cfg.Modules.Git.Symbols.Unstaged != "!" {
		t.Errorf("unstaged: got %q, want %q", cfg.Modules.Git.Symbols.Unstaged, "!")
	}
	// Empty symbols should be filled with defaults
	if cfg.Modules.Git.Symbols.Staged != "+" {
		t.Errorf("staged should default to +, got %q", cfg.Modules.Git.Symbols.Staged)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := Load(path)
	if cfg.Theme != "default" {
		t.Errorf("invalid yaml should return defaults, got theme=%q", cfg.Theme)
	}
}

func TestGenerateDefault(t *testing.T) {
	out := GenerateDefault()
	if out == "" {
		t.Error("GenerateDefault should return non-empty string")
	}
	// Should contain key config sections
	for _, want := range []string{"theme:", "format:", "modules:", "pwd:", "git:", "kube:", "gcp:"} {
		if !contains(out, want) {
			t.Errorf("GenerateDefault should contain %q", want)
		}
	}
}

func TestDefaultModuleOrder(t *testing.T) {
	cfg := Default()
	want := []string{"pwd", "git", "kube", "gcp", "claude"}
	if len(cfg.ModuleOrder) != len(want) {
		t.Fatalf("ModuleOrder length: got %d, want %d", len(cfg.ModuleOrder), len(want))
	}
	for i, name := range want {
		if cfg.ModuleOrder[i] != name {
			t.Errorf("ModuleOrder[%d]: got %q, want %q", i, cfg.ModuleOrder[i], name)
		}
	}
}

func TestModuleOrderFromConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	// claude before git, pwd last
	content := `
modules:
  claude:
    enabled: true
  git:
    enabled: true
  pwd:
    enabled: true
`
	os.WriteFile(path, []byte(content), 0o644)

	cfg := Load(path)
	// Should be: claude, git, pwd, then defaults not in config (kube, gcp)
	want := []string{"claude", "git", "pwd", "kube", "gcp"}
	if len(cfg.ModuleOrder) != len(want) {
		t.Fatalf("ModuleOrder length: got %d, want %d\norder: %v", len(cfg.ModuleOrder), len(want), cfg.ModuleOrder)
	}
	for i, name := range want {
		if cfg.ModuleOrder[i] != name {
			t.Errorf("ModuleOrder[%d]: got %q, want %q", i, cfg.ModuleOrder[i], name)
		}
	}
}

func TestModuleOrderPartial(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	// Only git specified
	content := `
modules:
  git:
    enabled: true
`
	os.WriteFile(path, []byte(content), 0o644)

	cfg := Load(path)
	// git first, then remaining defaults
	if cfg.ModuleOrder[0] != "git" {
		t.Errorf("first module should be git, got %q", cfg.ModuleOrder[0])
	}
	// All default modules should be present
	seen := make(map[string]bool)
	for _, name := range cfg.ModuleOrder {
		seen[name] = true
	}
	for _, name := range DefaultModuleOrder {
		if !seen[name] {
			t.Errorf("missing default module %q in order", name)
		}
	}
}

func TestModuleOrderMissing(t *testing.T) {
	cfg := Load("/nonexistent/path")
	// Should fall back to default order
	for i, name := range DefaultModuleOrder {
		if cfg.ModuleOrder[i] != name {
			t.Errorf("fallback ModuleOrder[%d]: got %q, want %q", i, cfg.ModuleOrder[i], name)
		}
	}
}

func TestExtractModuleOrderInvalidYAML(t *testing.T) {
	order := extractModuleOrder([]byte("{{invalid"))
	for i, name := range DefaultModuleOrder {
		if order[i] != name {
			t.Errorf("invalid yaml order[%d]: got %q, want %q", i, order[i], name)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
