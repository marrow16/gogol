package widgets

import (
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/logic/meta"
	"image"
	"maps"
	"slices"
	"strconv"
)

type collectedRulesPopout struct {
	parent           *menuPopup
	core             *Core
	list             widget.List
	listStyle        material.ListStyle
	ruleClicks       []widget.Clickable
	btnClear         *button
	btnAddCurrent    *button
	btnRemoveCurrent *button
	commonEdit       *input
	cached           map[int]bool
	rules            []logic.Rule
	commonality      string
	commonCount      int
}

func newCollectedRulesPopout(p *menuPopup, c *Core) *collectedRulesPopout {
	result := &collectedRulesPopout{
		parent:           p,
		core:             c,
		btnClear:         newButton(c.theme, "Clear"),
		btnAddCurrent:    newButton(c.theme, "Add Current"),
		btnRemoveCurrent: newButton(c.theme, "Remove Current"),
	}
	result.listStyle = material.List(c.theme, &result.list)
	result.list.Axis = layout.Vertical
	result.commonEdit = newInput(c.theme, nil, 256, func(text string) {}).maximumWidth(20).onSubmit(result.submitCommonality)
	return result
}

func (p *collectedRulesPopout) submitCommonality(text string) {
	if ev, err := meta.ParseRule(text); err == nil {
		p.core.settings.CollectedRules = make(map[int]bool)
		for perm := range ev.MatchingPermutations() {
			p.core.settings.CollectedRules[int(perm)] = true
		}
	}
}

func (p *collectedRulesPopout) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	p.checkFoundRulesChanged()
	m := measureText(gtx, theme, "M")
	ht := m.Size.Y * 16
	wd := m.Size.X * 40
	if p.btnClear.Clicked(gtx) {
		p.core.settings.CollectedRules = make(map[int]bool)
	}
	curr := p.core.gridHolder.grid.Rule.Permutation()
	if p.btnAddCurrent.Clicked(gtx) {
		p.core.settings.CollectedRules[curr] = true
	}
	if p.btnRemoveCurrent.Clicked(gtx) {
		delete(p.core.settings.CollectedRules, curr)
	}
	return layout.Inset{Left: 8, Right: 8, Top: 8, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min.X = wd
		gtx.Constraints.Min.Y = ht
		return layout.Flex{Axis: layout.Vertical, Gap: 10}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutList(gtx, theme)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Gap: 30}.Layout(gtx,
					layout.Rigid(p.btnClear.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if p.core.settings.CollectedRules[curr] {
							return p.btnRemoveCurrent.Layout(gtx)
						}
						return p.btnAddCurrent.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Gap: 10}.Layout(gtx,
					layout.Rigid(label(theme, "Commonality:")),
					layout.Flexed(1, p.commonEdit.layout),
					layout.Rigid(label(theme, "("+strconv.Itoa(len(p.rules))+"/"+strconv.Itoa(p.commonCount)+")")),
				)
			}),
		)
	})
}

func (p *collectedRulesPopout) layoutList(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	if len(p.rules) != len(p.ruleClicks) {
		p.ruleClicks = make([]widget.Clickable, len(p.rules))
	}
	m := measureText(gtx, theme, "M")
	ht := m.Size.Y * 13
	gtx.Constraints.Min.Y = ht
	gtx.Constraints.Max.Y = ht
	return widget.Border{Color: popupBorder, CornerRadius: 3, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return p.listStyle.Layout(
				gtx, len(p.rules),
				func(gtx layout.Context, index int) layout.Dimensions {
					r := p.rules[index]
					name := r.Rle()
					if !r.IsCustom() {
						name += ` "` + r.Name() + `"`
					}
					for p.ruleClicks[index].Clicked(gtx) {
						p.core.setRule(r)
					}
					if r.Permutation() == p.core.gridHolder.grid.Rule.Permutation() {
						lineHt := measureText(gtx, theme, "(Xy)").Size.Y
						paint.FillShape(
							gtx.Ops,
							popupSelectedBackground,
							clip.Rect{Max: image.Pt(gtx.Constraints.Max.X, lineHt)}.Op(),
						)
					}
					return p.ruleClicks[index].Layout(gtx, label(theme, name+" ("+strconv.Itoa(r.Permutation())+")"))
				},
			)
		})
	})
}

func (p *collectedRulesPopout) checkFoundRulesChanged() {
	changed := len(p.core.settings.CollectedRules) != len(p.cached)
	if !changed {
		for k, v := range p.core.settings.CollectedRules {
			if p.cached[k] != v {
				changed = true
				break
			}
		}
	}
	if changed {
		p.cached = maps.Clone(p.core.settings.CollectedRules)
		p.rules = make([]logic.Rule, 0, len(p.core.settings.CollectedRules))
		for perm := range p.core.settings.CollectedRules {
			if rule, err := logic.NewRuleFromPermutation(perm); err == nil {
				p.rules = append(p.rules, rule)
			}
		}
		slices.SortFunc(p.rules, func(a, b logic.Rule) int {
			if a.Permutation() < b.Permutation() {
				return -1
			}
			return 1
		})
		if ev, err := meta.CommonalityFromRules(p.rules...); err == nil {
			p.commonality = ev.String()
			p.commonEdit.setText(p.commonality)
			p.commonEdit.editor.SetCaret(len(p.commonality), 0)
			p.commonCount = meta.Count(ev)
		} else {
			p.commonality = ""
			p.commonEdit.setText("")
			p.commonCount = 0
		}
	}
}

func (p *collectedRulesPopout) hasFocus(gtx layout.Context) bool {
	return p.btnClear.isFocused(gtx) || p.btnRemoveCurrent.isFocused(gtx) || p.btnAddCurrent.isFocused(gtx) ||
		p.commonEdit.isFocused(gtx)
}

func (p *collectedRulesPopout) reset() {
	//nothing to reset
}
