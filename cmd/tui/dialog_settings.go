package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"errors"
	"github.com/marrow16/gogol/cmd/tui/layout"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"image/color"
	"io/fs"
	"os"
	"sort"
	"strconv"
	"strings"
)

type settingsDialog struct {
	m               *model
	tab             int
	currentForm     *layout.Form[*settingsDialog]
	clickPts        layout.ClickPoints[*settingsDialog]
	customRuleName  string
	exportResult    *exportResult
	importFrom      string
	importNoResize  bool
	importResult    *importResult
	sortedRuleNames []string
}

const (
	settingsTabGrid = iota
	settingsTabRule
	settingsTabExport
)

var settingsTabs = tabs{
	{"Grid", 0, settingsTabGrid},
	{"Rule", 0, settingsTabRule},
	{"Export/Import", 1, settingsTabExport},
}

func (d *settingsDialog) title() string {
	return "[settings]"
}

func (d *settingsDialog) render(rgn layout.Surface) *tea.Cursor {
	d.clickPts = layout.ClickPoints[*settingsDialog]{}
	// outer draw...
	rgn.FillWith(0, 0, rgn.Height(), rgn.Width(), '\u00A0', dialogBgStyle)
	rgn.BoxRounded(0, 0, rgn.Height(), rgn.Width(), dialogTextStyle)
	rgn.TextCenter(0, 0, rgn.Width(), "Grid Settings", dialogTextStyle)
	renderTabs(rgn, d.clickPts, settingsTabs, d.tab, func(t int) {
		d.tab = t
	})
	d.currentForm = nil
	var csr *tea.Cursor
	switch d.tab {
	case settingsTabGrid:
		d.currentForm = gridForm
	case settingsTabRule:
		d.currentForm = ruleForm
	case settingsTabExport:
		d.currentForm = exportForm
	}
	if d.currentForm != nil {
		csr = d.currentForm.Render(d, d.clickPts, rgn)
	}
	return csr
}

