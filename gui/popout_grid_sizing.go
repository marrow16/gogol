package gui

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"github.com/marrow16/gogol/gui/controls"
	"github.com/marrow16/gogol/logic"
)

type sizingPopout struct {
	parent              *menuPopup
	core                *Core
	height              *controls.NumberInput[int]
	width               *controls.NumberInput[int]
	cellSize            *controls.NumberInput[int]
	btnResize           *controls.Button
	btnFitScreen        *controls.Button
	chkKeepOnResize     *controls.Checkbox
	wrapMode            *widget.Enum
	radioWrapNone       *controls.RadioButton
	radioWrapHorizontal *controls.RadioButton
	radioWrapVertical   *controls.RadioButton
	radioWrapAll        *controls.RadioButton
	boundaryMode        *widget.Enum
	radioBoundaryDead   *controls.RadioButton
	radioBoundaryAlive  *controls.RadioButton
	randomize           *controls.NumberInput[int]
	inputs              []*controls.NumberInput[int]
}

func newSizingPopout(p *menuPopup, c *Core) *sizingPopout {
	result := &sizingPopout{
		parent:          p,
		core:            c,
		btnResize:       controls.NewButton(theme, "Resize"),
		btnFitScreen:    controls.NewButton(theme, "Fit screen"),
		chkKeepOnResize: controls.NewCheckBox(theme, "Keep cells", c.settings.KeepCellsOnResize),
	}
	result.height = controls.NewNumberInput(theme, 4, 2, 9999, 10, nil).SetValue(c.settings.Height)
	result.width = controls.NewNumberInput(theme, 4, 2, 9999, 10, nil).SetValue(c.settings.Width)
	result.cellSize = controls.NewNumberInput(theme, 3, 1, 32, 1, func(v int) {
		result.core.setCellSize(v)
	}).SetValue(c.settings.CellSize)
	result.randomize = controls.NewNumberInput(theme, 3, 0, 100, 10, func(v int) {
		result.core.setRandomization(v)
	}).SetValue(c.settings.Randomization)
	result.inputs = []*controls.NumberInput[int]{
		result.height, result.width, result.cellSize, result.randomize,
	}
	result.wrapMode = &widget.Enum{Value: p.core.gridHolder.grid.WrapMode().String()}
	result.radioWrapNone = controls.NewRadioButton(theme, result.wrapMode, logic.WrapNone.String(), "None")
	result.radioWrapHorizontal = controls.NewRadioButton(theme, result.wrapMode, logic.WrapHorizontal.String(), "Horizontal")
	result.radioWrapVertical = controls.NewRadioButton(theme, result.wrapMode, logic.WrapVertical.String(), "Vertical")
	result.radioWrapAll = controls.NewRadioButton(theme, result.wrapMode, logic.WrapAll.String(), "Toroidal")
	result.boundaryMode = &widget.Enum{Value: p.core.gridHolder.grid.BoundaryMode().String()}
	result.radioBoundaryDead = controls.NewRadioButton(theme, result.boundaryMode, logic.DeadBoundary.String(), "Dead cells")
	result.radioBoundaryAlive = controls.NewRadioButton(theme, result.boundaryMode, logic.AliveBoundary.String(), "Alive cells")
	c.settingsChangeListen(result.reset)
	return result
}

func (p *sizingPopout) reset() {
	p.height.SetValue(p.core.gridHolder.grid.Height())
	p.width.SetValue(p.core.gridHolder.grid.Width())
	p.chkKeepOnResize.SetChecked(p.core.settings.KeepCellsOnResize)
	p.cellSize.SetValue(p.core.settings.CellSize)
	p.wrapMode.Value = p.core.gridHolder.grid.WrapMode().String()
	p.boundaryMode.Value = p.core.gridHolder.grid.BoundaryMode().String()
	p.randomize.SetValue(p.core.settings.Randomization)
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
	p.height.SetValue(p.core.settings.Height)
	p.width.SetValue(p.core.settings.Width)
}

func (p *sizingPopout) resize() {
	h, w := p.height.Current(), p.width.Current()
	if h < 2 {
		h = 2
	}
	if w < 2 {
		w = 2
	}
	p.core.gridResize(h, w)
	p.height.SetValue(p.core.settings.Height)
	p.width.SetValue(p.core.settings.Width)
}

func (p *sizingPopout) layout(gtx layout.Context) layout.Dimensions {
	for _, inp := range p.inputs {
		inp.Update(gtx)
	}
	if p.btnFitScreen.Clicked(gtx) {
		p.fitScreen(gtx)
	}
	if p.btnResize.Clicked(gtx) {
		p.resize()
	}
	if p.chkKeepOnResize.Update(gtx) {
		p.core.settings.KeepCellsOnResize = p.chkKeepOnResize.Checked()
	}
	if p.wrapMode.Update(gtx) {
		p.core.setWrapMode(logic.WrapModeFromString(p.wrapMode.Value, p.core.gridHolder.grid.WrapMode()))
	}
	if p.boundaryMode.Update(gtx) {
		p.core.setBoundaryMode(logic.BoundaryModeFromString(p.boundaryMode.Value, p.core.gridHolder.grid.BoundaryMode()))
	}
	labelMax := measureMaxText(gtx, font.Normal, "Grid size: ", "Cell size: ", "Wrapping mode: ", "Boundary mode: ", "Randomize %: ")
	return popoutLayout(gtx, flexVertical(8,
		rigid(flexHorizontal(20,
			rigidLabel("Grid size:", text.End, 0, labelMax),
			rigidLabel("Width:", 0, 0, 0),
			flexed(p.width.Layout),
			rigidLabel("x Height:", 0, 0, 0),
			flexed(p.height.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidFixedWidth(nil, labelMax, 0),
			rigid(p.btnResize.Layout),
			rigid(p.btnFitScreen.Layout),
			rigid(p.chkKeepOnResize.Layout),
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
			flexed(p.cellSize.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Randomize %:", text.End, 0, labelMax),
			flexed(p.randomize.Layout),
		)),
	))
}

func (p *sizingPopout) hasFocus(gtx layout.Context) bool {
	_, radiosWrap := p.wrapMode.Focused()
	_, radiosBoundary := p.boundaryMode.Focused()
	return p.height.IsFocused(gtx) || p.width.IsFocused(gtx) || p.cellSize.IsFocused(gtx) || p.randomize.IsFocused(gtx) ||
		p.btnResize.IsFocused(gtx) || p.btnFitScreen.IsFocused(gtx) || p.chkKeepOnResize.IsFocused(gtx) ||
		radiosWrap || radiosBoundary
}
