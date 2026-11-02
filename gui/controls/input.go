package controls

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
)

type Input struct {
	theme            *material.Theme
	editor           widget.Editor
	style            material.EditorStyle
	onChangeFn       func(text string)
	onSubmitFn       func(text string)
	upDownFn         func(k key.Name, text string) (string, bool)
	maxWidth         int
	decorator        layout.Widget
	decoratorOnClick func()
}

func NewInput(theme *material.Theme, filter any, maxLen int, fn func(text string)) *Input {
	result := &Input{
		theme:      theme,
		onChangeFn: fn,
		editor: widget.Editor{
			Alignment:  text.Start,
			Filter:     buildFilterString(filter),
			MaxLen:     maxLen,
			SingleLine: true,
			Submit:     true,
		},
	}
	result.style = material.Editor(theme, &result.editor, "")
	return result
}

func (i *Input) UpDownSupport(fn func(k key.Name, text string) (string, bool)) *Input {
	i.upDownFn = fn
	return i
}

func (i *Input) MaximumWidth(w int) *Input {
	i.maxWidth = w
	return i
}

func (i *Input) OnSubmit(fn func(string)) *Input {
	i.onSubmitFn = fn
	return i
}

func (i *Input) ReadOnly(ro bool) *Input {
	i.editor.ReadOnly = ro
	return i
}

func (i *Input) SetCaret(start, end int) {
	i.editor.SetCaret(start, end)
}

func (i *Input) SetText(text string) {
	if text != i.editor.Text() {
		i.editor.SetText(text)
		i.editor.SetCaret(len(text), -len(text))
	}
}

func (i *Input) Text() string {
	return i.editor.Text()
}

func (i *Input) SetFocused(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: &i.editor})
}

func (i *Input) IsFocused(gtx layout.Context) bool {
	return gtx.Focused(&i.editor)
}

func (i *Input) Layout(gtx layout.Context) layout.Dimensions {
	i.Update(gtx)
	maxWd := 0
	if i.maxWidth > 0 {
		maxWd = measureText(gtx, i.theme, "\u2003").Size.X * i.maxWidth
	} else if i.editor.MaxLen > 0 {
		maxWd = measureText(gtx, i.theme, "\u2003").Size.X * i.editor.MaxLen
	}
	bc, bt := Style.bordering(!i.editor.ReadOnly && i.IsFocused(gtx))
	if i.decorator != nil {
		if maxWd == 0 {
			maxWd = measureText(gtx, i.theme, "\u2003").Size.X * 6
		}
		// need to add decorator width
		mi := op.Record(gtx.Ops)
		iDims := i.style.Layout(gtx)
		_ = mi.Stop()
		macro := op.Record(gtx.Ops)
		dgtx := gtx
		dgtx.Constraints.Max.Y = iDims.Size.Y + gtx.Dp(4)
		dDims := i.decorator(dgtx)
		gtx.Constraints.Max.X = maxWd + dDims.Size.X + 4
		gtx.Constraints.Min.X = maxWd + dDims.Size.X + 4
		dCall := macro.Stop()
		return widget.Border{Color: bc, CornerRadius: 3, Width: bt}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Gap: 4, Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X, gtx.Constraints.Max.X = maxWd, maxWd
					return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, i.style.Layout)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					dCall.Add(gtx.Ops)
					lineHt := iDims.Size.Y + (gtx.Dp(2) * 2)
					paint.FillShape(gtx.Ops, Style.BorderColor, clip.Rect{Max: image.Point{X: gtx.Dp(1), Y: lineHt}}.Op())
					return dDims
				}),
			)
		})
	}
	if maxWd > 0 {
		if maxWd < gtx.Constraints.Max.X {
			gtx.Constraints.Max.X = maxWd
		}
		if gtx.Constraints.Min.X > gtx.Constraints.Max.X {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
		}
	}
	return widget.Border{Color: bc, CornerRadius: 3, Width: bt}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, i.style.Layout)
	})
}

