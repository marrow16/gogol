package main

import (
	"bytes"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"encoding/json"
	"errors"
	"github.com/marrow16/gogol/cmd/tui/layout"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/recipes"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	recipesTabSelect = iota
	recipesTabLoad
)

var recipesTabs = tabs{
	{"Select", 0, recipesTabSelect},
	{"Load", 1, recipesTabLoad},
}

type recipesDialog struct {
	m            *model
	tab          int
	currentForm  *layout.Form[*recipesDialog]
	clickPts     layout.ClickPoints[*recipesDialog]
	loadFrom     string
	loadResult   *recipeAddResult
	selected     string
	jsonYOffset  int
	recipeErrors map[string]error
}

func (d *recipesDialog) title() string {
	return "[recipes]"
}

func (r *recipesDialog) render(rgn layout.Surface) *tea.Cursor {
	if r.recipeErrors == nil {
		r.recipeErrors = make(map[string]error)
	}
	r.clickPts = layout.ClickPoints[*recipesDialog]{}
	// outer draw...
	rgn.FillWith(0, 0, rgn.Height(), rgn.Width(), '\u00A0', dialogBgStyle)
	rgn.BoxRounded(0, 0, rgn.Height(), rgn.Width(), dialogTextStyle)
	rgn.TextCenter(0, 0, rgn.Width(), "Grid Recipes", dialogTextStyle)
	// tabs...
	if r.tab == recipesTabSelect && len(r.m.prefs.Recipes) == 0 {
		r.tab = recipesTabLoad
	}
	renderTabs(rgn, r.clickPts, recipesTabs, r.tab, func(t int) {
		r.tab = t
	})
	r.currentForm = nil
	switch r.tab {
	case recipesTabSelect:
		if r.selected == "" && len(r.m.prefs.Recipes) > 0 {
			rs := slices.Clone(r.m.prefs.Recipes)
			slices.Sort(rs)
			r.selected = rs[0]
		}
		r.currentForm = recipeSelectForm
	case recipesTabLoad:
		r.currentForm = recipeLoadForm
	}
	var csr *tea.Cursor
	if r.currentForm != nil {
		csr = r.currentForm.Render(r, r.clickPts, rgn)
	}
	return csr
}

func (r *recipesDialog) runSelected() (err error) {
	if r.recipeErrors == nil {
		r.recipeErrors = make(map[string]error)
	}
	if r.selected != "" {
		var recipe *recipes.Recipe
		if recipe, err = recipes.Load(r.selected); err == nil {
			_, _, err = recipe.Run(r.m.grid, false)
		}
		if err != nil {
			r.recipeErrors[r.selected] = err
		}
	}
	return err
}

var recipeSelectForm = &layout.Form[*recipesDialog]{
	Style:        dialogTextStyle,
	FocusedStyle: dialogTabStyle,
	FormRows: layout.FormRows[*recipesDialog]{
		3: {
			2: {Item: "Recipe:"},
			10: {
				Item: layout.NewDropdownSelect(
					func(r *recipesDialog) []string {
						result := slices.Clone(r.m.prefs.Recipes)
						slices.Sort(result)
						return result
					}, -1,
					func(r *recipesDialog) string {
						return r.selected
					},
					func(r *recipesDialog, value string) tea.Cmd {
						r.jsonYOffset = 0
						r.selected = value
						return nil
					}, true),
			},
		},
		4: {
			1: {
				Item: &recipePreview[*recipesDialog]{},
			},
		},
		17: {
			45: {
				Item: layout.NewButton("Run",
					func(r *recipesDialog) tea.Cmd {
						if r.selected != "" {
							s := r.selected
							delete(r.recipeErrors, s)
							return func() tea.Msg {
								recipe, err := recipes.Load(s)
								return recipeLoadResult{
									filename: s,
									recipe:   recipe,
									error:    err,
								}
							}
						}
						return nil
					}),
			},
			50: {
				Item: layout.NewButton("Save rle",
					func(r *recipesDialog) tea.Cmd {
						if r.selected != "" {
							s := r.selected
							delete(r.recipeErrors, s)
							return func() tea.Msg {
								recipe, err := recipes.Load(s)
								return recipeLoadResult{
									filename: s,
									recipe:   recipe,
									error:    err,
									saveRle:  true,
								}
							}
						}
						return nil
					}),
			},
		},
	},
}

type recipePreview[T any] struct{}

func (p *recipePreview[T]) Update(parent T, msg tea.Msg, focused bool) tea.Cmd {
	if focused {
		r := asRecipes(parent)
		switch mt := msg.(type) {
		case tea.KeyPressMsg:
			switch mt.String() {
			case "up":
				if r.jsonYOffset > 0 {
					r.jsonYOffset--
				}
			case "down":
				r.jsonYOffset++
			}
		}
	}
	return nil
}

func (p *recipePreview[T]) Reset(parent T) {}

