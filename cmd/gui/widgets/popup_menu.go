package widgets

import (
	"gioui.org/font"
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

func newMenuPopup(parent *statusBar) *menuPopup {
	result := &menuPopup{
		core:   parent.core,
		parent: parent,
		tag:    &struct{}{},
	}
	result.menuItems = menuItems{
		{
			parent: result,
			label:  "Patterns",
			popout: popoutPatterns,
		},
		{
			parent: result,
			label:  "Captured Patterns",
			popout: popoutCapturedPatterns,
		},
		{
			parent: result,
			label:  "Load Patterns",
			popout: popoutLoadPatterns,
		},
		{},
		{
			parent: result,
			label:  "Instrumentation",
			popout: popoutInstrumentation,
		},
		{
			parent: result,
			label:  "Grid Recipes",
			popout: popoutGridRecipes,
		},
		{},
		{
			parent: result,
			label:  "Size/Wrapping/Boundaries",
			popout: popoutSizeWrappingBoundaries,
		},
		{
			parent: result,
			label:  "Stepping",
			popout: popoutStepping,
		},
		{
			parent: result,
			label:  "Colors",
			popout: popoutColors,
		},
		{},
		{
			parent: result,
			label:  "Snapshot",
			key:    "S",
			fn: func() {
				parent.core.snapshot()
			},
		},
		{
			parent: result,
			label:  "Undo to snapshot",
			key:    "Z",
			fn: func() {
				parent.core.undoToSnapshot()
			},
		},
		{
			parent: result,
			label:  "Export Grid",
			key:    "X",
			fn: func() {
				parent.core.export()
			},
		},
		{
			parent: result,
			label:  "Import Grid",
			popout: popoutImportGrid,
		},
		{},
		{
			parent: result,
			label:  "Edit mode",
			key:    "E",
			fn: func() {
				parent.core.startEditMode()
			},
		},
		{},
		{
			parent: result,
			label:  "Randomize",
			key:    "R",
			fn: func() {
				parent.core.randomize()
			},
		},
		{
			parent: result,
			label:  "Random Noise",
			key:    "N",
			fn: func() {
				parent.core.randomChanges()
			},
		},
		{
			parent: result,
			label:  "Clear",
			key:    "C",
			fn: func() {
				parent.core.clear()
			},
		},
	}
	result.patternsPopout = newPatternsPopout(result, result.core)
	result.capturedPatternsPopout = newCapturedPatternsPopout(result, result.core)
	result.popouts = map[popoutType]popout{
		popoutColors:                 newColorsPopout(result, result.core),
		popoutSizeWrappingBoundaries: newSizingPopout(result, result.core),
		popoutStepping:               newSteppingPopout(result, result.core),
		popoutCapturedPatterns:       result.capturedPatternsPopout,
		popoutPatterns:               result.patternsPopout,
		popoutLoadPatterns:           newLoadPatternsPopout(result, result.core),
		popoutImportGrid:             newImportGridPopout(result, result.core),
		popoutGridRecipes:            newGridRecipesPopout(result, result.core),
		popoutInstrumentation:        newInstrumentationPopout(result, result.core),
	}
	result.selected = len(result.menuItems) - 1
	return result
}

type menuPopup struct {
	core                   *Core
	parent                 *statusBar
	tag                    *struct{}
	focused                bool
	menuItems              menuItems
	popouts                map[popoutType]popout
	patternsPopout         *patternsPopout
	capturedPatternsPopout *capturedPatternsPopout
	selected               int
	poppedOut              bool
	itemWidth              int
	width                  int
	right                  int
	bottom                 int
}

func (p *menuPopup) setSelected(n int) {
	if n != p.selected {
		p.resetPopout(n)
	}
	p.selected = n
}

func (p *menuPopup) resetPopout(n int) {
	if pout, ok := p.popouts[p.menuItems[n].popout]; ok {
		pout.reset()
	}
}

func (p *menuPopup) layout(gtx layout.Context) layout.Dimensions {
	p.right = gtx.Constraints.Max.X
	p.bottom = gtx.Constraints.Max.Y
	p.handleKeys(gtx)
	if p.itemWidth == 0 {
		p.itemWidth = p.menuItems.measureWidth(gtx, p.core.theme)
	}
	macro := op.Record(gtx.Ops)
	dims := layout.Inset{Top: 2, Left: 2, Bottom: 2, Right: 2}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx, p.menuItems.layout(p.core.theme, p.itemWidth)...)
	})
	p.width = dims.Size.X
	call := macro.Stop()
	paint.FillShape(gtx.Ops, popupBackground, clip.Rect{Max: dims.Size}.Op())
	border(gtx, dims, true, true, false, false)
	call.Add(gtx.Ops)
	p.layoutPopouts(gtx, p.core.theme)
	return dims
}

