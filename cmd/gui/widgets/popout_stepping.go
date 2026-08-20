package widgets

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
)

type steppingPopout struct {
	parent               *menuPopup
	core                 *Core
	stepDelay            *numberInput[int]
	stepAheadSize        *numberInput[int]
	chkStepAheadSnapshot *checkbox
}

func newSteppingPopout(p *menuPopup, c *Core) *steppingPopout {
	result := &steppingPopout{
		parent:               p,
		core:                 c,
		chkStepAheadSnapshot: newCheckBox("Snapshot on step ahead", c.settings.StepAheadSnapshot),
	}
	result.stepDelay = newNumberInput(4, 0, 2000, 10, func(v int) {
		if v >= 0 {
			result.core.settings.StepDelay = v
		} else {
			result.stepDelay.setValue(0)
		}
	}).setValue(int(c.settings.StepDelay))
	result.stepAheadSize = newNumberInput(4, 0, 9999, 100, func(v int) {
		result.core.settings.StepAheadBy = v
	}).setValue(int(c.settings.StepAheadBy))
	return result
}

func (p *steppingPopout) reset() {
	p.stepDelay.setValue(p.core.settings.StepDelay)
	p.stepAheadSize.setValue(p.core.settings.StepAheadBy)
	p.chkStepAheadSnapshot.SetChecked(p.core.settings.StepAheadSnapshot)
}

func (p *steppingPopout) hasFocus(gtx layout.Context) bool {
	return p.stepDelay.isFocused(gtx) || p.stepAheadSize.isFocused(gtx) || p.chkStepAheadSnapshot.isFocused(gtx)
}

func (p *steppingPopout) layout(gtx layout.Context) layout.Dimensions {
	p.stepDelay.update(gtx)
	p.stepAheadSize.update(gtx)
	if ok := p.chkStepAheadSnapshot.Update(gtx); ok {
		p.core.settings.StepAheadSnapshot = p.chkStepAheadSnapshot.Checked()
	}
	labelMax := measureMaxText(gtx, font.Normal, "Step delay (ms): ", "Step ahead size: ").Size.X
	return layout.Inset{Left: 8, Right: 8, Top: 8, Bottom: 4}.Layout(gtx, flexVertical(8,
		layout.Rigid(flexHorizontal(20,
			rigidLabel("Step delay (ms):", text.End, 0, labelMax),
			layout.Flexed(1, p.stepDelay.layout),
		)),
		layout.Rigid(flexHorizontal(20,
			rigidLabel("Step ahead size:", text.End, 0, labelMax),
			layout.Flexed(1, p.stepAheadSize.layout),
		)),
		layout.Rigid(p.chkStepAheadSnapshot.Layout),
	))
}
