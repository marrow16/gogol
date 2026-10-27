package gui

import (
	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/gui/controls"
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
	p.rleInput = controls.NewInput(theme, "sbSB/012345678", 21, p.rleChanged)
	p.permInput = controls.NewNumberInput[int](theme, 6, 0, (1<<18)-1, 1<<9, p.permChanged)
	p.intInput = controls.NewNumberInput[int](theme, 6, 0, (1<<18)-1, 1<<9, p.intChanged)
	p.nameInput = controls.NewInput(theme, "", 40, p.nameChanged).MaximumWidth(21)
	p.btnSaveName = controls.NewButton(theme, "Save")
	p.btnAddCollected = controls.NewButton(theme, "Add Collected")
	p.btnRemoveCollected = controls.NewButton(theme, "Remove Collected")
	p.refreshRules()
	return p
}

type rulesPopup struct {
	core               *Core
	parent             *statusBar
	selectedIndex      int
	sortedRules        []logic.Rule
	list               widget.List
	ruleClicks         []widget.Clickable
	rleInput           *controls.Input
	permInput          *controls.NumberInput[int]
	intInput           *controls.NumberInput[int]
	nameInput          *controls.Input
	btnSaveName        *controls.Button
	inputsDirty        bool
	btnAddCollected    *controls.Button
	btnRemoveCollected *controls.Button
}

