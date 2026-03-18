package modules

import (
	"os"
	"path/filepath"
	"strings"
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
	t.Setenv("CLOUDSDK_CORE_ACCOUNT", "user@example.com")
	t.Setenv("CLOUDSDK_COMPUTE_REGION", "us-central1")
	t.Setenv("CLOUDSDK_CONFIG_DIR", t.TempDir()) // prevent reading real config

	m := &GcpModule{}
	cfg := config.Default()
	cfg.Modules.Gcp.Enabled = true
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("gcp module with env var should return output")
	}

	// Check rows
	rowMap := make(map[string]string)
	for _, row := range out.Rows {
		rowMap[row.Key] = segmentsText(row.Segments)
	}

	if rowMap["gcp.project"] != "env-project" {
		t.Errorf("project: got %q, want %q", rowMap["gcp.project"], "env-project")
	}
	if rowMap["gcp.account"] != "user@example.com" {
		t.Errorf("account: got %q, want %q", rowMap["gcp.account"], "user@example.com")
	}
	if rowMap["gcp.region"] != "us-central1" {
		t.Errorf("region: got %q, want %q", rowMap["gcp.region"], "us-central1")
	}
}

func TestGcpModuleWithActiveConfig(t *testing.T) {
	t.Setenv("CLOUDSDK_CORE_PROJECT", "")
	t.Setenv("CLOUDSDK_CORE_ACCOUNT", "")
	t.Setenv("CLOUDSDK_COMPUTE_REGION", "")

	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG_DIR", dir)

	// Write active config
	os.WriteFile(filepath.Join(dir, "active_config"), []byte("myprofile"), 0o644)

	// Write config file
	confDir := filepath.Join(dir, "configurations")
	os.MkdirAll(confDir, 0o755)
	os.WriteFile(filepath.Join(confDir, "config_myprofile"), []byte(`[core]
project = profile-project
account = profile@example.com

[compute]
region = asia-northeast1
`), 0o644)

	m := &GcpModule{}
	cfg := config.Default()
	cfg.Modules.Gcp.Enabled = true
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("gcp module with active config should return output")
	}

	rowMap := make(map[string]string)
	for _, row := range out.Rows {
		rowMap[row.Key] = segmentsText(row.Segments)
	}

	if rowMap["gcp.project"] != "profile-project" {
		t.Errorf("project: got %q, want %q", rowMap["gcp.project"], "profile-project")
	}
	if rowMap["gcp.account"] != "profile@example.com" {
		t.Errorf("account: got %q, want %q", rowMap["gcp.account"], "profile@example.com")
	}
	if rowMap["gcp.region"] != "asia-northeast1" {
		t.Errorf("region: got %q, want %q", rowMap["gcp.region"], "asia-northeast1")
	}
}

func TestGcpModuleNonDefaultConfig(t *testing.T) {
	t.Setenv("CLOUDSDK_CORE_PROJECT", "")
	t.Setenv("CLOUDSDK_CORE_ACCOUNT", "")
	t.Setenv("CLOUDSDK_COMPUTE_REGION", "")

	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG_DIR", dir)

	os.WriteFile(filepath.Join(dir, "active_config"), []byte("staging"), 0o644)

	confDir := filepath.Join(dir, "configurations")
	os.MkdirAll(confDir, 0o755)
	os.WriteFile(filepath.Join(confDir, "config_staging"), []byte(`[core]
project = staging-project
`), 0o644)

	m := &GcpModule{}
	cfg := config.Default()
	cfg.Modules.Gcp.Enabled = true
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("gcp module should return output")
	}

	// gcp.config should show "staging" (non-default)
	found := false
	for _, row := range out.Rows {
		if row.Key == "gcp.config" {
			text := segmentsText(row.Segments)
			if text != "staging" {
				t.Errorf("config: got %q, want %q", text, "staging")
			}
			found = true
		}
	}
	if !found {
		t.Error("should have gcp.config row for non-default config")
	}
}

func TestGcpModuleDefaultConfigHidden(t *testing.T) {
	t.Setenv("CLOUDSDK_CORE_PROJECT", "")
	t.Setenv("CLOUDSDK_CORE_ACCOUNT", "")
	t.Setenv("CLOUDSDK_COMPUTE_REGION", "")

	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG_DIR", dir)

	// No active_config file → defaults to "default"
	confDir := filepath.Join(dir, "configurations")
	os.MkdirAll(confDir, 0o755)
	os.WriteFile(filepath.Join(confDir, "config_default"), []byte(`[core]
project = default-project
`), 0o644)

	m := &GcpModule{}
	cfg := config.Default()
	cfg.Modules.Gcp.Enabled = true
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("gcp module should return output")
	}

	// gcp.config should NOT appear for "default"
	for _, row := range out.Rows {
		if row.Key == "gcp.config" {
			t.Error("gcp.config should not appear for default config")
		}
	}
}

func TestGcpModuleInline(t *testing.T) {
	t.Setenv("CLOUDSDK_CORE_PROJECT", "my-project")
	t.Setenv("CLOUDSDK_CORE_ACCOUNT", "me@example.com")
	t.Setenv("CLOUDSDK_COMPUTE_REGION", "")
	t.Setenv("CLOUDSDK_CONFIG_DIR", t.TempDir())

	m := &GcpModule{}
	cfg := config.Default()
	cfg.Modules.Gcp.Enabled = true
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("gcp module should return output")
	}

	text := segmentsText(out.Segments)
	if !strings.Contains(text, "my-project") {
		t.Errorf("inline should contain project, got %q", text)
	}
	if !strings.Contains(text, "me@example.com") {
		t.Errorf("inline should contain account, got %q", text)
	}
}
