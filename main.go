package main

import (
	_ "image/jpeg"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func main() {
	ScreenWidth, ScreenHeight = getDisplayResolution()

	a = &App{App: app.New()}
	a.loadState()
	if a.state.Theme == "dark" {
		a.SetIcon(appIconDark)
	} else {
		a.SetIcon(appIconLight)
	}
	a.topWindow = a.NewWindow(a.Metadata().Name)
	a.topWindowTitle = binding.NewString()
	a.topWindowTitle.AddListener(binding.NewDataListener(func() {
		path, _ := a.topWindowTitle.Get()
		a.topWindow.SetTitle(a.Metadata().Name + ": " + path)
	}))
	a.Shortcuts()
	a.Settings().SetTheme(&Theme{})
	a.topWindowTitle.Set(rootURI.Path())
	a.topWindow.SetOnClosed(a.saveState)
	a.topWindow.SetMaster()
	a.topWindow.Resize(a.state.WindowSize)
	go func() {
		a.newPhotoList()
		a.newLayout()
	}()
	a.topWindow.CenterOnScreen()
	a.topWindow.Show()
	a.Run()
}

func getDisplayResolution() (float32, float32) {
	glfw.Init()
	defer glfw.Terminate()
	monitor := glfw.GetPrimaryMonitor()
	return float32(monitor.GetVideoMode().Width), float32(monitor.GetVideoMode().Height)
}
