package gui

import (
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/gui/controls"
	"github.com/marrow16/gogol/gui/icons"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"image"
	"image/color"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func newStatusBar(c *Core) *statusBar {
	sb := &statusBar{
		core:          c,
		height:        30,
		btnRotate:     controls.NewFlatButton(theme, "Rotate "+patterns.Rotate0.String()).Margin(layout.Inset{Top: 2, Bottom: 1}),
		btnInterlace:  controls.NewFlatButton(theme, "Overlay").Margin(layout.Inset{Top: 2, Bottom: 1}),
		btnModeCancel: controls.NewIconButton(icons.Close, -3),
		btnCapture:    controls.NewFlatButton(theme, "Capture").Margin(layout.Inset{Top: 2, Bottom: 1}).Colors(white, &button),
		btnFill:       controls.NewFlatButton(theme, "Fill").Margin(layout.Inset{Top: 2, Bottom: 1}),
		btnClear:      controls.NewFlatButton(theme, "Clear").Margin(layout.Inset{Top: 2, Bottom: 1}),
		btnPlace:      controls.NewFlatButton(theme, "Place").Margin(layout.Inset{Top: 2, Bottom: 1}).Colors(white, &button),
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
		layout.Rigid(newStatusBarButton(func() { sb.extended("A") }, icons.ButtonA).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("B") }, icons.ButtonB).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("C") }, icons.ButtonC).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("D") }, icons.ButtonD).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("E") }, icons.ButtonE).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("F") }, icons.ButtonF).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("G") }, icons.ButtonG).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("H") }, icons.ButtonH).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("I") }, icons.ButtonI).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("J") }, icons.ButtonJ).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("K") }, icons.ButtonK).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("L") }, icons.ButtonL).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("M") }, icons.ButtonM).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("N") }, icons.ButtonN).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("O") }, icons.ButtonO).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("P") }, icons.ButtonP).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("R") }, icons.ButtonR).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("S") }, icons.ButtonS).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("T") }, icons.ButtonT).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("U") }, icons.ButtonU).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("V") }, icons.ButtonV).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("W") }, icons.ButtonW).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("X") }, icons.ButtonX).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("Y") }, icons.ButtonY).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("Z") }, icons.ButtonZ).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("-") }, icons.Minus).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("+") }, icons.Plus).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("0") }, icons.Button0).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("1") }, icons.Button1).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("2") }, icons.Button2).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("3") }, icons.Button3).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("4") }, icons.Button4).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("5") }, icons.Button5).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("6") }, icons.Button6).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("7") }, icons.Button7).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("8") }, icons.Button8).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("9") }, icons.Button9).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("Γ") }, icons.ButtonGamma).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("Δ") }, icons.ButtonDelta).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("Θ") }, icons.ButtonTheta).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("Λ") }, icons.ButtonLambda).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("Ξ") }, icons.ButtonXi).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("Π") }, icons.ButtonPi).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("Σ") }, icons.ButtonSigma).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("Φ") }, icons.ButtonPhi).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("Ψ") }, icons.ButtonPsi).layout),
		layout.Rigid(newStatusBarButton(func() { sb.extended("Ω") }, icons.ButtonOmega).layout),
	}
	sb.rulesPopup = newRulesPopup(sb)
	sb.menuPopup = newMenuPopup(sb)
	return sb
}

func isValidButton(btn string) bool {
	return utf8.RuneCountInString(btn) == 1 && strings.Contains("ABCDEFGHIJKLMNOPRSTUVWXYZ+-0123456789ΓΔΘΛΞΠΣΦΨΩ", btn)
}

type popup int

