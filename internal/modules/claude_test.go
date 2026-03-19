package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

func TestBuildBar(t *testing.T) {
	tests := []struct {
		name       string
		pct, width int
		style      string
		want       string
	}{
		{"block 0%", 0, 10, "block", "▱▱▱▱▱▱▱▱▱▱"},
		{"block 50%", 50, 10, "block", "▰▰▰▰▰▱▱▱▱▱"},
		{"block 100%", 100, 10, "block", "▰▰▰▰▰▰▰▰▰▰"},
		{"dot 30%", 30, 10, "dot", "●●●○○○○○○○"},
		{"fill 70%", 70, 10, "fill", "███████░░░"},
		{"block 14%", 14, 10, "block", "▰▱▱▱▱▱▱▱▱▱"},
		{"over 100%", 150, 10, "block", "▰▰▰▰▰▰▰▰▰▰"},
		{"negative", -10, 10, "block", "▱▱▱▱▱▱▱▱▱▱"},
		{"unknown style defaults to block", 50, 10, "unknown", "▰▰▰▰▰▱▱▱▱▱"},
	}

	for _, tt := range tests {
		got := buildBar(tt.pct, tt.width, tt.style)
		if got != tt.want {
			t.Errorf("%s: buildBar(%d, %d, %q) = %q, want %q",
				tt.name, tt.pct, tt.width, tt.style, got, tt.want)
		}
	}
}

func TestPctColor(t *testing.T) {
	tests := []struct {
		pct  int
		want module.SemanticColor
	}{
		{0, module.Success},
		{59, module.Success},
		{60, module.Warning},
		{79, module.Warning},
		{80, module.Danger},
		{100, module.Danger},
	}

	for _, tt := range tests {
		got := pctColor(tt.pct)
		if got != tt.want {
			t.Errorf("pctColor(%d) = %v, want %v", tt.pct, got, tt.want)
		}
	}
}

func TestFormatResetAbsolute(t *testing.T) {
	// Use a fixed time in UTC and test absolute formatting
	ts := "2026-03-19T15:30:00Z"

	// time style
	got := formatReset(ts, "time", "absolute")
	if got == "?" {
		t.Errorf("formatReset time should not return ?, got %q", got)
	}
	// Should contain am/pm
	if !strings.Contains(got, "am") && !strings.Contains(got, "pm") {
		t.Errorf("absolute time should contain am/pm, got %q", got)
	}

	// datetime style
	got = formatReset(ts, "datetime", "absolute")
	if !strings.Contains(got, "Mar") {
		t.Errorf("absolute datetime should contain month, got %q", got)
	}
}

func TestFormatResetRelative(t *testing.T) {
	future := time.Now().Add(2 * time.Hour).Format(time.RFC3339)

	got := formatReset(future, "time", "relative")
	if !strings.Contains(got, "left") {
		t.Errorf("relative should contain 'left', got %q", got)
	}
	if !strings.Contains(got, "h") {
		t.Errorf("2 hours from now should contain 'h', got %q", got)
	}
}

func TestFormatResetEmpty(t *testing.T) {
	if got := formatReset("", "time", "absolute"); got != "?" {
		t.Errorf("empty string should return ?, got %q", got)
	}
	if got := formatReset("invalid", "time", "absolute"); got != "?" {
		t.Errorf("invalid string should return ?, got %q", got)
	}
}

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	// Past time
	got := formatRelativeTime(now.Add(-1 * time.Hour))
	if got != "now" {
		t.Errorf("past: got %q, want %q", got, "now")
	}

	// ~30 minutes
	got = formatRelativeTime(now.Add(31 * time.Minute))
	if !strings.Contains(got, "m left") {
		t.Errorf("30m: should contain 'm left', got %q", got)
	}

	// ~2 hours
	got = formatRelativeTime(now.Add(2*time.Hour + 30*time.Minute))
	if !strings.Contains(got, "2h") || !strings.Contains(got, "left") {
		t.Errorf("2h: should contain '2h' and 'left', got %q", got)
	}

	// ~1 day
	got = formatRelativeTime(now.Add(25 * time.Hour))
	if !strings.Contains(got, "1d") || !strings.Contains(got, "left") {
		t.Errorf("1d: should contain '1d' and 'left', got %q", got)
	}

	// Days should suppress minutes
	got = formatRelativeTime(now.Add(73 * time.Hour))
	if strings.Contains(got, "m ") {
		t.Errorf("days: should not contain minutes, got %q", got)
	}
}

