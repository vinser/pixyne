package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const (
	InitListPos   = 0
	InitFrameSize = 3
	MinFrameSize  = 1
	MaxFrameSize  = 6
)

const (
	AddColumn = iota
	RemoveColumn
)

// Photolist
type PhotoList struct {
	Folder string
	List   []*Photo
	Order  func(i, j int) bool
}

// create new PhotoList object for the folder
func (a *App) newPhotoList(folder string) {
	a.folder, _ = filepath.Abs(folder)
	files, err := os.ReadDir(a.folder)
	if err != nil {
		log.Fatalf("Can't list photo files from folder \"%s\". Error: %v\n", a.folder, err)
	}
	photos := []*Photo(nil)
	for _, f := range files {
		fName := strings.ToLower(f.Name())
		if strings.HasSuffix(fName, ".jpg") || strings.HasSuffix(fName, ".jpeg") {
			p := &Photo{
				File:       filepath.Join(folder, f.Name()),
				Droped:     false,
				DateChoice: ChoiceExifDate,
				Dates:      [3]string{},
			}
			p.Dates[ChoiceExifDate] = GetExifDate(p.File)
			p.Dates[ChoiceFileDate] = p.GetModifyDate()
			if len(p.Dates[ChoiceExifDate]) != len(DisplyDateFormat) {
				p.DateChoice = ChoiceFileDate
			}
			photos = append(photos, p)
		}
	}
	a.PhotoList = &PhotoList{
		List:  photos,
		Order: a.orderByFileNameAsc,
	}
}

// Save choosed photos:
// 1. move dropped photo to droppped folder
// 2. update exif dates with file modify date or input date
func (a *App) savePhotoList() {
	dateFileNames := false
	dateFileFormat := time.Now().Format(FileNameDateFormat)
	content := container.NewVBox(
		widget.NewLabel("Ready to save changes?"),
		widget.NewCheck("Rename files to date taken format "+dateFileFormat, func(b bool) { dateFileNames = b }),
	)
	d := dialog.NewCustomConfirm(
		"Save changes",
		"Proceed",
		"Cancel",
		content,
		func(b bool) {
			if b {
				dropDirOk := false
				dropDirName := filepath.Join(a.folder, "dropped")
				backupDirOk := false
				backupDirName := filepath.Join(a.folder, "original")
				for _, p := range a.List {
					if p.Droped {
						// move file to drop dir
						if !dropDirOk {
							err := os.Mkdir(dropDirName, 0775)
							if err != nil && !errors.Is(err, fs.ErrExist) {
								dialog.ShowError(err, a.topWindow)
							}
						}
						os.Rename(p.File, filepath.Join(dropDirName, filepath.Base(p.File)))
						continue
					}
					if p.DateChoice != ChoiceExifDate {
						// backup original file and make file copy with modified exif
						if !backupDirOk {
							err := os.Mkdir(backupDirName, 0775)
							if err != nil && !errors.Is(err, fs.ErrExist) {
								dialog.ShowError(err, a.topWindow)
							}
						}
						if UpdateExifDate(p.File, backupDirName, p.Dates[p.DateChoice]) == nil {
							if dateFileNames {
								os.Rename(p.File, pathNameToDate(p.File, p.Dates[p.DateChoice]))
							}
							continue
						}
					}
					if dateFileNames {
						// backup original file and rename file by date format "20060102_150405"
						if !backupDirOk {
							err := os.Mkdir(backupDirName, 0775)
							if err != nil && !errors.Is(err, fs.ErrExist) {
								dialog.ShowError(err, a.topWindow)
							}
						}
						fileCopy(p.File, filepath.Join(backupDirName, filepath.Base(p.File)))
						os.Rename(p.File, pathNameToDate(p.File, p.Dates[p.DateChoice]))
					}
				}
			}
		},
		a.topWindow)
	d.Show()
}

const (
	unordered = iota
	orderAsc
	orderDesc
)

var orderSymbols = []string{" ", " ↓", " ↑"}

