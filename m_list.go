package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
	// "github.com/DmitriyVTitov/size"
)

const (
	BackupDirName = "backup"
)

var ScreenWidth, ScreenHeight float32

// File date to use
const (
	UseExifDate = iota
	UseFileDate
	UseEnteredDate
)

// Photo
type Photo struct {
	id            int
	fileURI       fyne.URI
	width         int
	height        int
	byteSize      int64
	Drop          bool            `json:"drop"`
	Dates         [3]string       `json:"dates"`
	DateUsed      int             `json:"date_used"`
	CropRectangle image.Rectangle `json:"crop_rectangle"`
	Adjust        []float64       `json:"adjust"`
}

// get canvas image from file
func GetListImageAt(p *Photo) *canvas.Image {
	// frame.StatusText.Set(fmt.Sprintf("Loading...%s - %.2f MB", p.fileURI.Name(), float64(p.byteSize)/1024./1024.))
	m, err := imaging.Open(p.fileURI.Path(), imaging.AutoOrientation(true))
	if err != nil {
		log.Fatal(err)
	}
	// bytesBefore := float64(size.Of(m)) / 1024. / 1024.
	filter := imaging.Box
	nf := normFactor(m)
	if nf > 0 {
		m = imaging.Resize(m, int(ScreenWidth*nf), 0, filter)
	}
	img := canvas.NewImageFromImage(m)
	img.FillMode = canvas.ImageFillContain
	img.ScaleMode = canvas.ImageScaleFastest
	// bytesAfter := float64(size.Of(m)) / 1024. / 1024.
	// log.Printf("%s scaleDx %f scaleDy %f (%.2f MB)->(%.2f MB)", path.Base(list[pos].fileURI.Name()), scaleDx, scaleDy, bytesBefore, bytesAfter)

	return img
}

const downscaleFactor float32 = 0.75

// screen normalization factor
func normFactor(m image.Image) float32 {
	rowNum := 1
	if a.state.FrameSize > 3 {
		rowNum = 2
	}
	scaleDx := ScreenWidth / float32(m.Bounds().Dx()) / float32(rowNum)
	scaleDy := ScreenHeight / float32(m.Bounds().Dy()) / float32(rowNum)
	if scaleDx <= scaleDy {
		if scaleDx < 1 {
			return scaleDx
		}
	} else {
		if scaleDy < 1 {
			return scaleDy
		}
	}
	return -1
}

// create new PhotoList object for the folder
func (a *App) newPhotoList() {
	start := time.Now()
	photos := []*Photo(nil)
	files, _ := rootURI.List()
	for i, f := range files {
		ext := strings.ToLower(f.Extension())
		if ext == ".jpg" || ext == ".jpeg" {
			p := &Photo{
				id:       i,
				fileURI:  f,
				Drop:     false,
				DateUsed: UseExifDate,
				Dates:    [3]string{},
			}
			photos = append(photos, p)
		}
	}
	if len(photos) > 0 {
		drv := a.Driver()
		dd, ok := drv.(desktop.Driver)
		if !ok {
			log.Fatal("App can run only on desktops!!!")
		}
		w := dd.CreateSplashWindow()
		defer w.Close()
		var logo *canvas.Image
		if a.state.Theme == "dark" {
			logo = canvas.NewImageFromResource(appIconDark)
		} else {
			logo = canvas.NewImageFromResource(appIconLight)
		}
		logo.SetMinSize(fyne.NewSquareSize(100))
		text := canvas.NewText("Reading folder "+rootURI.Path(), theme.PrimaryColor())
		text.TextSize = theme.TextSize() * 1.3
		text.Alignment = fyne.TextAlignCenter
		progress := widget.NewProgressBar()
		progress.TextFormatter = func() string {
			return fmt.Sprintf("processing %d of %d files", int(progress.Value), int(progress.Max))
		}
		progress.Min = 0
		progress.Max = float64(len(photos))
		content := container.NewBorder(container.NewCenter(logo), progress, nil, nil, text)
		w.SetContent(content)
		w.SetPadded(true)
		w.Resize(fyne.NewSize(600, 100))
		w.Show()
		for i, p := range photos {
			p.Adjust = make([]float64, len(adjustFiltersDict))
			for k := range p.Adjust {
				p.Adjust[k] = adjustFiltersDict[k].zero
			}
			p.GetPhotoProperties(p.fileURI)
			if len(p.Dates[UseExifDate]) != len(ListDateFormat) {
				p.DateUsed = UseFileDate
			}
			if s, ok := a.state.List[p.fileURI.Name()]; ok {
				p.Drop = s.Drop
				p.DateUsed = s.DateUsed
				p.Dates = s.Dates
				p.CropRectangle = s.CropRectangle
				p.Adjust = s.Adjust
			}
			progress.SetValue(float64(i + 1))
		}
	}
	list = photos
	if a.state.FramePos >= len(list) {
		a.state.FramePos = len(list) - 1
	}
	if len(list) < a.state.FrameSize {
		a.state.FrameSize = len(list)
	}
	sortList(a.state.ListOrderColumn, a.state.ListOrder)
	time.Sleep(2*time.Second - time.Since(start))
}