var gridForm = &layout.Form[*settingsDialog]{
	Style:        dialogTextStyle,
	FocusedStyle: dialogTabStyle,
	FormRows: layout.FormRows[*settingsDialog]{
		3: {
			2: {
				Item: layout.NewButton("Clear", func(d *settingsDialog) tea.Cmd {
					d.m.grid.Clear()
					return nil
				}),
			},
			10: {
				Item: layout.NewButton("Randomize", func(d *settingsDialog) tea.Cmd {
					d.m.grid.Randomize(d.m.random)
					return nil
				}),
			},
		},
		5: {
			2: {Item: "Randomization %:"},
			20: {
				Item: layout.NewNumberInput(4, 0, 100, func(d *settingsDialog) int {
					return d.m.random
				}, func(d *settingsDialog, value int) tea.Cmd {
					d.m.random = value
					d.m.prefs.Random = value
					return d.m.savePrefs()
				}),
			},
		},
		6: {
			2: {Item: "Step delay (ms):"},
			20: {
				Item: layout.NewNumberInput(4, 1, 2000, func(d *settingsDialog) int {
					return d.m.stepDelay
				}, func(d *settingsDialog, value int) tea.Cmd {
					d.m.stepDelay = value
					d.m.prefs.StepDelay = value
					return d.m.savePrefs()
				}),
			},
		},
		7: {
			2: {Item: "Step ahead size:"},
			20: {
				Item: layout.NewNumberInput(4, 1, 9999,
					func(d *settingsDialog) int {
						return d.m.stepAheadBy
					},
					func(d *settingsDialog, value int) tea.Cmd {
						d.m.stepAheadBy = value
						d.m.prefs.StepAheadBy = value
						return d.m.savePrefs()
					}),
			},
			30: {Item: "Snapshot before:"},
			47: {
				Item: layout.NewRadio([]string{"No", "Yes"},
					func(d *settingsDialog) int {
						if d.m.snapshotBefore {
							return 1
						}
						return 0
					},
					func(d *settingsDialog, value int) tea.Cmd {
						d.m.snapshotBefore = value != 0
						d.m.prefs.SnapshotBefore = d.m.snapshotBefore
						return d.m.savePrefs()
					}),
			},
		},
		9: {
			2: {Item: "  Wrapping mode:"},
			19: {
				Item: layout.NewRadio([]string{"None", "Horizontal", "Vertical", "Toroidal"}, func(d *settingsDialog) int {
					return int(d.m.grid.WrapMode)
				}, func(d *settingsDialog, value int) tea.Cmd {
					d.m.grid.WrapMode = logic.WrapMode(value)
					return nil
				}),
			},
		},
		11: {
			2: {Item: " Boundary cells:"},
			19: {
				Item: layout.NewRadio([]string{"Dead", "Alive"}, func(d *settingsDialog) int {
					return int(d.m.grid.BoundaryMode)
				}, func(d *settingsDialog, value int) tea.Cmd {
					d.m.grid.BoundaryMode = logic.BoundaryMode(value)
					return nil
				}),
			},
		},
		14: {
			2: {Item: "Height:       Width:"},
			10: {
				Item: layout.NewNumberInput(4, 2, 9999, func(d *settingsDialog) int {
					return d.m.gridHeight
				}, func(d *settingsDialog, value int) tea.Cmd {
					d.m.gridHeight = value
					return nil
				}),
			},
			23: {
				Item: layout.NewNumberInput(4, 2, 9999, func(d *settingsDialog) int {
					return d.m.gridWidth
				}, func(d *settingsDialog, value int) tea.Cmd {
					d.m.gridWidth = value
					return nil
				}),
			},
			30: {
				Item: layout.NewButton("Resize", func(d *settingsDialog) tea.Cmd {
					if d.m.gridHeight == d.m.grid.Height && d.m.gridWidth == d.m.grid.Width {
						// no resize needed
						return nil
					}
					return func() tea.Msg {
						if grid, err := logic.NewGrid(d.m.gridHeight, d.m.gridWidth, d.m.grid.WrapMode, d.m.grid.BoundaryMode); err == nil {
							grid.Rule = d.m.grid.Rule
							return gridResizeResult{
								surface: newGridSurface(grid, d.m.cellStyle),
								grid:    grid,
							}
						}
						return nil
					}
				}),
			},
			39: {
				Item: layout.NewButton("Fit Screen", func(d *settingsDialog) tea.Cmd {
					if d.m.gridHeight == d.m.height*2 && d.m.gridWidth == d.m.width*2 {
						// no resize needed
						return nil
					}
					d.m.gridHeight, d.m.gridWidth = d.m.height*2, d.m.width*2
					return func() tea.Msg {
						if grid, err := logic.NewGrid(d.m.gridHeight, d.m.gridWidth, d.m.grid.WrapMode, d.m.grid.BoundaryMode); err == nil {
							grid.Rule = d.m.grid.Rule
							return gridResizeResult{
								surface: newGridSurface(grid, d.m.cellStyle),
								grid:    grid,
							}
						}
						return nil
					}
				}),
			},
		},
		17: {
			2: {
				Item: "Foreground  R:      G:      B:",
			},
			17: {
				Item: layout.NewNumberInput(3, 0, 255, func(d *settingsDialog) int {
					v, _, _ := rgb(d.m.cellStyle.GetForeground())
					return v
				}, func(d *settingsDialog, value int) tea.Cmd {
					_, g, b := rgb(d.m.cellStyle.GetForeground())
					d.adjustCellColor(true, value, g, b)
					d.m.prefs.setCellStyle(d.m.cellStyle)
					return d.m.savePrefs()
				}),
			},
			24: {
				Item: layout.NewNumberInput(3, 0, 255, func(d *settingsDialog) int {
					_, v, _ := rgb(d.m.cellStyle.GetForeground())
					return v
				}, func(d *settingsDialog, value int) tea.Cmd {
					r, _, b := rgb(d.m.cellStyle.GetForeground())
					d.adjustCellColor(true, r, value, b)
					d.m.prefs.setCellStyle(d.m.cellStyle)
					return d.m.savePrefs()

				}),
			},
			33: {
				Item: layout.NewNumberInput(3, 0, 255, func(d *settingsDialog) int {
					_, _, v := rgb(d.m.cellStyle.GetForeground())
					return v
				}, func(d *settingsDialog, value int) tea.Cmd {
					r, g, _ := rgb(d.m.cellStyle.GetForeground())
					d.adjustCellColor(true, r, g, value)
					d.m.prefs.setCellStyle(d.m.cellStyle)
					return d.m.savePrefs()

				}),
			},
		},
		18: {
			2: {
				Item: "Background  R:      G:      B:",
			},
			17: {
				Item: layout.NewNumberInput(3, 0, 255, func(d *settingsDialog) int {
					v, _, _ := rgb(d.m.cellStyle.GetBackground())
					return v
				}, func(d *settingsDialog, value int) tea.Cmd {
					_, g, b := rgb(d.m.cellStyle.GetBackground())
					d.adjustCellColor(false, value, g, b)
					d.m.prefs.setCellStyle(d.m.cellStyle)
					return d.m.savePrefs()

				}),
			},
			24: {
				Item: layout.NewNumberInput(3, 0, 255, func(d *settingsDialog) int {
					_, v, _ := rgb(d.m.cellStyle.GetBackground())
					return v
				}, func(d *settingsDialog, value int) tea.Cmd {
					r, _, b := rgb(d.m.cellStyle.GetBackground())
					d.adjustCellColor(false, r, value, b)
					d.m.prefs.setCellStyle(d.m.cellStyle)
					return d.m.savePrefs()

				}),
			},
			33: {
				Item: layout.NewNumberInput(3, 0, 255, func(d *settingsDialog) int {
					_, _, v := rgb(d.m.cellStyle.GetBackground())
					return v
				}, func(d *settingsDialog, value int) tea.Cmd {
					r, g, _ := rgb(d.m.cellStyle.GetBackground())
					d.adjustCellColor(false, r, g, value)
					d.m.prefs.setCellStyle(d.m.cellStyle)
					return d.m.savePrefs()

				}),
			},
		},
	},
}

