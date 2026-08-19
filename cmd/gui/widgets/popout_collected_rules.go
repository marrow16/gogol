package widgets

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/logic/meta"
	"maps"
	"slices"
	"strconv"
)

type collectedRulesPopout struct {
	parent           *menuPopup
	core             *Core
	rulesList        *listControl[logic.Rule]
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
	result.commonEdit = newInput(c.theme, nil, 256, func(text string) {}).maximumWidth(20).onSubmit(result.submitCommonality)
	result.rulesList = newListControl[logic.Rule](c.theme, result.rules, true).
		rowRenderer(result.layoutRule).
		onItemSelect(func(r logic.Rule) {
			c.gridHolder.grid.SetRule(r)
			c.window.Invalidate()
		}).
		onIsSelected(func(index int, r logic.Rule) bool {
			return r.Permutation() == c.gridHolder.grid.Rule.Permutation()
		})
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
				rowHt := measureText(gtx, theme, "Xy").Size.Y
				gtx.Constraints.Min.Y = rowHt * 14
				gtx.Constraints.Max.Y = rowHt * 14
				return p.rulesList.Layout(gtx)
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

func (p *collectedRulesPopout) layoutRule(gtx layout.Context, index int, r logic.Rule) layout.Dimensions {
	name := r.Rle()
	if !r.IsCustom() {
		name += ` "` + r.Name() + `"`
	}
	return layout.Inset{Left: 4, Right: 4}.Layout(gtx, label(p.core.theme, name+" ("+strconv.Itoa(r.Permutation())+")"))
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
		p.rulesList.resetItems(p.rules)
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
		p.commonEdit.isFocused(gtx) || p.rulesList.isFocused(gtx)
}

func (p *collectedRulesPopout) reset() {
	//nothing to reset
}
