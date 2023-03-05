package settings

import (
	"runtime"

	"fyne.io/fyne/v2/widget"
)

const (
	systemThemeName = "system"
)

func (s *Settings) themesRow() *widget.RadioGroup {
	def := s.fyneSettings.ThemeName
	themeNames := []string{"dark", "light"}
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		themeNames = append(themeNames, systemThemeName)
		if s.fyneSettings.ThemeName == "" {
			def = systemThemeName
		}
	}
	themes := widget.NewRadioGroup(themeNames, s.chooseTheme)
	themes.SetSelected(def)
	themes.Horizontal = true
	return themes
}

func (s *Settings) chooseTheme(name string) {
	if name == systemThemeName {
		name = ""
	}
	s.fyneSettings.ThemeName = name
	s.applySettings()
}
