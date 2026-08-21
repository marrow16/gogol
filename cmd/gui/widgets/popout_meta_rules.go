package widgets

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/logic/meta"
	"slices"
	"strconv"
	"strings"
)

type metaRulesPopout struct {
	parent       *menuPopup
	core         *Core
	chooser      *chooser[string]
	btnCreate    *button
	btnDelete    *button
	editor       widget.Editor
	mode         *widget.Enum
	radioEdit    *radioButton
	radioMatches *radioButton
	parseError   error
	linkHelp     widget.Clickable
	listMatched  *listControl[logic.Rule]
	btnAddAll    *button
	current      string
	currentRule  meta.Evaluator
	matched      []logic.Rule
}

const (
	metaRuleEdit    = "edit"
	metaRuleMatches = "matches"
)

func newMetaRulesPopout(p *menuPopup, c *Core) *metaRulesPopout {
	result := &metaRulesPopout{
		parent:    p,
		core:      c,
		btnCreate: newButton("Create"),
		btnDelete: newButton("Delete"),
		btnAddAll: newButton("Add all to collected rules"),
	}
	result.chooser = newChooser[string](38,
		result.sortedMetaRules(),
		result.metaRuleSelected,
		func(name string) string {
			return name
		},
	)
	result.mode = &widget.Enum{Value: metaRuleEdit}
	result.radioEdit = newRadioButton(result.mode, metaRuleEdit, "Edit")
	result.radioMatches = newRadioButton(result.mode, metaRuleMatches, "Matching Rules")
	result.listMatched = newListControl[logic.Rule]([]logic.Rule{}, true).
		rowRenderer(result.layoutMatchedRule).
		onItemSelect(func(r logic.Rule) {
			c.gridHolder.grid.SetRule(r)
			window.Invalidate()
		}).
		onIsSelected(func(index int, r logic.Rule) bool {
			return r.Permutation() == c.gridHolder.grid.Rule.Permutation()
		})
	return result
}

func (p *metaRulesPopout) layout(gtx layout.Context) layout.Dimensions {
	selected := p.chooser.currentItem()
	currName := p.chooser.editor.Text()
	if p.btnCreate.Clicked(gtx) {
		if selected == nil && currName != "" {
			if _, exists := p.core.settings.MetaRules[currName]; !exists {
				p.core.settings.MetaRules[currName] = "B() / S()"
				p.chooser.resetItems(p.sortedMetaRules())
				p.editor.SetText("B() / S()")
			}
		}
	}
	mt := measureText(gtx, "Xy")
	ht := mt.Size.Y * 15
	editorHt := ht - mt.Size.Y
	return popoutLayout(gtx, func(gtx layout.Context) layout.Dimensions {
		dims := flexVertical(10,
			rigid(p.chooser.layout),
			rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Max.X = p.chooser.dims.Size.X
				if selected != nil {
					gtx.Constraints.Min.Y = ht
					return p.layoutMetaRule(gtx, editorHt)
				} else if currName == "" {
					return label("No meta rule selected (select or enter new name)")(gtx)
				} else {
					return flexHorizontal(0, rigid(p.btnCreate.Layout))(gtx)
				}
			}),
		)(gtx)
		p.chooser.layoutDropdown(gtx)
		return dims
	})
}

func (p *metaRulesPopout) layoutMetaRule(gtx layout.Context, editorHt int) layout.Dimensions {
	if p.btnDelete.Clicked(gtx) {
		name := p.chooser.editor.Text()
		delete(p.core.settings.MetaRules, name)
		p.chooser.resetItems(p.sortedMetaRules())
		return layout.Dimensions{}
	}
	p.mode.Update(gtx)
	p.updateEditor(gtx)
	return flexVertical(10,
		rigid(flexHorizontal(20,
			rigid(p.btnDelete.Layout),
			rigid(p.radioEdit.Layout),
			rigid(p.radioMatches.Layout),
		)),
		rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.Y = editorHt
			gtx.Constraints.Min.Y = editorHt
			gtx.Constraints.Min.X = p.chooser.dims.Size.X
			switch p.mode.Value {
			case metaRuleMatches:
				return p.layoutMatchedRules(gtx, editorHt)
			default:
				return p.layoutEditor(gtx, editorHt)
			}
		}),
		rigid(func(gtx layout.Context) layout.Dimensions {
			switch {
			case p.parseError != nil:
				return errorLabel(p.parseError)(gtx)
			case p.mode.Value == metaRuleMatches && p.currentRule != nil:
				return p.layoutNavButtons(gtx)
			case p.currentRule != nil:
				return p.layoutReport(gtx)
			}
			return layout.Dimensions{}
		}),
	)(gtx)
}

