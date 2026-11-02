package controls

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
	"image/color"
)

func NewFlatButton(theme *material.Theme, label string) *FlatButton {
	b := &FlatButton{
		inset: layout.Inset{Left: 4, Right: 4},
	}
	s := material.Label(theme, theme.TextSize, label)
	s.Color = Style.ForegroundColor
	b.style = s
	return b
}

type FlatButton struct {
	clickable  widget.Clickable
	inset      layout.Inset
	margin     layout.Inset
	style      material.LabelStyle
	background *color.NRGBA
}

func (b *FlatButton) Layout(gtx layout.Context) layout.Dimensions {
	if b.background != nil {
		return b.margin.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return b.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return widget.Border{
					Color:        *b.background,
					CornerRadius: 3,
					Width:        1,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					macro := op.Record(gtx.Ops)
					dims := b.inset.Layout(gtx, b.style.Layout)
					content := macro.Stop()
					fill := op.Record(gtx.Ops)
					r := gtx.Dp(3)
					rr := clip.RRect{
						Rect: image.Rectangle{Max: dims.Size},
						NE:   r,
						NW:   r,
						SE:   r,
						SW:   r,
					}
					paint.FillShape(gtx.Ops, *b.background, rr.Op(gtx.Ops))
					background := fill.Stop()
					background.Add(gtx.Ops)
					content.Add(gtx.Ops)
					if gtx.Focused(&b.clickable) || b.clickable.Hovered() {
						paint.FillShape(gtx.Ops, Style.ShadeColor, clip.UniformRRect(image.Rectangle{Max: dims.Size}, r).Op(gtx.Ops))
					}
					return dims
				})
			})
		})
	}
	return b.margin.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return b.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return widget.Border{Color: Style.BorderColor, CornerRadius: 3, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				dims := b.inset.Layout(gtx, b.style.Layout)
				if gtx.Focused(&b.clickable) || b.clickable.Hovered() {
					r := gtx.Dp(3)
					paint.FillShape(gtx.Ops, Style.ShadeColor, clip.UniformRRect(image.Rectangle{Max: dims.Size}, r).Op(gtx.Ops))
				}
				return dims
			})
		})
	})
}

func (b *FlatButton) Clicked(gtx layout.Context) bool {
	return b.clickable.Clicked(gtx)
}

func (b *FlatButton) IsFocused(gtx layout.Context) bool {
	return gtx.Focused(&b.clickable)
}

func (b *FlatButton) SetFocused(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: &b.clickable})
}

func (b *FlatButton) HandleKeys(gtx layout.Context, fn func(k key.Name), keyNames ...key.Name) {
	if fn == nil || len(keyNames) == 0 {
		return
	}
	filters := make([]event.Filter, len(keyNames))
	for i, kn := range keyNames {
		filters[i] = key.Filter{Focus: &b.clickable, Name: kn}
	}
	for {
		ev, ok := gtx.Event(filters...)
		if !ok {
			break
		}
		if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
			fn(kev.Name)
		}
	}
}

func (b *FlatButton) Label(label string) *FlatButton {
	b.style.Text = label
	return b
}

func (b *FlatButton) Inset(i layout.Inset) *FlatButton {
	b.inset = i
	return b
}

func (b *FlatButton) Margin(i layout.Inset) *FlatButton {
	b.margin = i
	return b
}

func (b *FlatButton) Colors(fg color.NRGBA, bg *color.NRGBA) *FlatButton {
	b.style.Color = fg
	b.background = bg
	return b
}
