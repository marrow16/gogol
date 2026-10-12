package controls

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
)

type List[T any] struct {
	Items          []T
	border         bool
	theme          *material.Theme
	list           widget.List
	clickables     []widget.Clickable
	selectedIndex  int
	focused        bool
	tag            struct{}
	rowRenderFn    func(gtx layout.Context, index int, item T) layout.Dimensions
	selectItemFn   func(item T, keyboard bool)
	navigateItemFn func(item T)
	isSelectedFn   func(index int, item T) bool
}

func NewList[T any](theme *material.Theme, items []T, border bool) *List[T] {
	result := &List[T]{
		border: border,
		theme:  theme,
		Items:  items,
	}
	result.list.Axis = layout.Vertical
	result.isSelectedFn = result.IsSelected
	return result
}

func (l *List[T]) Focus(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: &l.tag})
}

func (l *List[T]) IsSelected(index int, _ T) bool {
	return l.selectedIndex == index
}

func (l *List[T]) SetSelected(index int) {
	if index >= 0 && index < len(l.Items) {
		l.selectedIndex = index
	} else {
		l.selectedIndex = 0
	}
}

func (l *List[T]) Selected() int {
	return l.selectedIndex
}

func (l *List[T]) RowRenderer(fn func(gtx layout.Context, index int, item T) layout.Dimensions) *List[T] {
	l.rowRenderFn = fn
	return l
}

func (l *List[T]) OnItemSelect(fn func(item T, keyboard bool)) *List[T] {
	l.selectItemFn = fn
	return l
}

func (l *List[T]) OnItemNavigate(fn func(item T)) *List[T] {
	l.navigateItemFn = fn
	return l
}

func (l *List[T]) OnIsSelected(fn func(index int, item T) bool) *List[T] {
	l.isSelectedFn = fn
	return l
}

func (l *List[T]) ResetItems(items []T) {
	l.selectedIndex = 0
	l.Items = items
	l.clickables = make([]widget.Clickable, 0)
}

func (l *List[T]) IsFocused(_ layout.Context) bool {
	return l.focused
}

func (l *List[T]) Layout(gtx layout.Context) layout.Dimensions {
	if len(l.clickables) != len(l.Items) {
		l.clickables = make([]widget.Clickable, len(l.Items))
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
				if l.selectedIndex < len(l.Items)-1 {
					l.selectedIndex++
					l.EnsureVisible(l.selectedIndex)
					if l.navigateItemFn != nil {
						l.navigateItemFn(l.Items[l.selectedIndex])
					} else if l.selectItemFn != nil {
						l.selectItemFn(l.Items[l.selectedIndex], true)
					}
				}
			case key.NameUpArrow:
				if l.selectedIndex > 0 {
					l.selectedIndex--
					l.EnsureVisible(l.selectedIndex)
					if l.navigateItemFn != nil {
						l.navigateItemFn(l.Items[l.selectedIndex])
					} else if l.selectItemFn != nil {
						l.selectItemFn(l.Items[l.selectedIndex], true)
					}
				}
			case key.NamePageUp:
				np := max(l.selectedIndex-(l.list.Position.Count-1), 0)
				l.selectedIndex = np
				l.EnsureVisible(l.selectedIndex)
				if l.navigateItemFn != nil {
					l.navigateItemFn(l.Items[l.selectedIndex])
				} else if l.selectItemFn != nil {
					l.selectItemFn(l.Items[l.selectedIndex], true)
				}
			case key.NamePageDown:
				np := l.selectedIndex + (l.list.Position.Count - 1)
				if np >= len(l.Items) {
					np = len(l.Items) - 1
				}
				l.selectedIndex = np
				l.EnsureVisible(l.selectedIndex)
				if l.navigateItemFn != nil {
					l.navigateItemFn(l.Items[l.selectedIndex])
				} else if l.selectItemFn != nil {
					l.selectItemFn(l.Items[l.selectedIndex], true)
				}
			case key.NameHome:
				l.selectedIndex = 0
				l.EnsureVisible(l.selectedIndex)
				if l.navigateItemFn != nil {
					l.navigateItemFn(l.Items[l.selectedIndex])
				} else if l.selectItemFn != nil {
					l.selectItemFn(l.Items[l.selectedIndex], true)
				}
			case key.NameEnd:
				l.selectedIndex = len(l.Items) - 1
				l.EnsureVisible(l.selectedIndex)
				if l.navigateItemFn != nil {
					l.navigateItemFn(l.Items[l.selectedIndex])
				} else if l.selectItemFn != nil {
					l.selectItemFn(l.Items[l.selectedIndex], true)
				}
			case key.NameReturn, key.NameEnter, key.NameSpace:
				if l.selectedIndex >= 0 &&
					l.selectedIndex < len(l.Items) && l.selectItemFn != nil {
					l.selectItemFn(l.Items[l.selectedIndex], true)
				}
			}
		}
	}
	macro := op.Record(gtx.Ops)
	var dims layout.Dimensions
	if l.border {
		bc, bt := Style.bordering(l.focused)
		dims = widget.Border{Color: bc, Width: bt, CornerRadius: 3}.Layout(gtx, l.layoutList)
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

func (l *List[T]) EnsureVisible(idx int) {
	if idx >= (l.list.Position.First + l.list.Position.Count - 1) {
		l.list.ScrollTo(idx)
	}
	if l.selectedIndex < l.list.Position.First {
		l.list.ScrollTo(idx)
	}
}

func (l *List[T]) ScrollTo(idx int) {
	l.list.ScrollTo(idx)
}

func (l *List[T]) layoutList(gtx layout.Context) layout.Dimensions {
	return material.List(l.theme, &l.list).Layout(
		gtx,
		len(l.Items),
		func(gtx layout.Context, index int) layout.Dimensions {
			btn := &l.clickables[index]
			if btn.Clicked(gtx) {
				l.selectedIndex = index
				gtx.Execute(key.FocusCmd{Tag: &l.tag})
				if l.selectItemFn != nil {
					l.selectItemFn(l.Items[index], false)
				}
			}
			return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				macro := op.Record(gtx.Ops)
				dims := l.rowRenderFn(gtx, index, l.Items[index])
				call := macro.Stop()
				if l.isSelectedFn(index, l.Items[index]) {
					l.selectedIndex = index
					bg := Style.SelectedColor
					if l.focused {
						bg = Style.SelectedFocusedColor
					}
					paint.FillShape(gtx.Ops, bg, clip.Rect{Max: image.Point{X: gtx.Constraints.Max.X, Y: dims.Size.Y}}.Op())
				}
				call.Add(gtx.Ops)
				return dims
			})
		},
	)
}