func TestHasClaudeFiles(t *testing.T) {
	// No claude files
	dir := t.TempDir()
	if hasClaudeFiles(dir) {
		t.Error("empty dir should not have claude files")
	}

	// With .claude directory
	os.Mkdir(filepath.Join(dir, ".claude"), 0o755)
	if !hasClaudeFiles(dir) {
		t.Error("dir with .claude should be detected")
	}

	// With CLAUDE.md only
	dir2 := t.TempDir()
	os.WriteFile(filepath.Join(dir2, "CLAUDE.md"), []byte("# Claude"), 0o644)
	if !hasClaudeFiles(dir2) {
		t.Error("dir with CLAUDE.md should be detected")
	}
}

func TestDetectClaudeProject(t *testing.T) {
	// No claude files, not a git repo
	dir := t.TempDir()
	if detectClaudeProject(dir) {
		t.Error("empty dir should not detect claude project")
	}

	// With .claude in cwd
	os.Mkdir(filepath.Join(dir, ".claude"), 0o755)
	if !detectClaudeProject(dir) {
		t.Error("dir with .claude should detect claude project")
	}
}

func TestClaudeModuleDisabled(t *testing.T) {
	m := &ClaudeModule{}
	cfg := config.Default()
	cfg.Modules.Claude.Enabled = false
	ctx := &module.Context{Cwd: t.TempDir(), Config: cfg}

	out := m.Run(ctx)
	if out != nil {
		t.Error("disabled claude module should return nil")
	}
}

func TestClaudeModuleAutoNoClaudeFiles(t *testing.T) {
	m := &ClaudeModule{}
	cfg := config.Default()
	cfg.Modules.Claude.Mode = "auto"
	ctx := &module.Context{Cwd: t.TempDir(), Config: cfg}

	out := m.Run(ctx)
	if out != nil {
		t.Error("auto mode without .claude should return nil")
	}
}

func TestClaudeModuleName(t *testing.T) {
	m := &ClaudeModule{}
	if m.Name() != "claude" {
		t.Errorf("Name() = %q, want %q", m.Name(), "claude")
	}
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{"valid", `{"claudeAiOauth":{"accessToken":"sk-123"}}`, "sk-123"},
		{"null token", `{"claudeAiOauth":{"accessToken":"null"}}`, ""},
		{"empty token", `{"claudeAiOauth":{"accessToken":""}}`, ""},
		{"invalid json", `{invalid`, ""},
		{"missing field", `{"other":"value"}`, ""},
	}
	for _, tt := range tests {
		got := extractToken([]byte(tt.input))
		if got != tt.want {
			t.Errorf("%s: extractToken() = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestGetOAuthTokenFromEnv(t *testing.T) {
	t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "env-token-123")
	got := getOAuthToken()
	if got != "env-token-123" {
		t.Errorf("getOAuthToken() = %q, want %q", got, "env-token-123")
	}
}

func TestGetOAuthTokenEmpty(t *testing.T) {
	t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "")
	// Without keychain or credentials file, should return empty
	// (keychain will fail in test env, credentials file won't exist)
	// Just verify it doesn't panic
	_ = getOAuthToken()
}

func TestBuildWindowRow(t *testing.T) {
	window := &usageWindow{Utilization: 42.0, ResetsAt: "2026-03-19T15:00:00Z"}
	segs, row := buildWindowRow("claude.usage.5h", "current", "time", window, 10, "block", "absolute")

	if len(segs) == 0 {
		t.Error("buildWindowRow should return inline segments")
	}
	if row.Key != "claude.usage.5h" {
		t.Errorf("row key: got %q, want %q", row.Key, "claude.usage.5h")
	}
	text := segmentsText(segs)
	if !strings.Contains(text, "42%") {
		t.Errorf("inline should contain 42%%, got %q", text)
	}
	if !strings.Contains(text, "current") {
		t.Errorf("inline should contain label, got %q", text)
	}
}

func TestBuildConfigOutput(t *testing.T) {
	dir := t.TempDir()

	// Empty dir
	segs, row := buildConfigOutput(dir, "auto")
	if segs != nil || row != nil {
		t.Error("empty dir should return nil")
	}

	// With CLAUDE.md
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# test"), 0o644)
	segs, row = buildConfigOutput(dir, "auto")
	if segs == nil || row == nil {
		t.Fatal("dir with CLAUDE.md should return output")
	}
	if row.Key != "claude.config" {
		t.Errorf("row key: got %q, want %q", row.Key, "claude.config")
	}
}

