package main

import (
	"encoding/json"
	"image/color"
	"os"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
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

// a new settings instance with the current configuration loaded
func NewSettings() *Settings {
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

// scale
type scaleItems struct {
	scale   float32
	name    string
	preview *canvas.Text
	button  *widget.Button
}

var scales = []*scaleItems{
	{scale: 0.5, name: "Tiny"},
	{scale: 0.8, name: "Small"},
	{scale: 1, name: "Normal"},
	{scale: 1.3, name: "Large"},
	{scale: 1.8, name: "Huge"}}

func (s *Settings) appliedScale(value float32) {
	for _, scale := range scales {
		scale.preview.TextSize = theme.TextSize() * scale.scale / value
	}
}

func (s *Settings) chooseScale(value float32) {
	s.fyneSettings.Scale = value

	for _, scale := range scales {
		if scale.scale == value {
			scale.button.Importance = widget.HighImportance
		} else {
			scale.button.Importance = widget.MediumImportance
		}

		scale.button.Refresh()
	}
	s.applySettings()
}

func (s *Settings) scalesRow() *fyne.Container {
	var buttons = make([]fyne.CanvasObject, len(scales))
	for i, scale := range scales {
		value := scale.scale
		button := widget.NewButton(scale.name, func() {
			s.chooseScale(value)
		})
		if s.fyneSettings.Scale == scale.scale {
			button.Importance = widget.HighImportance
		}

		scale.button = button
		buttons[i] = button
	}
	return container.NewGridWithColumns(len(scales), buttons...)
}

func (s *Settings) scalePreviewsRow(value float32) *fyne.Container {
	var previews = make([]fyne.CanvasObject, len(scales))
	for i, scale := range scales {
		text := canvas.NewText("A", theme.ForegroundColor())
		text.Alignment = fyne.TextAlignCenter
		text.TextSize = theme.TextSize() * scale.scale / value

		scale.preview = text
		previews[i] = text
	}
	return container.NewGridWithColumns(len(scales), previews...)
}

// color
type colorButton struct {
	widget.BaseWidget
	name  string
	color color.Color

	s *Settings
}

func (s *Settings) colorsRow() *fyne.Container {
	for _, c := range theme.PrimaryColorNames() {
		b := newColorButton(c, theme.PrimaryColorNamed(c), s)
		s.colors = append(s.colors, b)
	}
	return container.NewGridWithColumns(len(s.colors), s.colors...)
}

func newColorButton(n string, c color.Color, s *Settings) *colorButton {
	b := &colorButton{name: n, color: c, s: s}
	b.ExtendBaseWidget(b)
	return b
}

func (b *colorButton) CreateRenderer() fyne.WidgetRenderer {
	r := canvas.NewRectangle(b.color)
	r.StrokeWidth = 2

	if b.name == b.s.fyneSettings.PrimaryColor {
		r.StrokeColor = theme.PrimaryColor()
	}

	return &colorRenderer{btn: b, rect: r, objs: []fyne.CanvasObject{r}}
}

func (b *colorButton) Tapped(_ *fyne.PointEvent) {
	b.s.fyneSettings.PrimaryColor = b.name
	for _, child := range b.s.colors {
		child.Refresh()
	}
	b.s.applySettings()
}

type colorRenderer struct {
	btn  *colorButton
	rect *canvas.Rectangle
	objs []fyne.CanvasObject
}

func (r *colorRenderer) Layout(s fyne.Size) {
	r.rect.Resize(s)
}

func (r *colorRenderer) MinSize() fyne.Size {
	return fyne.NewSize(20, 20)
}

func (r *colorRenderer) Refresh() {
	if r.btn.name == r.btn.s.fyneSettings.PrimaryColor {
		r.rect.StrokeColor = theme.PrimaryColor()
	} else {
		r.rect.StrokeColor = color.Transparent
	}
	r.rect.FillColor = r.btn.color

	r.rect.Refresh()
}

func (r *colorRenderer) Objects() []fyne.CanvasObject {
	return r.objs
}

func (r *colorRenderer) Destroy() {
}

// theme
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
