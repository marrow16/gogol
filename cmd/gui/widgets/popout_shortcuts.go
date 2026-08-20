package widgets

import (
	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/go-andiamo/splitter"
	"slices"
	"strings"
)

func newShortcutsPopout(p *menuPopup, c *Core) *shortcutsPopout {
	result := &shortcutsPopout{
		parent: p,
		core:   c,
	}
	result.key = newInput("", 3, result.keyChanged).upDownSupport(result.keyUpDown).maximumWidth(3)
	return result
}

type shortcutsPopout struct {
	parent     *menuPopup
	core       *Core
	key        *input
	onShortcut string
	editor     widget.Editor
	linkHelp   widget.Clickable
}

func (p *shortcutsPopout) keyUpDown(k key.Name, text string) (string, bool) {
	shortcuts := make([]string, 0)
	for sc := range p.core.settings.Shortcuts {
		if p.isAllowedKey(strings.ToUpper(sc)) {
			shortcuts = append(shortcuts, sc)
		}
	}
	if len(shortcuts) == 0 {
		return "", false
	}
	slices.Sort(shortcuts)
	idx := slices.Index(shortcuts, text)
	switch k {
	case key.NameUpArrow:
		if idx == -1 {
			idx = len(shortcuts) - 1
		} else {
			idx--
		}
	case key.NameDownArrow:
		if idx == -1 {
			idx = 0
		} else {
			idx++
		}
	}
	if idx < 0 {
		idx = len(shortcuts) - 1
	} else if idx >= len(shortcuts) {
		idx = 0
	}
	return shortcuts[idx], true
}

func (p *shortcutsPopout) isAllowedKey(k string) bool {
	if len(k) == 1 {
		return true
	}
	return len(k) > 1 && strings.Contains("F1,F2,F3,F4,F5,F6,F7,F8,F9,F10,F11,F12,", k+",")
}

func (p *shortcutsPopout) keyChanged(s string) {
	us := strings.ToUpper(s)
	if us != s {
		p.key.setText(us)
	}
	if !p.isAllowedKey(us) {
		return
	} else if p.onShortcut != us {
		if sc, ok := p.core.settings.Shortcuts[us]; ok {
			p.editor.SetText(strings.Join(sc, "\n"))
		} else {
			p.editor.SetText("")
		}
	}
	p.onShortcut = us
}

var lineSplitter = splitter.MustCreateSplitter('\n').AddDefaultOptions(splitter.TrimSpaces, splitter.IgnoreEmpties)

func (p *shortcutsPopout) layout(gtx layout.Context) layout.Dimensions {
	kw := measureText(gtx, "Key: ")
	ht := kw.Size.Y * 16
	m := measureText(gtx, "M")
	ew := m.Size.X * 30
	k := strings.ToUpper(p.key.editor.Text())
	isAllowedKey := p.isAllowedKey(k)
	if p.linkHelp.Clicked(gtx) {
		_ = openURL(shortcutsHelp)
	}
	if isAllowedKey {
		for {
			ev, ok := p.editor.Update(gtx)
			if !ok {
				break
			}
			switch ev.(type) {
			case widget.ChangeEvent:
				if lines, err := lineSplitter.Split(p.editor.Text()); err == nil {
					if len(lines) == 0 {
						delete(p.core.settings.Shortcuts, k)
					} else {
						p.core.settings.Shortcuts[k] = lines
					}
				}
			}
		}
	}
	return layout.Inset{Left: 8, Right: 8, Top: 8, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min.X = ew
		gtx.Constraints.Min.Y = ht
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(flexHorizontal(20,
				rigidLabel("Key:", text.End, font.Bold, kw.Size.X),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X = m.Size.X * 2
					return p.key.layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !isAllowedKey {
						return layout.Dimensions{}
					}
					return label(altKeyName + k)(gtx)
				}),
			)),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: 3}.Layout(gtx, flexHorizontal(20,
					rigidLabel("Actions:", 0, 0, 0),
					layout.Rigid(linkLabel(&p.linkHelp, "(see help)")),
				))
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if !isAllowedKey {
					return layout.Dimensions{}
				}
				gtx.Constraints.Max.X = ew
				gtx.Constraints.Max.Y = ht - kw.Size.Y*2
				gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
				style := material.Editor(theme, &p.editor, "")
				bc, bt := focusedBorder(gtx.Focused(&p.editor))
				return widget.Border{Color: bc, CornerRadius: 3, Width: bt}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, style.Layout)
				})
			}),
		)
	})
}

func (p *shortcutsPopout) hasFocus(gtx layout.Context) bool {
	return p.key.isFocused(gtx) || gtx.Focused(&p.editor)
}

func (p *shortcutsPopout) reset() {}
