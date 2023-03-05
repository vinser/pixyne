package settings

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

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
