package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type order int

const (
	natOrder order = iota
	ascOrder
	descOrder

	DefaultListOrderColumn int   = 0
	DefaultListOrder       order = ascOrder
)

type ListColumn struct {
	Name     string
	Sortable bool
	Order    order
	Width    float32
}

func (a *App) newListView() {
	a.listColumns = []*ListColumn{
		{Name: "File name", Sortable: true},
		{Name: "Dropped"},
		{Name: "Exif date", Sortable: true},
		{Name: "File date", Sortable: true},
		{Name: "Entered date", Sortable: true},
		{Name: "Width x Height"},
		{Name: "Size MB", Sortable: true},
	}

	a.listColumns[0].Width = columnWidth(a.listColumns[0].Name, DisplayDateFormat, FileNameDateFormat+".000")
	a.listColumns[1].Width = columnWidth(a.listColumns[1].Name)
	a.listColumns[2].Width = columnWidth(a.listColumns[2].Name, DisplayDateFormat)
	a.listColumns[3].Width = columnWidth(a.listColumns[3].Name, DisplayDateFormat)
	a.listColumns[4].Width = columnWidth(a.listColumns[4].Name, DisplayDateFormat)
	a.listColumns[5].Width = columnWidth(a.listColumns[5].Name, "0000x0000")
	a.listColumns[6].Width = columnWidth(a.listColumns[6].Name, "999.9")
	for i := 0; i < len(a.listColumns); i++ {
		if a.listColumns[i].Sortable && a.state.ListOrderColumn == i {
			a.listColumns[i].Order = a.state.ListOrder
		}
	}

	a.listTable = widget.NewTableWithHeaders(a.dataLength, a.dataCreate, a.dataUpdate)
	a.listTable.CreateHeader = a.headerCreate
	a.listTable.UpdateHeader = a.headerUpdate
	for i := 0; i < len(a.listColumns); i++ {
		a.listTable.SetColumnWidth(i, a.listColumns[i].Width)
	}

	bottomCount := widget.NewLabel(fmt.Sprintf("total: %d", len(list)))
	bottomCount.TextStyle.Bold = true

	a.listView = container.NewBorder(nil, nil, nil, nil, container.NewBorder(nil, bottomCount, nil, nil, a.listTable))
}

func columnWidth(labels ...string) float32 {
	width := float32(0)
	re := regexp.MustCompile(`\w`)
	for _, l := range labels {
		l := re.ReplaceAllString(l, "0")
		if w := widget.NewLabelWithStyle(l, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}).MinSize().Width; w > width {
			width = w
		}
	}
	return width
}

func (a *App) dataLength() (rows int, cols int) {
	return len(list), len(a.listColumns)
}

func (a *App) dataCreate() fyne.CanvasObject {
	data := widget.NewLabel(DisplayDateFormat)
	data.Truncation = fyne.TextTruncateEllipsis
	return data
}

func (a *App) dataUpdate(id widget.TableCellID, o fyne.CanvasObject) {
	if id.Row == -1 {
		return
	}
	text := ""
	photo := list[id.Row]
	data := o.(*widget.Label)
	switch id.Col {
	case 0:
		text = filepath.Base(photo.File)
		data.TextStyle.Bold = false
		data.Alignment = fyne.TextAlignLeading
	case 1:
		if photo.Dropped {
			text = "Yes"
		} else {
			text = ""
		}
		data.TextStyle.Bold = true
		data.Alignment = fyne.TextAlignCenter
	case 2, 3, 4:
		text = listDateToDisplayDate(photo.Dates[id.Col-2])
		if id.Col-2 == photo.DateUsed {
			data.TextStyle.Bold = true
		} else {
			data.TextStyle.Bold = false
		}
		data.Alignment = fyne.TextAlignLeading
	case 5:
		text = fmt.Sprintf("%dx%d", photo.Width, photo.Height)
		data.TextStyle.Bold = false
		data.Alignment = fyne.TextAlignLeading
	case 6:
		text = fmt.Sprintf("%.1f", float64(photo.ByteSize)/1000000.0)
		data.TextStyle.Bold = false
		data.Alignment = fyne.TextAlignTrailing
	case 7:
		text = " "
		data.TextStyle.Bold = false
		data.Alignment = fyne.TextAlignLeading
	}
	data.SetText(text)
	a.listTable.OnSelected = func(id widget.TableCellID) {
		a.listTable.Unselect(id)
	}
}

func (a *App) headerCreate() fyne.CanvasObject {
	return widget.NewButton("000", nil)

}
func (a *App) headerUpdate(id widget.TableCellID, o fyne.CanvasObject) {
	header := o.(*widget.Button)
	if id.Col == -1 {
		header.SetText(strconv.Itoa(id.Row + 1))
		if id.Row >= frame.Pos && id.Row < frame.Pos+frame.Size {
			header.Importance = widget.HighImportance
		} else {
			header.Importance = widget.MediumImportance
		}
		// header.Icon = theme.NavigateBackIcon()
		header.OnTapped = func() {
			frame.Pos = id.Row
			a.toggleView()
		}
		header.Refresh()
	} else {
		orderIcons := []fyne.Resource{nil, theme.MoveUpIcon(), theme.MoveDownIcon()}
		header.Icon = orderIcons[a.listColumns[id.Col].Order]
		header.SetText(a.listColumns[id.Col].Name)
		if a.listColumns[id.Col].Sortable {
			header.Importance = widget.MediumImportance
		} else {
			// header.Importance = widget.LowImportance
			header.Disable()
		}
		header.OnTapped = func() {
			if a.listColumns[id.Col].Sortable {
				a.reorderList(id.Col)
				header.Refresh()
			}
		}
	}
}

func (a *App) reorderList(col int) {
	order := a.listColumns[col].Order
	order++
	if order > descOrder {
		order = natOrder
	}
	for i := 0; i < len(a.listColumns); i++ {
		a.listColumns[i].Order = natOrder
	}
	a.listColumns[col].Order = order

	posId := list[frame.Pos].id
	sortList(col, order)
	for i := 0; i < len(list); i++ {
		if list[i].id == posId {
			frame.Pos = i
			break
		}
	}
	if frame.Pos+frame.Size > len(list) {
		frame.Pos = len(list) - frame.Size
	}
	a.listTable.ScrollTo(widget.TableCellID{Col: 0, Row: frame.Pos})
	a.listTable.Refresh()
}

func sortList(column int, order order) {
	sort.Slice(list, func(i, j int) bool {
		a := list[i]
		b := list[j]

		if order == natOrder {
			return a.id < b.id
		}

		switch column {
		case 0:
			if order == ascOrder {
				return a.File < b.File
			}
			return a.File > b.File
		case 2:
			if order == ascOrder {
				return a.Dates[UseExifDate] < b.Dates[UseExifDate]
			}
			return a.Dates[UseExifDate] > b.Dates[UseExifDate]
		case 3:
			if order == ascOrder {
				return a.Dates[UseFileDate] < b.Dates[UseFileDate]
			}
			return a.Dates[UseFileDate] > b.Dates[UseFileDate]
		case 4:
			if order == ascOrder {
				return a.Dates[UseEnteredDate] < b.Dates[UseEnteredDate]
			}
			return a.Dates[UseEnteredDate] > b.Dates[UseEnteredDate]
		case 6:
			if order == ascOrder {
				return a.ByteSize < b.ByteSize
			}
			return a.ByteSize > b.ByteSize
		default:
			if order == descOrder {
				return a.id > b.id
			}
			return a.id < b.id
		}
	})

}