func (i *Input) Update(gtx layout.Context) {
	currentlyFocused := gtx.Focused(&i.editor)
	if currentlyFocused && (i.upDownFn != nil || i.decoratorOnClick != nil) {
		filters := make([]event.Filter, 0, 6)
		if i.upDownFn != nil {
			filters = append(filters, pointer.Filter{
				Target:  &i.editor,
				Kinds:   pointer.Scroll,
				ScrollY: pointer.ScrollRange{Min: 0, Max: 100000},
			},
				key.Filter{Focus: &i.editor, Name: key.NameUpArrow},
				key.Filter{Focus: &i.editor, Name: key.NameDownArrow},
				key.Filter{Focus: &i.editor, Name: key.NamePageUp},
				key.Filter{Focus: &i.editor, Name: key.NamePageDown})
		}
		if i.decoratorOnClick != nil {
			filters = append(filters, key.Filter{Focus: &i.editor, Required: key.ModCtrl, Name: key.NameEnter},
				key.Filter{Focus: &i.editor, Required: key.ModCtrl, Name: key.NameReturn})
		}
		for {
			ev, ok := gtx.Event(filters...)
			if !ok {
				break
			}
			switch evt := ev.(type) {
			case key.Event:
				if evt.State == key.Press {
					switch evt.Name {
					case key.NameEnter, key.NameReturn:
						if i.decoratorOnClick != nil {
							i.decoratorOnClick()
						}
					case key.NameUpArrow, key.NameDownArrow, key.NamePageUp, key.NamePageDown:
						if i.upDownFn != nil {
							if s, ok := i.upDownFn(evt.Name, i.editor.Text()); ok {
								i.SetText(s)
							}
						}
					}
				}
			case pointer.Event:
				if i.upDownFn != nil {
					if evt.Scroll.Y == 0 {
						if s, ok := i.upDownFn(key.NameUpArrow, i.editor.Text()); ok {
							i.SetText(s)
						}
					} else {
						if s, ok := i.upDownFn(key.NameDownArrow, i.editor.Text()); ok {
							i.SetText(s)
						}
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
			if i.onSubmitFn != nil {
				i.onSubmitFn(i.editor.Text())
			}
		}
	}
	if currentlyFocused != gtx.Focused(&i.editor) {
		gtx.Execute(op.InvalidateCmd{})
	}
}

func (i *Input) WithDecorator(w layout.Widget) *Input {
	i.decorator = w
	return i
}

func (i *Input) WithEllipsisButton(onClick func()) *Input {
	if onClick == nil {
		i.decorator = nil
		return i
	}
	clickable := Clickable{}
	lbl := material.Label(i.theme, i.theme.TextSize, "…")
	lbl.MaxLines = 1
	lbl.Alignment = text.Middle
	i.decoratorOnClick = onClick
	i.decorator = func(gtx layout.Context) layout.Dimensions {
		if clickable.Clicked(gtx) {
			onClick()
		}
		return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			macro := op.Record(gtx.Ops)
			dims := layout.Inset{Left: 2, Right: 2, Bottom: 2}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
				return layout.S.Layout(gtx, lbl.Layout)
			})
			call := macro.Stop()
			if clickable.Hovered() {
				paint.FillShape(gtx.Ops, Style.ShadeColor, clip.Rect{Max: image.Point{X: dims.Size.X, Y: dims.Size.Y}}.Op())
			}
			call.Add(gtx.Ops)
			return dims
		})
	}
	return i
}

type InputFilterType int

func filterFromType(v any) InputFilterType {
	switch v.(type) {
	case int:
		return FilterInt
	case uint:
		return FilterUint
	case float64:

		return FilterFloat
	}
	return FilterNone
}

func (f InputFilterType) String() string {
	switch f {
	case FilterAlphabet:
		return "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	case FilterAlphabetUpper:
		return "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case FilterAlphabetLower:
		return "abcdefghijklmnopqrstuvwxyz"
	case FilterInt:
		return "-0123456789"
	case FilterUint:
		return "0123456789"
	case FilterFloat:
		return "-.0123456789"
	case FilterFloatPositive:
		return ".0123456789"
	}
	return ""
}

const (
	FilterNone InputFilterType = iota
	FilterAlphabet
	FilterAlphabetLower
	FilterAlphabetUpper
	FilterInt
	FilterUint
	FilterFloat
	FilterFloatPositive
)

func buildFilterString(filter any) string {
	switch ft := filter.(type) {
	case string:
		return ft
	case InputFilterType:
		return ft.String()
	}
	return ""
}
