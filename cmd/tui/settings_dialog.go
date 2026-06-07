package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"errors"
	"fmt"
	"github.com/marrow16/gogol/cmd/tui/layout"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"image/color"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type settings struct {
	m                  *model
	tab                int
	currentForm        *layout.Form[*settings]
	clickPts           layout.ClickPoints[*settings]
	absTop, absLeft    int
	width, height      int
	currentPattern     *patterns.Pattern
	patternPlaceY      int
	patternPlaceX      int
	patternRotate      patterns.Rotation
	patternOffY        int
	patternOffX        int
	patternInfo        bool
	loadFrom           string
	loadPatternsResult *loadPatternsResult
	customRuleName     string
}

var (
	settingsBgStyle   = lipgloss.NewStyle().Background(lipgloss.Color("#eeeeee"))
	settingsTextStyle = lipgloss.NewStyle().Background(lipgloss.Color("#eeeeee")).
				Foreground(lipgloss.Color("#6680e6"))
	settingsTextUlStyle = lipgloss.NewStyle().Background(lipgloss.Color("#eeeeee")).
				Foreground(lipgloss.Color("#6680e6")).
				Underline(true)
	settingsTabStyle     = lipgloss.NewStyle().Background(lipgloss.Color("#6680e6")).Foreground(lipgloss.Color("#ffffff"))
	settingsPreviewStyle = lipgloss.NewStyle().Background(lipgloss.Color("#ffffff")).
				Foreground(lipgloss.Color("#000000"))
	settingsPreviewFocusedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#ffffff")).
					Foreground(lipgloss.Color("#6680e6"))
)

func (s *settings) render(rgn layout.Surface) *tea.Cursor {
	s.clickPts = layout.ClickPoints[*settings]{}
	s.absTop, s.absLeft, s.width, s.height = rgn.AbsoluteTop(), rgn.AbsoluteLeft(), rgn.Width(), rgn.Height()
	// outer draw...
	rgn.FillWith(0, 0, s.height, s.width, '\u00A0', settingsBgStyle)
	rgn.BoxRounded(0, 0, s.height, s.width, settingsTextStyle)
	rgn.TextCenter(0, 0, s.width, "Settings", settingsTextStyle)
	// tabs...
	s.renderTabs(rgn)
	s.currentForm = nil
	var csr *tea.Cursor
	switch s.tab {
	case settingsTabGrid:
		s.currentForm = gridForm
	case settingsTabRule:
		s.currentForm = ruleForm
	case settingsTabPatterns:
		s.currentForm = patternsForm
	case settingsTabLoad:
		s.currentForm = loadPatternsForm
	}
	if s.currentForm != nil {
		csr = s.currentForm.Render(s, s.clickPts, rgn)
	}
	return csr
}

