package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// application App
type App struct {
	fyne.App
	topWindow fyne.Window

	// Current folder state
	state State
	// List of photos in folder
	*PhotoList
	// Frame with photos
	frame *Frame

	mainContent *fyne.Container

	// Toolbar
	toolBar *widget.Toolbar
	// Toolbar actions
	actAbout       *widget.ToolbarAction
	actOpenFolder  *widget.ToolbarAction
	actSaveList    *widget.ToolbarAction
	actSettings    *widget.ToolbarAction
	actStrechFrame *widget.ToolbarAction
	actShrinkFrame *widget.ToolbarAction
	actToggleView  *widget.ToolbarAction

	// Frame view
	frameView *fyne.Container
	// Frame view scroll buttons
	bottomButtons *fyne.Container
	// Frame scroll buttons
	scrollButton []*widget.Button

	// List view
	listView *fyne.Container
	// List headers settings
	listHeaders    []*ActiveHeader
	listColumnsNum int
	// List table
	listTable *widget.Table
}

// make main window newLayout
func (a *App) newLayout() {
	a.reorderList(a.Order)
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
	a.actStrechFrame = widget.NewToolbarAction(theme.ContentAddIcon(), func() { a.resizeFrame(AddColumn) })
	a.actShrinkFrame = widget.NewToolbarAction(theme.ContentRemoveIcon(), func() { a.resizeFrame(RemoveColumn) })
	a.actToggleView = widget.NewToolbarAction(theme.ListIcon(), a.toggleView)

	a.toolBar = widget.NewToolbar()
}

func (a *App) toggleView() {
	if a.frameView.Hidden {
		a.showFrameToolbar()
		a.frameView.Show()
		a.listView.Hide()
		a.actToggleView.SetIcon(theme.ListIcon())
	} else {
		a.showListToolbar()
		a.frameView.Hide()
		a.listView.Show()
		a.actToggleView.SetIcon(theme.GridIcon())
	}
	a.toolBar.Refresh()
}

func (a *App) showFrameToolbar() {
	a.toolBar.Items = []widget.ToolbarItem{}
	a.toolBar.Append(widget.NewToolbarSpacer())
	a.toolBar.Append(a.actToggleView)
	a.toolBar.Append(widget.NewToolbarSeparator())
	a.toolBar.Append(a.actSettings)
	a.toolBar.Append(a.actAbout)
	if len(a.List) > 0 {
		if a.frame.Size < MaxFrameSize && a.frame.Size < len(a.List) {
			a.toolBar.Prepend(a.actStrechFrame)
		}
		if a.frame.Size > MinFrameSize {
			a.toolBar.Prepend(a.actShrinkFrame)
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
			a.topWindow.Close()
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
	appearance := widget.NewForm(
		widget.NewFormItem("", s.scalePreviewsRow(a.topWindow.Canvas().Scale())),
		widget.NewFormItem("Scale", s.scalesRow()),
		widget.NewFormItem("Main Color", s.colorsRow()),
		widget.NewFormItem("Theme", s.themesRow()),
	)
	dialog.ShowCustom("Appearance", "Close", appearance, a.topWindow)
}
