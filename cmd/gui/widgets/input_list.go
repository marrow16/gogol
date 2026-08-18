package widgets

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
)

type listControl[T any] struct {
	border        bool
	theme         *material.Theme
	list          widget.List
	clickables    []widget.Clickable
	selectedIndex int
	focused       bool
	tag           struct{}
	items         []T
	rowRenderFn   func(gtx layout.Context, index int, item T) layout.Dimensions
	selectItemFn  func(item T)
	isSelectedFn  func(index int, item T) bool
}

func newListControl[T any](theme *material.Theme, items []T, border bool) *listControl[T] {
	result := &listControl[T]{
		border: border,
		theme:  theme,
		items:  items,
	}
	result.list.Axis = layout.Vertical
	result.isSelectedFn = result.isSelected
	return result
}

func (l *listControl[T]) isSelected(index int, item T) bool {
	return l.selectedIndex == index
}

func (l *listControl[T]) rowRenderer(fn func(gtx layout.Context, index int, item T) layout.Dimensions) *listControl[T] {
	l.rowRenderFn = fn
	return l
}

func (l *listControl[T]) onItemSelect(fn func(item T)) *listControl[T] {
	l.selectItemFn = fn
	return l
}

func (l *listControl[T]) onIsSelected(fn func(index int, item T) bool) *listControl[T] {
	l.isSelectedFn = fn
	return l
}

func (l *listControl[T]) resetItems(items []T) {
	l.items = items
	l.clickables = make([]widget.Clickable, 0)
}

func (l *listControl[T]) isFocused(gtx layout.Context) bool {
	return l.focused
}

func (l *listControl[T]) Layout(gtx layout.Context) layout.Dimensions {
	if len(l.clickables) != len(l.items) {
		l.clickables = make([]widget.Clickable, len(l.items))
	}
	for {
		ev, ok := gtx.Event(
			key.FocusFilter{Target: &l.tag},
			key.Filter{Focus: &l.tag, Name: key.NameDownArrow},
			key.Filter{Focus: &l.tag, Name: key.NameUpArrow},
			key.Filter{Focus: &l.tag, Name: key.NameReturn},
			key.Filter{Focus: &l.tag, Name: key.NameEnter},
			key.Filter{Focus: &l.tag, Name: key.NameSpace},
			key.Filter{Focus: &l.tag, Name: key.NamePageUp},
			key.Filter{Focus: &l.tag, Name: key.NamePageDown},
			key.Filter{Focus: &l.tag, Name: key.NameHome},
			key.Filter{Focus: &l.tag, Name: key.NameEnd},
		)
		if !ok {
			break
		}
		switch evt := ev.(type) {
		case key.FocusEvent:
			l.focused = evt.Focus
		case key.Event:
			if evt.State != key.Press {
				continue
			}
			switch evt.Name {
			case key.NameDownArrow:
				if l.selectedIndex < len(l.items)-1 {
					l.selectedIndex++
					l.list.Position.Offset = 0
					l.list.ScrollTo(l.selectedIndex)
					if l.selectItemFn != nil {
						l.selectItemFn(l.items[l.selectedIndex])
					}
				}
			case key.NameUpArrow:
				if l.selectedIndex > 0 {
					l.selectedIndex--
					l.list.Position.Offset = 0
					l.list.ScrollTo(l.selectedIndex)
					if l.selectItemFn != nil {
						l.selectItemFn(l.items[l.selectedIndex])
					}
				}
			case key.NamePageUp:
				np := l.selectedIndex - (l.list.Position.Count - 1)
				if np < 0 {
					np = 0
				}
				l.selectedIndex = np
				l.list.Position.Offset = 0
				l.list.ScrollTo(l.selectedIndex)
				if l.selectItemFn != nil {
					l.selectItemFn(l.items[l.selectedIndex])
				}
			case key.NamePageDown:
				np := l.selectedIndex + (l.list.Position.Count - 1)
				if np >= len(l.items) {
					np = len(l.items) - 1
				}
				l.selectedIndex = np
				l.list.Position.Offset = 0
				l.list.ScrollTo(l.selectedIndex)
				if l.selectItemFn != nil {
					l.selectItemFn(l.items[l.selectedIndex])
				}
			case key.NameHome:
				l.selectedIndex = 0
				l.list.Position.Offset = 0
				l.list.ScrollTo(l.selectedIndex)
				if l.selectItemFn != nil {
					l.selectItemFn(l.items[l.selectedIndex])
				}
			case key.NameEnd:
				l.selectedIndex = len(l.items) - 1
				l.list.Position.Offset = 0
				l.list.ScrollTo(l.selectedIndex)
				if l.selectItemFn != nil {
					l.selectItemFn(l.items[l.selectedIndex])
				}
			case key.NameReturn, key.NameEnter, key.NameSpace:
				if l.selectedIndex >= 0 &&
					l.selectedIndex < len(l.items) &&
					l.selectItemFn != nil {
					l.selectItemFn(l.items[l.selectedIndex])
				}
			}
		}
	}
	macro := op.Record(gtx.Ops)
	var dims layout.Dimensions
	if l.border {
		clr := popupBorder
		bw := unit.Dp(1)
		if l.focused {
			clr = popupBorderFocused
			bw = unit.Dp(2)
		}
		dims = widget.Border{
			Color:        clr,
			Width:        bw,
			CornerRadius: unit.Dp(3),
		}.Layout(gtx, l.layoutList)
	} else {
		dims = l.layoutList(gtx)
	}
	call := macro.Stop()
	defer clip.Rect{
		Max: dims.Size,
	}.Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, &l.tag)
	call.Add(gtx.Ops)
	return dims
}

func (l *listControl[T]) layoutList(gtx layout.Context) layout.Dimensions {
	rowDims := measureText(gtx, l.theme, "Xy")
	return material.List(l.theme, &l.list).Layout(
		gtx,
		len(l.items),
		func(gtx layout.Context, index int) layout.Dimensions {
			btn := &l.clickables[index]
			if btn.Clicked(gtx) {
				l.selectedIndex = index
				gtx.Execute(key.FocusCmd{Tag: &l.tag})
				if l.selectItemFn != nil {
					l.selectItemFn(l.items[index])
				}
			}
			if l.isSelectedFn(index, l.items[index]) {
				l.selectedIndex = index
				bg := popupSelectedBackground
				if l.focused {
					bg = popupSelectedFocusedBackground
				}
				paint.FillShape(
					gtx.Ops,
					bg,
					clip.Rect{
						Max: image.Pt(gtx.Constraints.Max.X, rowDims.Size.Y),
					}.Op(),
				)
			}
			return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return l.rowRenderFn(gtx, index, l.items[index])
			})
		},
	)
}
