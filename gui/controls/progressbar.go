package controls

import (
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"image"
	"image/color"
)

type ProgressBar struct {
	Color      color.NRGBA
	Height     unit.Dp
	Radius     unit.Dp
	TrackColor color.NRGBA
	Progress   float32
	Reversed   bool
}

func (p ProgressBar) Layout(gtx layout.Context) layout.Dimensions {
	shader := func(width int, color color.NRGBA) layout.Dimensions {
		d := image.Point{X: width, Y: gtx.Dp(p.Height)}
		rr := gtx.Dp(p.Radius)
		defer clip.UniformRRect(image.Rectangle{Max: image.Pt(width, d.Y)}, rr).Push(gtx.Ops).Pop()
		paint.ColorOp{Color: color}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		return layout.Dimensions{Size: d}
	}
	progressBarWidth := gtx.Constraints.Max.X
	if p.Progress == 0 {
		if p.Reversed {
			return widget.Border{Color: Style.BorderColor, CornerRadius: p.Radius, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return shader(progressBarWidth, p.TrackColor)
			})
		}
		return widget.Border{Color: Style.BorderColor, CornerRadius: p.Radius, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return shader(progressBarWidth, p.Color)
		})
	}
	direction := layout.W
	if p.Reversed {
		direction = layout.E
	}
	return widget.Border{Color: Style.BorderColor, CornerRadius: p.Radius, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{Alignment: direction}.Layout(gtx,
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return shader(progressBarWidth, p.Color)
			}),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				fillWidth := int(float32(progressBarWidth) * min(1, max(0, p.Progress)))
				if fillWidth < int(p.Radius*2) {
					fillWidth = int(p.Radius * 2)
				}
				return shader(fillWidth, p.TrackColor)
			}),
		)
	})
}
