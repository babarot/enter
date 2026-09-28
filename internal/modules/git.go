package modules

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss/tree"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

type GitModule struct{}

func (m *GitModule) Name() string { return config.ModuleGit }

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
	repoSlug  string // "owner/repo"
	repoURL   string // "https://github.com/owner/repo"
	repoRoot  string // absolute path to repo root
	relPath   string // cwd relative to repo root (empty = at root)
	upstream  string // "origin/main" (empty = no upstream)
	gitDir    string // absolute path to this worktree's git dir
}

// worktreeInfo describes which worktree cwd belongs to.
type worktreeInfo struct {
	linked   bool   // true = a worktree added by "git worktree add"
	mainPath string // path of the main worktree
}

// baseInfo describes how far HEAD is from the base branch.
type baseInfo struct {
	ref       string    // "origin/main"
	ahead     int       // commits in HEAD but not in ref
	behind    int       // commits in ref but not in HEAD
	fetchedAt time.Time // zero = never fetched
}

func (m *GitModule) Run(ctx *module.Context) *module.Output {
	if !ctx.Config.Modules.Git.Enabled {
		return nil
	}

	gitCfg := &ctx.Config.Modules.Git

	// indicator: show whether we're in a git repo
	info := getGitInfo(ctx.Cwd)
	if info == nil {
		if gitCfg.Indicator {
			return &module.Output{
				Name:     m.Name(),
				Segments: []module.Segment{module.NewSegment("not a git repo", module.Muted)},
			}
		}
		return nil
	}
	if info.branch == "" {
		return nil
	}

	// Build summary segments: (branch *+$% ↑1↓2|REBASE)
	var statusSegs []module.Segment
	if gitCfg.Fields.Summary.Present() {
		symbols := gitCfg.Fields.Summary.Get().Symbols

		branchColor := module.Success
		if info.detached {
			branchColor = module.Danger
		}

		statusSegs = append(statusSegs, module.NewSegment("(", branchColor))
		branchSeg := module.NewSegment(info.branch, branchColor)
		if info.repoURL != "" && !info.detached {
			branchSeg.Link = info.repoURL + "/tree/" + info.branch
		}
		statusSegs = append(statusSegs, branchSeg)

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
			statusSegs = append(statusSegs, module.Plain(" "))
			statusSegs = append(statusSegs, flags...)
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
			statusSegs = append(statusSegs, module.Plain(" "))
			for i, seg := range upstream {
				if i > 0 {
					statusSegs = append(statusSegs, module.Plain(" "))
				}
				statusSegs = append(statusSegs, seg)
			}
		}

		// Operation
		if info.operation != "" {
			statusSegs = append(statusSegs, module.NewSegment("|"+info.operation, module.Accent))
		}

		statusSegs = append(statusSegs, module.NewSegment(")", branchColor))
	}

	// Build cwd segments (show current position in repo)
	var cwdSegs []module.Segment
	if gitCfg.Fields.Cwd.Present() {
		cwdText := formatTree(info.repoRoot, info.relPath, gitCfg.Fields.Cwd.Get().Style)
		cwdSegs = append(cwdSegs, module.NewSegment(cwdText, module.Muted))
	}

	var worktreeSegs []module.Segment
	if gitCfg.Fields.Worktree.Present() {
		worktreeSegs = worktreeSegments(getWorktreeInfo(ctx.Cwd, info.gitDir), gitCfg.Fields.Worktree.Get())
	}

	var baseSegs []module.Segment
	if gitCfg.Fields.Base.Present() {
		baseCfg := gitCfg.Fields.Base.Get()
		symbols := config.DefaultGitSymbols()
		if gitCfg.Fields.Summary.Present() {
			symbols = gitCfg.Fields.Summary.Get().Symbols
		}
		if base := getBaseInfo(ctx.Cwd, baseCfg.Ref); base != nil {
			if baseCfg.Mode == config.GitModeAlways || base.ref != info.upstream {
				baseSegs = baseSegments(base, symbols, baseCfg.ShowFetchAge(), time.Now())
			}
		}
	}

	// Build inline segments (all in one line)
	var segments []module.Segment
	if gitCfg.Fields.Url.Present() && info.repoURL != "" {
		segments = append(segments, module.NewSegment(info.repoURL, module.Primary))
		segments = append(segments, module.Plain(" "))
	}
	segments = append(segments, statusSegs...)
	for _, extra := range [][]module.Segment{worktreeSegs, baseSegs} {
		if len(extra) > 0 {
			segments = append(segments, module.Plain(" "))
			segments = append(segments, extra...)
		}
	}
	if len(cwdSegs) > 0 {
		segments = append(segments, module.Plain(" "))
		segments = append(segments, cwdSegs...)
	}

	// Build rows for table format
	var rows []module.Row
	if gitCfg.Fields.Url.Present() && info.repoURL != "" {
		rows = append(rows, module.Row{
			Key:      "git.url",
			Segments: []module.Segment{module.NewSegment(info.repoURL, module.Primary)},
		})
	}
	if gitCfg.Fields.Cwd.Present() {
		rows = append(rows, module.Row{
			Key:      "git.cwd",
			Segments: cwdSegs,
		})
	}
	if gitCfg.Fields.Summary.Present() {
		rows = append(rows, module.Row{
			Key:      "git.summary",
			Segments: statusSegs,
		})
	}
	if len(worktreeSegs) > 0 {
		rows = append(rows, module.Row{
			Key:      "git.worktree",
			Segments: worktreeSegs,
		})
	}
	if len(baseSegs) > 0 {
		rows = append(rows, module.Row{
			Key:      "git.base",
			Segments: baseSegs,
		})
	}
	if gitCfg.Fields.Status.Present() {
		statusSegs := getGitStatusSegments(ctx.Cwd, gitCfg.Fields.Status.Get().Style)
		if len(statusSegs) > 0 {
			rows = append(rows, module.Row{
				Key:      "git.status",
				Segments: statusSegs,
			})
		}
	}

	return &module.Output{
		Name:     m.Name(),
		Segments: segments,
		Rows:     rows,
	}
}

