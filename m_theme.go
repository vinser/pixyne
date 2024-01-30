//go:generate fyne bundle --package main --name appIconDark --output m_bundled.go icons/appIconDark.png
//go:generate fyne bundle --name appIconLight --output m_bundled.go --append icons/appIconLight.png
//go:generate fyne bundle --name iconScrollBack --output m_bundled.go --append icons/scroll-back.svg
//go:generate fyne bundle --name iconAdjust --output m_bundled.go --append icons/adjust.svg
//go:generate fyne bundle --name iconBlank --output m_bundled.go --append icons/blank.svg

package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Application custom theme and interface inplementation
type Theme struct{}

func (t *Theme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	variant := theme.VariantDark
	switch a.state.Theme {
	case "light":
		variant = theme.VariantLight
	case "dark":
		variant = theme.VariantDark
	}
	switch {
	case name == theme.ColorNamePrimary:
		return theme.PrimaryColorNamed(a.state.Color)
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
	return theme.DefaultTheme().Font(style)
}

func (t *Theme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *Theme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name) * a.state.Scale
}
