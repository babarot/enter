package modules

import (
	"os"
	"strings"

	"github.com/babarot/enter/internal/module"
	"github.com/goccy/go-yaml"
)

type KubeModule struct{}

func (m *KubeModule) Name() string { return "kube" }

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

	symbol := ctx.Config.Modules.Kube.Symbol

	// Inline: ⎈ context (cluster/namespace)
	var segments []module.Segment
	segments = append(segments, module.NewSegment(symbol, module.Primary))
	segments = append(segments, module.Plain(" "))
	segments = append(segments, module.NewSegment(info.context, module.Primary))
	if info.namespace != "" {
		segments = append(segments, module.NewSegment("/"+info.namespace, module.Muted))
	}

	// Rows
	var rows []module.Row
	rows = append(rows, module.Row{
		Key:      "kube.context",
		Segments: []module.Segment{module.NewSegment(info.context, module.Primary)},
	})
	if info.cluster != "" {
		rows = append(rows, module.Row{
			Key:      "kube.cluster",
			Segments: []module.Segment{module.NewSegment(info.cluster, module.Secondary)},
		})
	}
	if info.namespace != "" {
		rows = append(rows, module.Row{
			Key:      "kube.namespace",
			Segments: []module.Segment{module.NewSegment(info.namespace, module.Accent)},
		})
	}

	return &module.Output{
		Name:     m.Name(),
		Segments: segments,
		Rows:     rows,
		RowOrder: ctx.Config.Modules.Kube.Order,
	}
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
	CurrentContext string            `yaml:"current-context"`
	Contexts       []kubeconfigCtx  `yaml:"contexts"`
}

type kubeconfigCtx struct {
	Name    string             `yaml:"name"`
	Context kubeconfigCtxData  `yaml:"context"`
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

	// Find matching context to get cluster and namespace
	for _, ctx := range doc.Contexts {
		if ctx.Name == doc.CurrentContext {
			info.cluster = ctx.Context.Cluster
			info.namespace = ctx.Context.Namespace
			break
		}
	}

	return info
}
