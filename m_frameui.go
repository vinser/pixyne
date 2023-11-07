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
	DefaultListPos   = 0
	DefaultFrameSize = 3
	MinFrameSize     = 1
	MaxFrameSize     = 6
	MaxFrameColumn   = 3
)

// Choice tab frame - rows with photos
type Frame struct {
	Container *fyne.Container `json:"-"`
	Pos       int             `json:"pos"`
	Size      int             `json:"size"`
}

// fill frame with photo images starting from pos = 0.
func (a *App) initFrame() {
	frame = &Frame{}
	if len(list) == 0 {
		dialog.ShowInformation("No photos", "There are no JPEG photos in the current folder,\nplease choose another one", a.topWindow)
		frame.Container = container.NewGridWithColumns(1, canvas.NewText("", color.Black))
		return
	}
	frame.Pos = a.state.FramePos
	if frame.Size = a.state.FrameSize; frame.Size == 0 {
		frame.Size = DefaultFrameSize
	}
	if frame.Size > len(list) {
		frame.Size = len(list)
	}
	for i := frame.Pos; i < frame.Pos+frame.Size; i++ {
		list[i].SetImage(frame.Size)
	}
	frame.Container = container.NewGridWithColumns(getFrameColumnNum(frame.Size))
	for i := 0; i < frame.Size && i < len(list); i++ {
		frame.Container.Add(list[frame.Pos+i].NewFrameColumn(a.simpleMode))
	}
}

func getFrameColumnNum(frameSize int) int {
	switch {
	case frame.Size > 4:
		return MaxFrameColumn
	case frame.Size == 4:
		return 2
	case frame.Size < 1:
		return 1
	default:
		return frame.Size
	}
}

// scrollFrame frame at position pos
func (a *App) scrollFrame(newPos int) {
	switch {
	case newPos < 0:
		newPos = 0
	case newPos > len(list)-frame.Size:
		newPos = len(list) - frame.Size
	}

	switch {
	// case newPos == frame.Pos:
	// 	return
	case newPos-frame.Pos >= frame.Size || frame.Pos-newPos >= frame.Size || newPos == frame.Pos:
		for i := 0; i < frame.Size; i++ {
			list[frame.Pos+i].Img = nil
			list[newPos+i].SetImage(frame.Size)
		}
	case newPos > frame.Pos:
		for i := 0; i < newPos-frame.Pos; i++ {
			list[frame.Pos+i].Img = nil
			list[frame.Pos+frame.Size+i].SetImage(frame.Size)
		}
	case newPos < frame.Pos:
		for i := 0; i < frame.Pos-newPos; i++ {
			list[newPos+frame.Size+i].Img = nil
			list[newPos+i].SetImage(frame.Size)
		}
	}
	frame.Container.RemoveAll()
	for i := 0; i < frame.Size; i++ {
		frame.Container.Add(list[newPos+i].NewFrameColumn(a.simpleMode))
	}
	frame.Container.Refresh()
	frame.Pos = newPos
	a.updateFrameScrollButtons()
}

const (
	MorePhoto = 1
	LessPhoto = -1
)

// resizeFrame frame
func (a *App) resizeFrame(zoom int) {
	switch zoom {
	case LessPhoto:
		switch {
		case frame.Size-1 < MinFrameSize:
			return
		case frame.Size == 6: // skip 5 photos frame layout
			zoom--
		}
		for i := zoom; i < 0; i++ {
			list[frame.Pos+frame.Size+i].Img = nil
		}
		frame.Size += zoom
	case MorePhoto:
		switch {
		case frame.Size == MaxFrameSize || frame.Size == len(list):
			return
		case frame.Size == 4 && len(list) > 5: // skip 5 photos frame layout
			zoom++
		}
		if frame.Pos+frame.Size+zoom > len(list) {
			frame.Pos = frame.Pos - zoom
			for i := 0; i < zoom; i++ {
				list[frame.Pos+i].SetImage(frame.Size)
			}
		} else {
			for i := 0; i < zoom; i++ {
				list[frame.Pos+frame.Size+i].SetImage(frame.Size)
			}
		}
		frame.Size += zoom
	}
	frame.Container.RemoveAll()
	for i := 0; i < frame.Size; i++ {
		frame.Container.Add(list[frame.Pos+i].NewFrameColumn(a.simpleMode))
	}
	frame.Container.Layout = layout.NewGridLayoutWithColumns(getFrameColumnNum(len(frame.Container.Objects)))
	frame.Container.Refresh()
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