func (p *menuPopup) layoutPopouts(gtx layout.Context, theme *material.Theme) {
	if !p.poppedOut {
		return
	}
	item := p.menuItems[p.selected]
	if item.popout == popoutNone {
		return
	}
	pop := p.popouts[item.popout]
	if pop == nil {
		return
	}
	macro := op.Record(gtx.Ops)
	dims := pop.layout(gtx, theme)
	call := macro.Stop()
	x := -dims.Size.X
	y := p.selectedItemY()
	stack := op.Offset(image.Pt(x, y)).Push(gtx.Ops)
	defer stack.Pop()
	paint.FillShape(
		gtx.Ops,
		popupBackground,
		clip.Rect{Max: dims.Size}.Op(),
	)
	border(gtx, dims, true, true, true, true)
	call.Add(gtx.Ops)
}

func (p *menuPopup) selectedItemY() int {
	y := 0
	for i := 0; i < p.selected && i < len(p.menuItems); i++ {
		y += p.menuItems[i].height
	}
	return y
}

func (p *menuPopup) handleKeys(gtx layout.Context) {
	if !p.focused {
		p.focused = true
		gtx.Execute(key.FocusCmd{Tag: p.tag})
	}
	if p.poppedOut {
		if pop, ok := p.popouts[p.menuItems[p.selected].popout]; ok && pop != nil && pop.hasFocus(gtx) {
			return
		}
	}
	for {
		ev, ok := gtx.Event(
			key.Filter{Name: key.NameUpArrow},
			key.Filter{Name: key.NameDownArrow},
			key.Filter{Name: key.NameLeftArrow},
			key.Filter{Name: key.NameRightArrow},
			key.Filter{Name: key.NameHome},
			key.Filter{Name: key.NameEnd},
			key.Filter{Name: key.NameReturn},
			key.Filter{Name: key.NameEnter},
			key.Filter{Name: key.NameSpace},
		)
		if !ok {
			break
		}
		kev, ok := ev.(key.Event)
		if !ok || kev.State != key.Press {
			continue
		}
		if kev.Name == key.NameTab {
			if kev.Modifiers == key.ModShift {
				kev.Name = key.NameUpArrow
			} else {
				kev.Name = key.NameDownArrow
			}
		}
		switch kev.Name {
		case key.NameUpArrow:
			if p.selected > 0 {
				n := p.selected - 1
				if len(p.menuItems[n].label) == 0 {
					n--
				}
				p.setSelected(n)
			} else {
				p.setSelected(len(p.menuItems) - 1)
			}
		case key.NameDownArrow:
			if p.selected < len(p.menuItems)-1 {
				n := p.selected + 1
				if len(p.menuItems[n].label) == 0 {
					n++
				}
				p.setSelected(n)
			} else {
				p.setSelected(0)
			}
		case key.NameHome:
			p.setSelected(0)
		case key.NameEnd:
			p.setSelected(len(p.menuItems) - 1)
		case key.NameReturn, key.NameEnter, key.NameSpace:
			if p.menuItems[p.selected].fn != nil {
				p.menuItems[p.selected].fn()
			} else {
				p.poppedOut = !p.poppedOut
				if p.poppedOut {
					p.resetPopout(p.selected)
				}
			}
		case key.NameLeftArrow:
			if p.menuItems[p.selected].popout != popoutNone {
				p.poppedOut = true
				p.resetPopout(p.selected)
			}
		case key.NameRightArrow:
			if p.menuItems[p.selected].popout != popoutNone {
				p.poppedOut = false
			}
		}
	}
}

