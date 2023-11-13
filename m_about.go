package main

import (
	"fmt"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// show info about app
func (a *App) aboutDialog() {
	logo := canvas.NewImageFromResource(appIcon)
	logo.FillMode = canvas.ImageFillOriginal
	logoRow := container.NewGridWithColumns(8, logo)
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
	aboutDialog.Show()
}
