package main

import (
	"os"

	_ "image/jpeg"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	a := &App{App: app.New()}
	a.SetIcon(appIcon)
	a.Settings().SetTheme(&Theme{})
	a.topWindow = a.NewWindow(a.Metadata().Name)
	wd, _ := os.Getwd()
	a.newPhotoList(a.Preferences().StringWithFallback("folder", wd))
	a.newLayout()
	a.topWindow.SetMaster()
	a.topWindow.Resize(fyne.NewSize(1344, 756))
	a.topWindow.CenterOnScreen()
	a.topWindow.Show()
	a.Run()
}
