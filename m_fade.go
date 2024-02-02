package main

import (
	"image"
	"image/color"
	"image/draw"

	"fyne.io/fyne/v2/canvas"
)

type cropMask struct {
	bounds image.Rectangle
	crop   image.Rectangle
}

func (c *cropMask) ColorModel() color.Model {
	return color.AlphaModel
}

func (c *cropMask) Bounds() image.Rectangle {
	return c.bounds
}

func (c *cropMask) At(x, y int) color.Color {
	if p := image.Pt(x, y); p.In(c.crop) {
		return color.Alpha{255}
	}
	fade := uint8(256 * (1 - DefaultTranslucency))
	return color.Alpha{fade}
}

// fade photo image outside crop bounds
func (p *Photo) fadeByCrop(m *canvas.Image) {
	src := m.Image
	dst := image.NewNRGBA(src.Bounds())
	if !p.isCropped() { // TODO maybe simply return src?
		draw.Draw(dst, dst.Bounds(), src, image.Point{}, draw.Src)
		m.Image = dst
		return
	}
	factor := float32(p.height) / float32(src.Bounds().Dy())
	crop := image.Rectangle{
		Min: image.Point{
			X: int(float32(p.CropRectangle.Min.X) / factor),
			Y: int(float32(p.CropRectangle.Min.Y) / factor),
		},
		Max: image.Point{
			X: int(float32(p.CropRectangle.Max.X) / factor),
			Y: int(float32(p.CropRectangle.Max.Y) / factor),
		},
	}
	draw.DrawMask(dst, dst.Bounds(), src, image.Point{}, &cropMask{src.Bounds(), crop}, image.Point{}, draw.Over)
	m.Image = dst
}
