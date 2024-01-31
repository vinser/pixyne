package main

import (
	"errors"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// show info about app
func (a *App) aboutDialog() {
	a.disableShortcuts()
	var logo *canvas.Image
	if a.state.Theme == "dark" {
		logo = canvas.NewImageFromResource(appIconDark)
	} else {
		logo = canvas.NewImageFromResource(appIconLight)
	}
	logo.FillMode = canvas.ImageFillOriginal
	logoRow := container.NewBorder(nil, nil, logo, nil)
	infoRow := widget.NewRichTextFromMarkdown(`
## Pixyne - photo picker
---
Use [Pixyne](http://vinser.github.io/pixyne) to quickly review your photo folder and safely delete bad and similar shots.

You may also set EXIF shooting date to the file date or to a manually entered date.

---`)
	var buildVersion, buildFor, buildTime, goVersion, versionLine, buildLine string
	if buildVersion = a.Metadata().Version; buildVersion == "" {
		buildVersion = "selfcrafted"
	}
	versionLine = "Version: " + buildVersion

	if buildFor = a.Metadata().Custom["BuildForOS"]; buildFor != "" {
		buildLine = fmt.Sprintf("Build for: %s ", buildFor)
	}
	if buildTime = a.Metadata().Custom["BuildTime"]; buildTime != "" {
		buildLine = buildLine + fmt.Sprintf(" | Build time: %s ", buildTime)
	}
	if goVersion = a.Metadata().Custom["GoVersion"]; goVersion != "" {
		buildLine = buildLine + fmt.Sprintf(" | Go version: %s ", goVersion)
	}

	tecRow := widget.NewRichTextFromMarkdown(
		`Licence: MIT | [GitHub](https://github.com/vinser/pixyne) repo | ` + versionLine + `

` + buildLine)

	noteRow := widget.NewRichTextFromMarkdown(`
---
*Created using* [Fyne](https://fyne.io) *GUI library*.

*App icon designed by* [Icon8](https://icon8.com).`)

	aboutDialog := dialog.NewCustom("About", "Ok", container.NewVBox(logoRow, infoRow, tecRow, noteRow), a.topWindow)
	aboutDialog.SetOnClosed(func() {
		a.enableShortcuts()
	})
	aboutDialog.Show()
}

// show info about app hotkeys
func (a *App) hotkeysDialog() {
	a.disableShortcuts()
	ctrlForm := widget.NewForm()
	for i := range a.ControlShortCuts {
		item := widget.FormItem{Text: a.ControlShortCuts[i].Name, Widget: widget.NewLabel("Ctrl + " + string(a.ControlShortCuts[i].KeyName))}
		ctrlForm.AppendItem(&item)
	}
	altForm := widget.NewForm()
	for i := range a.AltShortCuts {
		item := widget.FormItem{Text: a.AltShortCuts[i].Name, Widget: widget.NewLabel("Alt + " + string(a.AltShortCuts[i].KeyName))}
		altForm.AppendItem(&item)
	}

	keysDialog := dialog.NewCustom("Hotkeys", "Ok", container.NewGridWithColumns(2, ctrlForm, altForm), a.topWindow)
	keysDialog.SetOnClosed(func() {
		a.enableShortcuts()
	})
	keysDialog.Show()
}

// open photo folder dialog
func (a *App) openFolderDialog() {
	a.disableShortcuts()
	d := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
		frame.ShowProgress()
		defer frame.HideProgress()
		frame.StatusText.Set("")
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
		a.defaultState(false)
		rootURI = list
		a.topWindowTitle.Set(rootURI.Path())
		a.newPhotoList()
		a.newLayout()
		a.enableShortcuts()
	}, a.topWindow)
	d.SetLocation(rootURI)
	d.Resize(fyne.NewSize(672, 378))
	d.Show()
}

// Save choosed photos:
// 1. move dropped photo to droppped folder
// 2. update exif dates with file modify date or input date
func (a *App) savePhotoListDialog() {
	a.disableShortcuts()
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
				frame.ShowProgress()
				defer frame.HideProgress()
				frame.StatusText.Set("")
				a.SavePhotoList(renameFiles)
				a.defaultState(false)
				a.topWindowTitle.Set(a.state.Folder)
				a.newPhotoList()
				a.newLayout()
			}
			a.enableShortcuts()
		},
		a.topWindow)
	d.Show()
}

func (a *App) settingsDialog() {
	a.disableShortcuts()
	s := NewSettings()
	settingsForm := widget.NewForm(
		widget.NewFormItem("", s.scalePreviewsRow(a.topWindow.Canvas().Scale())),
		widget.NewFormItem("Scale", s.scalesRow()),
		widget.NewFormItem("Main Color", s.colorsRow()),
		widget.NewFormItem("Theme", s.themesRow()),
		widget.NewFormItem("Mode", s.modeRow(a)),
		widget.NewFormItem("Date Format", s.datesRow(a)),
		widget.NewFormItem("Hotkeys", s.hotkeysRow()),
	)

	d := dialog.NewCustom("Settings", "Ok", settingsForm, a.topWindow)
	d.SetOnClosed(func() {
		frame.ShowProgress()
		defer frame.HideProgress()
		if len(list) > 0 {
			if !a.frameView.Hidden {
				a.showFrameToolbar()
			}
			a.topWindow.Content().Refresh()
			frame.At(a.state.FramePos)
			frame.ItemEndingAt(a.state.ItemPos)
		}
		a.enableShortcuts()
	})
	resettButton := widget.NewButton("Reset all", func() { a.approveResetAllDialog(d) })
	resettButton.Importance = widget.DangerImportance
	settingsForm.Append("", container.NewBorder(nil, nil, nil, resettButton))
	d.Show()
}

func (a *App) approveResetAllDialog(parent *dialog.CustomDialog) {
	d := dialog.NewCustomConfirm(
		"Reset all",
		"Proceed",
		"Cancel",
		container.NewVBox(
			widget.NewLabel("WARNING!"),
			widget.NewLabel("All settings will be reset to default values."),
		),
		func(b bool) {
			if b {
				frame.StatusText.Set("")
				a.defaultState(true)
				a.topWindowTitle.Set(a.state.Folder)
				a.newPhotoList()
				a.newLayout()
				parent.Hide()
				parent = nil
			}
		},
		a.topWindow)
	d.Show()
}
