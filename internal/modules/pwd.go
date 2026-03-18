package modules

import (
	"os"
	"strings"

	"github.com/babarot/enter/internal/module"
)

type PwdModule struct{}

func (m *PwdModule) Name() string { return "pwd" }

func (m *PwdModule) Run(ctx *module.Context) *module.Output {
	if !ctx.Config.Modules.Pwd.Enabled {
		return nil
	}

	home, _ := os.UserHomeDir()
	display := formatPath(ctx.Cwd, home, ctx.Config.Modules.Pwd.Style)

	return &module.Output{
		Name: m.Name(),
		Segments: []module.Segment{
			module.NewSegment(display, module.Secondary),
		},
	}
}

func formatPath(path, home, style string) string {
	if home != "" && strings.HasPrefix(path, home) {
		path = "~" + path[len(home):]
	}

	switch style {
	case "full":
		return path
	case "short":
		return shortenPath(path)
	case "basename":
		parts := strings.Split(path, "/")
		return parts[len(parts)-1]
	default: // "parent"
		parts := strings.Split(path, "/")
		// Filter empty strings
		var filtered []string
		for _, p := range parts {
			if p != "" {
				filtered = append(filtered, p)
			}
		}
		switch len(filtered) {
		case 0:
			return "/"
		case 1:
			if strings.HasPrefix(path, "~") {
				return "~/" + filtered[0]
			}
			return "/" + filtered[0]
		default:
			return filtered[len(filtered)-2] + "/" + filtered[len(filtered)-1]
		}
	}
}

func shortenPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) <= 2 {
		return path
	}

	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "" || parts[i] == "~" {
			continue
		}
		runes := []rune(parts[i])
		parts[i] = string(runes[0])
	}

	return strings.Join(parts, "/")
}
