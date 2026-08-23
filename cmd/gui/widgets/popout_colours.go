package widgets

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"image/color"
)

type colorsPopout struct {
	parent     *menuPopup
	core       *Core
	aliveR     *numberInput[int]
	aliveG     *numberInput[int]
	aliveB     *numberInput[int]
	deadR      *numberInput[int]
	deadG      *numberInput[int]
	deadB      *numberInput[int]
	borderR    *numberInput[int]
	borderG    *numberInput[int]
	borderB    *numberInput[int]
	inputs     []*numberInput[int]
	chkBorders *checkbox
}

func newColorsPopout(p *menuPopup, c *Core) *colorsPopout {
	result := &colorsPopout{
		parent:     p,
		core:       c,
		chkBorders: newCheckBox("Show Borders", c.settings.CellBorders),
	}
	result.aliveR = newNumberInput(3, 0, 255, 16, func(v int) {
		result.colorChanged(0, 0, v)
	}).setValue(int(c.settings.CellAliveColor.R))
	result.aliveG = newNumberInput(3, 0, 255, 16, func(v int) {
		result.colorChanged(0, 1, v)
	}).setValue(int(c.settings.CellAliveColor.G))
	result.aliveB = newNumberInput(3, 0, 255, 16, func(v int) {
		result.colorChanged(0, 2, v)
	}).setValue(int(c.settings.CellAliveColor.B))
	result.deadR = newNumberInput(3, 0, 255, 16, func(v int) {
		result.colorChanged(1, 0, v)
	}).setValue(int(c.settings.CellDeadColor.R))
	result.deadG = newNumberInput(3, 0, 255, 16, func(v int) {
		result.colorChanged(1, 1, v)
	}).setValue(int(c.settings.CellDeadColor.G))
	result.deadB = newNumberInput(3, 0, 255, 16, func(v int) {
		result.colorChanged(1, 2, v)
	}).setValue(int(c.settings.CellDeadColor.B))
	result.borderR = newNumberInput(3, 0, 255, 16, func(v int) {
		result.colorChanged(2, 0, v)
	}).setValue(int(c.settings.CellBorderColor.R))
	result.borderG = newNumberInput(3, 0, 255, 16, func(v int) {
		result.colorChanged(2, 1, v)
	}).setValue(int(c.settings.CellBorderColor.G))
	result.borderB = newNumberInput(3, 0, 255, 16, func(v int) {
		result.colorChanged(2, 2, v)
	}).setValue(int(c.settings.CellBorderColor.B))
	result.inputs = []*numberInput[int]{
		result.aliveR, result.aliveG, result.aliveB,
		result.deadR, result.deadG, result.deadB,
		result.borderR, result.borderG, result.borderB,
	}
	c.settingsChangeListen(result.reset)
	return result
}

func (p *colorsPopout) reset() {
	p.aliveR.setValue(int(p.core.settings.CellAliveColor.R))
	p.aliveG.setValue(int(p.core.settings.CellAliveColor.G))
	p.aliveB.setValue(int(p.core.settings.CellAliveColor.B))
	p.deadR.setValue(int(p.core.settings.CellDeadColor.R))
	p.deadG.setValue(int(p.core.settings.CellDeadColor.G))
	p.deadB.setValue(int(p.core.settings.CellDeadColor.B))
	p.borderR.setValue(int(p.core.settings.CellBorderColor.R))
	p.borderG.setValue(int(p.core.settings.CellBorderColor.G))
	p.borderB.setValue(int(p.core.settings.CellBorderColor.B))
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
	return p.chkBorders.isFocused(gtx) ||
		p.aliveR.isFocused(gtx) || p.aliveG.isFocused(gtx) || p.aliveB.isFocused(gtx) ||
		p.deadR.isFocused(gtx) || p.deadG.isFocused(gtx) || p.deadB.isFocused(gtx) ||
		p.borderR.isFocused(gtx) || p.borderG.isFocused(gtx) || p.borderB.isFocused(gtx)
}

func (p *colorsPopout) layout(gtx layout.Context) layout.Dimensions {
	for _, inp := range p.inputs {
		inp.update(gtx)
	}
	if ok := p.chkBorders.Update(gtx); ok {
		p.core.setCellBorders(p.chkBorders.Checked())
	}
	labelMax := measureMaxText(gtx, font.Normal, "Alive cells", "Dead cells", "Cell Border").Size.X
	return popoutLayout(gtx, flexVertical(4,
		rigid(flexHorizontal(20,
			rigidLabel("Alive cells", text.End, 0, labelMax),
			rigidLabel("R:", 0, 0, 0),
			flexed(p.aliveR.layout),
			rigidLabel("G:", 0, 0, 0),
			flexed(p.aliveG.layout),
			rigidLabel("B:", 0, 0, 0),
			flexed(p.aliveB.layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Dead cells", text.End, 0, labelMax),
			rigidLabel("R:", 0, 0, 0),
			flexed(p.deadR.layout),
			rigidLabel("G:", 0, 0, 0),
			flexed(p.deadG.layout),
			rigidLabel("B:", 0, 0, 0),
			flexed(p.deadB.layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Cell Border", text.End, 0, labelMax),
			rigidLabel("R:", 0, 0, 0),
			flexed(p.borderR.layout),
			rigidLabel("G:", 0, 0, 0),
			flexed(p.borderG.layout),
			rigidLabel("B:", 0, 0, 0),
			flexed(p.borderB.layout),
		)),
		rigid(p.chkBorders.Layout),
	))
}