const (
	popupNone popup = iota
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
	middleClickable controls.Clickable
	ruleDims        layout.Dimensions
	stepDims        layout.Dimensions
	buttonsStepping []layout.FlexChild
	buttonsRight    []layout.FlexChild
	buttonsRecord   []layout.FlexChild
	buttonsExtra    []layout.FlexChild
	btnRotate       *controls.FlatButton
	btnInterlace    *controls.FlatButton
	btnModeCancel   *controls.IconButton
	btnCapture      *controls.FlatButton
	btnClear        *controls.FlatButton
	btnFill         *controls.FlatButton
	btnPlace        *controls.FlatButton
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

func (sb *statusBar) extended(btn string) {
	sb.core.stopShortcuts()
	sb.showingPopup = popupNone
	if scn, ok := sb.core.settings.ExtraButtons[btn]; ok {
		sb.core.runShortcut(scn)
	}
}

func (sb *statusBar) popupShowing() bool {
	if sb.showingPopup == popupNone {
		return false
	}
	return true
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
	switch {
	case sb.showingPopup == popupRule:
		x := sb.stepDims.Size.X + gtx.Dp(unit.Dp(6)) + 3
		pgtx := gtx
		pgtx.Constraints = layout.Constraints{Max: image.Point{X: sb.ruleDims.Size.X - gtx.Dp(unit.Dp(12)) - 3, Y: sb.top}}
		macro := op.Record(gtx.Ops)
		dims := sb.rulesPopup.layout(pgtx)
		call := macro.Stop()
		stack := op.Offset(image.Point{X: x, Y: sb.top - dims.Size.Y}).Push(gtx.Ops)
		call.Add(gtx.Ops)
		stack.Pop()
	case sb.showingPopup == popupMenu:
		pgtx := gtx
		pgtx.Constraints = layout.Constraints{Max: image.Point{X: sb.right, Y: sb.top}}
		macro := op.Record(gtx.Ops)
		dims := sb.menuPopup.layout(pgtx)
		x := sb.right - dims.Size.X
		call := macro.Stop()
		stack := op.Offset(image.Point{X: x, Y: sb.top - dims.Size.Y}).Push(gtx.Ops)
		call.Add(gtx.Ops)
		stack.Pop()
	case sb.core.shortcutRunning && sb.core.shortcutPromptMenu != nil:
		pgtx := gtx
		pgtx.Constraints = layout.Constraints{Max: image.Point{X: sb.stepDims.Size.X - 16, Y: sb.top}}
		macro := op.Record(gtx.Ops)
		dims := sb.core.shortcutPromptMenu.Layout(pgtx)
		call := macro.Stop()
		stack := op.Offset(image.Point{X: 12, Y: sb.top - dims.Size.Y}).Push(gtx.Ops)
		call.Add(gtx.Ops)
		stack.Pop()
	case sb.core.shortcutRunning && sb.core.shortcutPromptDialog != nil:
		pgtx := gtx
		pgtx.Constraints = layout.Constraints{Min: image.Point{X: sb.stepDims.Size.X - 16}, Max: image.Point{X: sb.stepDims.Size.X + sb.ruleDims.Size.X, Y: sb.top}}
		macro := op.Record(gtx.Ops)
		dims := sb.core.shortcutPromptDialog.Layout(pgtx)
		call := macro.Stop()
		stack := op.Offset(image.Point{X: 12, Y: sb.top - dims.Size.Y}).Push(gtx.Ops)
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
		sb.leftArea(gtx),
		sb.middleArea(gtx),
		sb.rightArea(),
	)
	return layout.Dimensions{Size: size}
}

func (sb *statusBar) leftArea(gtx layout.Context) layout.FlexChild {
	wt := float32(1.0)
	if sb.core.mode != noMode {
		wt = 0.5
		if sb.btnModeCancel.Clicked(gtx) {
			sb.core.clearMode()
		}
	}
	return flexed(func(gtx layout.Context) layout.Dimensions {
		switch {
		case sb.core.isShortcutsRunning():
			if sb.core.shortcutPrompt != nil {
				return sb.inset(gtx, sb.core.shortcutPrompt.Layout)
			} else if sb.core.shortcutPromptMenu != nil {
				sb.stepDims = sb.insetLabel(gtx, sb.core.shortcutPromptMenu.prompt, text.Start)
			} else if sb.core.shortcutPromptDialog != nil {
				sb.stepDims = sb.insetLabel(gtx, sb.core.shortcutPromptDialog.prompt, text.Start)
			} else {
				st := sb.core.getShortcutsStatus()
				sb.stepDims = sb.insetLabel(gtx, "Running Shortcut"+st, text.Start)
			}
		case sb.core.mode == editMode:
			return sb.inset(gtx, flexHorizontal(8,
				rigid(sb.btnModeCancel.Layout),
				rigid(verticalCenter(label("Edit Mode"))),
			))
		case sb.core.mode == placePatternMode:
			return sb.inset(gtx, flexHorizontal(8,
				rigid(sb.btnModeCancel.Layout),
				rigid(verticalCenter(label("Place Pattern Mode"))),
			))
		case sb.core.mode == heatMapMode:
			return sb.inset(gtx, flexHorizontal(8,
				rigid(sb.btnModeCancel.Layout),
				rigid(verticalCenter(label("Heat Map Display"))),
			))
		case sb.core.status != "":
			sb.stepDims = sb.insetLabel(gtx, sb.core.status, text.Start)
		default:
			var repeat string
			if sb.core.instrumentRepeat != nil && sb.core.instrumentRepeat.Found {
				repeat = " [Repeat found]"
			}
			perc := (float64(sb.core.changes.Load()) / float64(sb.core.gridHolder.grid.Width()*sb.core.gridHolder.grid.Height())) * 100.0
			if hz := sb.core.hertz.Load(); hz > 0 {
				if hz > 1000 {
					khz := float64(hz) / 1000.0
					sb.stepDims = sb.insetLabel(gtx, "Step: "+commas(strconv.FormatUint(sb.core.gridHolder.grid.StepCount.Load(), 10))+" ("+strconv.FormatFloat(perc, 'f', 1, 64)+"% "+strconv.FormatFloat(khz, 'f', 2, 64)+"kHz)"+repeat, text.Start)
				} else {
					sb.stepDims = sb.insetLabel(gtx, "Step: "+commas(strconv.FormatUint(sb.core.gridHolder.grid.StepCount.Load(), 10))+" ("+strconv.FormatFloat(perc, 'f', 1, 64)+"% "+strconv.FormatUint(hz, 10)+"Hz)"+repeat, text.Start)
				}
			} else {
				sb.stepDims = sb.insetLabel(gtx, "Step: "+commas(strconv.FormatUint(sb.core.gridHolder.grid.StepCount.Load(), 10))+" ("+strconv.FormatFloat(perc, 'f', 1, 64)+"%)"+repeat, text.Start)
			}
		}
		return sb.stepDims
	}, wt)
}

func (sb *statusBar) middleArea(gtx layout.Context) layout.FlexChild {
	switch sb.core.mode {
	case editMode:
		if sb.btnRotate.Clicked(gtx) {
			rot := sb.core.gridHolder.editor.patternRotation + 1
			if rot > patterns.Rotate270 {
				rot = patterns.Rotate0
			}
			sb.core.gridHolder.editor.patternRotation = rot
		}
		if sb.btnInterlace.Clicked(gtx) {
			sb.core.gridHolder.editor.patternInterlace = !sb.core.gridHolder.editor.patternInterlace
		}
		if sb.btnCapture.Clicked(gtx) {
			sb.core.gridHolder.editor.endMarking(true)
		}
		if sb.btnClear.Clicked(gtx) {
			sb.core.gridHolder.editor.markingFillClear(false)
		}
		if sb.btnFill.Clicked(gtx) {
			sb.core.gridHolder.editor.markingFillClear(true)
		}
		if sb.core.gridHolder.editor.patternInterlace {
			sb.btnInterlace.Label("Interlaced")
		} else {
			sb.btnInterlace.Label("Overlayed")
		}
		sb.btnRotate.Label("Rotate " + sb.core.gridHolder.editor.patternRotation.String())
		return flexed(func(gtx layout.Context) layout.Dimensions {
			return sb.inset(gtx, flexHorizontalSpaced(0, layout.SpaceBetween,
				rigid(flexHorizontal(8,
					rigid(verticalCenter(label("@ X:"+strconv.Itoa(sb.core.gridHolder.editor.col)+" Y:"+strconv.Itoa(sb.core.gridHolder.editor.row)))),
					conditionalRigid(sb.core.gridHolder.editor.marking, flexHorizontal(8,
						rigid(verticalCenter(label("  Marking W:"+strconv.Itoa(max(sb.core.gridHolder.editor.col, sb.core.gridHolder.editor.markStartCol)-min(sb.core.gridHolder.editor.col, sb.core.gridHolder.editor.markStartCol)+1)+
							" H:"+strconv.Itoa(max(sb.core.gridHolder.editor.row, sb.core.gridHolder.editor.markStartRow)-min(sb.core.gridHolder.editor.row, sb.core.gridHolder.editor.markStartRow)+1)))),
					), nil),
				)),
				conditionalRigid(sb.core.gridHolder.editor.marking,
					flexHorizontal(8,
						rigid(sb.btnCapture.Layout),
						rigid(sb.btnClear.Layout),
						rigid(sb.btnFill.Layout),
					),
					flexHorizontal(8,
						rigid(sb.btnRotate.Layout),
						rigid(sb.btnInterlace.Layout),
					)),
			))
		}, 2)
	case placePatternMode:
		if sb.btnRotate.Clicked(gtx) && sb.core.gridHolder.overlay != nil {
			rot := sb.core.gridHolder.overlay.rotation + 1
			if rot > patterns.Rotate270 {
				rot = patterns.Rotate0
			}
			sb.core.gridHolder.overlay.rotation = rot
			sb.core.gridHolder.overlay.moved = true
		}
		if sb.btnInterlace.Clicked(gtx) && sb.core.gridHolder.overlay != nil {
			sb.core.gridHolder.overlay.interlaced = !sb.core.gridHolder.overlay.interlaced
			sb.core.gridHolder.overlay.moved = true
		}
		if sb.btnPlace.Clicked(gtx) && sb.core.gridHolder.overlay != nil {
			sb.core.placePattern(gtx)
		}
		lbl := ""
		if sb.core.gridHolder.overlay != nil {
			lbl = "@ X:" + strconv.Itoa(sb.core.gridHolder.overlay.col) + " Y:" + strconv.Itoa(sb.core.gridHolder.overlay.row)
			if sb.core.gridHolder.overlay.interlaced {
				sb.btnInterlace.Label("Interlaced")
			} else {
				sb.btnInterlace.Label("Overlayed")
			}
			sb.btnRotate.Label("Rotate " + sb.core.gridHolder.overlay.rotation.String())
		}
		return flexed(func(gtx layout.Context) layout.Dimensions {
			return sb.inset(gtx, flexHorizontalSpaced(0, layout.SpaceBetween,
				rigid(verticalCenter(label(lbl))),
				rigid(flexHorizontal(8,
					rigid(sb.btnPlace.Layout),
					rigid(sb.btnRotate.Layout),
					rigid(sb.btnInterlace.Layout),
				)),
			))
		}, 2)
	case heatMapMode:
		if sb.middleClickable.Clicked(gtx) {
			if sb.core.instrumentHeatMap != nil && sb.core.instrumentHeatMap.Type() == logic.AllHeatMapper {
				sb.core.cycleShowingHeatMapType()
				sb.core.showHeatMap()
			}
		}
		return flexed(func(gtx layout.Context) layout.Dimensions {
			return sb.middleClickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				hmt := logic.NoHeatMapper
				if sb.core.instrumentHeatMap != nil {
					hmt = sb.core.instrumentHeatMap.Type()
					if hmt == logic.AllHeatMapper {
						hmt = sb.core.showHeatMapperType
					}
				}
				sb.ruleDims = sb.insetLabel(gtx, hmt.String(), text.Middle)
				return sb.ruleDims
			})
		}, 2)
	default:
		if sb.middleClickable.Clicked(gtx) {
			sb.showHidePopup(popupRule)
		}
		return flexed(func(gtx layout.Context) layout.Dimensions {
			return sb.middleClickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				sb.ruleDims = sb.insetLabel(gtx, "Rule: "+sb.core.gridHolder.grid.Rule().Name(), text.Middle)
				return sb.ruleDims
			})
		}, 2)
	}
}

