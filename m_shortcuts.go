package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

type ShortCutInfo struct {
	Name    string
	KeyName fyne.KeyName
	handle  func(fyne.Shortcut)
}

func (a *App) Shortcuts() {
	// Control shortcuts
	a.ControlShortCuts = []ShortCutInfo{
		{"Zoom frame", fyne.KeyPlus, func(s fyne.Shortcut) { frame.AddItem() }},
		{"Shrink frame", fyne.KeyMinus, func(s fyne.Shortcut) { frame.RemoveItem() }},
		{"Previous photo", fyne.KeyLeft, func(s fyne.Shortcut) { frame.PrevItem() }},
		{"Next photo", fyne.KeyRight, func(s fyne.Shortcut) { frame.NextItem() }},
		{"Previous frame", fyne.KeyPageUp, func(s fyne.Shortcut) { frame.Prev() }},
		{"Next frame", fyne.KeyPageDown, func(s fyne.Shortcut) { frame.Next() }},
		{"First photo", fyne.KeyHome, func(s fyne.Shortcut) { frame.First() }},
		{"Last photo", fyne.KeyEnd, func(s fyne.Shortcut) { frame.Last() }},
		{"Drop/Undrop photo", fyne.KeyDelete, func(s fyne.Shortcut) { toggleDrop(a.state.FramePos+a.state.ItemPos, frame.Items[a.state.ItemPos]) }},
		{"Open new folder", fyne.KeyO, func(s fyne.Shortcut) { a.openFolderDialog() }},
		{"Save result", fyne.KeyS, func(s fyne.Shortcut) { a.savePhotoListDialog() }},
		{"Exit", fyne.KeyQ, func(s fyne.Shortcut) { a.topWindow.Close() }},
	}
	for _, v := range a.ControlShortCuts {
		sc := &desktop.CustomShortcut{KeyName: v.KeyName, Modifier: fyne.KeyModifierControl}
		a.topWindow.Canvas().AddShortcut(sc, v.handle)
	}
	// Alt shortcuts
	a.AltShortCuts = []ShortCutInfo{
		{"Toggle view", fyne.KeyV, func(s fyne.Shortcut) { a.toggleView() }},
		{"Full screen mode", fyne.KeyF, func(s fyne.Shortcut) { a.toggleFullScreen() }},
		{"Settings", fyne.KeyS, func(s fyne.Shortcut) { a.settingsDialog() }},
		{"About", fyne.KeyA, func(s fyne.Shortcut) { a.aboutDialog() }},
		{"Hotkeys", fyne.KeyK, func(s fyne.Shortcut) { a.hotkeysDialog() }},
		{"Mode simple/advanced", fyne.KeyM, func(s fyne.Shortcut) { a.toggleMode() }},
	}
	for _, v := range a.AltShortCuts {
		sc := &desktop.CustomShortcut{KeyName: v.KeyName, Modifier: fyne.KeyModifierAlt}
		a.topWindow.Canvas().AddShortcut(sc, v.handle)
	}

}
