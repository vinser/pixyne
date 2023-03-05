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
func aboutScreen() {
	hlPhotofyne := parseURL("https://github.com/vinser/photofyne")
	c := container.NewCenter(container.NewVBox(
		widget.NewLabelWithStyle("Photofyne App", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("Version: %s (%s)\n", version, target)),
		widget.NewLabel(fmt.Sprintf("Build time: %s\n", buildTime)),
		widget.NewLabel(fmt.Sprintf("Golang version: %s\n", goversion)),
		widget.NewHyperlink("Photofyne on GitHub", hlPhotofyne),
	))
	dialog.ShowCustom("About", "Ok", c, wMain)
}
