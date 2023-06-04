package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
)

// File date to use
const (
	UseExifDate = iota
	UseFileDate
	UseEnteredDate
)

// Photo
type Photo struct {
	File     string        `json:"-"`
	Dropped  bool          `json:"dropped"`
	Img      *canvas.Image `json:"-"`
	Dates    [3]string     `json:"dates"`
	DateUsed int           `json:"date_used"`
}

// frame column that contains button with photo image as background and date fix input
func (p *Photo) FrameColumn() *fyne.Container {
	fileLabel := widget.NewLabelWithStyle(filepath.Base(p.File), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	return container.NewBorder(fileLabel, p.dateInput(), nil, nil, p.imgButton())
}

// button with photo image as background
func (p *Photo) imgButton() *fyne.Container {
	var btn *widget.Button
	btn = widget.NewButton(
		"",
		func() {
			if p.Dropped {
				btn.SetText("")
				p.Img.Translucency = 0
				p.Dropped = false
			} else {
				btn.SetText("DROPPED")
				p.Img.Translucency = 0.5
				p.Dropped = true
			}
		},
	)
	if p.Dropped {
		btn.SetText("DROPPED")
		p.Img.Translucency = 0.5
	}
	return container.NewMax(p.Img, btn)
}

// single photo date fix input
func (p *Photo) dateInput() *fyne.Container {
	choices := []string{"EXIF", "File", "Input"}
	d := p.Dates[p.DateUsed]

	eDate := widget.NewEntry()
	eDate.Validator = validation.NewTime(DisplyDateFormat)
	eDate.SetText(d)
	eDate.Disable()

	rgDateChoice := widget.NewRadioGroup(
		choices,
		func(s string) {
			switch s {
			case "EXIF":
				p.Dates[UseEnteredDate] = ""
				p.DateUsed = UseExifDate
				eDate.SetText(p.Dates[p.DateUsed])
				eDate.Disable()
			case "File":
				p.Dates[UseEnteredDate] = ""
				p.DateUsed = UseFileDate
				eDate.SetText(p.Dates[p.DateUsed])
				eDate.Disable()
			case "Input":
				p.DateUsed = UseEnteredDate
				if p.Dates[p.DateUsed] == "" {
					if p.Dates[UseExifDate] == "" {
						p.Dates[p.DateUsed] = p.Dates[UseFileDate]
					} else {
						p.Dates[p.DateUsed] = p.Dates[UseExifDate]
					}
				}
				eDate.SetText(p.Dates[p.DateUsed])
				eDate.Enable()
			}
		})
	rgDateChoice.SetSelected(choices[p.DateUsed])
	rgDateChoice.Horizontal = true

	gr := container.NewVBox(rgDateChoice, eDate)

	return container.NewCenter(gr)
}

// get canvas image from file
func (p *Photo) GetImage(frameSize int) (img *canvas.Image) {
	m, err := imaging.Open(p.File, imaging.AutoOrientation(true))
	if err != nil {
		log.Fatal(err)
	}
	filter := imaging.CatmullRom
	// filter := imaging.Lanczos
	if frameSize > 1 {
		m = imaging.Resize(m, m.Bounds().Dx()/frameSize, 0, filter)
	}
	img = canvas.NewImageFromImage(m)
	img.FillMode = canvas.ImageFillContain
	img.ScaleMode = canvas.ImageScaleFastest
	if runtime.GOOS == "linux" { // !!! TO REMOVE fyne issue #3891 workaround
		switch runtime.GOARCH {
		case "arm64", "arm":
			img.ScaleMode = canvas.ImageScaleSmooth
		}
	}
	return
}

// get file modify date string
func (p *Photo) GetModifyDate() string {
	fi, err := os.Stat(p.File)
	if err != nil {
		return ""
	}
	fileModifyDate := fi.ModTime()
	return fileModifyDate.Format(DisplyDateFormat)
}
