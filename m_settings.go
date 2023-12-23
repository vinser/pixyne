package main

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	DefaultScale = 1.3
	DefaultTheme = "dark"
)

// access to user interfaces to control Fyne settings
type Settings struct {
	colors []fyne.CanvasObject
}

// a new settings instance with the current configuration loaded
func NewSettings() *Settings {
	s := &Settings{}
	return s
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
	{scale: 1.0, name: "Normal"},
	{scale: 1.3, name: "Big"},
	{scale: 1.8, name: "Large"},
	{scale: 2.2, name: "Huge"}}

func (s *Settings) choosedScale(value float32) {
	a.state.Scale = value

	for _, scale := range scales {
		if scale.scale == value {
			scale.button.Importance = widget.HighImportance
			scale.button.Refresh()
		} else {
			scale.button.Importance = widget.MediumImportance
			scale.button.Refresh()
		}
	}
}

func (s *Settings) scalesRow() *fyne.Container {
	var buttons = make([]fyne.CanvasObject, len(scales))
	for i, scale := range scales {
		value := scale.scale
		button := widget.NewButton(scale.name, func() {
			s.choosedScale(value)
		})
		if a.state.Scale == scale.scale {
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

	if b.name == a.state.Color {
		r.StrokeColor = theme.PrimaryColor()
	}

	return &colorRenderer{btn: b, rect: r, objs: []fyne.CanvasObject{r}}
}

func (b *colorButton) Tapped(_ *fyne.PointEvent) {
	a.state.Color = b.name
	for _, child := range b.s.colors {
		child.Refresh()
	}
	a.Settings().SetTheme(&Theme{})
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
	if r.btn.name == a.state.Color {
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
	themeNames := []string{"Dark", "Light"}
	themes := widget.NewRadioGroup(themeNames, func(selected string) {
		if selected == "Dark" {
			a.state.Theme = "dark"

		} else {
			a.state.Theme = "light"
		}
		a.Settings().SetTheme(&Theme{})
	})
	if a.state.Theme == "dark" {
		themes.SetSelected("Dark")
	} else {
		themes.SetSelected("Light")
	}
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
		display.SetText(time.Now().Format(DisplayDateFormat))
	}
	return container.NewGridWithColumns(3, choice, label, display)
}

func (s *Settings) modeRow(a *App) *widget.RadioGroup {
	mode := widget.NewRadioGroup([]string{"Simple", "Advanced"},
		func(selected string) {
			if selected == "Simple" {
				a.state.Simple = true
			} else {
				a.state.Simple = false
			}
		})
	if a.state.Simple {
		mode.SetSelected("Simple")
	} else {
		mode.SetSelected("Advanced")
	}
	mode.Horizontal = true
	return mode
}

func (s *Settings) hotkeysRow() *widget.Label {
	hotkey := "No shortcut for hotkeys"
	for i := range a.ControlShortCuts {
		if a.ControlShortCuts[i].Name == "Hotkeys" {
			hotkey = "To see hotkeys press Ctrl + " + string(a.ControlShortCuts[i].KeyName)
		}
	}
	for i := range a.AltShortCuts {
		if a.AltShortCuts[i].Name == "Hotkeys" {
			hotkey = "To see hotkeys press Alt + " + string(a.AltShortCuts[i].KeyName)
		}
	}
	lbl := widget.NewLabel(hotkey)
	return lbl
}
