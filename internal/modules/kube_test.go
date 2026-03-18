package modules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

func TestReadCurrentContext(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")

	content := `apiVersion: v1
kind: Config
current-context: prod-cluster
contexts:
- context:
    cluster: prod
  name: prod-cluster
`
	os.WriteFile(path, []byte(content), 0o644)

	ctx := readCurrentContext(path)
	if ctx != "prod-cluster" {
		t.Errorf("got %q, want %q", ctx, "prod-cluster")
	}
}

func TestReadCurrentContextMissing(t *testing.T) {
	ctx := readCurrentContext("/nonexistent/path")
	if ctx != "" {
		t.Errorf("missing file should return empty, got %q", ctx)
	}
}

func TestReadCurrentContextNoField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	os.WriteFile(path, []byte("apiVersion: v1\nkind: Config\n"), 0o644)

	ctx := readCurrentContext(path)
	if ctx != "" {
		t.Errorf("no current-context should return empty, got %q", ctx)
	}
}

func TestKubeModuleDisabled(t *testing.T) {
	m := &KubeModule{}
	cfg := config.Default()
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out != nil {
		t.Error("disabled kube module should return nil")
	}
}

func TestKubeModuleWithKUBECONFIG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	content := `apiVersion: v1
current-context: test-ctx
`
	os.WriteFile(path, []byte(content), 0o644)

	t.Setenv("KUBECONFIG", path)

	m := &KubeModule{}
	cfg := config.Default()
	cfg.Modules.Kube.Enabled = true
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("kube module with valid KUBECONFIG should return output")
	}

	found := false
	for _, seg := range out.Segments {
		if seg.Text == "test-ctx" {
			found = true
		}
	}
	if !found {
		t.Error("output should contain context name 'test-ctx'")
	}
}
