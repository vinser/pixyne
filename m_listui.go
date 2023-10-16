package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	unordered = iota
	orderAsc
	orderDesc
)

var orderSymbols = []string{" ", " ↓", " ↑"}

func (a *App) newListView() {
	a.listHeaders = []*ActiveHeader{
		{Name: "File Name", Order: orderAsc, SortAsc: a.orderByFileNameAsc, SortDesc: a.orderByFileNameDesc},
		{Name: "Exif Date"},
		{Name: "File Date", SortAsc: a.orderByFileDateAsc, SortDesc: a.orderByFileDateDesc},
		{Name: "Entered Date"},
		{Name: "Dropped"},
	}
	a.listColumnsNum = len(a.listHeaders)
	fileNameTemplate := ""
	for _, ph := range a.List {
		if fName := filepath.Base(ph.File); len(fName) > len(fileNameTemplate) {
			fileNameTemplate = fName
		}
	}

	a.listHeaders[0].Width = labelMaxWidth(a.listHeaders[0].Name, DisplyDateFormat, fileNameTemplate)
	a.listHeaders[1].Width = labelMaxWidth(a.listHeaders[1].Name, DisplyDateFormat)
	a.listHeaders[2].Width = labelMaxWidth(a.listHeaders[2].Name, DisplyDateFormat)
	a.listHeaders[3].Width = labelMaxWidth(a.listHeaders[3].Name, DisplyDateFormat)
	a.listHeaders[4].Width = labelMaxWidth(a.listHeaders[4].Name)

	a.listTable = widget.NewTableWithHeaders(a.dataLength, a.dataCreate, a.dataUpdate)
	a.listTable.CreateHeader = a.headerCreate
	a.listTable.UpdateHeader = a.headerUpdate

	for i := 0; i < len(a.listHeaders); i++ {
		a.listTable.SetColumnWidth(i, a.listHeaders[i].Width)
	}
	bottomCount := widget.NewLabel(fmt.Sprintf("total: %d", len(a.List)))
	a.listView = container.NewBorder(nil, bottomCount, nil, nil, a.listTable)
}

func labelMaxWidth(labels ...string) float32 {
	width := widget.NewLabel("").MinSize().Width
	for _, l := range labels {
		if w := widget.NewLabel(l).MinSize().Width; w > width {
			width = w
		}
	}
	return width
}

type ActiveCell struct {
	widget.Label
	OnTapped func()
}

func newActiveCell(label string) *ActiveCell {
	c := &ActiveCell{}
	c.ExtendBaseWidget(c)
	c.Label.SetText(label)
	return c
}

func (h *ActiveCell) Tapped(_ *fyne.PointEvent) {
	if h.OnTapped != nil {
		h.OnTapped()
	}
}

func (h *ActiveCell) TappedSecondary(_ *fyne.PointEvent) {
}
func (a *App) dataLength() (rows int, cols int) {
	return len(a.List), a.listColumnsNum
}

func (a *App) dataCreate() fyne.CanvasObject {
	data := newActiveCell(DisplyDateFormat)
	return data

}
func (a *App) dataUpdate(id widget.TableCellID, o fyne.CanvasObject) {
	if id.Row == -1 {
		return
	}
	text := ""
	photo := a.List[id.Row]
	data := o.(*ActiveCell)
	switch id.Col {
	case 0:
		text = filepath.Base(photo.File)
		data.TextStyle.Bold = false
	case 1, 2, 3:
		text = photo.Dates[id.Col-1]
		if id.Col-1 == photo.DateUsed {
			data.TextStyle.Bold = true
		} else {
			data.TextStyle.Bold = false
		}
	case 4:
		if photo.Dropped {
			text = "Yes"
			data.TextStyle.Bold = true
		}
	}
	data.SetText(text)
	data.OnTapped = func() {
		a.scrollFrame(id.Row)
		a.toggleView()
	}

}

type ActiveHeader struct {
	Name     string
	Width    float32
	Order    int
	SortAsc  func(i, j int) bool
	SortDesc func(i, j int) bool
	OnTapped func()
	widget.Label
}

func newActiveHeader(label string) *ActiveHeader {
	h := &ActiveHeader{}
	h.ExtendBaseWidget(h)
	h.Label.SetText(label)
	return h
}

func (h *ActiveHeader) Tapped(_ *fyne.PointEvent) {
	if h.OnTapped != nil {
		h.OnTapped()
	}
}

func (h *ActiveHeader) TappedSecondary(_ *fyne.PointEvent) {
}

func (a *App) headerCreate() fyne.CanvasObject {
	header := newActiveHeader("000")
	return header

}
func (a *App) headerUpdate(id widget.TableCellID, o fyne.CanvasObject) {
	header := o.(*ActiveHeader)
	header.TextStyle.Bold = true
	if id.Col == -1 {
		header.Label.SetText(strconv.Itoa(id.Row + 1))
		return
	} else {
		header.Label.SetText(a.listHeaders[id.Col].Name + orderSymbols[a.listHeaders[id.Col].Order])
		if a.listHeaders[id.Col].SortAsc == nil && a.listHeaders[id.Col].SortDesc == nil {
			header.TextStyle.Italic = true
			return
		}
	}
	header.OnTapped = func() {
		for j, h := range a.listHeaders {
			if j == id.Col {
				switch h.Order {
				case unordered, orderDesc:
					h.Order = orderAsc
					a.reorderList(a.listHeaders[id.Col].SortAsc)
				case orderAsc:
					h.Order = orderDesc
					a.reorderList(a.listHeaders[id.Col].SortDesc)
				}
				continue
			} else {
				h.Order = unordered
			}
			h.Refresh()
		}
		a.listView.Refresh()
	}
}

// List sort functions

func (a *App) reorderList(less func(i int, j int) bool) {
	sort.Slice(a.List, less)
}

func (a *App) orderByFileNameAsc(i, j int) bool {
	return a.List[i].File < a.List[j].File
}

func (a *App) orderByFileNameDesc(i, j int) bool {
	return a.List[j].File < a.List[i].File
}

func (a *App) orderByFileDateAsc(i, j int) bool {
	return a.List[i].Dates[UseFileDate] < a.List[j].Dates[UseFileDate]
}

func (a *App) orderByFileDateDesc(i, j int) bool {
	return a.List[j].Dates[UseFileDate] < a.List[i].Dates[UseFileDate]
}
