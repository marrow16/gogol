package controls

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"strings"
)

type Editor struct {
	theme      *material.Theme
	editor     widget.Editor
	style      material.EditorStyle
	onChangeFn func(text string, gtx layout.Context)
}

func NewEditor(theme *material.Theme, fn func(text string, gtx layout.Context)) *Editor {
	result := &Editor{
		theme:      theme,
		onChangeFn: fn,
		editor: widget.Editor{
			Alignment:  text.Start,
			SingleLine: false,
			Submit:     false,
		},
	}
	result.style = material.Editor(theme, &result.editor, "")
	return result
}

func (i *Editor) SetText(text string) {
	if text != i.editor.Text() {
		i.editor.SetText(strings.ReplaceAll(text, "\t", "    "))
	}
}

func (i *Editor) Text() string {
	return i.editor.Text()
}

func (i *Editor) SetFocused(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: &i.editor})
}

func (i *Editor) IsFocused(gtx layout.Context) bool {
	return gtx.Focused(&i.editor)
}

func (i *Editor) Layout(gtx layout.Context) layout.Dimensions {
	i.Update(gtx)
	bc, bt := Style.bordering(i.IsFocused(gtx))
	return widget.Border{Color: bc, CornerRadius: 3, Width: bt}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, i.style.Layout)
	})
}

func (i *Editor) Update(gtx layout.Context) {
	for {
		ev, ok := i.editor.Update(gtx)
		if !ok {
			break
		}
		switch ev.(type) {
		case widget.ChangeEvent:
			if i.onChangeFn != nil {
				i.onChangeFn(i.editor.Text(), gtx)
			}
		}
	}
}
