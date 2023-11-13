package main

import (
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

const (
	BackupDirName = "originals"
	DropDirName   = "dropped"
)

// create new PhotoList object for the folder
func (a *App) newPhotoList() {
	photos := []*Photo(nil)
	files, _ := rootURI.List()
	for i, f := range files {
		ext := strings.ToLower(f.Extension())
		if ext == ".jpg" || ext == ".jpeg" {
			p := &Photo{
				id:       i,
				fileURI:  f,
				Dropped:  false,
				DateUsed: UseExifDate,
				Dates:    [3]string{},
			}
			p.GetPhotoProperties(p.fileURI)
			if len(p.Dates[UseExifDate]) != len(ListDateFormat) {
				p.DateUsed = UseFileDate
			}
			if s, ok := a.state.List[f.Name()]; ok {
				p.Dropped = s.Dropped
				p.DateUsed = s.DateUsed
				p.Dates = s.Dates
			}
			photos = append(photos, p)
		}
	}
	list = photos
	sortList(a.state.ListOrderColumn, a.state.ListOrder)
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
		case p.Dropped:
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
		case p.Dropped:
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
		if p.Dropped {
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
