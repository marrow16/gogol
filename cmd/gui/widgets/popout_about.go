package widgets

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
)

const (
	gogolGuiVersion = "1.0.36"
	gogolRepo       = "https://github.com/marrow16/gogol"
	helpLink        = "https://github.com/marrow16/gogol/blob/main/HELP_GUI.md"
	shortcutsHelp   = "https://github.com/marrow16/gogol/blob/main/cmd/gui/SHORTCUTS.md"
	metaRuleHelp    = "https://github.com/marrow16/gogol/blob/main/logic/meta/README.md"
	gridRecipesHelp = "https://github.com/marrow16/gogol/blob/main/recipes/README.md"
)

func newAboutPopout(p *menuPopup) *aboutPopout {
	return &aboutPopout{
		parent: p,
	}
}

type aboutPopout struct {
	parent          *menuPopup
	linkRepo        widget.Clickable
	linkHelp        widget.Clickable
	linkShortcuts   widget.Clickable
	linkMetaRule    widget.Clickable
	linkGridRecipes widget.Clickable
}

func (p *aboutPopout) layout(gtx layout.Context) layout.Dimensions {
	m := measureText(gtx, "M")
	minX := m.Size.X * 30
	if p.linkRepo.Clicked(gtx) {
		_ = openURL(gogolRepo)
	}
	if p.linkHelp.Clicked(gtx) {
		_ = openURL(helpLink)
	}
	if p.linkShortcuts.Clicked(gtx) {
		_ = openURL(shortcutsHelp)
	}
	if p.linkMetaRule.Clicked(gtx) {
		_ = openURL(metaRuleHelp)
	}
	if p.linkGridRecipes.Clicked(gtx) {
		_ = openURL(gridRecipesHelp)
	}
	return layout.Inset{Left: 8, Right: 8, Top: 8, Bottom: 8}.Layout(gtx, flexVertical(0,
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
