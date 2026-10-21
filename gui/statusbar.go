package gui

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
	"github.com/marrow16/gogol/gui/icons"
	"image"
	"image/color"
	"strconv"
	"time"
)

func newStatusBar(c *Core) *statusBar {
	sb := &statusBar{
		core:   c,
		height: 30,
	}
	sb.buttonsStepping = []layout.FlexChild{
		layout.Rigid(newStatusBarButton(sb.play, icons.Play).alt(
			func() bool {
				return sb.core.running
			}, sb.pause, icons.Pause).layout),
		layout.Rigid(newStatusBarButton(sb.step, icons.Step).layout),
		layout.Rigid(newStatusBarButton(sb.stepAhead, icons.SkipForward).layout),
	}
	sb.buttonsRight = []layout.FlexChild{
		layout.Rigid(newStatusBarButton(sb.zoomIn, icons.ZoomIn).layout),
		layout.Rigid(newStatusBarButton(sb.zoomOut, icons.ZoomOut).layout),
		layout.Rigid(newStatusBarButton(sb.menu, icons.Burger).highlightWhen(
			func() bool {
				return sb.showingPopup == popupMenu
			}).layout),
	}
	sb.buttonsRecord = []layout.FlexChild{
		layout.Rigid(newStatusBarButton(sb.stepBack, icons.Reverse).layout),
		layout.Rigid(newStatusBarButton(sb.skipBack, icons.SkipBackward).layout),
	}
	sb.buttonsExtra = []layout.FlexChild{
		layout.Rigid(newStatusBarButton(sb.extendedA, icons.ButtonA).layout),
		layout.Rigid(newStatusBarButton(sb.extendedB, icons.ButtonB).layout),
		layout.Rigid(newStatusBarButton(sb.extendedC, icons.ButtonC).layout),
		layout.Rigid(newStatusBarButton(sb.extendedD, icons.ButtonD).layout),
		layout.Rigid(newStatusBarButton(sb.extendedE, icons.ButtonE).layout),
		layout.Rigid(newStatusBarButton(sb.extendedF, icons.ButtonF).layout),
		layout.Rigid(newStatusBarButton(sb.extendedPlus, icons.Plus).layout),
		layout.Rigid(newStatusBarButton(sb.extendedMinus, icons.Minus).layout),
	}
	sb.rulesPopup = newRulesPopup(sb)
	sb.menuPopup = newMenuPopup(sb)
	return sb
}

func isValidButton(btn string) bool {
	return btn == "A" || btn == "B" || btn == "C" || btn == "D" || btn == "E" || btn == "F" || btn == "+" || btn == "-"
}

type popup int

const (
	popupNone = iota
	popupRule
	popupMenu
)

type statusBar struct {
	core            *Core
	height          unit.Dp
	rulesPopup      *rulesPopup
	menuPopup       *menuPopup
	showingPopup    popup
	top             int
	right           int
	ruleClickable   widget.Clickable
	ruleDims        layout.Dimensions
	stepDims        layout.Dimensions
	buttonsStepping []layout.FlexChild
	buttonsRight    []layout.FlexChild
	buttonsRecord   []layout.FlexChild
	buttonsExtra    []layout.FlexChild
}

func (sb *statusBar) play() {
	sb.core.stopShortcuts()
	sb.showingPopup = popupNone
	if !sb.core.running {
		sb.core.start()
	}
}

func (sb *statusBar) pause() {
	sb.core.stopShortcuts()
	sb.showingPopup = popupNone
	if sb.core.running {
		sb.core.stop()
	}
}

func (sb *statusBar) step() {
	sb.core.stopShortcuts()
	sb.showingPopup = popupNone
	sb.core.step()
}

func (sb *statusBar) stepAhead() {
	sb.core.stopShortcuts()
	sb.showingPopup = popupNone
	sb.core.stepAhead()
}

func (sb *statusBar) stepBack() {
	sb.core.stopShortcuts()
	sb.showingPopup = popupNone
	sb.core.stepBack()
}

func (sb *statusBar) skipBack() {
	sb.core.stopShortcuts()
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
	sb.core.stopShortcuts()
	sb.showHidePopup(popupMenu)
}

func (sb *statusBar) extendedA() {
	sb.core.stopShortcuts()
	sb.showingPopup = popupNone
	if scn, ok := sb.core.settings.ExtraButtons["A"]; ok {
		sb.core.runShortcut(scn)
	}
}

func (sb *statusBar) extendedB() {
	sb.core.stopShortcuts()
	sb.showingPopup = popupNone
	if scn, ok := sb.core.settings.ExtraButtons["B"]; ok {
		sb.core.runShortcut(scn)
	}
}

func (sb *statusBar) extendedC() {
	sb.core.stopShortcuts()
	sb.showingPopup = popupNone
	if scn, ok := sb.core.settings.ExtraButtons["C"]; ok {
		sb.core.runShortcut(scn)
	}
}

