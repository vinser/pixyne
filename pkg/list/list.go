package list

import (
	"errors"
	"image/color"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
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
	Folder    string
	List      []*Photo
	Order     func(i, j int) bool
	Frame     *fyne.Container
	FrameSize int
	FramePos  int
}

// create new PhotoList object for the folder
func newPhotoList(folder string) *PhotoList {
	folder, _ = filepath.Abs(folder)
	files, err := os.ReadDir(folder)
	if err != nil {
		log.Fatalf("Can't list photo files from folder \"%s\". Error: %v\n", folder, err)
	}
	photos := []*Photo(nil)
	for _, f := range files {
		fName := strings.ToLower(f.Name())
		if strings.HasSuffix(fName, ".jpg") || strings.HasSuffix(fName, ".jpeg") {
			photo := &Photo{
				File:       filepath.Join(folder, f.Name()),
				Droped:     false,
				DateChoice: ChoiceExifDate,
				Dates:      [3]string{},
			}
			photo.Dates[ChoiceExifDate] = getExifDate(photo.File)
			photo.Dates[ChoiceFileDate] = photo.getModifyDate()
			if len(photo.Dates[ChoiceExifDate]) != len(DateFormat) {
				photo.DateChoice = ChoiceFileDate
			}
			photos = append(photos, photo)
		}
	}
	l := &PhotoList{
		Folder:    folder,
		List:      photos,
		FrameSize: InitFrameSize,
		FramePos:  InitListPos,
	}
	l.Order = l.orderByFileNameAsc
	return l
}

// make main window layout
func MainLayout(l *PhotoList) {
	l.reorder(l.Order)
	l.initFrame()
	contentTabs := container.NewAppTabs(l.newChoiceTab(), l.newListTab())
	contentTabs.SetTabLocation(container.TabLocationBottom)
	wMain.SetContent(contentTabs)
}

// Save choosed photos:
// 1. move dropped photo to droppped folder
// 2. update exif dates with file modify date or input date
func (l *PhotoList) savePhotoList() {
	dialog.ShowConfirm("Ready to save changes", "Proceed?",
		func(b bool) {
			if b {
				dropDirOk := false
				dropDirName := filepath.Join(l.Folder, "dropped")
				backupDirOk := false
				backupDirName := filepath.Join(l.Folder, "original")
				for _, p := range l.List {
					if p.Droped {
						// move file to drop dir
						if !dropDirOk {
							err := os.Mkdir(dropDirName, 0775)
							if err != nil && !errors.Is(err, fs.ErrExist) {
								dialog.ShowError(err, wMain)
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
								dialog.ShowError(err, wMain)
							}
						}
						updateExifDate(p.File, backupDirName, p.Dates[p.DateChoice])
					}
				}
			}
		},
		wMain)
}

// create new photos tab container
func (l *PhotoList) newListTab() *container.TabItem {
	toolBar := widget.NewToolbar(
		widget.NewToolbarAction(theme.FolderOpenIcon(), chooseFolder),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), l.savePhotoList),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.SettingsIcon(), settingsScreen),
		widget.NewToolbarAction(theme.HelpIcon(), aboutScreen),
	)
	return container.NewTabItemWithIcon("List", theme.ListIcon(), container.NewBorder(toolBar, nil, nil, nil, l.newListTabTable()))
}

const (
	unordered = iota
	orderAsc
	orderDesc
)

var orderSymbols = []string{" ↕", " ↑", " ↓"}

type ActiveHeader struct {
	widget.Label
	Sortable bool
	Order    int
	OnTapped func()
}

func newActiveHeader(label string, tapped func()) *ActiveHeader {
	h := &ActiveHeader{
		Sortable: false,
		Order:    0,
		OnTapped: tapped,
	}
	h.ExtendBaseWidget(h)
	h.Label.SetText(label)
	return h
}

func (h *ActiveHeader) SetText(label string) {
	h.Label.SetText(label + orderSymbols[h.Order])
}

func (h *ActiveHeader) Tapped(_ *fyne.PointEvent) {
	if h.OnTapped != nil {
		h.OnTapped()
	}
}

func (h *ActiveHeader) TappedSecondary(_ *fyne.PointEvent) {
}

