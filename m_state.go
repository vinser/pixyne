package main

import (
	"encoding/json"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
)

type State struct {
	Version           string            `json:"version"`
	WindowSize        fyne.Size         `json:"window_size"`
	Theme             string            `json:"theme"`
	Scale             float32           `json:"scale"`
	Color             string            `json:"color"`
	Simple            bool              `json:"simple_mode"`
	DisplayDateFormat string            `json:"display_date_format"`
	Folder            string            `json:"folder"`
	FramePos          int               `json:"frame_pos"`
	FrameSize         int               `json:"frame_size"`
	ItemPos           int               `json:"item_pos"`
	ListOrderColumn   int               `json:"list_order_column"`
	ListOrder         order             `json:"list_order"`
	List              map[string]*Photo `json:"list"`
}

func (a *App) saveState() {
	a.state.WindowSize = a.topWindow.Canvas().Size()
	a.state.Folder = rootURI.Path()
	for i, v := range a.listColumns {
		if v.Order != natOrder {
			a.state.ListOrderColumn = i
			a.state.ListOrder = v.Order
			break
		}
	}
	stateList := map[string]*Photo{}
	for _, photo := range list {
		if photo.isDroped() || photo.isDated() || photo.isCropped() || photo.isAjusted() {
			stateList[photo.fileURI.Name()] = photo
		}
	}
	a.state.List = stateList
	bytes, _ := json.Marshal(a.state)
	a.Preferences().SetString("state", string(bytes))
}

func (p *Photo) isDroped() bool {
	return p.Drop
}

func (p *Photo) isDated() bool {
	return p.DateUsed != UseExifDate
}

func (p *Photo) isAjusted() bool {
	for i, v := range p.Adjust {
		if v != adjustFiltersDict[i].zero {
			return true
		}
	}
	return false
}

func (p *Photo) isCropped() bool {
	return !p.CropRectangle.Empty()
}

func (a *App) loadState() {
	if state := a.Preferences().String("state"); state != "" {
		if err := json.Unmarshal([]byte(state), &a.state); err == nil {
			if rootURI, err = storage.ListerForURI(storage.NewFileURI(a.state.Folder)); err == nil {
				return
			}
		}
	}
	a.defaultState(true)
}

func (a *App) defaultState(init bool) {
	a.state.Version = a.Metadata().Version
	if init {
		a.state.WindowSize = fyne.NewSize(1280, 720)
		a.state.Theme = DefaultTheme
		a.state.Scale = DefaultScale
		a.state.Color = theme.ColorOrange
		a.state.Simple = true
		a.state.DisplayDateFormat = DefaultDisplayDateFormat
	}
	a.state.List = map[string]*Photo{}
	a.state.FramePos = DefaultListPos
	a.state.FrameSize = DefaultFrameSize
	a.state.ItemPos = DefaultItemPos
	a.state.ListOrderColumn = DefaultListOrderColumn
	a.state.ListOrder = DefaultListOrder
	home, _ := os.UserHomeDir()
	a.state.Folder = home
	uri := storage.NewFileURI(home)
	rootURI, _ = storage.ListerForURI(uri)
}