func (sb *statusBar) extendedD() {
	sb.core.stopShortcuts()
	sb.showingPopup = popupNone
	if scn, ok := sb.core.settings.ExtraButtons["D"]; ok {
		sb.core.runShortcut(scn)
	}
}

func (sb *statusBar) extendedE() {
	sb.core.stopShortcuts()
	sb.showingPopup = popupNone
	if scn, ok := sb.core.settings.ExtraButtons["E"]; ok {
		sb.core.runShortcut(scn)
	}
}

func (sb *statusBar) extendedF() {
	sb.core.stopShortcuts()
	sb.showingPopup = popupNone
	if scn, ok := sb.core.settings.ExtraButtons["F"]; ok {
		sb.core.runShortcut(scn)
	}
}

func (sb *statusBar) extendedPlus() {
	sb.core.stopShortcuts()
	sb.showingPopup = popupNone
	if scn, ok := sb.core.settings.ExtraButtons["+"]; ok {
		sb.core.runShortcut(scn)
	}
}

func (sb *statusBar) extendedMinus() {
	sb.core.stopShortcuts()
	sb.showingPopup = popupNone
	if scn, ok := sb.core.settings.ExtraButtons["-"]; ok {
		sb.core.runShortcut(scn)
	}
}

func (sb *statusBar) showHidePopup(p popup) {
	sb.core.stopShortcuts()
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
			sb.menuPopup.reset()
		}
	}
}

func (sb *statusBar) showPopups(gtx layout.Context) {
	switch sb.showingPopup {
	case popupRule:
		x := sb.stepDims.Size.X + gtx.Dp(unit.Dp(6)) + 3
		pgtx := gtx
		pgtx.Constraints = layout.Constraints{Max: image.Point{X: sb.ruleDims.Size.X - gtx.Dp(unit.Dp(12)) - 3, Y: sb.top}}
		macro := op.Record(gtx.Ops)
		dims := sb.rulesPopup.layout(pgtx)
		call := macro.Stop()
		stack := op.Offset(image.Point{X: x, Y: sb.top - dims.Size.Y}).Push(gtx.Ops)
		call.Add(gtx.Ops)
		stack.Pop()
	case popupMenu:
		pgtx := gtx
		pgtx.Constraints = layout.Constraints{Max: image.Point{X: sb.right, Y: sb.top}}
		macro := op.Record(gtx.Ops)
		dims := sb.menuPopup.layout(pgtx)
		x := sb.right - dims.Size.X
		call := macro.Stop()
		stack := op.Offset(image.Point{X: x, Y: sb.top - dims.Size.Y}).Push(gtx.Ops)
		call.Add(gtx.Ops)
		stack.Pop()
	}
}

func (sb *statusBar) layout(gtx layout.Context, windowRect clip.Rect) layout.Dimensions {
	if !sb.core.statusExpires.IsZero() {
		if time.Now().Before(sb.core.statusExpires) {
			gtx.Execute(op.InvalidateCmd{At: sb.core.statusExpires})
		} else {
			sb.core.status = ""
			sb.core.statusExpires = time.Time{}
		}
	}
	height := gtx.Dp(sb.height)
	size := image.Point{X: gtx.Constraints.Max.X, Y: height}
	sb.top = windowRect.Max.Y - height
	sb.right = windowRect.Max.X
	paint.FillShape(gtx.Ops, backgroundColor, clip.Rect{Max: size}.Op())
	horizontalLine(gtx, borderColor, size.X, 1)
	gtx.Constraints = layout.Exact(size)
	layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			switch {
			case sb.core.isShortcutsRunning():
				if sb.core.shortcutPrompt != nil {
					return sb.layoutShortcutPrompt(gtx)
				}
				st := sb.core.getShortcutsStatus()
				sb.stepDims = sb.label(gtx, "Running Shortcut"+st, text.Start)
			case sb.core.mode != noMode:
				sb.stepDims = sb.label(gtx, sb.core.modeDisplay(), text.Start)
			case sb.core.status != "":
				sb.stepDims = sb.label(gtx, sb.core.status, text.Start)
			default:
				var repeat string
				if sb.core.instrumentRepeat != nil && sb.core.instrumentRepeat.Found {
					repeat = " [Repeat found]"
				}
				perc := (float64(sb.core.changes.Load()) / float64(sb.core.gridHolder.grid.Width()*sb.core.gridHolder.grid.Height())) * 100.0
				if hz := sb.core.hertz.Load(); hz > 0 {
					if hz > 1000 {
						khz := float64(hz) / 1000.0
						sb.stepDims = sb.label(gtx, "Step: "+commas(strconv.FormatUint(sb.core.gridHolder.grid.StepCount.Load(), 10))+" ("+strconv.FormatFloat(perc, 'f', 1, 64)+"% "+strconv.FormatFloat(khz, 'f', 2, 64)+"kHz)"+repeat, text.Start)
					} else {
						sb.stepDims = sb.label(gtx, "Step: "+commas(strconv.FormatUint(sb.core.gridHolder.grid.StepCount.Load(), 10))+" ("+strconv.FormatFloat(perc, 'f', 1, 64)+"% "+strconv.FormatUint(hz, 10)+"Hz)"+repeat, text.Start)
					}
				} else {
					sb.stepDims = sb.label(gtx, "Step: "+commas(strconv.FormatUint(sb.core.gridHolder.grid.StepCount.Load(), 10))+" ("+strconv.FormatFloat(perc, 'f', 1, 64)+"%)"+repeat, text.Start)
				}
			}
			return sb.stepDims
		}),
		layout.Flexed(2, func(gtx layout.Context) layout.Dimensions {
			for sb.ruleClickable.Clicked(gtx) {
				sb.showHidePopup(popupRule)
			}
			return sb.ruleClickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				sb.ruleDims = sb.label(gtx, "Rule: "+sb.core.gridHolder.grid.Rule().Name(), text.Middle)
				return sb.ruleDims
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.End,
					Gap:       3,
				}.Layout(gtx, sb.buttons()...)
			})
		}),
	)
	return layout.Dimensions{Size: size}
}

