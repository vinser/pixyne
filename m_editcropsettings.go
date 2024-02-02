package main

import (
	"fmt"
	"slices"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (e *Editor) editCropSettingsDialog() {
	e.clearCrop()
	var d *dialog.CustomDialog
	cc := container.NewVBox()
	for i := 0; i < len(a.state.FavoriteCrops); i++ {
		crop := a.state.FavoriteCrops[i]
		row := e.cropSettingsDeleteRow(crop, cc, i)
		cc.Add(row)
	}
	c := container.NewVBox()
	c.Add(cc)
	// insert new crop aspect
	newCropRow := e.cropSettingsAddRow(c, cc)
	c.Add(newCropRow)
	d = dialog.NewCustom("Set favorite crops", "Ok", c, a.topWindow)
	d.SetOnClosed(func() {
		e.fixedCrops.RemoveAll()
		for i := 0; i < len(a.state.FavoriteCrops); i++ {
			crop := a.state.FavoriteCrops[i]
			e.fixedCrops.Objects = append(e.fixedCrops.Objects, widget.NewButton(fmt.Sprintf("%.f:%.f", crop.X, crop.Y), func() { e.fixCrop(crop.X / crop.Y) }))
		}
		e.fixedCrops.Refresh()
	})
	d.Show()
}

func (e *Editor) cropSettingsDeleteRow(crop fyne.Position, cc *fyne.Container, i int) *fyne.Container {
	label := widget.NewLabel(fmt.Sprintf("%.f:%.f", crop.X, crop.Y))
	label.Alignment = fyne.TextAlignCenter
	button := widget.NewButtonWithIcon("",
		theme.ContentRemoveIcon(),
		func() {
			a.state.FavoriteCrops = slices.Delete(a.state.FavoriteCrops, i, i+1)
			cc.RemoveAll()
			for i := 0; i < len(a.state.FavoriteCrops); i++ {
				crop := a.state.FavoriteCrops[i]
				row := e.cropSettingsDeleteRow(crop, cc, i)
				cc.Add(row)
			}
			cc.Refresh()
		},
	)
	return container.NewGridWithColumns(2, label, button)
}

func (e *Editor) cropSettingsAddRow(c, cc *fyne.Container) *fyne.Container {
	newX := widget.NewEntry()
	newX.SetPlaceHolder("10")
	newX.Validator = validation.NewRegexp(`\d{1,2}`, "invalid width")
	newY := widget.NewEntry()
	newY.SetPlaceHolder("10")
	newY.Validator = validation.NewRegexp(`\d{1,2}`, "invalid height")

	return container.NewGridWithColumns(2,
		container.NewHBox(newX, widget.NewLabel(":"), newY),
		widget.NewButtonWithIcon("",
			theme.ContentAddIcon(),
			func() {
				if err := newX.Validate(); err != nil {
					return
				}
				if err := newY.Validate(); err != nil {
					return
				}
				x, _ := strconv.ParseFloat(newX.Text, 32)
				y, _ := strconv.ParseFloat(newY.Text, 32)
				a.state.FavoriteCrops = append(a.state.FavoriteCrops, fyne.NewPos(float32(x), float32(y)))
				cc.RemoveAll()
				for i := 0; i < len(a.state.FavoriteCrops); i++ {
					crop := a.state.FavoriteCrops[i]
					row := e.cropSettingsDeleteRow(crop, cc, i)
					cc.Add(row)
				}
				newX.Text = ""
				newY.Text = ""
				c.Refresh()
			},
		))
}
