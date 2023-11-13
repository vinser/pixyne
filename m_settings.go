package main

import (
	"encoding/json"
	"image/color"
	"os"
	"path/filepath"
	"runtime"
	"time"

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
		s.choosedScale(1.0)
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
	{scale: 1.3, name: "Big"},
	{scale: 1.8, name: "Large"},
	{scale: 2.2, name: "Huge"}}

func (s *Settings) appliedScale(value float32) {
	for _, scale := range scales {
		scale.preview.TextSize = theme.TextSize() * scale.scale / value
	}
}

func (s *Settings) choosedScale(value float32) {
	s.fyneSettings.Scale = value

	for _, scale := range scales {
		if scale.scale == value {
			scale.button.Importance = widget.HighImportance
			scale.button.Refresh()
		} else {
			scale.button.Importance = widget.MediumImportance
			scale.button.Refresh()
		}
	}
	s.applySettings()
}

func (s *Settings) scalesRow() *fyne.Container {
	var buttons = make([]fyne.CanvasObject, len(scales))
	for i, scale := range scales {
		value := scale.scale
		button := widget.NewButton(scale.name, func() {
			s.choosedScale(value)
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
func (s *Settings) themesRow() *widget.RadioGroup {
	def := s.fyneSettings.ThemeName
	themeNames := []string{"dark", "light"}
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		themeNames = append(themeNames, "system")
		if s.fyneSettings.ThemeName == "" {
			def = "system"
		}
	}
	themes := widget.NewRadioGroup(themeNames, func(selected string) {
		if selected == "system" {
			selected = ""
		}
		s.fyneSettings.ThemeName = selected
		s.applySettings()
	})
	themes.SetSelected(def)
	themes.Horizontal = true
	return themes
}

// date format
func (s *Settings) datesRow(a *App) *fyne.Container {
	toFormat := map[string]string{
		"DD.MM.YYYY": "02.01.2006 15:04:05",
		"MM.DD.YYYY": "01.02.2006 15:04:05",
		"YYYY.MM.DD": "2006.01.02 15:04:05",
	}
	byFormat := make(map[string]string, len(toFormat))
	for k, v := range toFormat {
		byFormat[v] = k
	}
	display := widget.NewLabel(time.Now().Format(DisplayDateFormat))
	label := widget.NewLabel("Sample:")
	label.Alignment = fyne.TextAlignTrailing
	options := []string{}
	for k := range toFormat {
		options = append(options, k)
	}
	choice := widget.NewSelectEntry(options)
	choice.SetText(byFormat[DisplayDateFormat])
	choice.OnChanged = func(s string) {
		DisplayDateFormat = toFormat[s]
		if a.frameView.Hidden {
			a.listTable.Refresh()
		} else {
			a.scrollFrame(frame.Pos)
		}
		display.SetText(time.Now().Format(DisplayDateFormat))
	}
	return container.NewGridWithColumns(3, choice, label, display)
}

func (s *Settings) modeRow(a *App) *widget.RadioGroup {
	mode := widget.NewRadioGroup([]string{"full", "simple"}, func(selected string) {
		if selected == "simple" {
			a.simpleMode = true
		} else {
			a.simpleMode = false
		}
		a.scrollFrame(frame.Pos)

	})
	if a.simpleMode {
		mode.SetSelected("simple")
	} else {
		mode.SetSelected("full")
	}
	mode.Horizontal = true
	return mode
}