func worktreeSegments(wt *worktreeInfo, cfg config.GitWorktreeConfig) []module.Segment {
	if wt == nil {
		return nil
	}
	if !wt.linked {
		if cfg.Mode != config.GitModeAlways {
			return nil
		}
		return []module.Segment{module.NewSegment("main worktree", module.Muted)}
	}
	home, _ := os.UserHomeDir()
	return []module.Segment{
		module.NewSegment("linked", module.Warning),
		module.NewSegment(" ← "+formatPath(wt.mainPath, home, config.CwdStyleFull), module.Muted),
	}
}

func baseSegments(b *baseInfo, symbols config.GitSymbols, showFetchAge bool, now time.Time) []module.Segment {
	segs := []module.Segment{module.NewSegment(b.ref, module.Secondary)}
	switch {
	case b.ahead == 0 && b.behind == 0:
		segs = append(segs, module.NewSegment(" =", module.Success))
	default:
		if b.ahead > 0 {
			segs = append(segs, module.NewSegment(fmt.Sprintf(" %s%d", symbols.Ahead, b.ahead), module.Success))
		}
		if b.behind > 0 {
			segs = append(segs, module.NewSegment(fmt.Sprintf(" %s%d", symbols.Behind, b.behind), module.Danger))
		}
	}
	if showFetchAge {
		segs = append(segs, module.NewSegment(" ("+formatFetchAge(b.fetchedAt, now)+")", module.Muted))
	}
	return segs
}

