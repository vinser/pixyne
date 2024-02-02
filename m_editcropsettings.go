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
	var d *dialog.CustomDialog
	c := container.NewVBox()
	for i, crop := range a.state.FavoriteCrops {
		row := e.cropSettingsDeleteRow(crop, c, i)
		c.Add(row)
	}
	// insert new crop aspect
	newCropRow := e.cropSettingsAddRow(c)
	c.Add(newCropRow)
	d = dialog.NewCustom("Set favorite crops", "Ok", c, a.topWindow)
	d.SetOnClosed(func() {
		e.fixedCrops.RemoveAll()
		for _, crop := range a.state.FavoriteCrops {
			e.fixedCrops.Objects = append(e.fixedCrops.Objects, widget.NewButton(fmt.Sprintf("%.f:%.f", crop.X, crop.Y), func() { e.fixCrop(crop.X / crop.Y) }))
		}
		e.fixedCrops.Refresh()
	})
	d.Show()
}

func (e *Editor) cropSettingsDeleteRow(crop fyne.Position, c *fyne.Container, i int) *fyne.Container {
	label := widget.NewLabel(fmt.Sprintf("%.f:%.f", crop.X, crop.Y))
	label.Alignment = fyne.TextAlignCenter
	button := widget.NewButtonWithIcon("",
		theme.ContentRemoveIcon(),
		func() {
			a.state.FavoriteCrops = slices.Delete(a.state.FavoriteCrops, i, i+1)
			c.Objects = slices.Delete(c.Objects, i, i+1)
			c.Refresh()
		},
	)
	return container.NewGridWithColumns(2, label, button)
}

func (e *Editor) cropSettingsAddRow(c *fyne.Container) *fyne.Container {
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
				lastRow := c.Objects[len(c.Objects)-1].(*fyne.Container)
				c.Remove(lastRow)
				c.Add(e.cropSettingsDeleteRow(fyne.NewPos(float32(x), float32(y)), c, len(a.state.FavoriteCrops)-1))
				c.Add(lastRow)
				newX.Text = ""
				newY.Text = ""
				c.Refresh()
			},
		))
}