func (p *rulesPopup) rleChanged(text string) {
	if !strings.EqualFold(text, p.core.gridHolder.grid.Rule().Rle()) {
		if r, err := logic.NewRuleRle("", text); err == nil {
			p.core.gridHolder.grid.SetRule(r)
			idx, found := slices.BinarySearchFunc(p.sortedRules, p.core.gridHolder.grid.Rule(), func(a, b logic.Rule) int {
				return strings.Compare(a.Name(), b.Name())
			})
			p.permInput.SetValue(r.Permutation())
			p.intInput.SetValue(r.Integer())
			p.nameInput.SetText(r.Name())
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
		idx, found := slices.BinarySearchFunc(p.sortedRules, p.core.gridHolder.grid.Rule(), func(a, b logic.Rule) int {
			return strings.Compare(a.Name(), b.Name())
		})
		p.rleInput.SetText(r.Rle())
		p.nameInput.SetText(r.Name())
		p.intInput.SetValue(r.Integer())
		if found {
			p.selectedIndex = idx
			p.list.ScrollTo(idx)
		} else {
			p.selectedIndex = -1
		}
	}
}

func (p *rulesPopup) intChanged(n int) {
	if r, err := logic.NewRuleFromInteger(n); err == nil {
		p.core.gridHolder.grid.SetRule(r)
		idx, found := slices.BinarySearchFunc(p.sortedRules, p.core.gridHolder.grid.Rule(), func(a, b logic.Rule) int {
			return strings.Compare(a.Name(), b.Name())
		})
		p.rleInput.SetText(r.Rle())
		p.nameInput.SetText(r.Name())
		p.permInput.SetValue(r.Permutation())
		if found {
			p.selectedIndex = idx
			p.list.ScrollTo(idx)
		} else {
			p.selectedIndex = -1
		}
	}
}

func (p *rulesPopup) nameChanged(_ string) {}

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
	idx, found := slices.BinarySearchFunc(p.sortedRules, p.core.gridHolder.grid.Rule(), func(a, b logic.Rule) int {
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
	p.inputsDirty = true
}

func (p *rulesPopup) refreshInputs() {
	p.inputsDirty = false
	rule := p.core.gridHolder.grid.Rule()
	p.rleInput.SetText(rule.Rle())
	p.permInput.SetValue(rule.Permutation())
	p.intInput.SetValue(rule.Integer())
	p.nameInput.SetText(rule.Name())
}

func (p *rulesPopup) saveRuleName() {
	if name := p.nameInput.Text(); len(name) > 0 {
		rle := p.core.gridHolder.grid.Rule().Rle()
		if r, err := logic.NewRuleRle(name, rle); err == nil {
			if ok := logic.AddRule(name, r); ok {
				p.core.settings.Rules[name] = rle
				p.refreshRules()
			}
		}
	}
}

func (p *rulesPopup) layout(gtx layout.Context) layout.Dimensions {
	if p.inputsDirty {
		p.refreshInputs()
	}
	if p.btnSaveName.Clicked(gtx) {
		p.saveRuleName()
	}
	if p.btnAddCollected.Clicked(gtx) {
		p.core.settings.CollectedRules[p.core.gridHolder.grid.Rule().Permutation()] = true
	}
	if p.btnRemoveCollected.Clicked(gtx) {
		delete(p.core.settings.CollectedRules, p.core.gridHolder.grid.Rule().Permutation())
	}
	p.handleEvents(gtx)
	rowDims := measureText(gtx, "Xy")
	macro := op.Record(gtx.Ops)
	dims := layout.Inset{Top: 1, Left: 1, Bottom: 1, Right: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(
			gtx,
			p.layoutList(rowDims),
			p.layoutDetails(),
		)
	})
	call := macro.Stop()
	paint.FillShape(gtx.Ops, backgroundColor, clip.Rect{Max: dims.Size}.Op())
	borders(gtx, dims, borderColor, true, true, false, true)
	call.Add(gtx.Ops)
	return dims
}

func (p *rulesPopup) layoutDetails() layout.FlexChild {
	custom := p.core.gridHolder.grid.Rule().IsCustom()
	canSave := false
	if custom {
		if name := p.nameInput.Text(); len(name) > 0 && name != p.core.gridHolder.grid.Rule().Name() {
			canSave = true
		}
	}
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		horizontalLine(gtx, borderColor, gtx.Constraints.Max.X, 1)
		maxText := measureMaxText(gtx, font.Bold, "Rule: ", "Perm.: ", "Name: ", "Integer: ")
		return layout.Inset{Top: 4, Bottom: 4, Left: 4, Right: 4}.Layout(gtx, flexVertical(10,
			rigid(flexHorizontal(0,
				rigidLabel("Name: ", text.End, font.Bold, maxText),
				conditionalFlexed(custom, p.nameInput.Layout, borderedInset(2, 2, 4, 4, label(p.core.gridHolder.grid.Rule().Name()))),
				conditionalRigid(custom && canSave, label(" "), nil),
				conditionalRigid(custom && canSave, p.btnSaveName.Layout, nil),
			)),
			rigid(flexHorizontal(0,
				rigidLabel("Rule: ", text.End, font.Bold, maxText),
				flexed(p.rleInput.Layout),
			)),
			rigid(flexHorizontal(0,
				rigidLabel("Perm.: ", text.End, font.Bold, maxText),
				flexed(p.permInput.Layout),
			)),
			rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
					rigid(flexHorizontal(0,
						rigidLabel("Integer: ", text.End, font.Bold, maxText),
						flexed(p.intInput.Layout),
					)),
					conditionalRigid(
						p.core.settings.CollectedRules[p.core.gridHolder.grid.Rule().Permutation()],
						p.btnRemoveCollected.Layout,
						p.btnAddCollected.Layout,
					),
				)
			}),
		))
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
		idx, ok := slices.BinarySearchFunc(p.sortedRules, p.core.gridHolder.grid.Rule(), func(a, b logic.Rule) int {
			return strings.Compare(a.Name(), b.Name())
		})
		if !ok {
			idx = -1
		}
		return material.List(theme, &p.list).Layout(
			pgtx,
			len(p.sortedRules),
			func(gtx layout.Context, index int) layout.Dimensions {
				for p.ruleClicks[index].Clicked(gtx) {
					gtx.Execute(key.FocusCmd{Tag: &p.list})
					p.core.stop()
					p.core.gridHolder.grid.SetRule(p.sortedRules[index])
					p.refreshInputs()
					window.Invalidate()
				}
				return p.ruleClicks[index].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					if index == idx {
						bg := selectedColor
						if !p.rleInput.IsFocused(gtx) && !p.permInput.IsFocused(gtx) && !p.intInput.IsFocused(gtx) && !p.nameInput.IsFocused(gtx) {
							bg = selectedFocusedColor
						}
						paint.FillShape(gtx.Ops, bg, clip.Rect{Max: image.Point{X: pgtx.Constraints.Max.X, Y: rowDims.Size.Y}}.Op())
					}
					gtx.Constraints.Min.X = pgtx.Constraints.Max.X
					r := p.sortedRules[index]
					return inset(0, 0, 3, 3,
						flexHorizontal(32,
							flexed(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(theme, theme.TextSize, r.Name())
								lbl.Alignment = text.Start
								lbl.MaxLines = 1
								return lbl.Layout(gtx)
							}, 1.5),
							flexed(func(gtx layout.Context) layout.Dimensions {
								s := strconv.Itoa(r.Permutation())
								if len(s) < 6 {
									s = strings.Repeat("\u2007", 6-len(s)) + s
								}
								lbl := material.Label(theme, theme.TextSize, r.Rle()+"  "+s)
								lbl.Alignment = text.End
								lbl.MaxLines = 1
								return lbl.Layout(gtx)
							}, 2.5),
						),
					)(gtx)
				})
			},
		)
	})
}

func (p *rulesPopup) handleEvents(gtx layout.Context) {
	switch {
	case p.rleInput.IsFocused(gtx):
		p.rleInput.Update(gtx)
		return
	case p.permInput.IsFocused(gtx):
		p.permInput.Update(gtx)
		return
	case p.intInput.IsFocused(gtx):
		p.intInput.Update(gtx)
		return
	case p.nameInput.IsFocused(gtx):
		p.nameInput.Update(gtx)
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
				p.permInput.SetFocused(gtx)
			} else {
				p.rleInput.SetFocused(gtx)
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
			p.refreshInputs()
			window.Invalidate()
		}
	}
}
