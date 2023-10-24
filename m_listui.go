package main

import (
	"fmt"
	"path/filepath"
	"regexp"
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
	a.listColumns = []*ListCell{
		{Name: "File Name", Order: orderAsc, SortAsc: a.orderByFileNameAsc, SortDesc: a.orderByFileNameDesc},
		{Name: "Exif Date", SortAsc: a.orderByExifDateAsc, SortDesc: a.orderByExifDateDesc},
		{Name: "File Date", SortAsc: a.orderByFileDateAsc, SortDesc: a.orderByFileDateDesc},
		{Name: "Entered Date", SortAsc: a.orderByEnteredDateAsc, SortDesc: a.orderByEnteredDateDesc},
		{Name: "Dropped"},
	}

	a.listColumns[0].Width = columnWidth(a.listColumns[0].Name, DisplayDateFormat, FileNameDateFormat+".000")
	a.listColumns[1].Width = columnWidth(a.listColumns[1].Name, DisplayDateFormat)
	a.listColumns[2].Width = columnWidth(a.listColumns[2].Name, DisplayDateFormat)
	a.listColumns[3].Width = columnWidth(a.listColumns[3].Name, DisplayDateFormat)
	a.listColumns[4].Width = columnWidth(a.listColumns[4].Name)

	a.listTable = widget.NewTableWithHeaders(a.dataLength, a.dataCreate, a.dataUpdate)
	a.listTable.CreateHeader = a.headerCreate
	a.listTable.UpdateHeader = a.headerUpdate

	for i := 0; i < len(a.listColumns); i++ {
		a.listTable.SetColumnWidth(i, a.listColumns[i].Width)
	}
	bottomCount := widget.NewLabel(fmt.Sprintf("total: %d", len(list)))
	a.listView = container.NewBorder(nil, nil, nil, nil, container.NewBorder(nil, bottomCount, nil, nil, a.listTable))
}

func columnWidth(labels ...string) float32 {
	width := widget.NewLabel("").MinSize().Width
	re := regexp.MustCompile(`\w`)
	for _, l := range labels {
		l := re.ReplaceAllString(l, "0")
		if w := widget.NewLabelWithStyle(l, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}).MinSize().Width; w > width {
			width = w
		}
	}
	return width
}

type ListCell struct {
	widget.Label
	Name     string
	Width    float32
	Order    int
	OnTapped func()
	SortAsc  func(i, j int) bool
	SortDesc func(i, j int) bool
}

func newListCell(label string) *ListCell {
	h := &ListCell{}
	h.ExtendBaseWidget(h)
	h.Label.SetText(label)
	return h
}

func (c *ListCell) Tapped(_ *fyne.PointEvent) {
	if c.OnTapped != nil {
		c.OnTapped()
	}
}

func (c *ListCell) TappedSecondary(_ *fyne.PointEvent) {
}

func (a *App) dataLength() (rows int, cols int) {
	return len(list), len(a.listColumns)
}

func (a *App) dataCreate() fyne.CanvasObject {
	data := newListCell(DisplayDateFormat)
	data.Truncation = fyne.TextTruncateEllipsis
	return data
}

func (a *App) dataUpdate(id widget.TableCellID, o fyne.CanvasObject) {
	if id.Row == -1 {
		return
	}
	text := ""
	photo := list[id.Row]
	data := o.(*ListCell)
	switch id.Col {
	case 0:
		text = filepath.Base(photo.File)
		data.TextStyle.Bold = false
	case 1, 2, 3:
		text = listDateToDisplayDate(photo.Dates[id.Col-1])
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

func (a *App) headerCreate() fyne.CanvasObject {
	header := newListCell("000")
	return header

}
func (a *App) headerUpdate(id widget.TableCellID, o fyne.CanvasObject) {
	header := o.(*ListCell)
	header.TextStyle.Bold = true
	if id.Col == -1 {
		header.Label.SetText(strconv.Itoa(id.Row + 1))
		return
	} else {
		header.Label.SetText(a.listColumns[id.Col].Name + orderSymbols[a.listColumns[id.Col].Order])
		if a.listColumns[id.Col].SortAsc == nil && a.listColumns[id.Col].SortDesc == nil {
			header.TextStyle.Italic = true
			return
		}
	}
	header.OnTapped = func() {
		for j, h := range a.listColumns {
			if j == id.Col {
				switch h.Order {
				case unordered, orderDesc:
					h.Order = orderAsc
					a.reorderList(a.listColumns[id.Col].SortAsc)
				case orderAsc:
					h.Order = orderDesc
					a.reorderList(a.listColumns[id.Col].SortDesc)
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
	sort.Slice(list, less)
}

func (a *App) orderByFileNameAsc(i, j int) bool {
	return list[i].File < list[j].File
}

func (a *App) orderByFileNameDesc(i, j int) bool {
	return list[j].File < list[i].File
}

func (a *App) orderByExifDateAsc(i, j int) bool {
	return list[i].Dates[UseExifDate] < list[j].Dates[UseExifDate]
}

func (a *App) orderByExifDateDesc(i, j int) bool {
	return list[j].Dates[UseExifDate] < list[i].Dates[UseExifDate]
}

func (a *App) orderByFileDateAsc(i, j int) bool {
	return list[i].Dates[UseFileDate] < list[j].Dates[UseFileDate]
}

func (a *App) orderByFileDateDesc(i, j int) bool {
	return list[j].Dates[UseFileDate] < list[i].Dates[UseFileDate]
}

func (a *App) orderByEnteredDateAsc(i, j int) bool {
	return list[i].Dates[UseEnteredDate] < list[j].Dates[UseEnteredDate]
}

func (a *App) orderByEnteredDateDesc(i, j int) bool {
	return list[j].Dates[UseEnteredDate] < list[i].Dates[UseEnteredDate]
}
