package modules

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/babarot/enter/internal/module"
)

type GcpModule struct{}

func (m *GcpModule) Name() string { return "gcp" }

func (m *GcpModule) Run(ctx *module.Context) *module.Output {
	if !ctx.Config.Modules.Gcp.Enabled {
		return nil
	}

	project := getGcpProject()
	if project == "" {
		return nil
	}

	symbol := ctx.Config.Modules.Gcp.Symbol

	return &module.Output{
		Name: m.Name(),
		Segments: []module.Segment{
			module.NewSegment(symbol, module.Accent),
			module.Plain(" "),
			module.NewSegment(project, module.Accent),
		},
	}
}

func getGcpProject() string {
	// 1. Environment variable takes priority
	if project := os.Getenv("CLOUDSDK_CORE_PROJECT"); project != "" {
		return project
	}

	gcloudDir := gcloudConfigDir()
	if gcloudDir == "" {
		return ""
	}

	// 2. Global properties file
	if project := readIniValue(filepath.Join(gcloudDir, "properties"), "core", "project"); project != "" {
		return project
	}

	// 3. Active configuration
	activeConfig := "default"
	if data, err := os.ReadFile(filepath.Join(gcloudDir, "active_config")); err == nil {
		if name := strings.TrimSpace(string(data)); name != "" {
			activeConfig = name
		}
	}

	configPath := filepath.Join(gcloudDir, "configurations", "config_"+activeConfig)
	return readIniValue(configPath, "core", "project")
}

func gcloudConfigDir() string {
	if dir := os.Getenv("CLOUDSDK_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "gcloud")
}

func readIniValue(path, section, key string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	sectionHeader := "[" + section + "]"
	inSection := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") {
			inSection = line == sectionHeader
			continue
		}
		if inSection {
			if k, v, ok := strings.Cut(line, "="); ok {
				if strings.TrimSpace(k) == key {
					val := strings.TrimSpace(v)
					if val != "" {
						return val
					}
				}
			}
		}
	}

	return ""
}
