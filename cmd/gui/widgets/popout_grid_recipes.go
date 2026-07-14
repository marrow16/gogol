package widgets

import (
	"errors"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/recipes"
	"slices"
	"sort"
	"strings"
)

type gridRecipesPopout struct {
	parent     *menuPopup
	core       *Core
	chooser    *chooser[string]
	btnPath    *pathButton
	btnRun     *button
	btnSaveRle *button
	error      error
}

func newGridRecipesPopout(p *menuPopup, c *Core) *gridRecipesPopout {
	result := &gridRecipesPopout{
		parent:     p,
		core:       c,
		btnPath:    newPathButton(c.theme),
		btnRun:     newButton(c.theme, "Run"),
		btnSaveRle: newButton(c.theme, "Save as RLE"),
	}
	result.chooser = newChooser[string](c.theme, 38,
		result.sortedRecipes(),
		result.recipeSelected,
		func(recipe string) string {
			return recipe
		},
	)
	result.chooser.middleEllipsis = true
	result.chooser.onSubmit(result.submitFilename)
	c.gridRecipes = result
	return result
}

func (p *gridRecipesPopout) sortedRecipes() []string {
	result := slices.Clone(p.core.settings.Recipes)
	sort.Strings(result)
	return result
}

func (p *gridRecipesPopout) recipeSelected(recipe *string) {
	p.error = nil
}

func (p *gridRecipesPopout) submitFilename(recipe string) {
	if recipe == "" {
		p.error = nil
		return
	}
	_, p.error = recipes.Load(recipe)
	if p.error == nil {
		p.core.settings.AddRecipe(recipe)
		p.chooser.resetItems(p.sortedRecipes())
	}
}

func (p *gridRecipesPopout) reset() {
}

func (p *gridRecipesPopout) getCurrentRecipe() (*recipes.Recipe, string) {
	p.error = nil
	filename := p.chooser.currentItem()
	if filename == nil {
		p.error = errors.New("No recipe selected")
		return nil, ""
	}
	var recipe *recipes.Recipe
	if recipe, p.error = recipes.Load(*filename); p.error == nil {
		return recipe, *filename
	}
	return nil, ""
}

func (p *gridRecipesPopout) runRecipe() {
	if recipe, _ := p.getCurrentRecipe(); recipe != nil {
		grid, resized, err := recipe.Run(p.core.gridHolder.grid, true)
		if err != nil {
			p.error = err
			return
		}
		if resized {
			p.core.settings.Height, p.core.settings.Width, p.core.settings.WrapMode, p.core.settings.BoundaryMode = grid.Height, grid.Width, grid.WrapMode, grid.BoundaryMode
			p.core.gridHolder.replaceGrid(grid)
			p.core.resetInstrumentation()
			p.core.window.Invalidate()
		} else {
			p.core.resetInstrumentation()
		}
	}
}

func (p *gridRecipesPopout) saveRecipeRle() {
	if recipe, filename := p.getCurrentRecipe(); recipe != nil {
		p.error = recipe.SaveAsRle(p.core.gridHolder.grid, filename, p.core.settings.Originator)
	}
}

func (p *gridRecipesPopout) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	if p.btnPath.Clicked(gtx) {
		filePicker(func(filename string) {
			path := strings.TrimSpace(string(filename))
			p.core.settings.AddRecipe(path)
			p.chooser.resetItems(p.sortedRecipes())
			p.chooser.setText(path)
		})
	}
	if p.btnRun.Clicked(gtx) {
		p.runRecipe()
	}
	if p.btnSaveRle.Clicked(gtx) {
		p.saveRecipeRle()
	}
	return layout.Inset{
		Left: unit.Dp(8), Right: unit.Dp(8),
		Top: unit.Dp(8), Bottom: unit.Dp(8),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		var chooserDims layout.Dimensions
		dims := layout.Flex{Axis: layout.Vertical, Gap: 10}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						chooserDims = p.chooser.layout(gtx)
						return chooserDims
					}),
					layout.Rigid(p.btnPath.Layout),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Max.X = p.chooser.dims.Size.X
				return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
					layout.Rigid(p.btnRun.Layout),
					layout.Rigid(p.btnSaveRle.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						switch {
						case p.error != nil:
							return errorLabel(theme, p.error)(gtx)
						case p.chooser.currentItem() != nil:
							return label(theme, "(Press "+modKeyName+"G to run)")(gtx)
						default:
							return layout.Dimensions{}
						}
					}),
				)
			}),
		)
		p.chooser.layoutDropdown(gtx, chooserDims)
		return dims
	})
}

func (p *gridRecipesPopout) hasFocus(gtx layout.Context) bool {
	return p.chooser.isFocused(gtx) || p.btnRun.isFocused(gtx) || p.btnSaveRle.isFocused(gtx) || p.btnPath.isFocused(gtx)
}
