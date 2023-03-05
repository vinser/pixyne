package settings

import (
	"encoding/json"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// access to user interfaces to control Fyne settings
type Settings struct {
	fyneSettings app.SettingsSchema
	colors       []fyne.CanvasObject
}

func (s *Settings) load() {
	err := s.loadFromFile(s.fyneSettings.StoragePath())
	if err != nil {
		fyne.LogError("Settings load error:", err)
	}
}

func (s *Settings) loadFromFile(path string) error {
	file, err := os.Open(path) // #nosec
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(filepath.Dir(path), 0700)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	decode := json.NewDecoder(file)

	return decode.Decode(&s.fyneSettings)
}

func (s *Settings) save() error {
	return s.saveToFile(s.fyneSettings.StoragePath())
}

func (s *Settings) saveToFile(path string) error {
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil { // this is not an exists error according to docs
		return err
	}

	data, err := json.Marshal(&s.fyneSettings)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func settingsScreen() {
	s := newSettings()
	appearance := widget.NewForm(
		widget.NewFormItem("", s.scalePreviewsRow(wMain.Canvas().Scale())),
		widget.NewFormItem("Scale", s.scalesRow()),
		widget.NewFormItem("Main Color", s.colorsRow()),
		widget.NewFormItem("Theme", s.themesRow()),
	)
	dialog.ShowCustom("Appearance", "Ok", appearance, wMain)
}

// a new settings instance with the current configuration loaded
func newSettings() *Settings {
	s := &Settings{}
	s.load()
	if s.fyneSettings.Scale == 0 {
		s.fyneSettings.Scale = 1
	}
	return s
}

func (s *Settings) applySettings() {
	if s.fyneSettings.Scale == 0.0 {
		s.chooseScale(1.0)
	}
	err := s.save()
	if err != nil {
		fyne.LogError("Failed on saving", err)
	}

	s.appliedScale(s.fyneSettings.Scale)
}
