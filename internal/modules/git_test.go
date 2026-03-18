package modules

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name             string
		raw              string
		wantSlug, wantURL string
	}{
		{
			"scp style",
			"git@github.com:babarot/enter.git",
			"babarot/enter",
			"https://github.com/babarot/enter",
		},
		{
			"ssh scheme with port",
			"ssh://git@ssh.github.com:443/babarot/enter.git",
			"babarot/enter",
			"https://github.com/babarot/enter",
		},
		{
			"https",
			"https://github.com/babarot/enter.git",
			"babarot/enter",
			"https://github.com/babarot/enter",
		},
		{
			"https no .git",
			"https://github.com/babarot/enter",
			"babarot/enter",
			"https://github.com/babarot/enter",
		},
		{
			"http",
			"http://github.com/babarot/enter.git",
			"babarot/enter",
			"http://github.com/babarot/enter",
		},
		{
			"gitlab scp",
			"git@gitlab.com:org/repo.git",
			"org/repo",
			"https://gitlab.com/org/repo",
		},
		{
			"empty",
			"",
			"", "",
		},
		{
			"invalid",
			"not-a-url",
			"", "",
		},
	}

	for _, tt := range tests {
		slug, url := parseRemoteURL(tt.raw)
		if slug != tt.wantSlug {
			t.Errorf("%s: slug = %q, want %q", tt.name, slug, tt.wantSlug)
		}
		if url != tt.wantURL {
			t.Errorf("%s: url = %q, want %q", tt.name, url, tt.wantURL)
		}
	}
}

func TestStatusCodeColor(t *testing.T) {
	tests := []struct {
		x, y byte
		want module.SemanticColor
	}{
		{'A', ' ', module.Success},
		{' ', 'A', module.Success},
		{'D', ' ', module.Danger},
		{' ', 'D', module.Danger},
		{'M', ' ', module.Warning},
		{' ', 'M', module.Warning},
		{'R', ' ', module.Accent},
		{'?', '?', module.Muted},
		{'M', 'M', module.Warning},
	}

	for _, tt := range tests {
		got := statusCodeColor(tt.x, tt.y)
		if got != tt.want {
			t.Errorf("statusCodeColor(%c, %c) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestFormatTree(t *testing.T) {
	tests := []struct {
		name, root, rel, style, want string
	}{
		{"root breadcrumb", "/home/user/project", "", "breadcrumb", "/"},
		{"root tree", "/home/user/project", "", "tree", "/"},
		{"subdir breadcrumb", "/home/user/project", "cmd/enter", "breadcrumb", "/project → cmd → enter"},
		{"single dir breadcrumb", "/home/user/project", "src", "breadcrumb", "/project → src"},
	}

	for _, tt := range tests {
		got := formatTree(tt.root, tt.rel, tt.style)
		if got != tt.want {
			t.Errorf("%s: formatTree(%q, %q, %q) = %q, want %q",
				tt.name, tt.root, tt.rel, tt.style, got, tt.want)
		}
	}
}

func TestFormatTreeTreeStyle(t *testing.T) {
	got := formatTree("/home/user/project", "cmd/enter", "tree")
	if !strings.Contains(got, "/project") {
		t.Errorf("tree style should contain /project root, got %q", got)
	}
	if !strings.Contains(got, "cmd") {
		t.Errorf("tree style should contain cmd, got %q", got)
	}
	if !strings.Contains(got, "← here") {
		t.Errorf("tree style should contain ← here marker, got %q", got)
	}
}

// --- Integration tests using real git repos ---

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git init failed: %s: %s", err, out)
		}
	}

	// Create a file and commit
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "README.md"},
		{"git", "commit", "-m", "initial"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %s: %s", args, err, out)
		}
	}

	return dir
}

func TestGetGitInfoBasic(t *testing.T) {
	dir := initTestRepo(t)

	info := getGitInfo(dir)
	if info == nil {
		t.Fatal("getGitInfo returned nil for valid repo")
	}
	if info.branch == "" {
		t.Error("branch should not be empty")
	}
	if info.detached {
		t.Error("should not be detached")
	}
}

func TestGetGitInfoNotARepo(t *testing.T) {
	dir := t.TempDir()
	info := getGitInfo(dir)
	if info != nil {
		t.Error("getGitInfo should return nil for non-repo")
	}
}

func TestGetGitInfoWithChanges(t *testing.T) {
	dir := initTestRepo(t)

	// Create untracked file
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0o644)

	info := getGitInfo(dir)
	if info == nil {
		t.Fatal("getGitInfo returned nil")
	}
	if !info.untracked {
		t.Error("should have untracked files")
	}

	// Modify tracked file
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# modified"), 0o644)

	info = getGitInfo(dir)
	if info == nil {
		t.Fatal("getGitInfo returned nil")
	}
	if !info.unstaged {
		t.Error("should have unstaged changes")
	}
}

func TestGetGitInfoStagedChanges(t *testing.T) {
	dir := initTestRepo(t)

	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# staged"), 0o644)
	cmd := exec.Command("git", "add", "README.md")
	cmd.Dir = dir
	cmd.Run()

	info := getGitInfo(dir)
	if info == nil {
		t.Fatal("getGitInfo returned nil")
	}
	if !info.staged {
		t.Error("should have staged changes")
	}
}

func TestGitModuleDisabled(t *testing.T) {
	m := &GitModule{}
	cfg := config.Default()
	cfg.Modules.Git.Enabled = false
	ctx := &module.Context{Cwd: "/tmp", Config: cfg}

	out := m.Run(ctx)
	if out != nil {
		t.Error("disabled git module should return nil")
	}
}

