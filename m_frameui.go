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
	*fyne.Container `json:"-"`
	Pos             int `json:"pos"`
	Size            int `json:"size"`
}

// fill frame with photo images starting from pos = 0.
func (a *App) initFrame() {
	a.frame = &Frame{}
	f := a.frame
	if len(a.List) == 0 {
		dialog.ShowInformation("No photos", "There are no JPEG photos in the current folder,\nplease choose another one", a.topWindow)
		f.Container = container.NewGridWithColumns(1, canvas.NewText("", color.Black))
		return
	}
	f.Pos = a.state.Pos
	if f.Size = a.state.Size; f.Size == 0 {
		f.Size = InitFrameSize
	}
	if f.Size > len(a.List) {
		f.Size = len(a.List)
	}
	for i := f.Pos; i < f.Pos+f.Size; i++ {
		a.List[i].Img = a.List[i].GetImage(f.Size)
	}
	f.Container = container.NewGridWithColumns(frameColumnNum(f.Size))
	for i := 0; i < f.Size && i < len(a.List); i++ {
		f.Add(a.List[f.Pos+i].FrameColumn())
	}
}

// frameColumnNum calculates the number of columns for the frame.
func frameColumnNum(frameSize int) int {
	switch frameSize {
	case 0:
		return 1
	case 4:
		return 2
	case 5, 6:
		return MaxFrameColumn
	default:
		return frameSize
	}
}

// scrollFrame frame at position pos
func (a *App) scrollFrame(pos int) {
	f := a.frame
	switch {
	case pos < 0:
		pos = 0
	case pos > len(a.List)-f.Size:
		pos = len(a.List) - f.Size
	}

	switch {
	case pos-f.Pos >= f.Size || f.Pos-pos >= f.Size:
		for i := f.Pos; i < f.Pos+f.Size; i++ {
			a.List[i].Img = nil
		}
		for i := pos; i < pos+f.Size; i++ {
			a.List[i].Img = a.List[i].GetImage(f.Size)
			if a.List[i].Dropped {
				a.List[i].Img.Translucency = 0.5
			}
		}
	case pos > f.Pos:
		for i := f.Pos; i < pos; i++ {
			a.List[i].Img = nil
			a.List[i+f.Size].Img = a.List[i+f.Size].GetImage(f.Size)
			if a.List[i+f.Size].Dropped {
				a.List[i+f.Size].Img.Translucency = 0.5
			}
		}
	case f.Pos > pos:
		for i := pos; i < f.Pos; i++ {
			a.List[i+f.Size].Img = nil
			a.List[i].Img = a.List[i].GetImage(f.Size)
			if a.List[i].Dropped {
				a.List[i].Img.Translucency = 0.5
			}
		}
	}

	// TODO: may be optimized when for scroll les than frame size by not all objects deletion/addition? Somwthing like this:
	// https://stackoverflow.com/questions/63995289/how-to-remove-objects-from-golang-fyne-container
	f.RemoveAll()
	for i := 0; i < f.Size; i++ {
		f.Add(a.List[pos+i].FrameColumn())
	}
	f.Refresh()
	f.Pos = pos
	a.updateFrameScrollButtons()
}

const (
	AddColumn = iota
	RemoveColumn
)

// resizeFrame frame
func (a *App) resizeFrame(zoom int) {
	f := a.frame
	switch zoom {
	case RemoveColumn:
		if f.Size-1 < MinFrameSize {
			return
		}
		a.List[f.Pos+f.Size-1].Img = nil
		f.Size--
	case AddColumn:
		if f.Size+1 > MaxFrameSize || f.Size+1 > len(a.List) {
			return
		}
		i := f.Pos + f.Size
		if i == len(a.List) {
			f.Pos--
			i = f.Pos
		}
		a.List[i].Img = a.List[i].GetImage(f.Size)
		if a.List[i].Dropped {
			a.List[i].Img.Translucency = 0.5
		}
		f.Size++
	}
	//      0-1-2-3-4-5-6-7-8
	//          2-3-4			p=2, s=3
	// 		0-1-2				p=0, s=3
	// 					6-7-8	p=6, s=3

	// TODO: may be optimized when for scroll les than frame size by not all objects deletion/addition? Somwthing like this:
	// https://stackoverflow.com/questions/63995289/how-to-remove-objects-from-golang-fyne-container
	f.RemoveAll()
	for i := 0; i < f.Size; i++ {
		f.Add(a.List[f.Pos+i].FrameColumn())
	}
	f.Layout = layout.NewGridLayoutWithColumns(frameColumnNum(len(f.Objects)))
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
		prevFrameBtn:  {label: "<<", icon: theme.MediaFastRewindIcon(), tapped: func() { a.scrollFrame(a.frame.Pos - a.frame.Size) }},
		prevPhotoBtn:  {label: "<", icon: theme.NewThemedResource(iconScrollBack), tapped: func() { a.scrollFrame(a.frame.Pos - 1) }},
		nextPhotoBtn:  {label: ">", icon: theme.MediaPlayIcon(), tapped: func() { a.scrollFrame(a.frame.Pos + 1) }},
		nextFrameBtn:  {label: ">>", icon: theme.MediaFastForwardIcon(), tapped: func() { a.scrollFrame(a.frame.Pos + a.frame.Size) }},
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
	if a.frame.Pos == 0 {
		a.scrollButton[prevPhotoBtn].Disable()
		a.scrollButton[prevFrameBtn].Disable()
		a.scrollButton[firstPhotoBtn].Disable()
	}
	if a.frame.Pos+a.frame.Size == len(a.List) {
		a.scrollButton[nextPhotoBtn].Disable()
		a.scrollButton[nextFrameBtn].Disable()
		a.scrollButton[lastPhotoBtn].Disable()
	}
}
