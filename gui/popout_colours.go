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
	alive      *controls.ColorPicker
	dead       *controls.ColorPicker
	border     *controls.ColorPicker
	chkBorders *controls.Checkbox
}

func newColorsPopout(p *menuPopup, c *Core) *colorsPopout {
	result := &colorsPopout{
		parent:     p,
		core:       c,
		chkBorders: controls.NewCheckBox(theme, "Show Borders", c.settings.CellBorders),
	}
	result.alive = controls.NewColorPicker(theme, c.settings.CellAliveColor, func(c color.NRGBA) {
		p.core.settings.CellAliveColor = c
		p.core.gridHolder.grid.Draw()
	}).Direction(true)
	result.dead = controls.NewColorPicker(theme, c.settings.CellDeadColor, func(c color.NRGBA) {
		p.core.settings.CellDeadColor = c
		p.core.gridHolder.grid.Draw()
	}).Direction(true)
	result.border = controls.NewColorPicker(theme, c.settings.CellBorderColor, func(c color.NRGBA) {
		p.core.settings.CellBorderColor = c
		p.core.gridHolder.rebuild()
	}).Direction(true)
	c.settingsChangeListen(result.reset)
	return result
}

func (p *colorsPopout) reset() {
	p.alive.SetValue(p.core.settings.CellAliveColor)
	p.dead.SetValue(p.core.settings.CellDeadColor)
	p.border.SetValue(p.core.settings.CellBorderColor)
	p.chkBorders.SetChecked(p.core.settings.CellBorders)
}

func (p *colorsPopout) hasFocus(gtx layout.Context) bool {
	return p.chkBorders.IsFocused(gtx) ||
		p.alive.IsFocused(gtx) || p.dead.IsFocused(gtx) || p.border.IsFocused(gtx)
}

func (p *colorsPopout) layout(gtx layout.Context) layout.Dimensions {
	if ok := p.chkBorders.Update(gtx); ok {
		p.core.setCellBorders(p.chkBorders.Checked())
	}
	labelMax := measureMaxText(gtx, font.Normal, "Alive cells:", "Dead cells:", "Cell border:")
	return popoutLayout(gtx, func(gtx layout.Context) layout.Dimensions {
		dims := flexVertical(4,
			rigid(flexHorizontal(20,
				rigidLabel("Alive cells:", text.End, 0, labelMax),
				rigid(func(gtx layout.Context) layout.Dimensions {
					dims := p.alive.Layout(gtx)
					p.alive.LayoutDropdown(gtx)
					return dims
				}),
			)),
			rigid(flexHorizontal(20,
				rigidLabel("Dead cells:", text.End, 0, labelMax),
				rigid(func(gtx layout.Context) layout.Dimensions {
					dims := p.dead.Layout(gtx)
					p.dead.LayoutDropdown(gtx)
					return dims
				}),
			)),
			rigid(flexHorizontal(20,
				rigidLabel("Cell border:", text.End, 0, labelMax),
				rigid(func(gtx layout.Context) layout.Dimensions {
					dims := p.border.Layout(gtx)
					p.border.LayoutDropdown(gtx)
					return dims
				}),
			)),
			rigid(p.chkBorders.Layout),
		)(gtx)
		return dims
	})
}
