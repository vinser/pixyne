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
	// objects := []fyne.CanvasObject{}

	// if name := a.Metadata().Name; name != "" {
	// 	objects = append(objects, widget.NewLabel("Use Pixyne to quickly review your photo folder and safely drop bad and similar shots"))
	// }
	// if appUrl, _ := url.Parse(a.Metadata().Custom["OnGitHub"]); appUrl != nil {
	// 	c := container.NewHBox(widget.NewHyperlink("View on GitHub", appUrl), widget.NewLabel("Lecense: MIT"))
	// 	objects = append(objects, c)
	// }

	// if version := a.Metadata().Custom["Version"]; version != "" {
	// 	objects = append(objects, widget.NewLabel(fmt.Sprintf("Version: %s \n", version)))
	// } else {
	// 	if version := a.Metadata().Version; version != "" {
	// 		objects = append(objects, widget.NewLabel(fmt.Sprintf("Version: %s \n", version)))
	// 	}
	// }
	// if buildTime := a.Metadata().Custom["BuildTime"]; buildTime != "" {
	// 	objects = append(objects, widget.NewLabel(fmt.Sprintf("Build time: %s \n", buildTime)))
	// }
	// if buildHost := a.Metadata().Custom["BuildHost"]; buildHost != "" {
	// 	objects = append(objects, widget.NewLabel(fmt.Sprintf("Build host: %s \n", buildHost)))
	// }
	// if goVersion := a.Metadata().Custom["GoVersion"]; goVersion != "" {
	// 	objects = append(objects, widget.NewLabel(fmt.Sprintf("Go version: %s \n", goVersion)))
	// }
	// if fyneUrl, _ := url.Parse("https://fyne.io/"); fyneUrl != nil {
	// 	c := container.NewHBox(widget.NewLabel("Created using"), widget.NewHyperlink("Fyne GUI library", fyneUrl))
	// 	objects = append(objects, c)
	// }
	// if icon8Url, _ := url.Parse("https://icon8.com"); icon8Url != nil {
	// 	c := container.NewHBox(widget.NewLabel("App icon design"), widget.NewIcon(appIcon), widget.NewHyperlink("by Icon8", icon8Url))
	// 	objects = append(objects, c)
	// }

	// content := container.NewCenter(container.NewVBox(objects...))
	logo := canvas.NewImageFromResource(appIcon)
	logo.FillMode = canvas.ImageFillOriginal
	logoRow := container.NewGridWithColumns(8, logo)
	infoRow := widget.NewRichTextFromMarkdown(`
## Pixyne - photo picker
---
Use [Pixyne](http://vinser.github.io/pixyne) to quickly review your photo folder and safely delete bad and similar shots.

You may also set EXIF shooting date to the file date or to a manually entered date.

---`)
	var version, buildHost, buildTime, goVersion, versionLine, buildLine string
	if version = a.Metadata().Custom["Version"]; version == "" {
		if version = a.Metadata().Version; version == "" {
			version = "selfcrafted"
		}
	}
	versionLine = "Version :" + version

	if buildHost = a.Metadata().Custom["BuildHost"]; buildHost != "" {
		buildLine = fmt.Sprintf("Build host: %s ", buildHost)
	}
	if buildTime = a.Metadata().Custom["BuildTime"]; buildTime != "" {
		buildLine = buildLine + fmt.Sprintf(" | Build time: %s ", buildTime)
	}
	if goVersion = a.Metadata().Custom["GoVersion"]; goVersion != "" {
		buildLine = buildLine + fmt.Sprintf(" | Go version: %s ", goVersion)
	}
	tecRow := widget.NewRichTextFromMarkdown(`
Licence: MIT | [GitHub](https://github.com/vinser/pixyne) repo | ` + versionLine + `

` + buildLine)

	noteRow := widget.NewRichTextFromMarkdown(`
---
*Created using* [Fyne](https://fyne.io) *GUI library*.

*App icon designed by* [Icon8](https://icon8.com).`)

	aboutDialog := dialog.NewCustom("About", "Ok", container.NewVBox(logoRow, infoRow, tecRow, noteRow), a.topWindow)
	aboutDialog.Show()
}
