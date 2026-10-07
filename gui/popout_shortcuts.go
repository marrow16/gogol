package gui

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/gui/controls"
	"github.com/marrow16/gogol/gui/help"
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
	}
	result.chooser = newChooser[string](38,
		result.sortedShortcuts(),
		result.shortcutSelected,
		func(name string) string {
			return name
		},
	)
	c.settings.ShortcutVariables.NotifyChanges(result.variableChanges)
	result.sortVariables(c.settings.ShortcutVariables.Clone())
	return result
}

type shortcutsPopout struct {
	parent        *menuPopup
	core          *Core
	chooser       *chooser[string]
	btnCreate     *controls.Button
	btnDelete     *controls.Button
	btnRun        *controls.Button
	editor        widget.Editor
	syntaxCheckAt time.Time
	syntaxErrors  []error
	errorsList    widget.List
	linkHelp      widget.Clickable
	// variables...
	varsMutex       sync.Mutex
	variablesList   widget.List
	variables       [][2]string
	btnVarsClearAll *controls.Button
}

func (p *shortcutsPopout) variableChanges(_ string, _ bool, all map[string]string) {
	p.sortVariables(all)
}

func (p *shortcutsPopout) sortVariables(all map[string]string) {
	p.varsMutex.Lock()
	defer p.varsMutex.Unlock()
	p.variables = make([][2]string, 0, len(all))
	for k, v := range all {
		p.variables = append(p.variables, [2]string{k, v})
	}
	slices.SortFunc(p.variables, func(a, b [2]string) int {
		return strings.Compare(a[0], b[0])
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
	selected := p.chooser.currentItem()
	curr := p.chooser.editor.Text()
	isVars := selected != nil && *selected == variablesShortcutName
	if p.btnCreate.Clicked(gtx) {
		if selected == nil && curr != "" {
			if _, exists := p.core.settings.Shortcuts[curr]; !exists {
				p.core.settings.Shortcuts[curr] = []string{}
				p.chooser.resetItems(p.sortedShortcuts())
				p.editor.SetText("")
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
				gtx.Constraints.Max.X = p.chooser.dims.Size.X
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
		p.chooser.layoutDropdown(gtx)
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
		name := p.chooser.editor.Text()
		delete(p.core.settings.Shortcuts, name)
		p.chooser.resetItems(p.sortedShortcuts())
		p.chooser.setText("")
		return layout.Dimensions{}
	}
	key := p.chooser.editor.Text()
	if p.btnRun.Clicked(gtx) {
		p.core.runShortcut(key)
	}
	if p.linkHelp.Clicked(gtx) {
		p.core.showHelp(help.ShortCutsRef)
	}
	canKey := p.isAllowedKey(key)
	conflict := p.isConflictKey(key)
	p.updateEditor(gtx)
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
			gtx.Constraints.Min.X = p.chooser.dims.Size.X
			style := material.Editor(theme, &p.editor, "")
			bc, bt := focusedBorder(gtx.Focused(&p.editor))
			return widget.Border{Color: bc, CornerRadius: 3, Width: bt}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, style.Layout)
			})
		}),
	)(gtx)
}

func (p *shortcutsPopout) updateEditor(gtx layout.Context) {
	for {
		ev, ok := p.editor.Update(gtx)
		if !ok {
			break
		}
		if _, ok = ev.(widget.ChangeEvent); ok {
			p.syntaxCheckAt = gtx.Now.Add(250 * time.Millisecond)
			name := p.chooser.editor.Text()
			if _, ok = p.core.settings.Shortcuts[name]; ok {
				p.core.settings.Shortcuts[name] = strings.Split(p.editor.Text(), "\n")
			}
		}
	}
	if !p.syntaxCheckAt.IsZero() {
		if !gtx.Now.Before(p.syntaxCheckAt) {
			p.syntaxCheckAt = time.Time{}
			p.checkSyntax()
		} else {
			gtx.Execute(op.InvalidateCmd{At: p.syntaxCheckAt})
		}
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
	return flexVertical(10,
		rigidLeftRight(0,
			label("Variables:"),
			p.btnVarsClearAll.Layout,
		),
		rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			gtx.Constraints.Min.Y, gtx.Constraints.Max.Y = editorHt, editorHt
			return widget.Border{Color: popupBorder, Width: 1, CornerRadius: 3}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return material.List(theme, &p.variablesList).Layout(gtx, len(p.variables), func(gtx layout.Context, index int) layout.Dimensions {
					return label(" $" + p.variables[index][0] + ` = "` + p.variables[index][1] + `"`)(gtx)
				})
			})
		}),
	)(gtx)
}

func (p *shortcutsPopout) hasFocus(gtx layout.Context) bool {
	return p.chooser.isFocused(gtx) ||
		p.btnCreate.IsFocused(gtx) || p.btnDelete.IsFocused(gtx) || p.btnRun.IsFocused(gtx) ||
		gtx.Focused(&p.editor)
}

func (p *shortcutsPopout) reset() {}
