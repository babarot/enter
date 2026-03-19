package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Theme != "default" {
		t.Errorf("theme: got %q, want %q", cfg.Theme, "default")
	}
	if cfg.Format != "table" {
		t.Errorf("format: got %q, want %q", cfg.Format, "table")
	}
	if !cfg.Modules.Cwd.Enabled {
		t.Error("cwd should be enabled by default")
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
  cwd:
    enabled: false
    style: "full"
  git:
    enabled: true
    fields:
      url:
        enabled: true
      summary:
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
	if cfg.Modules.Cwd.Enabled {
		t.Error("cwd should be disabled")
	}
	if cfg.Modules.Cwd.Style != "full" {
		t.Errorf("cwd style: got %q, want %q", cfg.Modules.Cwd.Style, "full")
	}
	if !cfg.Modules.Git.Fields.Url.Enabled {
		t.Error("git show_repo should be true")
	}
	if cfg.Modules.Git.Fields.Summary.Symbols.Unstaged != "!" {
		t.Errorf("unstaged: got %q, want %q", cfg.Modules.Git.Fields.Summary.Symbols.Unstaged, "!")
	}
	// Empty symbols should be filled with defaults
	if cfg.Modules.Git.Fields.Summary.Symbols.Staged != "+" {
		t.Errorf("staged should default to +, got %q", cfg.Modules.Git.Fields.Summary.Symbols.Staged)
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
	for _, want := range []string{"theme:", "format:", "modules:", "cwd:", "git:", "kube:", "gcp:"} {
		if !contains(out, want) {
			t.Errorf("GenerateDefault should contain %q", want)
		}
	}
}

func TestDefaultModuleOrder(t *testing.T) {
	cfg := Default()
	want := []string{"cwd", "git", "kube", "gcp", "claude", "codex"}
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
	// claude before git, cwd last
	content := `
modules:
  claude:
    enabled: true
  git:
    enabled: true
  cwd:
    enabled: true
`
	os.WriteFile(path, []byte(content), 0o644)

	cfg := Load(path)
	// Should be: claude, git, cwd, then defaults not in config (kube, gcp, codex)
	want := []string{"claude", "git", "cwd", "kube", "gcp", "codex"}
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
	order, _ := extractOrder([]byte("{{invalid"))
	for i, name := range DefaultModuleOrder {
		if order[i] != name {
			t.Errorf("invalid yaml order[%d]: got %q, want %q", i, order[i], name)
		}
	}
}

func TestValidate(t *testing.T) {
	cfg := Default()

	// Set invalid values
	cfg.Format = "invalid"
	cfg.Trigger = "invalid"
	cfg.KeyStyle = "invalid"
	cfg.Modules.Git.Fields.Cwd.Style = "invalid"
	cfg.Modules.Git.Fields.Status.Style = "invalid"
	cfg.Modules.Claude.Mode = "invalid"
	cfg.Modules.Claude.Fields.Usage.BarStyle = "invalid"
	cfg.Modules.Claude.Fields.Usage.TimeStyle = "invalid"
	cfg.Modules.Claude.Fields.Usage.CacheTTL = -1

	cfg.validate()

	d := Default()
	if cfg.Format != d.Format {
		t.Errorf("format: got %q, want %q", cfg.Format, d.Format)
	}
	if cfg.Trigger != d.Trigger {
		t.Errorf("trigger: got %q, want %q", cfg.Trigger, d.Trigger)
	}
	if cfg.KeyStyle != d.KeyStyle {
		t.Errorf("key_style: got %q, want %q", cfg.KeyStyle, d.KeyStyle)
	}
	if cfg.Modules.Git.Fields.Cwd.Style != d.Modules.Git.Fields.Cwd.Style {
		t.Errorf("git.cwd.style: got %q, want %q", cfg.Modules.Git.Fields.Cwd.Style, d.Modules.Git.Fields.Cwd.Style)
	}
	if cfg.Modules.Git.Fields.Status.Style != d.Modules.Git.Fields.Status.Style {
		t.Errorf("git.status.style: got %q, want %q", cfg.Modules.Git.Fields.Status.Style, d.Modules.Git.Fields.Status.Style)
	}
	if cfg.Modules.Claude.Mode != d.Modules.Claude.Mode {
		t.Errorf("claude.mode: got %q, want %q", cfg.Modules.Claude.Mode, d.Modules.Claude.Mode)
	}
	if cfg.Modules.Claude.Fields.Usage.BarStyle != d.Modules.Claude.Fields.Usage.BarStyle {
		t.Errorf("claude.usage.bar_style: got %q, want %q", cfg.Modules.Claude.Fields.Usage.BarStyle, d.Modules.Claude.Fields.Usage.BarStyle)
	}
	if cfg.Modules.Claude.Fields.Usage.TimeStyle != d.Modules.Claude.Fields.Usage.TimeStyle {
		t.Errorf("claude.usage.time_style: got %q, want %q", cfg.Modules.Claude.Fields.Usage.TimeStyle, d.Modules.Claude.Fields.Usage.TimeStyle)
	}
	if cfg.Modules.Claude.Fields.Usage.CacheTTL != d.Modules.Claude.Fields.Usage.CacheTTL {
		t.Errorf("claude.usage.cache_ttl: got %d, want %d", cfg.Modules.Claude.Fields.Usage.CacheTTL, d.Modules.Claude.Fields.Usage.CacheTTL)
	}
}

func TestValidateValidValues(t *testing.T) {
	cfg := Default()
	cfg.Format = "inline"
	cfg.Trigger = "on_cd"
	cfg.KeyStyle = "flat"
	cfg.Modules.Git.Fields.Cwd.Style = "breadcrumb"
	cfg.Modules.Git.Fields.Status.Style = "long"
	cfg.Modules.Claude.Mode = "always"
	cfg.Modules.Claude.Fields.Usage.BarStyle = "dot"
	cfg.Modules.Claude.Fields.Usage.TimeStyle = "relative"
	cfg.Modules.Claude.Fields.Usage.CacheTTL = 60

	cfg.validate()

	// All should remain as set
	if cfg.Format != "inline" {
		t.Errorf("format should stay inline, got %q", cfg.Format)
	}
	if cfg.Modules.Claude.Fields.Usage.BarStyle != "dot" {
		t.Errorf("bar_style should stay dot, got %q", cfg.Modules.Claude.Fields.Usage.BarStyle)
	}
}

func TestExtractSubKeyOrder(t *testing.T) {
	content := `
modules:
  git:
    fields:
      status:
        enabled: true
      summary:
        symbols: {}
      url:
        enabled: true
`
	_, subKeyOrder := extractOrder([]byte(content))
	if subKeyOrder == nil {
		t.Fatal("subKeyOrder should not be nil")
	}
	gitOrder, ok := subKeyOrder["git"]
	if !ok {
		t.Fatal("should have git sub-key order")
	}
	// Should be: status, sign, url
	want := []string{"status", "summary", "url"}
	if len(gitOrder) != len(want) {
		t.Fatalf("git sub-key order length: got %d, want %d (%v)", len(gitOrder), len(want), gitOrder)
	}
	for i, name := range want {
		if gitOrder[i] != name {
			t.Errorf("git sub-key order[%d]: got %q, want %q", i, gitOrder[i], name)
		}
	}
}

func TestStringOrSliceSingle(t *testing.T) {
	content := `dir: "foo"`
	var w When
	if err := yaml.Unmarshal([]byte(content), &w); err != nil {
		t.Fatal(err)
	}
	if len(w.Dir) != 1 || w.Dir[0] != "foo" {
		t.Errorf("got %v, want [foo]", w.Dir)
	}
}

func TestStringOrSliceMulti(t *testing.T) {
	content := `dir:
  - "foo"
  - "bar"
`
	var w When
	if err := yaml.Unmarshal([]byte(content), &w); err != nil {
		t.Fatal(err)
	}
	if len(w.Dir) != 2 || w.Dir[0] != "foo" || w.Dir[1] != "bar" {
		t.Errorf("got %v, want [foo bar]", w.Dir)
	}
}

func TestWhenMatchNil(t *testing.T) {
	var w *When
	if !w.Match("/any/path") {
		t.Error("nil When should match everything")
	}
}

func TestWhenMatchEmpty(t *testing.T) {
	w := &When{}
	if !w.Match("/any/path") {
		t.Error("empty When should match everything")
	}
}

func TestWhenMatchGlob(t *testing.T) {
	w := &When{Dir: StringOrSlice{"/home/user/src/**"}}
	if !w.Match("/home/user/src/github.com/foo") {
		t.Error("should match nested path")
	}
	if w.Match("/home/user/documents/foo") {
		t.Error("should not match unrelated path")
	}
}

func TestWhenMatchMultiplePatterns(t *testing.T) {
	w := &When{Dir: StringOrSlice{"/a/**", "/b/**"}}
	if !w.Match("/b/foo") {
		t.Error("should match second pattern")
	}
	if w.Match("/c/foo") {
		t.Error("should not match any pattern")
	}
}

func TestWhenMatchTildeExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}
	w := &When{Dir: StringOrSlice{"~/src/**"}}
	if !w.Match(filepath.Join(home, "src", "github.com", "foo")) {
		t.Error("should match path under ~/src")
	}
	if w.Match(filepath.Join(home, "documents", "foo")) {
		t.Error("should not match path outside ~/src")
	}
}

