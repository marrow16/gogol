package controls

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func NewPathButton(theme *material.Theme) *PathButton {
	b := &PathButton{}
	s := material.Button(theme, &b.clickable, "…")
	s.Inset = layout.Inset{Top: 4, Bottom: 0, Left: 4, Right: 4}
	s.Background = Style.BackgroundColor
	s.Color = Style.ForegroundColor
	s.TextSize = theme.TextSize
	b.style = s
	return b
}

type PathButton struct {
	clickable widget.Clickable
	style     material.ButtonStyle
}

func (b *PathButton) Layout(gtx layout.Context) layout.Dimensions {
	bc, bt := Style.bordering(b.IsFocused(gtx))
	return widget.Border{Color: bc, CornerRadius: 3, Width: bt}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return b.style.Layout(gtx)
	})
}

func (b *PathButton) Clicked(gtx layout.Context) bool {
	return b.clickable.Clicked(gtx)
}

func (b *PathButton) IsFocused(gtx layout.Context) bool {
	return gtx.Focused(&b.clickable)
}
