package widgets

import (
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/cmd/gui/icons"
	"image"
	"strconv"
)

func newStatusBar(c *Core) *statusBar {
	sb := &statusBar{
		core:   c,
		height: 30,
	}
	sb.buttons = []layout.FlexChild{
		layout.Rigid(newStatusBarButton(sb.play, icons.Play).alt(
			func() bool {
				return sb.core.running
			}, sb.pause, icons.Pause).layout),
		layout.Rigid(newStatusBarButton(sb.step, icons.Step).layout),
		layout.Rigid(newStatusBarButton(sb.stepAhead, icons.SkipForward).layout),
		layout.Rigid(newStatusBarButton(sb.zoomIn, icons.ZoomIn).layout),
		layout.Rigid(newStatusBarButton(sb.zoomOut, icons.ZoomOut).layout),
		layout.Rigid(newStatusBarButton(sb.menu, icons.Burger).highlightWhen(
			func() bool {
				return sb.showingPopup == popupMenu
			}).layout),
	}
	sb.buttonsRecord = []layout.FlexChild{
		layout.Rigid(newStatusBarButton(sb.play, icons.Play).alt(
			func() bool {
				return sb.core.running
			}, sb.pause, icons.Pause).layout),
		layout.Rigid(newStatusBarButton(sb.step, icons.Step).layout),
		layout.Rigid(newStatusBarButton(sb.stepAhead, icons.SkipForward).layout),

		layout.Rigid(newStatusBarButton(sb.stepBack, icons.Reverse).layout),
		layout.Rigid(newStatusBarButton(sb.skipBack, icons.SkipBackward).layout),

		layout.Rigid(newStatusBarButton(sb.zoomIn, icons.ZoomIn).layout),
		layout.Rigid(newStatusBarButton(sb.zoomOut, icons.ZoomOut).layout),
		layout.Rigid(newStatusBarButton(sb.menu, icons.Burger).highlightWhen(
			func() bool {
				return sb.showingPopup == popupMenu
			}).layout),
	}
	sb.rulesPopup = newRulesPopup(sb)
	sb.menuPopup = newMenuPopup(sb)
	return sb
}

type popup int

const (
	popupNone = iota
	popupRule
	popupMenu
)

type statusBar struct {
	core          *Core
	height        unit.Dp
	rulesPopup    *rulesPopup
	menuPopup     *menuPopup
	showingPopup  popup
	top           int
	right         int
	ruleClickable widget.Clickable
	ruleDims      layout.Dimensions
	stepDims      layout.Dimensions
	buttons       []layout.FlexChild
	buttonsRecord []layout.FlexChild
}

func (sb *statusBar) play() {
	sb.showingPopup = popupNone
	if !sb.core.running {
		sb.core.start()
	}
}

func (sb *statusBar) pause() {
	sb.showingPopup = popupNone
	if sb.core.running {
		sb.core.stop()
	}
}

func (sb *statusBar) step() {
	sb.showingPopup = popupNone
	sb.core.step()
}

func (sb *statusBar) stepAhead() {
	sb.showingPopup = popupNone
	sb.core.stepAhead()
}

func (sb *statusBar) stepBack() {
	sb.showingPopup = popupNone
	sb.core.stepBack()
}

func (sb *statusBar) skipBack() {
	sb.showingPopup = popupNone
	sb.core.skipBack()
}

func (sb *statusBar) zoomIn() {
	sb.showingPopup = popupNone
	sb.core.zoomIn()
}

func (sb *statusBar) zoomOut() {
	sb.showingPopup = popupNone
	sb.core.zoomOut()
}

func (sb *statusBar) menu() {
	sb.showHidePopup(popupMenu)
}

func (sb *statusBar) showHidePopup(p popup) {
	if sb.showingPopup == p {
		sb.showingPopup = popupNone
	} else {
		sb.showingPopup = p
		switch sb.showingPopup {
		case popupRule:
			sb.core.stop()
			sb.rulesPopup.setSelected()
		case popupMenu:
			sb.core.stop()
			sb.core.clearMode()
			sb.menuPopup.focused = false
		}
	}
}

