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
	case "inline":
		return renderInline(outputs, cfg, theme)
	default:
		return renderTable(outputs, cfg, theme)
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
			// Build a set of parent keys to skip children
			// e.g. if "claude.usage" exists, skip "claude.usage.5h" and "claude.usage.7d"
			parentKeys := make(map[string]bool)
			for _, row := range out.Rows {
				parentKeys[row.Key] = true
			}

			for _, row := range out.Rows {
				// Skip if a parent key exists (e.g. skip "x.y.z" if "x.y" exists)
				if dot := strings.LastIndex(row.Key, "."); dot > 0 {
					parent := row.Key[:dot]
					if parentKeys[parent] {
						continue
					}
				}

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

func renderTable(outputs []*module.Output, cfg *config.Config, theme *ThemePalette) string {
	entries := flattenRows(outputs, theme)
	if len(entries) == 0 {
		return ""
	}

	if cfg.KeyStyle == "tree" {
		entries = treeifyKeys(entries)
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


// treeifyKeys transforms flat dotted keys into tree-structured display keys.
// Keys without a dot prefix are kept as-is.
// Keys sharing a prefix are grouped: the first gets a group header,
// children get ├── / └── prefixes.
func treeifyKeys(entries []struct{ key, value string }) []struct{ key, value string } {
	type entry = struct{ key, value string }
	var result []entry

	// Group entries by their prefix (part before first dot)
	i := 0
	for i < len(entries) {
		key := entries[i].key
		dot := strings.Index(key, ".")
		if dot < 0 {
			// No dot — standalone key, keep as-is
			result = append(result, entries[i])
			i++
			continue
		}

		// Find all entries with the same prefix
		prefix := key[:dot]
		groupStart := i
		for i < len(entries) {
			k := entries[i].key
			d := strings.Index(k, ".")
			if d < 0 || k[:d] != prefix {
				break
			}
			i++
		}
		group := entries[groupStart:i]

		if len(group) == 1 {
			// Single entry with dot — keep flat
			result = append(result, group[0])
			continue
		}

		// Multiple entries — emit group header + tree children
		result = append(result, entry{key: prefix, value: ""})
		for j, e := range group {
			child := e.key[dot+1:] // strip prefix + dot
			isLast := j == len(group)-1
			var connector, continuation string
			if isLast {
				connector = "└── "
				continuation = "    "
			} else {
				connector = "├── "
				continuation = "│   "
			}

			// If value is multiline, pad the key column with continuation lines
			// so the tree connector doesn't break visually
			valueLines := strings.Count(e.value, "\n")
			key := connector + child
			for k := 0; k < valueLines; k++ {
				key += "\n" + continuation
			}

			result = append(result, entry{key: key, value: e.value})
		}
	}

	return result
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
