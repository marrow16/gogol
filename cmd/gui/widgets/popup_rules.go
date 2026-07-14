package widgets

import (
	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/logic"
	"image"
	"slices"
	"strconv"
	"strings"
)

func newRulesPopup(parent *statusBar) *rulesPopup {
	p := &rulesPopup{
		core:   parent.core,
		parent: parent,
	}
	p.rleInput = newInput(parent.core.theme, "sbSB/012345678", 21, p.rleChanged)
	p.permInput = newNumberInput[int](p.core.theme, 6, 0, (1<<18)-1, 1<<9, p.permChanged)
	p.refreshRules()
	return p
}

type rulesPopup struct {
	core          *Core
	parent        *statusBar
	selectedIndex int
	sortedRules   []logic.Rule
	list          widget.List
	ruleClicks    []widget.Clickable
	rleInput      *input
	permInput     *numberInput[int]
}

func (p *rulesPopup) rleChanged(text string) {
	if !strings.EqualFold(text, p.core.gridHolder.grid.Rule.Rle()) {
		if r, err := logic.NewRuleRle("", text); err == nil {
			p.core.gridHolder.grid.SetRule(r)
			idx, found := slices.BinarySearchFunc(p.sortedRules, p.core.gridHolder.grid.Rule, func(a, b logic.Rule) int {
				return strings.Compare(a.Name(), b.Name())
			})
			p.permInput.setValue(r.Permutation())
			if found {
				p.selectedIndex = idx
				p.list.ScrollTo(idx)
			} else {
				p.selectedIndex = -1
			}
		}
	}
}

func (p *rulesPopup) permChanged(n int) {
	if r, err := logic.NewRuleFromPermutation(n); err == nil {
		p.core.gridHolder.grid.SetRule(r)
		idx, found := slices.BinarySearchFunc(p.sortedRules, p.core.gridHolder.grid.Rule, func(a, b logic.Rule) int {
			return strings.Compare(a.Name(), b.Name())
		})
		p.rleInput.setText(r.Rle())
		if found {
			p.selectedIndex = idx
			p.list.ScrollTo(idx)
		} else {
			p.selectedIndex = -1
		}
	}
}

func (p *rulesPopup) refreshRules() {
	p.sortedRules = make([]logic.Rule, 0, len(logic.Rules))
	for _, r := range logic.Rules {
		p.sortedRules = append(p.sortedRules, r)
	}
	slices.SortFunc(p.sortedRules, func(a, b logic.Rule) int {
		return strings.Compare(a.Name(), b.Name())
	})
	p.setSelected()
}

func (p *rulesPopup) setSelected() {
	idx, found := slices.BinarySearchFunc(p.sortedRules, p.core.gridHolder.grid.Rule, func(a, b logic.Rule) int {
		return strings.Compare(a.Name(), b.Name())
	})
	if found {
		p.selectedIndex = idx
		p.list.ScrollTo(idx)
	} else {
		p.selectedIndex = -1
	}
	p.updateInputs()
}

func (p *rulesPopup) updateInputs() {
	p.rleInput.setText(p.core.gridHolder.grid.Rule.Rle())
	p.permInput.setValue(p.core.gridHolder.grid.Rule.Permutation())
}

func (p *rulesPopup) layout(gtx layout.Context) layout.Dimensions {
	p.handleEvents(gtx)

	rowDims := measureText(gtx, p.core.theme, "Xy")
	macro := op.Record(gtx.Ops)
	dims := layout.Inset{Top: 1, Left: 1, Bottom: 1, Right: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(
			gtx,
			p.layoutList(rowDims),
			p.layoutDetails(rowDims),
		)
	})
	call := macro.Stop()
	paint.FillShape(
		gtx.Ops,
		popupBackground,
		clip.Rect{Max: dims.Size}.Op(),
	)
	border(gtx, dims, true, true, false, true)
	call.Add(gtx.Ops)
	return dims
}

func (p *rulesPopup) layoutDetails(rowDims layout.Dimensions) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		paint.FillShape(gtx.Ops, popupBorder, clip.Rect(image.Rect(0, 0, gtx.Constraints.Max.X, 1)).Op())
		maxText := measureMaxText(gtx, p.core.theme, font.Bold, "Rule: ", "Perm.: ").Size.X
		pad := unit.Dp(float32(rowDims.Size.Y/3) / gtx.Metric.PxPerDp)
		return layout.Inset{
			Top:    pad,
			Bottom: pad,
			Left:   pad,
			Right:  pad,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
				Gap:  10,
			}.Layout(gtx,
				row(p.core.theme, maxText, "Rule: ", p.rleInput.layout),
				row(p.core.theme, maxText, "Perm.: ", p.permInput.layout),
			)
		})
	})
}

