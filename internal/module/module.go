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
	Text  string
	Color SemanticColor
}

func NewSegment(text string, color SemanticColor) Segment {
	return Segment{Text: text, Color: color}
}

func Plain(text string) Segment {
	return Segment{Text: text, Color: Default}
}

type Output struct {
	Name     string
	Segments []Segment
}

type Context struct {
	Cwd    string
	Config *config.Config
}

type Module interface {
	Name() string
	Run(ctx *Context) *Output
}