var gridForm = &layout.Form[*settings]{
	Style:        settingsTextStyle,
	FocusedStyle: settingsTabStyle,
	FormRows: layout.FormRows[*settings]{
		3: {
			2: {
				Item: layout.NewButton("Clear", func(s *settings) tea.Cmd {
					s.m.grid.Clear()
					return nil
				}),
			},
		},
		5: {
			2: {Item: "Randomization %:"},
			20: {
				Item: layout.NewNumberInput(3, 0, 100, func(s *settings) int {
					return s.m.random
				}, func(s *settings, value int) tea.Cmd {
					s.m.random = value
					s.m.prefs.Random = value
					return s.m.savePrefs()
				}),
			},
			26: {
				Item: layout.NewButton("Randomize", func(s *settings) tea.Cmd {
					s.m.grid.Randomize(s.m.random)
					return nil
				}),
			},
		},
		7: {
			2: {Item: "Step delay (ms):"},
			20: {
				Item: layout.NewNumberInput(4, 1, 2000, func(s *settings) int {
					return s.m.stepDelay
				}, func(s *settings, value int) tea.Cmd {
					s.m.stepDelay = value
					s.m.prefs.StepDelay = value
					return s.m.savePrefs()
				}),
			},
		},
		9: {
			2: {Item: "  Wrapping mode:"},
			19: {
				Item: layout.NewRadio([]string{"None", "Horizontal", "Vertical", "Toroidal"}, func(s *settings) int {
					return int(s.m.grid.WrapMode)
				}, func(s *settings, value int) tea.Cmd {
					s.m.grid.WrapMode = logic.WrapMode(value)
					return nil
				}),
			},
		},
		11: {
			2: {Item: " Boundary cells:"},
			19: {
				Item: layout.NewRadio([]string{"Dead", "Alive"}, func(s *settings) int {
					return int(s.m.grid.BoundaryMode)
				}, func(s *settings, value int) tea.Cmd {
					s.m.grid.BoundaryMode = logic.BoundaryMode(value)
					return nil
				}),
			},
		},
		14: {
			2: {Item: "Height:       Width:"},
			10: {
				Item: layout.NewNumberInput(4, 2, 9999, func(s *settings) int {
					return s.m.gridHeight
				}, func(s *settings, value int) tea.Cmd {
					s.m.gridHeight = value
					return nil
				}),
			},
			23: {
				Item: layout.NewNumberInput(4, 2, 9999, func(s *settings) int {
					return s.m.gridWidth
				}, func(s *settings, value int) tea.Cmd {
					s.m.gridWidth = value
					return nil
				}),
			},
			30: {
				Item: layout.NewButton("Resize", func(s *settings) tea.Cmd {
					if s.m.gridHeight == s.m.grid.Height && s.m.gridWidth == s.m.grid.Width {
						// no resize needed
						return nil
					}
					return func() tea.Msg {
						if grid, err := logic.NewGrid(s.m.gridHeight, s.m.gridWidth, s.m.grid.WrapMode, s.m.grid.BoundaryMode); err == nil {
							grid.Rule = s.m.grid.Rule
							return gridResizeResult{
								surface: newGridSurface(grid, s.m.cellStyle),
								grid:    grid,
							}
						}
						return nil
					}
				}),
			},
			39: {
				Item: layout.NewButton("Fit Screen", func(s *settings) tea.Cmd {
					if s.m.gridHeight == s.m.height*2 && s.m.gridWidth == s.m.width*2 {
						// no resize needed
						return nil
					}
					s.m.gridHeight, s.m.gridWidth = s.m.height*2, s.m.width*2
					return func() tea.Msg {
						if grid, err := logic.NewGrid(s.m.gridHeight, s.m.gridWidth, s.m.grid.WrapMode, s.m.grid.BoundaryMode); err == nil {
							grid.Rule = s.m.grid.Rule
							return gridResizeResult{
								surface: newGridSurface(grid, s.m.cellStyle),
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
				Item: layout.NewNumberInput(3, 0, 255, func(s *settings) int {
					v, _, _ := rgb(s.m.cellStyle.GetForeground())
					return v
				}, func(s *settings, value int) tea.Cmd {
					_, g, b := rgb(s.m.cellStyle.GetForeground())
					s.adjustCellColor(true, value, g, b)
					s.m.prefs.setCellStyle(s.m.cellStyle)
					return s.m.savePrefs()
				}),
			},
			24: {
				Item: layout.NewNumberInput(3, 0, 255, func(s *settings) int {
					_, v, _ := rgb(s.m.cellStyle.GetForeground())
					return v
				}, func(s *settings, value int) tea.Cmd {
					r, _, b := rgb(s.m.cellStyle.GetForeground())
					s.adjustCellColor(true, r, value, b)
					s.m.prefs.setCellStyle(s.m.cellStyle)
					return s.m.savePrefs()

				}),
			},
			33: {
				Item: layout.NewNumberInput(3, 0, 255, func(s *settings) int {
					_, _, v := rgb(s.m.cellStyle.GetForeground())
					return v
				}, func(s *settings, value int) tea.Cmd {
					r, g, _ := rgb(s.m.cellStyle.GetForeground())
					s.adjustCellColor(true, r, g, value)
					s.m.prefs.setCellStyle(s.m.cellStyle)
					return s.m.savePrefs()

				}),
			},
		},
		18: {
			2: {
				Item: "Background  R:      G:      B:",
			},
			17: {
				Item: layout.NewNumberInput(3, 0, 255, func(s *settings) int {
					v, _, _ := rgb(s.m.cellStyle.GetBackground())
					return v
				}, func(s *settings, value int) tea.Cmd {
					_, g, b := rgb(s.m.cellStyle.GetBackground())
					s.adjustCellColor(false, value, g, b)
					s.m.prefs.setCellStyle(s.m.cellStyle)
					return s.m.savePrefs()

				}),
			},
			24: {
				Item: layout.NewNumberInput(3, 0, 255, func(s *settings) int {
					_, v, _ := rgb(s.m.cellStyle.GetBackground())
					return v
				}, func(s *settings, value int) tea.Cmd {
					r, _, b := rgb(s.m.cellStyle.GetBackground())
					s.adjustCellColor(false, r, value, b)
					s.m.prefs.setCellStyle(s.m.cellStyle)
					return s.m.savePrefs()

				}),
			},
			33: {
				Item: layout.NewNumberInput(3, 0, 255, func(s *settings) int {
					_, _, v := rgb(s.m.cellStyle.GetBackground())
					return v
				}, func(s *settings, value int) tea.Cmd {
					r, g, _ := rgb(s.m.cellStyle.GetBackground())
					s.adjustCellColor(false, r, g, value)
					s.m.prefs.setCellStyle(s.m.cellStyle)
					return s.m.savePrefs()

				}),
			},
		},
	},
}

var ruleForm = &layout.Form[*settings]{
	Style:        settingsTextStyle,
	FocusedStyle: settingsTabStyle,
	FormRows: layout.FormRows[*settings]{
		3: {
			2: {Item: "       Name:"},
			15: {
				Item: layout.NewDropdownSelect(func(s *settings) []string {
					return sortedRuleNames
				}, -1, func(s *settings) string {
					return s.m.grid.Rule.Name()
				}, func(s *settings, value string) tea.Cmd {
					if nr, ok := logic.Rules[value]; ok {
						s.m.grid.Rule = nr
						s.m.prefs.setRule(s.m.grid.Rule)
						return s.m.savePrefs()
					} else if strings.HasPrefix(value, "Custom ") {
						s.customRuleName = ""
						if nr, err := logic.NewRuleRle("", strings.TrimPrefix(value, "Custom ")); err == nil {
							s.m.grid.Rule = nr
							s.m.prefs.setRule(s.m.grid.Rule)
							return s.m.savePrefs()
						}
					}
					return nil
				}),
			},
		},
		5: {
			2: {Item: "       Rule:"},
			15: {
				Item: layout.NewTextInput(21, "BbSs/012345678", func(s *settings) string {
					return s.m.grid.Rule.Rle()
				}, func(s *settings, value string) tea.Cmd {
					if nr, err := logic.NewRuleRle("", value); err == nil {
						s.customRuleName = ""
						s.m.grid.Rule = nr
						s.m.prefs.setRule(s.m.grid.Rule)
						return s.m.savePrefs()
					}
					return nil
				}),
			},
		},
		7: {
			2: {Item: "Permutation:"},
			15: {Item: layout.NewNumberInput(6, 0, (1<<18)-1, func(s *settings) int {
				return s.m.grid.Rule.Permutation()
			}, func(s *settings, value int) tea.Cmd {
				if nr, err := logic.NewRuleFromPermutation(value); err == nil {
					s.customRuleName = ""
					s.m.grid.Rule = nr
					s.m.prefs.setRule(s.m.grid.Rule)
					return s.m.savePrefs()
				}
				return nil
			})},
		},
		17: {
			6: {
				Item: "As name:",
				Condition: func(s *settings) bool {
					return s.m.grid.Rule.IsCustom()
				},
			},
			15: {
				Item: layout.NewTextInput(25, "", func(s *settings) string {
					return s.customRuleName
				}, func(s *settings, value string) tea.Cmd {
					s.customRuleName = value
					return nil
				}),
				Condition: func(s *settings) bool {
					return s.m.grid.Rule.IsCustom()
				},
			},
			42: {
				Item: layout.NewButton("Save", func(s *settings) tea.Cmd {
					if s.customRuleName != "" {
						s.m.prefs.addRule(s.customRuleName, s.m.grid.Rule.Rle())
						logic.AddRule(s.customRuleName, s.m.grid.Rule)
						return s.m.savePrefs()
					}
					return nil
				}),
				Condition: func(s *settings) bool {
					return s.m.grid.Rule.IsCustom() && len(s.customRuleName) > 0
				},
			},
		},
	},
}

var sortedRuleNames = func() []string {
	names := make([]string, 0, len(logic.Rules))
	for name := range logic.Rules {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}()

func (s *settings) adjustRule(born bool, add bool, digit string) {
	bw, sw := s.m.grid.Rule.BornWith(), s.m.grid.Rule.SurvivesWith()
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
		s.m.grid.Rule = nr
	}
}

func rgb(c color.Color) (int, int, int) {
	r, g, b, _ := c.RGBA()
	return int(r) / 256, int(g) / 256, int(b) / 256
}

func (s *settings) adjustCellColor(fg bool, r, g, b int) {
	if fg {
		s.m.cellStyle = s.m.cellStyle.Foreground(lipgloss.RGBColor{R: clampRGB(r), G: clampRGB(g), B: clampRGB(b)})
	} else {
		s.m.cellStyle = s.m.cellStyle.Background(lipgloss.RGBColor{R: clampRGB(r), G: clampRGB(g), B: clampRGB(b)})
	}
	s.m.gridSurface.ClearStyle(0, 0, -1, -1, &s.m.cellStyle)
}

func clampRGB(c int) uint8 {
	if c < 0 {
		return 0
	} else if c > 255 {
		return 255
	}
	return uint8(c)
}

var patternsForm = &layout.Form[*settings]{
	Style:        settingsTextStyle,
	FocusedStyle: settingsTabStyle,
	FormRows: layout.FormRows[*settings]{
		3: {
			1: {Item: "Name:"},
			7: {
				Item: layout.NewDropdownSelect(func(s *settings) []string {
					return sortedPatterns
				}, -1, func(s *settings) string {
					if s.currentPattern == nil {
						if p, ok := patterns.PatternLibrary[sortedPatterns[0]]; ok {
							s.currentPattern = &p
						} else {
							return ""
						}
					}
					return s.currentPattern.Name
				}, func(s *settings, value string) tea.Cmd {
					s.patternOffY, s.patternOffX = 0, 0
					if p, ok := patterns.PatternLibrary[value]; ok {
						s.currentPattern = &p
					}
					return nil
				}),
			},
		},
		4: {
			0: {Item: &settingsPatternPreview[*settings]{}},
		},
		18: {
			1: {Item: "At  Y:      X:       Rotate:"},
			7: {
				Item: layout.NewNumberInput(4, 0,
					func(s *settings) int {
						return s.m.grid.Height - 1
					},
					func(s *settings) int {
						return s.patternPlaceY
					},
					func(s *settings, value int) tea.Cmd {
						if value < s.m.grid.Height {
							s.patternPlaceY = value
						}
						return nil
					}),
			},
			15: {
				Item: layout.NewNumberInput(4, 0,
					func(s *settings) int {
						return s.m.grid.Width - 1
					},
					func(s *settings) int {
						return s.patternPlaceX
					},
					func(s *settings, value int) tea.Cmd {
						if value < s.m.grid.Width {
							s.patternPlaceX = value
						}
						return nil
					}),
			},
			30: {
				Item: layout.NewRadio([]string{"0°", "90°", "180°", "270°"}, func(s *settings) int {
					return int(s.patternRotate)
				}, func(s *settings, value int) tea.Cmd {
					s.patternRotate = patterns.Rotation(value)
					return nil
				}),
			},
			52: {
				Item: layout.NewButton(" Place ", func(s *settings) tea.Cmd {
					if s.currentPattern != nil {
						s.currentPattern.Draw(s.m.grid, s.patternPlaceY, s.patternPlaceX, s.patternRotate)
					}
					return nil
				}, false),
			},
		},
	},
}

type settingsPatternPreview[T any] struct{}

func (p *settingsPatternPreview[T]) Update(parent T, msg tea.Msg, focused bool) tea.Cmd {
	if focused {
		s := asSettings(parent)
		switch mt := msg.(type) {
		case tea.KeyPressMsg:
			switch mt.String() {
			case "up":
				if s.patternOffY > 0 {
					s.patternOffY--
				}
			case "down":
				s.patternOffY++
			case "left":
				if s.patternOffX > 0 {
					s.patternOffX--
				}
			case "right":
				s.patternOffX++
			case "ctrl+k":
				s.patternInfo = !s.patternInfo
			}
		}
	}
	return nil
}

func (p *settingsPatternPreview[T]) Render(parent T, form *layout.Form[T], inputNo int, sf layout.Surface, clickPts layout.ClickPoints[T], row, col int, focused bool, style lipgloss.Style, focusedStyle lipgloss.Style) *tea.Cursor {
	rgn := sf.Region(row, col+1, 14, sf.Width()-2)
	s := asSettings(parent)
	if s.currentPattern == nil {
		rgn.TextCenter(7, 0, rgn.Width(), "No preview/info available", settingsTextStyle)
	} else if s.patternInfo {
		rgn.TextCenter(1, 0, rgn.Width(), "Pattern Information", settingsTextStyle)
		rgn.Text(2, 2, "Height: "+strconv.Itoa(s.currentPattern.Height), settingsTextStyle)
		rgn.Text(2, 17, "Width: "+strconv.Itoa(s.currentPattern.Width), settingsTextStyle)
		clickPts.Add(rgn.Text(2, rgn.Width()-17, "ctrl+k", settingsTextUlStyle), func(parent T) tea.Cmd {
			s.patternInfo = false
			return nil
		})
		rgn.Text(2, rgn.Width()-10, "- Preview", settingsTextStyle)
		if s.currentPattern.Rule != nil {
			rgn.Text(3, 4, "Rule:", settingsTextStyle)
			rn := s.currentPattern.Rule.Rle()
			if n, ok := logic.RleToName(rn); ok {
				rn = n
			}
			clickPts.Add(rgn.Text(3, 10, rn, settingsTextUlStyle), func(parent T) tea.Cmd {
				s.m.grid.Rule = s.currentPattern.Rule
				s.m.prefs.setRule(s.m.grid.Rule)
				return s.m.savePrefs()
			})
		}
		rgn.Text(5, 2, "Origin: "+s.currentPattern.Origination, settingsTextStyle)
		if len(s.currentPattern.Comments) > 0 {
			rgn.Text(6, 1, "Comment:", settingsTextStyle)
			y := 6
			for _, comment := range s.currentPattern.Comments {
				y += rgn.TextWrapped(y, 10, rgn.Width()-11, comment, settingsTextStyle)
			}
		}
	} else {
		wd := s.currentPattern.Width - s.patternOffX
		if wd > rgn.Width() {
			wd = rgn.Width()
		}
		ht := s.currentPattern.Height - s.patternOffY
		if ht > rgn.Height() {
			ht = rgn.Height()
		}
		if preview := rgn.Region(0, 0, ht, wd); preview != nil {
			for y := 0; y < ht; y++ {
				clickPts.Add(layout.Placement{
					Extent: wd,
					Row:    preview.AbsoluteTop() + y,
					Col:    preview.AbsoluteLeft(),
				}, func(parent T) tea.Cmd {
					s := asSettings(parent)
					s.patternInfo = true
					return nil
				})
			}
			useStyle := settingsPreviewStyle
			if focused {
				useStyle = settingsPreviewFocusedStyle
			}
			preview.FillWith(0, 0, preview.Height(), preview.Width(), '\u00A0', useStyle)
			aliveStyle := lipgloss.NewStyle().Foreground(useStyle.GetBackground()).Background(useStyle.GetForeground())
			s.currentPattern.DrawTo(patterns.Rotate0, func(row, col int, alive bool) {
				if alive {
					ar, ac := row-s.patternOffY, col-s.patternOffX
					if ar >= 0 && ac >= 0 {
						preview.Text(ar, ac, "\u00A0", aliveStyle)
					}
				}
			})
		}
	}
	return nil
}

func asSettings(parent any) *settings {
	return parent.(*settings)
}

var sortedPatterns = sortPatterns()

func sortPatterns() []string {
	names := make([]string, 0, len(patterns.PatternLibrary))
	for name := range patterns.PatternLibrary {
		names = append(names, name)
	}
	slices.SortStableFunc(names, func(a, b string) int {
		return strings.Compare(
			strings.ToLower(a),
			strings.ToLower(b),
		)
	})
	return names
}

var loadPatternsForm = &layout.Form[*settings]{
	Style:        settingsTextStyle,
	FocusedStyle: settingsTabStyle,
	FormRows: layout.FormRows[*settings]{
		3: {
			2: {Item: "From:"},
			8: {Item: layout.NewTextInput(-1, "", func(s *settings) string {
				return s.loadFrom
			}, func(s *settings, value string) tea.Cmd {
				s.loadFrom = value
				return nil
			}, true)},
		},
		5: {
			0: {Item: "Enter path to directory or file and press...", Alignment: layout.AlignCenter},
		},
		7: {
			0: {
				Item: layout.NewButton("Load", func(s *settings) tea.Cmd {
					s.loadPatternsResult = nil
					if s.loadFrom != "" {
						return func() tea.Msg {
							fs, err := os.Stat(s.loadFrom)
							if err != nil {
								return loadPatternsResult{error: "Invalid filepath"}
							}
							if fs.IsDir() {
								return loadPatternsLibrary(s.loadFrom, false)
							} else if err = loadPattern(s.loadFrom); err == nil {
								sortedPatterns = sortPatterns()
								return loadPatternsResult{loaded: 1, filename: s.loadFrom}
							} else {
								return loadPatternsResult{error: err.Error()}
							}
						}
					}
					return nil
				}).Align(layout.AlignCenter),
			},
		},
		9: {
			0: {
				Item: func(s *settings) any {
					if s.loadPatternsResult != nil {
						if s.loadPatternsResult.error != "" {
							return "Error: " + s.loadPatternsResult.error
						} else {
							return fmt.Sprintf("Successfully loaded %d pattern(s)", s.loadPatternsResult.loaded)
						}
					}
					return ""
				},
				Alignment: layout.AlignCenter,
				Width:     -1,
			},
		},
		15: {
			0: {Item: "To clear all loaded patterns and libraries...", Alignment: layout.AlignCenter},
		},
		17: {
			0: {
				Item: layout.NewButton("Clear", func(s *settings) tea.Cmd {
					s.loadPatternsResult = nil
					patterns.ResetLibrary()
					sortedPatterns = sortPatterns()
					s.m.prefs.clearPatterns()
					return s.m.savePrefs()
				}).Align(layout.AlignCenter),
			},
		},
	},
}

type loadPatternsResult struct {
	error     string
	loaded    int
	directory string
	filename  string
}

func loadPatternsLibrary(fp string, noSort bool) loadPatternsResult {
	count := 0
	filepath.Walk(fp, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".rle") {
			if err := loadPattern(path); err == nil {
				count++
			}
		}
		return nil
	})
	if count > 0 {
		if !noSort {
			sortedPatterns = sortPatterns()
		}
		return loadPatternsResult{loaded: count, directory: fp}
	} else {
		return loadPatternsResult{error: "No .rle files found"}
	}
}

func loadPattern(fp string) error {
	f, err := os.Open(fp)
	if err != nil {
		return errors.New("Unable to open file")
	}
	defer func() {
		_ = f.Close()
	}()
	if p, err := patterns.NewPatternFromRle(f); err != nil {
		return err
	} else {
		if p.Name == "" {
			p.Name = filepath.Base(fp)
		}
		patterns.PatternLibrary[p.Name] = p
		return nil
	}
}

const (
	settingsTabGrid = iota
	settingsTabRule
	settingsTabPatterns
	settingsTabLoad
)

var settingsTabs = []struct {
	title string
	ul    int
	tabNo int
}{
	{"Grid", 0, settingsTabGrid},
	{"Rule", 0, settingsTabRule},
	{"Patterns", 0, settingsTabPatterns},
	{"Load", 1, settingsTabLoad},
}

func (s *settings) renderTabs(rgn layout.Surface) {
	x := 3
	for _, tab := range settingsTabs {
		if tab.tabNo == s.tab {
			rgn.Text(1, x-1, " "+tab.title+" ", settingsTabStyle)
		} else {
			s.clickPts.Add(rgn.Text(1, x, tab.title, settingsTextStyle), func(s *settings) tea.Cmd {
				s.tab = tab.tabNo
				return nil
			})
			if tab.ul != -1 {
				rgn.Text(1, x+tab.ul, tab.title[tab.ul:tab.ul+1], settingsTextUlStyle)
			}
		}
		x += len(tab.title) + 3
	}
}

func (s *settings) update(msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		switch mt.String() {
		case "ctrl+c":
			return tea.Quit
		case "esc":
			s.m.settingsShowing = false
		case "ctrl+g":
			s.tab = settingsTabGrid
		case "ctrl+r":
			s.tab = settingsTabRule
		case "ctrl+p":
			s.tab = settingsTabPatterns
		case "ctrl+o":
			s.tab = settingsTabLoad
		default:
			if s.currentForm != nil {
				return s.currentForm.Update(s, msg)
			}
		}
	case tea.MouseClickMsg:
		clickX, clickY := mt.X-s.absLeft, mt.Y-s.absTop
		if clickX < 0 || clickX >= s.width || clickY < 0 || clickY >= s.height {
			s.m.settingsShowing = false
		} else if fn, ok := s.clickPts[layout.ClickPoint{mt.Y, mt.X}]; ok {
			return fn(s)
		} else if s.currentForm != nil {
			return s.currentForm.Update(s, msg)
		}
	case tea.MouseWheelMsg:
		if s.currentForm != nil {
			return s.currentForm.Update(s, msg)
		}
	case tea.PasteMsg:
		if s.currentForm != nil {
			if cmd := s.currentForm.Update(s, mt); cmd != nil {
				return cmd
			}
		}
	case loadPatternsResult:
		s.loadPatternsResult = &mt
		if mt.error == "" && mt.loaded > 0 {
			if mt.filename != "" {
				s.m.prefs.addPattern(mt.filename)
			} else {
				s.m.prefs.addPatternLibrary(mt.directory)
			}
			return s.m.savePrefs()
		}
	default:
		if s.currentForm != nil {
			return s.currentForm.Update(s, msg)
		}
	}
	return nil
}
