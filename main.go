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
	"fyne.io/fyne/v2/widget"
)

var initCh = make(chan struct{})
var respCh = make(chan struct{})
var progress *widget.ProgressBar

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
	go initScreen()
	go mainJob()
	a.Run()
}

func initScreen() {
	w := a.NewWindow("Probe")
	defer w.Close()
	// w := a.w
	w.SetFullScreen(true)
	w.Show()
	time.Sleep(time.Second * 1)
	var logo *canvas.Image
	if a.state.Theme == "dark" {
		logo = canvas.NewImageFromResource(appIconDark)
	} else {
		logo = canvas.NewImageFromResource(appIconLight)
	}
	logo.SetMinSize(fyne.NewSquareSize(400))
	emptyObj := canvas.NewRectangle(color.Transparent)
	progress = widget.NewProgressBar()
	content := container.NewGridWithColumns(3,
		emptyObj,
		container.NewGridWithRows(3,
			emptyObj,
			container.NewCenter(logo),
			container.NewVBox(
				widget.NewLabel("Processing folder with photos"),
				progress)),
		emptyObj)
	w.SetContent(content)
	a.newPhotoList()

	ScreenWidth = int(w.Canvas().Size().Width)
	ScreenHeight = int(w.Canvas().Size().Height)
	initCh <- struct{}{}
	<-respCh
}

func mainJob() {
	<-initCh
	// log.Printf("Screen: %d x %d", ScreenWidth, ScreenHeight)
	a.topWindowTitle.Set(rootURI.Path())
	a.topWindow.SetOnClosed(a.saveState)
	a.topWindow.SetMaster()
	a.topWindow.Resize(fyne.NewSize(float32(ScreenWidth)*0.6, float32(ScreenHeight)*0.6))
	a.topWindow.CenterOnScreen()
	a.newLayout()
	a.topWindow.Show()
	respCh <- struct{}{}
	time.Sleep(time.Second * 1)
	// log.Printf("Window: %f x %f", a.topWindow.Content().Size().Width, a.topWindow.Content().Size().Height)
}
