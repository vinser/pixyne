package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// In the case where the application was built bypassing the fyne command and using the standard go command
func init() {
	app.SetMetadata(fyne.AppMetadata{
		ID:      "com.github.vinser.pixyne",
		Name:    "Pixyne - photo picker",
		Version: "selfcrafted",
		Custom:  map[string]string{"OnGitHub": "https://github.com/vinser/pixyne"},
	})
}
