package widgets

import (
	"errors"
	"gioui.org/layout"
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
		btnPath:    newPathButton(),
		btnRun:     newButton("Run"),
		btnSaveRle: newButton("Save as RLE"),
	}
	result.chooser = newChooser[string](38,
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
			window.Invalidate()
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

func (p *gridRecipesPopout) layout(gtx layout.Context) layout.Dimensions {
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
	return popoutLayout(gtx, func(gtx layout.Context) layout.Dimensions {
		dims := flexVertical(10,
			rigid(flexHorizontal(0,
				rigid(p.chooser.layout),
				rigid(p.btnPath.Layout),
			)),
			rigid(flexHorizontal(20,
				rigid(p.btnRun.Layout),
				rigid(p.btnSaveRle.Layout),
				rigid(errorLabel(p.error)),
				conditionalRigid(p.error == nil && p.chooser.currentItem() != nil, label("(Press "+modKeyName+"G to run)"), nil),
			)),
		)(gtx)
		p.chooser.layoutDropdown(gtx)
		return dims
	})
}

func (p *gridRecipesPopout) hasFocus(gtx layout.Context) bool {
	return p.chooser.isFocused(gtx) || p.btnRun.isFocused(gtx) || p.btnSaveRle.isFocused(gtx) || p.btnPath.isFocused(gtx)
}
