package gui

import (
	"gioui.org/layout"
	"github.com/marrow16/gogol/gui/controls"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/logic/meta"
	"maps"
	"slices"
	"strconv"
)

type collectedRulesPopout struct {
	parent           *menuPopup
	core             *Core
	rulesList        *controls.List[logic.Rule]
	btnClear         *controls.Button
	btnAddCurrent    *controls.Button
	btnRemoveCurrent *controls.Button
	commonEdit       *controls.Input
	cached           map[int]bool
	rules            []logic.Rule
	commonality      string
	commonCount      int
}

func newCollectedRulesPopout(p *menuPopup, c *Core) *collectedRulesPopout {
	result := &collectedRulesPopout{
		parent:           p,
		core:             c,
		btnClear:         controls.NewButton(theme, "Clear"),
		btnAddCurrent:    controls.NewButton(theme, "Add Current"),
		btnRemoveCurrent: controls.NewButton(theme, "Remove Current"),
	}
	result.commonEdit = controls.NewInput(theme, nil, 256, func(text string) {}).MaximumWidth(20).OnSubmit(result.submitCommonality)
	result.rulesList = controls.NewList[logic.Rule](theme, result.rules, true).
		RowRenderer(result.layoutRule).
		OnItemSelect(func(r logic.Rule, _ bool) {
			c.gridHolder.grid.SetRule(r)
			window.Invalidate()
		}).
		OnIsSelected(func(index int, r logic.Rule) bool {
			return r.Permutation() == c.gridHolder.grid.Rule().Permutation()
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

func (p *collectedRulesPopout) layout(gtx layout.Context) layout.Dimensions {
	p.checkFoundRulesChanged()
	m := measureText(gtx, "M")
	ht := m.Size.Y * 16
	wd := m.Size.X * 40
	if p.btnClear.Clicked(gtx) {
		p.core.settings.CollectedRules = make(map[int]bool)
	}
	curr := p.core.gridHolder.grid.Rule().Permutation()
	if p.btnAddCurrent.Clicked(gtx) {
		p.core.settings.CollectedRules[curr] = true
	}
	if p.btnRemoveCurrent.Clicked(gtx) {
		delete(p.core.settings.CollectedRules, curr)
	}
	gtx.Constraints.Min.X = wd
	gtx.Constraints.Min.Y = ht
	return popoutLayout(gtx, flexVertical(10,
		rigid(func(gtx layout.Context) layout.Dimensions {
			rowHt := measureText(gtx, "Xy").Size.Y
			gtx.Constraints.Min.Y = rowHt * 14
			gtx.Constraints.Max.Y = gtx.Constraints.Min.Y
			return p.rulesList.Layout(gtx)
		}),
		rigid(flexHorizontal(30,
			rigid(p.btnClear.Layout),
			conditionalRigid(p.core.settings.CollectedRules[curr], p.btnRemoveCurrent.Layout, p.btnAddCurrent.Layout),
		)),
		rigid(flexHorizontal(10,
			rigidLabel("Commonality:", 0, 0, 0),
			flexed(p.commonEdit.Layout),
			rigidLabel("("+strconv.Itoa(len(p.rules))+"/"+strconv.Itoa(p.commonCount)+")", 0, 0, 0),
		)),
	))
}

func (p *collectedRulesPopout) layoutRule(gtx layout.Context, _ int, r logic.Rule) layout.Dimensions {
	name := r.Rle()
	if !r.IsCustom() {
		name += ` "` + r.Name() + `"`
	}
	return inset(0, 0, 4, 4, label(name+" ("+strconv.Itoa(r.Permutation())+")"))(gtx)
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
		p.rulesList.ResetItems(p.rules)
		if ev, err := meta.CommonalityFromRules(p.rules...); err == nil {
			p.commonality = ev.String()
			p.commonEdit.SetText(p.commonality)
			p.commonEdit.SetCaret(len(p.commonality), 0)
			p.commonCount = meta.Count(ev)
		} else {
			p.commonality = ""
			p.commonEdit.SetText("")
			p.commonCount = 0
		}
	}
}

func (p *collectedRulesPopout) hasFocus(gtx layout.Context) bool {
	return p.btnClear.IsFocused(gtx) || p.btnRemoveCurrent.IsFocused(gtx) || p.btnAddCurrent.IsFocused(gtx) ||
		p.commonEdit.IsFocused(gtx) || p.rulesList.IsFocused(gtx)
}

func (p *collectedRulesPopout) reset() {
	//nothing to reset
}