// formatFetchAge describes how old the remote-tracking refs are.
func formatFetchAge(fetchedAt, now time.Time) string {
	if fetchedAt.IsZero() {
		return "never fetched"
	}
	d := now.Sub(fetchedAt)
	switch {
	case d < time.Minute:
		return "fetched just now"
	case d < time.Hour:
		return fmt.Sprintf("fetched %dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("fetched %dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("fetched %dd ago", int(d.Hours()/24))
	}
}

// getWorktreeInfo reports whether cwd is in a linked worktree.
// A linked worktree has its own git dir under <common-dir>/worktrees/,
// so the two differ; in the main worktree they are the same.
func getWorktreeInfo(cwd, gitDir string) *worktreeInfo {
	commonDir, ok := execGit(cwd, "rev-parse", "--git-common-dir")
	if !ok || gitDir == "" {
		return nil
	}
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(cwd, commonDir)
	}
	wt := &worktreeInfo{linked: cleanPath(gitDir) != cleanPath(commonDir)}
	if !wt.linked {
		return wt
	}
	// The first entry of "git worktree list" is always the main worktree.
	if list, ok := execGit(cwd, "worktree", "list", "--porcelain"); ok {
		first, _, _ := strings.Cut(list, "\n")
		wt.mainPath = strings.TrimPrefix(first, "worktree ")
	}
	return wt
}

// getBaseInfo counts commits between HEAD and the base ref.
// It does not fetch, so the counts are as fresh as the last fetch.
func getBaseInfo(cwd, ref string) *baseInfo {
	if ref == "" {
		ref = defaultBaseRef(cwd)
	}
	if ref == "" {
		return nil
	}
	counts, ok := execGit(cwd, "rev-list", "--left-right", "--count", "HEAD..."+ref)
	if !ok {
		return nil
	}
	b := &baseInfo{ref: ref}
	if _, err := fmt.Sscanf(counts, "%d %d", &b.ahead, &b.behind); err != nil {
		return nil
	}
	b.fetchedAt = lastFetchTime(cwd, ref)
	return b
}

// lastFetchTime returns the latest of two signals, since neither covers
// every fetch on its own:
//   - the reflog of ref, which records a fetch from any worktree but only
//     when the ref actually moved
//   - FETCH_HEAD, which is written on every fetch but lives in each
//     worktree's own git dir, so all of them are checked
func lastFetchTime(cwd, ref string) time.Time {
	var latest time.Time
	if sel, ok := execGit(cwd, "reflog", "show", "-1", "--date=unix", "--format=%gd", "refs/remotes/"+ref); ok {
		// "origin/main@{1759080445}"
		if _, stamp, found := strings.Cut(sel, "@{"); found {
			var sec int64
			if _, err := fmt.Sscanf(strings.TrimSuffix(stamp, "}"), "%d", &sec); err == nil {
				latest = time.Unix(sec, 0)
			}
		}
	}
	commonDir, ok := execGit(cwd, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if !ok {
		return latest
	}
	heads, _ := filepath.Glob(filepath.Join(commonDir, "worktrees", "*", "FETCH_HEAD"))
	for _, path := range append(heads, filepath.Join(commonDir, "FETCH_HEAD")) {
		if st, err := os.Stat(path); err == nil && st.ModTime().After(latest) {
			latest = st.ModTime()
		}
	}
	return latest
}

// defaultBaseRef picks the remote's default branch: origin/HEAD if set,
// otherwise the first of origin/main and origin/master that exists.
func defaultBaseRef(cwd string) string {
	if ref, ok := execGit(cwd, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); ok && ref != "" {
		return ref
	}
	for _, ref := range []string{"origin/main", "origin/master"} {
		if _, ok := execGit(cwd, "rev-parse", "--verify", "--quiet", "refs/remotes/"+ref); ok {
			return ref
		}
	}
	return ""
}

// cleanPath resolves symlinks so that /tmp and /private/tmp compare equal.
func cleanPath(p string) string {
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	return filepath.Clean(p)
}

func getGitStatusSegments(cwd, style string) []module.Segment {
	switch style {
	case config.GitStatusStyleLong:
		return getGitStatusLong(cwd)
	default:
		return getGitStatusShort(cwd)
	}
}

func getGitStatusShort(cwd string) []module.Segment {
	output, ok := execGitRaw(cwd, "status", "--short")
	if !ok || output == "" {
		return nil
	}

	var segments []module.Segment
	for i, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		if i > 0 {
			segments = append(segments, module.Plain("\n"))
		}
		if len(line) >= 3 {
			x, y := line[0], line[1]
			filename := strings.TrimLeft(line[2:], " ")

			var code string
			switch {
			case x != ' ' && y != ' ':
				code = string([]byte{x, y})
			case x != ' ':
				code = string(x)
			case y != ' ':
				code = string(y)
			default:
				code = "?"
			}

			codeColor := statusCodeColor(x, y)
			segments = append(segments, module.NewSegment(fmt.Sprintf("%-2s", code), codeColor))
			segments = append(segments, module.NewSegment(" "+filename, codeColor))
		} else {
			segments = append(segments, module.NewSegment(line, module.Muted))
		}
	}
	return segments
}

