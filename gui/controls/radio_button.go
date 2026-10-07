package controls

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func NewRadioButton(theme *material.Theme, enum *widget.Enum, key string, label string) *RadioButton {
	s := material.RadioButton(theme, enum, key, label)
	s.Size = 18
	s.TextSize = theme.TextSize
	return &RadioButton{
		style: s,
	}
}

type RadioButton struct {
	style material.RadioButtonStyle
}

func (b *RadioButton) Layout(gtx layout.Context) layout.Dimensions {
	return b.style.Layout(gtx)
}