func (l *PhotoList) newListTabTable() *fyne.Container {
	listTitle := []string{"File Name", "Exif Date", "File Date", "Entered Date", "Dropped"}

	table := widget.NewTable(
		func() (int, int) {
			return len(l.List), len(listTitle)
		},
		func() fyne.CanvasObject {
			text := DateFormat
			for _, ph := range l.List {
				fName := filepath.Base(ph.File)
				if len(fName) > len(text) {
					text = fName
				}
			}
			data := widget.NewLabel(text)
			return data
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			text := ""
			ph := l.List[i.Row]
			data := o.(*widget.Label)
			switch i.Col {
			case 0:
				text = filepath.Base(ph.File)
				data.TextStyle.Bold = false
			case 1, 2, 3:
				text = ph.Dates[i.Col-1]
				if i.Col-1 == ph.DateChoice {
					data.TextStyle.Bold = true
				} else {
					data.TextStyle.Bold = false
				}
			case 4:
				if ph.Droped {
					text = "Yes"
					data.TextStyle.Bold = true
				}
			}
			data.SetText(text)
		})

	header := widget.NewTable(
		func() (int, int) {
			return 1, len(listTitle)
		},
		func() fyne.CanvasObject {
			text := DateFormat
			for _, ph := range l.List {
				fName := filepath.Base(ph.File)
				if len(fName) > len(text) {
					text = fName
				}
			}
			h := newActiveHeader(text, nil)
			return h
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			h := o.(*ActiveHeader)
			h.SetText(listTitle[i.Col])
			h.TextStyle.Bold = true
			// switch i.Col {
			// case 0:
			// 	h.OnTapped = func() {
			// 		if l.Order == l.orderByFileDateAsc {
			// 			l.reorder(l.orderByFileNameDesc)
			// 		}
			// 	}
			// case 3:
			// 	h.OnTapped = func() { l.reorder(l.orderByFileDateDesc) }
			// }
		})
	return container.NewBorder(header, nil, nil, nil, table)
}

// create new photos tab container
func (l *PhotoList) newChoiceTab() *container.TabItem {
	actOpenFolder := widget.NewToolbarAction(theme.FolderOpenIcon(), chooseFolder)
	actDecFrame := widget.NewToolbarAction(theme.ContentRemoveIcon(), func() { l.resizeFrame(RemoveColumn) })
	actIncFrame := widget.NewToolbarAction(theme.ContentAddIcon(), func() { l.resizeFrame(AddColumn) })
	toolBar := widget.NewToolbar(
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.SettingsIcon(), settingsScreen),
		widget.NewToolbarAction(theme.HelpIcon(), aboutScreen),
	)
	if len(l.List) > 0 {
		toolBar.Prepend(actIncFrame)
		toolBar.Prepend(actDecFrame)
	} else {
		toolBar.Prepend(actOpenFolder)
	}

	prevPhotoBtn := widget.NewButton("<", func() {
		l.scrollFrame(l.FramePos - 1)
	})
	prevFrameBtn := widget.NewButton("<<", func() {
		l.scrollFrame(l.FramePos - l.FrameSize)
	})
	firstPhotoBtn := widget.NewButton("|<", func() {
		l.scrollFrame(0)
	})

	nextPhotoBtn := widget.NewButton(">", func() {
		l.scrollFrame(l.FramePos + 1)
	})
	nextFrameBtn := widget.NewButton(">>", func() {
		l.scrollFrame(l.FramePos + l.FrameSize)
	})
	lastPhotoBtn := widget.NewButton(">|", func() {
		l.scrollFrame(len(l.List))
	})
	bottomButtons := container.NewGridWithColumns(6, firstPhotoBtn, prevFrameBtn, prevPhotoBtn, nextPhotoBtn, nextFrameBtn, lastPhotoBtn)

	return container.NewTabItemWithIcon("Choice", theme.GridIcon(), container.NewBorder(toolBar, bottomButtons, nil, nil, l.Frame))
}

// scroll frame at position pos
func (l *PhotoList) scrollFrame(pos int) {

	switch {
	case pos < 0:
		pos = 0
	case pos > len(l.List)-l.FrameSize:
		pos = len(l.List) - l.FrameSize
	}

	switch {
	case pos-l.FramePos >= l.FrameSize || l.FramePos-pos >= l.FrameSize:
		for i := l.FramePos; i < l.FramePos+l.FrameSize; i++ {
			l.List[i].Img = nil
		}
		for i := pos; i < pos+l.FrameSize; i++ {
			l.List[i].Img = l.List[i].img(l.FrameSize)
			if l.List[i].Droped {
				l.List[i].Img.Translucency = 0.5
			}
		}
	case pos > l.FramePos:
		for i := l.FramePos; i < pos; i++ {
			l.List[i].Img = nil
			l.List[i+l.FrameSize].Img = l.List[i+l.FrameSize].img(l.FrameSize)
			if l.List[i+l.FrameSize].Droped {
				l.List[i+l.FrameSize].Img.Translucency = 0.5
			}
		}
	case l.FramePos > pos:
		for i := pos; i < l.FramePos; i++ {
			l.List[i+l.FrameSize].Img = nil
			l.List[i].Img = l.List[i].img(l.FrameSize)
			if l.List[i].Droped {
				l.List[i].Img.Translucency = 0.5
			}
		}
	}

	// TODO: may be optimized when for scroll les than frame size by not all objects deletion/addition? Somwthing like this:
	// https://stackoverflow.com/questions/63995289/how-to-remove-objects-from-golang-fyne-container
	l.Frame.RemoveAll()
	for i := 0; i < l.FrameSize; i++ {
		l.Frame.Add(l.List[pos+i].FrameColumn())
	}
	l.Frame.Refresh()

	l.FramePos = pos
}

