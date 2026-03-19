package modules

import (
	"os"
	"regexp"
	"strings"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
	"github.com/goccy/go-yaml"
)

type KubeModule struct{}

func (m *KubeModule) Name() string { return config.ModuleKube }

type kubeInfo struct {
	context   string
	cluster   string
	namespace string
}

func (m *KubeModule) Run(ctx *module.Context) *module.Output {
	if !ctx.Config.Modules.Kube.Enabled {
		return nil
	}

	info := getKubeInfo()
	if info == nil {
		return nil
	}

	// Default namespace
	if info.namespace == "" {
		info.namespace = "default"
	}

	// Clean context name
	displayContext := info.context
	if ctx.Config.Modules.Kube.Fields.Context.Clean {
		displayContext = cleanContext(info.context)
	}

	// Inline: context/namespace
	segments := []module.Segment{
		module.NewSegment(displayContext, module.Primary),
		module.NewSegment("/"+info.namespace, module.Muted),
	}

	// Rows
	rows := []module.Row{
		{
			Key:      "kube.context",
			Segments: []module.Segment{module.NewSegment(displayContext, module.Primary)},
		},
		{
			Key:      "kube.namespace",
			Segments: []module.Segment{module.NewSegment(info.namespace, module.Accent)},
		},
	}
	if info.cluster != "" {
		rows = append(rows, module.Row{
			Key:      "kube.cluster",
			Segments: []module.Segment{module.NewSegment(info.cluster, module.Secondary)},
		})
	}

	return &module.Output{
		Name:     m.Name(),
		Segments: segments,
		Rows:     rows,
	}
}

// cleanContext strips cloud provider prefixes and regions from context names.
//
//	GKE: gke_project_asia-northeast1-a_cluster → project/cluster
//	EKS: arn:aws:eks:us-west-2:123456:cluster/name → name
//	AKS: aks_project_eastus_cluster → project/cluster
//	Other: returned as-is
func cleanContext(ctx string) string {
	switch {
	case strings.HasPrefix(ctx, "gke_"):
		return cleanGKE(ctx)
	case strings.HasPrefix(ctx, "arn:aws:eks:") || strings.HasPrefix(ctx, "eks_"):
		return cleanEKS(ctx)
	case strings.HasPrefix(ctx, "aks_"):
		return cleanAKS(ctx)
	default:
		return ctx
	}
}

var gkeRegionRe = regexp.MustCompile(`_[a-z]+-[a-z]+\d+-[a-z]_?`)

func cleanGKE(ctx string) string {
	// gke_project_asia-northeast1-a_cluster → project/cluster
	ctx = strings.TrimPrefix(ctx, "gke_")
	ctx = gkeRegionRe.ReplaceAllString(ctx, "/")
	ctx = strings.TrimSuffix(ctx, "/")
	return ctx
}

func cleanEKS(ctx string) string {
	// arn:aws:eks:us-west-2:123456:cluster/name → name
	if strings.HasPrefix(ctx, "arn:") {
		if idx := strings.LastIndex(ctx, "/"); idx >= 0 {
			return ctx[idx+1:]
		}
	}
	// eks_name_region → name
	ctx = strings.TrimPrefix(ctx, "eks_")
	parts := strings.SplitN(ctx, "_", 2)
	return parts[0]
}

var aksRegionRe = regexp.MustCompile(`_(east|west|central|north|south)(us|eu|asia|uk|au|in|japan|korea|canada|france|germany|brazil|uae)\d*_`)

func cleanAKS(ctx string) string {
	// aks_project_eastus_cluster → project/cluster
	ctx = strings.TrimPrefix(ctx, "aks_")
	ctx = aksRegionRe.ReplaceAllString(ctx, "/")
	return ctx
}

func getKubeInfo() *kubeInfo {
	paths := kubeconfigPaths()

	for _, p := range paths {
		if info := readKubeInfo(p); info != nil {
			return info
		}
	}

	return nil
}

func kubeconfigPaths() []string {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return strings.Split(kubeconfig, ":")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{home + "/.kube/config"}
}

// kubeconfig YAML structure (minimal)
type kubeconfigDoc struct {
	CurrentContext string           `yaml:"current-context"`
	Contexts       []kubeconfigCtx `yaml:"contexts"`
}

type kubeconfigCtx struct {
	Name    string            `yaml:"name"`
	Context kubeconfigCtxData `yaml:"context"`
}

type kubeconfigCtxData struct {
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace"`
}

func readKubeInfo(path string) *kubeInfo {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var doc kubeconfigDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil
	}

	if doc.CurrentContext == "" {
		return nil
	}

	info := &kubeInfo{context: doc.CurrentContext}

	for _, ctx := range doc.Contexts {
		if ctx.Name == doc.CurrentContext {
			info.cluster = ctx.Context.Cluster
			info.namespace = ctx.Context.Namespace
			break
		}
	}

	return info
}
