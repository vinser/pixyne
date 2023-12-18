package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
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

type Shape int

const (
	shapeDefault Shape = iota
	shapeBigger
	shapeSmaller
)

func shapeFame(shape Shape) (cols, size int) {
	if size > len(list) {
		size = len(list)
	}
	switch {
	case shape == shapeBigger:
		switch {
		case frame.Size >= 4:
			cols = MaxFrameColumn
			size = MaxFrameSize
		case frame.Size == 3:
			cols = 2
			size = 4
		default:
			cols = frame.Size + 1
			size = frame.Size + 1
		}
		return
	case shape == shapeSmaller:
		switch {
		case frame.Size >= 5:
			cols = 2
			size = 4
		case frame.Size <= 4:
			cols = frame.Size - 1
			size = frame.Size - 1
		}
		return
	case shape == shapeDefault:
		switch {
		case frame.Size > 4:
			cols = MaxFrameColumn
			size = MaxFrameSize
		case frame.Size == 4:
			cols = 2
			size = 4
		default:
			cols = frame.Size
			size = frame.Size
		}
		return
	}
	return DefaultFrameSize, DefaultFrameSize
}

// Choice tab frame - rows with photos
type Frame struct {
	Content *fyne.Container  `json:"-"`
	ListPos int              `json:"list_pos"`
	Size    int              `json:"frame_size"`
	Cols    int              `json:"-"`
	ItemPos int              `json:"-"`
	Items   []*FrameItem     `json:"-"`
	Simple  bool             `json:"simple_mode"`
	Buttons []*widget.Button `json:"-"`
}

func (a *App) newFrame() {
	frame = &Frame{}
	if len(list) == 0 {
		dialog.ShowInformation("No photos", "There are no JPEG photos in the current folder,\nplease choose another one", a.topWindow)
		frame.Content = container.NewGridWithColumns(1, canvas.NewText("", color.Black))
		return
	}
	frame.ListPos = a.state.FramePos
	if frame.Size = a.state.FrameSize; frame.Size == 0 {
		frame.Size = DefaultFrameSize
	}
	frame.Cols, frame.Size = shapeFame(shapeDefault)
	frame.Simple = a.state.Simple
	frame.Content = container.NewGridWithColumns(frame.Cols)
	frame.Items = make([]*FrameItem, frame.Size)
	frame.At(frame.ListPos)
}

func (f *Frame) ShowProgress() {
	f.DisableButtons()
	loadProgress.Show()
}

func (f *Frame) HideProgress() {
	loadProgress.Hide()
	f.EnableButtons()
}
func (f *Frame) At(pos int) {
	if pos < 0 || pos >= len(list) {
		return
	}
	f.Content.RemoveAll()
	for i := 0; i < f.Size; i++ {
		f.Items[i] = NewFrameItem(pos+i, f.Simple)
	}
	for i := 0; i < f.Size; i++ {
		f.Content.Add(f.Items[i].Content)
	}
	f.Content.Refresh()
	f.ListPos = pos
}

func (f *Frame) First() {
	f.ShowProgress()
	defer f.HideProgress()
	f.At(0)
	f.updateFrameScrollButtons()
}

func (f *Frame) Last() {
	f.ShowProgress()
	defer f.HideProgress()
	pos := len(list) - f.Size
	f.At(pos)
	f.updateFrameScrollButtons()
}

func (f *Frame) Prev() {
	f.ShowProgress()
	defer f.HideProgress()
	pos := f.ListPos - f.Size
	if pos < 0 {
		f.First()
		return
	}
	f.At(pos)
	f.updateFrameScrollButtons()
}

func (f *Frame) Next() {
	f.ShowProgress()
	defer f.HideProgress()
	pos := f.ListPos + f.Size
	if pos > len(list)-f.Size {
		f.Last()
		return
	}
	f.At(pos)
	f.updateFrameScrollButtons()
}

