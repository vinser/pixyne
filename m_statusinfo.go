package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

type StatusInfo struct {
	widget.BaseWidget
	text     binding.String
	label    *widget.Label
	progress *widget.ProgressBarInfinite
}

func newStatusInfo() *StatusInfo {
	status := &StatusInfo{}
	status.text = binding.NewString()
	status.showPosition()
	label := widget.NewLabelWithData(status.text)
	label.Alignment = fyne.TextAlignCenter
	progress := widget.NewProgressBarInfinite()
	progress.Hide()
	status.label = label
	status.progress = progress

	status.ExtendBaseWidget(status)
	return status
}

func (si *StatusInfo) showPosition() {
	if len(list) == 0 {
		si.text.Set("")
	} else {
		si.text.Set(fmt.Sprintf("%d/%d", a.state.FramePos+a.state.ItemPos+1, len(list)))
	}
}

func (si *StatusInfo) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewStack(si.label, si.progress, newFixedSpacer(fyne.NewSize(400, 0)))
	return widget.NewSimpleRenderer(c)
}
func (si *StatusInfo) ShowProgress() {
	frame.DisableButtons()
	si.progress.Show()
}
func (si *StatusInfo) HideProgress() {
	si.showPosition()
	si.progress.Hide()
	frame.EnableButtons()
}

func (si *StatusInfo) ToolbarObject() fyne.CanvasObject {
	return si
}
