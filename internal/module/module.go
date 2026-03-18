package module

import "github.com/babarot/enter/internal/config"

type SemanticColor int

const (
	Default SemanticColor = iota
	Primary
	Secondary
	Success
	Warning
	Danger
	Muted
	Accent
)

type Segment struct {
	Text      string
	Color     SemanticColor
	Underline bool
}

func NewSegment(text string, color SemanticColor) Segment {
	return Segment{Text: text, Color: color}
}

func Plain(text string) Segment {
	return Segment{Text: text, Color: Default}
}

// Row is a key-value pair for structured output (table format).
type Row struct {
	Key      string
	Segments []Segment
}

type Output struct {
	Name     string
	Segments []Segment // used for inline format
	Rows     []Row     // used for table format (optional)
}

type Context struct {
	Cwd    string
	Config *config.Config
}

// Module is the interface that all pluggable modules implement.
// Run returns nil if the module is disabled or has nothing to display.
type Module interface {
	Name() string
	Run(ctx *Context) *Output
}
