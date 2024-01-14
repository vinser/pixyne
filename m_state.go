package main

import (
	"encoding/json"
	"os"

	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
)

type State struct {
	Version           string            `json:"version"`
	Theme             string            `json:"theme"`
	Scale             float32           `json:"scale"`
	Color             string            `json:"color"`
	Simple            bool              `json:"simple_mode"`
	DisplayDateFormat string            `json:"display_date_format"`
	Folder            string            `json:"folder"`
	FramePos          int               `json:"frame_pos"`
	FrameSize         int               `json:"frame_size"`
	ListOrderColumn   int               `json:"list_order_column"`
	ListOrder         order             `json:"list_order"`
	List              map[string]*Photo `json:"list"`
}

func (a *App) saveState() {
	a.Preferences().SetInt("screen_width", ScreenWidth)
	a.Preferences().SetInt("screen_height", ScreenHeight)
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
		if photo.Drop || photo.DateUsed != UseExifDate || !photo.CropRectangle.Empty() {
			stateList[photo.fileURI.Name()] = photo
		}
	}
	a.state.List = stateList
	bytes, _ := json.Marshal(a.state)
	a.Preferences().SetString("state", string(bytes))
}

func (a *App) loadState() {
	ScreenWidth = a.Preferences().IntWithFallback("screen_width", 0)
	ScreenHeight = a.Preferences().IntWithFallback("screen_height", 0)
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
		a.state.Theme = DefaultTheme
		a.state.Scale = DefaultScale
		a.state.Color = theme.ColorOrange
		a.state.Simple = true
		a.state.DisplayDateFormat = DefaultDisplayDateFormat
	}
	a.state.List = map[string]*Photo{}
	a.state.FramePos = DefaultListPos
	a.state.FrameSize = DefaultFrameSize
	a.state.ListOrderColumn = DefaultListOrderColumn
	a.state.ListOrder = DefaultListOrder
	home, _ := os.UserHomeDir()
	a.state.Folder = home
	uri := storage.NewFileURI(home)
	rootURI, _ = storage.ListerForURI(uri)
}