func (sb *statusBar) buttons() []layout.FlexChild {
	result := make([]layout.FlexChild, 0, 14)
	for i, f := range sb.buttonsExtra {
		btn := string(rune('A' + i))
		if btn == "G" {
			btn = "+"
		}
		if btn == "H" {
			btn = "-"
		}
		if scn, ok := sb.core.settings.ExtraButtons[btn]; ok {
			if _, ok = sb.core.settings.Shortcuts[scn]; ok {
				result = append(result, f)
			}
		}
	}
	result = append(result, sb.buttonsStepping...)
	if sb.core.isRecording() {
		result = append(result, sb.buttonsRecord...)
	}
	return append(result, sb.buttonsRight...)
}

func (sb *statusBar) layoutShortcutPrompt(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Left: 8, Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Max
		sb.drawInsetBorder(gtx, image.Rectangle{Max: image.Point{X: size.X, Y: size.Y}})
		gtx.Constraints.Min.Y = size.Y
		gtx.Constraints.Max.Y = size.Y
		return layout.Inset{Left: 6, Right: 6, Top: 3}.Layout(gtx, sb.core.shortcutPrompt.layout)
	})
}

func (sb *statusBar) label(gtx layout.Context, s string, align text.Alignment) layout.Dimensions {
	return layout.Inset{Left: 8, Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Max
		sb.drawInsetBorder(gtx, image.Rectangle{Max: image.Point{X: size.X, Y: size.Y}})
		lbl := material.Label(theme, theme.TextSize, s)
		lbl.Alignment = align
		lbl.MaxLines = 1
		gtx.Constraints.Min.Y = size.Y
		gtx.Constraints.Max.Y = size.Y
		return layout.Inset{Left: 6, Right: 6, Top: 3}.Layout(gtx, lbl.Layout)
	})
}

var borderLightColor = color.NRGBA{R: 250, G: 250, B: 250, A: 255}

func (sb *statusBar) drawInsetBorder(gtx layout.Context, r image.Rectangle) {
	// top
	paint.FillShape(gtx.Ops, borderColor, clip.Rect{Min: image.Point{X: r.Min.X, Y: r.Min.Y + 4}, Max: image.Point{X: r.Max.X, Y: r.Min.Y + 5}}.Op())
	// left
	paint.FillShape(gtx.Ops, borderColor, clip.Rect{Min: image.Point{X: r.Min.X, Y: r.Min.Y + 4}, Max: image.Point{X: r.Min.X + 1, Y: r.Max.Y - 5}}.Op())
	// bottom
	paint.FillShape(gtx.Ops, borderLightColor, clip.Rect{Min: image.Point{X: r.Min.X, Y: r.Max.Y - 4}, Max: image.Point{X: r.Max.X, Y: r.Max.Y - 5}}.Op())
	// right
	paint.FillShape(gtx.Ops, borderLightColor, clip.Rect{Min: image.Point{X: r.Max.X - 1, Y: r.Min.Y + 4}, Max: image.Point{X: r.Max.X, Y: r.Max.Y - 5}}.Op())
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
	return layout.Inset{Top: 3, Bottom: 3}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		wh := min(gtx.Constraints.Max.Y, gtx.Constraints.Max.X)
		size := image.Point{X: wh, Y: wh}
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
				paint.FillShape(gtx.Ops, selectedFocusedColor, clip.UniformRRect(image.Rectangle{Max: size}, 4).Op(gtx.Ops))
			}
			defer op.Affine(
				f32.Affine2D{}.Scale(
					f32.Point{X: 0, Y: 0},
					f32.Point{X: float32(size.X) / float32(r.Dx()), Y: float32(size.Y) / float32(r.Dy())}),
			).Push(gtx.Ops).Pop()
			paint.NewImageOp(useImg).Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			return layout.Dimensions{Size: size}
		})
	})
}
