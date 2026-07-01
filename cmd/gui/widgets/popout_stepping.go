package widgets

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type steppingPopout struct {
	parent            *menuPopup
	core              *Core
	stepDelay         *numberInput[int]
	stepAheadSize     *numberInput[int]
	stepAheadSnapshot *widget.Bool
}

func newSteppingPopout(p *menuPopup, c *Core) *steppingPopout {
	result := &steppingPopout{
		parent:            p,
		core:              c,
		stepAheadSnapshot: &widget.Bool{Value: c.settings.StepAheadSnapshot},
	}
	result.stepDelay = newNumberInput(c.theme, 4, 0, 2000, 10, func(v int) {
		if v >= 0 {
			result.core.settings.StepDelay = v
		} else {
			result.stepDelay.setValue(0)
		}
	}).setValue(int(c.settings.StepDelay))
	result.stepAheadSize = newNumberInput(c.theme, 4, 0, 9999, 100, func(v int) {
		result.core.settings.StepAheadBy = v
	}).setValue(int(c.settings.StepAheadBy))
	return result
}

func (p *steppingPopout) reset() {
	p.stepDelay.setValue(p.core.settings.StepDelay)
	p.stepAheadSize.setValue(p.core.settings.StepAheadBy)
	p.stepAheadSnapshot.Value = p.core.settings.StepAheadSnapshot
}

func (p *steppingPopout) hasFocus(gtx layout.Context) bool {
	return p.stepDelay.isFocused(gtx) || p.stepAheadSize.isFocused(gtx) || gtx.Focused(p.stepAheadSnapshot)
}

func (p *steppingPopout) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	p.stepDelay.update(gtx)
	p.stepAheadSize.update(gtx)
	if ok := p.stepAheadSnapshot.Update(gtx); ok {
		p.core.settings.StepAheadSnapshot = p.stepAheadSnapshot.Value
	}
	labelMax := measureMaxText(gtx, theme, font.Normal, "Step delay (ms): ", "Step ahead size: ")
	return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8), Top: unit.Dp(8), Bottom: unit.Dp(4)}.
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = labelMax.Size.X
								return rightLabel(p.core.theme, "Step delay (ms):").Layout(gtx)
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return p.stepDelay.layout(gtx)
							}),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = labelMax.Size.X
								return rightLabel(p.core.theme, "Step ahead size:").Layout(gtx)
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return p.stepAheadSize.layout(gtx)
							}),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						chkBox := material.CheckBox(p.core.theme, p.stepAheadSnapshot, "Snapshot on step ahead")
						chkBox.TextSize = unit.Sp(16)
						chkBox.Size = 18
						return chkBox.Layout(gtx)
					})
				}),
			)
		})
}