// resize frame
func (l *PhotoList) resizeFrame(zoom int) {

	switch zoom {
	case RemoveColumn:
		if l.FrameSize-1 < MinFrameSize {
			return
		}
		l.List[l.FramePos+l.FrameSize-1].Img = nil
		l.FrameSize--
	case AddColumn:
		if l.FrameSize+1 > MaxFrameSize || l.FrameSize+1 > len(l.List) {
			return
		}
		i := l.FramePos + l.FrameSize
		if i == len(l.List) {
			l.FramePos--
			i = l.FramePos
		}
		l.List[i].Img = l.List[i].img(l.FrameSize)
		if l.List[i].Droped {
			l.List[i].Img.Translucency = 0.5
		}
		l.FrameSize++
	}
	//      0-1-2-3-4-5-6-7-8
	//          2-3-4			p=2, s=3
	// 		0-1-2				p=0, s=3
	// 					6-7-8	p=6, s=3

	// TODO: may be optimized when for scroll les than frame size by not all objects deletion/addition? Somwthing like this:
	// https://stackoverflow.com/questions/63995289/how-to-remove-objects-from-golang-fyne-container
	l.Frame.RemoveAll()
	for i := 0; i < l.FrameSize; i++ {
		l.Frame.Add(l.List[l.FramePos+i].FrameColumn())
	}
	l.Frame.Layout = layout.NewGridLayoutWithColumns(len(l.Frame.Objects))
	l.Frame.Refresh()
}

// fill frame Num photo images starting with Pos = 0.
func (l *PhotoList) initFrame() {
	if l.FrameSize > len(l.List) {
		l.FrameSize = len(l.List)
	}
	if l.FrameSize == 0 { // Workaround for NewGridWithColumns(0) main window shrink on Windows OS
		l.Frame = container.NewGridWithColumns(1, canvas.NewText("", color.Black))
		return
	}
	for i := l.FramePos; i < l.FramePos+l.FrameSize && i < len(l.List); i++ {
		l.List[i].Img = l.List[i].img(l.FrameSize)
	}
	l.Frame = container.NewGridWithColumns(l.FrameSize)
	for i := 0; i < l.FrameSize && i < len(l.List); i++ {
		l.Frame.Add(l.List[l.FramePos+i].FrameColumn())
	}
}

// open photo folder dialog
func chooseFolder() {
	folder := ""

	fd := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, wMain)
			return
		}
		if list == nil {
			wMain.Close()
			return
		}
		folder = list.Path()
		fyne.CurrentApp().Preferences().SetString("folder", folder)
		pl = newPhotoList(folder)
		MainLayout(pl)
	}, wMain)
	wd, _ := os.Getwd()
	savedLocation := fyne.CurrentApp().Preferences().StringWithFallback("folder", wd)
	locationUri, _ := storage.ListerForURI(storage.NewFileURI(savedLocation))
	fd.SetLocation(locationUri)
	fd.Resize(fyne.NewSize(672, 378))
	fd.Show()
}

func (l *PhotoList) reorder(less func(i int, j int) bool) {
	sort.Slice(l.List, less)
}

func (l *PhotoList) orderByFileNameAsc(i, j int) bool {
	return l.List[i].File < l.List[j].File
}

func (l *PhotoList) orderByFileNameDesc(i, j int) bool {
	return l.List[j].File < l.List[i].File
}

func (l *PhotoList) orderByFileDateAsc(i, j int) bool {
	return l.List[i].Dates[ChoiceFileDate] < l.List[j].Dates[ChoiceFileDate]
}

func (l *PhotoList) orderByFileDateDesc(i, j int) bool {
	return l.List[j].Dates[ChoiceFileDate] < l.List[i].Dates[ChoiceFileDate]
}
