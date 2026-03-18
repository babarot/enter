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

// flattenRows expands outputs into key-value pairs.
// If a module has Rows, each row becomes a separate entry.
// Otherwise, the module's Name + Segments become a single entry.
func flattenRows(outputs []*module.Output, theme *ThemePalette) []struct{ key, value string } {
	borderColor := lipgloss.Color(toHex(theme.Muted))
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	var result []struct{ key, value string }
	for _, out := range outputs {
		if len(out.Rows) > 0 {
			for _, row := range out.Rows {
				value := renderSegments(row.Segments, theme)
				if value != "" {
					// Wrap multiline values in a nested box
					if strings.Contains(value, "\n") {
						value = boxStyle.Render(value)
					}
					result = append(result, struct{ key, value string }{row.Key, value})
				}
			}
		} else {
			value := renderSegments(out.Segments, theme)
			if value != "" {
				if strings.Contains(value, "\n") {
					value = boxStyle.Render(value)
				}
				result = append(result, struct{ key, value string }{out.Name, value})
			}
		}
	}
	return result
}

func renderTable(outputs []*module.Output, _ *config.Config, theme *ThemePalette) string {
	entries := flattenRows(outputs, theme)
	if len(entries) == 0 {
		return ""
	}

	var rows [][]string
	for _, e := range entries {
		label := Paint(e.key, module.Muted, theme)
		rows = append(rows, []string{label, e.value})
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
	entries := flattenRows(outputs, theme)
	var lines []string
	for _, e := range entries {
		label := Paint(e.key, module.Muted, theme)
		lines = append(lines, label+"  "+e.value)
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