func TestGitModuleNotARepo(t *testing.T) {
	m := &GitModule{}
	cfg := config.Default()
	cfg.Modules.Git.ShowIndicator = false
	ctx := &module.Context{Cwd: t.TempDir(), Config: cfg}

	out := m.Run(ctx)
	if out != nil {
		t.Error("git module outside repo with show_indicator=false should return nil")
	}
}

func TestGitModuleShowIndicator(t *testing.T) {
	m := &GitModule{}
	cfg := config.Default()
	cfg.Modules.Git.ShowIndicator = true
	ctx := &module.Context{Cwd: t.TempDir(), Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("show_indicator should return output outside repo")
	}
	found := false
	for _, seg := range out.Segments {
		if strings.Contains(seg.Text, "not a git repo") {
			found = true
			break
		}
	}
	if !found {
		t.Error("should contain 'not a git repo'")
	}
}

func TestGitModuleInRepo(t *testing.T) {
	dir := initTestRepo(t)
	m := &GitModule{}
	cfg := config.Default()
	ctx := &module.Context{Cwd: dir, Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("git module in repo should return output")
	}
	if out.Name != "git" {
		t.Errorf("name: got %q, want %q", out.Name, "git")
	}

	// Should contain branch info in segments
	text := segmentsText(out.Segments)
	if !strings.Contains(text, "(") {
		t.Errorf("should contain branch parens, got %q", text)
	}
}

func TestGitModuleShowRepo(t *testing.T) {
	dir := initTestRepo(t)

	// Add a remote
	cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
	cmd.Dir = dir
	cmd.Run()

	m := &GitModule{}
	cfg := config.Default()
	cfg.Modules.Git.ShowRepo = true
	ctx := &module.Context{Cwd: dir, Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("git module should return output")
	}

	text := segmentsText(out.Segments)
	if !strings.Contains(text, "https://github.com/test/repo") {
		t.Errorf("show_repo should include URL, got %q", text)
	}

	// Check rows
	found := false
	for _, row := range out.Rows {
		if row.Key == "git.url" {
			found = true
			break
		}
	}
	if !found {
		t.Error("should have git.url row")
	}
}

func TestGitModuleShowTree(t *testing.T) {
	dir := initTestRepo(t)
	subdir := filepath.Join(dir, "sub", "dir")
	os.MkdirAll(subdir, 0o755)

	m := &GitModule{}
	cfg := config.Default()
	cfg.Modules.Git.ShowTree = true
	ctx := &module.Context{Cwd: subdir, Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("git module should return output")
	}

	found := false
	for _, row := range out.Rows {
		if row.Key == "git.cwd" {
			found = true
			break
		}
	}
	if !found {
		t.Error("should have git.cwd row in subdir")
	}
}

func TestGitModuleShowStatus(t *testing.T) {
	dir := initTestRepo(t)
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0o644)

	m := &GitModule{}
	cfg := config.Default()
	cfg.Modules.Git.ShowStatus = true
	ctx := &module.Context{Cwd: dir, Config: cfg}

	out := m.Run(ctx)
	if out == nil {
		t.Fatal("git module should return output")
	}

	found := false
	for _, row := range out.Rows {
		if row.Key == "git.status" {
			found = true
			text := segmentsText(row.Segments)
			if !strings.Contains(text, "new.txt") {
				t.Errorf("git.status should show new.txt, got %q", text)
			}
			break
		}
	}
	if !found {
		t.Error("should have git.status row")
	}
}

func TestDetectOperation(t *testing.T) {
	dir := t.TempDir()

	// No operation
	if op := detectOperation(dir); op != "" {
		t.Errorf("no operation expected, got %q", op)
	}

	// Simulate merge
	os.WriteFile(filepath.Join(dir, "MERGE_HEAD"), []byte("abc123"), 0o644)
	if op := detectOperation(dir); op != "MERGING" {
		t.Errorf("got %q, want %q", op, "MERGING")
	}
	os.Remove(filepath.Join(dir, "MERGE_HEAD"))

	// Simulate cherry-pick
	os.WriteFile(filepath.Join(dir, "CHERRY_PICK_HEAD"), []byte("abc123"), 0o644)
	if op := detectOperation(dir); op != "CHERRY-PICKING" {
		t.Errorf("got %q, want %q", op, "CHERRY-PICKING")
	}
	os.Remove(filepath.Join(dir, "CHERRY_PICK_HEAD"))

	// Simulate revert
	os.WriteFile(filepath.Join(dir, "REVERT_HEAD"), []byte("abc123"), 0o644)
	if op := detectOperation(dir); op != "REVERTING" {
		t.Errorf("got %q, want %q", op, "REVERTING")
	}
	os.Remove(filepath.Join(dir, "REVERT_HEAD"))

	// Simulate bisect
	os.WriteFile(filepath.Join(dir, "BISECT_LOG"), []byte("log"), 0o644)
	if op := detectOperation(dir); op != "BISECTING" {
		t.Errorf("got %q, want %q", op, "BISECTING")
	}
	os.Remove(filepath.Join(dir, "BISECT_LOG"))

	// Simulate rebase
	os.MkdirAll(filepath.Join(dir, "rebase-merge"), 0o755)
	os.WriteFile(filepath.Join(dir, "rebase-merge", "msgnum"), []byte("2"), 0o644)
	os.WriteFile(filepath.Join(dir, "rebase-merge", "end"), []byte("5"), 0o644)
	if op := detectOperation(dir); op != "REBASE 2/5" {
		t.Errorf("got %q, want %q", op, "REBASE 2/5")
	}
}

func segmentsText(segs []module.Segment) string {
	var b strings.Builder
	for _, s := range segs {
		b.WriteString(s.Text)
	}
	return b.String()
}
