package main

import (
	"fmt"
	"path/filepath"
	"sort"

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
		{},
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

	a.listHeaders[0].Width = labelMaxWidth("000")
	a.listHeaders[1].Width = labelMaxWidth(a.listHeaders[1].Name, DisplyDateFormat, fileNameTemplate)
	a.listHeaders[2].Width = labelMaxWidth(a.listHeaders[2].Name, DisplyDateFormat)
	a.listHeaders[3].Width = labelMaxWidth(a.listHeaders[3].Name, DisplyDateFormat)
	a.listHeaders[4].Width = labelMaxWidth(a.listHeaders[4].Name, DisplyDateFormat)
	a.listHeaders[5].Width = labelMaxWidth(a.listHeaders[5].Name)

	a.headerRow = widget.NewTable(
		func() (int, int) {
			return 1, a.listColumnsNum
		},
		func() fyne.CanvasObject {
			header := newActiveHeader(DisplyDateFormat)
			return header
		},
		a.headerAction,
	)
	a.dataRows = widget.NewTable(
		func() (int, int) {
			return len(a.List), a.listColumnsNum
		},
		func() fyne.CanvasObject {
			data := newActiveCell(DisplyDateFormat)
			return data
		},
		a.dataAction,
	)

	for i := 0; i < len(a.listHeaders); i++ {
		a.headerRow.SetColumnWidth(i, a.listHeaders[i].Width)
		a.dataRows.SetColumnWidth(i, a.listHeaders[i].Width)
	}
	a.dataRows.OnSelected = a.syncHeader // silly attempt to synchronize table scrolling
	bottomCount := widget.NewLabel(fmt.Sprintf("total: %d", len(a.List)))
	a.listView = container.NewBorder(a.headerRow, bottomCount, nil, nil, a.dataRows)
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

func (a *App) dataAction(cell widget.TableCellID, o fyne.CanvasObject) {
	text := ""
	photo := a.List[cell.Row]
	data := o.(*ActiveCell)
	switch cell.Col {
	case 0:
		text = fmt.Sprint(cell.Row + 1)
	case 1:
		text = filepath.Base(photo.File)
		data.TextStyle.Bold = false
	case 2, 3, 4:
		text = photo.Dates[cell.Col-2]
		if cell.Col-2 == photo.DateUsed {
			data.TextStyle.Bold = true
		} else {
			data.TextStyle.Bold = false
		}
	case 5:
		if photo.Dropped {
			text = "Yes"
			data.TextStyle.Bold = true
		}
	}
	data.SetText(text)
	data.OnTapped = func() {
		a.scrollFrame(cell.Row)
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

func (a *App) headerAction(cell widget.TableCellID, o fyne.CanvasObject) {
	header := o.(*ActiveHeader)
	header.TextStyle.Bold = true
	header.Label.SetText(a.listHeaders[cell.Col].Name + orderSymbols[a.listHeaders[cell.Col].Order])
	if a.listHeaders[cell.Col].SortAsc == nil && a.listHeaders[cell.Col].SortDesc == nil {
		header.TextStyle.Italic = true
		return
	}
	header.OnTapped = func() {
		for j, h := range a.listHeaders {
			if j == cell.Col {
				switch h.Order {
				case unordered, orderDesc:
					h.Order = orderAsc
					a.reorderList(a.listHeaders[cell.Col].SortAsc)
				case orderAsc:
					h.Order = orderDesc
					a.reorderList(a.listHeaders[cell.Col].SortDesc)
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

func (a *App) syncHeader(cell widget.TableCellID) {
	cell.Row = 0
	a.headerRow.ScrollTo(cell)
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
