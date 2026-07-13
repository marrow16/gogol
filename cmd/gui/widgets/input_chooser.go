package widgets

import (
	"fmt"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
	"slices"
	"strings"
)

type chooser[T any] struct {
	labeller       func(T) string
	theme          *material.Theme
	dropdownRows   int
	maxWidth       int
	editor         widget.Editor
	button         widget.Clickable
	list           widget.List
	opened         bool
	items          []T
	filteredItems  []T
	rowClicks      []widget.Clickable
	dims           layout.Dimensions
	onChangeFn     func(item *T)
	onSubmitFn     func(text string)
	noRefilter     bool
	middleEllipsis bool
}

func defaultLabeller(item any) string {
	switch it := item.(type) {
	case string:
		return it
	case fmt.Stringer:
		return it.String()
	default:
		return fmt.Sprintf("%v", item)
	}
}

func newChooser[T any](theme *material.Theme, maxWidth int, items []T, onChange func(item *T), labeller func(T) string) *chooser[T] {
	result := &chooser[T]{
		labeller: labeller,
		editor: widget.Editor{
			Alignment:  text.Start,
			SingleLine: true,
			Submit:     true,
		},
		theme:         theme,
		dropdownRows:  10,
		maxWidth:      maxWidth,
		items:         items,
		filteredItems: items,
		onChangeFn:    onChange,
	}
	if labeller == nil {
		result.labeller = func(item T) string {
			return defaultLabeller(item)
		}
	}
	return result
}

func (i *chooser[T]) onSubmit(fn func(text string)) *chooser[T] {
	i.onSubmitFn = fn
	return i
}

func (i *chooser[T]) currentItem() *T {
	curr := i.editor.Text()
	if idx := slices.IndexFunc(i.items, func(item T) bool {
		return i.labeller(item) == curr
	}); idx >= 0 {
		return &i.items[idx]
	}
	return nil
}

func (i *chooser[T]) resetItems(items []T) {
	i.items = items
	i.filteredItems = items
}

func (i *chooser[T]) layout(gtx layout.Context) layout.Dimensions {
	i.update(gtx)
	borderColor := popupBorder
	borderThickness := unit.Dp(1)
	if i.isFocused(gtx) {
		borderColor = popupBorderFocused
		borderThickness = unit.Dp(2)
	}
	i.dims = widget.Border{
		Color:        borderColor,
		CornerRadius: unit.Dp(3),
		Width:        borderThickness,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(2),
			Bottom: unit.Dp(2),
			Left:   unit.Dp(4),
			Right:  unit.Dp(4),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					maxChWd := measureText(gtx, i.theme, "M")
					maxWidth := maxChWd.Size.X * i.maxWidth
					if maxWidth < gtx.Constraints.Max.X {
						gtx.Constraints.Max.X = maxWidth
					}
					if gtx.Constraints.Min.X > gtx.Constraints.Max.X {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
					}
					return material.Editor(i.theme, &i.editor, "").Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					ch := "▼"
					if i.opened {
						ch = "▲"
					}
					btn := material.Button(i.theme, &i.button, ch)
					btn.Inset = layout.Inset{Top: 3, Bottom: 3, Left: 3, Right: 3}
					btn.Background = popupBackground
					btn.Color = popupForeground
					btn.TextSize = unit.Sp(16)
					return btn.Layout(gtx)
				}),
			)
		})
	})
	return i.dims
}

