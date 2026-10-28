package gui

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/gui/controls"
	"github.com/marrow16/gogol/gui/help"
	"github.com/marrow16/gogol/gui/icons"
	"github.com/marrow16/gogol/gui/settings"
	"github.com/marrow16/gogol/gui/shortcuts"
	"slices"
	"strings"
	"sync"
	"time"
)

func newShortcutsPopout(p *menuPopup, c *Core) *shortcutsPopout {
	result := &shortcutsPopout{
		parent:          p,
		core:            c,
		btnCreate:       controls.NewButton(theme, "Create"),
		btnDelete:       controls.NewButton(theme, "Delete"),
		btnRun:          controls.NewButton(theme, "Run"),
		errorsList:      widget.List{Axis: layout.Vertical},
		variablesList:   widget.List{Axis: layout.Vertical},
		btnVarsClearAll: controls.NewButton(theme, "Clear All"),
		addVarName:      controls.NewInput(theme, nil, 0, nil),
		btnAddVariable:  controls.NewIconButton(icons.Plus, (theme.TextSize*5)/4),
	}
	result.chooser = controls.NewChooser[string](theme, 38,
		result.sortedShortcuts(),
		result.shortcutSelected,
		func(name string) string {
			return name
		},
	)
	result.editor = controls.NewEditor(theme, result.editorChange).Mono()
	c.settings.ShortcutVariables.NotifyChanges(result.variableChanges)
	result.sortVariables(c.settings.ShortcutVariables.Clone())
	return result
}

type shortcutsPopout struct {
	parent        *menuPopup
	core          *Core
	chooser       *controls.Chooser[string]
	btnCreate     *controls.Button
	btnDelete     *controls.Button
	btnRun        *controls.Button
	editor        *controls.Editor
	syntaxCheckAt time.Time
	syntaxErrors  []error
	errorsList    widget.List
	linkHelp      widget.Clickable
	// variables...
	varsMutex       sync.Mutex
	variablesList   widget.List
	variables       []*variableHolder
	btnAddVariable  *controls.IconButton
	addVarName      *controls.Input
	varNameFocus    string
	btnVarsClearAll *controls.Button
}

func (p *shortcutsPopout) variableChanges(name string, del bool, all map[string]string) {
	if len(name) != 0 && !del {
		for _, h := range p.variables {
			if h.name == name {
				h.value = all[name]
				h.editor.SetText(h.value)
				return
			}
		}
	}
	p.sortVariables(all)
}

func (p *shortcutsPopout) sortVariables(all map[string]string) {
	p.varsMutex.Lock()
	defer p.varsMutex.Unlock()
	p.variables = make([]*variableHolder, 0, len(all))
	for k, v := range all {
		p.variables = append(p.variables, newVariableHolder(p.core.settings.ShortcutVariables, k, v))
	}
	slices.SortFunc(p.variables, func(a, b *variableHolder) int {
		return strings.Compare(a.name, b.name)
	})
}

const variablesShortcutName = "[$variables]"

func (p *shortcutsPopout) sortedShortcuts() []string {
	result := make([]string, 0, len(p.core.settings.Shortcuts))
	for name := range p.core.settings.Shortcuts {
		result = append(result, name)
	}
	slices.Sort(result)
	result = append(result, variablesShortcutName)
	return result
}

func (p *shortcutsPopout) shortcutSelected(name *string) {
	if name != nil {
		if sc, ok := p.core.settings.Shortcuts[*name]; ok {
			p.syntaxErrors = make([]error, 0)
			p.editor.SetText(strings.Join(sc, "\n"))
		} else {
			p.editor.SetText("")
		}
	}
}

