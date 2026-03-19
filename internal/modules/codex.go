package modules

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

type CodexModule struct{}

func (m *CodexModule) Name() string { return config.ModuleCodex }

func (m *CodexModule) Run(ctx *module.Context) *module.Output {
	cfg := &ctx.Config.Modules.Codex
	if !cfg.Enabled {
		return nil
	}

	if cfg.Mode == config.CodexModeAuto && !detectCodexProject(ctx.Cwd) {
		return nil
	}

	if !cfg.Fields.Config.Present() {
		return nil
	}

	segs, row := buildCodexConfigOutput(ctx.Cwd, cfg.Fields.Config.Get().Mode)
	if row == nil {
		return nil
	}

	return &module.Output{
		Name:     m.Name(),
		Segments: segs,
		Rows:     []module.Row{*row},
	}
}

// detectCodexProject checks if cwd (or git root) contains .codex/ or AGENTS.md
func detectCodexProject(cwd string) bool {
	if hasCodexFiles(cwd) {
		return true
	}

	if root, ok := execGit(cwd, "rev-parse", "--show-toplevel"); ok {
		if root != cwd && hasCodexFiles(root) {
			return true
		}
	}

	return false
}

func hasCodexFiles(dir string) bool {
	for _, name := range []string{".codex", "AGENTS.md", ".agents"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}

func buildCodexConfigOutput(cwd, mode string) ([]module.Segment, *module.Row) {
	segs := buildCodexConfigView(cwd, mode)
	if len(segs) == 0 {
		return nil, nil
	}
	return segs, &module.Row{
		Key:      "codex.config",
		Segments: segs,
	}
}

func buildCodexConfigView(cwd, mode string) []module.Segment {
	root := cwd
	if toplevel, ok := execGit(cwd, "rev-parse", "--show-toplevel"); ok {
		root = toplevel
	}

	items := []configItem{
		checkFile(root, "AGENTS.md"),
		checkFile(root, ".codex/config.toml"),
		checkFile(root, ".codex/instructions.md"),
		checkDir(root, ".agents/skills"),
	}

	var segments []module.Segment
	first := true
	for _, item := range items {
		if mode == config.CodexModeAuto && !item.exists {
			continue
		}

		if !first {
			segments = append(segments, module.Plain("\n"))
		}
		first = false

		if item.exists {
			segments = append(segments, module.NewSegment("✓ ", module.Success))
		} else {
			segments = append(segments, module.NewSegment("✗ ", module.Muted))
		}

		label := item.label
		if item.count >= 0 && item.exists {
			label = fmt.Sprintf("%s (%d)", item.label, item.count)
		}

		color := module.Secondary
		if !item.exists {
			color = module.Muted
		}
		segments = append(segments, module.NewSegment(label, color))
	}

	return segments
}
