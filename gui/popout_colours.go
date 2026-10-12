package gui

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"github.com/marrow16/gogol/gui/controls"
	"image/color"
)

type colorsPopout struct {
	parent     *menuPopup
	core       *Core
	aliveR     *controls.NumberInput[int]
	aliveG     *controls.NumberInput[int]
	aliveB     *controls.NumberInput[int]
	deadR      *controls.NumberInput[int]
	deadG      *controls.NumberInput[int]
	deadB      *controls.NumberInput[int]
	borderR    *controls.NumberInput[int]
	borderG    *controls.NumberInput[int]
	borderB    *controls.NumberInput[int]
	inputs     []*controls.NumberInput[int]
	chkBorders *controls.Checkbox
}

func newColorsPopout(p *menuPopup, c *Core) *colorsPopout {
	result := &colorsPopout{
		parent:     p,
		core:       c,
		chkBorders: controls.NewCheckBox(theme, "Show Borders", c.settings.CellBorders),
	}
	result.aliveR = controls.NewNumberInput(theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(0, 0, v)
	}).SetValue(int(c.settings.CellAliveColor.R))
	result.aliveG = controls.NewNumberInput(theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(0, 1, v)
	}).SetValue(int(c.settings.CellAliveColor.G))
	result.aliveB = controls.NewNumberInput(theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(0, 2, v)
	}).SetValue(int(c.settings.CellAliveColor.B))
	result.deadR = controls.NewNumberInput(theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(1, 0, v)
	}).SetValue(int(c.settings.CellDeadColor.R))
	result.deadG = controls.NewNumberInput(theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(1, 1, v)
	}).SetValue(int(c.settings.CellDeadColor.G))
	result.deadB = controls.NewNumberInput(theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(1, 2, v)
	}).SetValue(int(c.settings.CellDeadColor.B))
	result.borderR = controls.NewNumberInput(theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(2, 0, v)
	}).SetValue(int(c.settings.CellBorderColor.R))
	result.borderG = controls.NewNumberInput(theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(2, 1, v)
	}).SetValue(int(c.settings.CellBorderColor.G))
	result.borderB = controls.NewNumberInput(theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(2, 2, v)
	}).SetValue(int(c.settings.CellBorderColor.B))
	result.inputs = []*controls.NumberInput[int]{
		result.aliveR, result.aliveG, result.aliveB,
		result.deadR, result.deadG, result.deadB,
		result.borderR, result.borderG, result.borderB,
	}
	c.settingsChangeListen(result.reset)
	return result
}

func (p *colorsPopout) reset() {
	p.aliveR.SetValue(int(p.core.settings.CellAliveColor.R))
	p.aliveG.SetValue(int(p.core.settings.CellAliveColor.G))
	p.aliveB.SetValue(int(p.core.settings.CellAliveColor.B))
	p.deadR.SetValue(int(p.core.settings.CellDeadColor.R))
	p.deadG.SetValue(int(p.core.settings.CellDeadColor.G))
	p.deadB.SetValue(int(p.core.settings.CellDeadColor.B))
	p.borderR.SetValue(int(p.core.settings.CellBorderColor.R))
	p.borderG.SetValue(int(p.core.settings.CellBorderColor.G))
	p.borderB.SetValue(int(p.core.settings.CellBorderColor.B))
	p.chkBorders.SetChecked(p.core.settings.CellBorders)
}

func (p *colorsPopout) colorChanged(c int, component int, v int) {
	switch c {
	case 0:
		//alive
		if nc, changed := p.colorComponentChanged(p.core.settings.CellAliveColor, component, v); changed {
			p.core.settings.CellAliveColor = nc
			p.core.gridHolder.grid.Draw()
		}
	case 1:
		//dead
		if nc, changed := p.colorComponentChanged(p.core.settings.CellDeadColor, component, v); changed {
			p.core.settings.CellDeadColor = nc
			p.core.gridHolder.grid.Draw()
		}
	case 2:
		//border
		if nc, changed := p.colorComponentChanged(p.core.settings.CellBorderColor, component, v); changed {
			p.core.settings.CellBorderColor = nc
			p.core.gridHolder.rebuild()
		}
	}
}

func (p *colorsPopout) colorComponentChanged(c color.NRGBA, component int, v int) (color.NRGBA, bool) {
	switch component {
	case 0:
		//red
		if int(c.R) != v {
			return color.NRGBA{R: uint8(v), G: c.G, B: c.B, A: c.A}, true
		}
	case 1:
		//green
		if int(c.G) != v {
			return color.NRGBA{R: c.R, G: uint8(v), B: c.B, A: c.A}, true
		}
	case 2:
		//blue
		if int(c.B) != v {
			return color.NRGBA{R: c.R, G: c.G, B: uint8(v), A: c.A}, true
		}
	}
	return c, false
}

func (p *colorsPopout) hasFocus(gtx layout.Context) bool {
	return p.chkBorders.IsFocused(gtx) ||
		p.aliveR.IsFocused(gtx) || p.aliveG.IsFocused(gtx) || p.aliveB.IsFocused(gtx) ||
		p.deadR.IsFocused(gtx) || p.deadG.IsFocused(gtx) || p.deadB.IsFocused(gtx) ||
		p.borderR.IsFocused(gtx) || p.borderG.IsFocused(gtx) || p.borderB.IsFocused(gtx)
}

func (p *colorsPopout) layout(gtx layout.Context) layout.Dimensions {
	for _, inp := range p.inputs {
		inp.Update(gtx)
	}
	if ok := p.chkBorders.Update(gtx); ok {
		p.core.setCellBorders(p.chkBorders.Checked())
	}
	labelMax := measureMaxText(gtx, font.Normal, "Alive cells", "Dead cells", "Cell border")
	return popoutLayout(gtx, flexVertical(4,
		rigid(flexHorizontal(20,
			rigidLabel("Alive cells", text.End, 0, labelMax),
			rigidLabel("R:", 0, 0, 0),
			flexed(p.aliveR.Layout),
			rigidLabel("G:", 0, 0, 0),
			flexed(p.aliveG.Layout),
			rigidLabel("B:", 0, 0, 0),
			flexed(p.aliveB.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Dead cells", text.End, 0, labelMax),
			rigidLabel("R:", 0, 0, 0),
			flexed(p.deadR.Layout),
			rigidLabel("G:", 0, 0, 0),
			flexed(p.deadG.Layout),
			rigidLabel("B:", 0, 0, 0),
			flexed(p.deadB.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Cell border", text.End, 0, labelMax),
			rigidLabel("R:", 0, 0, 0),
			flexed(p.borderR.Layout),
			rigidLabel("G:", 0, 0, 0),
			flexed(p.borderG.Layout),
			rigidLabel("B:", 0, 0, 0),
			flexed(p.borderB.Layout),
		)),
		rigid(p.chkBorders.Layout),
	))
}
