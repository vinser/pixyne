package main

import (
	"errors"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// List folder URI
var rootURI fyne.ListableURI

// List of photos
var list []*Photo

// Frame with photos
var frame *Frame

var loadProgress *widget.ProgressBarInfinite

// application App
type App struct {
	fyne.App
	topWindow      fyne.Window
	topWindowTitle binding.String

	// Current folder state
	state State

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

	// List view
	listView *fyne.Container
	// List table
	listTable *widget.Table
	// List columns settings
	listColumns []*ListColumn
}

var a *App

// make main window newLayout
func (a *App) newLayout() {
	loadProgress = widget.NewProgressBarInfinite()
	loadProgress.Hide()
	a.newToolBar()
	a.newFrame()
	a.showFrameToolbar()
	a.frameView = frame.newFrameView()
	a.listView = a.newListView()
	a.listView.Hide()
	top := container.NewStack(a.toolBar, container.NewGridWithColumns(3, widget.NewLabel(""), loadProgress))
	a.topWindow.SetContent(container.NewBorder(top, nil, nil, nil, container.NewStack(a.frameView, a.listView)))
}

func (a *App) newToolBar() {
	a.actAbout = widget.NewToolbarAction(theme.InfoIcon(), a.aboutDialog)
	a.actOpenFolder = widget.NewToolbarAction(theme.FolderOpenIcon(), a.openFolderDialog)
	a.actSaveList = widget.NewToolbarAction(theme.DocumentSaveIcon(), a.savePhotoListDialog)
	a.actSettings = widget.NewToolbarAction(theme.SettingsIcon(), a.settingsDialog)
	a.actAddPhoto = widget.NewToolbarAction(theme.ContentAddIcon(), func() { frame.AddItem() })
	a.actRemovePhoto = widget.NewToolbarAction(theme.ContentRemoveIcon(), func() { frame.RemoveItem() })
	a.actToggleView = widget.NewToolbarAction(theme.ListIcon(), a.toggleView)
	a.actToggleFullScreen = widget.NewToolbarAction(theme.ViewFullScreenIcon(), a.toggleFullScreen)
	a.actNoAction = widget.NewToolbarAction(theme.NewThemedResource(iconBlank), func() {})

	a.toolBar = widget.NewToolbar()
}

func (a *App) toggleView() {
	if a.frameView.Hidden {
		a.showFrameToolbar()
		a.frameView.Show()
		a.listView.Hide()
		a.actToggleView.SetIcon(theme.ListIcon())
		frame.At(a.state.FramePos)
		a.frameView.Refresh()
	} else {
		a.showListToolbar()
		a.listTable.ScrollTo(widget.TableCellID{Col: 0, Row: a.state.FramePos})
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
		if a.state.FrameSize < MaxFrameSize && a.state.FrameSize < len(list) {
			a.toolBar.Prepend(a.actAddPhoto)
		} else {
			a.toolBar.Prepend(a.actNoAction)
		}
		if a.state.FrameSize > MinFrameSize {
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
			return
		}
		if list.Scheme() != "file" {
			dialog.ShowError(errors.New("only local files are supported"), a.topWindow)
			return
		}
		a.defaultState()
		rootURI = list
		a.topWindowTitle.Set(rootURI.Path())
		a.newPhotoList()
		a.newLayout()
	}, a.topWindow)
	d.SetLocation(rootURI)
	d.Resize(fyne.NewSize(672, 378))
	d.Show()
}

// Save choosed photos:
// 1. move dropped photo to droppped folder
// 2. update exif dates with file modify date or input date
func (a *App) savePhotoListDialog() {
	renameFiles := false
	datedFileFormat := time.Now().Format(FileNameDateFormat)
	content := container.NewVBox(
		widget.NewLabel("Ready to save changes?"),
		widget.NewCheck("Rename files to date taken format "+datedFileFormat, func(b bool) { renameFiles = b }),
	)
	d := dialog.NewCustomConfirm(
		"Save changes",
		"Proceed",
		"Cancel",
		content,
		func(b bool) {
			if b {
				a.SavePhotoList(renameFiles)
				a.defaultState()
				a.newPhotoList()
				a.newLayout()
			}
		},
		a.topWindow)
	d.Show()
}

func (a *App) settingsDialog() {
	s := NewSettings()
	settingsForm := widget.NewForm(
		widget.NewFormItem("", s.scalePreviewsRow(a.topWindow.Canvas().Scale())),
		widget.NewFormItem("Scale", s.scalesRow()),
		widget.NewFormItem("Main Color", s.colorsRow()),
		widget.NewFormItem("Theme", s.themesRow()),
		widget.NewFormItem("Mode", s.modeRow(a)),
		widget.NewFormItem("Date Format", s.datesRow(a)),
	)
	dialog.ShowCustom("Settings", "Close", settingsForm, a.topWindow)
}
