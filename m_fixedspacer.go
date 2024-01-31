package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

type fixedSpacer struct {
	*canvas.Rectangle
}

func newFixedSpacer(size fyne.Size) *fixedSpacer {
	background := canvas.NewRectangle(color.Transparent)
	background.StrokeColor = color.Transparent
	background.StrokeWidth = theme.Padding()
	background.CornerRadius = 0
	background.SetMinSize(size)
	spacer := &fixedSpacer{Rectangle: background}
	return spacer
}

func (s *fixedSpacer) ToolbarObject() fyne.CanvasObject {
	return s
}
