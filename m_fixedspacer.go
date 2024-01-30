package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type fixedSpacerRenderer struct {
	size fyne.Size

	background *canvas.Rectangle
	objects    []fyne.CanvasObject
	spacer     *fixedSpacer
}

func (r *fixedSpacerRenderer) Layout(_ fyne.Size) {
}

func (r *fixedSpacerRenderer) MinSize() fyne.Size {
	return r.size
}

func (r *fixedSpacerRenderer) Refresh() {
	r.background.Refresh()
}

func (r *fixedSpacerRenderer) Destroy() {

}

func (r *fixedSpacerRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *fixedSpacerRenderer) SetObjects(objects []fyne.CanvasObject) {
	r.objects = objects
}

type fixedSpacer struct {
	widget.BaseWidget
	size fyne.Size
}

func (s *fixedSpacer) CreateRenderer() fyne.WidgetRenderer {
	s.ExtendBaseWidget(s)
	background := canvas.NewRectangle(color.Transparent)
	background.StrokeColor = color.Transparent
	background.StrokeWidth = theme.Padding()
	background.CornerRadius = 0
	r := &fixedSpacerRenderer{
		size:       s.size,
		background: background,
		objects:    []fyne.CanvasObject{background},
		spacer:     s,
	}
	return r
}

func (s *fixedSpacer) Resize(_ fyne.Size) {
}

func newFixedSpacer(size fyne.Size) *fixedSpacer {
	spacer := &fixedSpacer{size: size}
	spacer.ExtendBaseWidget(spacer)
	return spacer
}

func (s *fixedSpacer) ToolbarObject() fyne.CanvasObject {
	return s
}
