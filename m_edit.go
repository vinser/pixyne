package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
)

type Editor struct {
	w fyne.Window
	// Edit dialog
	dialog *dialog.ConfirmDialog
	photo  *Photo
	imgSrc image.Image
	img    *canvas.Image
	// Adjust
	adjustFilters *AdjustFilters
	// Crop
	cropsMenu        *fyne.Container
	fixedCrops       *fyne.Container
	sliderTopEdge    *widget.Slider
	sliderBottomEdge *widget.Slider
	sliderLeftEdge   *widget.Slider
	sliderRightEdge  *widget.Slider
	//
	imgSize fyne.Size
	border  *canvas.Rectangle
	posTL   fyne.Position
	posBR   fyne.Position
}

func (a *App) editDialog() {
	a.disableShortcuts()
	e := &Editor{
		adjustFilters: &AdjustFilters{
			Slider: []*widget.Slider{},
		},
	}
	e.photo = list[a.state.FramePos+a.state.ItemPos]
	e.w = a.topWindow

	// Edit frame
	borderedImage := e.newBorderedImage()
	top := container.NewStack(e.sliderRightEdge)
	bottom := container.NewStack(e.sliderLeftEdge)
	right := container.NewStack(e.sliderTopEdge)
	left := container.NewStack(e.sliderBottomEdge)
	// centerEditor := container.NewCenter(container.NewBorder(top, bottom, left, right, borderedImage))
	centerEditor := container.NewBorder(top, bottom, left, right, borderedImage)

	// Right panel
	editBrightness := e.newAdjustControl(AdjustBrightness)
	editContrast := e.newAdjustControl(AdjustContrast)
	editHue := e.newAdjustControl(AdjustHue)
	editSaturation := e.newAdjustControl(AdjustSaturation)
	editGamma := e.newAdjustControl(AdjustGamma)
	widthSpacer := newFixedSpacer(widget.NewLabel("______________________").MinSize())
	e.newCropsMenu()
	groupEditCrop := widget.NewAccordionItem(
		"Crop",
		e.cropsMenu,
	)
	groupEditCrop.Open = true

	groupEditAdjust := widget.NewAccordionItem(
		"Adjust",
		container.NewVBox(
			editBrightness,
			editContrast,
			editHue,
			editSaturation,
			editGamma,
			widget.NewButtonWithIcon("Reset adjusts", theme.ContentClearIcon(), func() { e.resetAdjusts() }),
		),
	)

	rightOptions := container.NewVScroll(
		container.NewVBox(
			widget.NewAccordion(
				groupEditCrop,
				groupEditAdjust,
			),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("Reset all edits", theme.ContentClearIcon(), func() { e.ResetAllEdits() }),
			widthSpacer,
		),
	)

	content := container.NewBorder(nil, nil, nil, rightOptions, centerEditor)
	e.freeCrop() // default active crop
	e.photo.adjustByFilters(e.img)
	e.dialog = dialog.NewCustomConfirm(
		"Edit image",
		"Apply edits",
		"Cancel",
		content,
		func(b bool) {
			if b {
				e.saveEdits()
			}
			a.enableShortcuts()
			e.w.SetFixedSize(false)
		},
		e.w,
	)
	e.dialog.Show()
	jolt(e.w) // TODO: remove this workaround after fix dialog widgets malfunctioning
	e.w.SetFixedSize(true)
}

func (e *Editor) newCropsMenu() {
	c := container.NewVBox()
	c.Objects = append(c.Objects, widget.NewButton("Free", func() { e.freeCrop() }))
	e.fixedCrops = container.NewVBox()
	for i := 0; i < len(a.state.FavoriteCrops); i++ {
		crop := a.state.FavoriteCrops[i]
		button := widget.NewButton(fmt.Sprintf("%.f:%.f", crop.X, crop.Y), func() { e.fixCrop(crop.X / crop.Y) })
		e.fixedCrops.Objects = append(e.fixedCrops.Objects, button)
	}
	c.Objects = append(c.Objects, e.fixedCrops)
	c.Objects = append(c.Objects, widget.NewButtonWithIcon("Clear crop", theme.ContentClearIcon(), func() { e.clearCrop() }))
	c.Objects = append(c.Objects, widget.NewButtonWithIcon("Set crops", theme.SettingsIcon(), func() { e.editCropSettingsDialog() }))
	e.cropsMenu = c
}

func (e *Editor) newAdjustControl(filter int) *fyne.Container {
	text := widget.NewLabel(adjustFiltersDict[filter].name)

	data := binding.NewFloat()
	value := widget.NewLabelWithData(binding.FloatToStringWithFormat(data, "%.1f"))

	slider := widget.NewSliderWithData(adjustFiltersDict[filter].min, adjustFiltersDict[filter].max, data)
	slider.Value = e.photo.Adjust[filter]
	slider.Step = adjustFiltersDict[filter].step
	slider.OnChangeEnded = func(f float64) { e.adjustOnChangeEnded(filter, f) }
	e.adjustFilters.Slider = append(e.adjustFilters.Slider, slider)
	data.Set(e.photo.Adjust[filter])

	return container.NewVBox(
		container.NewHBox(
			text,
			layout.NewSpacer(),
			value,
		),
		slider,
	)
}