func (i *chooser[T]) update(gtx layout.Context) {
	currentlyFocused := gtx.Focused(&i.editor)
	if i.button.Clicked(gtx) {
		i.opened = !i.opened
		if i.opened {
			gtx.Execute(key.FocusCmd{Tag: &i.editor})
		}
	}
	i.handleKeys(gtx)
	for {
		ev, ok := i.editor.Update(gtx)
		if !ok {
			break
		}
		switch ev.(type) {
		case widget.ChangeEvent:
			i.searchAndFilter()
			i.changed()
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

func (i *chooser[T]) handleKeys(gtx layout.Context) {
	if i.opened {
		i.handleOpenedKeys(gtx)
		return
	}
	for {
		ev, ok := gtx.Event(
			key.Filter{Focus: &i.editor, Name: key.NameUpArrow},
			key.Filter{Focus: &i.editor, Name: key.NameDownArrow},
		)
		if !ok {
			break
		}
		if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
			curr := i.editor.Text()
			idx := 0
			if len(curr) > 0 {
				idx = slices.IndexFunc(i.items, func(item T) bool {
					return strings.EqualFold(i.labeller(item), curr)
				})
				if idx == -1 {
					// no exact find, find starts with...
					curr = strings.ToLower(curr)
					idx = slices.IndexFunc(i.filteredItems, func(item T) bool {
						return strings.HasPrefix(strings.ToLower(i.labeller(item)), curr)
					})
					if idx != -1 {
						lbl := i.labeller(i.items[idx])
						i.editor.SetText(lbl)
						i.editor.SetCaret(0, len(lbl))
					}
					return
				}
			}
			maxi := len(i.items) - 1
			switch kev.Name {
			case key.NameUpArrow:
				if idx-1 > 0 {
					idx--
				} else {
					idx = maxi
				}
			case key.NameDownArrow:
				if idx < maxi {
					idx++
				} else {
					idx = 0
				}
			}
			if idx >= 0 && idx <= maxi {
				lbl := i.labeller(i.items[idx])
				i.editor.SetText(lbl)
				i.editor.SetCaret(0, len(lbl))
			}
		}
	}
}

func (i *chooser[T]) handleOpenedKeys(gtx layout.Context) {
	for {
		ev, ok := gtx.Event(
			key.Filter{Name: key.NameUpArrow},
			key.Filter{Name: key.NameDownArrow},
			key.Filter{Name: key.NamePageUp},
			key.Filter{Name: key.NamePageDown},
			key.Filter{Name: key.NameHome},
			key.Filter{Name: key.NameEnd})
		if !ok {
			break
		}
		if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
			curr := i.editor.Text()
			maxi := len(i.filteredItems) - 1
			idx := 0
			if len(curr) > 0 {
				idx = slices.IndexFunc(i.filteredItems, func(item T) bool {
					return strings.EqualFold(i.labeller(item), curr)
				})
			}
			switch kev.Name {
			case key.NameUpArrow:
				if idx-1 > 0 {
					idx--
				} else {
					idx = maxi
				}
			case key.NameDownArrow:
				if idx < maxi {
					idx++
				} else {
					idx = 0
				}
			case key.NamePageUp:
				if idx-(i.dropdownRows-1) > 0 {
					idx -= i.dropdownRows - 1
				} else {
					idx = 0
				}
			case key.NamePageDown:
				if idx+(i.dropdownRows-1) <= maxi {
					idx += i.dropdownRows - 1
				} else {
					idx = maxi
				}
			case key.NameHome:
				idx = 0
			case key.NameEnd:
				idx = maxi
			}
			if idx >= 0 && idx <= maxi {
				i.noRefilter = true
				i.list.ScrollTo(idx)
				lbl := i.labeller(i.filteredItems[idx])
				i.editor.SetText(lbl)
				i.editor.SetCaret(0, len(lbl))
			}
		}
	}
}

func (i *chooser[T]) searchAndFilter() {
	if i.noRefilter {
		i.noRefilter = false
		return
	}
	curr := i.editor.Text()
	curr = strings.ToLower(curr)
	switch len(curr) {
	case 0:
		i.filteredItems = i.items
	case 1:
		// just scroll to first that begins with this letter
		i.filteredItems = i.items
		if idx := slices.IndexFunc(i.filteredItems, func(item T) bool {
			return strings.HasPrefix(strings.ToLower(i.labeller(item)), curr)
		}); idx != -1 {
			i.list.ScrollTo(idx)
		}
	default:
		// filter list down to those that contain this value
		startsWith := make([]T, 0, len(i.items))
		contains := make([]T, 0, len(i.items))
		for _, item := range i.items {
			label := strings.ToLower(i.labeller(item))
			if strings.HasPrefix(label, curr) {
				startsWith = append(startsWith, item)
			} else if strings.Contains(label, curr) {
				contains = append(contains, item)
			}
		}
		i.filteredItems = append(startsWith, contains...)
		i.list.ScrollTo(0)
	}
}

func (i *chooser[T]) changed() {
	if i.onChangeFn != nil {
		curr := i.editor.Text()
		idx := slices.IndexFunc(i.filteredItems, func(item T) bool {
			return i.labeller(item) == curr
		})
		if idx != -1 {
			i.onChangeFn(&i.filteredItems[idx])
		} else {
			i.onChangeFn(nil)
		}
	}
}

func (i *chooser[T]) layoutDropdown(gtx layout.Context, dims layout.Dimensions) {
	if i.opened && len(i.filteredItems) > 0 {
		stack := op.Offset(image.Pt(0, dims.Size.Y)).Push(gtx.Ops)
		rowDims := measureText(gtx, i.theme, "Xy")
		showRows := i.dropdownRows
		if len(i.filteredItems) < showRows {
			showRows = len(i.filteredItems)
		}
		height := rowDims.Size.Y * showRows
		width := dims.Size.X
		layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				pgtx := gtx
				pgtx.Constraints.Min.X = width
				pgtx.Constraints.Max.X = width
				pgtx.Constraints.Min.Y = height
				pgtx.Constraints.Max.Y = height
				paint.FillShape(
					pgtx.Ops,
					popupBackground,
					clip.Rect{Max: image.Pt(width, height)}.Op(),
				)
				i.list.Axis = layout.Vertical
				if len(i.rowClicks) != len(i.filteredItems) {
					i.rowClicks = make([]widget.Clickable, len(i.filteredItems))
				}
				listDims := material.List(i.theme, &i.list).Layout(pgtx, len(i.filteredItems), func(gtx layout.Context, index int) layout.Dimensions {
					lbl := i.labeller(i.filteredItems[index])
					for i.rowClicks[index].Clicked(gtx) {
						i.setText(lbl)
						gtx.Execute(op.InvalidateCmd{})
						gtx.Execute(key.FocusCmd{Tag: &i.list})
						i.opened = false
					}
					return i.rowClicks[index].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						if lbl == i.editor.Text() {
							paint.FillShape(
								gtx.Ops,
								popupSelectedBackground,
								clip.Rect{Max: image.Pt(width, rowDims.Size.Y)}.Op(),
							)
						}
						gtx.Constraints.Min.X = pgtx.Constraints.Max.X
						return layout.Inset{Left: 3, Right: 3}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							if i.middleEllipsis {
								return label(i.theme, middleEllipsis(gtx, i.theme, lbl, gtx.Constraints.Max.X))(gtx)
							} else {
								return label(i.theme, lbl)(gtx)
							}
						})
					})
				})
				border(pgtx, layout.Dimensions{
					Size:     image.Point{X: width, Y: height},
					Baseline: 0,
				}, true, true, true, true)
				return listDims
			}),
		)
		stack.Pop()
	}
}

func middleEllipsis(gtx layout.Context, theme *material.Theme, s string, maxWidth int) string {
	if measureText(gtx, theme, s).Size.X <= maxWidth {
		return s
	}
	r := []rune(s)
	left := len(r) / 2
	right := left
	for left > 0 || right < len(r) {
		if left > 0 {
			left--
		}
		if right < len(r) {
			right++
		}

		t := string(r[:left]) + "…" + string(r[right:])
		if measureText(gtx, theme, t).Size.X <= maxWidth {
			return t
		}
	}
	return "…"
}

func (i *chooser[T]) setText(text string) {
	if text != i.editor.Text() {
		i.editor.SetText(text)
		i.editor.SetCaret(0, len(text))
	}
}

func (i *chooser[T]) setFocused(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: &i.editor})
}

func (i *chooser[T]) isFocused(gtx layout.Context) bool {
	return gtx.Focused(&i.editor) || i.opened
}