func (p *shortcutsPopout) layout(gtx layout.Context) layout.Dimensions {
	selected := p.chooser.CurrentItem()
	curr := p.chooser.Text()
	isVars := selected != nil && *selected == variablesShortcutName
	if p.btnCreate.Clicked(gtx) {
		if selected == nil && curr != "" {
			if _, exists := p.core.settings.Shortcuts[curr]; !exists {
				p.core.settings.Shortcuts[curr] = []string{}
				p.chooser.ResetItems(p.sortedShortcuts())
				p.editor.SetText("")
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
				switch {
				case isVars:
					gtx.Constraints.Min.Y = ht
					return p.layoutVariables(gtx, editorHt)
				case selected != nil:
					gtx.Constraints.Min.Y = ht
					return p.layoutShortcut(gtx, editorHt)
				case curr == "":
					return label("No shortcut selected (select or enter new name)")(gtx)
				}
				canKey := p.isAllowedKey(curr)
				conflict := p.isConflictKey(curr)
				return flexHorizontal(20,
					rigid(p.btnCreate.Layout),
					conditionalRigid(canKey, label("Key invocable"), nil),
					conditionalRigid(canKey && conflict, label("(Overrides application key!)"), nil),
				)(gtx)
			}),
			conditionalRigid(!isVars && selected != nil && len(p.syntaxErrors) > 0, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Max.X = p.chooser.Width()
				return flexHorizontal(8,
					rigid(label("Errors:")),
					rigid(func(gtx layout.Context) layout.Dimensions {
						lineHt := measureText(gtx, "Xy").Size.Y
						maxHt := lineHt * min(len(p.syntaxErrors), 3)
						gtx.Constraints.Max.Y = maxHt
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						return material.List(theme, &p.errorsList).Layout(gtx, len(p.syntaxErrors), func(gtx layout.Context, index int) layout.Dimensions {
							lbl := material.Label(theme, theme.TextSize, p.syntaxErrors[index].Error())
							lbl.MaxLines = 1
							lbl.Color = errorColor
							return lbl.Layout(gtx)
						})
					}),
				)(gtx)
			}, nil),
		)(gtx)
		p.chooser.LayoutDropdown(gtx)
		return dims
	})
}

func (p *shortcutsPopout) isAllowedKey(k string) bool {
	if len(k) == 1 {
		return k == strings.ToUpper(k)
	}
	return len(k) > 1 && strings.Contains("F1,F2,F3,F4,F5,F6,F7,F8,F9,F10,F11,F12,", k+",")
}

func (p *shortcutsPopout) isConflictKey(k string) bool {
	return strings.Contains("012345678 BCEGHLMNPRSXZ,.[];'", k)
}

func (p *shortcutsPopout) layoutShortcut(gtx layout.Context, editorHt int) layout.Dimensions {
	if p.btnDelete.Clicked(gtx) {
		name := p.chooser.Text()
		delete(p.core.settings.Shortcuts, name)
		p.chooser.ResetItems(p.sortedShortcuts())
		p.chooser.SetText("")
		return layout.Dimensions{}
	}
	key := p.chooser.Text()
	if p.btnRun.Clicked(gtx) {
		p.core.runShortcut(key)
	}
	if p.linkHelp.Clicked(gtx) {
		p.core.showHelp(help.ShortCutsRef)
	}
	canKey := p.isAllowedKey(key)
	conflict := p.isConflictKey(key)
	if !p.syntaxCheckAt.IsZero() {
		if !gtx.Now.Before(p.syntaxCheckAt) {
			p.syntaxCheckAt = time.Time{}
			p.checkSyntax()
		} else {
			gtx.Execute(op.InvalidateCmd{At: p.syntaxCheckAt})
		}
	}
	return flexVertical(10,
		rigidLeftRight(0,
			flexHorizontal(20,
				rigid(p.btnRun.Layout),
				conditionalRigid(canKey, label("Key: "+altKeyName+key), nil),
				conditionalRigid(canKey && conflict, label("(Overrides application key!)"), nil),
			),
			p.btnDelete.Layout,
		),
		rigid(flexHorizontal(20,
			rigid(label("Actions:")),
			rigid(linkLabel(&p.linkHelp, "(see help)")),
		)),
		rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.Y = editorHt
			gtx.Constraints.Min.Y = editorHt
			gtx.Constraints.Min.X = p.chooser.Width()
			return p.editor.Layout(gtx)
		}),
	)(gtx)
}

