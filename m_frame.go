package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Choice tab frame - row with photos
type Frame struct {
	*fyne.Container
	pos  int
	size int
}

// fill frame with photo images starting from pos = 0.
func (a *App) initFrame() {
	a.frame = &Frame{
		pos:  InitListPos,
		size: InitFrameSize,
	}
	f := a.frame
	if f.size > len(a.List) {
		f.size = len(a.List)
	}
	if f.size == 0 { // Workaround for fyne v.2.3.1 NewGridWithColumns(0) main window shrink on Windows OS  !!! Remove this when issue #3669 will be closed
		f.Container = container.NewGridWithColumns(1, canvas.NewText("", color.Black))
		return
	}
	for i := f.pos; i < f.pos+f.size; i++ {
		a.List[i].Img = a.List[i].GetImage(f.size)
	}
	f.Container = container.NewGridWithColumns(f.size)
	for i := 0; i < f.size && i < len(a.List); i++ {
		f.Add(a.List[f.pos+i].FrameColumn())
	}
}

// scrollFrame frame at position pos
func (a *App) scrollFrame(pos int) {
	f := a.frame
	switch {
	case pos < 0:
		pos = 0
	case pos > len(a.List)-f.size:
		pos = len(a.List) - f.size
	}

	switch {
	case pos-f.pos >= f.size || f.pos-pos >= f.size:
		for i := f.pos; i < f.pos+f.size; i++ {
			a.List[i].Img = nil
		}
		for i := pos; i < pos+f.size; i++ {
			a.List[i].Img = a.List[i].GetImage(f.size)
			if a.List[i].Droped {
				a.List[i].Img.Translucency = 0.5
			}
		}
	case pos > f.pos:
		for i := f.pos; i < pos; i++ {
			a.List[i].Img = nil
			a.List[i+f.size].Img = a.List[i+f.size].GetImage(f.size)
			if a.List[i+f.size].Droped {
				a.List[i+f.size].Img.Translucency = 0.5
			}
		}
	case f.pos > pos:
		for i := pos; i < f.pos; i++ {
			a.List[i+f.size].Img = nil
			a.List[i].Img = a.List[i].GetImage(f.size)
			if a.List[i].Droped {
				a.List[i].Img.Translucency = 0.5
			}
		}
	}

	// TODO: may be optimized when for scroll les than frame size by not all objects deletion/addition? Somwthing like this:
	// https://stackoverflow.com/questions/63995289/how-to-remove-objects-from-golang-fyne-container
	f.RemoveAll()
	for i := 0; i < f.size; i++ {
		f.Add(a.List[pos+i].FrameColumn())
	}
	f.Refresh()
	f.pos = pos
	a.updateFrameScrollButtons()
}

// resizeFrame frame
func (a *App) resizeFrame(zoom int) {
	f := a.frame
	switch zoom {
	case RemoveColumn:
		if f.size-1 < MinFrameSize {
			return
		}
		a.List[f.pos+f.size-1].Img = nil
		f.size--
	case AddColumn:
		if f.size+1 > MaxFrameSize || f.size+1 > len(a.List) {
			return
		}
		i := f.pos + f.size
		if i == len(a.List) {
			f.pos--
			i = f.pos
		}
		a.List[i].Img = a.List[i].GetImage(f.size)
		if a.List[i].Droped {
			a.List[i].Img.Translucency = 0.5
		}
		f.size++
	}
	//      0-1-2-3-4-5-6-7-8
	//          2-3-4			p=2, s=3
	// 		0-1-2				p=0, s=3
	// 					6-7-8	p=6, s=3

	// TODO: may be optimized when for scroll les than frame size by not all objects deletion/addition? Somwthing like this:
	// https://stackoverflow.com/questions/63995289/how-to-remove-objects-from-golang-fyne-container
	f.RemoveAll()
	for i := 0; i < f.size; i++ {
		f.Add(a.List[f.pos+i].FrameColumn())
	}
	f.Layout = layout.NewGridLayoutWithColumns(len(f.Objects))
	f.Refresh()
	a.showFrameToolbar()
	a.updateFrameScrollButtons()
}

// Frame scroll button names
const (
	firstPhotoBtn = iota
	prevFrameBtn
	prevPhotoBtn
	nextPhotoBtn
	nextFrameBtn
	lastPhotoBtn
)

type scrollButtonOpts struct {
	label  string
	icon   fyne.Resource
	tapped func()
}

func (a *App) newFrameView() {
	sbo := map[int]scrollButtonOpts{
		firstPhotoBtn: {label: "|<", icon: theme.MediaSkipPreviousIcon(), tapped: func() { a.scrollFrame(0) }},
		prevFrameBtn:  {label: "<<", icon: theme.MediaFastRewindIcon(), tapped: func() { a.scrollFrame(a.frame.pos - a.frame.size) }},
		prevPhotoBtn:  {label: "<", icon: theme.NewThemedResource(iconScrollBack), tapped: func() { a.scrollFrame(a.frame.pos - 1) }},
		nextPhotoBtn:  {label: ">", icon: theme.MediaPlayIcon(), tapped: func() { a.scrollFrame(a.frame.pos + 1) }},
		nextFrameBtn:  {label: ">>", icon: theme.MediaFastForwardIcon(), tapped: func() { a.scrollFrame(a.frame.pos + a.frame.size) }},
		lastPhotoBtn:  {label: ">|", icon: theme.MediaSkipNextIcon(), tapped: func() { a.scrollFrame(len(a.List)) }},
	}
	o := make([]fyne.CanvasObject, len(sbo))
	a.scrollButton = make([]*widget.Button, len(sbo))
	for i, opt := range sbo {
		b := widget.NewButtonWithIcon("", opt.icon, opt.tapped)
		b.Importance = widget.HighImportance
		o[i] = b
		a.scrollButton[i] = b
	}

	a.bottomButtons = container.NewGridWithColumns(len(o), o...)
	a.frameView = container.NewBorder(nil, a.bottomButtons, nil, nil, a.frame.Container)
	a.updateFrameScrollButtons()
}

func (a *App) updateFrameScrollButtons() {
	a.scrollButton[prevPhotoBtn].Enable()
	a.scrollButton[prevFrameBtn].Enable()
	a.scrollButton[firstPhotoBtn].Enable()
	a.scrollButton[nextPhotoBtn].Enable()
	a.scrollButton[nextFrameBtn].Enable()
	a.scrollButton[lastPhotoBtn].Enable()
	if a.frame.pos == 0 {
		a.scrollButton[prevPhotoBtn].Disable()
		a.scrollButton[prevFrameBtn].Disable()
		a.scrollButton[firstPhotoBtn].Disable()
	}
	if a.frame.pos+a.frame.size == len(a.List) {
		a.scrollButton[nextPhotoBtn].Disable()
		a.scrollButton[nextFrameBtn].Disable()
		a.scrollButton[lastPhotoBtn].Disable()
	}
}
