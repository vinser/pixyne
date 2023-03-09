package main

import (
	"os"
	"path/filepath"
	"strings"

	_ "image/jpeg"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

var version, buildTime, target, goversion string

func main() {
	appID := "com.github/vinser/pixyne"
	a := &App{
		App: app.NewWithID(appID),
	}
	a.Settings().SetTheme(&Theme{})
	a.name = strings.Title(filepath.Base(appID))
	a.wMain = a.NewWindow(a.name)
	a.wMain.SetMaster()
	wd, _ := os.Getwd()
	a.newPhotoList(a.Preferences().StringWithFallback("folder", wd))
	a.newLayout()
	a.wMain.Resize(fyne.NewSize(1344, 756))
	a.wMain.CenterOnScreen()
	a.wMain.Show()
	a.inited = true
	a.Run()
}