func (p *shortcutsPopout) editorChange(text string, gtx layout.Context) {
	p.syntaxCheckAt = gtx.Now.Add(200 * time.Millisecond)
	name := p.chooser.Text()
	if _, ok := p.core.settings.Shortcuts[name]; ok {
		p.core.settings.Shortcuts[name] = strings.Split(text, "\n")
	}
}

func (p *shortcutsPopout) checkSyntax() {
	lines := strings.Split(p.editor.Text(), "\n")
	_, p.syntaxErrors = shortcuts.ParseShortcut("", lines)
}

func (p *shortcutsPopout) layoutVariables(gtx layout.Context, editorHt int) layout.Dimensions {
	if p.btnVarsClearAll.Clicked(gtx) {
		p.core.settings.ShortcutVariables.DeleteAll()
		p.sortVariables(p.core.settings.ShortcutVariables.Clone())
	}
	if p.btnAddVariable.Clicked(gtx) {
		name := p.addVarName.Text()
		if validVariableName(name) {
			if _, exists := p.core.settings.ShortcutVariables.Get(name); !exists {
				p.core.settings.ShortcutVariables.Set(name, "")
				p.addVarName.SetText("")
				p.varNameFocus = name
				window.Invalidate()
			}
		}
	}
	return flexVertical(10,
		rigidLeftRight(0,
			label("Variables:"),
			p.btnVarsClearAll.Layout,
		),
		rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			gtx.Constraints.Min.Y, gtx.Constraints.Max.Y = editorHt, editorHt
			names := make([]string, len(p.variables))
			for i, v := range p.variables {
				names[i] = "M$" + v.name
			}
			nw := measureMaxText(gtx, font.Bold, names...)
			variables := p.variables
			return widget.Border{Color: borderColor, Width: 1, CornerRadius: 3}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return material.List(theme, &p.variablesList).Layout(gtx, len(variables)+1, func(gtx layout.Context, index int) layout.Dimensions {
					if index >= len(p.variables) {
						return layout.Inset{Top: 4, Left: 4, Right: 4}.Layout(gtx, flexHorizontal(10,
							rigid(p.btnAddVariable.Layout),
							rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = gtx.Constraints.Max.X
								return p.addVarName.Layout(gtx)
							}),
						))
					}
					if p.varNameFocus == variables[index].name {
						p.varNameFocus = ""
						variables[index].editor.SetFocused(gtx)
					}
					return variables[index].layout(gtx, nw)
				})
			})
		}),
	)(gtx)
}

type variableHolder struct {
	variables *settings.Variables
	name      string
	value     string
	editor    *controls.Input
	btnDelete *controls.IconButton
}

func (h *variableHolder) layout(gtx layout.Context, nameWidth int) layout.Dimensions {
	if h.btnDelete.Clicked(gtx) {
		h.variables.Delete(h.name)
		window.Invalidate()
	}
	return layout.Inset{Top: 4, Left: 4, Right: 4}.Layout(gtx, flexHorizontal(10,
		rigid(h.btnDelete.Layout),
		rigidLabel("$"+h.name, 0, font.Bold, nameWidth),
		rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return h.editor.Layout(gtx)
		})))
}

func (h *variableHolder) valueChanged(s string) {
	if h.value != s {
		h.value = s
		h.variables.Set(h.name, s)
	}
}

func newVariableHolder(variables *settings.Variables, name string, value string) *variableHolder {
	result := &variableHolder{
		variables: variables,
		name:      name,
		value:     value,
		btnDelete: controls.NewIconButton(icons.Minus, (theme.TextSize*5)/4),
	}
	result.editor = controls.NewInput(theme, nil, 0, result.valueChanged)
	result.editor.SetText(value)
	return result
}

func (p *shortcutsPopout) hasFocus(gtx layout.Context) bool {
	for _, h := range p.variables {
		if h.editor.IsFocused(gtx) {
			return true
		}
	}
	return p.chooser.IsFocused(gtx) ||
		p.btnCreate.IsFocused(gtx) || p.btnDelete.IsFocused(gtx) || p.btnRun.IsFocused(gtx) ||
		p.editor.IsFocused(gtx) || p.addVarName.IsFocused(gtx)
}

func (p *shortcutsPopout) reset() {}
