package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

func Render(outputs []*module.Output, cfg *config.Config) string {
	theme := GetTheme(cfg.Theme)

	switch cfg.Format {
	case "table":
		return renderTable(outputs, cfg, theme)
	case "compact":
		return renderCompact(outputs, cfg, theme)
	default:
		return renderInline(outputs, cfg, theme)
	}
}

func renderInline(outputs []*module.Output, cfg *config.Config, theme *ThemePalette) string {
	sep := Dim(cfg.Separator)

	var parts []string
	for _, out := range outputs {
		rendered := renderSegments(out.Segments, theme)
		if rendered != "" {
			parts = append(parts, rendered)
		}
	}

	return strings.Join(parts, sep)
}

func renderTable(outputs []*module.Output, _ *config.Config, theme *ThemePalette) string {
	var rows [][]string
	for _, out := range outputs {
		label := Paint(out.Name, module.Muted, theme)
		value := renderSegments(out.Segments, theme)
		if value != "" {
			rows = append(rows, []string{label, value})
		}
	}

	if len(rows) == 0 {
		return ""
	}

	borderColor := lipgloss.Color(toHex(theme.Muted))

	t := table.New().
		Rows(rows...).
		BorderStyle(lipgloss.NewStyle().Foreground(borderColor)).
		StyleFunc(func(row, col int) lipgloss.Style {
			if col == 0 {
				return lipgloss.NewStyle().PaddingRight(1)
			}
			return lipgloss.NewStyle().PaddingLeft(1)
		})

	return t.Render()
}

func renderCompact(outputs []*module.Output, _ *config.Config, theme *ThemePalette) string {
	var lines []string
	for _, out := range outputs {
		value := renderSegments(out.Segments, theme)
		if value == "" {
			continue
		}
		label := Paint(out.Name, module.Muted, theme)
		lines = append(lines, label+"  "+value)
	}

	return strings.Join(lines, "\n")
}

func renderSegments(segments []module.Segment, theme *ThemePalette) string {
	var buf strings.Builder
	for _, seg := range segments {
		buf.WriteString(Paint(seg.Text, seg.Color, theme))
	}
	return buf.String()
}

func Paint(text string, color module.SemanticColor, theme *ThemePalette) string {
	rgb := ColorForSemantic(color, theme)
	if rgb == nil {
		return text
	}
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(toHex(*rgb)))
	return style.Render(text)
}

func Dim(text string) string {
	style := lipgloss.NewStyle().Faint(true)
	return style.Render(text)
}

func toHex(c RGB) string {
	return "#" + hexByte(c.R) + hexByte(c.G) + hexByte(c.B)
}

func hexByte(b uint8) string {
	const hex = "0123456789abcdef"
	return string([]byte{hex[b>>4], hex[b&0x0f]})
}
