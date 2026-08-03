package widgets

import (
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
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
	listStyle    material.ListStyle
	ruleClicks   []widget.Clickable
	parseError   error
	linkHelp     widget.Clickable

	btnFirst    *button
	btnLast     *button
	btnNext     *button
	btnPrev     *button
	current     string
	currentRule meta.Evaluator
	matched     []logic.Rule
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
		btnFirst:  newButton(c.theme, "First"),
		btnLast:   newButton(c.theme, "Last"),
		btnNext:   newButton(c.theme, "Next"),
		btnPrev:   newButton(c.theme, "Prev"),
	}
	result.listStyle = material.List(c.theme, &result.list)
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
	ht := mt.Size.Y * 15
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
					return p.layoutMetaRule(gtx, theme, editorHt)
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

func (p *metaRulesPopout) layoutMetaRule(gtx layout.Context, theme *material.Theme, editorHt int) layout.Dimensions {
	if p.btnDelete.Clicked(gtx) {
		name := p.chooser.editor.Text()
		delete(p.core.settings.MetaRules, name)
		p.chooser.resetItems(p.sortedMetaRules())
		return layout.Dimensions{}
	}
	p.mode.Update(gtx)
	p.updateEditor(gtx)
	return layout.Flex{Axis: layout.Vertical, Gap: 10}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
				layout.Rigid(p.btnDelete.Layout),
				layout.Rigid(p.radioEdit.Layout),
				layout.Rigid(p.radioMatches.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: image.Point{X: 100}}
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
			switch {
			case p.parseError != nil:
				return errorLabel(theme, p.parseError)(gtx)
			case p.mode.Value == metaRuleMatches && p.currentRule != nil:
				return p.layoutNavButtons(gtx, theme)
			case p.currentRule != nil:
				return p.layoutReport(gtx, theme)
			}
			return layout.Dimensions{}
		}),
	)
}

func (p *metaRulesPopout) layoutReport(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	if p.linkHelp.Clicked(gtx) {
		_ = openURL("https://github.com/marrow16/gogol/blob/main/logic/meta/README.md")
	}
	return layout.Flex{Axis: layout.Horizontal, Gap: 30}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Clickable(gtx, &p.linkHelp, label(theme, "(see help)"))
		}),
		layout.Rigid(label(theme, "Matched rules: "+strconv.Itoa(len(p.matched)))),
	)
}

func (p *metaRulesPopout) layoutNavButtons(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	if len(p.matched) > 0 {
		if p.btnFirst.Clicked(gtx) {
			p.core.setRule(p.matched[0])
			p.list.ScrollTo(0)
		}
		if p.btnLast.Clicked(gtx) {
			p.core.setRule(p.matched[len(p.matched)-1])
			p.list.ScrollTo(len(p.matched) - 1)
		}
		if p.btnNext.Clicked(gtx) {
			currPerm := p.core.gridHolder.grid.Rule.Permutation()
			if next := p.currentRule.Next(currPerm); next != currPerm {
				if idx := slices.IndexFunc(p.matched, func(r logic.Rule) bool {
					return r.Permutation() == next
				}); idx != -1 {
					p.core.setRule(p.matched[idx])
					p.list.ScrollTo(idx)
				}
			} else {
				p.core.setRule(p.matched[0])
				p.list.ScrollTo(0)
			}
		}
		if p.btnPrev.Clicked(gtx) {
			currPerm := p.core.gridHolder.grid.Rule.Permutation()
			if prev := p.currentRule.Previous(currPerm); prev != currPerm {
				if idx := slices.IndexFunc(p.matched, func(r logic.Rule) bool {
					return r.Permutation() == prev
				}); idx != -1 {
					p.core.setRule(p.matched[idx])
					p.list.ScrollTo(idx)
				}
			} else {
				p.core.setRule(p.matched[len(p.matched)-1])
				p.list.ScrollTo(len(p.matched) - 1)
			}
		}
		return layout.Flex{Axis: layout.Horizontal, Gap: 30}.Layout(gtx,
			layout.Rigid(p.btnFirst.Layout),
			layout.Rigid(p.btnLast.Layout),
			layout.Rigid(p.btnNext.Layout),
			layout.Rigid(p.btnPrev.Layout),
		)
	} else {
		return label(theme, "No matched rules found")(gtx)
	}
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
		return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, style.Layout)
	})
}

func (p *metaRulesPopout) layoutMatchedRules(gtx layout.Context, theme *material.Theme, editorHt int) layout.Dimensions {
	gtx.Constraints.Max.Y = editorHt
	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	gtx.Constraints.Min.X = p.chooser.dims.Size.X
	p.parseCurrent()
	if p.parseError != nil || p.currentRule == nil {
		return widget.Border{Color: popupBorder, CornerRadius: 3, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return label(theme, "Error - No matching rules")(gtx)
			})
		})
	}
	count := len(p.matched)
	if count == 0 {
		return widget.Border{Color: popupBorder, CornerRadius: 3, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return label(theme, "No matching rules")(gtx)
			})
		})
	}
	if count != len(p.ruleClicks) {
		p.ruleClicks = make([]widget.Clickable, len(p.matched))
	}
	return widget.Border{Color: popupBorder, CornerRadius: 3, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return p.listStyle.Layout(
				gtx, len(p.matched),
				func(gtx layout.Context, index int) layout.Dimensions {
					r := p.matched[index]
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
					return p.ruleClicks[index].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Left: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal, Gap: 32}.Layout(gtx,
								layout.Flexed(3.5, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(p.core.theme, p.core.theme.TextSize, name+" ("+strconv.Itoa(r.Permutation())+")")
									lbl.Alignment = text.Start
									lbl.MaxLines = 1
									return lbl.Layout(gtx)
								}),
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(p.core.theme, p.core.theme.TextSize, strconv.Itoa(index+1))
									lbl.Alignment = text.End
									lbl.MaxLines = 1
									return lbl.Layout(gtx)
								}),
							)
						})
					})
				},
			)
		})
	})
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
				p.clearList()
				name := p.chooser.editor.Text()
				if _, ok = p.core.settings.MetaRules[name]; ok {
					p.core.settings.MetaRules[name] = strings.ReplaceAll(p.current, "    ", "\t")
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
		gtx.Focused(&p.editor) ||
		p.btnFirst.isFocused(gtx) || p.btnLast.isFocused(gtx) ||
		p.btnPrev.isFocused(gtx) || p.btnNext.isFocused(gtx) ||
		gtx.Focused(&p.list)
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
