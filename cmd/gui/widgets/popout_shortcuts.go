package widgets

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/unit"
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
	result.key = newInput(c.theme, "", 1, result.keyChanged).upDownSupport(result.keyUpDown).maximumWidth(2)
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
		if len(sc) == 1 {
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

func (p *shortcutsPopout) keyChanged(s string) {
	us := strings.ToUpper(s)
	if us != s {
		p.key.setText(us)
	}
	if len(us) != 1 {
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

func (p *shortcutsPopout) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	kw := measureText(gtx, theme, "Key: ")
	ht := kw.Size.Y * 16
	m := measureText(gtx, theme, "M")
	ew := m.Size.X * 30
	k := p.key.editor.Text()
	if p.linkHelp.Clicked(gtx) {
		_ = openURL("https://github.com/marrow16/gogol/tree/main/cmd/gui/SHORTCUTS.md")
	}
	if len(k) == 1 {
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
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
					layout.Rigid(rightAlignedBoldLabel(theme, "Key:", kw.Size.X)),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X = m.Size.X * 2
						return p.key.layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if len(k) != 1 {
							return layout.Dimensions{}
						}
						return label(theme, altKeyName+k)(gtx)
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
					layout.Rigid(label(theme, "Actions:")),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Clickable(gtx, &p.linkHelp, label(theme, "(see help)"))
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if len(k) != 1 {
					return layout.Dimensions{}
				}
				gtx.Constraints.Max.X = ew
				gtx.Constraints.Max.Y = ht - kw.Size.Y*2
				gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
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
			}),
		)
	})
}

func (p *shortcutsPopout) hasFocus(gtx layout.Context) bool {
	return p.key.isFocused(gtx) || gtx.Focused(&p.editor)
}

func (p *shortcutsPopout) reset() {}
