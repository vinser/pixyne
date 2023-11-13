package main

import (
	_ "image/jpeg"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
)

func main() {
	a := &App{App: app.New()}
	a.SetIcon(appIcon)
	a.Settings().SetTheme(&Theme{})
	a.topWindow = a.NewWindow(a.Metadata().Name)
	a.topWindowTitle = binding.NewString()
	a.topWindowTitle.AddListener(binding.NewDataListener(func() {
		path, _ := a.topWindowTitle.Get()
		a.topWindow.SetTitle(a.Metadata().Name + ": " + path)
	}))
	a.topWindow.SetOnClosed(a.saveState)
	a.topWindow.SetMaster()
	a.topWindow.Resize(fyne.NewSize(1344, 756))
	a.topWindow.CenterOnScreen()
	a.topWindow.Show()
	a.loadState()
	a.newPhotoList()
	a.newLayout()

	a.Run()
}
