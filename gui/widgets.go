package gui

import (
	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
	"image/color"
	"strings"
)

func horizontalLine(gtx layout.Context, c color.NRGBA, width, thickness int) {
	paint.FillShape(gtx.Ops, c, clip.Rect{Max: image.Point{X: width, Y: thickness}}.Op())
}

func borders(gtx layout.Context, dims layout.Dimensions, c color.NRGBA, top, left, bottom, right bool) {
	if top {
		paint.FillShape(gtx.Ops, c, clip.Rect{Max: image.Point{X: dims.Size.X, Y: 1}}.Op())
	}
	if left {
		paint.FillShape(gtx.Ops, c, clip.Rect{Max: image.Point{X: 1, Y: dims.Size.Y}}.Op())
	}
	if bottom {
		paint.FillShape(gtx.Ops, c, clip.Rect{Min: image.Point{X: 0, Y: dims.Size.Y - 1}, Max: image.Point{X: dims.Size.X, Y: dims.Size.Y}}.Op())
	}
	if right {
		paint.FillShape(gtx.Ops, c, clip.Rect{Min: image.Point{X: dims.Size.X - 1, Y: 0}, Max: image.Point{X: dims.Size.X, Y: dims.Size.Y}}.Op())
	}
}

func errorLabel(err error) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if err == nil {
			return layout.Dimensions{}
		}
		lbl := material.Label(theme, theme.TextSize, err.Error())
		lbl.MaxLines = 1
		lbl.Color = errorColor
		return lbl.Layout(gtx)
	}
}

func insetErrorLabel(err error) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if err == nil {
			return layout.Dimensions{}
		}
		lbl := material.Label(theme, theme.TextSize, err.Error())
		lbl.MaxLines = 1
		lbl.Color = errorColor
		return layout.Inset{Top: 3}.Layout(gtx, lbl.Layout)
	}
}

func label(s string) layout.Widget {
	lbl := material.Label(theme, theme.TextSize, s)
	lbl.MaxLines = 1
	return lbl.Layout
}

func labelItalic(s string) layout.Widget {
	lbl := material.Label(theme, theme.TextSize, s)
	lbl.MaxLines = 1
	lbl.Font.Style = font.Italic
	return lbl.Layout
}

func textLabel(s string) layout.Widget {
	return material.Label(theme, theme.TextSize, s).Layout
}

func labelRight(s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(theme, theme.TextSize, s)
		lbl.Alignment = text.End
		lbl.MaxLines = 1
		return lbl.Layout(gtx)
	}
}

func insetLabel(s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(theme, theme.TextSize, s)
		lbl.MaxLines = 1
		return layout.Inset{Top: 3}.Layout(gtx, lbl.Layout)
	}
}

func linkLabel(btn *widget.Clickable, s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return material.Clickable(gtx, btn, func(gtx layout.Context) layout.Dimensions {
			return layout.Stack{}.Layout(gtx,
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(theme, theme.TextSize, s)
					lbl.MaxLines = 1
					lbl.Color = linkColor
					pointer.CursorPointer.Add(gtx.Ops)
					return lbl.Layout(gtx)
				}),
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					thickness := gtx.Dp(1)
					rect := clip.Rect{
						Min: image.Point{X: 0, Y: gtx.Constraints.Min.Y - thickness},
						Max: image.Point{X: gtx.Constraints.Min.X, Y: gtx.Constraints.Min.Y},
					}
					defer rect.Push(gtx.Ops).Pop()
					paint.ColorOp{Color: linkColor}.Add(gtx.Ops)
					paint.PaintOp{}.Add(gtx.Ops)
					return layout.Dimensions{Size: gtx.Constraints.Min}
				}),
			)
		})
	}
}

func inset(top, bottom, left, right unit.Dp, w layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: top, Bottom: bottom, Left: left, Right: right}.Layout(gtx, w)
	}
}

