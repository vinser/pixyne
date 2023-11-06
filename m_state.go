package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type State struct {
	Folder    string            `json:"folder"`
	FramePos  int               `json:"frame_pos"`
	FrameSize int               `json:"frame_size"`
	List      map[string]*Photo `json:"list"`
}

func (a *App) saveState() {
	a.Preferences().SetBool("simple_mode", a.simpleMode)
	a.Preferences().SetString("display_date_format", DisplayDateFormat)
	sl := map[string]*Photo{}
	for _, p := range list {
		if p.Dropped || p.DateUsed != UseExifDate {
			sl[filepath.Base(p.File)] = p
		}
	}
	a.state.List = sl
	a.state.FramePos = frame.Pos
	a.state.FrameSize = frame.Size
	bytes, _ := json.Marshal(a.state)
	a.Preferences().SetString("state", string(bytes))
}

func (a *App) loadState() {
	DisplayDateFormat = a.Preferences().StringWithFallback("display_date_format", DefaultDisplayDateFormat)
	a.simpleMode = a.Preferences().BoolWithFallback("simple_mode", false)
	if state := a.Preferences().String("state"); state != "" {
		if err := json.Unmarshal([]byte(state), &a.state); err == nil {
			if _, err = os.Stat(a.state.Folder); err == nil {
				return
			}
		}
	}
	wd, _ := os.Getwd()
	a.state.Folder = wd
}

func (a *App) clearState() {
	a.state.List = map[string]*Photo{}
	a.state.FramePos = InitListPos
	a.state.FrameSize = InitFrameSize
}
