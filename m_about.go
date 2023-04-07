package main

import (
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// show info about app
func (a *App) aboutDialog() {
	objects := []fyne.CanvasObject{}

	if name := a.Metadata().Name; name != "" {
		objects = append(objects, widget.NewLabelWithStyle(fmt.Sprintf("%s \n", name), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
	}
	if version := a.Metadata().Custom["Version"]; version != "" {
		objects = append(objects, widget.NewLabel(fmt.Sprintf("Version: %s \n", version)))
	} else {
		if version := a.Metadata().Version; version != "" {
			objects = append(objects, widget.NewLabel(fmt.Sprintf("Version: %s \n", version)))
		}
	}
	if buildTime := a.Metadata().Custom["BuildTime"]; buildTime != "" {
		objects = append(objects, widget.NewLabel(fmt.Sprintf("Build time: %s \n", buildTime)))
	}
	if buildHost := a.Metadata().Custom["BuildHost"]; buildHost != "" {
		objects = append(objects, widget.NewLabel(fmt.Sprintf("Build host: %s \n", buildHost)))
	}
	if goVersion := a.Metadata().Custom["GoVersion"]; goVersion != "" {
		objects = append(objects, widget.NewLabel(fmt.Sprintf("Go version: %s \n", goVersion)))
	}
	if url, _ := url.Parse(a.Metadata().Custom["OnGitHub"]); url != nil {
		objects = append(objects, widget.NewHyperlink("On GitHub", url))
	}

	content := container.NewCenter(container.NewVBox(objects...))
	dialog.ShowCustom("About", "Ok", content, a.topWindow)
}
