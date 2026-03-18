package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

func TestCleanContext(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		// GKE
		{"gke full", "gke_my-project_asia-northeast1-a_my-cluster", "my-project/my-cluster"},
		{"gke us", "gke_project_us-central1-b_cluster", "project/cluster"},

		// EKS ARN
		{"eks arn", "arn:aws:eks:us-west-2:123456789:cluster/my-cluster", "my-cluster"},

		// EKS prefix
		{"eks prefix", "eks_my-cluster_us-west-2", "my-cluster"},

		// AKS
		{"aks full", "aks_my-project_eastus_my-cluster", "my-project/my-cluster"},
		{"aks europe", "aks_project_westeu_cluster", "project/cluster"},

		// Passthrough
		{"plain", "my-context", "my-context"},
		{"docker", "docker-desktop", "docker-desktop"},
		{"minikube", "minikube", "minikube"},
	}

	for _, tt := range tests {
		got := cleanContext(tt.input)
		if got != tt.want {
			t.Errorf("%s: cleanContext(%q) = %q, want %q", tt.name, tt.input, got, tt.want)
		}
	}
}

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
	if info.namespace != "" {
		t.Errorf("raw namespace should be empty, got %q", info.namespace)
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

func TestKubeModuleDefaultNamespace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	// No namespace specified
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

	// Should have namespace row with "default"
	for _, row := range out.Rows {
		if row.Key == "kube.namespace" {
			text := segmentsText(row.Segments)
			if text != "default" {
				t.Errorf("namespace should be 'default', got %q", text)
			}
			return
		}
	}
	t.Error("should have kube.namespace row")
}

func TestKubeModuleCleanContext(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	content := `apiVersion: v1
current-context: gke_my-project_asia-northeast1-a_my-cluster
contexts:
- context:
    cluster: gke-cluster
    namespace: prod
  name: gke_my-project_asia-northeast1-a_my-cluster
`
	os.WriteFile(path, []byte(content), 0o644)

	t.Setenv("KUBECONFIG", path)

	m := &KubeModule{}
	cfg := config.Default()
	cfg.Modules.Kube.Enabled = true
	cfg.Modules.Kube.CleanContext = true
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("kube module should return output")
	}

	// Context row should be cleaned
	for _, row := range out.Rows {
		if row.Key == "kube.context" {
			text := segmentsText(row.Segments)
			if strings.Contains(text, "gke_") {
				t.Errorf("clean_context should strip gke_ prefix, got %q", text)
			}
			if !strings.Contains(text, "my-project") {
				t.Errorf("clean_context should contain project, got %q", text)
			}
			return
		}
	}
	t.Error("should have kube.context row")
}

func TestKubeModuleRawContext(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	content := `apiVersion: v1
current-context: gke_my-project_asia-northeast1-a_my-cluster
contexts:
- context:
    cluster: gke-cluster
  name: gke_my-project_asia-northeast1-a_my-cluster
`
	os.WriteFile(path, []byte(content), 0o644)

	t.Setenv("KUBECONFIG", path)

	m := &KubeModule{}
	cfg := config.Default()
	cfg.Modules.Kube.Enabled = true
	cfg.Modules.Kube.CleanContext = false
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("kube module should return output")
	}

	for _, row := range out.Rows {
		if row.Key == "kube.context" {
			text := segmentsText(row.Segments)
			if !strings.Contains(text, "gke_") {
				t.Errorf("raw context should keep gke_ prefix, got %q", text)
			}
			return
		}
	}
	t.Error("should have kube.context row")
}
