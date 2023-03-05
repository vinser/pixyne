package main

import (
	"os"
	"path/filepath"
	"strings"

	_ "image/jpeg"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/vinser/pixyne/pkg/ui"
)

var version, buildTime, target, goversion string

var wMain fyne.Window

var pl *list.PhotoList

func main() {
	a := app.NewWithID("com.github/vinser/photofine")
	t := &ui.Theme{}

	a.Settings().SetTheme(t)

	wMain = a.NewWindow(strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0])))

	wd, _ := os.Getwd()
	pl = list.newPhotoList(a.Preferences().StringWithFallback("folder", wd))
	MainLayout(pl)
	wMain.Resize(fyne.NewSize(1344, 756))
	wMain.CenterOnScreen()
	wMain.SetMaster()
	wMain.Show()
	a.Run()
}
