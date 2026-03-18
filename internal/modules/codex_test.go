package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

func TestCodexModuleName(t *testing.T) {
	m := &CodexModule{}
	if m.Name() != "codex" {
		t.Errorf("Name() = %q, want %q", m.Name(), "codex")
	}
}

func TestCodexModuleDisabled(t *testing.T) {
	m := &CodexModule{}
	cfg := config.Default()
	cfg.Modules.Codex.Enabled = false
	ctx := &module.Context{Cwd: t.TempDir(), Config: cfg}

	out := m.Run(ctx)
	if out != nil {
		t.Error("disabled codex module should return nil")
	}
}

func TestCodexModuleAutoNoCodexFiles(t *testing.T) {
	m := &CodexModule{}
	cfg := config.Default()
	cfg.Modules.Codex.Mode = "auto"
	ctx := &module.Context{Cwd: t.TempDir(), Config: cfg}

	out := m.Run(ctx)
	if out != nil {
		t.Error("auto mode without codex files should return nil")
	}
}

func TestHasCodexFiles(t *testing.T) {
	// No codex files
	dir := t.TempDir()
	if hasCodexFiles(dir) {
		t.Error("empty dir should not have codex files")
	}

	// With .codex directory
	os.Mkdir(filepath.Join(dir, ".codex"), 0o755)
	if !hasCodexFiles(dir) {
		t.Error("dir with .codex should be detected")
	}

	// With AGENTS.md only
	dir2 := t.TempDir()
	os.WriteFile(filepath.Join(dir2, "AGENTS.md"), []byte("# Agents"), 0o644)
	if !hasCodexFiles(dir2) {
		t.Error("dir with AGENTS.md should be detected")
	}

	// With .agents directory only
	dir3 := t.TempDir()
	os.Mkdir(filepath.Join(dir3, ".agents"), 0o755)
	if !hasCodexFiles(dir3) {
		t.Error("dir with .agents should be detected")
	}
}

func TestDetectCodexProject(t *testing.T) {
	// No codex files, not a git repo
	dir := t.TempDir()
	if detectCodexProject(dir) {
		t.Error("empty dir should not detect codex project")
	}

	// With .codex in cwd
	os.Mkdir(filepath.Join(dir, ".codex"), 0o755)
	if !detectCodexProject(dir) {
		t.Error("dir with .codex should detect codex project")
	}
}

func TestDetectCodexProjectGitRoot(t *testing.T) {
	dir := initTestRepo(t)
	subdir := filepath.Join(dir, "sub")
	os.MkdirAll(subdir, 0o755)

	// No codex files anywhere
	if detectCodexProject(subdir) {
		t.Error("should not detect without codex files")
	}

	// Add AGENTS.md at git root
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# test"), 0o644)
	if !detectCodexProject(subdir) {
		t.Error("should detect AGENTS.md at git root from subdir")
	}
}

func TestBuildCodexConfigOutput(t *testing.T) {
	dir := t.TempDir()

	// Empty dir
	segs, row := buildCodexConfigOutput(dir, "auto")
	if segs != nil || row != nil {
		t.Error("empty dir should return nil")
	}

	// With AGENTS.md
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# test"), 0o644)
	segs, row = buildCodexConfigOutput(dir, "auto")
	if segs == nil || row == nil {
		t.Fatal("dir with AGENTS.md should return output")
	}
	if row.Key != "codex.config" {
		t.Errorf("row key: got %q, want %q", row.Key, "codex.config")
	}
}

func TestBuildCodexConfigViewAuto(t *testing.T) {
	dir := t.TempDir()

	// Empty dir — auto returns nothing
	segs := buildCodexConfigView(dir, "auto")
	if len(segs) != 0 {
		t.Error("auto mode with empty dir should return no segments")
	}

	// Add AGENTS.md and skills dir
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# test"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".agents", "skills"), 0o755)
	os.WriteFile(filepath.Join(dir, ".agents", "skills", "s1.md"), []byte("s"), 0o644)

	segs = buildCodexConfigView(dir, "auto")
	text := segmentsText(segs)
	if !strings.Contains(text, "✓") {
		t.Errorf("auto mode should show ✓, got %q", text)
	}
	if !strings.Contains(text, "AGENTS.md") {
		t.Errorf("should contain AGENTS.md, got %q", text)
	}
	if !strings.Contains(text, "skills (1)") {
		t.Errorf("should contain skills (1), got %q", text)
	}
	// Should NOT contain missing items
	if strings.Contains(text, "✗") {
		t.Errorf("auto mode should not show ✗, got %q", text)
	}
}

func TestBuildCodexConfigViewAlways(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# test"), 0o644)

	segs := buildCodexConfigView(dir, "always")
	text := segmentsText(segs)

	// Should show both ✓ and ✗
	if !strings.Contains(text, "✓") {
		t.Errorf("always mode should show ✓, got %q", text)
	}
	if !strings.Contains(text, "✗") {
		t.Errorf("always mode should show ✗ for missing items, got %q", text)
	}
	if !strings.Contains(text, "AGENTS.md") {
		t.Errorf("should contain AGENTS.md, got %q", text)
	}
	if !strings.Contains(text, "config.toml") {
		t.Errorf("always mode should show config.toml, got %q", text)
	}
}

func TestCodexModuleAlwaysMode(t *testing.T) {
	m := &CodexModule{}
	cfg := config.Default()
	cfg.Modules.Codex.Mode = "always"
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# test"), 0o644)
	ctx := &module.Context{Cwd: dir, Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Error("always mode with AGENTS.md should return output")
	}
	if out != nil && out.Name != "codex" {
		t.Errorf("output name: got %q, want %q", out.Name, "codex")
	}
}
