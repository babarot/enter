package modules

import (
	"testing"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

func TestLsModuleName(t *testing.T) {
	m := &LsModule{}
	if m.Name() != "ls" {
		t.Errorf("got %q, want %q", m.Name(), "ls")
	}
}

func TestLsModuleDisabled(t *testing.T) {
	cfg := config.Default()
	cfg.Modules.Ls.Enabled = false
	ctx := &module.Context{Cwd: t.TempDir(), Config: cfg}
	m := &LsModule{}
	if out := m.Run(ctx); out != nil {
		t.Error("disabled module should return nil")
	}
}

func TestLsModuleEmptyCmd(t *testing.T) {
	cfg := config.Default()
	cfg.Modules.Ls.Enabled = true
	cfg.Modules.Ls.Cmd = ""
	ctx := &module.Context{Cwd: t.TempDir(), Config: cfg}
	m := &LsModule{}
	if out := m.Run(ctx); out != nil {
		t.Error("empty cmd should return nil")
	}
}

func TestLsModuleEcho(t *testing.T) {
	cfg := config.Default()
	cfg.Modules.Ls.Enabled = true
	cfg.Modules.Ls.Cmd = "echo hello"
	ctx := &module.Context{Cwd: t.TempDir(), Config: cfg}
	m := &LsModule{}
	out := m.Run(ctx)
	if out == nil {
		t.Fatal("expected output, got nil")
	}
	if len(out.Segments) != 1 || out.Segments[0].Text != "hello" {
		t.Errorf("got %q, want %q", out.Segments[0].Text, "hello")
	}
}

func TestLsModuleUsesWorkingDir(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default()
	cfg.Modules.Ls.Enabled = true
	cfg.Modules.Ls.Cmd = "pwd"
	ctx := &module.Context{Cwd: dir, Config: cfg}
	m := &LsModule{}
	out := m.Run(ctx)
	if out == nil {
		t.Fatal("expected output, got nil")
	}
	if out.Segments[0].Text != dir {
		t.Errorf("got %q, want %q", out.Segments[0].Text, dir)
	}
}

func TestLsModuleFailingCommand(t *testing.T) {
	cfg := config.Default()
	cfg.Modules.Ls.Enabled = true
	cfg.Modules.Ls.Cmd = "false"
	ctx := &module.Context{Cwd: t.TempDir(), Config: cfg}
	m := &LsModule{}
	out := m.Run(ctx)
	if out == nil {
		t.Fatal("failing command should return error output")
	}
	if out.Segments[0].Color != module.Danger {
		t.Error("failing command should use Danger color")
	}
}
