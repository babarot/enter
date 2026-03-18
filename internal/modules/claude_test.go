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
