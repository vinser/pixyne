package main

import (
	"fmt"
	"image/color"
	"runtime/debug"
	"sync"

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
	DefaultItemPos   = 0
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
		case a.state.FrameSize >= 4:
			cols = MaxFrameColumn
			size = MaxFrameSize
		case a.state.FrameSize == 3:
			cols = 2
			size = 4
		default:
			cols = a.state.FrameSize + 1
			size = a.state.FrameSize + 1
		}
		return
	case shape == shapeSmaller:
		switch {
		case a.state.FrameSize >= 5:
			cols = 2
			size = 4
		case a.state.FrameSize <= 4:
			cols = a.state.FrameSize - 1
			size = a.state.FrameSize - 1
		}
		return
	case shape == shapeDefault:
		switch {
		case a.state.FrameSize > 4:
			cols = MaxFrameColumn
			size = MaxFrameSize
		case a.state.FrameSize == 4:
			cols = 2
			size = 4
		default:
			cols = a.state.FrameSize
			size = a.state.FrameSize
		}
		return
	}
	return DefaultFrameSize, DefaultFrameSize
}

// Choice tab frame - rows with photos
type Frame struct {
	Content *fyne.Container
	Cols    int
	// ItemPos    int
	Items   []*FrameItem
	Buttons []*widget.Button
}

func (a *App) newFrame() {
	frame = &Frame{}
	if len(list) == 0 {
		dialog.ShowInformation("No photos", "There are no JPEG photos in the current folder,\nplease choose another one", a.topWindow)
		frame.Content = container.NewGridWithColumns(1, canvas.NewText("", color.Black))
		return
	}
	if a.state.FrameSize == 0 {
		a.state.FrameSize = DefaultFrameSize
	}
	frame.Cols, a.state.FrameSize = shapeFame(shapeDefault)
	frame.Content = container.NewGridWithColumns(frame.Cols)
	frame.Items = make([]*FrameItem, a.state.FrameSize)
	frame.At(a.state.FramePos)
	frame.ItemEndingAt(a.state.ItemPos)
}

func (f *Frame) At(pos int) {
	if pos < 0 || pos >= len(list) {
		return
	}
	f.Content.RemoveAll()
	var wg sync.WaitGroup
	for i := 0; i < a.state.FrameSize; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			f.Items[i] = NewFrameItem(pos+i, a.state.Simple)
		}(i)
	}
	wg.Wait()
	for i := 0; i < a.state.FrameSize; i++ {
		f.Content.Add(f.Items[i].Content)
	}
	f.Content.Refresh()
	a.state.FramePos = pos
}

func (f *Frame) First() {
	if a.state.FramePos == 0 {
		return
	}
	a.statusInfo.ShowProgress()
	defer a.statusInfo.HideProgress()
	f.At(0)
	f.ItemEndingAt(0)
	f.updateFrameScrollButtons()
}

func (f *Frame) Last() {
	if a.state.FramePos == len(list)-a.state.FrameSize {
		return
	}
	pos := len(list) - a.state.FrameSize
	a.statusInfo.ShowProgress()
	defer a.statusInfo.HideProgress()
	f.At(pos)
	f.ItemEndingAt(a.state.FrameSize - 1)
	f.updateFrameScrollButtons()
}

func (f *Frame) Prev() {
	if a.state.FramePos == 0 {
		return
	}
	a.statusInfo.ShowProgress()
	defer a.statusInfo.HideProgress()
	pos := a.state.FramePos - a.state.FrameSize
	if pos < 0 {
		f.First()
		return
	}
	f.At(pos)
	f.ItemEndingAt(0)
	f.updateFrameScrollButtons()
}

func (f *Frame) Next() {
	if a.state.FramePos == len(list)-a.state.FrameSize {
		return
	}
	a.statusInfo.ShowProgress()
	defer a.statusInfo.HideProgress()
	pos := a.state.FramePos + a.state.FrameSize
	if pos > len(list)-a.state.FrameSize {
		f.Last()
		return
	}
	f.At(pos)
	f.ItemEndingAt(a.state.FrameSize - 1)
	f.updateFrameScrollButtons()
}

func (f *Frame) PrevItem() {
	a.statusInfo.ShowProgress()
	defer a.statusInfo.HideProgress()
	if a.state.ItemPos > 0 {
		f.ItemEndingAt(a.state.ItemPos - 1)
		f.Content.Refresh()
		return
	}
	if a.state.FramePos > 0 {
		f.Items = f.Items[:len(f.Items)-1]
		f.Items = append([]*FrameItem{NewFrameItem(a.state.FramePos-1, a.state.Simple)}, f.Items...)
		f.Content.RemoveAll()
		for i := 0; i < a.state.FrameSize; i++ {
			f.Content.Add(f.Items[i].Content)
		}
		f.ItemEndingAt(0)
		f.Content.Refresh()
		a.state.FramePos--
		f.updateFrameScrollButtons()
	}
}

func (f *Frame) NextItem() {
	a.statusInfo.ShowProgress()
	defer a.statusInfo.HideProgress()
	if a.state.ItemPos < a.state.FrameSize-1 {
		f.ItemEndingAt(a.state.ItemPos + 1)
		f.Content.Refresh()
		return
	}
	if a.state.FramePos < len(list)-a.state.FrameSize {
		f.Items = f.Items[1:]
		f.Items = append(f.Items, NewFrameItem(a.state.FramePos+a.state.FrameSize, a.state.Simple))
		f.Content.RemoveAll()
		for i := 0; i < a.state.FrameSize; i++ {
			f.Content.Add(f.Items[i].Content)
		}
		f.ItemEndingAt(a.state.FrameSize - 1)
		f.Content.Refresh()
		a.state.FramePos++
		f.updateFrameScrollButtons()
	}
}

