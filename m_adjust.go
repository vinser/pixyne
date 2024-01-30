package main

import (
	"image"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
)

const (
	AdjustBrightness = iota
	AdjustContrast
	AdjustHue
	AdjustSaturation
	AdjustGamma
)

type adjustFilterOptions struct {
	name   string
	zero   float64
	min    float64
	max    float64
	step   float64
	adjust func(image.Image, float64) *image.NRGBA
}

var adjustFiltersDict = map[int]adjustFilterOptions{
	AdjustBrightness: {"Brightness", 0, -100, 100, 1, imaging.AdjustBrightness},
	AdjustContrast:   {"Contrast", 0, -100, 100, 1, imaging.AdjustContrast},
	AdjustHue:        {"Hue", 0, -180, 180, 1, imaging.AdjustHue},
	AdjustSaturation: {"Saturation", 0, -100, 100, 1, imaging.AdjustSaturation},
	AdjustGamma:      {"Gamma", 1, 0.1, 5, 0.1, imaging.AdjustGamma},
}

type AdjustFilters struct {
	Slider []*widget.Slider
}

func (f *AdjustFilters) doAdjust(dst *canvas.Image, src image.Image, filter int, value float64) {
	var tmp image.Image = imaging.Clone(src)
	for i, s := range f.Slider {
		if i == filter && filter >= 0 && filter < len(adjustFiltersDict) {
			s.Value = value
		}
		if s.Value != adjustFiltersDict[i].zero {
			tmp = adjustFiltersDict[i].adjust(tmp, s.Value)
		}
	}
	dst.Image = tmp
	dst.Refresh()
}

func (p *Photo) adjustByFilters(m *canvas.Image) {
	src := m.Image
	var tmp image.Image = imaging.Clone(src)
	for i, a := range p.Adjust {
		if a != adjustFiltersDict[i].zero {
			tmp = adjustFiltersDict[i].adjust(tmp, a)
		}
	}
	m.Image = tmp
}
