package controls

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"image"
	"image/color"
)

var Style = Styling{
	BorderColor:            color.NRGBA{R: 181, G: 181, B: 181, A: 255},
	BorderFocusedColor:     color.NRGBA{R: 102, G: 128, B: 230, A: 230},
	BorderThickness:        1,
	BorderThicknessFocused: 2,
	BackgroundColor:        color.NRGBA{R: 240, G: 240, B: 239, A: 255},
	ForegroundColor:        color.NRGBA{R: 0, G: 0, B: 0, A: 255},
	SelectedColor:          color.NRGBA{R: 102, G: 128, B: 230, A: 80},
	SelectedFocusedColor:   color.NRGBA{R: 102, G: 128, B: 230, A: 160},
}

type Styling struct {
	BorderColor            color.NRGBA
	BorderFocusedColor     color.NRGBA
	BorderThickness        unit.Dp
	BorderThicknessFocused unit.Dp
	BackgroundColor        color.NRGBA
	ForegroundColor        color.NRGBA
	SelectedColor          color.NRGBA
	SelectedFocusedColor   color.NRGBA
	MonoFont               *font.Font
}

func (t Styling) bordering(focused bool) (color.NRGBA, unit.Dp) {
	if focused {
		return t.BorderFocusedColor, t.BorderThicknessFocused
	}
	return t.BorderColor, t.BorderThickness
}

func measureText(gtx layout.Context, theme *material.Theme, text string) layout.Dimensions {
	gtx.Constraints.Min = image.Point{}
	gtx.Constraints.Max = image.Point{X: 1e6, Y: 1e6}
	macro := op.Record(gtx.Ops)
	dims := material.Label(theme, theme.TextSize, text).Layout(gtx)
	_ = macro.Stop()
	return dims
}