func (e *Editor) adjustOnChangeEnded(filter int, value float64) {
	e.adjustFilters.doAdjust(e.img, e.imgSrc, filter, value)
}

func (e *Editor) resetAdjusts() {
	for i, s := range e.adjustFilters.Slider {
		s.Value = adjustFiltersDict[i].zero
		s.OnChanged(s.Value)
		s.OnChangeEnded(s.Value)
		s.Refresh()
	}
	// e.adjustFilters.doAdjust(e.img, e.imgSrc, -1, 0)
}

func (e *Editor) refreshCropSliders() {
	e.sliderTopEdge.Value = float64(e.imgSize.Height - e.posTL.Y)
	e.sliderTopEdge.Refresh()
	e.sliderLeftEdge.Value = float64(e.posTL.X)
	e.sliderLeftEdge.Refresh()
	e.sliderBottomEdge.Value = float64(e.imgSize.Height - e.posBR.Y)
	e.sliderBottomEdge.Refresh()
	e.sliderRightEdge.Value = float64(e.posBR.X)
	e.sliderRightEdge.Refresh()
}

func (e *Editor) refreshCropBorder() {
	e.border.Move(e.posTL)
	e.border.Resize(fyne.NewSize(e.posBR.X-e.posTL.X, e.posBR.Y-e.posTL.Y))
}

func (e *Editor) newBorderedImage() *fyne.Container {
	srcImg := GetListImageAt(e.photo).Image
	sc := e.w.Canvas().Scale()
	wWidth := e.w.Canvas().Size().Width
	wHeight := e.w.Canvas().Size().Height
	iWidth := float32(srcImg.Bounds().Dx()) / sc
	iHeight := float32(srcImg.Bounds().Dy()) / sc
	fX := iWidth / wWidth
	fY := iHeight / wHeight
	if fX < fY {
		dy := int(e.w.Canvas().Size().Height*sc*downscaleFactor - 100)
		e.imgSrc = imaging.Resize(srcImg, 0, dy, imaging.Box)
	} else {
		dx := int(e.w.Canvas().Size().Width*sc*downscaleFactor - 300)
		e.imgSrc = imaging.Resize(srcImg, dx, 0, imaging.Box)
	}

	e.img = canvas.NewImageFromImage(e.imgSrc)
	e.img.FillMode = canvas.ImageFillOriginal
	e.img.ScaleMode = canvas.ImageScaleFastest
	e.imgSize = fyne.NewSize(float32(e.img.Image.Bounds().Dx())/sc, float32(e.img.Image.Bounds().Dy())/sc)
	e.img.Move(fyne.NewPos(0, 0))
	e.img.Resize(e.imgSize)

	e.border = canvas.NewRectangle(color.Transparent)
	e.border.StrokeColor = theme.PrimaryColor()
	e.border.StrokeWidth = 5

	e.sliderTopEdge = widget.NewSlider(0, float64(e.imgSize.Height))
	e.sliderTopEdge.Orientation = widget.Vertical
	e.sliderLeftEdge = widget.NewSlider(0, float64(e.imgSize.Width))
	e.sliderLeftEdge.Orientation = widget.Horizontal
	e.sliderBottomEdge = widget.NewSlider(0, float64(e.imgSize.Height))
	e.sliderBottomEdge.Orientation = widget.Vertical
	e.sliderRightEdge = widget.NewSlider(0, float64(e.imgSize.Width))
	e.sliderRightEdge.Orientation = widget.Horizontal

	return container.NewWithoutLayout(e.img, e.border)
}

// Clear image crop
func (e *Editor) clearCrop() {
	e.photo.CropRectangle = image.Rect(0, 0, 0, 0)
	e.posTL = fyne.NewPos(0, 0)
	e.posBR = fyne.NewPos(e.imgSize.Width, e.imgSize.Height)
	e.refreshCropSliders()
	e.refreshCropBorder()
}

