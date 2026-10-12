package gui

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"github.com/marrow16/gogol/gui/controls"
)

type steppingPopout struct {
	parent               *menuPopup
	core                 *Core
	stepDelay            *controls.NumberInput[int]
	stepAheadSize        *controls.NumberInput[int]
	chkStepAheadSnapshot *controls.Checkbox
}

func newSteppingPopout(p *menuPopup, c *Core) *steppingPopout {
	result := &steppingPopout{
		parent:               p,
		core:                 c,
		chkStepAheadSnapshot: controls.NewCheckBox(theme, "Snapshot on step ahead", c.settings.StepAheadSnapshot),
	}
	result.stepDelay = controls.NewNumberInput(theme, 4, 0, 2000, 10, func(v int) {
		if v >= 0 {
			result.core.settings.StepDelay = v
		} else {
			result.stepDelay.SetValue(0)
		}
	}).SetValue(c.settings.StepDelay)
	result.stepAheadSize = controls.NewNumberInput(theme, 4, 1, 9999, 100, func(v int) {
		result.core.settings.StepAheadBy = v
	}).SetValue(c.settings.StepAheadBy)
	return result
}

func (p *steppingPopout) reset() {
	p.stepDelay.SetValue(p.core.settings.StepDelay)
	p.stepAheadSize.SetValue(p.core.settings.StepAheadBy)
	p.chkStepAheadSnapshot.SetChecked(p.core.settings.StepAheadSnapshot)
}

func (p *steppingPopout) hasFocus(gtx layout.Context) bool {
	return p.stepDelay.IsFocused(gtx) || p.stepAheadSize.IsFocused(gtx) || p.chkStepAheadSnapshot.IsFocused(gtx)
}

func (p *steppingPopout) layout(gtx layout.Context) layout.Dimensions {
	p.stepDelay.Update(gtx)
	p.stepAheadSize.Update(gtx)
	if ok := p.chkStepAheadSnapshot.Update(gtx); ok {
		p.core.settings.StepAheadSnapshot = p.chkStepAheadSnapshot.Checked()
	}
	labelMax := measureMaxText(gtx, font.Normal, "Step delay (ms): ", "Step ahead size: ")
	return popoutLayout(gtx, flexVertical(8,
		rigid(flexHorizontal(20,
			rigidLabel("Step delay (ms):", text.End, 0, labelMax),
			flexed(p.stepDelay.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Step ahead size:", text.End, 0, labelMax),
			flexed(p.stepAheadSize.Layout),
		)),
		rigid(p.chkStepAheadSnapshot.Layout),
	))
}
