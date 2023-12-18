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
	a.SetIcon(appIcon)
	a.Settings().SetTheme(&Theme{})
	a.topWindow = a.NewWindow(a.Metadata().Name)
	a.topWindowTitle = binding.NewString()
	a.topWindowTitle.AddListener(binding.NewDataListener(func() {
		path, _ := a.topWindowTitle.Get()
		a.topWindow.SetTitle(a.Metadata().Name + ": " + path)
	}))
	a.loadState()
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
	logo := canvas.NewImageFromResource(appIcon)
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

	W4K = int(w.Canvas().Size().Width)
	H4K = int(w.Canvas().Size().Height)
	initCh <- struct{}{}
	<-respCh
}

func mainJob() {
	<-initCh
	// log.Printf("Screen: %d x %d", W4K, H4K)
	a.topWindow.SetOnClosed(a.saveState)
	a.topWindow.SetMaster()
	a.topWindow.Resize(fyne.NewSize(1344, 756))
	a.topWindow.CenterOnScreen()
	a.newLayout()
	a.topWindow.Show()
	respCh <- struct{}{}
	time.Sleep(time.Second * 1)
	// log.Printf("Window: %f x %f", a.topWindow.Content().Size().Width, a.topWindow.Content().Size().Height)
}