func (e *Editor) freeCrop() {
	if e.photo.isCropped() {
		sc := e.w.Canvas().Scale()
		factor := float32(e.photo.height) / float32(e.img.Image.Bounds().Dy())
		e.posTL.X = float32(e.photo.CropRectangle.Min.X) / factor / sc
		e.posTL.Y = float32(e.photo.CropRectangle.Min.Y) / factor / sc
		e.posBR.X = float32(e.photo.CropRectangle.Max.X) / factor / sc
		e.posBR.Y = float32(e.photo.CropRectangle.Max.Y) / factor / sc
	} else {
		e.posTL = fyne.NewPos(0, 0)
		e.posBR = fyne.NewPos(e.imgSize.Width, e.imgSize.Height)
	}
	e.sliderTopEdge.Value = float64(e.imgSize.Height)
	e.sliderLeftEdge.Value = 0
	e.sliderBottomEdge.Value = 0
	e.sliderRightEdge.Value = float64(e.imgSize.Width)

	e.sliderTopEdge.OnChanged = func(v float64) {
		if v <= e.sliderBottomEdge.Value {
			return
		}
		e.posTL.Y = e.imgSize.Height - float32(v)
		e.refreshCropBorder()
	}
	e.sliderLeftEdge.OnChanged = func(v float64) {
		if v >= e.sliderRightEdge.Value {
			return
		}
		e.posTL.X = float32(v)
		e.refreshCropBorder()
	}
	e.sliderBottomEdge.OnChanged = func(v float64) {
		if v >= e.sliderTopEdge.Value {
			return
		}
		e.posBR.Y = e.imgSize.Height - float32(v)
		e.refreshCropBorder()
	}
	e.sliderRightEdge.OnChanged = func(v float64) {
		if v <= e.sliderLeftEdge.Value {
			return
		}
		e.posBR.X = float32(v)
		e.refreshCropBorder()
	}
	e.refreshCropSliders()
	e.refreshCropBorder()
}

// Crop the image
func (e *Editor) fixCrop(dstAspect float32) {
	if dstAspect <= e.img.Aspect() {
		e.posTL = fyne.NewPos(0, 0)
		e.posBR = fyne.NewPos(e.imgSize.Height*dstAspect, e.imgSize.Height)
	} else {
		e.posTL = fyne.NewPos(0, e.imgSize.Height-e.imgSize.Width/dstAspect)
		e.posBR = fyne.NewPos(e.imgSize.Width, e.imgSize.Height)

	}
	e.sliderLeftEdge.OnChanged = func(v float64) {
		x := float32(v)
		dx := e.posBR.X - e.posTL.X
		if x >= e.posBR.X || x+dx >= e.imgSize.Width {
			return
		}
		e.posTL.X = x
		e.posBR.X = e.posTL.X + dx
		e.refreshCropBorder()
		e.refreshCropSliders()
	}
	e.sliderBottomEdge.OnChanged = func(v float64) {
		y := e.imgSize.Height - float32(v)
		dy := e.posBR.Y - e.posTL.Y
		if v >= e.sliderTopEdge.Value || dy >= y {
			return
		}
		e.posBR.Y = y
		e.posTL.Y = e.posBR.Y - dy
		e.refreshCropBorder()
		e.refreshCropSliders()
	}
	e.sliderRightEdge.OnChanged = func(v float64) {
		x := float32(v)
		if x <= e.posTL.X || e.posBR.Y <= (x-e.posTL.X)/dstAspect {
			return
		}
		e.posBR.X = x
		e.posTL.Y = e.posBR.Y - (x-e.posTL.X)/dstAspect
		e.refreshCropBorder()
		e.refreshCropSliders()
	}
	e.sliderTopEdge.OnChanged = func(v float64) {
		y := e.imgSize.Height - float32(v)
		if y >= e.posBR.Y || (e.posBR.Y-y)*dstAspect+e.posTL.X >= e.imgSize.Width {
			return
		}
		e.posTL.Y = y
		e.posBR.X = (e.posBR.Y-y)*dstAspect + e.posTL.X
		e.refreshCropBorder()
		e.refreshCropSliders()
	}
	e.refreshCropSliders()
	e.refreshCropBorder()
}

func (e *Editor) ResetAllEdits() {
	e.resetAdjusts()
	e.clearCrop()
}

// Save photo edits
func (e *Editor) saveEdits() {
	p := e.photo
	mi := frame.Items[a.state.ItemPos].Img
	mi.Image = GetListImageAt(p).Image
	// save crops if any
	if e.posTL.X != 0 || e.posTL.Y != 0 || e.posBR.X != e.imgSize.Width || e.posBR.Y != e.imgSize.Height {
		sc := e.w.Canvas().Scale()
		factor := float32(p.height) / float32(e.img.Image.Bounds().Dy()) * sc
		p.CropRectangle = image.Rect(int(e.posTL.X*factor), int(e.posTL.Y*factor), int(e.posBR.X*factor), int(e.posBR.Y*factor))
		p.fadeByCrop(mi)
	} else {
		p.CropRectangle = image.Rect(0, 0, 0, 0)
	}
	// save adjusts if any
	for i, s := range e.adjustFilters.Slider {
		p.Adjust[i] = s.Value
	}
	if p.isAjusted() {
		p.adjustByFilters(mi)
	}
	mi.Refresh()
}

// Workaround for dialog malfunctioning
const joltAmplitude = 1.0

func jolt(w fyne.Window) {
	s := w.Content().Size()
	w.Resize(s.AddWidthHeight(-joltAmplitude, -joltAmplitude))
	time.Sleep(10 * time.Millisecond)
	w.Resize(s)
}