func borderedInset(top, bottom, left, right unit.Dp, w layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return widget.Border{
			Color:        borderColor,
			CornerRadius: 3,
			Width:        1,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: top, Bottom: bottom, Left: left, Right: right}.Layout(gtx, w)
		})
	}
}

func flexed(w layout.Widget, weighted ...float32) layout.FlexChild {
	weight := float32(1.0)
	if len(weighted) > 0 {
		weight = weighted[0]
	}
	return layout.Flexed(weight, func(gtx layout.Context) layout.Dimensions {
		return w(gtx)
	})
}

func conditionalFlexed(cond bool, w1, w2 layout.Widget) layout.FlexChild {
	return layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		if cond {
			return w1(gtx)
		} else if w2 != nil {
			return w2(gtx)
		}
		return layout.Dimensions{}
	})
}

func conditionalRigid(cond bool, w1, w2 layout.Widget) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		if cond {
			return w1(gtx)
		} else if w2 != nil {
			return w2(gtx)
		}
		return layout.Dimensions{}
	})
}

func rigid(w layout.Widget) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return w(gtx)
	})
}

func rigidLeftRight(minWd int, wl, wr layout.Widget) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		if minWd > 0 {
			gtx.Constraints.Min.X = minWd
		} else {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
		}
		return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
			rigid(wl),
			rigid(wr),
		)
	})
}

func rigidImage(img *image.NRGBA) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		size := img.Bounds().Size()
		stack := clip.Rect{Max: size}.Push(gtx.Ops)
		defer stack.Pop()
		paint.NewImageOp(img).Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		return layout.Dimensions{Size: size}
	})
}

func rigidLabel(label string, align text.Alignment, weight font.Weight, fixedWidth int) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		if fixedWidth > 0 {
			gtx.Constraints.Min.X = fixedWidth
			gtx.Constraints.Max.X = fixedWidth
		}
		lbl := material.Label(theme, theme.TextSize, label)
		lbl.MaxLines = 1
		lbl.Alignment = align
		lbl.Font.Weight = weight
		return lbl.Layout(gtx)
	})
}

func rigidFixedWidth(w layout.Widget, width int, direction layout.Direction) layout.FlexChild {
	if w == nil {
		return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = width
			return layout.Dimensions{Size: gtx.Constraints.Min}
		})
	}
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		if width > 0 {
			gtx.Constraints.Min.X = width
			gtx.Constraints.Max.X = width
		}
		return direction.Layout(gtx, w)
	})
}

func rigidSpacerVertical(ht int) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Point{Y: ht}}
	})
}

func flexVertical(gap int, children ...layout.FlexChild) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Gap: gap}.Layout(gtx, children...)
	}
}

func flexHorizontal(gap int, children ...layout.FlexChild) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Gap: gap}.Layout(gtx, children...)
	}
}

func flexHorizontalSpaced(gap int, spacing layout.Spacing, children ...layout.FlexChild) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Gap: gap, Spacing: spacing}.Layout(gtx, children...)
	}
}

func popoutLayout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	return layout.Inset{Left: 8, Right: 8, Top: 8, Bottom: 8}.Layout(gtx, w)
}

func measureText(gtx layout.Context, text string) layout.Dimensions {
	gtx.Constraints.Min = image.Point{}
	gtx.Constraints.Max = image.Point{X: 1e6, Y: 1e6}
	macro := op.Record(gtx.Ops)
	dims := material.Label(theme, theme.TextSize, text).Layout(gtx)
	_ = macro.Stop()
	return dims
}

func measureMaxText(gtx layout.Context, weight font.Weight, texts ...string) int {
	gtx.Constraints.Min = image.Point{}
	gtx.Constraints.Max = image.Point{X: 1e6, Y: 1e6}
	macro := op.Record(gtx.Ops)
	// join all the texts as multi-line - the width will be the widest line...
	lbl := material.Label(theme, theme.TextSize, strings.Join(texts, "\n"))
	lbl.Font.Weight = weight
	dims := lbl.Layout(gtx)
	_ = macro.Stop()
	return dims.Size.X
}
