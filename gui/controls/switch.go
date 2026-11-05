package controls

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func NewSwitch(theme *material.Theme, initial bool) *Switch {
	result := &Switch{
		theme: theme,
		value: &widget.Bool{Value: initial},
	}
	result.control = material.Switch(theme, result.value, "")
	return result
}

type Switch struct {
	theme   *material.Theme
	value   *widget.Bool
	control material.SwitchStyle
}

func (s *Switch) Layout(gtx layout.Context) layout.Dimensions {
	if s.IsFocused(gtx) {
		s.control.Color.Enabled = s.theme.Palette.ContrastBg
		s.control.Color.Disabled = s.theme.Palette.Bg
	} else {
		s.control.Color.Enabled = Style.ButtonColor
		s.control.Color.Disabled = s.theme.Palette.Bg
	}
	return s.control.Layout(gtx)
}

func (s *Switch) Update(gtx layout.Context) bool {
	return s.value.Update(gtx)
}

func (s *Switch) IsFocused(gtx layout.Context) bool {
	return gtx.Focused(s.value)
}

func (s *Switch) SetFocused(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: s.value})
}

func (s *Switch) Checked() bool {
	return s.value.Value
}

func (s *Switch) SetChecked(b bool) {
	s.value.Value = b
}
