package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Photolist
type PhotoList struct {
	Folder string
	List   []*Photo
	Order  func(i, j int) bool
}

// create new PhotoList object for the folder
func (a *App) newPhotoList() {
	files, err := os.ReadDir(a.state.Folder)
	if err != nil {
		log.Fatalf("Can't list photo files from folder \"%s\". Error: %v\n", a.state.Folder, err)
	}
	photos := []*Photo(nil)
	for _, f := range files {
		fName := strings.ToLower(f.Name())
		if strings.HasSuffix(fName, ".jpg") || strings.HasSuffix(fName, ".jpeg") {
			p := &Photo{
				File:     filepath.Join(a.state.Folder, f.Name()),
				Dropped:  false,
				DateUsed: UseExifDate,
				Dates:    [3]string{},
			}
			p.Dates[UseExifDate] = GetExifDate(p.File)
			p.Dates[UseFileDate] = p.GetModifyDate()
			if len(p.Dates[UseExifDate]) != len(DisplyDateFormat) {
				p.DateUsed = UseFileDate
			}
			if s, ok := a.state.List[fName]; ok {
				p.Dropped = s.Dropped
				p.DateUsed = s.DateUsed
				p.Dates = s.Dates
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
				dropDirName := filepath.Join(a.state.Folder, "dropped")
				backupDirOk := false
				backupDirName := filepath.Join(a.state.Folder, "original")
				for _, p := range a.List {
					if p.Dropped {
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
					if p.DateUsed != UseExifDate {
						// backup original file and make file copy with modified exif
						if !backupDirOk {
							err := os.Mkdir(backupDirName, 0775)
							if err != nil && !errors.Is(err, fs.ErrExist) {
								dialog.ShowError(err, a.topWindow)
							}
						}
						if UpdateExifDate(p.File, backupDirName, p.Dates[p.DateUsed]) == nil {
							if dateFileNames {
								os.Rename(p.File, fileNameToDate(p.File, p.Dates[p.DateUsed]))
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
						copyPhoto(p.File, filepath.Join(backupDirName, filepath.Base(p.File)))
						os.Rename(p.File, fileNameToDate(p.File, p.Dates[p.DateUsed]))
					}
				}
				a.clearState()
			}
		},
		a.topWindow)
	d.Show()
}

type State struct {
	Folder string `json:"folder"`
	Frame  `json:"frame"`
	List   map[string]*Photo `json:"list"`
}

func (a *App) saveState() {
	list := map[string]*Photo{}
	for _, p := range a.List {
		if p.Dropped || p.DateUsed != UseExifDate {
			list[filepath.Base(p.File)] = p
		}
	}
	a.state.List = list
	a.state.Pos = a.frame.Pos
	a.state.Size = a.frame.Size
	bytes, _ := json.Marshal(a.state)
	a.Preferences().SetString("state", string(bytes))
}

func (a *App) loadState() {
	if state := a.Preferences().String("state"); state != "" {
		if err := json.Unmarshal([]byte(state), &a.state); err == nil {
			return
		}
	}
	wd, _ := os.Getwd()
	a.state.Folder = wd
}

func (a *App) clearState() {
	a.state.List = map[string]*Photo{}
	a.state.Pos = InitListPos
	a.state.Size = InitFrameSize
	// a.Preferences().RemoveValue("state")
}
