package render

import "github.com/babarot/enter/internal/module"

type RGB struct {
	R, G, B uint8
}

type ThemePalette struct {
	Primary   RGB
	Secondary RGB
	Success   RGB
	Warning   RGB
	Danger    RGB
	Muted     RGB
	Accent    RGB
}

func GetTheme(name string) *ThemePalette {
	switch name {
	case "tokyo-night":
		return &ThemePalette{
			Primary:   RGB{122, 162, 247},
			Secondary: RGB{192, 202, 245},
			Success:   RGB{158, 206, 106},
			Warning:   RGB{224, 175, 104},
			Danger:    RGB{247, 118, 142},
			Muted:     RGB{108, 115, 148},
			Accent:    RGB{187, 154, 247},
		}
	case "catppuccin":
		return &ThemePalette{
			Primary:   RGB{137, 180, 250},
			Secondary: RGB{205, 214, 244},
			Success:   RGB{166, 227, 161},
			Warning:   RGB{249, 226, 175},
			Danger:    RGB{243, 139, 168},
			Muted:     RGB{127, 132, 156},
			Accent:    RGB{203, 166, 247},
		}
	case "dracula":
		return &ThemePalette{
			Primary:   RGB{139, 233, 253},
			Secondary: RGB{248, 248, 242},
			Success:   RGB{80, 250, 123},
			Warning:   RGB{241, 250, 140},
			Danger:    RGB{255, 85, 85},
			Muted:     RGB{98, 114, 164},
			Accent:    RGB{189, 147, 249},
		}
	case "nord":
		return &ThemePalette{
			Primary:   RGB{136, 192, 208},
			Secondary: RGB{216, 222, 233},
			Success:   RGB{163, 190, 140},
			Warning:   RGB{235, 203, 139},
			Danger:    RGB{191, 97, 106},
			Muted:     RGB{76, 86, 106},
			Accent:    RGB{180, 142, 173},
		}
	case "gruvbox":
		return &ThemePalette{
			Primary:   RGB{131, 165, 152},
			Secondary: RGB{235, 219, 178},
			Success:   RGB{184, 187, 38},
			Warning:   RGB{250, 189, 47},
			Danger:    RGB{251, 73, 52},
			Muted:     RGB{146, 131, 116},
			Accent:    RGB{211, 134, 155},
		}
	default:
		return &ThemePalette{
			Primary:   RGB{96, 165, 250},
			Secondary: RGB{209, 213, 219},
			Success:   RGB{74, 222, 128},
			Warning:   RGB{251, 191, 36},
			Danger:    RGB{248, 113, 113},
			Muted:     RGB{156, 163, 175},
			Accent:    RGB{192, 132, 252},
		}
	}
}

func ColorForSemantic(color module.SemanticColor, theme *ThemePalette) *RGB {
	switch color {
	case module.Primary:
		return &theme.Primary
	case module.Secondary:
		return &theme.Secondary
	case module.Success:
		return &theme.Success
	case module.Warning:
		return &theme.Warning
	case module.Danger:
		return &theme.Danger
	case module.Muted:
		return &theme.Muted
	case module.Accent:
		return &theme.Accent
	default:
		return nil
	}
}
