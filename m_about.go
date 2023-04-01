package main

import (
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}

// show info about app
func (a *App) aboutDialog() {
	lblName := widget.NewLabelWithStyle(fmt.Sprintf("%s \n", a.Metadata().Name), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	lblVersion := widget.NewLabel(fmt.Sprintf("Version %s \n", a.Metadata().Custom["Version"]))
	lblBuildTime := widget.NewLabel(fmt.Sprintf("Build time %s \n", a.Metadata().Custom["BuildTime"]))
	lblGoVersion := widget.NewLabel(fmt.Sprintf("Go version %s \n", a.Metadata().Custom["GoVersion"]))
	lblOnGitHub := widget.NewHyperlink("On GitHub", parseURL(a.Metadata().Custom["OnGitHub"]))
	content := container.NewCenter(container.NewVBox(lblName, lblVersion, lblBuildTime, lblGoVersion, lblOnGitHub))
	dialog.ShowCustom("About", "Ok", content, a.wMain)
}
