package widgets

import (
	"errors"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"io/fs"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type importGridPopout struct {
	parent  *menuPopup
	core    *Core
	path    *input
	btnLoad widget.Clickable
	btnPath widget.Clickable
	resize  *widget.Bool
	error   error
}

func newImportGridPopout(p *menuPopup, c *Core) *importGridPopout {
	result := &importGridPopout{
		parent: p,
		core:   c,
		path:   newInput(c.theme, "", 256, nil).maximumWidth(30),
		resize: &widget.Bool{Value: true},
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
			resizeReqd := p.resize.Value && pattern.Height != p.core.gridHolder.grid.Height && pattern.Width != p.core.gridHolder.grid.Width
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

func (p *importGridPopout) filePicker() {
	if runtime.GOOS == "darwin" {
		out, err := exec.Command(
			"osascript",
			"-e",
			`POSIX path of (choose file)`,
		).Output()
		if err == nil {
			p.path.setText(strings.TrimSpace(string(out)))
		}
	}
}

func (p *importGridPopout) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	if p.btnLoad.Clicked(gtx) {
		p.importGrid()
	}
	if p.btnPath.Clicked(gtx) {
		p.filePicker()
	}
	labelMax := measureMaxText(gtx, theme, font.Normal, "Path: ")
	btnMax := measureMaxText(gtx, theme, font.Normal, "  Import  ")
	return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8), Top: unit.Dp(8), Bottom: unit.Dp(4)}.
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{Top: unit.Dp(2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Min.X = labelMax.Size.X
									return rightLabel(theme, "Path:").Layout(gtx)
								})
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return p.path.layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if !isMac {
									return layout.Dimensions{}
								}
								btn := material.Button(theme, &p.btnPath, "…")
								btn.Inset = layout.Inset{Bottom: 2, Left: 3, Right: 3}
								btn.Background = popupBackground
								btn.Color = popupForeground
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
								return layout.Dimensions{Size: gtx.Constraints.Min}
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									chkBox := material.CheckBox(theme, p.resize, "Resize grid")
									chkBox.TextSize = unit.Sp(16)
									chkBox.Size = 18
									return chkBox.Layout(gtx)
								})
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Max.X = btnMax.Size.X
								btn := material.Button(theme, &p.btnLoad, "Import")
								btn.Inset = layout.Inset{Bottom: 2}
								btn.TextSize = unit.Sp(16)
								return btn.Layout(gtx)
							}),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if p.error == nil {
						return layout.Dimensions{}
					}
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body1(theme, p.error.Error())
						lbl.Color = errorColor
						return lbl.Layout(gtx)

					})
				}),
			)
		})
}

func (p *importGridPopout) hasFocus(gtx layout.Context) bool {
	return p.path.isFocused(gtx) || gtx.Focused(p.resize) || gtx.Focused(&p.btnLoad) || gtx.Focused(&p.btnPath)
}
