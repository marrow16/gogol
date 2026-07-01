package widgets

import (
	"errors"
	"fmt"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/cmd/gui/settings"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type loadPatternsPopout struct {
	parent  *menuPopup
	core    *Core
	path    *input
	btnLoad widget.Clickable
	btnPath widget.Clickable
	error   error
	loaded  *int
}

func newLoadPatternsPopout(p *menuPopup, c *Core) *loadPatternsPopout {
	result := &loadPatternsPopout{
		parent: p,
		core:   c,
		path:   newInput(c.theme, "", 256, nil).maximumWidth(30),
	}
	result.path.onChangeFn = result.clearError
	return result
}

func (p *loadPatternsPopout) reset() {
}

func (p *loadPatternsPopout) clearError(text string) {
	p.error = nil
	p.loaded = nil
}

func (p *loadPatternsPopout) filePicker(gtx layout.Context) {
	if runtime.GOOS == "darwin" {
		out, err := exec.Command(
			"osascript",
			"-e",
			`POSIX path of (choose file)`,
		).Output()
		if err == nil {
			p.path.setText(strings.TrimSpace(string(out)))
			p.path.setFocused(gtx)
		}
	}
}

func (p *loadPatternsPopout) loadPatterns() {
	path := p.path.editor.Text()
	if path == "" {
		p.error = errors.New("Please enter a path")
		return
	}
	fs, err := os.Stat(path)
	if err != nil || fs == nil {
		p.error = errors.New("Invalid filepath")
		return
	}
	if fs.IsDir() {
		p.loaded, p.error = settings.LoadPatternsLibrary(path)
		if p.error == nil && p.loaded != nil && *p.loaded > 0 {
			p.core.settings.AddPatternLibrary(path)
		}
	} else if err = settings.LoadPattern(path); err == nil {
		loaded := 1
		p.core.settings.AddPattern(path)
		p.loaded = &loaded
	} else {
		p.error = err
	}
}

func (p *loadPatternsPopout) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	if p.btnLoad.Clicked(gtx) {
		p.loadPatterns()
	}
	if p.btnPath.Clicked(gtx) {
		p.filePicker(gtx)
	}
	labelMax := measureMaxText(gtx, theme, font.Normal, "Path: ")
	btnMax := measureMaxText(gtx, theme, font.Normal, "  Load  ")
	return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8), Top: unit.Dp(8), Bottom: unit.Dp(4)}.
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body1(theme, "Enter pattern filename or directory (for library)...")
						lbl.TextSize = lbl.TextSize - 2
						return lbl.Layout(gtx)
					})
				}),
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
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Max.X = btnMax.Size.X
								btn := material.Button(theme, &p.btnLoad, "Load")
								btn.Inset = layout.Inset{Bottom: 2}
								btn.TextSize = unit.Sp(16)
								return btn.Layout(gtx)
							}),
						)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					switch {
					case p.error != nil:
						return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body1(theme, p.error.Error())
							lbl.Color = errorColor
							return lbl.Layout(gtx)
						})
					case p.loaded != nil:
						return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return material.Body1(theme, fmt.Sprintf("Successfully loaded %d pattern(s)", *p.loaded)).Layout(gtx)
						})
					default:
						return layout.Dimensions{}
					}
				}),
			)
		})
}

func (p *loadPatternsPopout) hasFocus(gtx layout.Context) bool {
	return p.path.isFocused(gtx) || gtx.Focused(&p.btnLoad) || gtx.Focused(&p.btnPath)
}