func (a *App) newListView() {
	a.listColumnsNum = 6
	a.listHeaders = make([]*ActiveHeader, a.listColumnsNum)

	dateTemplate := DisplyDateFormat
	dateWidth := widget.NewLabel(dateTemplate).MinSize().Width
	fileNameWidth := dateWidth
	for _, ph := range a.List {
		fName := filepath.Base(ph.File)
		if len(fName) > len(dateTemplate) {
			fileNameWidth = widget.NewLabel(fName).MinSize().Width
		}
	}
	a.headerRow = widget.NewTable(
		func() (int, int) {
			return 1, a.listColumnsNum
		},
		func() fyne.CanvasObject {
			header := newActiveHeader(dateTemplate)
			return header
		},
		a.headerAction,
	)
	a.headerRow.SetColumnWidth(0, 20)
	a.headerRow.SetColumnWidth(1, fileNameWidth)
	a.headerRow.SetColumnWidth(2, dateWidth)
	a.headerRow.SetColumnWidth(3, dateWidth)
	a.headerRow.SetColumnWidth(4, dateWidth)
	a.headerRow.SetColumnWidth(5, 80)

	a.dataRows = widget.NewTable(
		func() (int, int) {
			return len(a.List), a.listColumnsNum
		},
		func() fyne.CanvasObject {
			data := newActiveCell(dateTemplate)
			return data
		},
		a.dataAction,
	)
	a.dataRows.SetColumnWidth(0, 20)
	a.dataRows.SetColumnWidth(1, fileNameWidth)
	a.dataRows.SetColumnWidth(2, dateWidth)
	a.dataRows.SetColumnWidth(3, dateWidth)
	a.dataRows.SetColumnWidth(4, dateWidth)
	a.dataRows.SetColumnWidth(5, 80)

	a.dataRows.OnSelected = a.syncHeader // poor attempt to synchronize table scrolling
	bottomCount := widget.NewLabel(fmt.Sprintf("total: %d", len(a.List)))
	a.listView = container.NewBorder(a.headerRow, bottomCount, nil, nil, a.dataRows)
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
		data.OnTapped = func() {
			a.scrollFrame(cell.Row)
			a.toggleView()
		}
	case 2, 3, 4:
		text = photo.Dates[cell.Col-2]
		if cell.Col-1 == photo.DateChoice {
			data.TextStyle.Bold = true
		} else {
			data.TextStyle.Bold = false
		}
	case 5:
		if photo.Droped {
			text = "Yes"
			data.TextStyle.Bold = true
		}
	}
	data.SetText(text)
}

type ActiveHeader struct {
	widget.Label
	Name     string
	Order    int
	SortAsc  func(i, j int) bool
	SortDesc func(i, j int) bool
	OnTapped func()
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
	settings := []struct {
		Title    string
		SortAsc  func(i, j int) bool
		SortDesc func(i, j int) bool
		Default  bool
	}{
		{"", nil, nil, false},
		{"File Name", a.orderByFileNameAsc, a.orderByFileNameDesc, true},
		{"Exif Date", nil, nil, false},
		{"File Date", a.orderByFileDateAsc, a.orderByFileDateDesc, false},
		{"Entered Date", nil, nil, false},
		{"Dropped", nil, nil, false},
	}

	header := o.(*ActiveHeader)
	a.listHeaders[cell.Col] = header
	header.TextStyle.Bold = true
	if settings[cell.Col].Default && !a.inited {
		header.Order = orderAsc
	}
	header.Label.SetText(settings[cell.Col].Title + orderSymbols[header.Order])
	if settings[cell.Col].SortAsc == nil && settings[cell.Col].SortDesc == nil {
		header.TextStyle.Italic = true
		return
	}
	header.OnTapped = func() {
		for j, h := range a.listHeaders {
			if j == cell.Col {
				switch h.Order {
				case unordered, orderDesc:
					h.Order = orderAsc
					a.reorderList(settings[cell.Col].SortAsc)
				case orderAsc:
					h.Order = orderDesc
					a.reorderList(settings[cell.Col].SortDesc)
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
	return a.List[i].Dates[ChoiceFileDate] < a.List[j].Dates[ChoiceFileDate]
}

func (a *App) orderByFileDateDesc(i, j int) bool {
	return a.List[j].Dates[ChoiceFileDate] < a.List[i].Dates[ChoiceFileDate]
}
