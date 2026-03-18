package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

func TestReadKubeInfo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")

	content := `apiVersion: v1
kind: Config
current-context: prod-cluster
contexts:
- context:
    cluster: prod-k8s
    namespace: production
  name: prod-cluster
- context:
    cluster: dev-k8s
    namespace: dev
  name: dev-cluster
`
	os.WriteFile(path, []byte(content), 0o644)

	info := readKubeInfo(path)
	if info == nil {
		t.Fatal("readKubeInfo returned nil")
	}
	if info.context != "prod-cluster" {
		t.Errorf("context: got %q, want %q", info.context, "prod-cluster")
	}
	if info.cluster != "prod-k8s" {
		t.Errorf("cluster: got %q, want %q", info.cluster, "prod-k8s")
	}
	if info.namespace != "production" {
		t.Errorf("namespace: got %q, want %q", info.namespace, "production")
	}
}

func TestReadKubeInfoNoNamespace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")

	content := `apiVersion: v1
current-context: minimal
contexts:
- context:
    cluster: my-cluster
  name: minimal
`
	os.WriteFile(path, []byte(content), 0o644)

	info := readKubeInfo(path)
	if info == nil {
		t.Fatal("readKubeInfo returned nil")
	}
	if info.context != "minimal" {
		t.Errorf("context: got %q, want %q", info.context, "minimal")
	}
	if info.cluster != "my-cluster" {
		t.Errorf("cluster: got %q, want %q", info.cluster, "my-cluster")
	}
	if info.namespace != "" {
		t.Errorf("namespace should be empty, got %q", info.namespace)
	}
}

func TestReadKubeInfoMissing(t *testing.T) {
	info := readKubeInfo("/nonexistent/path")
	if info != nil {
		t.Error("missing file should return nil")
	}
}

func TestReadKubeInfoNoContext(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	os.WriteFile(path, []byte("apiVersion: v1\nkind: Config\n"), 0o644)

	info := readKubeInfo(path)
	if info != nil {
		t.Error("no current-context should return nil")
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
contexts:
- context:
    cluster: test-cluster
    namespace: test-ns
  name: test-ctx
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

	// Check inline segments contain context
	text := segmentsText(out.Segments)
	if !strings.Contains(text, "test-ctx") {
		t.Errorf("inline should contain context name, got %q", text)
	}
	if !strings.Contains(text, "/test-ns") {
		t.Errorf("inline should contain namespace, got %q", text)
	}

	// Check rows
	rowKeys := make(map[string]bool)
	for _, row := range out.Rows {
		rowKeys[row.Key] = true
	}
	for _, key := range []string{"kube.context", "kube.cluster", "kube.namespace"} {
		if !rowKeys[key] {
			t.Errorf("missing row key %q", key)
		}
	}
}

func TestKubeModuleNoNamespace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	content := `apiVersion: v1
current-context: minimal
contexts:
- context:
    cluster: my-cluster
  name: minimal
`
	os.WriteFile(path, []byte(content), 0o644)

	t.Setenv("KUBECONFIG", path)

	m := &KubeModule{}
	cfg := config.Default()
	cfg.Modules.Kube.Enabled = true
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("kube module should return output")
	}

	// namespace row should not exist
	for _, row := range out.Rows {
		if row.Key == "kube.namespace" {
			t.Error("namespace row should not exist when namespace is empty")
		}
	}
}