func (f *Frame) PrevItem() {
	f.ShowProgress()
	defer f.HideProgress()
	if f.ListPos > 0 {
		f.Items = f.Items[:len(f.Items)-1]
		f.Items = append([]*FrameItem{NewFrameItem(f.ListPos-1, f.Simple)}, f.Items...)
		f.Content.RemoveAll()
		for i := 0; i < f.Size; i++ {
			f.Content.Add(f.Items[i].Content)
		}
		f.Content.Refresh()
		f.ListPos--
		f.updateFrameScrollButtons()
	}
}

func (f *Frame) NextItem() {
	f.ShowProgress()
	defer f.HideProgress()
	if f.ListPos < len(list)-f.Size {
		f.Items = f.Items[1:]
		f.Items = append(f.Items, NewFrameItem(f.ListPos+f.Size, f.Simple))
		f.Content.RemoveAll()
		for i := 0; i < f.Size; i++ {
			f.Content.Add(f.Items[i].Content)
		}
		f.Content.Refresh()
		f.ListPos++
		f.updateFrameScrollButtons()
	}
}

func (f *Frame) RemoveItem() {
	if f.Size > MinFrameSize {
		f.ShowProgress()
		defer f.HideProgress()
		newCols, newSize := shapeFame(shapeSmaller)
		f.Items = f.Items[:len(f.Items)-f.Size+newSize]
		f.Size = newSize
		f.Cols = newCols
		f.Content.RemoveAll()
		f.Content.Layout = layout.NewGridLayoutWithColumns(newCols)
		for i := 0; i < f.Size; i++ {
			f.Content.Add(f.Items[i].Content)
		}
		f.Content.Refresh()
		f.updateFrameScrollButtons()
		a.showFrameToolbar()
	}
}

