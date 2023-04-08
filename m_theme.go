//go:generate fyne bundle --package main --name appIcon --output m_bundled.go appIcon.png
//go:generate fyne bundle --name fontRegular --output m_bundled.go --append fonts/Inter-Regular.ttf
//go:generate fyne bundle --name fontBold --output m_bundled.go --append fonts/Inter-Bold.ttf
//go:generate fyne bundle --name fontItalic --output m_bundled.go --append fonts/Inter-Italic.ttf
//go:generate fyne bundle --name fontBoldItalic --output m_bundled.go --append fonts/Inter-BoldItalic.ttf

//go:generate fyne bundle --name iconScrollBack --output m_bundled.go --append icons/scroll-back.svg

package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Application custom theme and interface inplementation
type Theme struct{}

func (t *Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch {
	case name == theme.ColorNameButton:
		return color.Transparent
	case name == theme.ColorNameDisabled && variant == theme.VariantDark:
		return color.NRGBA{R: 100, G: 100, B: 100, A: 255}
	case name == theme.ColorNameDisabled && variant == theme.VariantLight:
		return color.NRGBA{R: 180, G: 180, B: 180, A: 255}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (t *Theme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return fontRegular
	}
	if style.Bold {
		if style.Italic {
			return fontBoldItalic
		}
		return fontBold
	}
	if style.Italic {
		return fontItalic
	}
	if style.Symbol {
		return fontRegular
	}
	return fontRegular
}

func (t *Theme) Icon(name fyne.ThemeIconName) fyne.Resource {
	// switch name {
	// case theme.IconNameRadioButton:
	// 	return theme.NewThemedResource(iconRadioFree)
	// case theme.IconNameRadioButtonChecked:
	// 	return theme.NewThemedResource(iconRadioChecked)
	// }
	return theme.DefaultTheme().Icon(name)
}

func (t *Theme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
