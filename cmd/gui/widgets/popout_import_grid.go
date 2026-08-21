package widgets

import (
	"errors"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

type importGridPopout struct {
	parent    *menuPopup
	core      *Core
	path      *input
	btnImport *button
	btnPath   *pathButton
	chkResize *checkbox
	error     error
}

func newImportGridPopout(p *menuPopup, c *Core) *importGridPopout {
	result := &importGridPopout{
		parent:    p,
		core:      c,
		path:      newInput("", 256, nil).maximumWidth(30),
		btnImport: newButton("Import"),
		btnPath:   newPathButton(),
		chkResize: newCheckBox("Resize grid", true),
	}
	result.path.onChangeFn = result.clearError
	return result
}

func (p *importGridPopout) reset() {
}

func (p *importGridPopout) clearError(text string) {
	p.error = nil
}

func (p *importGridPopout) importGrid() {
	path := p.path.editor.Text()
	if path == "" {
		p.error = errors.New("Please enter a path")
		return
	}
	if f, err := os.Open(path); err == nil {
		defer func() {
			_ = f.Close()
		}()
		if pattern, err := patterns.NewPatternFromRle(f); err == nil {
			// parse wrap and boundary modes...
			wrap := p.core.gridHolder.grid.WrapMode
			boundary := p.core.gridHolder.grid.BoundaryMode
			step := uint64(0)
			for _, c := range pattern.Comments {
				if after, ok := strings.CutPrefix(c, "Wrap mode: "); ok {
					wrap = logic.WrapModeFromString(after, wrap)
				} else if after, ok := strings.CutPrefix(c, "Boundary mode: "); ok {
					boundary = logic.BoundaryModeFromString(after, boundary)
				} else if after, ok := strings.CutPrefix(c, "Step: "); ok {
					if n, err := strconv.ParseUint(after, 10, 64); err == nil {
						step = uint64(n)
					}
				}
			}
			p.core.stop()
			resizeReqd := p.chkResize.Checked() && pattern.Height != p.core.gridHolder.grid.Height && pattern.Width != p.core.gridHolder.grid.Width
			if resizeReqd {
				p.core.gridResize(pattern.Height, pattern.Width)
			}
			p.core.gridHolder.grid.Rule = pattern.Rule
			p.core.gridHolder.grid.SetBoundaryMode(boundary)
			p.core.gridHolder.grid.SetWrapMode(wrap)
			p.core.gridHolder.grid.StepCount.Store(step)
			pattern.Draw(p.core.gridHolder.grid, 0, 0, patterns.Rotate0)
		} else {
			p.error = err
		}
	} else if errors.Is(err, fs.ErrNotExist) {
		p.error = errors.New("File does not exist")
	} else {
		p.error = err
	}
}

func (p *importGridPopout) layout(gtx layout.Context) layout.Dimensions {
	if p.btnImport.Clicked(gtx) {
		p.importGrid()
	}
	if p.btnPath.Clicked(gtx) {
		filePicker(func(filename string) {
			p.path.setText(strings.TrimSpace(string(filename)))
		})
	}
	labelMax := measureMaxText(gtx, font.Normal, "Path: ").Size.X
	return popoutLayout(gtx, flexVertical(8,
		rigid(flexHorizontal(0,
			rigidLabel("Path: ", text.End, 0, labelMax),
			flexed(p.path.layout),
			rigid(p.btnPath.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidFixedWidth(nil, labelMax-20, 0),
			rigid(p.chkResize.Layout),
			rigid(p.btnImport.Layout),
		)),
		rigid(errorLabel(p.error)),
	))
}

func (p *importGridPopout) hasFocus(gtx layout.Context) bool {
	return p.path.isFocused(gtx) || p.chkResize.isFocused(gtx) ||
		p.btnImport.isFocused(gtx) || p.btnPath.isFocused(gtx)
}