func TestClaudeModuleAlwaysMode(t *testing.T) {
	m := &ClaudeModule{}
	cfg := config.Default()
	cfg.Modules.Claude.Mode = "always"
	// Disable config view to simplify
	cfg.Modules.Claude.Fields.Config.Enabled = false
	ctx := &module.Context{Cwd: t.TempDir(), Config: cfg}

	// In always mode, Run should not return nil just because of missing .claude
	// (it will return nil only if usage API fails, which it will in test)
	out := m.Run(ctx)
	// Usage API will fail in test env, but config view is disabled
	// so nil is expected — just verify no panic
	_ = out
}

func TestDetectClaudeProjectGitRoot(t *testing.T) {
	dir := initTestRepo(t)
	subdir := filepath.Join(dir, "sub")
	os.MkdirAll(subdir, 0o755)

	// No claude files anywhere
	if detectClaudeProject(subdir) {
		t.Error("should not detect without .claude")
	}

	// Add CLAUDE.md at git root
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# test"), 0o644)
	if !detectClaudeProject(subdir) {
		t.Error("should detect CLAUDE.md at git root from subdir")
	}
}

func TestCheckFile(t *testing.T) {
	dir := t.TempDir()

	// File doesn't exist
	item := checkFile(dir, "CLAUDE.md")
	if item.exists {
		t.Error("should not exist")
	}
	if item.count != -1 {
		t.Errorf("file check should have count -1, got %d", item.count)
	}

	// File exists
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# test"), 0o644)
	item = checkFile(dir, "CLAUDE.md")
	if !item.exists {
		t.Error("should exist")
	}
}

func TestCheckDir(t *testing.T) {
	dir := t.TempDir()

	// Dir doesn't exist
	item := checkDir(dir, "rules")
	if item.exists {
		t.Error("should not exist")
	}
	if item.count != 0 {
		t.Errorf("missing dir should have count 0, got %d", item.count)
	}

	// Dir exists with files
	rulesDir := filepath.Join(dir, "rules")
	os.Mkdir(rulesDir, 0o755)
	os.WriteFile(filepath.Join(rulesDir, "rule1.md"), []byte("rule"), 0o644)
	os.WriteFile(filepath.Join(rulesDir, "rule2.md"), []byte("rule"), 0o644)
	os.WriteFile(filepath.Join(rulesDir, ".hidden"), []byte("hidden"), 0o644)

	item = checkDir(dir, "rules")
	if !item.exists {
		t.Error("should exist")
	}
	if item.count != 2 {
		t.Errorf("should count 2 non-hidden files, got %d", item.count)
	}
}

func TestBuildConfigViewAuto(t *testing.T) {
	dir := t.TempDir()

	// Empty dir — auto returns nothing
	segs := buildConfigView(dir, "auto")
	if len(segs) != 0 {
		t.Error("auto mode with empty dir should return no segments")
	}

	// Add CLAUDE.md and rules
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# test"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".claude", "rules"), 0o755)
	os.WriteFile(filepath.Join(dir, ".claude", "rules", "r1.md"), []byte("r"), 0o644)

	segs = buildConfigView(dir, "auto")
	text := segmentsText(segs)
	if !strings.Contains(text, "✓") {
		t.Errorf("auto mode should show ✓, got %q", text)
	}
	if !strings.Contains(text, "CLAUDE.md") {
		t.Errorf("should contain CLAUDE.md, got %q", text)
	}
	if !strings.Contains(text, "rules (1)") {
		t.Errorf("should contain rules (1), got %q", text)
	}
	// Should NOT contain missing items
	if strings.Contains(text, "✗") {
		t.Errorf("auto mode should not show ✗, got %q", text)
	}
}

func TestBuildConfigViewAlways(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# test"), 0o644)

	segs := buildConfigView(dir, "always")
	text := segmentsText(segs)

	// Should show both ✓ and ✗
	if !strings.Contains(text, "✓") {
		t.Errorf("always mode should show ✓, got %q", text)
	}
	if !strings.Contains(text, "✗") {
		t.Errorf("always mode should show ✗ for missing items, got %q", text)
	}
	if !strings.Contains(text, "CLAUDE.md") {
		t.Errorf("should contain CLAUDE.md, got %q", text)
	}
	if !strings.Contains(text, ".mcp.json") {
		t.Errorf("always mode should show .mcp.json, got %q", text)
	}
}