func getGitStatusLong(cwd string) []module.Segment {
	output, ok := execGit(cwd, "status")
	if !ok || output == "" {
		return nil
	}

	// Color by section context
	var segments []module.Segment
	section := "header" // header, staged, unstaged, untracked
	for i, line := range strings.Split(output, "\n") {
		if i > 0 {
			segments = append(segments, module.Plain("\n"))
		}

		trimmed := strings.TrimSpace(line)

		// Detect section changes (headers always muted)
		switch {
		case strings.HasPrefix(trimmed, "Changes to be committed"):
			section = "staged"
			segments = append(segments, module.NewSegment(line, module.Muted))
			continue
		case strings.HasPrefix(trimmed, "Changes not staged"):
			section = "unstaged"
			segments = append(segments, module.NewSegment(line, module.Muted))
			continue
		case strings.HasPrefix(trimmed, "Untracked files"):
			section = "untracked"
			segments = append(segments, module.NewSegment(line, module.Muted))
			continue
		}

		// Color based on current section
		switch section {
		case "staged":
			if strings.HasPrefix(trimmed, "modified:") || strings.HasPrefix(trimmed, "new file:") ||
				strings.HasPrefix(trimmed, "deleted:") || strings.HasPrefix(trimmed, "renamed:") {
				segments = append(segments, module.NewSegment(line, module.Success))
			} else {
				segments = append(segments, module.NewSegment(line, module.Muted))
			}
		case "unstaged":
			if strings.HasPrefix(trimmed, "modified:") || strings.HasPrefix(trimmed, "deleted:") {
				segments = append(segments, module.NewSegment(line, module.Danger))
			} else {
				segments = append(segments, module.NewSegment(line, module.Muted))
			}
		case "untracked":
			segments = append(segments, module.NewSegment(line, module.Muted))
		default:
			segments = append(segments, module.NewSegment(line, module.Muted))
		}
	}
	return segments
}

// statusCodeColor returns the semantic color for a git status XY code pair.
// X = index (staged), Y = worktree (unstaged).
// Staged → green (Success), unstaged → red (Danger), both → yellow (Warning).
func statusCodeColor(x, y byte) module.SemanticColor {
	staged := x != ' ' && x != '?'
	unstaged := y != ' ' && y != '?'

	switch {
	case x == '?' && y == '?':
		return module.Muted // untracked
	case staged && unstaged:
		return module.Warning // both staged and unstaged changes
	case staged:
		return module.Success // staged only (index)
	case unstaged:
		return module.Danger // unstaged only (worktree)
	default:
		return module.Secondary
	}
}

// execGitRaw runs git and trims only trailing whitespace, preserving leading
// spaces that are significant in commands like "git status --short".
func execGitRaw(cwd string, args ...string) (string, bool) {
	allArgs := append([]string{"--no-optional-locks", "-C", cwd}, args...)
	cmd := exec.Command("git", allArgs...)
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return strings.TrimRight(string(out), " \t\n\r"), true
}