func (sb *statusBar) showPopups(gtx layout.Context) {
	switch sb.showingPopup {
	case popupRule:
		x := sb.stepDims.Size.X + gtx.Dp(unit.Dp(6)) + 3
		pgtx := gtx
		pgtx.Constraints = layout.Constraints{
			Max: image.Pt(
				sb.ruleDims.Size.X-gtx.Dp(unit.Dp(12))-3,
				sb.top,
			),
		}
		macro := op.Record(gtx.Ops)
		dims := sb.rulesPopup.layout(pgtx)
		call := macro.Stop()
		stack := op.Offset(image.Pt(x, sb.top-dims.Size.Y)).Push(gtx.Ops)
		call.Add(gtx.Ops)
		stack.Pop()
	case popupMenu:
		pgtx := gtx
		pgtx.Constraints = layout.Constraints{
			Max: image.Pt(sb.right, sb.top),
		}
		macro := op.Record(gtx.Ops)
		dims := sb.menuPopup.layout(pgtx)
		x := sb.right - dims.Size.X
		call := macro.Stop()
		stack := op.Offset(image.Pt(x, sb.top-dims.Size.Y)).Push(gtx.Ops)
		call.Add(gtx.Ops)
		stack.Pop()
	}
}

func (sb *statusBar) layout(gtx layout.Context, windowRect clip.Rect) layout.Dimensions {
	height := gtx.Dp(sb.height)
	size := image.Pt(gtx.Constraints.Max.X, height)
	sb.top = windowRect.Max.Y - height
	sb.right = windowRect.Max.X
	r := image.Rectangle{Max: size}
	paint.FillShape(gtx.Ops, popupBackground, clip.Rect(r).Op())
	paint.FillShape(gtx.Ops, popupBorder, clip.Rect(image.Rect(0, 0, size.X, 1)).Op())
	gtx.Constraints = layout.Exact(size)
	theme := sb.core.theme
	layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			switch {
			case sb.core.runningShortcut:
				sb.stepDims = sb.label(gtx, theme, "Running Shortcut", text.Start)
			case sb.core.status != "":
				sb.stepDims = sb.label(gtx, theme, sb.core.status, text.Start)
			case sb.core.mode != noMode:
				sb.stepDims = sb.label(gtx, theme, sb.core.modeDisplay(), text.Start)
			case sb.core.instrumentRepeat != nil && sb.core.instrumentRepeat.Found:
				sb.stepDims = sb.label(gtx, theme, "Repeat Found!", text.Start)
			default:
				sb.stepDims = sb.label(gtx, theme, "Step: "+commas(strconv.FormatUint(sb.core.gridHolder.grid.StepCount.Load(), 10)), text.Start)
			}
			return sb.stepDims
		}),
		layout.Flexed(2, func(gtx layout.Context) layout.Dimensions {
			for sb.ruleClickable.Clicked(gtx) {
				sb.showHidePopup(popupRule)
			}
			return sb.ruleClickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				sb.ruleDims = sb.label(gtx, theme, "Rule: "+sb.core.gridHolder.grid.Rule.Name(), text.Middle)
				return sb.ruleDims
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if sb.core.isRecording() {
				return layout.Inset{Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:      layout.Horizontal,
						Alignment: layout.End,
						Gap:       3,
					}.Layout(gtx, sb.buttonsRecord...)
				})
			} else {
				return layout.Inset{Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:      layout.Horizontal,
						Alignment: layout.End,
						Gap:       3,
					}.Layout(gtx, sb.buttons...)
				})
			}
		}),
	)
	return layout.Dimensions{Size: size}
}

