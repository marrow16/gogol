package widgets

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
)

const (
	gogolGuiVersion = "1.0.43"
	gogolRepo       = "https://github.com/marrow16/gogol"
	helpLink        = "https://github.com/marrow16/gogol/blob/main/HELP_GUI.md"
	shortcutsHelp   = "https://github.com/marrow16/gogol/blob/main/cmd/gui/SHORTCUTS.md"
	metaRuleHelp    = "https://github.com/marrow16/gogol/blob/main/logic/meta/README.md"
	gridRecipesHelp = "https://github.com/marrow16/gogol/blob/main/recipes/README.md"
)

func newAboutPopout(p *menuPopup) *aboutPopout {
	result := &aboutPopout{
		parent: p,
	}
	result.links = map[*widget.Clickable]string{
		&result.linkRepo:        gogolRepo,
		&result.linkHelp:        helpLink,
		&result.linkShortcuts:   shortcutsHelp,
		&result.linkMetaRule:    metaRuleHelp,
		&result.linkGridRecipes: gridRecipesHelp,
	}
	return result
}

type aboutPopout struct {
	parent          *menuPopup
	linkRepo        widget.Clickable
	linkHelp        widget.Clickable
	linkShortcuts   widget.Clickable
	linkMetaRule    widget.Clickable
	linkGridRecipes widget.Clickable
	links           map[*widget.Clickable]string
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
		rigidLabel("Version: "+gogolGuiVersion, text.Middle, 0, minX),
		rigidSpacerVertical(m.Size.Y/2),
		rigidLabel("Author: Martin \"Marrow\" Rowlinson", text.Middle, 0, minX),
		rigidFixedWidth(linkLabel(&p.linkRepo, gogolRepo), minX, layout.Center),
		rigidSpacerVertical(m.Size.Y/2),
		rigidFixedWidth(linkLabel(&p.linkHelp, "General UI help"), minX, layout.Center),
		rigidFixedWidth(linkLabel(&p.linkShortcuts, "Shortcuts help"), minX, layout.Center),
		rigidFixedWidth(linkLabel(&p.linkMetaRule, "Meta Rules help"), minX, layout.Center),
		rigidFixedWidth(linkLabel(&p.linkGridRecipes, "Grid Recipes help"), minX, layout.Center),
	))
}

func (p *aboutPopout) hasFocus(gtx layout.Context) bool {
	return false
}

func (p *aboutPopout) reset() {
	// nothing to reset
}
