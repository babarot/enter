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
	if !cfg.Modules.Git.Fields.Url.Present() {
		t.Error("git url field should be present")
	}
	if !cfg.Modules.Git.Fields.Summary.Present() {
		t.Error("git summary field should be present")
	}
	if cfg.Modules.Git.Fields.Cwd.Present() {
		t.Error("git cwd field should not be present (not in YAML)")
	}
	if cfg.Modules.Git.Fields.Status.Present() {
		t.Error("git status field should not be present (not in YAML)")
	}
	if cfg.Modules.Git.Fields.Summary.Get().Symbols.Unstaged != "!" {
		t.Errorf("unstaged: got %q, want %q", cfg.Modules.Git.Fields.Summary.Get().Symbols.Unstaged, "!")
	}
	// Empty symbols should be filled with defaults
	if cfg.Modules.Git.Fields.Summary.Get().Symbols.Staged != "+" {
		t.Errorf("staged should default to +, got %q", cfg.Modules.Git.Fields.Summary.Get().Symbols.Staged)
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
	order, _, _ := extractOrder([]byte("{{invalid"))
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
	cfg.Modules.Git.Fields.Cwd.Set(GitCwdConfig{Style: "invalid"})
	cfg.Modules.Git.Fields.Status.Set(GitStatusConfig{Style: "invalid"})
	cfg.Modules.Claude.Mode = "invalid"
	cfg.Modules.Claude.Fields.Usage.Set(ClaudeUsageConfig{BarStyle: "invalid", TimeStyle: "invalid", CacheTTL: -1})

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
	if cfg.Modules.Git.Fields.Cwd.Get().Style != d.Modules.Git.Fields.Cwd.Get().Style {
		t.Errorf("git.cwd.style: got %q, want %q", cfg.Modules.Git.Fields.Cwd.Get().Style, d.Modules.Git.Fields.Cwd.Get().Style)
	}
	if cfg.Modules.Git.Fields.Status.Get().Style != d.Modules.Git.Fields.Status.Get().Style {
		t.Errorf("git.status.style: got %q, want %q", cfg.Modules.Git.Fields.Status.Get().Style, d.Modules.Git.Fields.Status.Get().Style)
	}
	if cfg.Modules.Claude.Mode != d.Modules.Claude.Mode {
		t.Errorf("claude.mode: got %q, want %q", cfg.Modules.Claude.Mode, d.Modules.Claude.Mode)
	}
	if cfg.Modules.Claude.Fields.Usage.Get().BarStyle != d.Modules.Claude.Fields.Usage.Get().BarStyle {
		t.Errorf("claude.usage.bar_style: got %q, want %q", cfg.Modules.Claude.Fields.Usage.Get().BarStyle, d.Modules.Claude.Fields.Usage.Get().BarStyle)
	}
	if cfg.Modules.Claude.Fields.Usage.Get().TimeStyle != d.Modules.Claude.Fields.Usage.Get().TimeStyle {
		t.Errorf("claude.usage.time_style: got %q, want %q", cfg.Modules.Claude.Fields.Usage.Get().TimeStyle, d.Modules.Claude.Fields.Usage.Get().TimeStyle)
	}
	if cfg.Modules.Claude.Fields.Usage.Get().CacheTTL != d.Modules.Claude.Fields.Usage.Get().CacheTTL {
		t.Errorf("claude.usage.cache_ttl: got %d, want %d", cfg.Modules.Claude.Fields.Usage.Get().CacheTTL, d.Modules.Claude.Fields.Usage.Get().CacheTTL)
	}
}

func TestValidateValidValues(t *testing.T) {
	cfg := Default()
	cfg.Format = "inline"
	cfg.Trigger = "on_cd"
	cfg.KeyStyle = "flat"
	cfg.Modules.Git.Fields.Cwd.Set(GitCwdConfig{Style: "breadcrumb"})
	cfg.Modules.Git.Fields.Status.Set(GitStatusConfig{Style: "long"})
	cfg.Modules.Claude.Mode = "always"
	cfg.Modules.Claude.Fields.Usage.Set(ClaudeUsageConfig{BarStyle: "dot", TimeStyle: "relative", CacheTTL: 60})

	cfg.validate()

	// All should remain as set
	if cfg.Format != "inline" {
		t.Errorf("format should stay inline, got %q", cfg.Format)
	}
	if cfg.Modules.Claude.Fields.Usage.Get().BarStyle != "dot" {
		t.Errorf("bar_style should stay dot, got %q", cfg.Modules.Claude.Fields.Usage.Get().BarStyle)
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
	_, subKeyOrder, _ := extractOrder([]byte(content))
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

func TestDefaultFieldsAllPresent(t *testing.T) {
	cfg := Default()
	if !cfg.Modules.Git.Fields.Url.Present() {
		t.Error("default git.url should be present")
	}
	if !cfg.Modules.Git.Fields.Cwd.Present() {
		t.Error("default git.cwd should be present")
	}
	if !cfg.Modules.Git.Fields.Summary.Present() {
		t.Error("default git.summary should be present")
	}
	if !cfg.Modules.Git.Fields.Status.Present() {
		t.Error("default git.status should be present")
	}
	if !cfg.Modules.Kube.Fields.Context.Present() {
		t.Error("default kube.context should be present")
	}
	if !cfg.Modules.Claude.Fields.Usage.Present() {
		t.Error("default claude.usage should be present")
	}
	if !cfg.Modules.Claude.Fields.Config.Present() {
		t.Error("default claude.config should be present")
	}
	if !cfg.Modules.Codex.Fields.Config.Present() {
		t.Error("default codex.config should be present")
	}
}

func TestLoadWithoutFieldsRestoresDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	// No "fields:" key at all — should get all defaults
	content := `
modules:
  git:
    enabled: true
`
	os.WriteFile(path, []byte(content), 0o644)

	cfg := Load(path)
	if !cfg.Modules.Git.Fields.Url.Present() {
		t.Error("git.url should be present when fields: not specified")
	}
	if !cfg.Modules.Git.Fields.Summary.Present() {
		t.Error("git.summary should be present when fields: not specified")
	}
	if !cfg.Modules.Git.Fields.Cwd.Present() {
		t.Error("git.cwd should be present when fields: not specified")
	}
	if !cfg.Modules.Git.Fields.Status.Present() {
		t.Error("git.status should be present when fields: not specified")
	}
}

func TestLoadWithEmptyKeyField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	// "url:" is an empty key — should still be Present via ensureFieldsPresent
	content := `
modules:
  git:
    enabled: true
    fields:
      url:
      summary:
        symbols:
          unstaged: "!"
`
	os.WriteFile(path, []byte(content), 0o644)

	cfg := Load(path)
	if !cfg.Modules.Git.Fields.Url.Present() {
		t.Error("empty-key url should be present")
	}
	if !cfg.Modules.Git.Fields.Summary.Present() {
		t.Error("summary should be present")
	}
	if cfg.Modules.Git.Fields.Cwd.Present() {
		t.Error("cwd should not be present (omitted)")
	}
	if cfg.Modules.Git.Fields.Status.Present() {
		t.Error("status should not be present (omitted)")
	}
}

func TestLoadWithEmptyFieldsBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	// "fields:" is present but empty — all fields should be not present
	content := `
modules:
  git:
    enabled: true
    fields:
`
	os.WriteFile(path, []byte(content), 0o644)

	cfg := Load(path)
	if cfg.Modules.Git.Fields.Url.Present() {
		t.Error("url should not be present with empty fields block")
	}
	if cfg.Modules.Git.Fields.Cwd.Present() {
		t.Error("cwd should not be present with empty fields block")
	}
	if cfg.Modules.Git.Fields.Summary.Present() {
		t.Error("summary should not be present with empty fields block")
	}
	if cfg.Modules.Git.Fields.Status.Present() {
		t.Error("status should not be present with empty fields block")
	}
}

func TestExtractOrderFieldsPresent(t *testing.T) {
	content := `
modules:
  git:
    fields:
      url:
      summary:
  kube:
    enabled: true
`
	_, _, fieldsPresent := extractOrder([]byte(content))
	if !fieldsPresent["git"] {
		t.Error("git should have fieldsPresent=true")
	}
	if fieldsPresent["kube"] {
		t.Error("kube should have fieldsPresent=false (no fields: key)")
	}
}

func TestValidateSkipsNotPresentFields(t *testing.T) {
	cfg := Default()
	// Make cwd not present, then validate should not touch it
	cfg.Modules.Git.Fields.Cwd = Field[GitCwdConfig]{}
	cfg.validate()
	if cfg.Modules.Git.Fields.Cwd.Present() {
		t.Error("validate should not make absent fields present")
	}
}

func TestEnsureFieldsPresent(t *testing.T) {
	var fields GitFields
	// Simulate: extractOrder found "url" and "status" in YAML, but they were empty keys
	ensureFieldsPresent([]string{"url", "status"}, &fields)
	if !fields.Url.Present() {
		t.Error("url should be marked present")
	}
	if !fields.Status.Present() {
		t.Error("status should be marked present")
	}
	if fields.Cwd.Present() {
		t.Error("cwd should remain not present")
	}
	if fields.Summary.Present() {
		t.Error("summary should remain not present")
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

func boolPtr(b bool) *bool { return &b }

func TestWhenGitRepoTrue(t *testing.T) {
	// This test runs inside the enter repo, so git_repo: true should match.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	w := &When{GitRepo: boolPtr(true)}
	if !w.Match(cwd) {
		t.Error("git_repo: true should match inside a git repo")
	}
}

func TestWhenGitRepoFalse(t *testing.T) {
	// A temp dir is not a git repo, so git_repo: false should match.
	dir := t.TempDir()
	w := &When{GitRepo: boolPtr(false)}
	if !w.Match(dir) {
		t.Error("git_repo: false should match outside a git repo")
	}
}

func TestWhenGitRepoFalseInsideRepo(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	w := &When{GitRepo: boolPtr(false)}
	if w.Match(cwd) {
		t.Error("git_repo: false should NOT match inside a git repo")
	}
}

func TestWhenGitRepoAndDir(t *testing.T) {
	// Both conditions must match (AND logic).
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	w := &When{
		Dir:     StringOrSlice{"/nonexistent/**"},
		GitRepo: boolPtr(true),
	}
	if w.Match(cwd) {
		t.Error("should not match when dir pattern does not match")
	}
}

func TestLoadConfigWithWhenGitRepo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
modules:
  cwd:
    enabled: true
    when:
      git_repo: true
  gcp:
    enabled: true
    when:
      git_repo: false
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Load(path)
	if cfg.Modules.Cwd.When == nil || cfg.Modules.Cwd.When.GitRepo == nil {
		t.Fatal("cwd.when.git_repo should not be nil")
	}
	if *cfg.Modules.Cwd.When.GitRepo != true {
		t.Error("cwd.when.git_repo should be true")
	}
	if cfg.Modules.Gcp.When == nil || cfg.Modules.Gcp.When.GitRepo == nil {
		t.Fatal("gcp.when.git_repo should not be nil")
	}
	if *cfg.Modules.Gcp.When.GitRepo != false {
		t.Error("gcp.when.git_repo should be false")
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