func (sb *statusBar) label(gtx layout.Context, theme *material.Theme, s string, align text.Alignment) layout.Dimensions {
	return layout.Inset{
		Left:  unit.Dp(8),
		Right: unit.Dp(8),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Max
		sb.drawInsetRect(gtx, image.Rect(0, 0, size.X, size.Y))
		lbl := material.Body1(theme, s)
		lbl.Color = popupForeground
		lbl.Alignment = align
		lbl.MaxLines = 1
		gtx.Constraints.Min.Y = size.Y
		gtx.Constraints.Max.Y = size.Y
		return layout.Inset{
			Left:  unit.Dp(6),
			Right: unit.Dp(6),
			Top:   unit.Dp(3),
		}.Layout(gtx, lbl.Layout)
	})
}

type statusBarButton struct {
	clickable widget.Clickable
	fn        func()
	img       image.Image
	isAlt     bool
	altImg    image.Image
	altFn     func()
	altCheck  func() bool
	hltCheck  func() bool
}

func newStatusBarButton(fn func(), img image.Image) *statusBarButton {
	return &statusBarButton{
		fn:  fn,
		img: img,
	}
}

func (b *statusBarButton) alt(altCheck func() bool, altFn func(), altImg image.Image) *statusBarButton {
	b.altCheck, b.altFn, b.altImg = altCheck, altFn, altImg
	b.isAlt = b.altCheck != nil && b.altFn != nil && b.altImg != nil
	return b
}

func (b *statusBarButton) highlightWhen(fn func() bool) *statusBarButton {
	b.hltCheck = fn
	return b
}

func (b *statusBarButton) useImage() image.Image {
	if b.isAlt && b.altCheck() {
		return b.altImg
	}
	return b.img
}

func (b *statusBarButton) layout(gtx layout.Context) layout.Dimensions {
	return layout.Inset{
		Top:    unit.Dp(3),
		Bottom: unit.Dp(3),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		wh := gtx.Constraints.Max.Y
		if wh > gtx.Constraints.Max.X {
			wh = gtx.Constraints.Max.X
		}
		size := image.Pt(wh, wh)
		for b.clickable.Clicked(gtx) {
			if b.isAlt && b.altCheck() {
				b.altFn()
			} else {
				b.fn()
			}
		}
		return b.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			useImg := b.useImage()
			r := useImg.Bounds()
			if b.hltCheck != nil && b.hltCheck() {
				paint.FillShape(
					gtx.Ops,
					popupHighlightColor,
					clip.UniformRRect(image.Rectangle{Max: size}, 4).Op(gtx.Ops),
				)
			}
			defer op.Affine(
				f32.Affine2D{}.Scale(
					f32.Pt(0, 0),
					f32.Pt(
						float32(size.X)/float32(r.Dx()),
						float32(size.Y)/float32(r.Dy()),
					),
				),
			).Push(gtx.Ops).Pop()
			paint.NewImageOp(useImg).Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			return layout.Dimensions{Size: size}
		})
	})
}

func (sb *statusBar) drawInsetRect(gtx layout.Context, r image.Rectangle) {
	// top
	paint.FillShape(gtx.Ops, popupBorder,
		clip.Rect(image.Rect(
			r.Min.X,
			r.Min.Y+4,
			r.Max.X,
			r.Min.Y+5,
		)).Op(),
	)
	// left
	paint.FillShape(gtx.Ops, popupBorder,
		clip.Rect(image.Rect(
			r.Min.X,
			r.Min.Y+4,
			r.Min.X+1,
			r.Max.Y-5,
		)).Op(),
	)
	// bottom
	paint.FillShape(gtx.Ops, popupBorderLight,
		clip.Rect(image.Rect(
			r.Min.X,
			r.Max.Y-4,
			r.Max.X,
			r.Max.Y-5,
		)).Op(),
	)
	// right
	paint.FillShape(gtx.Ops, popupBorderLight,
		clip.Rect(image.Rect(
			r.Max.X-1,
			r.Min.Y+4,
			r.Max.X,
			r.Max.Y-5,
		)).Op(),
	)
}
