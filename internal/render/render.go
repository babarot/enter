package render

import (
	"fmt"
	"strings"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

func Render(outputs []*module.Output, cfg *config.Config) string {
	theme := GetTheme(cfg.Theme)
	sep := Dim(cfg.Separator)

	var parts []string
	for _, out := range outputs {
		var buf strings.Builder
		for _, seg := range out.Segments {
			buf.WriteString(Paint(seg.Text, seg.Color, theme))
		}
		if buf.Len() > 0 {
			parts = append(parts, buf.String())
		}
	}

	return strings.Join(parts, sep)
}

func Paint(text string, color module.SemanticColor, theme *ThemePalette) string {
	rgb := ColorForSemantic(color, theme)
	if rgb == nil {
		return text
	}
	return fmt.Sprintf("\033[38;2;%d;%d;%dm%s\033[0m", rgb.R, rgb.G, rgb.B, text)
}

func Dim(text string) string {
	return fmt.Sprintf("\033[2m%s\033[0m", text)
}
