package modules

import (
	"testing"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

func TestCwdModuleName(t *testing.T) {
	m := &CwdModule{}
	if m.Name() != "cwd" {
		t.Errorf("Name() = %q, want %q", m.Name(), "cwd")
	}
}

func TestCwdModuleEnabled(t *testing.T) {
	m := &CwdModule{}
	cfg := config.Default()
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("cwd module should return output")
	}
	if out.Name != "cwd" {
		t.Errorf("Name: got %q, want %q", out.Name, "cwd")
	}
	if len(out.Segments) == 0 {
		t.Error("should have segments")
	}
}

func TestCwdModuleDisabled(t *testing.T) {
	m := &CwdModule{}
	cfg := config.Default()
	cfg.Modules.Cwd.Enabled = false
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out != nil {
		t.Error("disabled cwd module should return nil")
	}
}

func TestCwdModuleStyles(t *testing.T) {
	m := &CwdModule{}

	styles := []string{"short", "parent", "full", "basename"}
	for _, style := range styles {
		cfg := config.Default()
		cfg.Modules.Cwd.Style = style
		ctx := &module.Context{Cwd: "/tmp/test/dir", Config: cfg}

		out := m.Run(ctx)
		if out == nil {
			t.Errorf("style %q: should return output", style)
			continue
		}
		if len(out.Segments) == 0 {
			t.Errorf("style %q: should have segments", style)
		}
	}
}

func TestFormatPath(t *testing.T) {
	home := "/Users/test"

	tests := []struct {
		name, path, style, want string
	}{
		// parent style
		{"parent basic", "/Users/test/src/project", "parent", "src/project"},
		{"parent home", "/Users/test/project", "parent", "~/project"},
		{"parent root", "/", "parent", "/"},
		{"parent single", "/Users/test/dir", "parent", "~/dir"},

		// full style
		{"full basic", "/Users/test/src/project", "full", "~/src/project"},
		{"full no home", "/opt/data", "full", "/opt/data"},

		// short style
		{"short deep", "/Users/test/src/github/com/babarot/enter", "short", "~/s/g/c/babarot/enter"},
		{"short 4 parts", "/Users/test/src/github/project", "short", "~/s/github/project"},
		{"short 3 parts", "/Users/test/src/project", "short", "~/src/project"},
		{"short shallow", "/Users/test/dir", "short", "~/dir"},

		// basename style
		{"basename basic", "/Users/test/src/project", "basename", "project"},
		{"basename root", "/", "basename", ""},
	}

	for _, tt := range tests {
		got := formatPath(tt.path, home, tt.style)
		if got != tt.want {
			t.Errorf("%s: formatPath(%q, %q, %q) = %q, want %q",
				tt.name, tt.path, home, tt.style, got, tt.want)
		}
	}
}

func TestShortenPath(t *testing.T) {
	tests := []struct {
		path, want string
	}{
		{"~/src/github/com/babarot/enter", "~/s/g/c/babarot/enter"},
		{"~/src/github/project", "~/s/github/project"},
		{"~/src/project", "~/src/project"},
		{"~/dir", "~/dir"},
		{"/a/b/c", "/a/b/c"},
		{"/usr/local/bin", "/u/local/bin"},
		{"/usr/local/share/bin", "/u/l/share/bin"},
	}

	for _, tt := range tests {
		got := shortenPath(tt.path)
		if got != tt.want {
			t.Errorf("shortenPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}
