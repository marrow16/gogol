package widgets

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/cmd/gui/help"
	"slices"
	"strings"
)

func newShortcutsPopout(p *menuPopup, c *Core) *shortcutsPopout {
	result := &shortcutsPopout{
		parent:    p,
		core:      c,
		btnCreate: newButton("Create"),
		btnDelete: newButton("Delete"),
		btnRun:    newButton("Run"),
	}
	result.chooser = newChooser[string](38,
		result.sortedShortcuts(),
		result.shortcutSelected,
		func(name string) string {
			return name
		},
	)
	return result
}

type shortcutsPopout struct {
	parent    *menuPopup
	core      *Core
	chooser   *chooser[string]
	btnCreate *button
	btnDelete *button
	btnRun    *button
	editor    widget.Editor
	linkHelp  widget.Clickable
}

func (p *shortcutsPopout) sortedShortcuts() []string {
	result := make([]string, 0, len(p.core.settings.Shortcuts))
	for name := range p.core.settings.Shortcuts {
		result = append(result, name)
	}
	slices.Sort(result)
	return result
}

func (p *shortcutsPopout) shortcutSelected(name *string) {
	if name != nil {
		if sc, ok := p.core.settings.Shortcuts[*name]; ok {
			p.editor.SetText(strings.Join(sc, "\n"))
		} else {
			p.editor.SetText("")
		}
	}
}

func (p *shortcutsPopout) layout(gtx layout.Context) layout.Dimensions {
	selected := p.chooser.currentItem()
	curr := p.chooser.editor.Text()
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
				if selected != nil {
					gtx.Constraints.Min.Y = ht
					return p.layoutShortcut(gtx, editorHt)
				} else if curr == "" {
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
		rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
				rigid(flexHorizontal(20,
					rigid(p.btnRun.Layout),
					conditionalRigid(canKey, label("Key: "+altKeyName+key), nil),
					conditionalRigid(canKey && conflict, label("(Overrides application key!)"), nil),
				)),
				rigid(p.btnDelete.Layout),
			)
		}),
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
			name := p.chooser.editor.Text()
			if _, ok = p.core.settings.Shortcuts[name]; ok {
				p.core.settings.Shortcuts[name] = strings.Split(p.editor.Text(), "\n")
			}
		}
	}
}

func (p *shortcutsPopout) hasFocus(gtx layout.Context) bool {
	return p.chooser.isFocused(gtx) ||
		p.btnCreate.isFocused(gtx) || p.btnDelete.isFocused(gtx) || p.btnRun.isFocused(gtx) ||
		gtx.Focused(&p.editor)
}

func (p *shortcutsPopout) reset() {}
