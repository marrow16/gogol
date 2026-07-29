package widgets

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/logic/meta"
	"image"
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
	list         widget.List
	ruleClicks   []widget.Clickable
	parseError   error
	linkHelp     widget.Clickable
}

const (
	metaRuleEdit    = "edit"
	metaRuleMatches = "matches"
)

func newMetaRulesPopout(p *menuPopup, c *Core) *metaRulesPopout {
	result := &metaRulesPopout{
		parent:    p,
		core:      c,
		btnCreate: newButton(c.theme, "Create"),
		btnDelete: newButton(c.theme, "Delete"),
	}
	result.chooser = newChooser[string](c.theme, 38,
		result.sortedMetaRules(),
		result.metaRuleSelected,
		func(name string) string {
			return name
		},
	)
	result.mode = &widget.Enum{Value: metaRuleEdit}
	result.radioEdit = newRadioButton(c.theme, result.mode, metaRuleEdit, "Edit")
	result.radioMatches = newRadioButton(c.theme, result.mode, metaRuleMatches, "Matching Rules")
	result.list.Axis = layout.Vertical
	return result
}

func (p *metaRulesPopout) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
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
	mt := measureText(gtx, theme, "Xy")
	ht := mt.Size.Y * 16
	editorHt := ht - mt.Size.Y
	return layout.Inset{Left: 8, Right: 8, Top: 8, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		var chooserDims layout.Dimensions
		dims := layout.Flex{Axis: layout.Vertical, Gap: 10}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						chooserDims = p.chooser.layout(gtx)
						return chooserDims
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Max.X = p.chooser.dims.Size.X
				if selected != nil {
					gtx.Constraints.Min.Y = ht
					return p.layoutMetaRule(gtx, theme, editorHt, chooserDims)
				} else if currName == "" {
					return label(theme, "No meta rule selected (select or enter new name)")(gtx)
				} else {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(p.btnCreate.Layout),
					)
				}
			}),
		)
		p.chooser.layoutDropdown(gtx, chooserDims)
		return dims
	})
}

func (p *metaRulesPopout) layoutMetaRule(gtx layout.Context, theme *material.Theme, editorHt int, chooserDims layout.Dimensions) layout.Dimensions {
	if p.btnDelete.Clicked(gtx) {
		name := p.chooser.editor.Text()
		delete(p.core.settings.MetaRules, name)
		p.chooser.resetItems(p.sortedMetaRules())
		return layout.Dimensions{}
	}
	p.mode.Update(gtx)
	p.updateEditor(gtx)
	if p.linkHelp.Clicked(gtx) {
		_ = openURL("https://github.com/marrow16/gogol/blob/main/logic/meta/README.md")
	}
	return layout.Flex{Axis: layout.Vertical, Gap: 10}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
				layout.Rigid(p.btnDelete.Layout),
				layout.Rigid(p.radioEdit.Layout),
				layout.Rigid(p.radioMatches.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: image.Point{X: 100}}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Top: 2}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return material.Clickable(gtx, &p.linkHelp, label(theme, "(see help)"))
					})
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.Y = editorHt
			gtx.Constraints.Min.Y = editorHt
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
			gtx.Constraints.Min.X = p.chooser.dims.Size.X
			switch p.mode.Value {
			case metaRuleMatches:
				return p.layoutMatchedRules(gtx, theme, editorHt)
			default:
				return p.layoutEditor(gtx, theme, editorHt)
			}
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if p.parseError != nil {
				return errorLabel(theme, p.parseError)(gtx)
			}
			return layout.Dimensions{}
		}),
	)
}

func (p *metaRulesPopout) layoutEditor(gtx layout.Context, theme *material.Theme, editorHt int) layout.Dimensions {
	gtx.Constraints.Max.Y = editorHt
	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	gtx.Constraints.Min.X = p.chooser.dims.Size.X
	style := material.Editor(theme, &p.editor, "")
	borderColor := popupBorder
	borderThickness := unit.Dp(1)
	if gtx.Focused(&p.editor) {
		borderColor = popupBorderFocused
		borderThickness = unit.Dp(2)
	}
	return widget.Border{
		Color:        borderColor,
		CornerRadius: 3,
		Width:        borderThickness,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    2,
			Bottom: 2,
			Left:   4,
			Right:  4,
		}.Layout(gtx, style.Layout)
	})
}

func (p *metaRulesPopout) layoutMatchedRules(gtx layout.Context, theme *material.Theme, editorHt int) layout.Dimensions {
	gtx.Constraints.Max.Y = editorHt
	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	gtx.Constraints.Min.X = p.chooser.dims.Size.X
	mrs, ok := p.core.settings.MetaRules[p.chooser.editor.Text()]
	if !ok {
		return layout.Dimensions{}
	}
	var mr meta.Evaluator
	mr, p.parseError = meta.ParseRule(mrs)
	if p.parseError != nil {
		return widget.Border{Color: popupBorder, CornerRadius: 3, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return label(theme, "Error - No matching rules")(gtx)
			})
		})
	}
	rules := make([]logic.Rule, 0)
	for r := range mr.MatchingRules() {
		rules = append(rules, r)
	}
	if len(rules) == 0 {
		return widget.Border{Color: popupBorder, CornerRadius: 3, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return label(theme, "No matching rules")(gtx)
			})
		})
	}
	if len(rules) != len(p.ruleClicks) {
		p.ruleClicks = make([]widget.Clickable, len(rules))
	}
	return widget.Border{Color: popupBorder, CornerRadius: 3, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return material.List(p.core.theme, &p.list).Layout(
				gtx, len(rules),
				func(gtx layout.Context, index int) layout.Dimensions {
					r := rules[index]
					name := r.Rle()
					if !r.IsCustom() {
						name = "\"" + r.Name() + "\""
					}
					for p.ruleClicks[index].Clicked(gtx) {
						gtx.Execute(key.FocusCmd{Tag: &p.list})
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
					return p.ruleClicks[index].Layout(gtx, label(theme, name+" ("+strconv.Itoa(r.Permutation())+")   #"+strconv.Itoa(index+1)+"/"+strconv.Itoa(len(rules))))
				},
			)
		})
	})
}

func (p *metaRulesPopout) updateEditor(gtx layout.Context) {
	for {
		ev, ok := p.editor.Update(gtx)
		if !ok {
			break
		}
		if _, ok = ev.(widget.ChangeEvent); ok {
			mrStr := p.editor.Text()
			_, p.parseError = meta.ParseRule(mrStr)
			if p.parseError == nil {
				p.clearList()
				name := p.chooser.editor.Text()
				if _, ok = p.core.settings.MetaRules[name]; ok {
					p.core.settings.MetaRules[name] = strings.ReplaceAll(mrStr, "    ", "\t")
				}
			}
		}
	}
}

func (p *metaRulesPopout) clearList() {
	p.ruleClicks = make([]widget.Clickable, 0)
}

func (p *metaRulesPopout) hasFocus(gtx layout.Context) bool {
	_, radiosFocused := p.mode.Focused()
	return radiosFocused || p.chooser.isFocused(gtx) ||
		p.btnCreate.isFocused(gtx) || p.btnDelete.isFocused(gtx) ||
		gtx.Focused(&p.editor) || gtx.Focused(&p.list)
}

func (p *metaRulesPopout) reset() {
	p.clearList()
}

func (p *metaRulesPopout) metaRuleSelected(name *string) {
	if name != nil {
		p.parseError = nil
		mrs := strings.ReplaceAll(p.core.settings.MetaRules[*name], "\t", "    ")
		p.editor.SetText(mrs)
		p.clearList()
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
