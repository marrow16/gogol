package widgets

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
)

const (
	gogolGuiVersion = "1.0.35"
	gogolRepo       = "https://github.com/marrow16/gogol"
	helpLink        = "https://github.com/marrow16/gogol/blob/main/HELP_GUI.md"
	shortcutsHelp   = "https://github.com/marrow16/gogol/blob/main/cmd/gui/SHORTCUTS.md"
	metaRuleHelp    = "https://github.com/marrow16/gogol/blob/main/logic/meta/README.md"
	gridRecipesHelp = "https://github.com/marrow16/gogol/blob/main/recipes/README.md"
)

func newAboutPopout(p *menuPopup, c *Core) *aboutPopout {
	return &aboutPopout{
		parent: p,
		core:   c,
	}
}

type aboutPopout struct {
	parent          *menuPopup
	core            *Core
	linkRepo        widget.Clickable
	linkHelp        widget.Clickable
	linkShortcuts   widget.Clickable
	linkMetaRule    widget.Clickable
	linkGridRecipes widget.Clickable
}

func (p *aboutPopout) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	m := measureText(gtx, theme, "M")
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
	return layout.Inset{Left: 8, Right: 8, Top: 8, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = minX
				gtx.Constraints.Max.X = minX
				lbl := material.Label(theme, theme.TextSize, "GoGoL")
				lbl.MaxLines = 1
				lbl.Alignment = text.Middle
				lbl.Font.Weight = font.Bold
				return lbl.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = minX
				gtx.Constraints.Max.X = minX
				lbl := material.Label(theme, theme.TextSize, "Version: "+gogolGuiVersion)
				lbl.MaxLines = 1
				lbl.Alignment = text.Middle
				return lbl.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Point{Y: m.Size.Y / 2}}
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = minX
				gtx.Constraints.Max.X = minX
				lbl := material.Label(theme, theme.TextSize, "Author: Martin \"Marrow\" Rowlinson")
				lbl.MaxLines = 1
				lbl.Alignment = text.Middle
				return lbl.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = minX
				gtx.Constraints.Max.X = minX
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return p.linkRepo.Layout(gtx, linkLabel(theme, gogolRepo))
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Point{Y: m.Size.Y / 2}}
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = minX
				gtx.Constraints.Max.X = minX
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return p.linkHelp.Layout(gtx, linkLabel(theme, "General UI help"))
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = minX
				gtx.Constraints.Max.X = minX
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return p.linkShortcuts.Layout(gtx, linkLabel(theme, "Shortcuts help"))
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = minX
				gtx.Constraints.Max.X = minX
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return p.linkMetaRule.Layout(gtx, linkLabel(theme, "Meta Rules help"))
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = minX
				gtx.Constraints.Max.X = minX
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return p.linkGridRecipes.Layout(gtx, linkLabel(theme, "Grid Recipes help"))
				})
			}),
		)
	})
}

func (p *aboutPopout) hasFocus(gtx layout.Context) bool {
	return false
}

func (p *aboutPopout) reset() {
	// nothing to reset
}
