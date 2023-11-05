package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// List of photos
var list []*Photo

// Frame with photos
var frame *Frame

// application App
type App struct {
	fyne.App
	topWindow fyne.Window

	// Current folder state
	state State

	mainContent *fyne.Container

	// Toolbar
	toolBar *widget.Toolbar
	// Toolbar actions
	actAbout            *widget.ToolbarAction
	actOpenFolder       *widget.ToolbarAction
	actSaveList         *widget.ToolbarAction
	actSettings         *widget.ToolbarAction
	actAddPhoto         *widget.ToolbarAction
	actRemovePhoto      *widget.ToolbarAction
	actToggleView       *widget.ToolbarAction
	actToggleFullScreen *widget.ToolbarAction
	actNoAction         *widget.ToolbarAction

	// Frame view
	frameView *fyne.Container
	// Frame view scroll buttons
	bottomButtons *fyne.Container
	// Frame scroll buttons
	scrollButton []*widget.Button

	// List view
	listView *fyne.Container
	// List table
	listTable *widget.Table
	// List columns settings
	listColumns []*ListColumn
}

// make main window newLayout
func (a *App) newLayout() {
	a.newToolBar()
	a.initFrame()
	a.showFrameToolbar()
	a.newFrameView()
	a.newListView()
	a.listView.Hide()
	a.mainContent = container.NewBorder(a.toolBar, nil, nil, nil, container.NewStack(a.frameView, a.listView))
	a.topWindow.SetContent(a.mainContent)
}

func (a *App) newToolBar() {
	a.actAbout = widget.NewToolbarAction(theme.InfoIcon(), a.aboutDialog)
	a.actOpenFolder = widget.NewToolbarAction(theme.FolderOpenIcon(), a.openFolderDialog)
	a.actSaveList = widget.NewToolbarAction(theme.DocumentSaveIcon(), a.savePhotoList)
	a.actSettings = widget.NewToolbarAction(theme.SettingsIcon(), a.settingsDialog)
	a.actAddPhoto = widget.NewToolbarAction(theme.ContentAddIcon(), func() { a.resizeFrame(AddPhoto) })
	a.actRemovePhoto = widget.NewToolbarAction(theme.ContentRemoveIcon(), func() { a.resizeFrame(RemovePhoto) })
	a.actToggleView = widget.NewToolbarAction(theme.ListIcon(), a.toggleView)
	a.actToggleFullScreen = widget.NewToolbarAction(theme.ViewFullScreenIcon(), a.toggleFullScreen)
	a.actNoAction = widget.NewToolbarAction(theme.NewThemedResource(iconBlank), func() {})

	a.toolBar = widget.NewToolbar()
}

func (a *App) toggleView() {
	if a.frameView.Hidden {
		a.showFrameToolbar()
		a.frameView.Refresh()
		a.frameView.Show()
		a.listView.Hide()
		a.actToggleView.SetIcon(theme.ListIcon())
	} else {
		a.showListToolbar()
		a.frameView.Hide()
		a.listView.Refresh()
		a.listView.Show()
		a.actToggleView.SetIcon(theme.GridIcon())
	}
	a.toolBar.Refresh()
}

func (a *App) toggleFullScreen() {
	if a.topWindow.FullScreen() {
		a.topWindow.SetFullScreen(false)
		a.actToggleFullScreen.SetIcon(theme.ViewFullScreenIcon())
	} else {
		a.topWindow.SetFullScreen(true)
		a.actToggleFullScreen.SetIcon(theme.ViewRestoreIcon())
	}
	a.toolBar.Refresh()
}

func (a *App) showFrameToolbar() {
	a.toolBar.Items = []widget.ToolbarItem{}
	a.toolBar.Append(widget.NewToolbarSpacer())
	a.toolBar.Append(a.actToggleView)
	a.toolBar.Append(a.actToggleFullScreen)
	a.toolBar.Append(widget.NewToolbarSeparator())
	a.toolBar.Append(a.actSettings)
	a.toolBar.Append(a.actAbout)
	if len(list) > 0 {
		if frame.Size < MaxFrameSize && frame.Size < len(list) {
			a.toolBar.Prepend(a.actAddPhoto)
		} else {
			a.toolBar.Prepend(a.actNoAction)
		}
		if frame.Size > MinFrameSize {
			a.toolBar.Prepend(a.actRemovePhoto)
		} else {
			a.toolBar.Prepend(a.actNoAction)
		}

	} else {
		a.toolBar.Prepend(a.actOpenFolder)
	}
}

func (a *App) showListToolbar() {
	a.toolBar.Items = []widget.ToolbarItem{}
	a.toolBar.Append(a.actOpenFolder)
	a.toolBar.Append(a.actSaveList)
	a.toolBar.Append(widget.NewToolbarSpacer())
	a.toolBar.Append(a.actToggleView)
	a.toolBar.Append(a.actToggleFullScreen)
	a.toolBar.Append(widget.NewToolbarSeparator())
	a.toolBar.Append(a.actSettings)
	a.toolBar.Append(a.actAbout)
}

// open photo folder dialog
func (a *App) openFolderDialog() {
	d := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, a.topWindow)
			return
		}
		if list == nil {
			// a.topWindow.Close()
			return
		}
		a.clearState()
		a.state.Folder = list.Path()
		a.newPhotoList()
		a.newLayout()
	}, a.topWindow)
	locationUri, _ := storage.ListerForURI(storage.NewFileURI(a.state.Folder))
	d.SetLocation(locationUri)
	d.Resize(fyne.NewSize(672, 378))
	d.Show()
}

func (a *App) settingsDialog() {
	s := NewSettings()
	settingsForm := widget.NewForm(
		widget.NewFormItem("", s.scalePreviewsRow(a.topWindow.Canvas().Scale())),
		widget.NewFormItem("Scale", s.scalesRow()),
		widget.NewFormItem("Main Color", s.colorsRow()),
		widget.NewFormItem("Theme", s.themesRow()),
		widget.NewFormItem("Date Format", s.datesRow(a)),
	)
	dialog.ShowCustom("Settings", "Close", settingsForm, a.topWindow)
}
