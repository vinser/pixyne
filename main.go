package main

import (
	"image/color"
	_ "image/jpeg"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
)

var initCh chan struct{}
var respCh chan struct{}

func main() {
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
	if ScreenWidth == 0 {
		go firstRun()
	} else {
		standardRun()
	}

	a.Run()
}

func standardRun() {
	a.topWindowTitle.Set(rootURI.Path())
	a.topWindow.SetOnClosed(a.saveState)
	a.topWindow.SetMaster()
	a.newPhotoList()
	topFit := downscaleFactor / fyne.CurrentApp().Settings().Scale()
	a.topWindow.Resize(fyne.NewSize(float32(ScreenWidth)*topFit, float32(ScreenHeight)*topFit))
	a.newLayout()
	a.topWindow.CenterOnScreen()
	a.topWindow.Show()
}
func firstRun() {

	a.topWindowTitle.Set(rootURI.Path())
	a.topWindow.SetOnClosed(a.saveState)
	a.topWindow.SetMaster()
	initCh = make(chan struct{})
	respCh = make(chan struct{})
	go initScreenRoutine()
	<-initCh
	a.newPhotoList()
	topFit := downscaleFactor / fyne.CurrentApp().Settings().Scale()
	a.topWindow.Resize(fyne.NewSize(float32(ScreenWidth)*topFit, float32(ScreenHeight)*topFit))
	a.newLayout()
	a.topWindow.CenterOnScreen()
	a.topWindow.Show()
	respCh <- struct{}{}
}

func initScreenRoutine() {
	w := a.NewWindow("Probe")
	defer func() {
		w.Close()
		close(initCh)
		close(respCh)
	}()
	w.SetFullScreen(true)
	w.Show()
	var logo *canvas.Image
	if a.state.Theme == "dark" {
		logo = canvas.NewImageFromResource(appIconDark)
	} else {
		logo = canvas.NewImageFromResource(appIconLight)
	}
	logo.SetMinSize(fyne.NewSquareSize(400))
	emptyObj := canvas.NewRectangle(color.Transparent)
	text := canvas.NewText("Optimizing Pixyne...", theme.PrimaryColor())
	text.TextSize = theme.TextSize() * 3
	text.Alignment = fyne.TextAlignCenter
	content := container.NewGridWithColumns(3,
		emptyObj,
		container.NewGridWithRows(3,
			emptyObj,
			container.NewCenter(logo),
			container.NewGridWithRows(3, text),
		),
		emptyObj)
	w.SetContent(content)
	time.Sleep(time.Second * 3)

	ScreenWidth = int(w.Canvas().Size().Width * fyne.CurrentApp().Settings().Scale())
	ScreenHeight = int(w.Canvas().Size().Height * fyne.CurrentApp().Settings().Scale())
	initCh <- struct{}{}
	<-respCh
}