func (p *rulesPopup) layoutList(rowDims layout.Dimensions) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		pgtx := gtx
		pgtx.Constraints.Max.Y = rowDims.Size.Y * 10
		pgtx.Constraints.Min.X = gtx.Constraints.Max.X
		p.list.Axis = layout.Vertical
		if len(p.ruleClicks) != len(p.sortedRules) {
			p.ruleClicks = make([]widget.Clickable, len(p.sortedRules))
		}
		idx, ok := slices.BinarySearchFunc(p.sortedRules, p.core.gridHolder.grid.Rule, func(a, b logic.Rule) int {
			return strings.Compare(a.Name(), b.Name())
		})
		if !ok {
			idx = -1
		}
		return material.List(p.core.theme, &p.list).Layout(
			pgtx,
			len(p.sortedRules),
			func(gtx layout.Context, index int) layout.Dimensions {
				for p.ruleClicks[index].Clicked(gtx) {
					gtx.Execute(key.FocusCmd{Tag: &p.list})
					p.core.stop()
					p.core.gridHolder.grid.SetRule(p.sortedRules[index])
					p.updateInputs()
					p.core.window.Invalidate()
				}
				return p.ruleClicks[index].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					if index == idx {
						bg := popupSelectedBackground
						if !p.rleInput.isFocused(gtx) && !p.permInput.isFocused(gtx) {
							bg = popupSelectedFocusedBackground
						}
						paint.FillShape(
							gtx.Ops,
							bg,
							clip.Rect{Max: image.Pt(pgtx.Constraints.Max.X, rowDims.Size.Y)}.Op(),
						)
					}
					gtx.Constraints.Min.X = pgtx.Constraints.Max.X
					r := p.sortedRules[index]
					return layout.Flex{Axis: layout.Horizontal, Gap: 32}.Layout(gtx,
						layout.Flexed(1.5, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(p.core.theme, p.core.theme.TextSize, r.Name())
							lbl.Alignment = text.Start
							lbl.MaxLines = 1
							return lbl.Layout(gtx)
						}),
						layout.Flexed(2.5, func(gtx layout.Context) layout.Dimensions {
							s := strconv.Itoa(r.Permutation())
							if len(s) < 6 {
								s = strings.Repeat("\u2007", 6-len(s)) + s
							}
							lbl := material.Label(p.core.theme, p.core.theme.TextSize, r.Rle()+"  "+s)
							lbl.Alignment = text.End
							lbl.MaxLines = 1
							return lbl.Layout(gtx)
						}),
					)
				})
			},
		)
	})
}

func (p *rulesPopup) handleEvents(gtx layout.Context) {
	switch {
	case p.rleInput.isFocused(gtx):
		p.rleInput.update(gtx)
		return
	case p.permInput.isFocused(gtx):
		p.permInput.update(gtx)
		return
	}
	for {
		ev, ok := gtx.Event(key.Filter{
			Optional: key.ModShift | key.ModCtrl | key.ModAlt | key.ModCommand,
			Name:     "",
		}, key.Filter{
			Optional: key.ModShift,
			Name:     key.NameTab,
		})
		if !ok {
			break
		}
		kev, ok := ev.(key.Event)
		if !ok || kev.State != key.Press {
			continue
		}
		changed := false
		switch kev.Name {
		case key.NameTab:
			if kev.Modifiers == key.ModShift {
				p.permInput.setFocused(gtx)
			} else {
				p.rleInput.setFocused(gtx)
			}
		case key.NameEnter, key.NameReturn:
			p.parent.showHidePopup(popupNone)
		case key.NameUpArrow, key.NameLeftArrow:
			if p.selectedIndex > 0 {
				changed = true
				p.selectedIndex--
			}
		case key.NameDownArrow, key.NameRightArrow:
			if p.selectedIndex < len(p.sortedRules)-1 {
				changed = true
				p.selectedIndex++
			}
		case key.NameHome:
			changed = true
			p.selectedIndex = 0
		case key.NameEnd:
			changed = true
			p.selectedIndex = len(p.sortedRules) - 1
		case key.NamePageUp:
			changed = true
			p.selectedIndex -= 9
		case key.NamePageDown:
			changed = true
			p.selectedIndex += 9
		default:
			ks := strings.ToUpper(string(kev.Name))
			if len(ks) == 1 && ((ks >= "A" && ks <= "Z") || (ks >= "0" && ks <= "9")) {
				if p.selectedIndex > 0 {
					curr := p.sortedRules[p.selectedIndex]
					if strings.HasPrefix(strings.ToUpper(curr.Name()), ks) && p.selectedIndex+1 < len(p.sortedRules)-1 && strings.HasPrefix(strings.ToUpper(p.sortedRules[p.selectedIndex+1].Name()), ks) {
						p.selectedIndex++
						changed = true
					}
				}
				if !changed {
					if idx := slices.IndexFunc(p.sortedRules, func(r logic.Rule) bool {
						return strings.HasPrefix(strings.ToUpper(r.Name()), ks)
					}); idx != -1 {
						changed = true
						p.selectedIndex = idx
					}
				}
			}
		}
		if changed {
			if p.selectedIndex < 0 {
				p.selectedIndex = 0
			}
			if p.selectedIndex >= len(p.sortedRules) {
				p.selectedIndex = len(p.sortedRules) - 1
			}
			// keep selected visible in 10-row list
			if p.selectedIndex < p.list.Position.First {
				p.list.Position.First = p.selectedIndex
			} else if p.selectedIndex >= p.list.Position.First+10 {
				p.list.Position.First = p.selectedIndex - 9
			}
			p.core.gridHolder.grid.SetRule(p.sortedRules[p.selectedIndex])
			p.updateInputs()
			p.core.window.Invalidate()
		}
	}
}