func (f *Frame) RemoveItem() {
	if a.state.FrameSize > MinFrameSize {
		a.statusInfo.ShowProgress()
		defer a.statusInfo.HideProgress()
		newCols, newSize := shapeFame(shapeSmaller)
		f.Items = f.Items[:len(f.Items)-a.state.FrameSize+newSize]
		a.state.FrameSize = newSize
		f.Cols = newCols
		f.Content.RemoveAll()
		f.Content.Layout = layout.NewGridLayoutWithColumns(newCols)
		for i := 0; i < a.state.FrameSize; i++ {
			f.Content.Add(f.Items[i].Content)
		}
		f.ItemEndingAt(0)
		f.Content.Refresh()
		f.updateFrameScrollButtons()
		a.showFrameToolbar()
	}
}

func (f *Frame) AddItem() {
	if a.state.FrameSize < MaxFrameSize {
		a.statusInfo.ShowProgress()
		defer a.statusInfo.HideProgress()
		newCols, newSize := shapeFame(shapeBigger)
		if a.state.FramePos+a.state.FrameSize >= len(list) {
			f.At(len(list) - newSize)
		}
		for i := 0; i < newSize-a.state.FrameSize; i++ {
			f.Items = append(f.Items, NewFrameItem(a.state.FramePos+a.state.FrameSize+i, a.state.Simple))
		}
		a.state.FrameSize = newSize
		f.Cols = newCols
		f.Content.RemoveAll()
		f.Content.Layout = layout.NewGridLayoutWithColumns(newCols)
		for i := 0; i < a.state.FrameSize; i++ {
			f.Content.Add(f.Items[i].Content)
		}
		f.ItemEndingAt(a.state.FrameSize - 1)
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
	a.statusInfo.HideProgress()
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
	if a.state.FramePos == 0 {
		f.Buttons[prevFrameBtn].Disable()
		f.Buttons[firstPhotoBtn].Disable()
	}
	if a.state.FramePos+a.state.FrameSize == len(list) {
		f.Buttons[nextFrameBtn].Disable()
		f.Buttons[lastPhotoBtn].Disable()
	}
	if a.state.ItemPos == 0 && a.state.FramePos == 0 {
		f.Buttons[prevPhotoBtn].Disable()
	}
	if a.state.ItemPos == a.state.FrameSize-1 && a.state.FramePos+a.state.FrameSize == len(list) {
		f.Buttons[nextPhotoBtn].Disable()
	}
}

func (f *Frame) ItemEndingAt(pos int) {
	if pos >= a.state.FrameSize {
		a.state.ItemPos = a.state.FrameSize - 1
	} else {
		a.state.ItemPos = pos
	}
	for i := 0; i < a.state.FrameSize; i++ {
		f.Items[i].Ending.StrokeWidth = 0
	}
	f.Items[a.state.ItemPos].Ending.StrokeWidth = 5
	f.Items[a.state.ItemPos].Content.Refresh()
}

type FrameItem struct {
	Content *fyne.Container
	Label   string
	Img     *canvas.Image
	Button  *widget.Button
	Ending  *canvas.Rectangle
}

var DefaultTranslucency = 0.75

// NewFrameItem creates a new FrameItem
func NewFrameItem(listPos int, simpleMode bool) *FrameItem {
	item := &FrameItem{}
	item.Img = GetListImageAt(list[listPos])
	item.Label = fmt.Sprint(list[listPos].fileURI.Name())
	item.Button = widget.NewButton("", func() { toggleDrop(listPos, item) })
	item.Ending = &canvas.Rectangle{FillColor: color.Transparent,
		StrokeColor: theme.PrimaryColor(),
		// StrokeColor:  color.White,
		StrokeWidth:  0,
		CornerRadius: 0,
	}
	if list[listPos].isDroped() {
		item.Button.SetText("DROPPED")
		item.Img.Translucency = DefaultTranslucency
	}
	if list[listPos].isCropped() {
		list[listPos].fadeByCrop(item.Img)
	}
	if list[listPos].isAjusted() {
		list[listPos].adjustByFilters(item.Img)
	}

	topLabel := widget.NewLabelWithStyle(item.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	topLabel.Truncation = fyne.TextTruncateEllipsis
	centerStack := container.NewStack()
	centerStack.Add(item.Button)
	centerStack.Add(item.Img)
	centerStack.Add(item.Ending)
	if simpleMode {
		item.Content = container.NewBorder(topLabel, nil, nil, nil, centerStack)
	} else {
		item.Content = container.NewBorder(topLabel, newDateInput(listPos), nil, nil, centerStack)
	}
	debug.FreeOSMemory() // keeps app slim
	return item
}
func newDateInput(listPos int) *fyne.Container {
	p := list[listPos]
	choices := []string{"EXIF", "File", "Input"}
	d := listDateToDisplayDate(p.Dates[p.DateUsed])

	eDate := widget.NewEntry()
	eDate.Validator = validation.NewTime(a.state.DisplayDateFormat)
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
		item.Img.Translucency = DefaultTranslucency
	} else {
		item.Button.SetText("")
		item.Img.Translucency = 0
	}
	for i := 0; i < a.state.FrameSize; i++ {
		if frame.Items[i] == item {
			frame.ItemEndingAt(i)
		}
	}
}
