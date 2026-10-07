package controls

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image/color"
)

func NewPathButton(theme *material.Theme, fg, bg color.NRGBA) *PathButton {
	b := &PathButton{}
	s := material.Button(theme, &b.clickable, "…")
	s.Inset = layout.Inset{Top: 4, Bottom: 4, Left: 4, Right: 4}
	s.Background = bg
	s.Color = fg
	s.TextSize = theme.TextSize
	b.style = s
	return b
}

type PathButton struct {
	clickable widget.Clickable
	style     material.ButtonStyle
}

func (b *PathButton) Layout(gtx layout.Context) layout.Dimensions {
	return b.style.Layout(gtx)
}

func (b *PathButton) Clicked(gtx layout.Context) bool {
	return b.clickable.Clicked(gtx)
}

func (b *PathButton) IsFocused(gtx layout.Context) bool {
	return gtx.Focused(&b.clickable)
}
