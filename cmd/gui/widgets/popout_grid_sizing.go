package widgets

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/logic"
)

type sizingPopout struct {
	parent       *menuPopup
	core         *Core
	height       *numberInput[int]
	width        *numberInput[int]
	cellSize     *numberInput[int]
	btnResize    widget.Clickable
	btnFitScreen widget.Clickable
	wrapMode     *widget.Enum
	boundaryMode *widget.Enum
	randomize    *numberInput[int]
	inputs       []*numberInput[int]
}

func newSizingPopout(p *menuPopup, c *Core) *sizingPopout {
	result := &sizingPopout{
		parent: p,
		core:   c,
	}
	result.height = newNumberInput(c.theme, 4, 2, 999, 10, nil).setValue(int(c.settings.Height))
	result.width = newNumberInput(c.theme, 4, 2, 999, 10, nil).setValue(int(c.settings.Width))
	result.cellSize = newNumberInput(c.theme, 3, 3, 32, 1, func(v int) {
		result.core.setCellSize(v)
	}).setValue(int(c.settings.CellSize))
	result.randomize = newNumberInput(c.theme, 3, 0, 100, 10, func(v int) {
		result.core.setRandomization(v)
	}).setValue(int(c.settings.Randomization))
	result.inputs = []*numberInput[int]{
		result.height, result.width, result.cellSize, result.randomize,
	}
	result.wrapMode = &widget.Enum{Value: p.core.gridHolder.grid.WrapMode.String()}
	result.boundaryMode = &widget.Enum{Value: p.core.gridHolder.grid.BoundaryMode.String()}
	return result
}

func (p *sizingPopout) reset() {
	p.height.setValue(p.core.gridHolder.grid.Height)
	p.width.setValue(p.core.gridHolder.grid.Width)
	p.cellSize.setValue(p.core.settings.CellSize)
	p.wrapMode.Value = p.core.gridHolder.grid.WrapMode.String()
	p.boundaryMode.Value = p.core.gridHolder.grid.BoundaryMode.String()
	p.randomize.setValue(p.core.settings.Randomization)
}

func (p *sizingPopout) fitScreen(gtx layout.Context) {
	sbHt := gtx.Dp(p.core.statusBar.height)
	cellSize := float32(p.core.settings.CellSize) * p.core.gridHolder.zoom
	if cellSize <= 0 {
		cellSize = 1
	}
	availH := float32(p.core.windowRect.Max.Y - sbHt)
	availW := float32(p.core.windowRect.Max.X)
	h := int(availH / cellSize)
	w := int(availW / cellSize)
	if h < 2 {
		h = 2
	}
	if w < 2 {
		w = 2
	}
	p.core.gridResize(h, w)
	p.height.setValue(p.core.settings.Height)
	p.width.setValue(p.core.settings.Width)
}

func (p *sizingPopout) resize() {
	h, w := p.height.current(), p.width.current()
	if h < 2 {
		h = 2
	}
	if w < 2 {
		w = 2
	}
	p.core.gridResize(h, w)
	p.height.setValue(p.core.settings.Height)
	p.width.setValue(p.core.settings.Width)
}

func (p *sizingPopout) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	for _, inp := range p.inputs {
		inp.update(gtx)
	}
	if p.btnFitScreen.Clicked(gtx) {
		p.fitScreen(gtx)
	}
	if p.btnResize.Clicked(gtx) {
		p.resize()
	}
	if p.wrapMode.Update(gtx) {
		p.core.setWrapMode(logic.WrapModeFromString(p.wrapMode.Value, p.core.gridHolder.grid.WrapMode))
	}
	if p.boundaryMode.Update(gtx) {
		p.core.setBoundaryMode(logic.BoundaryModeFromString(p.boundaryMode.Value, p.core.gridHolder.grid.BoundaryMode))
	}
	labelMax := measureMaxText(gtx, theme, font.Normal, "Grid size: ", "Cell size: ", "Wrapping mode: ", "Boundary mode: ", "Randomize %: ")
	btnMax := measureMaxText(gtx, theme, font.Normal, "  Resize  ", "  Fit screen  ")
	return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8), Top: unit.Dp(8), Bottom: unit.Dp(4)}.
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = labelMax.Size.X
								return rightLabel(p.core.theme, "Grid size:").Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return material.Body1(theme, "Width:").Layout(gtx)
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return p.width.layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return material.Body1(theme, "x Height:").Layout(gtx)
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return p.height.layout(gtx)
							}),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = labelMax.Size.X
								return layout.Dimensions{Size: gtx.Constraints.Min}
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Max.X = btnMax.Size.X
								btn := material.Button(p.core.theme, &p.btnResize, "Resize")
								btn.Inset = layout.Inset{Bottom: 2}
								btn.TextSize = unit.Sp(16)
								return btn.Layout(gtx)
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Max.X = btnMax.Size.X
								btn := material.Button(p.core.theme, &p.btnFitScreen, "Fit screen")
								btn.Inset = layout.Inset{Bottom: 2}
								btn.TextSize = unit.Sp(16)
								return btn.Layout(gtx)
							}),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = labelMax.Size.X
								return rightLabel(p.core.theme, "Wrapping mode:").Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								radio := material.RadioButton(p.core.theme, p.wrapMode, logic.WrapNone.String(), "None")
								radio.Size = 18
								radio.TextSize = unit.Sp(16)
								return radio.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								radio := material.RadioButton(p.core.theme, p.wrapMode, logic.WrapHorizontal.String(), "Horizontal")
								radio.Size = 18
								radio.TextSize = unit.Sp(16)
								return radio.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								radio := material.RadioButton(p.core.theme, p.wrapMode, logic.WrapVertical.String(), "Vertical")
								radio.Size = 18
								radio.TextSize = unit.Sp(16)
								return radio.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								radio := material.RadioButton(p.core.theme, p.wrapMode, logic.WrapAll.String(), "Toroidal")
								radio.Size = 18
								radio.TextSize = unit.Sp(16)
								return radio.Layout(gtx)
							}),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = labelMax.Size.X
								return rightLabel(p.core.theme, "Boundary mode:").Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								radio := material.RadioButton(p.core.theme, p.boundaryMode, logic.DeadBoundary.String(), "Dead cells")
								radio.Size = 18
								radio.TextSize = unit.Sp(16)
								return radio.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								radio := material.RadioButton(p.core.theme, p.boundaryMode, logic.AliveBoundary.String(), "Alive cells")
								radio.Size = 18
								radio.TextSize = unit.Sp(16)
								return radio.Layout(gtx)
							}),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = labelMax.Size.X
								return rightLabel(p.core.theme, "Cell size:").Layout(gtx)
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return p.cellSize.layout(gtx)
							}),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = labelMax.Size.X
								return rightLabel(p.core.theme, "Randomize %:").Layout(gtx)
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return p.randomize.layout(gtx)
							}),
						)
					})
				}),
			)
		})
}

func (p *sizingPopout) hasFocus(gtx layout.Context) bool {
	return p.height.isFocused(gtx) || p.width.isFocused(gtx) || p.cellSize.isFocused(gtx) || p.randomize.isFocused(gtx) ||
		gtx.Focused(&p.btnResize) || gtx.Focused(&p.btnFitScreen)
}