var extraButtonIndices = map[int]string{
	0: "A", 1: "B", 2: "C", 3: "D", 4: "E", 5: "F",
	6: "G", 7: "H", 8: "I", 9: "J", 10: "K", 11: "L",
	12: "M", 13: "N", 14: "O", 15: "P", 16: "R", 17: "S",
	18: "T", 19: "U", 20: "V", 21: "W", 22: "X", 23: "Y", 24: "Z",
	25: "-", 26: "+",
	27: "0", 28: "1", 29: "2", 30: "3", 31: "4", 32: "5", 33: "6", 34: "7", 35: "8", 36: "9",
	37: "Γ", 38: "Δ", 39: "Θ", 40: "Λ", 41: "Ξ", 42: "Π", 43: "Σ", 44: "Φ", 45: "Ψ", 46: "Ω",
}

func (sb *statusBar) rightArea() layout.FlexChild {
	buttons := make([]layout.FlexChild, 0, 14)
	if sb.core.mode == noMode && !sb.core.isShortcutsInteractive() {
		for i, f := range sb.buttonsExtra {
			btn := extraButtonIndices[i]
			if scn, ok := sb.core.settings.ExtraButtons[btn]; ok {
				if _, ok = sb.core.settings.Shortcuts[scn]; ok {
					buttons = append(buttons, f)
				}
			}
		}
		buttons = append(buttons, sb.buttonsStepping...)
		if sb.core.isRecording() {
			buttons = append(buttons, sb.buttonsRecord...)
		}
	}
	buttons = append(buttons, sb.buttonsRight...)
	return rigid(inset(0, 0, 0, 8, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.End,
			Gap:       3,
		}.Layout(gtx, buttons...)
	}))
}

func (sb *statusBar) inset(gtx layout.Context, w layout.Widget) layout.Dimensions {
	return layout.Inset{Left: 6, Right: 6}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Max
		sb.drawInsetBorder(gtx, image.Rectangle{Max: image.Point{X: size.X, Y: size.Y}})
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
		return layout.Inset{Top: 3, Bottom: 2, Left: 6, Right: 6}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				rigid(w),
			)
		})
	})
}

func (sb *statusBar) insetLabel(gtx layout.Context, s string, align text.Alignment) layout.Dimensions {
	return layout.Inset{Left: 6, Right: 6}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Max
		sb.drawInsetBorder(gtx, image.Rectangle{Max: image.Point{X: size.X, Y: size.Y}})
		lbl := material.Label(theme, theme.TextSize, s)
		lbl.Alignment = align
		lbl.MaxLines = 1
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
		return layout.Inset{Top: 3, Bottom: 3, Left: 6, Right: 6}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				rigid(verticalCenter(lbl.Layout)),
			)
		})
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
	clickable controls.Clickable
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
