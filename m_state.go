package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type State struct {
	Folder           string `json:"folder"`
	Frame            `json:"frame"`
	List             map[string]*Photo `json:"list"`
	DisplyDateFormat string            `json:"disply_date_format"`
}

func (a *App) saveState() {
	list := map[string]*Photo{}
	for _, p := range list {
		if p.Dropped || p.DateUsed != UseExifDate {
			list[filepath.Base(p.File)] = p
		}
	}
	a.state.List = list
	a.state.Pos = frame.Pos
	a.state.Size = frame.Size
	a.state.DisplyDateFormat = DisplayDateFormat
	bytes, _ := json.Marshal(a.state)
	a.Preferences().SetString("state", string(bytes))
}

func (a *App) loadState() {
	if state := a.Preferences().String("state"); state != "" {
		if err := json.Unmarshal([]byte(state), &a.state); err == nil {
			DisplayDateFormat = a.state.DisplyDateFormat
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
	// a.state.DisplyDateFormat = InitDisplayDateFormat
	// a.Preferences().RemoveValue("state")
}
