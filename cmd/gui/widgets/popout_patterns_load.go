package widgets

import (
	"errors"
	"fmt"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/cmd/gui/settings"
	"os"
	"strings"
)

type loadPatternsPopout struct {
	parent  *menuPopup
	core    *Core
	path    *input
	btnLoad *button
	btnPath *pathButton
	error   error
	loaded  *int
}

func newLoadPatternsPopout(p *menuPopup, c *Core) *loadPatternsPopout {
	result := &loadPatternsPopout{
		parent:  p,
		core:    c,
		path:    newInput("", 256, nil).maximumWidth(30),
		btnLoad: newButton("Load"),
		btnPath: newPathButton(),
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

func (p *loadPatternsPopout) layout(gtx layout.Context) layout.Dimensions {
	if p.btnLoad.Clicked(gtx) {
		p.loadPatterns()
	}
	if p.btnPath.Clicked(gtx) {
		filePicker(func(filename string) {
			p.path.setText(strings.TrimSpace(string(filename)))
			p.path.setFocused(gtx)

		})
	}
	labelMax := measureMaxText(gtx, font.Normal, "Path: ").Size.X
	return layout.Inset{Left: 8, Right: 8, Top: 8, Bottom: 4}.Layout(gtx, flexVertical(8,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: 4}.Layout(gtx, material.Label(theme, theme.TextSize-2, "Enter pattern filename or directory (for library)...").Layout)
		}),
		layout.Rigid(flexHorizontal(0,
			rigidLabel("Path: ", text.End, 0, labelMax),
			layout.Flexed(1, p.path.layout),
			layout.Rigid(p.btnPath.Layout),
		)),
		layout.Rigid(flexHorizontal(20,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = labelMax
				return layout.Dimensions{Size: gtx.Constraints.Min}
			}),
			layout.Rigid(p.btnLoad.Layout),
		)),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			switch {
			case p.error != nil:
				return layout.Inset{Bottom: 4}.Layout(gtx, errorLabel(p.error))
			case p.loaded != nil:
				return layout.Inset{Bottom: 4}.Layout(gtx, label(fmt.Sprintf("Successfully loaded %d pattern(s)", *p.loaded)))
			default:
				return layout.Dimensions{}
			}
		}),
	))
}

func (p *loadPatternsPopout) hasFocus(gtx layout.Context) bool {
	return p.path.isFocused(gtx) || p.btnLoad.isFocused(gtx) || p.btnPath.isFocused(gtx)
}
