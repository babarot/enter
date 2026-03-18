package render

import (
	"os"
	"strings"
	"testing"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

func TestMain(m *testing.M) {
	// Force lipgloss to output ANSI codes even in non-TTY (test) environment
	os.Setenv("CLICOLOR_FORCE", "1")
	os.Exit(m.Run())
}

func TestGetTheme(t *testing.T) {
	themes := []string{"default", "tokyo-night", "catppuccin", "dracula", "nord", "gruvbox"}
	for _, name := range themes {
		theme := GetTheme(name)
		if theme == nil {
			t.Errorf("GetTheme(%q) returned nil", name)
		}
	}

	// Unknown theme should return default
	unknown := GetTheme("nonexistent")
	def := GetTheme("default")
	if unknown.Primary != def.Primary {
		t.Error("unknown theme should return default palette")
	}
}

func TestColorForSemantic(t *testing.T) {
	theme := GetTheme("default")

	tests := []struct {
		color module.SemanticColor
		isNil bool
	}{
		{module.Primary, false},
		{module.Secondary, false},
		{module.Success, false},
		{module.Warning, false},
		{module.Danger, false},
		{module.Muted, false},
		{module.Accent, false},
		{module.Default, true},
	}

	for _, tt := range tests {
		rgb := ColorForSemantic(tt.color, theme)
		if tt.isNil && rgb != nil {
			t.Errorf("ColorForSemantic(%v) should be nil", tt.color)
		}
		if !tt.isNil && rgb == nil {
			t.Errorf("ColorForSemantic(%v) should not be nil", tt.color)
		}
	}
}

func TestPaint(t *testing.T) {
	theme := GetTheme("default")

	// Default color should return text as-is
	result := Paint("hello", module.Default, theme)
	if result != "hello" {
		t.Errorf("Paint with Default color should return plain text, got %q", result)
	}

	// Non-default color should add ANSI codes
	result = Paint("hello", module.Primary, theme)
	if !strings.Contains(result, "hello") {
		t.Error("Paint result should contain the original text")
	}
	if result == "hello" {
		t.Error("Paint with Primary should add color codes")
	}
}

func TestDim(t *testing.T) {
	result := Dim("hello")
	if !strings.Contains(result, "hello") {
		t.Error("Dim result should contain the original text")
	}
	if result == "hello" {
		t.Error("Dim should add formatting")
	}
}

func TestRenderInline(t *testing.T) {
	cfg := config.Default()
	outputs := []*module.Output{
		{
			Name:     "cwd",
			Segments: []module.Segment{module.NewSegment("test/dir", module.Secondary)},
		},
		{
			Name:     "git",
			Segments: []module.Segment{module.NewSegment("(main)", module.Success)},
		},
	}

	result := Render(outputs, cfg)
	if !strings.Contains(result, "test/dir") {
		t.Error("inline render should contain cwd")
	}
	if !strings.Contains(result, "main") {
		t.Error("inline render should contain git branch")
	}
}

func TestRenderTable(t *testing.T) {
	cfg := config.Default()
	cfg.Format = "table"
	outputs := []*module.Output{
		{
			Name:     "cwd",
			Segments: []module.Segment{module.NewSegment("test/dir", module.Secondary)},
		},
	}

	result := Render(outputs, cfg)
	if !strings.Contains(result, "test/dir") {
		t.Error("table render should contain cwd value")
	}
	if !strings.Contains(result, "╭") {
		t.Error("table render should contain border characters")
	}
}

func TestRenderWithRows(t *testing.T) {
	cfg := config.Default()
	cfg.Format = "table"
	outputs := []*module.Output{
		{
			Name:     "git",
			Segments: []module.Segment{module.NewSegment("(main)", module.Success)},
			Rows: []module.Row{
				{Key: "git.url", Segments: []module.Segment{module.NewSegment("https://github.com/test/repo", module.Primary)}},
				{Key: "git.sign", Segments: []module.Segment{module.NewSegment("(main)", module.Success)}},
			},
		},
	}

	result := Render(outputs, cfg)
	// Default key_style is "tree", so keys become "├── url", "└── sign"
	if !strings.Contains(result, "url") {
		t.Error("table render with rows should contain url key")
	}
	if !strings.Contains(result, "sign") {
		t.Error("table render with rows should contain sign key")
	}
	if !strings.Contains(result, "test/repo") {
		t.Error("table render with rows should contain repo URL")
	}
	// Should have group header
	if !strings.Contains(result, "git") {
		t.Error("table render with tree keys should contain git group header")
	}
}

func TestRenderWithRowsFlat(t *testing.T) {
	cfg := config.Default()
	cfg.Format = "table"
	cfg.KeyStyle = "flat"
	outputs := []*module.Output{
		{
			Name:     "git",
			Segments: []module.Segment{module.NewSegment("(main)", module.Success)},
			Rows: []module.Row{
				{Key: "git.url", Segments: []module.Segment{module.NewSegment("https://github.com/test/repo", module.Primary)}},
				{Key: "git.sign", Segments: []module.Segment{module.NewSegment("(main)", module.Success)}},
			},
		},
	}

	result := Render(outputs, cfg)
	if !strings.Contains(result, "git.url") {
		t.Error("flat render should contain git.url")
	}
	if !strings.Contains(result, "git.sign") {
		t.Error("flat render should contain git.sign")
	}
}

func TestReorderRows(t *testing.T) {
	rows := []module.Row{
		{Key: "git.url", Segments: []module.Segment{module.NewSegment("url", module.Primary)}},
		{Key: "git.sign", Segments: []module.Segment{module.NewSegment("sign", module.Success)}},
		{Key: "git.cwd", Segments: []module.Segment{module.NewSegment("cwd", module.Muted)}},
		{Key: "git.status", Segments: []module.Segment{module.NewSegment("status", module.Danger)}},
	}

	// Reverse order
	reordered := reorderRows(rows, "git", []string{"status", "cwd", "sign", "url"})
	want := []string{"git.status", "git.cwd", "git.sign", "git.url"}
	if len(reordered) != len(want) {
		t.Fatalf("reorderRows length: got %d, want %d", len(reordered), len(want))
	}
	for i, key := range want {
		if reordered[i].Key != key {
			t.Errorf("reorderRows[%d]: got %q, want %q", i, reordered[i].Key, key)
		}
	}
}

func TestReorderRowsPartial(t *testing.T) {
	rows := []module.Row{
		{Key: "git.url", Segments: []module.Segment{module.NewSegment("url", module.Primary)}},
		{Key: "git.sign", Segments: []module.Segment{module.NewSegment("sign", module.Success)}},
		{Key: "git.cwd", Segments: []module.Segment{module.NewSegment("cwd", module.Muted)}},
	}

	// Only specify sign first, rest appended
	reordered := reorderRows(rows, "git", []string{"sign"})
	if reordered[0].Key != "git.sign" {
		t.Errorf("first should be git.sign, got %q", reordered[0].Key)
	}
	if len(reordered) != 3 {
		t.Errorf("all rows should be present, got %d", len(reordered))
	}
}

func TestReorderRowsEmpty(t *testing.T) {
	rows := []module.Row{
		{Key: "git.url", Segments: []module.Segment{module.NewSegment("url", module.Primary)}},
		{Key: "git.sign", Segments: []module.Segment{module.NewSegment("sign", module.Success)}},
	}

	// Empty order should keep original order
	reordered := reorderRows(rows, "git", []string{})
	if reordered[0].Key != "git.url" {
		t.Errorf("empty order should keep original, got %q first", reordered[0].Key)
	}
}

func TestRenderWithRowOrder(t *testing.T) {
	cfg := config.Default()
	cfg.Format = "table"
	cfg.KeyStyle = "flat"
	outputs := []*module.Output{
		{
			Name:     "git",
			Segments: []module.Segment{module.NewSegment("(main)", module.Success)},
			Rows: []module.Row{
				{Key: "git.url", Segments: []module.Segment{module.NewSegment("https://example.com", module.Primary)}},
				{Key: "git.sign", Segments: []module.Segment{module.NewSegment("(main)", module.Success)}},
			},
			RowOrder: []string{"sign", "url"},
		},
	}

	result := Render(outputs, cfg)
	// sign should appear before url
	signIdx := strings.Index(result, "sign")
	urlIdx := strings.Index(result, "url")
	if signIdx < 0 || urlIdx < 0 {
		t.Fatalf("both sign and url should be present, got %q", result)
	}
	if signIdx > urlIdx {
		t.Error("sign should appear before url with RowOrder [sign, url]")
	}
}

func TestRenderEmpty(t *testing.T) {
	cfg := config.Default()
	result := Render(nil, cfg)
	if result != "" {
		t.Errorf("empty outputs should render empty string, got %q", result)
	}
}