var ruleForm = &layout.Form[*settingsDialog]{
	Style:        dialogTextStyle,
	FocusedStyle: dialogTabStyle,
	FormRows: layout.FormRows[*settingsDialog]{
		3: {
			2: {Item: "       Name:"},
			15: {
				Item: layout.NewDropdownSelect(func(d *settingsDialog) []string {
					return d.supplyRuleNames()
				}, -1, func(d *settingsDialog) string {
					return d.m.grid.Rule.Name()
				}, func(d *settingsDialog, value string) tea.Cmd {
					if nr, ok := logic.Rules[value]; ok {
						d.m.grid.Rule = nr
						d.m.prefs.setRule(d.m.grid.Rule)
						return d.m.savePrefs()
					} else if strings.HasPrefix(value, "Custom ") {
						d.customRuleName = ""
						if nr, err := logic.NewRuleRle("", strings.TrimPrefix(value, "Custom ")); err == nil {
							d.m.grid.Rule = nr
							d.m.prefs.setRule(d.m.grid.Rule)
							return d.m.savePrefs()
						}
					}
					return nil
				}),
			},
		},
		5: {
			2: {Item: "       Rule:"},
			15: {
				Item: layout.NewTextInput(21, "BbSs/012345678", func(d *settingsDialog) string {
					return d.m.grid.Rule.Rle()
				}, func(d *settingsDialog, value string) tea.Cmd {
					if nr, err := logic.NewRuleRle("", value); err == nil {
						d.customRuleName = ""
						d.m.grid.Rule = nr
						d.m.prefs.setRule(d.m.grid.Rule)
						return d.m.savePrefs()
					}
					return nil
				}),
			},
		},
		7: {
			2: {Item: "Permutation:"},
			15: {Item: layout.NewNumberInput(6, 0, (1<<18)-1, func(d *settingsDialog) int {
				return d.m.grid.Rule.Permutation()
			}, func(d *settingsDialog, value int) tea.Cmd {
				if nr, err := logic.NewRuleFromPermutation(value); err == nil {
					d.customRuleName = ""
					d.m.grid.Rule = nr
					d.m.prefs.setRule(d.m.grid.Rule)
					return d.m.savePrefs()
				}
				return nil
			})},
		},
		17: {
			6: {
				Item: "As name:",
				Condition: func(d *settingsDialog) bool {
					return d.m.grid.Rule.IsCustom()
				},
			},
			15: {
				Item: layout.NewTextInput(25, "", func(d *settingsDialog) string {
					return d.customRuleName
				}, func(d *settingsDialog, value string) tea.Cmd {
					d.customRuleName = value
					return nil
				}),
				Condition: func(d *settingsDialog) bool {
					return d.m.grid.Rule.IsCustom()
				},
			},
			42: {
				Item: layout.NewButton("Save", func(d *settingsDialog) tea.Cmd {
					if d.customRuleName != "" {
						d.m.prefs.addRule(d.customRuleName, d.m.grid.Rule.Rle())
						logic.AddRule(d.customRuleName, d.m.grid.Rule)
						return d.m.savePrefs()
					}
					return nil
				}),
				Condition: func(d *settingsDialog) bool {
					return d.m.grid.Rule.IsCustom() && len(d.customRuleName) > 0
				},
			},
		},
	},
}