func execGit(cwd string, args ...string) (string, bool) {
	allArgs := append([]string{"--no-optional-locks", "-C", cwd}, args...)
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

		case strings.HasPrefix(line, "# branch.upstream "):
			info.upstream = strings.TrimPrefix(line, "# branch.upstream ")

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
		info.gitDir = gitDir
		info.operation = detectOperation(gitDir)
	}

	if info.branch == "" && !info.detached {
		return nil
	}

	// Get remote URL for repo slug
	if remoteURL, ok := execGit(cwd, "remote", "get-url", "origin"); ok {
		info.repoSlug, info.repoURL = parseRemoteURL(remoteURL)
	}

	// Get repo root and relative path
	if toplevel, ok := execGit(cwd, "rev-parse", "--show-toplevel"); ok {
		info.repoRoot = toplevel
		if cwd != toplevel && strings.HasPrefix(cwd, toplevel+"/") {
			info.relPath = cwd[len(toplevel)+1:]
		}
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

// parseRemoteURL extracts "owner/repo" and HTTPS URL from a git remote URL.
// Supports:
//   - git@github.com:owner/repo.git
//   - ssh://git@ssh.github.com:443/owner/repo.git
//   - https://github.com/owner/repo.git
func parseRemoteURL(raw string) (slug, httpURL string) {
	raw = strings.TrimSpace(raw)

	// SSH with explicit scheme: ssh://git@ssh.github.com:443/owner/repo.git
	if strings.HasPrefix(raw, "ssh://") {
		// Remove scheme
		after := strings.TrimPrefix(raw, "ssh://")
		// Remove user@ prefix
		if idx := strings.Index(after, "@"); idx >= 0 {
			after = after[idx+1:]
		}
		// Split host(:port) and path
		// ssh.github.com:443/owner/repo.git -> find first /
		if slashIdx := strings.Index(after, "/"); slashIdx >= 0 {
			host := after[:slashIdx]
			path := after[slashIdx+1:]
			path = strings.TrimSuffix(path, ".git")
			slug = path
			// Normalize host: ssh.github.com -> github.com
			host = strings.TrimPrefix(host, "ssh.")
			// Remove port
			if colonIdx := strings.Index(host, ":"); colonIdx >= 0 {
				host = host[:colonIdx]
			}
			httpURL = fmt.Sprintf("https://%s/%s", host, path)
			return
		}
	}

	// SCP-style SSH: git@github.com:owner/repo.git
	if strings.HasPrefix(raw, "git@") {
		after := strings.TrimPrefix(raw, "git@")
		if host, path, ok := strings.Cut(after, ":"); ok {
			path = strings.TrimSuffix(path, ".git")
			slug = path
			httpURL = fmt.Sprintf("https://%s/%s", host, path)
			return
		}
	}

	// HTTPS/HTTP
	if strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "http://") {
		trimmed := strings.TrimSuffix(raw, ".git")
		// Extract last two path segments as owner/repo
		parts := strings.Split(trimmed, "/")
		if len(parts) >= 2 {
			slug = parts[len(parts)-2] + "/" + parts[len(parts)-1]
			httpURL = trimmed
			return
		}
	}

	return "", ""
}

// formatTree renders the current position within the repo.
// repoRoot: absolute path to repo root
// relPath: relative path from root to cwd (e.g. "cmd/enter")
// style: "breadcrumb" or "tree"
func formatTree(repoRoot, relPath, style string) string {
	rootName := filepath.Base(repoRoot)

	// At repo root
	if relPath == "" {
		return "/" + rootName
	}

	parts := strings.Split(relPath, "/")

	switch style {
	case config.GitCwdStyleTree:
		return formatLipglossTree(rootName, parts)
	default: // config.GitCwdStyleBreadcrumb
		all := append([]string{"/" + rootName}, parts...)
		return strings.Join(all, " → ")
	}
}

func formatLipglossTree(rootName string, parts []string) string {
	// enter
	// └── cmd
	//     └── enter  ← here

	// Build from innermost to outermost
	var inner any
	for i := len(parts) - 1; i >= 0; i-- {
		label := parts[i]
		if i == len(parts)-1 {
			label = label + "  ← here"
		}
		if inner == nil {
			inner = tree.Root(label)
		} else {
			inner = tree.Root(label).Child(inner)
		}
	}

	t := tree.Root("/" + rootName).Child(inner)
	return t.String()
}
