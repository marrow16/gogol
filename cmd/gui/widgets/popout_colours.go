package widgets

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/cmd/gui/settings"
	"image/color"
)

type popout interface {
	layout(gtx layout.Context, theme *material.Theme) layout.Dimensions
	hasFocus(gtx layout.Context) bool
	reset()
}

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
		chkBorders: newCheckBox(c.theme, "Show Borders", c.settings.CellBorders),
	}
	result.aliveR = newNumberInput(c.theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(0, 0, v)
	}).setValue(int(c.settings.CellAliveColor.R))
	result.aliveG = newNumberInput(c.theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(0, 1, v)
	}).setValue(int(c.settings.CellAliveColor.G))
	result.aliveB = newNumberInput(c.theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(0, 2, v)
	}).setValue(int(c.settings.CellAliveColor.B))
	result.deadR = newNumberInput(c.theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(1, 0, v)
	}).setValue(int(c.settings.CellDeadColor.R))
	result.deadG = newNumberInput(c.theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(1, 1, v)
	}).setValue(int(c.settings.CellDeadColor.G))
	result.deadB = newNumberInput(c.theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(1, 2, v)
	}).setValue(int(c.settings.CellDeadColor.B))
	result.borderR = newNumberInput(c.theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(2, 0, v)
	}).setValue(int(c.settings.CellBorderColor.R))
	result.borderG = newNumberInput(c.theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(2, 1, v)
	}).setValue(int(c.settings.CellBorderColor.G))
	result.borderB = newNumberInput(c.theme, 3, 0, 255, 16, func(v int) {
		result.colorChanged(2, 2, v)
	}).setValue(int(c.settings.CellBorderColor.B))
	result.inputs = []*numberInput[int]{
		result.aliveR, result.aliveG, result.aliveB,
		result.deadR, result.deadG, result.deadB,
		result.borderR, result.borderG, result.borderB,
	}
	c.settingsListen(result.settingsChanged)
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

func (p *colorsPopout) settingsChanged(settings *settings.Settings) {
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
			p.core.gridHolder.grid.Draw()
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
	return p.chkBorders.Focused(gtx) ||
		p.aliveR.isFocused(gtx) || p.aliveG.isFocused(gtx) || p.aliveB.isFocused(gtx) ||
		p.deadR.isFocused(gtx) || p.deadG.isFocused(gtx) || p.deadB.isFocused(gtx) ||
		p.borderR.isFocused(gtx) || p.borderG.isFocused(gtx) || p.borderB.isFocused(gtx)
}

func (p *colorsPopout) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	for _, inp := range p.inputs {
		inp.update(gtx)
	}
	if ok := p.chkBorders.Update(gtx); ok {
		p.core.setCellBorders(p.chkBorders.Checked())
	}
	labelMax := measureMaxText(gtx, theme, font.Normal, "Alive cells", "Dead cells", "Cell Border").Size.X
	return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8), Top: unit.Dp(8), Bottom: unit.Dp(4)}.
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(rightAlignedLabel(theme, "Alive cells", labelMax)),
							layout.Rigid(label(theme, "R:")),
							layout.Flexed(1, p.aliveR.layout),
							layout.Rigid(label(theme, "G:")),
							layout.Flexed(1, p.aliveG.layout),
							layout.Rigid(label(theme, "B:")),
							layout.Flexed(1, p.aliveB.layout),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(rightAlignedLabel(theme, "Dead cells", labelMax)),
							layout.Rigid(label(theme, "R:")),
							layout.Flexed(1, p.deadR.layout),
							layout.Rigid(label(theme, "G:")),
							layout.Flexed(1, p.deadG.layout),
							layout.Rigid(label(theme, "B:")),
							layout.Flexed(1, p.deadB.layout),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(rightAlignedLabel(theme, "Cell Border", labelMax)),
							layout.Rigid(label(theme, "R:")),
							layout.Flexed(1, p.borderR.layout),
							layout.Rigid(label(theme, "G:")),
							layout.Flexed(1, p.borderG.layout),
							layout.Rigid(label(theme, "B:")),
							layout.Flexed(1, p.borderB.layout),
						)
					})
				}),
				layout.Rigid(p.chkBorders.Layout),
			)
		})
}
