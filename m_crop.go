package main

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
)

func (a *App) newCropToolbar() {
	a.cropToolbar = container.NewHBox(
		widget.NewButton("X", func() { a.doCropDialog(0.) }),
		widget.NewButton("Free", func() { a.doCropDialog(-1.) }),
		widget.NewButton("1:1", func() { a.doCropDialog(1.) }),
		widget.NewButton("16:9", func() { a.doCropDialog(16. / 9.) }),
		widget.NewButton("9:16", func() { a.doCropDialog(9. / 16.) }),
		widget.NewButton("5:4", func() { a.doCropDialog(5. / 4.) }),
		widget.NewButton("4:5", func() { a.doCropDialog(4. / 5.) }),
	)
	a.cropToolbar.Hide()
}

func (a *App) doCropDialog(dstAspect float32) {
	p := list[a.state.FramePos+frame.ItemPos]
	w := a.topWindow
	mi := frame.Items[frame.ItemPos].Img
	m := GetListImageAt(p)
	if dstAspect == 0 {
		p.CropRectangle = image.Rect(0, 0, 0, 0)
		mi.Image = m.Image
		mi.Refresh()
		a.cropToolbar.Hide()
		a.toolBar.Show()
		return
	}
	scaleFactor := a.Settings().Scale()
	fitFactor := float32(ScreenHeight) / float32(m.Image.Bounds().Dy()) * downscaleFactor / scaleFactor
	img := canvas.NewImageFromImage(imaging.Resize(m.Image, 0, int(float32(m.Image.Bounds().Dy())*fitFactor+0.5), imaging.Box))
	img.FillMode = canvas.ImageFillOriginal
	img.ScaleMode = canvas.ImageScaleFastest
	border := canvas.NewRectangle(color.Transparent)
	border.StrokeColor = theme.PrimaryColor()
	border.StrokeWidth = 5
	center := container.NewWithoutLayout(img, border)
	imgSize := fyne.NewSize(float32(img.Image.Bounds().Dx())/scaleFactor, float32(img.Image.Bounds().Dy())/scaleFactor)
	img.Move(fyne.NewPos(0, 0))
	img.Resize(imgSize)
	var content fyne.CanvasObject
	var posTL, posBR fyne.Position
	refreshBorder := func() {
		border.Move(posTL)
		border.Resize(fyne.NewSize(posBR.X-posTL.X, posBR.Y-posTL.Y))
	}

	topEdge := widget.NewSlider(0, float64(imgSize.Height))
	topEdge.Orientation = widget.Vertical
	leftEdge := widget.NewSlider(0, float64(imgSize.Width))
	leftEdge.Orientation = widget.Horizontal
	bottomEdge := widget.NewSlider(0, float64(imgSize.Height))
	bottomEdge.Orientation = widget.Vertical
	rightEdge := widget.NewSlider(0, float64(imgSize.Width))
	rightEdge.Orientation = widget.Horizontal

	refreshSliders := func() {
		topEdge.Value = float64(imgSize.Height - posTL.Y)
		topEdge.Refresh()
		leftEdge.Value = float64(posTL.X)
		leftEdge.Refresh()
		bottomEdge.Value = float64(imgSize.Height - posBR.Y)
		bottomEdge.Refresh()
		rightEdge.Value = float64(posBR.X)
		rightEdge.Refresh()
	}

	srcAspect := img.Aspect()
	// log.Print(p.width, p.height)
	// log.Print(imgSize.Width, imgSize.Height)
	// log.Print(srcAspect, dstAspect)
	if dstAspect < 0 {
		posTL = fyne.NewPos(0, 0)
		posBR = fyne.NewPos(imgSize.Width, imgSize.Height)
		refreshBorder()

		topEdge.Value = float64(imgSize.Height)
		leftEdge.Value = 0
		bottomEdge.Value = 0
		rightEdge.Value = float64(imgSize.Width)

		topEdge.OnChanged = func(v float64) {
			if v <= bottomEdge.Value {
				return
			}
			posTL.Y = imgSize.Height - float32(v)
			refreshBorder()
		}
		leftEdge.OnChanged = func(v float64) {
			if v >= rightEdge.Value {
				return
			}
			posTL.X = float32(v)
			refreshBorder()
		}
		bottomEdge.OnChanged = func(v float64) {
			if v >= topEdge.Value {
				return
			}
			posBR.Y = imgSize.Height - float32(v)
			refreshBorder()
		}
		rightEdge.OnChanged = func(v float64) {
			if v <= leftEdge.Value {
				return
			}
			posBR.X = float32(v)
			refreshBorder()
		}

		top := container.NewVBox(rightEdge)
		bottom := container.NewVBox(leftEdge)
		right := container.NewStack(topEdge)
		left := container.NewStack(bottomEdge)
		content = container.NewBorder(top, bottom, left, right, center)
	} else {
		if dstAspect <= srcAspect {
			posTL = fyne.NewPos(0, 0)
			posBR = fyne.NewPos(imgSize.Height*dstAspect, imgSize.Height)
		} else {
			posTL = fyne.NewPos(0, imgSize.Height-imgSize.Width/dstAspect)
			posBR = fyne.NewPos(imgSize.Width, imgSize.Height)

		}
		refreshBorder()
		refreshSliders()

		leftEdge.OnChanged = func(v float64) {
			x := float32(v)
			dx := posBR.X - posTL.X
			if x >= posBR.X || x+dx >= imgSize.Width {
				return
			}
			posTL.X = x
			posBR.X = posTL.X + dx
			refreshBorder()
			refreshSliders()
		}
		bottomEdge.OnChanged = func(v float64) {
			y := imgSize.Height - float32(v)
			dy := posBR.Y - posTL.Y
			if v >= topEdge.Value || dy >= y {
				return
			}
			posBR.Y = y
			posTL.Y = posBR.Y - dy
			refreshBorder()
			refreshSliders()
		}
		rightEdge.OnChanged = func(v float64) {
			x := float32(v)
			if x <= posTL.X || posBR.Y <= (x-posTL.X)/dstAspect {
				return
			}
			posBR.X = x
			posTL.Y = posBR.Y - (x-posTL.X)/dstAspect
			refreshBorder()
			refreshSliders()
		}
		topEdge.OnChanged = func(v float64) {
			y := imgSize.Height - float32(v)
			if y >= posBR.Y || (posBR.Y-y)*dstAspect+posTL.X >= imgSize.Width {
				return
			}
			posTL.Y = y
			posBR.X = (posBR.Y-y)*dstAspect + posTL.X
			refreshBorder()
			refreshSliders()
		}

		top := container.NewVBox(rightEdge)
		bottom := container.NewVBox(leftEdge)
		right := container.NewStack(topEdge)
		left := container.NewStack(bottomEdge)
		content = container.NewBorder(top, bottom, left, right, center)
	}

	dlg := dialog.NewCustomConfirm("Crop image", "Crop", "Cancel", content, func(b bool) {
		if b {
			factor := float32(p.height) / float32(m.Image.Bounds().Dy()) * scaleFactor / fitFactor
			posTL.X *= factor
			posTL.Y *= factor
			posBR.X *= factor
			posBR.Y *= factor
			p.CropRectangle = image.Rect(int(posTL.X), int(posTL.Y), int(posBR.X), int(posBR.Y))
			mi.Image = m.Image
			p.fadeByCrop(mi)
			mi.Refresh()
			a.cropToolbar.Hide()
			a.toolBar.Show()
		}
	}, w)
	jolt(w)
	dlg.Show()
}

// Attempt workaround for sliders malfunctioning
const joltAmplitude = 5.0

func jolt(w fyne.Window) {
	s := w.Content().Size().AddWidthHeight(joltAmplitude, joltAmplitude)
	w.Resize(s)
	s = w.Content().Size().AddWidthHeight(-joltAmplitude, -joltAmplitude)
	w.Resize(s)
}