type menuItems []menuItem

func (m menuItems) layout(theme *material.Theme, width int) []layout.FlexChild {
	result := make([]layout.FlexChild, 0, len(m))
	for idx := range m {
		item := &m[idx]
		if len(item.label) > 0 {
			item.index = idx
		}
		result = append(result, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return item.layout(gtx, theme, width)
		}))
	}
	return result
}

func (m menuItems) measureWidth(gtx layout.Context, theme *material.Theme) int {
	lbls := make([]string, 0, len(m))
	addSpaceFor := " " + modKeyName + "WWWW"
	for _, item := range m {
		if len(item.key) > 0 {
			lbls = append(lbls, item.label+modKeyName+item.key+addSpaceFor)
		} else {
			lbls = append(lbls, "<< "+item.label)
		}
	}
	dims := measureMaxText(gtx, theme, font.Normal, lbls...)
	return dims.Size.X
}

type menuItem struct {
	parent    *menuPopup
	clickable widget.Clickable
	label     string
	key       string
	popout    popoutType
	fn        func()
	index     int
	height    int
}

func (i *menuItem) layout(gtx layout.Context, theme *material.Theme, width int) layout.Dimensions {
	if i.label == "" {
		dims := measureText(gtx, theme, "-")
		dims.Size.X = width
		dims.Size.Y = (dims.Size.Y * 2) / 3
		i.height = dims.Size.Y
		paint.FillShape(gtx.Ops, popupBorder,
			clip.Rect(image.Rect(
				0,
				dims.Size.Y/2,
				dims.Size.X,
				(dims.Size.Y/2)+1,
			)).Op(),
		)
		return dims
	} else {
		for i.clickable.Clicked(gtx) {
			if i.popout != popoutNone {
				i.parent.poppedOut = true
			}
			i.parent.setSelected(i.index)
			if i.fn != nil {
				i.fn()
			}
		}
		return material.Clickable(gtx, &i.clickable, func(gtx layout.Context) layout.Dimensions {
			extra := unit.Dp(0)
			if len(i.key) == 0 {
				sz1 := measureText(gtx, theme, "C")
				sz2 := measureText(gtx, theme, modKeyName)
				extra = gtx.Metric.PxToDp(max(sz1.Size.Y, sz2.Size.Y) - min(sz1.Size.Y, sz2.Size.Y))
			}
			gtx.Constraints.Min.X = width
			gtx.Constraints.Max.X = width
			macro := op.Record(gtx.Ops)
			dims := layout.Inset{
				Top:    unit.Dp(2),
				Left:   unit.Dp(4),
				Right:  unit.Dp(4),
				Bottom: unit.Dp(2 + extra),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if i.popout == popoutNone {
							return layout.Dimensions{}
						}
						return material.Label(theme, theme.TextSize, "< ").Layout(gtx)
					}),
					layout.Flexed(1, label(theme, i.label)),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if i.key == "" {
							return layout.Dimensions{}
						}
						return material.Label(theme, theme.TextSize, modKeyName+i.key).Layout(gtx)
					}),
				)
			})
			call := macro.Stop()
			if i.index == i.parent.selected {
				paint.FillShape(
					gtx.Ops,
					popupSelectedBackground,
					clip.Rect{Max: dims.Size}.Op(),
				)
			}
			call.Add(gtx.Ops)
			i.height = dims.Size.Y
			return dims
		})
	}
}

type popoutType int

const (
	popoutNone popoutType = iota
	popoutColors
	popoutSizeWrappingBoundaries
	popoutStepping
	popoutCapturedPatterns
	popoutPatterns
	popoutLoadPatterns
	popoutImportGrid
	popoutGridRecipes
	popoutInstrumentation
)
