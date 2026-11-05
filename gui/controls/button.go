package controls

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image/color"
)

func NewButton(theme *material.Theme, label string) *Button {
	b := &Button{
		theme: theme,
	}
	s := material.Button(theme, &b.clickable, label)
	s.Inset = layout.Inset{Bottom: 2, Left: 4, Right: 4}
	s.TextSize = theme.TextSize
	b.style = s
	return b
}

type Button struct {
	theme     *material.Theme
	clickable widget.Clickable
	style     material.ButtonStyle
}

func (b *Button) Layout(gtx layout.Context) layout.Dimensions {
	if b.IsFocused(gtx) || b.clickable.Hovered() {
		b.style.Color = b.theme.ContrastFg
		b.style.Background = b.theme.ContrastBg
	} else {
		b.style.Color = Style.BackgroundColor
		b.style.Background = Style.ButtonColor
	}
	return b.style.Layout(gtx)
}

func (b *Button) Clicked(gtx layout.Context) bool {
	return b.clickable.Clicked(gtx)
}

func (b *Button) IsFocused(gtx layout.Context) bool {
	return gtx.Focused(&b.clickable)
}

func (b *Button) SetFocused(gtx layout.Context) {

	gtx.Execute(key.FocusCmd{Tag: &b.clickable})
}

func (b *Button) BackgroundColor(c color.NRGBA) *Button {
	b.style.Background = c
	return b
}