func (p *metaRulesPopout) layoutReport(gtx layout.Context) layout.Dimensions {
	if p.linkHelp.Clicked(gtx) {
		_ = openURL(metaRuleHelp)
	}
	return flexHorizontal(30,
		rigid(linkLabel(&p.linkHelp, "(see help)")),
		rigidLabel("Matched rules: "+strconv.Itoa(len(p.matched)), 0, 0, 0),
	)(gtx)
}

func (p *metaRulesPopout) layoutNavButtons(gtx layout.Context) layout.Dimensions {
	if len(p.matched) > 0 {
		if p.btnAddAll.Clicked(gtx) {
			for perm := range p.currentRule.MatchingPermutations() {
				p.core.settings.CollectedRules[int(perm)] = true
			}
		}
		return flexHorizontal(0, rigid(p.btnAddAll.Layout))(gtx)
	} else {
		return label("No matched rules found")(gtx)
	}
}

func (p *metaRulesPopout) layoutEditor(gtx layout.Context, editorHt int) layout.Dimensions {
	gtx.Constraints.Max.Y = editorHt
	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	gtx.Constraints.Min.X = p.chooser.dims.Size.X
	style := material.Editor(theme, &p.editor, "")
	bc, bt := focusedBorder(gtx.Focused(&p.editor))
	return widget.Border{Color: bc, CornerRadius: 3, Width: bt}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, style.Layout)
	})
}

func (p *metaRulesPopout) layoutMatchedRules(gtx layout.Context, editorHt int) layout.Dimensions {
	gtx.Constraints.Max.Y = editorHt
	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	gtx.Constraints.Min.X = p.chooser.dims.Size.X
	p.parseCurrent()
	if p.parseError != nil || p.currentRule == nil {
		return widget.Border{Color: popupBorder, CornerRadius: 3, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return label("Error - No matching rules")(gtx)
			})
		})
	}
	count := len(p.matched)
	if count == 0 {
		return widget.Border{Color: popupBorder, CornerRadius: 3, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return label("No matching rules")(gtx)
			})
		})
	}
	return p.listMatched.Layout(gtx)
}

func (p *metaRulesPopout) layoutMatchedRule(gtx layout.Context, index int, r logic.Rule) layout.Dimensions {
	name := r.Rle()
	if !r.IsCustom() {
		name += ` "` + r.Name() + `"`
	}
	return inset(0, 0, 4, 4, flexHorizontal(32,
		flexed(label(name+" ("+strconv.Itoa(r.Permutation())+")"), 3.5),
		flexed(labelRight(strconv.Itoa(index+1))),
	))(gtx)
}

func (p *metaRulesPopout) parseCurrent() {
	mrs := p.editor.Text()
	if mrs != p.current || p.currentRule == nil {
		p.current = mrs
		if r, err := meta.ParseRule(mrs); err == nil {
			p.matched = make([]logic.Rule, 0)
			p.currentRule = r
			for r := range p.currentRule.MatchingRules() {
				p.matched = append(p.matched, r)
			}
			p.parseError = nil
		} else {
			p.matched = nil
			p.currentRule = nil
			p.parseError = err
		}
		p.listMatched.resetItems(p.matched)
	}
}

func (p *metaRulesPopout) updateEditor(gtx layout.Context) {
	for {
		ev, ok := p.editor.Update(gtx)
		if !ok {
			break
		}
		if _, ok = ev.(widget.ChangeEvent); ok {
			p.parseCurrent()
			if p.parseError == nil {
				name := p.chooser.editor.Text()
				if _, ok = p.core.settings.MetaRules[name]; ok {
					p.core.settings.MetaRules[name] = strings.ReplaceAll(p.current, "    ", "\t")
				}
			}
		}
	}
}

func (p *metaRulesPopout) hasFocus(gtx layout.Context) bool {
	on, radiosFocused := p.mode.Focused()
	return radiosFocused || p.chooser.isFocused(gtx) ||
		p.btnCreate.isFocused(gtx) || p.btnDelete.isFocused(gtx) ||
		gtx.Focused(&p.editor) || p.btnAddAll.isFocused(gtx) ||
		(on == metaRuleMatches && p.listMatched.isFocused(gtx))
}

func (p *metaRulesPopout) reset() {
	//??
}

func (p *metaRulesPopout) metaRuleSelected(name *string) {
	if name != nil {
		p.parseError = nil
		mrs := strings.ReplaceAll(p.core.settings.MetaRules[*name], "\t", "    ")
		p.editor.SetText(mrs)
	}
}

func (p *metaRulesPopout) sortedMetaRules() []string {
	result := make([]string, 0, len(p.core.settings.MetaRules))
	for name := range p.core.settings.MetaRules {
		result = append(result, name)
	}
	slices.Sort(result)
	return result
}