func (d *settingsDialog) supplyRuleNames() []string {
	if len(d.sortedRuleNames) != len(logic.Rules) {
		d.sortedRuleNames = make([]string, 0, len(logic.Rules))
		for name := range logic.Rules {
			d.sortedRuleNames = append(d.sortedRuleNames, name)
		}
		sort.Strings(d.sortedRuleNames)
	}
	return d.sortedRuleNames
}

func (d *settingsDialog) adjustRule(born bool, add bool, digit string) {
	bw, sw := d.m.grid.Rule.BornWith(), d.m.grid.Rule.SurvivesWith()
	if born {
		if add {
			bw += digit
		} else {
			bw = strings.ReplaceAll(bw, digit, "")
		}
	} else if add {
		sw += digit
	} else {
		sw = strings.ReplaceAll(sw, digit, "")
	}
	if nr, err := logic.NewRuleRle("", "B"+bw+"/S"+sw); err == nil {
		d.m.grid.Rule = nr
	}
}

func rgb(c color.Color) (int, int, int) {
	r, g, b, _ := c.RGBA()
	return int(r) / 256, int(g) / 256, int(b) / 256
}

func (d *settingsDialog) adjustCellColor(fg bool, r, g, b int) {
	if fg {
		d.m.cellStyle = d.m.cellStyle.Foreground(lipgloss.RGBColor{R: clampRGB(r), G: clampRGB(g), B: clampRGB(b)})
	} else {
		d.m.cellStyle = d.m.cellStyle.Background(lipgloss.RGBColor{R: clampRGB(r), G: clampRGB(g), B: clampRGB(b)})
	}
	d.m.gridSurface.ClearStyle(0, 0, -1, -1, &d.m.cellStyle)
}

func clampRGB(c int) uint8 {
	if c < 0 {
		return 0
	} else if c > 255 {
		return 255
	}
	return uint8(c)
}

