package controls

import (
	"fmt"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
	"image/color"
	"slices"
	"strings"
)

type Chooser[T any] struct {
	theme          *material.Theme
	labeller       func(T) string
	dropdownRows   int
	maxWidth       int
	editor         widget.Editor
	editorStyle    material.EditorStyle
	button         Clickable
	list           widget.List
	listStyle      material.ListStyle
	opened         bool
	items          []T
	filteredItems  []T
	rowClicks      []Clickable
	dims           layout.Dimensions
	onChangeFn     func(item *T)
	onSubmitFn     func(text string)
	noRefilter     bool
	middleEllipsis bool
	dropdownUp     bool
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

func NewChooser[T any](theme *material.Theme, maxWidth int, items []T, onChange func(item *T), labeller func(T) string) *Chooser[T] {
	result := &Chooser[T]{
		theme:    theme,
		labeller: labeller,
		editor: widget.Editor{
			Alignment:  text.Start,
			SingleLine: true,
			Submit:     true,
		},
		dropdownRows:  10,
		maxWidth:      maxWidth,
		items:         items,
		filteredItems: items,
		onChangeFn:    onChange,
	}
	result.listStyle = material.List(theme, &result.list)
	if labeller == nil {
		result.labeller = func(item T) string {
			return defaultLabeller(item)
		}
	}
	result.editorStyle = material.Editor(theme, &result.editor, "")
	return result
}

func (i *Chooser[T]) MaxRows(r int) *Chooser[T] {
	i.dropdownRows = max(r, 3)
	return i
}

func (i *Chooser[T]) OnSubmit(fn func(text string)) *Chooser[T] {
	i.onSubmitFn = fn
	return i
}

func (i *Chooser[T]) Opened(opened bool) {
	i.opened = opened
}

func (i *Chooser[T]) MiddleEllipsis(me bool) *Chooser[T] {
	i.middleEllipsis = me
	return i
}

func (i *Chooser[T]) Direction(up bool) *Chooser[T] {
	i.dropdownUp = up
	return i
}

func (i *Chooser[T]) CurrentItem() *T {
	curr := i.editor.Text()
	if idx := slices.IndexFunc(i.items, func(item T) bool {
		return i.labeller(item) == curr
	}); idx >= 0 {
		return &i.items[idx]
	}
	return nil
}

func (i *Chooser[T]) Text() string {
	return i.editor.Text()
}

func (i *Chooser[T]) Len() int {
	return len(i.items)
}

func (i *Chooser[T]) ResetItems(items []T) {
	i.items = items
	i.filteredItems = items
}

func (i *Chooser[T]) Width() int {
	return i.dims.Size.X
}

func (i *Chooser[T]) Layout(gtx layout.Context) layout.Dimensions {
	i.update(gtx)
	bc, bt := Style.bordering(i.IsFocused(gtx))
	i.dims = widget.Border{Color: bc, CornerRadius: 3, Width: bt}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		bw := 0
		dims := layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Gap: 10, Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					maxChWd := measureText(gtx, i.theme, "\u2003")
					maxWidth := maxChWd.Size.X * i.maxWidth
					if maxWidth < gtx.Constraints.Max.X {
						gtx.Constraints.Max.X = maxWidth
					}
					if gtx.Constraints.Min.X > gtx.Constraints.Max.X {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
					}
					return i.editorStyle.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					ch := "▼"
					if i.opened {
						ch = "▲"
					}
					macro := op.Record(gtx.Ops)
					dims := i.button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(i.theme, i.theme.TextSize, ch)
						lbl.Alignment = text.Middle
						lbl.MaxLines = 1
						return lbl.Layout(gtx)
					})
					call := macro.Stop()
					if i.button.Hovered() {
						paint.FillShape(gtx.Ops, Style.ShadeColor, clip.Rect{Max: image.Point{X: dims.Size.X, Y: dims.Size.Y}}.Op())
					}
					call.Add(gtx.Ops)
					bw = dims.Size.X
					return dims
				}),
			)
		})
		bl := dims.Size.X - bw - gtx.Dp(6)
		paint.FillShape(gtx.Ops, Style.BorderColor, clip.Rect{Min: image.Point{X: bl}, Max: image.Point{X: bl + gtx.Dp(1), Y: dims.Size.Y}}.Op())
		return dims
	})
	return i.dims
}

func (i *Chooser[T]) update(gtx layout.Context) {
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
			} else if !i.opened && len(i.filteredItems) > 0 {
				i.opened = true
			}
		}
	}
	if currentlyFocused != gtx.Focused(&i.editor) {
		gtx.Execute(op.InvalidateCmd{})
	}
}

