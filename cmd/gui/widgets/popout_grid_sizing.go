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
	parent              *menuPopup
	core                *Core
	height              *numberInput[int]
	width               *numberInput[int]
	cellSize            *numberInput[int]
	btnResize           *button
	btnFitScreen        *button
	wrapMode            *widget.Enum
	radioWrapNone       *radioButton
	radioWrapHorizontal *radioButton
	radioWrapVertical   *radioButton
	radioWrapAll        *radioButton
	boundaryMode        *widget.Enum
	radioBoundaryDead   *radioButton
	radioBoundaryAlive  *radioButton
	randomize           *numberInput[int]
	inputs              []*numberInput[int]
}

func newSizingPopout(p *menuPopup, c *Core) *sizingPopout {
	result := &sizingPopout{
		parent:       p,
		core:         c,
		btnResize:    newButton(c.theme, "Resize"),
		btnFitScreen: newButton(c.theme, "Fit screen"),
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
	result.radioWrapNone = newRadioButton(c.theme, result.wrapMode, logic.WrapNone.String(), "None")
	result.radioWrapHorizontal = newRadioButton(c.theme, result.wrapMode, logic.WrapHorizontal.String(), "Horizontal")
	result.radioWrapVertical = newRadioButton(c.theme, result.wrapMode, logic.WrapVertical.String(), "Vertical")
	result.radioWrapAll = newRadioButton(c.theme, result.wrapMode, logic.WrapAll.String(), "Toroidal")
	result.boundaryMode = &widget.Enum{Value: p.core.gridHolder.grid.BoundaryMode.String()}
	result.radioBoundaryDead = newRadioButton(c.theme, result.boundaryMode, logic.DeadBoundary.String(), "Dead cells")
	result.radioBoundaryAlive = newRadioButton(c.theme, result.boundaryMode, logic.AliveBoundary.String(), "Alive cells")
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
	labelMax := measureMaxText(gtx, theme, font.Normal, "Grid size: ", "Cell size: ", "Wrapping mode: ", "Boundary mode: ", "Randomize %: ").Size.X
	return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8), Top: unit.Dp(8), Bottom: unit.Dp(4)}.
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(rightAlignedLabel(theme, "Grid size:", labelMax)),
							layout.Rigid(label(theme, "Width:")),
							layout.Flexed(1, p.width.layout),
							layout.Rigid(label(theme, "x Height:")),
							layout.Flexed(1, p.height.layout),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = labelMax
								return layout.Dimensions{Size: gtx.Constraints.Min}
							}),
							layout.Rigid(p.btnResize.Layout),
							layout.Rigid(p.btnFitScreen.Layout),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(rightAlignedLabel(theme, "Wrapping mode:", labelMax)),
							layout.Rigid(p.radioWrapNone.Layout),
							layout.Rigid(p.radioWrapHorizontal.Layout),
							layout.Rigid(p.radioWrapVertical.Layout),
							layout.Rigid(p.radioWrapAll.Layout),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(rightAlignedLabel(theme, "Boundary mode:", labelMax)),
							layout.Rigid(p.radioBoundaryDead.Layout),
							layout.Rigid(p.radioBoundaryAlive.Layout),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(rightAlignedLabel(theme, "Cell size:", labelMax)),
							layout.Flexed(1, p.cellSize.layout),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
							layout.Rigid(rightAlignedLabel(theme, "Randomize %:", labelMax)),
							layout.Flexed(1, p.randomize.layout),
						)
					})
				}),
			)
		})
}

func (p *sizingPopout) hasFocus(gtx layout.Context) bool {
	_, radiosWrap := p.wrapMode.Focused()
	_, radiosBoundary := p.boundaryMode.Focused()
	return p.height.isFocused(gtx) || p.width.isFocused(gtx) || p.cellSize.isFocused(gtx) || p.randomize.isFocused(gtx) ||
		p.btnResize.isFocused(gtx) || p.btnFitScreen.isFocused(gtx) ||
		radiosWrap || radiosBoundary
}
