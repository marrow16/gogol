package widgets

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
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
		btnResize:    newButton("Resize"),
		btnFitScreen: newButton("Fit screen"),
	}
	result.height = newNumberInput(4, 2, 999, 10, nil).setValue(int(c.settings.Height))
	result.width = newNumberInput(4, 2, 999, 10, nil).setValue(int(c.settings.Width))
	result.cellSize = newNumberInput(3, 1, 32, 1, func(v int) {
		result.core.setCellSize(v)
	}).setValue(int(c.settings.CellSize))
	result.randomize = newNumberInput(3, 0, 100, 10, func(v int) {
		result.core.setRandomization(v)
	}).setValue(int(c.settings.Randomization))
	result.inputs = []*numberInput[int]{
		result.height, result.width, result.cellSize, result.randomize,
	}
	result.wrapMode = &widget.Enum{Value: p.core.gridHolder.grid.WrapMode.String()}
	result.radioWrapNone = newRadioButton(result.wrapMode, logic.WrapNone.String(), "None")
	result.radioWrapHorizontal = newRadioButton(result.wrapMode, logic.WrapHorizontal.String(), "Horizontal")
	result.radioWrapVertical = newRadioButton(result.wrapMode, logic.WrapVertical.String(), "Vertical")
	result.radioWrapAll = newRadioButton(result.wrapMode, logic.WrapAll.String(), "Toroidal")
	result.boundaryMode = &widget.Enum{Value: p.core.gridHolder.grid.BoundaryMode.String()}
	result.radioBoundaryDead = newRadioButton(result.boundaryMode, logic.DeadBoundary.String(), "Dead cells")
	result.radioBoundaryAlive = newRadioButton(result.boundaryMode, logic.AliveBoundary.String(), "Alive cells")
	c.settingsChangeListen(result.reset)
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

func (p *sizingPopout) layout(gtx layout.Context) layout.Dimensions {
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
	labelMax := measureMaxText(gtx, font.Normal, "Grid size: ", "Cell size: ", "Wrapping mode: ", "Boundary mode: ", "Randomize %: ").Size.X
	return popoutLayout(gtx, flexVertical(8,
		rigid(flexHorizontal(20,
			rigidLabel("Grid size:", text.End, 0, labelMax),
			rigidLabel("Width:", 0, 0, 0),
			flexed(p.width.layout),
			rigidLabel("x Height:", 0, 0, 0),
			flexed(p.height.layout),
		)),
		rigid(flexHorizontal(20,
			rigidFixedWidth(nil, labelMax, 0),
			rigid(p.btnResize.Layout),
			rigid(p.btnFitScreen.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Wrapping mode:", text.End, 0, labelMax),
			rigid(p.radioWrapNone.Layout),
			rigid(p.radioWrapHorizontal.Layout),
			rigid(p.radioWrapVertical.Layout),
			rigid(p.radioWrapAll.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Boundary mode:", text.End, 0, labelMax),
			rigid(p.radioBoundaryDead.Layout),
			rigid(p.radioBoundaryAlive.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Cell size:", text.End, 0, labelMax),
			flexed(p.cellSize.layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Randomize %:", text.End, 0, labelMax),
			flexed(p.randomize.layout),
		)),
	))
}

func (p *sizingPopout) hasFocus(gtx layout.Context) bool {
	_, radiosWrap := p.wrapMode.Focused()
	_, radiosBoundary := p.boundaryMode.Focused()
	return p.height.isFocused(gtx) || p.width.isFocused(gtx) || p.cellSize.isFocused(gtx) || p.randomize.isFocused(gtx) ||
		p.btnResize.isFocused(gtx) || p.btnFitScreen.isFocused(gtx) ||
		radiosWrap || radiosBoundary
}
