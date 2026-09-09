package widgets

import (
	"runtime"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
)

const (
	gogolGuiVersion = "1.2.50"
	gogolRepo       = "https://github.com/marrow16/gogol"
)

func newAboutPopout(p *menuPopup) *aboutPopout {
	result := &aboutPopout{
		parent: p,
		v:      "Version: " + gogolGuiVersion + " (Go: " + strings.TrimPrefix(runtime.Version(), "go") + ")",
	}
	result.links = map[*widget.Clickable]string{
		&result.linkRepo: gogolRepo,
	}
	return result
}

type aboutPopout struct {
	parent   *menuPopup
	linkRepo widget.Clickable
	links    map[*widget.Clickable]string
	v        string
}

func (p *aboutPopout) layout(gtx layout.Context) layout.Dimensions {
	m := measureText(gtx, "M")
	minX := m.Size.X * 30
	for c, l := range p.links {
		if c.Clicked(gtx) {
			_ = openURL(l)
		}
	}
	return popoutLayout(gtx, flexVertical(0,
		rigidLabel("GoGoL", text.Middle, font.Bold, minX),
		rigidLabel(p.v, text.Middle, 0, minX),
		rigidSpacerVertical(m.Size.Y/2),
		rigidLabel("Author: Martin \"Marrow\" Rowlinson", text.Middle, 0, minX),
		rigidFixedWidth(linkLabel(&p.linkRepo, gogolRepo), minX, layout.Center),
		rigidSpacerVertical(m.Size.Y/2),
	))
}

func (p *aboutPopout) hasFocus(gtx layout.Context) bool {
	return false
}

func (p *aboutPopout) reset() {
	// nothing to reset
}
