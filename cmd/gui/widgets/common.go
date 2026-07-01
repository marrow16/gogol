package widgets

import (
	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget/material"
	"image"
	"image/color"
	"runtime"
)

var modKeyName = func() string {
	if runtime.GOOS == "darwin" {
		return "⌥" //"⌘"
	}
	return "Ctrl+"
}()
var modKey = func() key.Modifiers {
	if runtime.GOOS == "darwin" {
		return key.ModAlt
	}
	return key.ModCtrl
}()

var isMac = runtime.GOOS == "darwin"

var (
	popupForeground                = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	popupBackground                = color.NRGBA{R: 240, G: 240, B: 239, A: 255}
	popupBorder                    = color.NRGBA{R: 181, G: 181, B: 181, A: 255}
	popupBorderFocused             = color.NRGBA{R: 102, G: 128, B: 230, A: 230}
	popupBorderLight               = color.NRGBA{R: 250, G: 250, B: 250, A: 255}
	popupSelectedBackground        = color.NRGBA{R: 102, G: 128, B: 230, A: 128}
	popupSelectedFocusedBackground = color.NRGBA{R: 102, G: 128, B: 230, A: 200}
	popupHighlightColor            = popupSelectedFocusedBackground
	errorColor                     = color.NRGBA{R: 200, G: 0, B: 0, A: 255}
)

func commas(s string) string {
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func handleEscape(gtx layout.Context) bool {
	if ev, ok := gtx.Event(key.Filter{Name: key.NameEscape}); ok {
		if kev, ok := ev.(key.Event); ok && kev.State == key.Press && kev.Name == key.NameEscape {
			return true
		}
	}
	return false
}

func border(gtx layout.Context, dims layout.Dimensions, top, left, bottom, right bool) {
	if top {
		paint.FillShape(gtx.Ops, popupBorder,
			clip.Rect(image.Rect(
				0,
				0,
				dims.Size.X,
				1,
			)).Op(),
		)
	}
	if left {
		paint.FillShape(gtx.Ops, popupBorder,
			clip.Rect(image.Rect(
				0,
				0,
				1,
				dims.Size.Y,
			)).Op(),
		)
	}
	if bottom {
		paint.FillShape(gtx.Ops, popupBorder,
			clip.Rect(image.Rect(
				0,
				dims.Size.Y-1,
				dims.Size.X,
				dims.Size.Y,
			)).Op(),
		)
	}
	if right {
		paint.FillShape(gtx.Ops, popupBorder,
			clip.Rect(image.Rect(
				dims.Size.X-1,
				0,
				dims.Size.X,
				dims.Size.Y,
			)).Op(),
		)
	}
}

func rightLabel(theme *material.Theme, s string) material.LabelStyle {
	lbl := material.Body1(theme, s)
	lbl.Alignment = text.End
	lbl.MaxLines = 1
	return lbl
}

func measureText(gtx layout.Context, theme *material.Theme, text string) layout.Dimensions {
	gtx.Constraints.Min = image.Point{}
	gtx.Constraints.Max = image.Pt(1e6, 1e6)
	macro := op.Record(gtx.Ops)
	dims := material.Body1(theme, text).Layout(gtx)
	_ = macro.Stop()
	return dims
}

func measureMaxText(gtx layout.Context, theme *material.Theme, weight font.Weight, text ...string) layout.Dimensions {
	mx := layout.Dimensions{}
	macro := op.Record(gtx.Ops)
	for _, t := range text {
		lbl := material.Body1(theme, t)
		lbl.Font.Weight = weight
		dims := lbl.Layout(gtx)
		if dims.Size.X > mx.Size.X {
			mx.Size.X = dims.Size.X
		}
		if dims.Size.Y > mx.Size.Y {
			mx.Size.Y = dims.Size.Y
		}
	}
	_ = macro.Stop()
	return mx
}
