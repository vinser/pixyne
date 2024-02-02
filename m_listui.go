package main

import (
	"fmt"
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

func (a *App) newListView() *fyne.Container {
	a.listColumns = []*ListColumn{
		{Name: "File name", Sortable: true},
		{Name: "Exif date", Sortable: true},
		{Name: "File date", Sortable: true},
		{Name: "Entered date", Sortable: true},
		{Name: "Drop"},
		{Name: "Crop"},
		{Name: "Ajust"},
		{Name: "Pixels WxH"},
		{Name: "Size MiB", Sortable: true},
	}

	a.listColumns[0].Width = columnWidth(a.listColumns[0].Name, a.state.DisplayDateFormat, FileNameDateFormat+".000")
	a.listColumns[1].Width = columnWidth(a.listColumns[1].Name, a.state.DisplayDateFormat)
	a.listColumns[2].Width = columnWidth(a.listColumns[2].Name, a.state.DisplayDateFormat)
	a.listColumns[3].Width = columnWidth(a.listColumns[3].Name, a.state.DisplayDateFormat)
	a.listColumns[4].Width = columnWidth(a.listColumns[4].Name)
	a.listColumns[5].Width = columnWidth(a.listColumns[5].Name)
	a.listColumns[6].Width = columnWidth(a.listColumns[6].Name)
	a.listColumns[7].Width = columnWidth(a.listColumns[7].Name, "0000x0000")
	a.listColumns[8].Width = columnWidth(a.listColumns[8].Name, "999.9")
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

	return container.NewBorder(nil, nil, nil, nil, container.NewBorder(nil, bottomCount, nil, nil, a.listTable))
}

func columnWidth(labels ...string) float32 {
	width := float32(0)
	re := regexp.MustCompile(`\w`)
	for _, l := range labels {
		l := re.ReplaceAllString(l, "0")
		if w := widget.NewLabelWithStyle(l, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}).MinSize().Width + theme.Padding(); w > width {
			width = w
		}
	}
	return width
}

func (a *App) dataLength() (rows int, cols int) {
	return len(list), len(a.listColumns)
}

func (a *App) dataCreate() fyne.CanvasObject {
	data := widget.NewLabel(a.state.DisplayDateFormat)
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
		text = photo.fileURI.Name()
		data.TextStyle.Bold = false
		data.Alignment = fyne.TextAlignLeading
	case 1, 2, 3:
		text = listDateToDisplayDate(photo.Dates[id.Col-1])
		if id.Col-1 == photo.DateUsed {
			data.TextStyle.Bold = true
		} else {
			data.TextStyle.Bold = false
		}
		data.Alignment = fyne.TextAlignLeading
	case 4:
		if photo.isDroped() {
			text = "Yes"
		} else {
			text = ""
		}
		data.TextStyle.Bold = true
		data.Alignment = fyne.TextAlignCenter
	case 5:
		if photo.isCropped() {
			text = "Yes"
		} else {
			text = ""
		}
		data.TextStyle.Bold = true
		data.Alignment = fyne.TextAlignCenter
	case 6:
		if photo.isAdjusted() {
			text = "Yes"
		} else {
			text = ""
		}
		data.TextStyle.Bold = true
		data.Alignment = fyne.TextAlignCenter
	case 7:
		text = fmt.Sprintf("%dx%d", photo.width, photo.height)
		data.TextStyle.Bold = false
		data.Alignment = fyne.TextAlignLeading
	case 8:
		text = fmt.Sprintf("%.1f", float64(photo.byteSize)/1000000.0)
		data.TextStyle.Bold = false
		data.Alignment = fyne.TextAlignTrailing
	case 9:
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
		if id.Row >= a.state.FramePos && id.Row < a.state.FramePos+a.state.FrameSize {
			header.Importance = widget.HighImportance
		} else {
			header.Importance = widget.MediumImportance
		}
		// header.Icon = theme.NavigateBackIcon()
		header.OnTapped = func() {
			a.statusInfo.ShowProgress()
			defer a.statusInfo.HideProgress()
			a.state.FramePos = id.Row
			a.state.ItemPos = 0
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

	posId := list[a.state.FramePos].id
	sortList(col, order)
	for i := 0; i < len(list); i++ {
		if list[i].id == posId {
			a.state.FramePos = i
			break
		}
	}
	if a.state.FramePos+a.state.FrameSize > len(list) {
		a.state.FramePos = len(list) - a.state.FrameSize
	}
	a.syncListViewScroll()
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
				return a.fileURI.Name() < b.fileURI.Name()
			}
			return a.fileURI.Name() > b.fileURI.Name()
		case 1:
			if order == ascOrder {
				return a.Dates[UseExifDate] < b.Dates[UseExifDate]
			}
			return a.Dates[UseExifDate] > b.Dates[UseExifDate]
		case 2:
			if order == ascOrder {
				return a.Dates[UseFileDate] < b.Dates[UseFileDate]
			}
			return a.Dates[UseFileDate] > b.Dates[UseFileDate]
		case 3:
			if order == ascOrder {
				return a.Dates[UseEnteredDate] < b.Dates[UseEnteredDate]
			}
			return a.Dates[UseEnteredDate] > b.Dates[UseEnteredDate]
		case 8:
			if order == ascOrder {
				return a.byteSize < b.byteSize
			}
			return a.byteSize > b.byteSize
		default:
			if order == descOrder {
				return a.id > b.id
			}
			return a.id < b.id
		}
	})

}
