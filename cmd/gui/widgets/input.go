package widgets

import (
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type input struct {
	editor     widget.Editor
	style      material.EditorStyle
	theme      *material.Theme
	onChangeFn func(text string)
	upDownFn   func(k key.Name, text string) (string, bool)
	maxWidth   int
}

func newInput(theme *material.Theme, filter any, maxLen int, fn func(text string)) *input {
	result := &input{
		onChangeFn: fn,
		editor: widget.Editor{
			Alignment:  text.Start,
			Filter:     buildFilterString(filter),
			MaxLen:     maxLen,
			SingleLine: true,
			Submit:     true,
		},
		theme: theme,
	}
	result.style = material.Editor(theme, &result.editor, "")
	return result
}

func (i *input) upDownSupport(fn func(k key.Name, text string) (string, bool)) *input {
	i.upDownFn = fn
	return i
}

func (i *input) maximumWidth(w int) *input {
	i.maxWidth = w
	return i
}

func (i *input) setText(text string) {
	if text != i.editor.Text() {
		i.editor.SetText(text)
		i.editor.SetCaret(len(text), -len(text))
	}
}

func (i *input) setFocused(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: &i.editor})
}

func (i *input) isFocused(gtx layout.Context) bool {
	return gtx.Focused(&i.editor)
}

func (i *input) layout(gtx layout.Context) layout.Dimensions {
	i.update(gtx)
	if i.maxWidth > 0 {
		maxChWd := measureText(gtx, i.theme, "M")
		maxWidth := maxChWd.Size.X * i.maxWidth
		if maxWidth < gtx.Constraints.Max.X {
			gtx.Constraints.Max.X = maxWidth
		}
		if gtx.Constraints.Min.X > gtx.Constraints.Max.X {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
		}
	} else if i.editor.MaxLen > 0 {
		maxChWd := measureText(gtx, i.theme, "M")
		maxWidth := maxChWd.Size.X * i.editor.MaxLen
		if maxWidth < gtx.Constraints.Max.X {
			gtx.Constraints.Max.X = maxWidth
		}
		if gtx.Constraints.Min.X > gtx.Constraints.Max.X {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
		}
	}
	borderColor := popupBorder
	borderThickness := unit.Dp(1)
	if i.isFocused(gtx) {
		borderColor = popupBorderFocused
		borderThickness = unit.Dp(2)
	}
	return widget.Border{
		Color:        borderColor,
		CornerRadius: 3,
		Width:        borderThickness,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, i.style.Layout)
	})
}

func (i *input) update(gtx layout.Context) {
	currentlyFocused := gtx.Focused(&i.editor)
	if currentlyFocused && i.upDownFn != nil {
		for {
			ev, ok := gtx.Event(
				pointer.Filter{
					Target:  &i.editor,
					Kinds:   pointer.Scroll,
					ScrollY: pointer.ScrollRange{Min: 0, Max: 100000},
				},
				key.Filter{Focus: &i.editor, Name: key.NameUpArrow},
				key.Filter{Focus: &i.editor, Name: key.NameDownArrow},
				key.Filter{Focus: &i.editor, Name: key.NamePageUp},
				key.Filter{Focus: &i.editor, Name: key.NamePageDown},
			)
			if !ok {
				break
			}
			switch evt := ev.(type) {
			case key.Event:
				if evt.State == key.Press {
					if s, ok := i.upDownFn(evt.Name, i.editor.Text()); ok {
						i.setText(s)
					}
				}
			case pointer.Event:
				if evt.Scroll.Y == 0 {
					if s, ok := i.upDownFn(key.NameUpArrow, i.editor.Text()); ok {
						i.setText(s)
					}
				} else {
					if s, ok := i.upDownFn(key.NameDownArrow, i.editor.Text()); ok {
						i.setText(s)
					}
				}
			}
		}
	}
	for {
		ev, ok := i.editor.Update(gtx)
		if !ok {
			break
		}
		switch ev.(type) {
		case widget.ChangeEvent:
			if i.onChangeFn != nil {
				i.onChangeFn(i.editor.Text())
			}
		case widget.SubmitEvent:
			// optional
		}
	}
	if currentlyFocused != gtx.Focused(&i.editor) {
		gtx.Execute(op.InvalidateCmd{})
	}
}

type filterType int

func filterFromType(v any) filterType {
	switch v.(type) {
	case int:
		return filterInt
	case uint:
		return filterUint
	case float64:

		return filterFloat
	}
	return filterNone
}

func (f filterType) String() string {
	switch f {
	case filterAlphabet:
		return "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	case filterAlphabetUpper:
		return "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case filterAlphabetLower:
		return "abcdefghijklmnopqrstuvwxyz"
	case filterInt:
		return "-0123456789"
	case filterUint:
		return "0123456789"
	case filterFloat:
		return "-.0123456789"
	case filterFloatPositive:
		return ".0123456789"
	}
	return ""
}

const (
	filterNone filterType = iota
	filterAlphabet
	filterAlphabetLower
	filterAlphabetUpper
	filterInt
	filterUint
	filterFloat
	filterFloatPositive
)

func buildFilterString(filter any) string {
	switch ft := filter.(type) {
	case string:
		return ft
	case filterType:
		return ft.String()
	}
	return ""
}
