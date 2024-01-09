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
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"github.com/disintegration/imaging"
	// "github.com/DmitriyVTitov/size"
)

const (
	BackupDirName = "backup"
)

var ScreenWidth, ScreenHeight int

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
}

// get canvas image from file
func GetListImageAt(p *Photo) *canvas.Image {
	frame.StatusText.Set(fmt.Sprintf("Loading...%s - %.2f MB", p.fileURI.Name(), float64(p.byteSize)/1024./1024.))
	m, err := imaging.Open(p.fileURI.Path(), imaging.AutoOrientation(true))
	if err != nil {
		log.Fatal(err)
	}
	// bytesBefore := float64(size.Of(m)) / 1024. / 1024.
	filter := imaging.Box
	nf := normFactor(m)
	if nf > 0 {
		m = imaging.Resize(m, int(float32(ScreenWidth)*nf), 0, filter)
	}
	img := canvas.NewImageFromImage(m)
	img.FillMode = canvas.ImageFillContain
	img.ScaleMode = canvas.ImageScaleFastest
	// bytesAfter := float64(size.Of(m)) / 1024. / 1024.
	// log.Printf("%s scaleDx %f scaleDy %f (%.2f MB)->(%.2f MB)", path.Base(list[pos].fileURI.Name()), scaleDx, scaleDy, bytesBefore, bytesAfter)

	return img
}

const downscaleFactor float32 = 1.0

// screen normalization factor
func normFactor(m image.Image) float32 {
	scaleDx := float32(ScreenWidth) / float32(m.Bounds().Dx())
	scaleDy := float32(ScreenHeight) / float32(m.Bounds().Dy())
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
		// progress.TextFormatter = func() string {
		// 	return fmt.Sprintf("processing %d of %d files", int(progress.Value), int(progress.Max))
		// }
		// progress.Min = 0
		// progress.Max = float64(len(photos))
		// for i, p := range photos {
		for _, p := range photos {
			p.GetPhotoProperties(p.fileURI)
			if len(p.Dates[UseExifDate]) != len(ListDateFormat) {
				p.DateUsed = UseFileDate
			}
			if s, ok := a.state.List[p.fileURI.Name()]; ok {
				p.Drop = s.Drop
				p.DateUsed = s.DateUsed
				p.Dates = s.Dates
				p.CropRectangle = s.CropRectangle
			}
			// progress.SetValue(float64(i + 1))
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
	time.Sleep(1*time.Second - time.Since(start))
}

func (a *App) SavePhotoList(rename bool) {
	modified := false
	for _, p := range list {
		if p.Drop {
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
		switch {
		case p.Drop:
			continue
		case p.DateUsed != UseExifDate || !p.CropRectangle.Empty():
			src = p.fileURI
			if rename {
				dst, _ = storage.Child(rootURI, listDateToFileNameDate(p.Dates[p.DateUsed])+p.fileURI.Extension())
			} else {
				dst, _ = storage.Child(rootURI, p.fileURI.Name())
			}
			dateTime := ""
			if p.DateUsed != UseExifDate {
				dateTime = p.Dates[p.DateUsed]
			}
			err = SaveUpdatedImage(src, dst, dateTime, p.CropRectangle)
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
