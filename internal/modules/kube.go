package modules

import (
	"os"
	"strings"

	"github.com/babarot/enter/internal/module"
	"gopkg.in/yaml.v3"
)

type KubeModule struct{}

func (m *KubeModule) Name() string { return "kube" }

func (m *KubeModule) Run(ctx *module.Context) *module.Output {
	if !ctx.Config.Modules.Kube.Enabled {
		return nil
	}

	context := getCurrentContext()
	if context == "" {
		return nil
	}

	symbol := ctx.Config.Modules.Kube.Symbol

	return &module.Output{
		Name: m.Name(),
		Segments: []module.Segment{
			module.NewSegment(symbol, module.Primary),
			module.Plain(" "),
			module.NewSegment(context, module.Primary),
		},
	}
}

func getCurrentContext() string {
	// Respect KUBECONFIG env var (colon-separated)
	kubeconfig := os.Getenv("KUBECONFIG")
	var paths []string
	if kubeconfig != "" {
		paths = strings.Split(kubeconfig, ":")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		paths = []string{home + "/.kube/config"}
	}

	for _, p := range paths {
		if ctx := readCurrentContext(p); ctx != "" {
			return ctx
		}
	}

	return ""
}

func readCurrentContext(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var doc struct {
		CurrentContext string `yaml:"current-context"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return ""
	}

	return doc.CurrentContext
}