func (i *Chooser[T]) handleKeys(gtx layout.Context) {
	if i.opened {
		i.handleOpenedKeys(gtx)
		return
	}
	for {
		ev, ok := gtx.Event(
			key.Filter{Focus: &i.editor, Name: key.NameUpArrow},
			key.Filter{Focus: &i.editor, Name: key.NameDownArrow},
			key.Filter{Focus: &i.editor, Name: key.NamePageUp},
			key.Filter{Focus: &i.editor, Name: key.NamePageDown},
			key.Filter{Focus: &i.editor, Name: key.NameHome},
			key.Filter{Focus: &i.editor, Name: key.NameEnd})
		if !ok {
			break
		}
		if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
			curr := i.editor.Text()
			currL := strings.ToLower(curr)
			idx := 0
			comp := 0
			compS := 0
			if len(curr) > 0 && kev.Name != key.NameHome && kev.Name != key.NameEnd {
				idx = -1
				for n := range len(i.items) {
					lbl := i.labeller(i.items[n])
					if lbl == curr {
						idx = n
						break
					}
					lbl = strings.ToLower(lbl)
					if sc := stringSimilarity(lbl, currL); sc > compS {
						compS = sc
						comp = n
					}
				}
			}
			maxi := len(i.items) - 1
			switch kev.Name {
			case key.NameUpArrow:
				if idx == -1 {
					idx = comp
				} else if idx > 0 {
					idx--
				} else {
					idx = maxi
				}
			case key.NameDownArrow:
				if idx == -1 {
					idx = comp
				} else if idx < maxi {
					idx++
				} else {
					idx = 0
				}
			case key.NamePageUp:
				if idx == -1 {
					idx = comp
				} else if idx-(i.dropdownRows-1) > 0 {
					idx -= i.dropdownRows - 1
				} else {
					idx = 0
				}
			case key.NamePageDown:
				if idx == -1 {
					idx = comp
				} else if idx+(i.dropdownRows-1) <= maxi {
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
				lbl := i.labeller(i.items[idx])
				i.editor.SetText(lbl)
				i.editor.SetCaret(len(lbl), 0)
			}
		}
	}
}

func (i *Chooser[T]) handleOpenedKeys(gtx layout.Context) {
	for {
		ev, ok := gtx.Event(
			key.Filter{Name: key.NameReturn},
			key.Filter{Name: key.NameEnter},
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
			currL := strings.ToLower(curr)
			maxi := len(i.filteredItems) - 1
			idx := 0
			comp := 0
			compS := 0
			if len(curr) > 0 && kev.Name != key.NameHome && kev.Name != key.NameEnd {
				idx = -1
				for n := range len(i.filteredItems) {
					lbl := i.labeller(i.filteredItems[n])
					if lbl == curr {
						idx = n
						break
					}
					lbl = strings.ToLower(lbl)
					if sc := stringSimilarity(lbl, currL); sc > compS {
						compS = sc
						comp = n
					}
				}
			}
			switch kev.Name {
			case key.NameEnter, key.NameReturn:
				if idx != -1 {
					i.opened = false
				}
			case key.NameUpArrow:
				if idx == -1 {
					idx = comp
				} else if idx > 0 {
					idx--
				} else {
					idx = maxi
				}
			case key.NameDownArrow:
				if idx == -1 {
					idx = comp
				} else if idx < maxi {
					idx++
				} else {
					idx = 0
				}
			case key.NamePageUp:
				if idx == -1 {
					comp -= i.dropdownRows - 1
				} else if idx-(i.dropdownRows-1) > 0 {
					idx -= i.dropdownRows - 1
				} else {
					idx = 0
				}
			case key.NamePageDown:
				if idx == -1 {
					comp += i.dropdownRows - 1
				} else if idx+(i.dropdownRows-1) <= maxi {
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
				i.editor.SetCaret(len(lbl), 0)
			} else {
				comp = max(min(comp, len(i.filteredItems)-1), 0)
				i.list.ScrollTo(comp)
			}
		}
	}
}

func (i *Chooser[T]) searchAndFilter() {
	if i.noRefilter {
		i.noRefilter = false
		return
	}
	actual := i.editor.Text()
	curr := strings.ToLower(actual)
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
		if idx := slices.IndexFunc(i.items, func(item T) bool {
			return i.labeller(item) == actual
		}); idx != -1 {
			i.filteredItems = i.items
			i.list.ScrollTo(idx)
		} else {
			// filter list down to those that contain this value
			startsWith := make([]T, 0, len(i.items))
			contains := make([]T, 0, len(i.items))
			comp := 0
			compS := 0
			for n, item := range i.items {
				lbl := strings.ToLower(i.labeller(item))
				if strings.HasPrefix(lbl, curr) {
					startsWith = append(startsWith, item)
				} else if strings.Contains(lbl, curr) {
					contains = append(contains, item)
				} else if sc := stringSimilarity(lbl, curr); sc > compS {
					compS = sc
					comp = n
				}
			}
			if len(startsWith) == 0 && len(contains) == 0 {
				i.filteredItems = i.items
				i.list.ScrollTo(comp)
			} else {
				i.filteredItems = append(startsWith, contains...)
				i.list.ScrollTo(0)
			}
		}
	}
}

