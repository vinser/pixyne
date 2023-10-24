package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	InitListPos    = 0
	InitFrameSize  = 3
	MinFrameSize   = 1
	MaxFrameSize   = 6
	MaxFrameColumn = 3
)

// Choice tab frame - row with photos
type Frame struct {
	Container *fyne.Container `json:"-"`
	Pos       int             `json:"pos"`
	Size      int             `json:"size"`
}

// fill frame with photo images starting from pos = 0.
func (a *App) initFrame() {
	frame = &Frame{}
	f := frame
	if len(list) == 0 {
		dialog.ShowInformation("No photos", "There are no JPEG photos in the current folder,\nplease choose another one", a.topWindow)
		f.Container = container.NewGridWithColumns(1, canvas.NewText("", color.Black))
		return
	}
	f.Pos = a.state.Pos
	if f.Size = a.state.Size; f.Size == 0 {
		f.Size = InitFrameSize
	}
	if f.Size > len(list) {
		f.Size = len(list)
	}
	for i := f.Pos; i < f.Pos+f.Size; i++ {
		list[i].Img = list[i].GetImage(f.Size)
	}
	f.Container = container.NewGridWithColumns(frameColumnNum(f.Size))
	for i := 0; i < f.Size && i < len(list); i++ {
		f.Container.Add(list[f.Pos+i].NewFrameColumn())
	}
}

// frameColumnNum calculates the number of columns for the frame.
func frameColumnNum(frameSize int) int {
	switch {
	case frameSize > 4:
		return MaxFrameColumn
	case frameSize == 4:
		return 2
	case frameSize < 1:
		return 1
	default:
		return frameSize
	}
}

// scrollFrame frame at position pos
func (a *App) scrollFrame(pos int) {
	f := frame
	switch {
	case pos < 0:
		pos = 0
	case pos > len(list)-f.Size:
		pos = len(list) - f.Size
	}

	switch {
	case pos-f.Pos >= f.Size || f.Pos-pos >= f.Size || pos == f.Pos:
		for i := f.Pos; i < f.Pos+f.Size; i++ {
			list[i].Img = nil
		}
		for i := pos; i < pos+f.Size; i++ {
			list[i].Img = list[i].GetImage(f.Size)
			if list[i].Dropped {
				list[i].Img.Translucency = 0.5
			}
		}
	case pos > f.Pos:
		for i := f.Pos; i < pos; i++ {
			list[i].Img = nil
			list[i+f.Size].Img = list[i+f.Size].GetImage(f.Size)
			if list[i+f.Size].Dropped {
				list[i+f.Size].Img.Translucency = 0.5
			}
		}
	case f.Pos > pos:
		for i := pos; i < f.Pos; i++ {
			list[i+f.Size].Img = nil
			list[i].Img = list[i].GetImage(f.Size)
			if list[i].Dropped {
				list[i].Img.Translucency = 0.5
			}
		}
	}

	// TODO: may be optimized when for scroll les than frame size by not all objects deletion/addition? Somwthing like this:
	// https://stackoverflow.com/questions/63995289/how-to-remove-objects-from-golang-fyne-container
	f.Container.RemoveAll()
	for i := 0; i < f.Size; i++ {
		f.Container.Add(list[pos+i].NewFrameColumn())
	}
	f.Container.Refresh()
	f.Pos = pos
	a.updateFrameScrollButtons()
}

const (
	AddColumn = iota
	RemoveColumn
)