func (f *Frame) AddItem() {
	if f.Size < MaxFrameSize {
		f.ShowProgress()
		defer f.HideProgress()
		newCols, newSize := shapeFame(shapeBigger)
		if f.ListPos+f.Size >= len(list) {
			f.At(len(list) - newSize)
		}
		for i := 0; i < newSize-f.Size; i++ {
			f.Items = append(f.Items, NewFrameItem(f.ListPos+f.Size+i, f.Simple))
		}
		f.Size = newSize
		f.Cols = newCols
		f.Content.RemoveAll()
		f.Content.Layout = layout.NewGridLayoutWithColumns(newCols)
		for i := 0; i < f.Size; i++ {
			f.Content.Add(f.Items[i].Content)
		}
		f.Content.Refresh()
		f.updateFrameScrollButtons()
		a.showFrameToolbar()
	}
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

func (f *Frame) newFrameView() *fyne.Container {
	opts := map[int]scrollButtonOpts{
		firstPhotoBtn: {label: "|<", icon: theme.MediaSkipPreviousIcon(), tapped: func() { frame.First() }},
		prevFrameBtn:  {label: "<<", icon: theme.MediaFastRewindIcon(), tapped: func() { frame.Prev() }},
		prevPhotoBtn:  {label: "<", icon: theme.NewThemedResource(iconScrollBack), tapped: func() { frame.PrevItem() }},
		nextPhotoBtn:  {label: ">", icon: theme.MediaPlayIcon(), tapped: func() { frame.NextItem() }},
		nextFrameBtn:  {label: ">>", icon: theme.MediaFastForwardIcon(), tapped: func() { frame.Next() }},
		lastPhotoBtn:  {label: ">|", icon: theme.MediaSkipNextIcon(), tapped: func() { frame.Last() }},
	}
	objs := make([]fyne.CanvasObject, len(opts))
	f.Buttons = make([]*widget.Button, len(opts))
	for i, opt := range opts {
		btn := widget.NewButtonWithIcon("", opt.icon, opt.tapped)
		btn.Importance = widget.HighImportance
		objs[i] = btn
		f.Buttons[i] = btn
	}

	btns := container.NewGridWithColumns(len(objs), objs...)
	f.updateFrameScrollButtons()
	return container.NewBorder(nil, btns, nil, nil, f.Content)
}

func (f *Frame) DisableButtons() {
	for i := 0; i < len(f.Buttons); i++ {
		f.Buttons[i].Disable()
	}
}

func (f *Frame) EnableButtons() {
	f.updateFrameScrollButtons()
}

func (f *Frame) updateFrameScrollButtons() {
	f.Buttons[prevPhotoBtn].Enable()
	f.Buttons[prevFrameBtn].Enable()
	f.Buttons[firstPhotoBtn].Enable()
	f.Buttons[nextPhotoBtn].Enable()
	f.Buttons[nextFrameBtn].Enable()
	f.Buttons[lastPhotoBtn].Enable()
	if frame.ListPos == 0 {
		f.Buttons[prevPhotoBtn].Disable()
		f.Buttons[prevFrameBtn].Disable()
		f.Buttons[firstPhotoBtn].Disable()
	}
	if frame.ListPos+frame.Size == len(list) {
		f.Buttons[nextPhotoBtn].Disable()
		f.Buttons[nextFrameBtn].Disable()
		f.Buttons[lastPhotoBtn].Disable()
	}
}

type FrameItem struct {
	Content *fyne.Container
	Label   string
	Img     *canvas.Image
	Button  *widget.Button
}

// NewFrameItem creates a new FrameItem
func NewFrameItem(listPos int, simpleMode bool) *FrameItem {
	item := &FrameItem{}
	item.Img = GetListImageAt(listPos)
	item.Label = fmt.Sprint(list[listPos].fileURI.Name())
	item.Button = widget.NewButton("", func() { toggleDrop(listPos, item) })
	if list[listPos].Drop {
		item.Button.SetText("DROPPED")
		item.Img.Translucency = 0.8
	}
	topLabel := widget.NewLabelWithStyle(item.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	topLabel.Truncation = fyne.TextTruncateEllipsis
	centerStack := container.NewStack()
	centerStack.Add(item.Img)
	centerStack.Add(item.Button)
	if simpleMode {
		item.Content = container.NewBorder(topLabel, nil, nil, nil, centerStack)
	} else {
		item.Content = container.NewBorder(topLabel, newDateInput(listPos), nil, nil, centerStack)
	}
	return item
}
func newDateInput(listPos int) *fyne.Container {
	p := list[listPos]
	choices := []string{"EXIF", "File", "Input"}
	d := listDateToDisplayDate(p.Dates[p.DateUsed])

	eDate := widget.NewEntry()
	eDate.Validator = validation.NewTime(DisplayDateFormat)
	eDate.SetText(d)
	eDate.OnChanged = func(e string) {
		p.Dates[p.DateUsed] = displayDateToListDate(e)
	}

	rgDateChoice := widget.NewRadioGroup(
		choices,
		func(s string) {
			switch s {
			case "EXIF":
				p.Dates[UseEnteredDate] = ""
				p.DateUsed = UseExifDate
				eDate.SetText(listDateToDisplayDate(p.Dates[p.DateUsed]))
				eDate.Disable()
			case "File":
				p.Dates[UseEnteredDate] = ""
				p.DateUsed = UseFileDate
				eDate.SetText(listDateToDisplayDate(p.Dates[p.DateUsed]))
				eDate.Disable()
			case "Input":
				p.DateUsed = UseEnteredDate
				if p.Dates[p.DateUsed] == "" {
					if p.Dates[UseExifDate] == "" {
						p.Dates[p.DateUsed] = p.Dates[UseFileDate]
					} else {
						p.Dates[p.DateUsed] = p.Dates[UseExifDate]
					}
				}
				eDate.SetText(listDateToDisplayDate(p.Dates[p.DateUsed]))
				eDate.Enable()
			}
		})
	rgDateChoice.SetSelected(choices[p.DateUsed])
	rgDateChoice.Horizontal = true

	gr := container.NewVBox(rgDateChoice, eDate)

	return container.NewCenter(gr)
}

func toggleDrop(pos int, item *FrameItem) {
	list[pos].Drop = !list[pos].Drop
	if list[pos].Drop {
		item.Button.SetText("DROPPED")
		item.Img.Translucency = 0.8
	} else {
		item.Button.SetText("")
		item.Img.Translucency = 0
	}
}
