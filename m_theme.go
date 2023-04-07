//go:generate fyne bundle --package main --name appIcon --output m_bundled.go appIcon.png
//go:generate fyne bundle --name fontRegular --output m_bundled.go --append fonts/Inter-Regular.ttf
//go:generate fyne bundle --name fontBold --output m_bundled.go --append fonts/Inter-Bold.ttf
//go:generate fyne bundle --name fontItalic --output m_bundled.go --append fonts/Inter-Italic.ttf
//go:generate fyne bundle --name fontBoldItalic --output m_bundled.go --append fonts/Inter-BoldItalic.ttf

//go:generate fyne bundle --name iconAbout --output m_bundled.go --append icons/about.svg
//go:generate fyne bundle --name iconHelp --output m_bundled.go --append icons/help.svg
//go:generate fyne bundle --name iconDistpayGrid --output m_bundled.go --append icons/display-grid.svg
//go:generate fyne bundle --name iconDistpayList --output m_bundled.go --append icons/display-list.svg
//go:generate fyne bundle --name iconFolderOpen --output m_bundled.go --append icons/folder-open.svg
//go:generate fyne bundle --name iconFolderSave --output m_bundled.go --append icons/folder-save.svg
//go:generate fyne bundle --name iconFrameShrink --output m_bundled.go --append icons/frame-shrink.svg
//go:generate fyne bundle --name iconFrameStretch --output m_bundled.go --append icons/frame-stretch.svg
//go:generate fyne bundle --name iconScrollFirst --output m_bundled.go --append icons/scroll-first.svg
//go:generate fyne bundle --name iconScrollLast --output m_bundled.go --append icons/scroll-last.svg
//go:generate fyne bundle --name iconScrollNextFrame --output m_bundled.go --append icons/scroll-forward.svg
//go:generate fyne bundle --name iconScrollNext --output m_bundled.go --append icons/scroll-next.svg
//go:generate fyne bundle --name iconScrollPrevFrame --output m_bundled.go --append icons/scroll-backward.svg
//go:generate fyne bundle --name iconScrollPrev --output m_bundled.go --append icons/scroll-previous.svg
//go:generate fyne bundle --name iconSettings --output m_bundled.go --append icons/settings.svg

//go:generate fyne bundle --name iconRadioFree --output m_bundled.go --append icons/radio-free.svg
//go:generate fyne bundle --name iconRadioChecked --output m_bundled.go --append icons/radio-checked.svg

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
	switch name {
	case theme.IconNameRadioButton:
		return theme.NewThemedResource(iconRadioFree)
	case theme.IconNameRadioButtonChecked:
		return theme.NewThemedResource(iconRadioChecked)
	}
	return theme.DefaultTheme().Icon(name)
}

func (t *Theme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