func TestWhenForAllModules(t *testing.T) {
	cfg := Default()
	conds := map[string]*When{
		ModuleCwd:    {Dir: StringOrSlice{"/cwd/**"}},
		ModuleGit:    {Dir: StringOrSlice{"/git/**"}},
		ModuleKube:   {Dir: StringOrSlice{"/kube/**"}},
		ModuleGcp:    {Dir: StringOrSlice{"/gcp/**"}},
		ModuleClaude: {Dir: StringOrSlice{"/claude/**"}},
		ModuleCodex:  {Dir: StringOrSlice{"/codex/**"}},
	}
	cfg.Modules.Cwd.When = conds[ModuleCwd]
	cfg.Modules.Git.When = conds[ModuleGit]
	cfg.Modules.Kube.When = conds[ModuleKube]
	cfg.Modules.Gcp.When = conds[ModuleGcp]
	cfg.Modules.Claude.When = conds[ModuleClaude]
	cfg.Modules.Codex.When = conds[ModuleCodex]

	for name, want := range conds {
		if got := cfg.Modules.WhenFor(name); got != want {
			t.Errorf("WhenFor(%s): got %v, want %v", name, got, want)
		}
	}
	if got := cfg.Modules.WhenFor("unknown"); got != nil {
		t.Error("WhenFor(unknown) should return nil")
	}
}

func TestLoadConfigWithWhen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
modules:
  gcp:
    enabled: true
    when:
      dir: "**/mycompany/**"
  kube:
    enabled: true
    when:
      dir:
        - "**/mycompany/**"
        - "**/k8s-*/**"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Load(path)

	if cfg.Modules.Gcp.When == nil {
		t.Fatal("gcp.when should not be nil")
	}
	if len(cfg.Modules.Gcp.When.Dir) != 1 {
		t.Errorf("gcp.when.dir: got %d patterns, want 1", len(cfg.Modules.Gcp.When.Dir))
	}
	if cfg.Modules.Kube.When == nil {
		t.Fatal("kube.when should not be nil")
	}
	if len(cfg.Modules.Kube.When.Dir) != 2 {
		t.Errorf("kube.when.dir: got %d patterns, want 2", len(cfg.Modules.Kube.When.Dir))
	}
	if cfg.Modules.Cwd.When != nil {
		t.Error("cwd.when should be nil when not set")
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
