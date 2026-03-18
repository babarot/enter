package modules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

func TestReadIniValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "properties")

	content := `[core]
project = my-project
account = test@example.com

[compute]
region = us-central1
`
	os.WriteFile(path, []byte(content), 0o644)

	tests := []struct {
		section, key, want string
	}{
		{"core", "project", "my-project"},
		{"core", "account", "test@example.com"},
		{"compute", "region", "us-central1"},
		{"core", "nonexistent", ""},
		{"nosection", "project", ""},
	}

	for _, tt := range tests {
		got := readIniValue(path, tt.section, tt.key)
		if got != tt.want {
			t.Errorf("readIniValue(%q, %q) = %q, want %q", tt.section, tt.key, got, tt.want)
		}
	}
}

func TestReadIniValueMissingFile(t *testing.T) {
	got := readIniValue("/nonexistent/path", "core", "project")
	if got != "" {
		t.Errorf("missing file should return empty, got %q", got)
	}
}

func TestGcpModuleDisabled(t *testing.T) {
	m := &GcpModule{}
	cfg := config.Default()
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out != nil {
		t.Error("disabled gcp module should return nil")
	}
}

func TestGcpModuleWithEnvVar(t *testing.T) {
	t.Setenv("CLOUDSDK_CORE_PROJECT", "env-project")

	m := &GcpModule{}
	cfg := config.Default()
	cfg.Modules.Gcp.Enabled = true
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("gcp module with env var should return output")
	}

	found := false
	for _, seg := range out.Segments {
		if seg.Text == "env-project" {
			found = true
		}
	}
	if !found {
		t.Error("output should contain project name 'env-project'")
	}
}

func TestGcpModuleWithActiveConfig(t *testing.T) {
	t.Setenv("CLOUDSDK_CORE_PROJECT", "")

	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG_DIR", dir)

	// Write active config
	os.WriteFile(filepath.Join(dir, "active_config"), []byte("myprofile"), 0o644)

	// Write config file
	confDir := filepath.Join(dir, "configurations")
	os.MkdirAll(confDir, 0o755)
	os.WriteFile(filepath.Join(confDir, "config_myprofile"), []byte(`[core]
project = profile-project
`), 0o644)

	m := &GcpModule{}
	cfg := config.Default()
	cfg.Modules.Gcp.Enabled = true
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("gcp module with active config should return output")
	}

	found := false
	for _, seg := range out.Segments {
		if seg.Text == "profile-project" {
			found = true
		}
	}
	if !found {
		t.Error("output should contain project name 'profile-project'")
	}
}
