package modules

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/babarot/enter/internal/module"
)

type GitModule struct{}

func (m *GitModule) Name() string { return "git" }

type gitInfo struct {
	branch    string
	detached  bool
	unstaged  bool
	staged    bool
	stash     bool
	untracked bool
	ahead     int
	behind    int
	operation string
}

func (m *GitModule) Run(ctx *module.Context) *module.Output {
	if !ctx.Config.Modules.Git.Enabled {
		return nil
	}

	info := getGitInfo(ctx.Cwd)
	if info == nil || info.branch == "" {
		return nil
	}

	symbols := &ctx.Config.Modules.Git.Symbols
	var segments []module.Segment

	branchColor := module.Success
	if info.detached {
		branchColor = module.Danger
	}

	segments = append(segments, module.NewSegment("(", branchColor))
	segments = append(segments, module.NewSegment(info.branch, branchColor))

	// State flags
	var flags []module.Segment
	if info.unstaged {
		flags = append(flags, module.NewSegment(symbols.Unstaged, module.Danger))
	}
	if info.staged {
		flags = append(flags, module.NewSegment(symbols.Staged, module.Success))
	}
	if info.stash {
		flags = append(flags, module.NewSegment(symbols.Stash, module.Primary))
	}
	if info.untracked {
		flags = append(flags, module.NewSegment(symbols.Untracked, module.Danger))
	}
	if len(flags) > 0 {
		segments = append(segments, module.Plain(" "))
		segments = append(segments, flags...)
	}

	// Ahead/behind
	var upstream []module.Segment
	if info.ahead > 0 {
		upstream = append(upstream, module.NewSegment(
			fmt.Sprintf("%s%d", symbols.Ahead, info.ahead), module.Success))
	}
	if info.behind > 0 {
		upstream = append(upstream, module.NewSegment(
			fmt.Sprintf("%s%d", symbols.Behind, info.behind), module.Danger))
	}
	if len(upstream) > 0 {
		segments = append(segments, module.Plain(" "))
		for i, seg := range upstream {
			if i > 0 {
				segments = append(segments, module.Plain(" "))
			}
			segments = append(segments, seg)
		}
	}

	// Operation
	if info.operation != "" {
		segments = append(segments, module.NewSegment("|"+info.operation, module.Accent))
	}

	segments = append(segments, module.NewSegment(")", branchColor))

	return &module.Output{
		Name:     m.Name(),
		Segments: segments,
	}
}

func execGit(cwd string, args ...string) (string, bool) {
	allArgs := append([]string{"-C", cwd}, args...)
	cmd := exec.Command("git", allArgs...)
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}

func getGitInfo(cwd string) *gitInfo {
	status, ok := execGit(cwd, "status", "--porcelain=v2", "--branch", "--show-stash")
	if !ok {
		return nil
	}

	info := &gitInfo{}
	var oid string

	for _, line := range strings.Split(status, "\n") {
		switch {
		case strings.HasPrefix(line, "# branch.head "):
			head := strings.TrimPrefix(line, "# branch.head ")
			if head == "(detached)" {
				info.detached = true
			} else {
				info.branch = head
			}

		case strings.HasPrefix(line, "# branch.ab "):
			ab := strings.TrimPrefix(line, "# branch.ab ")
			for _, part := range strings.Fields(ab) {
				if strings.HasPrefix(part, "+") {
					fmt.Sscanf(part, "+%d", &info.ahead)
				} else if strings.HasPrefix(part, "-") {
					fmt.Sscanf(part, "-%d", &info.behind)
				}
			}

		case strings.HasPrefix(line, "# branch.oid "):
			oid = strings.TrimPrefix(line, "# branch.oid ")

		case strings.HasPrefix(line, "# stash "):
			info.stash = true

		case strings.HasPrefix(line, "1 ") || strings.HasPrefix(line, "2 "):
			fields := strings.Fields(line)
			if len(fields) >= 2 && len(fields[1]) >= 2 {
				xy := fields[1]
				if xy[0] != '.' {
					info.staged = true
				}
				if xy[1] != '.' {
					info.unstaged = true
				}
			}

		case strings.HasPrefix(line, "u "):
			info.staged = true
			info.unstaged = true

		case strings.HasPrefix(line, "? "):
			info.untracked = true
		}
	}

	// Detached HEAD — describe with tag or short SHA
	if info.detached {
		if desc, ok := execGit(cwd, "describe", "--tags", "--exact-match", "HEAD"); ok {
			info.branch = fmt.Sprintf("(%s)", desc)
		} else if desc, ok := execGit(cwd, "describe", "--contains", "--all", "HEAD"); ok {
			info.branch = fmt.Sprintf("(%s)", desc)
		} else {
			short := oid
			if len(short) > 7 {
				short = short[:7]
			}
			info.branch = fmt.Sprintf("(%s...)", short)
		}
	}

	// Detect in-progress operation via git rev-parse --git-dir (worktree-safe)
	if gitDirRaw, ok := execGit(cwd, "rev-parse", "--git-dir"); ok {
		gitDir := gitDirRaw
		if !filepath.IsAbs(gitDir) {
			gitDir = filepath.Join(cwd, gitDir)
		}
		info.operation = detectOperation(gitDir)
	}

	if info.branch == "" && !info.detached {
		return nil
	}

	return info
}

func detectOperation(gitDir string) string {
	// Rebase (merge-based)
	if isDir(filepath.Join(gitDir, "rebase-merge")) {
		step := readFileTrimmed(filepath.Join(gitDir, "rebase-merge", "msgnum"))
		total := readFileTrimmed(filepath.Join(gitDir, "rebase-merge", "end"))
		return formatOperation("REBASE", step, total)
	}

	// Rebase (apply-based)
	if isDir(filepath.Join(gitDir, "rebase-apply")) {
		step := readFileTrimmed(filepath.Join(gitDir, "rebase-apply", "next"))
		total := readFileTrimmed(filepath.Join(gitDir, "rebase-apply", "last"))
		if fileExists(filepath.Join(gitDir, "rebase-apply", "rebasing")) {
			return formatOperation("REBASE", step, total)
		}
		if fileExists(filepath.Join(gitDir, "rebase-apply", "applying")) {
			return formatOperation("AM", step, total)
		}
		return formatOperation("AM/REBASE", step, total)
	}

	if fileExists(filepath.Join(gitDir, "MERGE_HEAD")) {
		return "MERGING"
	}
	if fileExists(filepath.Join(gitDir, "CHERRY_PICK_HEAD")) {
		return "CHERRY-PICKING"
	}
	if fileExists(filepath.Join(gitDir, "REVERT_HEAD")) {
		return "REVERTING"
	}

	// Sequencer
	if todo := readFileTrimmed(filepath.Join(gitDir, "sequencer", "todo")); todo != "" {
		firstLine := strings.SplitN(todo, "\n", 2)[0]
		if strings.HasPrefix(firstLine, "p ") || strings.HasPrefix(firstLine, "pick ") {
			return "CHERRY-PICKING"
		}
		if strings.HasPrefix(firstLine, "revert ") {
			return "REVERTING"
		}
	}

	if fileExists(filepath.Join(gitDir, "BISECT_LOG")) {
		return "BISECTING"
	}

	return ""
}

func formatOperation(op, step, total string) string {
	if step != "" && total != "" {
		return fmt.Sprintf("%s %s/%s", op, step, total)
	}
	return op
}

func readFileTrimmed(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
