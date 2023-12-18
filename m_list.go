package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"github.com/disintegration/imaging"
	// "github.com/DmitriyVTitov/size"
)

const (
	BackupDirName = "originals"
	DropDirName   = "dropped"
)

var W4K = 3840 * 3 / 4
var H4K = 2160 * 3 / 4

// File date to use
const (
	UseExifDate = iota
	UseFileDate
	UseEnteredDate
)

// Photo
type Photo struct {
	id       int
	fileURI  fyne.URI
	width    int
	height   int
	byteSize int64
	Drop     bool      `json:"drop"`
	Dates    [3]string `json:"dates"`
	DateUsed int       `json:"date_used"`
}

// get canvas image from file
func GetListImageAt(pos int) *canvas.Image {
	m, err := imaging.Open(list[pos].fileURI.Path(), imaging.AutoOrientation(true))
	if err != nil {
		log.Fatal(err)
	}
	// bytesBefore := float64(size.Of(m)) / 1024. / 1024.
	filter := imaging.Box
	scaleDx := float64(W4K) / float64(m.Bounds().Dx()) / float64(frame.Size)
	scaleDy := float64(H4K) / float64(m.Bounds().Dy()) / float64(frame.Size)
	if scaleDx < scaleDy {
		if scaleDx < 1 {
			m = imaging.Resize(m, int(float64(m.Bounds().Dx())*scaleDx), 0, filter)
		}

	} else {
		if scaleDy < 1 {
			m = imaging.Resize(m, 0, int(float64(m.Bounds().Dy())*scaleDy), filter)
		}
	}
	img := canvas.NewImageFromImage(m)
	img.FillMode = canvas.ImageFillContain
	img.ScaleMode = canvas.ImageScaleFastest
	// bytesAfter := float64(size.Of(m)) / 1024. / 1024.
	// log.Printf("%s scaleDx %f scaleDy %f (%.2f MB)->(%.2f MB)", path.Base(list[pos].fileURI.Name()), scaleDx, scaleDy, bytesBefore, bytesAfter)

	return img
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
		progress.TextFormatter = func() string {
			return fmt.Sprintf("processing %d of %d files", int(progress.Value), int(progress.Max))
		}
		progress.Min = 0
		progress.Max = float64(len(photos))
		for i, p := range photos {
			p.GetPhotoProperties(p.fileURI)
			if len(p.Dates[UseExifDate]) != len(ListDateFormat) {
				p.DateUsed = UseFileDate
			}
			if s, ok := a.state.List[p.fileURI.Name()]; ok {
				p.Drop = s.Drop
				p.DateUsed = s.DateUsed
				p.Dates = s.Dates
			}
			progress.SetValue(float64(i + 1))
		}
	}
	list = photos
	sortList(a.state.ListOrderColumn, a.state.ListOrder)
	time.Sleep(1*time.Second - time.Since(start))
}

func (a *App) SavePhotoList(rename bool) {
	backupURI, err := makeBackupDir()
	if err != nil {
		dialog.ShowError(err, a.topWindow)
		return
	}
	dropURI, err := makeDropDir()
	if err != nil {
		dialog.ShowError(err, a.topWindow)
		return
	}
	var src, dst fyne.URI
	for _, p := range list {
		switch {
		case p.Drop:
			src = p.fileURI
			dst, _ = storage.Child(dropURI, p.fileURI.Name())
			os.Rename(src.Path(), dst.Path())
		default:
			src = p.fileURI
			dst, _ = storage.Child(backupURI, p.fileURI.Name())
			os.Rename(src.Path(), dst.Path())
		}
		p.fileURI = dst
	}
	for _, p := range list {
		switch {
		case p.Drop:
			continue
		case p.DateUsed != UseExifDate:
			src = p.fileURI
			if rename {
				dst, _ = storage.Child(rootURI, listDateToFileNameDate(p.Dates[p.DateUsed])+p.fileURI.Extension())
			} else {
				dst, _ = storage.Child(rootURI, p.fileURI.Name())
			}
			err = UpdateExif(src, dst, p.Dates[p.DateUsed])
			if err != nil {
				dialog.ShowError(err, a.topWindow)
			}
			continue
		case rename:
			src = p.fileURI
			dst, _ = storage.Child(rootURI, listDateToFileNameDate(p.Dates[p.DateUsed])+p.fileURI.Extension())
			if p.fileURI.Name() == dst.Name() {
				os.Rename(src.Path(), dst.Path())
			} else {
				storage.Copy(src, dst)
			}
		default:
			src = p.fileURI
			dst, _ = storage.Child(rootURI, p.fileURI.Name())
			os.Rename(src.Path(), dst.Path())
		}
		p.fileURI = dst
	}
}

func makeBackupDir() (fyne.URI, error) {
	return makeSubfolder(BackupDirName)
}

func makeDropDir() (fyne.URI, error) {
	for _, p := range list {
		if p.Drop {
			return makeSubfolder(DropDirName)
		}
	}
	return nil, nil
}

func makeSubfolder(name string) (fyne.URI, error) {
	URI, _ := storage.Child(rootURI, name)
	yes, err := storage.Exists(URI)
	if err != nil {
		return nil, err
	}
	if !yes {
		if err := storage.CreateListable(URI); err != nil {
			return nil, err
		}
	}
	return URI, nil
}
