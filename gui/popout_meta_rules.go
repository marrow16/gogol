package gui

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/marrow16/gogol/gui/controls"
	"github.com/marrow16/gogol/gui/help"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/logic/meta"
	"slices"
	"strconv"
	"strings"
)

type metaRulesPopout struct {
	parent       *menuPopup
	core         *Core
	chooser      *controls.Chooser[string]
	btnCreate    *controls.Button
	btnDelete    *controls.Button
	editor       *controls.Editor
	mode         *widget.Enum
	radioEdit    *controls.RadioButton
	radioMatches *controls.RadioButton
	parseError   error
	linkHelp     widget.Clickable
	listMatched  *controls.List[logic.Rule]
	btnAddAll    *controls.Button
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
		btnCreate: controls.NewButton(theme, "Create"),
		btnDelete: controls.NewButton(theme, "Delete"),
		btnAddAll: controls.NewButton(theme, "Add all to collected rules"),
	}
	result.chooser = controls.NewChooser[string](theme, 38,
		result.sortedMetaRules(),
		result.metaRuleSelected,
		func(name string) string {
			return name
		},
	)
	result.chooser.SetText(c.settings.SelectedMetaRule)
	result.editor = controls.NewEditor(theme, result.editorChange).Mono()
	result.mode = &widget.Enum{Value: metaRuleEdit}
	result.radioEdit = controls.NewRadioButton(theme, result.mode, metaRuleEdit, "Edit")
	result.radioMatches = controls.NewRadioButton(theme, result.mode, metaRuleMatches, "Matching Rules")
	result.listMatched = controls.NewList[logic.Rule](theme, []logic.Rule{}, true).
		RowRenderer(result.layoutMatchedRule).
		OnItemSelect(func(r logic.Rule, _ bool) {
			c.gridHolder.grid.SetRule(r)
			window.Invalidate()
		}).
		OnIsSelected(func(index int, r logic.Rule) bool {
			return r.Permutation() == c.gridHolder.grid.Rule().Permutation()
		})
	return result
}

func (p *metaRulesPopout) layout(gtx layout.Context) layout.Dimensions {
	selected := p.chooser.CurrentItem()
	currName := p.chooser.Text()
	if p.btnCreate.Clicked(gtx) {
		if selected == nil && currName != "" {
			if _, exists := p.core.settings.MetaRules[currName]; !exists {
				p.core.settings.MetaRules[currName] = "B() / S()"
				p.chooser.ResetItems(p.sortedMetaRules())
				p.editor.SetText("B() / S()")
			}
		}
	}
	mt := measureText(gtx, "Xy")
	ht := mt.Size.Y * 15
	editorHt := ht - mt.Size.Y
	return popoutLayout(gtx, func(gtx layout.Context) layout.Dimensions {
		dims := flexVertical(10,
			rigid(p.chooser.Layout),
			rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Max.X = p.chooser.Width()
				if selected != nil {
					gtx.Constraints.Min.Y = ht
					return p.layoutMetaRule(gtx, editorHt)
				} else if currName == "" {
					return label("No meta rule selected (select or enter new name)")(gtx)
				}
				return flexHorizontal(0, rigid(p.btnCreate.Layout))(gtx)
			}),
		)(gtx)
		p.chooser.LayoutDropdown(gtx)
		return dims
	})
}

func (p *metaRulesPopout) layoutMetaRule(gtx layout.Context, editorHt int) layout.Dimensions {
	if p.btnDelete.Clicked(gtx) {
		name := p.chooser.Text()
		delete(p.core.settings.MetaRules, name)
		p.chooser.ResetItems(p.sortedMetaRules())
		return layout.Dimensions{}
	}
	p.mode.Update(gtx)
	return flexVertical(10,
		rigid(flexHorizontal(20,
			rigid(p.btnDelete.Layout),
			rigid(p.radioEdit.Layout),
			rigid(p.radioMatches.Layout),
		)),
		rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.Y = editorHt
			gtx.Constraints.Min.Y = editorHt
			gtx.Constraints.Min.X = p.chooser.Width()
			switch p.mode.Value {
			case metaRuleMatches:
				return p.layoutMatchedRules(gtx, editorHt)
			default:
				return p.editor.Layout(gtx)
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
		p.core.showHelp(help.MetaRulesRef)
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
	}
	return label("No matched rules found")(gtx)
}

func (p *metaRulesPopout) layoutMatchedRules(gtx layout.Context, editorHt int) layout.Dimensions {
	gtx.Constraints.Max.Y = editorHt
	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	gtx.Constraints.Min.X = p.chooser.Width()
	p.parseCurrent(p.editor.Text())
	if p.parseError != nil || p.currentRule == nil {
		return widget.Border{Color: borderColor, CornerRadius: 3, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return label("Error - No matching rules")(gtx)
			})
		})
	}
	count := len(p.matched)
	if count == 0 {
		return widget.Border{Color: borderColor, CornerRadius: 3, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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

func (p *metaRulesPopout) parseCurrent(text string) {
	//text := p.editor.Text()
	if text != p.current || p.currentRule == nil {
		p.current = text
		if r, err := meta.ParseRule(text); err == nil {
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
		p.listMatched.ResetItems(p.matched)
	}
}

func (p *metaRulesPopout) editorChange(text string, _ layout.Context) {
	p.parseCurrent(text)
	if p.parseError == nil {
		name := p.chooser.Text()
		if _, ok := p.core.settings.MetaRules[name]; ok {
			p.core.settings.MetaRules[name] = strings.ReplaceAll(text, "    ", "\t")
		}
	}
}

func (p *metaRulesPopout) hasFocus(gtx layout.Context) bool {
	on, radiosFocused := p.mode.Focused()
	return radiosFocused || p.chooser.IsFocused(gtx) ||
		p.btnCreate.IsFocused(gtx) || p.btnDelete.IsFocused(gtx) ||
		p.editor.IsFocused(gtx) || p.btnAddAll.IsFocused(gtx) ||
		(on == metaRuleMatches && p.listMatched.IsFocused(gtx))
}

func (p *metaRulesPopout) reset() {
	// nothing to reset
}

func (p *metaRulesPopout) metaRuleSelected(name *string) {
	if name != nil {
		p.core.settings.SelectedMetaRule = *name
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