// resizeFrame frame
func (a *App) resizeFrame(zoom int) {
	f := frame
	switch zoom {
	case RemoveColumn:
		if f.Size-1 < MinFrameSize {
			return
		}
		list[f.Pos+f.Size-1].Img = nil
		f.Size--
	case AddColumn:
		if f.Size+1 > MaxFrameSize || f.Size+1 > len(list) {
			return
		}
		i := f.Pos + f.Size
		if i == len(list) {
			f.Pos--
			i = f.Pos
		}
		list[i].Img = list[i].GetImage(f.Size)
		if list[i].Dropped {
			list[i].Img.Translucency = 0.5
		}
		f.Size++
	}
	//      0-1-2-3-4-5-6-7-8
	//          2-3-4			p=2, s=3
	// 		0-1-2				p=0, s=3
	// 					6-7-8	p=6, s=3

	// TODO: may be optimized when for scroll les than frame size by not all objects deletion/addition? Somwthing like this:
	// https://stackoverflow.com/questions/63995289/how-to-remove-objects-from-golang-fyne-container
	f.Container.RemoveAll()
	for i := 0; i < f.Size; i++ {
		f.Container.Add(list[f.Pos+i].NewFrameColumn())
	}
	f.Container.Layout = layout.NewGridLayoutWithColumns(frameColumnNum(len(f.Container.Objects)))
	f.Container.Refresh()
	a.showFrameToolbar()
	a.updateFrameScrollButtons()
}

// Fill new frame with photo images starting from frame.Pos
func (a *App) fillFrame() {
	frame.Container = container.NewGridWithColumns(frameColumnNum(frame.Size))
	for i := 0; i < frame.Size; i++ {
		frame.Container.Add(list[frame.Pos+i].NewFrameColumn())
	}
}

// Append new n photos to frame
func (a *App) appendFrame(n int) {
	frameObjects := frame.Container.Objects
	frame.Container = container.NewGridWithColumns(frameColumnNum(frame.Size), frameObjects...)
	for i := 0; i < n; i++ {
		frame.Container.Add(list[frame.Pos+frame.Size+i].NewFrameColumn())
	}
}

// Trim n photos from frame tail
func (a *App) trimFrameTail(n int) {
	frameObjects := frame.Container.Objects[0 : frame.Size-n-1]
	frame.Container = container.NewGridWithColumns(frameColumnNum(frame.Size), frameObjects...)
}

// Scroll frame

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
		prevFrameBtn:  {label: "<<", icon: theme.MediaFastRewindIcon(), tapped: func() { a.scrollFrame(frame.Pos - frame.Size) }},
		prevPhotoBtn:  {label: "<", icon: theme.NewThemedResource(iconScrollBack), tapped: func() { a.scrollFrame(frame.Pos - 1) }},
		nextPhotoBtn:  {label: ">", icon: theme.MediaPlayIcon(), tapped: func() { a.scrollFrame(frame.Pos + 1) }},
		nextFrameBtn:  {label: ">>", icon: theme.MediaFastForwardIcon(), tapped: func() { a.scrollFrame(frame.Pos + frame.Size) }},
		lastPhotoBtn:  {label: ">|", icon: theme.MediaSkipNextIcon(), tapped: func() { a.scrollFrame(len(list)) }},
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
	// a.bottomButtons = container.NewCenter(container.NewHBox(o...))
	a.frameView = container.NewBorder(nil, a.bottomButtons, nil, nil, frame.Container)
	a.updateFrameScrollButtons()
}

func (a *App) updateFrameScrollButtons() {
	a.scrollButton[prevPhotoBtn].Enable()
	a.scrollButton[prevFrameBtn].Enable()
	a.scrollButton[firstPhotoBtn].Enable()
	a.scrollButton[nextPhotoBtn].Enable()
	a.scrollButton[nextFrameBtn].Enable()
	a.scrollButton[lastPhotoBtn].Enable()
	if frame.Pos == 0 {
		a.scrollButton[prevPhotoBtn].Disable()
		a.scrollButton[prevFrameBtn].Disable()
		a.scrollButton[firstPhotoBtn].Disable()
	}
	if frame.Pos+frame.Size == len(list) {
		a.scrollButton[nextPhotoBtn].Disable()
		a.scrollButton[nextFrameBtn].Disable()
		a.scrollButton[lastPhotoBtn].Disable()
	}
}