func stringSimilarity(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	l := min(len(ra), len(rb))
	result := 0
	for i := range l {
		if ra[i] != rb[i] {
			break
		}
		result++
	}
	return result
}

func (i *Chooser[T]) changed() {
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

func (i *Chooser[T]) LayoutDropdown(gtx layout.Context) {
	if i.opened && len(i.filteredItems) > 0 {
		rowDims := measureText(gtx, i.theme, "Xy")
		showRows := min(len(i.filteredItems), i.dropdownRows)
		height := rowDims.Size.Y * showRows
		width := i.dims.Size.X
		y := i.dims.Size.Y
		if i.dropdownUp {
			y = -height
		}
		stack := op.Offset(image.Point{X: 0, Y: y}).Push(gtx.Ops)
		layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				pgtx := gtx
				pgtx.Constraints.Min.X = width
				pgtx.Constraints.Max.X = width
				pgtx.Constraints.Min.Y = height
				pgtx.Constraints.Max.Y = height
				paint.FillShape(gtx.Ops, Style.BackgroundColor, clip.Rect{Max: image.Point{X: width, Y: height}}.Op())
				i.list.Axis = layout.Vertical
				if len(i.rowClicks) != len(i.filteredItems) {
					i.rowClicks = make([]Clickable, len(i.filteredItems))
				}
				fillAdj := 0
				if len(i.filteredItems) <= i.dropdownRows {
					i.listStyle.AnchorStrategy = material.Overlay
				} else {
					i.listStyle.AnchorStrategy = material.Occupy
					fillAdj = gtx.Dp(i.listStyle.Width())
				}
				listDims := i.listStyle.Layout(pgtx, len(i.filteredItems), func(gtx layout.Context, index int) layout.Dimensions {
					lbl := i.labeller(i.filteredItems[index])
					for i.rowClicks[index].Clicked(gtx) {
						i.SetText(lbl)
						gtx.Execute(op.InvalidateCmd{})
						gtx.Execute(key.FocusCmd{Tag: &i.editor})
						i.editor.SetCaret(len(lbl), 0)
						i.opened = false
					}
					return i.rowClicks[index].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						if lbl == i.editor.Text() {
							paint.FillShape(gtx.Ops, Style.SelectedColor, clip.Rect{Max: image.Point{X: width - fillAdj, Y: rowDims.Size.Y}}.Op())
						}
						gtx.Constraints.Min.X = pgtx.Constraints.Max.X
						return layout.Inset{Left: 3, Right: 3}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							if i.middleEllipsis {
								lb := material.Label(i.theme, i.theme.TextSize, middleEllipsis(gtx, i.theme, lbl, gtx.Constraints.Max.X))
								lb.MaxLines = 1
								return lb.Layout(gtx)
							}
							lb := material.Label(i.theme, i.theme.TextSize, lbl)
							lb.MaxLines = 1
							return lb.Layout(gtx)
						})
					})
				})
				borders(pgtx, layout.Dimensions{
					Size:     image.Point{X: width, Y: height},
					Baseline: 0,
				}, Style.BorderColor, true, true, true, true)
				return listDims
			}),
		)
		stack.Pop()
	}
}

func borders(gtx layout.Context, dims layout.Dimensions, c color.NRGBA, top, left, bottom, right bool) {
	if top {
		paint.FillShape(gtx.Ops, c, clip.Rect{Max: image.Point{X: dims.Size.X, Y: 1}}.Op())
	}
	if left {
		paint.FillShape(gtx.Ops, c, clip.Rect{Max: image.Point{X: 1, Y: dims.Size.Y}}.Op())
	}
	if bottom {
		paint.FillShape(gtx.Ops, c, clip.Rect{Min: image.Point{X: 0, Y: dims.Size.Y - 1}, Max: image.Point{X: dims.Size.X, Y: dims.Size.Y}}.Op())
	}
	if right {
		paint.FillShape(gtx.Ops, c, clip.Rect{Min: image.Point{X: dims.Size.X - 1, Y: 0}, Max: image.Point{X: dims.Size.X, Y: dims.Size.Y}}.Op())
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

func (i *Chooser[T]) SetText(text string) {
	if text != i.editor.Text() {
		i.editor.SetText(text)
		i.editor.SetCaret(len(text), 0)
	}
}

func (i *Chooser[T]) SetFocused(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: &i.editor})
}

func (i *Chooser[T]) IsFocused(gtx layout.Context) bool {
	return gtx.Focused(&i.editor) || i.opened || gtx.Focused(&i.button)
}