func (a *App) SavePhotoList(rename bool) {
	modified := false
	for _, p := range list {
		if p.Drop || p.DateUsed != UseExifDate || !p.CropRectangle.Empty() {
			modified = true
			break
		}
	}
	if !modified && !rename {
		return
	}
	backupURI, err := makeChildFolder(BackupDirName)
	if err != nil {
		dialog.ShowError(err, a.topWindow)
		return
	}
	var src, dst fyne.URI
	for _, p := range list {
		src = p.fileURI
		dst, _ = storage.Child(backupURI, p.fileURI.Name())
		os.Rename(src.Path(), dst.Path())
		p.fileURI = dst
	}
	for _, p := range list {
		src = p.fileURI
		switch {
		case p.isDroped():
		case p.isDated() || p.isCropped() || p.isAdjusted():
			if rename && !isDateSimilarToFileName(p) {
				dst, _ = storage.Child(rootURI, listDateToFileNameDate(p.Dates[p.DateUsed])+p.fileURI.Extension())
			} else {
				dst, _ = storage.Child(rootURI, p.fileURI.Name())
			}
			err = p.SaveUpdatedImage(src, dst)
			if err != nil {
				dialog.ShowError(err, a.topWindow)
			}
		case rename:
			if !isDateSimilarToFileName(p) {
				dst, _ = storage.Child(rootURI, listDateToFileNameDate(p.Dates[UseExifDate])+p.fileURI.Extension())
				storage.Copy(src, dst)
				break
			}
			fallthrough
		default:
			dst, _ = storage.Child(rootURI, p.fileURI.Name())
			os.Rename(src.Path(), dst.Path())
		}
	}
}

func makeChildFolder(name string) (fyne.URI, error) {
	URI, _ := storage.Child(rootURI, name)
	exists, err := storage.Exists(URI)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := storage.CreateListable(URI); err != nil {
			return nil, err
		}
	}
	return URI, nil
}

func isDateSimilarToFileName(p *Photo) bool {
	similar := func(a, b time.Time) bool {
		diff := a.Sub(b)
		if diff < 0 {
			diff = -diff
		}
		return diff < time.Second*2
	}
	nameDate, err := time.Parse(FileNameDateFormat, strings.TrimSuffix(p.fileURI.Name(), p.fileURI.Extension()))
	if err != nil {
		return false
	}
	switch p.DateUsed {
	case UseExifDate:
		exifDate, _ := time.Parse(ListDateFormat, p.Dates[UseExifDate])
		return similar(nameDate, exifDate)
	case UseFileDate:
		fileDate, _ := time.Parse(ListDateFormat, p.Dates[UseFileDate])
		return similar(nameDate, fileDate)
	case UseEnteredDate:
		enteredDate, _ := time.Parse(ListDateFormat, p.Dates[UseEnteredDate])
		return similar(nameDate, enteredDate)
	}
	return true
}
