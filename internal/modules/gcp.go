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

type gcpInfo struct {
	project string
	account string
	region  string
	config  string // active configuration name
}

func (m *GcpModule) Run(ctx *module.Context) *module.Output {
	if !ctx.Config.Modules.Gcp.Enabled {
		return nil
	}

	info := getGcpInfo()
	if info == nil {
		return nil
	}

	// Inline: project (account) or account
	var segments []module.Segment
	if info.project != "" {
		segments = append(segments, module.NewSegment(info.project, module.Accent))
		if info.account != "" {
			segments = append(segments, module.NewSegment(" ("+info.account+")", module.Muted))
		}
	} else if info.account != "" {
		segments = append(segments, module.NewSegment(info.account, module.Accent))
	}

	// Rows
	var rows []module.Row
	if info.project != "" {
		rows = append(rows, module.Row{
			Key:      "gcp.project",
			Segments: []module.Segment{module.NewSegment(info.project, module.Accent)},
		})
	}
	if info.account != "" {
		rows = append(rows, module.Row{
			Key:      "gcp.account",
			Segments: []module.Segment{module.NewSegment(info.account, module.Muted)},
		})
	}
	if info.region != "" {
		rows = append(rows, module.Row{
			Key:      "gcp.region",
			Segments: []module.Segment{module.NewSegment(info.region, module.Secondary)},
		})
	}
	if info.config != "" && info.config != "default" {
		rows = append(rows, module.Row{
			Key:      "gcp.config",
			Segments: []module.Segment{module.NewSegment(info.config, module.Muted)},
		})
	}

	return &module.Output{
		Name:     m.Name(),
		Segments: segments,
		Rows:     rows,
	}
}

func getGcpInfo() *gcpInfo {
	info := &gcpInfo{}

	// 1. Environment variables take priority
	info.project = os.Getenv("CLOUDSDK_CORE_PROJECT")
	info.account = os.Getenv("CLOUDSDK_CORE_ACCOUNT")
	info.region = os.Getenv("CLOUDSDK_COMPUTE_REGION")

	gcloudDir := gcloudConfigDir()
	if gcloudDir == "" {
		if info.project != "" {
			return info
		}
		return nil
	}

	// 2. Determine active config
	activeConfig := "default"
	if data, err := os.ReadFile(filepath.Join(gcloudDir, "active_config")); err == nil {
		if name := strings.TrimSpace(string(data)); name != "" {
			activeConfig = name
		}
	}
	info.config = activeConfig

	// 3. Read from active config file
	configPath := filepath.Join(gcloudDir, "configurations", "config_"+activeConfig)

	if info.project == "" {
		info.project = readIniValue(configPath, "core", "project")
	}
	if info.account == "" {
		info.account = readIniValue(configPath, "core", "account")
	}
	if info.region == "" {
		info.region = readIniValue(configPath, "compute", "region")
	}

	// 4. Fallback to global properties
	propsPath := filepath.Join(gcloudDir, "properties")
	if info.project == "" {
		info.project = readIniValue(propsPath, "core", "project")
	}
	if info.account == "" {
		info.account = readIniValue(propsPath, "core", "account")
	}
	if info.region == "" {
		info.region = readIniValue(propsPath, "compute", "region")
	}

	// Return nil only if nothing was found
	if info.project == "" && info.account == "" && info.region == "" {
		return nil
	}

	return info
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
