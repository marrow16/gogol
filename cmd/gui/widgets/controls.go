package widgets

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func newButton(theme *material.Theme, label string) *button {
	b := &button{}
	s := material.Button(theme, &b.clickable, label)
	s.Inset = layout.Inset{Bottom: 2, Left: 4, Right: 4}
	s.TextSize = theme.TextSize
	b.style = s
	return b
}

type button struct {
	clickable widget.Clickable
	style     material.ButtonStyle
}

func (b *button) Layout(gtx layout.Context) layout.Dimensions {
	return b.style.Layout(gtx)
}

func (b *button) Clicked(gtx layout.Context) bool {
	return b.clickable.Clicked(gtx)
}

func (b *button) isFocused(gtx layout.Context) bool {
	return gtx.Focused(&b.clickable)
}

func newPathButton(theme *material.Theme) *pathButton {
	b := &pathButton{}
	s := material.Button(theme, &b.clickable, "…")
	s.Inset = layout.Inset{Top: 4, Bottom: 4, Left: 4, Right: 4}
	s.Background = popupBackground
	s.Color = popupForeground
	s.TextSize = theme.TextSize
	b.style = s
	return b
}

type pathButton struct {
	clickable widget.Clickable
	style     material.ButtonStyle
}

func (b *pathButton) Layout(gtx layout.Context) layout.Dimensions {
	if !isMac {
		return layout.Dimensions{}
	}
	return b.style.Layout(gtx)
}

func (b *pathButton) Clicked(gtx layout.Context) bool {
	return b.clickable.Clicked(gtx)
}

func (b *pathButton) isFocused(gtx layout.Context) bool {
	return gtx.Focused(&b.clickable)
}

func newRadioButton(theme *material.Theme, enum *widget.Enum, key string, label string) *radioButton {
	s := material.RadioButton(theme, enum, key, label)
	s.Size = 18
	s.TextSize = theme.TextSize
	return &radioButton{
		style: s,
	}
}

type radioButton struct {
	style material.RadioButtonStyle
}

func (b *radioButton) Layout(gtx layout.Context) layout.Dimensions {
	return b.style.Layout(gtx)
}

func newCheckBox(theme *material.Theme, label string, initial bool) *checkbox {
	v := &widget.Bool{Value: initial}
	s := material.CheckBox(theme, v, label)
	s.TextSize = theme.TextSize
	s.Size = 18
	return &checkbox{
		value: v,
		style: s,
	}
}

type checkbox struct {
	value *widget.Bool
	style material.CheckBoxStyle
}

func (c *checkbox) Layout(gtx layout.Context) layout.Dimensions {
	return c.style.Layout(gtx)
}

func (c *checkbox) Update(gtx layout.Context) bool {
	return c.value.Update(gtx)
}

func (c *checkbox) isFocused(gtx layout.Context) bool {
	return gtx.Focused(c.value)
}

func (c *checkbox) Checked() bool {
	return c.value.Value
}

func (c *checkbox) SetChecked(b bool) {
	c.value.Value = b
}