var exportForm = &layout.Form[*settingsDialog]{
	Style:        dialogTextStyle,
	FocusedStyle: dialogTabStyle,
	FormRows: layout.FormRows[*settingsDialog]{
		4: {
			25: {
				Item: layout.NewButton("Export Grid", func(d *settingsDialog) tea.Cmd {
					d.exportResult = nil
					return func() tea.Msg {
						fn, err := d.m.exportGrid()
						return exportResult{
							filename: fn,
							error:    err,
						}
					}
				}),
			},
		},
		6: {
			1: {
				Item: func(d *settingsDialog) any {
					if d.exportResult != nil {
						if d.exportResult.error != nil {
							return d.exportResult.error.Error()
						}
						return "Saved to " + d.exportResult.filename
					}
					return ""
				},
				Alignment: layout.AlignCenter,
				Condition: func(d *settingsDialog) bool {
					return d.exportResult != nil
				},
			},
		},
		8: {
			1: {Item: strings.Repeat("─", 58)},
		},
		10: {
			2: {Item: "Import from:"},
			15: {
				Item: layout.NewTextInput(43, "",
					func(d *settingsDialog) string {
						return d.importFrom
					},
					func(d *settingsDialog, value string) tea.Cmd {
						d.importFrom = value
						return nil
					}, true, true),
			},
		},
		11: {
			7: {Item: "Resize:"},
			15: {
				Item: layout.NewRadio([]string{"Yes", "No"},
					func(d *settingsDialog) int {
						if d.importNoResize {
							return 1
						}
						return 0
					},
					func(d *settingsDialog, value int) tea.Cmd {
						d.importNoResize = value != 0
						return nil
					}),
			},
		},
		13: {
			25: {
				Item: layout.NewButton("Import Grid", func(d *settingsDialog) tea.Cmd {
					d.importResult = nil
					if len(d.importFrom) == 0 {
						return nil
					}
					return func() tea.Msg {
						f, err := os.Open(d.importFrom)
						if err != nil {
							if errors.Is(err, fs.ErrNotExist) {
								return importResult{
									error: errors.New("File does not exist"),
								}
							} else {
								return importResult{error: err}
							}
						}
						defer func() {
							_ = f.Close()
						}()
						if p, err := patterns.NewPatternFromRle(f); err == nil {
							// parse wrap and boundary modes...
							wrap := d.m.grid.WrapMode
							boundary := d.m.grid.BoundaryMode
							step := uint64(0)
							for _, c := range p.Comments {
								if after, ok := strings.CutPrefix(c, "Wrap mode: "); ok {
									wrap = logic.WrapModeFromString(after, wrap)
								} else if after, ok := strings.CutPrefix(c, "Boundary mode: "); ok {
									boundary = logic.BoundaryModeFromString(after, boundary)
								} else if after, ok := strings.CutPrefix(c, "Step: "); ok {
									if n, err := strconv.ParseUint(after, 10, 64); err == nil {
										step = uint64(n)
									}
								}
							}
							noResize := d.importNoResize || (p.Height == d.m.grid.Height && p.Width == d.m.grid.Width)
							if noResize {
								// if no resize, draw it now...
								d.importResult = nil
								d.m.grid.Clear()
								p.Draw(d.m.grid, 0, 0, patterns.Rotate0)
								d.m.grid.StepCount.Store(step)
								d.m.grid.Rule = p.Rule
								d.m.grid.BoundaryMode = boundary
								d.m.grid.WrapMode = wrap
								d.m.prefs.setRule(d.m.grid.Rule)
								d.m.prefs.setWrapMode(wrap)
								d.m.prefs.setBoundaryMode(boundary)
								return d.m.savePrefs()
							} else if grid, err := logic.NewGrid(p.Height, p.Width, wrap, boundary); err == nil {
								d.importResult = nil
								grid.Rule = p.Rule
								p.Draw(grid, 0, 0, patterns.Rotate0)
								d.m.grid.StepCount.Store(step)
								d.m.prefs.setRule(grid.Rule)
								d.m.prefs.setWrapMode(wrap)
								d.m.prefs.setBoundaryMode(boundary)
								d.m.gridHeight, d.m.gridWidth = grid.Height, grid.Width
								d.m.prefs.Height, d.m.prefs.Width = grid.Height, grid.Width
								gridForm.Reset(d)
								return gridResizeResult{
									surface:     newGridSurface(grid, d.m.cellStyle),
									grid:        grid,
									noRandomize: true,
								}
							} else {
								return importResult{error: err}
							}
						} else {
							return importResult{error: err}
						}
					}
				}),
			},
		},
		15: {
			1: {
				Item: func(d *settingsDialog) any {
					if d.importResult != nil && d.importResult.error != nil {
						return d.importResult.error.Error()
					}
					return ""
				},
				Alignment: layout.AlignCenter,
				Condition: func(d *settingsDialog) bool {
					return d.importResult != nil
				},
			},
		},
	},
}

type exportResult struct {
	error    error
	filename string
}

type importResult struct {
	error  error
	resize bool
}

func (d *settingsDialog) update(msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		switch mt.String() {
		case "ctrl+c":
			return tea.Quit
		case "esc":
			d.m.stopped()
		case "ctrl+g":
			d.tab = settingsTabGrid
		case "ctrl+r":
			d.tab = settingsTabRule
		case "ctrl+x":
			d.tab = settingsTabExport
		default:
			if d.currentForm != nil {
				return d.currentForm.Update(d, msg)
			}
		}
	case tea.MouseClickMsg:
		if fn, ok := d.clickPts[layout.ClickPoint{mt.Y, mt.X}]; ok {
			return fn(d)
		} else if d.currentForm != nil {
			return d.currentForm.Update(d, msg)
		}
	case tea.MouseWheelMsg:
		if d.currentForm != nil {
			return d.currentForm.Update(d, msg)
		}
	case tea.PasteMsg:
		if d.currentForm != nil {
			if cmd := d.currentForm.Update(d, mt); cmd != nil {
				return cmd
			}
		}
	case exportResult:
		d.exportResult = &mt
	case importResult:
		d.importResult = &mt
	default:
		if d.currentForm != nil {
			return d.currentForm.Update(d, msg)
		}
	}
	return nil
}
