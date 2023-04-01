package main

import (
	"os"

	_ "image/jpeg"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	a := &App{App: app.New()}
	a.Settings().SetTheme(&Theme{})
	a.wMain = a.NewWindow(a.Metadata().Custom["AppName"])
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
