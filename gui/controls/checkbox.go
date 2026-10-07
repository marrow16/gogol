package controls

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func NewCheckBox(theme *material.Theme, label string, initial bool) *Checkbox {
	v := &widget.Bool{Value: initial}
	s := material.CheckBox(theme, v, label)
	s.TextSize = theme.TextSize
	s.Size = 18
	return &Checkbox{
		value: v,
		style: s,
	}
}

type Checkbox struct {
	value *widget.Bool
	style material.CheckBoxStyle
}

func (c *Checkbox) Layout(gtx layout.Context) layout.Dimensions {
	return c.style.Layout(gtx)
}

func (c *Checkbox) Update(gtx layout.Context) bool {
	return c.value.Update(gtx)
}

func (c *Checkbox) IsFocused(gtx layout.Context) bool {
	return gtx.Focused(c.value)
}

func (c *Checkbox) Checked() bool {
	return c.value.Value
}

func (c *Checkbox) SetChecked(b bool) {
	c.value.Value = b
}
