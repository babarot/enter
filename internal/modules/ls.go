package modules

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

type LsModule struct{}

func (m *LsModule) Name() string { return config.ModuleLs }

func (m *LsModule) Run(ctx *module.Context) *module.Output {
	if !ctx.Config.Modules.Ls.Enabled {
		return nil
	}

	cmdStr := ctx.Config.Modules.Ls.Cmd
	if cmdStr == "" {
		return nil
	}

	out, stderr, err := execShell(ctx.Cwd, cmdStr)
	if err != nil {
		errMsg := stderr
		if errMsg == "" {
			errMsg = err.Error()
		}
		return &module.Output{
			Name:     m.Name(),
			Segments: []module.Segment{module.NewSegment(errMsg, module.Danger)},
		}
	}
	if out == "" {
		return nil
	}

	return &module.Output{
		Name: m.Name(),
		Segments: []module.Segment{
			module.NewSegment(out, module.Default),
		},
	}
}

func execShell(cwd, cmdStr string) (stdout, stderr string, err error) {
	// Normalize YAML block scalar: collapse newlines into spaces.
	cmdStr = strings.TrimSpace(strings.ReplaceAll(cmdStr, "\n", " "))
	if cmdStr == "" {
		return "", "", nil
	}
	shell := fmt.Sprintf("cd %s && %s", shellescape(cwd), cmdStr)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", shell)
	cmd.Stdin = nil
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf
	out, err := cmd.Output()
	return strings.TrimRight(string(out), "\n"), strings.TrimSpace(stderrBuf.String()), err
}

// shellescape wraps s in single quotes, escaping any embedded single quotes.
func shellescape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
