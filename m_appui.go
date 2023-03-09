package main

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// application App
type App struct {
	name string
	fyne.App
	wMain  fyne.Window
	inited bool

	// Current folder
	folder string
	// List of photos in folder
	*PhotoList
	// Frame with photos
	frame *Frame

	mainContent *fyne.Container

	// Toolbar
	toolBar *widget.Toolbar
	// Toolbar actions
	actAbout       *widget.ToolbarAction
	actSettings    *widget.ToolbarAction
	actOpenFolder  *widget.ToolbarAction
	actSaveList    *widget.ToolbarAction
	actStrechFrame *widget.ToolbarAction
	actShrinkFrame *widget.ToolbarAction
	actToggleView  *widget.ToolbarAction

	// Frame view
	frameView *fyne.Container
	// Frame view scroll buttons
	bottomButtons *fyne.Container
	// Frame scroll buttons
	prevPhotoBtn  *widget.Button
	prevFrameBtn  *widget.Button
	firstPhotoBtn *widget.Button
	nextPhotoBtn  *widget.Button
	nextFrameBtn  *widget.Button
	lastPhotoBtn  *widget.Button

	// List view
	listView       *fyne.Container
	listColumnsNum int
	// List view table column listHeaders
	listHeaders []*ActiveHeader
}

// make main window newLayout
func (a *App) newLayout() {
	a.reorderList(a.Order)
	a.initFrame()
	a.newToolBar()
	a.newFrameView()
	a.newListView()
	a.listView.Hide()
	a.mainContent = container.NewBorder(a.toolBar, nil, nil, nil, container.NewMax(a.frameView, a.listView))
	a.wMain.SetContent(a.mainContent)
}

func (a *App) newToolBar() {
	a.actAbout = widget.NewToolbarAction(theme.HelpIcon(), a.aboutDialog)
	a.actSettings = widget.NewToolbarAction(theme.SettingsIcon(), a.settingsDialog)
	a.actOpenFolder = widget.NewToolbarAction(theme.FolderOpenIcon(), a.openFolderDialog)
	a.actSaveList = widget.NewToolbarAction(theme.DocumentSaveIcon(), a.savePhotoList)
	a.actStrechFrame = widget.NewToolbarAction(theme.ContentAddIcon(), func() { a.resizeFrame(AddColumn) })
	a.actShrinkFrame = widget.NewToolbarAction(theme.ContentRemoveIcon(), func() { a.resizeFrame(RemoveColumn) })
	a.actToggleView = widget.NewToolbarAction(theme.ListIcon(), a.toggleView)

	a.toolBar = widget.NewToolbar()
	a.toolBar.Append(widget.NewToolbarSpacer())
	a.toolBar.Append(a.actToggleView)
	a.toolBar.Append(widget.NewToolbarSeparator())
	a.toolBar.Append(a.actSettings)
	a.toolBar.Append(a.actAbout)

	if len(a.List) > 0 {
		a.toolBar.Prepend(a.actStrechFrame)
		a.toolBar.Prepend(a.actShrinkFrame)
	} else {
		a.toolBar.Prepend(a.actOpenFolder)
	}
}

func (a *App) toggleView() {
	if a.frameView.Hidden {
		a.toolBar.Items = []widget.ToolbarItem{}
		a.toolBar.Append(widget.NewToolbarSpacer())
		a.toolBar.Append(a.actToggleView)
		a.toolBar.Append(widget.NewToolbarSeparator())
		a.toolBar.Append(a.actSettings)
		a.toolBar.Append(a.actAbout)
		if len(a.List) > 0 {
			a.toolBar.Prepend(a.actStrechFrame)
			a.toolBar.Prepend(a.actShrinkFrame)
		} else {
			a.toolBar.Prepend(a.actOpenFolder)
		}

		a.frameView.Show()
		a.listView.Hide()
		a.actToggleView.SetIcon(theme.ListIcon())
	} else {
		a.toolBar.Items = []widget.ToolbarItem{}
		a.toolBar.Append(a.actOpenFolder)
		a.toolBar.Append(a.actSaveList)
		a.toolBar.Append(widget.NewToolbarSpacer())
		a.toolBar.Append(a.actToggleView)
		a.toolBar.Append(widget.NewToolbarSeparator())
		a.toolBar.Append(a.actSettings)
		a.toolBar.Append(a.actAbout)

		a.frameView.Hide()
		a.listView.Show()
		a.actToggleView.SetIcon(theme.GridIcon())
	}
	a.toolBar.Refresh()
}

// open photo folder dialog
func (a *App) openFolderDialog() {
	folder := ""

	d := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, a.wMain)
			return
		}
		if list == nil {
			a.wMain.Close()
			return
		}
		folder = list.Path()
		a.Preferences().SetString("folder", folder)
		a.newPhotoList(folder)
		a.newLayout()
	}, a.wMain)
	wd, _ := os.Getwd()
	location := a.Preferences().StringWithFallback("folder", wd)
	locationUri, _ := storage.ListerForURI(storage.NewFileURI(location))
	d.SetLocation(locationUri)
	d.Resize(fyne.NewSize(672, 378))
	d.Show()
}

func (a *App) settingsDialog() {
	s := NewSettings()
	appearance := widget.NewForm(
		widget.NewFormItem("", s.scalePreviewsRow(a.wMain.Canvas().Scale())),
		widget.NewFormItem("Scale", s.scalesRow()),
		widget.NewFormItem("Main Color", s.colorsRow()),
		widget.NewFormItem("Theme", s.themesRow()),
	)
	dialog.ShowCustom("Appearance", "Ok", appearance, a.wMain)
}