func (p *recipePreview[T]) Render(parent T, form *layout.Form[T], inputNo int, sf layout.Surface, clickPts layout.ClickPoints[T], row, col int, focused bool, style lipgloss.Style, focusedStyle lipgloss.Style) *tea.Cursor {
	rgn := sf.Region(row, col, 12, sf.Width()-2)
	r := asRecipes(parent)
	s := dialogTextStyle
	if focused {
		s = dialogTabStyle
	}
	if r.selected == "" {
		rgn.TextCenter(7, 0, rgn.Width(), "No preview/info available", s)
	} else if err, ok := r.recipeErrors[r.selected]; ok {
		rgn.TextCenter(1, 0, rgn.Width(), "Recipe run error", s)
		rgn.TextWrapped(3, 1, rgn.Width()-2, err.Error(), s)
	} else {
		f, err := os.Open(r.selected)
		if err != nil {
			if os.IsNotExist(err) {
				rgn.TextCenter(7, 0, rgn.Width(), "Recipe file not found", dialogTextStyle)
			} else {
				rgn.TextCenter(7, 0, rgn.Width(), "Error loading recipe file", dialogTextStyle)
			}
			return nil
		}
		defer func() {
			_ = f.Close()
		}()
		var m map[string]any
		if err = json.NewDecoder(f).Decode(&m); err != nil {
			rgn.TextCenter(7, 0, rgn.Width(), "Error parsing recipe file", dialogTextStyle)
			return nil
		}
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "  ")
		err = enc.Encode(m)
		if err != nil {
			rgn.TextCenter(7, 0, rgn.Width(), "Unable to display recipe json", dialogTextStyle)
			return nil
		}
		rgn.Fill(0, 0, rgn.Height(), rgn.Width(), s)
		lines := strings.Split(buf.String(), "\n")
		for i := 0; i < len(lines); i++ {
			if lines[i] != "" {
				rgn.TextFixed(i-r.jsonYOffset, 0, rgn.Width(), lines[i], s)
			}
		}
	}
	return nil
}

func asRecipes(parent any) *recipesDialog {
	return parent.(*recipesDialog)
}

var recipeLoadForm = &layout.Form[*recipesDialog]{
	Style:        dialogTextStyle,
	FocusedStyle: dialogTabStyle,
	FormRows: layout.FormRows[*recipesDialog]{
		3: {
			2: {Item: "From:"},
			8: {
				Item: layout.NewTextInput(50, "",
					func(r *recipesDialog) string {
						return r.loadFrom
					},
					func(r *recipesDialog, value string) tea.Cmd {
						r.loadFrom = value
						return nil
					}, true, true),
			},
		},
		5: {
			0: {
				Item: layout.NewButton("Load", func(r *recipesDialog) tea.Cmd {
					r.loadResult = nil
					return func() tea.Msg {
						if f, err := os.Open(r.loadFrom); err == nil {
							defer func() {
								_ = f.Close()
							}()
							var m map[string]any
							if err = json.NewDecoder(f).Decode(&m); err == nil {
								return recipeAddResult{filename: r.loadFrom}
							} else {
								return recipeAddResult{error: errors.New("Invalid json")}
							}
						} else if os.IsNotExist(err) {
							return recipeAddResult{error: errors.New("File does not exist")}
						} else {
							return recipeAddResult{error: err}
						}
					}
				}).Align(layout.AlignCenter),
				Condition: func(r *recipesDialog) bool {
					return r.loadFrom != ""
				},
			},
		},
		7: {
			1: {
				Item: func(r *recipesDialog) any {
					if r.loadResult != nil && r.loadResult.error != nil {
						return "Error: " + r.loadResult.error.Error()
					}
					return ""
				},
				Alignment: layout.AlignCenter,
				Condition: func(r *recipesDialog) bool {
					return r.loadResult != nil && r.loadResult.error != nil
				},
			},
		},
	},
}

type recipeAddResult struct {
	filename string
	error    error
}

type recipeLoadResult struct {
	filename string
	recipe   *recipes.Recipe
	error    error
	saveRle  bool
}

type recipeRunResult struct {
	filename string
	resized  bool
	grid     *logic.Grid
	error    error
}

func (r *recipesDialog) update(msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		switch mt.String() {
		case "ctrl+c":
			return tea.Quit
		case "esc":
			r.m.stopped()
		case "ctrl+s":
			r.tab = recipesTabSelect
		case "ctrl+o":
			r.tab = recipesTabLoad
		default:
			if r.currentForm != nil {
				return r.currentForm.Update(r, msg)
			}
		}
	case recipeAddResult:
		if mt.error == nil {
			r.tab = recipesTabSelect
			r.selected, _ = filepath.Abs(mt.filename)
			r.m.prefs.addRecipe(r.selected)
			return func() tea.Msg {
				return r.m.savePrefs()
			}
		} else {
			r.loadResult = &mt
		}
	case recipeLoadResult:
		if mt.error != nil {
			r.recipeErrors[mt.filename] = mt.error
		} else if mt.saveRle {
			return func() tea.Msg {
				err := mt.recipe.SaveAsRle(r.m.grid, mt.filename, r.m.prefs.Originator)
				return recipeRunResult{
					filename: mt.filename,
					resized:  false,
					error:    err,
				}
			}
		} else {
			return func() tea.Msg {
				ng, resized, err := mt.recipe.Run(r.m.grid, true)
				return recipeRunResult{
					filename: mt.filename,
					resized:  resized,
					grid:     ng,
					error:    err,
				}
			}
		}
	case recipeRunResult:
		if mt.error != nil {
			r.recipeErrors[mt.filename] = mt.error
		} else {
			r.m.stopped()
			if mt.resized {
				grid := mt.grid
				return func() tea.Msg {
					return gridResizeResult{
						surface:     newGridSurface(grid, r.m.cellStyle),
						grid:        grid,
						noRandomize: true,
					}
				}
			}
		}
	case tea.MouseClickMsg:
		if fn, ok := r.clickPts[layout.ClickPoint{mt.Y, mt.X}]; ok {
			return fn(r)
		} else if r.currentForm != nil {
			return r.currentForm.Update(r, msg)
		}
	case tea.MouseWheelMsg:
		if r.currentForm != nil {
			return r.currentForm.Update(r, msg)
		}
	case tea.PasteMsg:
		if r.currentForm != nil {
			if cmd := r.currentForm.Update(r, mt); cmd != nil {
				return cmd
			}
		}
	default:
		if r.currentForm != nil {
			return r.currentForm.Update(r, msg)
		}
	}
	return nil
}
