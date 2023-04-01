package main

import (
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
)

// File date choices
const (
	ChoiceExifDate = iota
	ChoiceFileDate
	ChoiceEnteredDate
)
const DateFormat = "2006:01:02 03:04:05"

// Photo
type Photo struct {
	File       string
	Droped     bool
	Img        *canvas.Image
	Dates      [3]string
	DateChoice int
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
			if p.Droped {
				btn.SetText("")
				p.Img.Translucency = 0
				p.Droped = false
			} else {
				btn.SetText("DROPPED")
				p.Img.Translucency = 0.5
				p.Droped = true
			}
		},
	)
	if p.Droped {
		btn.SetText("DROPPED")
		p.Img.Translucency = 0.5
	}
	return container.NewMax(p.Img, btn)
}

// single photo date fix input
func (p *Photo) dateInput() *fyne.Container {
	choices := []string{"EXIF", "File", "Input"}
	d := p.Dates[p.DateChoice]

	eDate := widget.NewEntry()
	eDate.Validator = validation.NewTime(DateFormat)
	eDate.SetText(d)
	eDate.Disable()

	rgDateChoice := widget.NewRadioGroup(
		choices,
		func(s string) {
			switch s {
			case "EXIF":
				p.Dates[ChoiceEnteredDate] = ""
				p.DateChoice = ChoiceExifDate
				eDate.SetText(p.Dates[p.DateChoice])
				eDate.Disable()
			case "File":
				p.Dates[ChoiceEnteredDate] = ""
				p.DateChoice = ChoiceFileDate
				eDate.SetText(p.Dates[p.DateChoice])
				eDate.Disable()
			case "Input":
				p.DateChoice = ChoiceEnteredDate
				if p.Dates[p.DateChoice] == "" {
					if p.Dates[ChoiceExifDate] == "" {
						p.Dates[p.DateChoice] = p.Dates[ChoiceFileDate]
					} else {
						p.Dates[p.DateChoice] = p.Dates[ChoiceExifDate]
					}
				}
				eDate.SetText(p.Dates[p.DateChoice])
				eDate.Enable()
			}
		})
	rgDateChoice.SetSelected(choices[p.DateChoice])
	rgDateChoice.Horizontal = true

	gr := container.NewVBox(rgDateChoice, eDate)

	return container.NewCenter(gr)
}

// get canvas image from file
func (p *Photo) GetImage(scale int) (img *canvas.Image) {
	m, err := imaging.Open(p.File, imaging.AutoOrientation(true))
	if err != nil {
		log.Fatal(err)
	}
	if scale > 1 {
		width := (m.Bounds().Max.X - m.Bounds().Min.X) / scale
		// m = imaging.Resize(m, width, 0, imaging.Lanczos)
		m = imaging.Resize(m, width, 0, imaging.CatmullRom)
	}
	img = canvas.NewImageFromImage(m)
	img.FillMode = canvas.ImageFillContain
	// img.ScaleMode = canvas.ImageScalePixels
	img.ScaleMode = canvas.ImageScaleFastest
	return
}

// get file modify date string
func (p *Photo) GetModifyDate() string {
	fi, err := os.Stat(p.File)
	if err != nil {
		return ""
	}
	fileModifyDate := fi.ModTime()
	return fileModifyDate.Format(DateFormat)
}
