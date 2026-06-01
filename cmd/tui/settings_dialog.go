package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	clickPts           clickPts
	absTop, absLeft    int
	width, height      int
	currentPattern     *patterns.Pattern
	patternPlaceY      int
	patternPlaceX      int
	patternOffY        int
	patternOffX        int
	loadFrom           string
	loadPatternsResult *loadPatternsResult
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
	settingsPreviewAliveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).
					Background(lipgloss.Color("#000000"))
)

func (s *settings) render(rgn layout.Surface) *tea.Cursor {
	s.clickPts = clickPts{}
	s.absTop, s.absLeft, s.width, s.height = rgn.AbsoluteTop(), rgn.AbsoluteLeft(), rgn.Width(), rgn.Height()

	// outer draw...
	rgn.FillWith(0, 0, s.height, s.width, '\u00A0', settingsBgStyle)
	rgn.BoxRounded(0, 0, s.height, s.width, settingsTextStyle)
	rgn.TextCenter(0, 0, s.width, "Settings", settingsTextStyle)
	// tabs...
	s.renderTabs(rgn)

	var csr *tea.Cursor
	switch s.tab {
	case tabGrid:
		csr = s.renderGridSettings(rgn)
	case tabRule:
		csr = s.renderRuleSettings(rgn)
	case tabColors:
		csr = s.renderColorsSettings(rgn)
	case tabPatterns:
		csr = s.renderPatternSettings(rgn)
	case tabLoad:
		csr = s.renderLoadPatternsSettings(rgn)
	}
	return csr
}

func (s *settings) renderGridSettings(rgn layout.Surface) *tea.Cursor {
	rgn.BoxRounded(2, 1, 3, 7, settingsTextStyle)
	s.clickPts.add(rgn.Text(3, 2, "Clear", settingsTextStyle), func() tea.Cmd {
		s.m.grid.Clear()
		return nil
	})

	rgn.Text(5, 2, "Randomization %: ▲    ▼", settingsTextStyle)
	rgn.TextRight(5, 20, 4, strconv.Itoa(s.m.random), settingsTextStyle)
	if s.m.random < 100 {
		s.clickPts.add(rgn.Text(5, 19, "▲", settingsTextStyle), func() tea.Cmd {
			s.m.random++
			return nil
		})
	}
	if s.m.random > 0 {
		s.clickPts.add(rgn.Text(5, 24, "▼", settingsTextStyle), func() tea.Cmd {
			s.m.random--
			return nil
		})
	}
	rgn.BoxRounded(4, 26, 3, 11, settingsTextStyle)
	s.clickPts.add(rgn.Text(5, 27, "Randomize", settingsTextStyle), func() tea.Cmd {
		s.m.grid.Randomize(s.m.random)
		return nil
	})

	rgn.Text(7, 2, "Step delay (ms): ▲    ▼", settingsTextStyle)
	rgn.TextRight(7, 20, 4, strconv.Itoa(s.m.stepDelay), settingsTextStyle)
	if s.m.stepDelay < 2000 {
		s.clickPts.add(rgn.Text(7, 19, "▲", settingsTextStyle), func() tea.Cmd {
			s.m.stepDelay++
			return nil
		})
	}
	if s.m.stepDelay > 1 {
		s.clickPts.add(rgn.Text(7, 24, "▼", settingsTextStyle), func() tea.Cmd {
			s.m.stepDelay--
			return nil
		})
	}

	rgn.Text(9, 2, "  Grid wrapping:", settingsTextStyle)
	if s.m.grid.WrapMode == logic.WrapNone {
		rgn.Text(9, 19, " None ", settingsTabStyle)
	} else {
		s.clickPts.add(rgn.Text(9, 20, "None", settingsTextUlStyle), func() tea.Cmd {
			s.m.grid.SetWrapMode(logic.WrapNone)
			return nil
		})
	}
	if s.m.grid.WrapMode == logic.WrapHorizontal {
		rgn.Text(9, 25, " Horizontal ", settingsTabStyle)
	} else {
		s.clickPts.add(rgn.Text(9, 26, "Horizontal", settingsTextUlStyle), func() tea.Cmd {
			s.m.grid.SetWrapMode(logic.WrapHorizontal)
			return nil
		})
	}
	if s.m.grid.WrapMode == logic.WrapVertical {
		rgn.Text(9, 37, " Vertical ", settingsTabStyle)
	} else {
		s.clickPts.add(rgn.Text(9, 38, "Vertical", settingsTextUlStyle), func() tea.Cmd {
			s.m.grid.SetWrapMode(logic.WrapVertical)
			return nil
		})
	}
	if s.m.grid.WrapMode == logic.WrapAll {
		rgn.Text(9, 47, " Toroidal ", settingsTabStyle)
	} else {
		s.clickPts.add(rgn.Text(9, 48, "Toroidal", settingsTextUlStyle), func() tea.Cmd {
			s.m.grid.SetWrapMode(logic.WrapAll)
			return nil
		})
	}
	rgn.Text(11, 2, " Boundary cells:", settingsTextStyle)
	if s.m.grid.BoundaryMode == logic.DeadBoundary {
		rgn.Text(11, 19, " Dead ", settingsTabStyle)
	} else {
		s.clickPts.add(rgn.Text(11, 20, "Dead", settingsTextUlStyle), func() tea.Cmd {
			s.m.grid.SetBoundaryMode(logic.DeadBoundary)
			return nil
		})
	}
	if s.m.grid.BoundaryMode == logic.AliveBoundary {
		rgn.Text(11, 25, " Alive ", settingsTabStyle)
	} else {
		s.clickPts.add(rgn.Text(11, 26, "Alive", settingsTextUlStyle), func() tea.Cmd {
			s.m.grid.SetBoundaryMode(logic.AliveBoundary)
			return nil
		})
	}
	/*
			Grid strl+g
		todo		* size
	*/
	return nil
}

func (s *settings) renderRuleSettings(rgn layout.Surface) *tea.Cursor {
	rgn.Text(3, 2, "       Name: ▲▼", settingsTextStyle)
	rgn.Text(3, 18, s.m.grid.Rule.Name(), settingsTextStyle)
	s.clickPts.add(rgn.Text(3, 15, "▲", settingsTextStyle), func() tea.Cmd {
		idx := slices.Index(sortedRuleNames, s.m.grid.Rule.Name())
		if idx <= 0 {
			idx = len(sortedRuleNames) - 1
		} else {
			idx--
		}
		s.m.grid.Rule = logic.Rules[sortedRuleNames[idx]]
		return nil
	})
	s.clickPts.add(rgn.Text(3, 16, "▼", settingsTextStyle), func() tea.Cmd {
		idx := slices.Index(sortedRuleNames, s.m.grid.Rule.Name())
		if idx == -1 || idx >= len(sortedRuleNames)-1 {
			idx = 0
		} else {
			idx++
		}
		s.m.grid.Rule = logic.Rules[sortedRuleNames[idx]]
		return nil
	})
	rgn.Text(5, 2, "        RLE:", settingsTextStyle)
	rgn.Text(5, 15, s.m.grid.Rule.Rle(), settingsTextStyle)
	rgn.Text(6, 2, "       Born:", settingsTextStyle)
	bw := s.m.grid.Rule.BornWith()
	for w := 0; w < 9; w++ {
		digit := strconv.Itoa(w)
		if strings.Contains(bw, digit) {
			s.clickPts.add(rgn.Text(6, 15+(w*2), digit, settingsTabStyle), func() tea.Cmd {
				s.adjustRule(true, false, digit)
				return nil
			})
		} else {
			s.clickPts.add(rgn.Text(6, 15+(w*2), digit, settingsTextUlStyle), func() tea.Cmd {
				s.adjustRule(true, true, digit)
				return nil
			})
		}
	}
	rgn.Text(7, 2, "   Survives:", settingsTextStyle)
	sw := s.m.grid.Rule.SurvivesWith()
	for w := 0; w < 9; w++ {
		digit := strconv.Itoa(w)
		if strings.Contains(sw, digit) {
			s.clickPts.add(rgn.Text(7, 15+(w*2), digit, settingsTabStyle), func() tea.Cmd {
				s.adjustRule(false, false, digit)
				return nil
			})
		} else {
			s.clickPts.add(rgn.Text(7, 15+(w*2), digit, settingsTextUlStyle), func() tea.Cmd {
				s.adjustRule(false, true, digit)
				return nil
			})
		}
	}

	rgn.Text(9, 2, "Permutation: ▲▼", settingsTextStyle)
	perm := s.m.grid.Rule.Permutation()
	rgn.Text(9, 18, strconv.Itoa(perm), settingsTextStyle)
	if perm > 0 {
		s.clickPts.add(rgn.Text(9, 16, "▼", settingsTextStyle), func() tea.Cmd {
			if nr, err := logic.NewRuleFromPermutation(s.m.grid.Rule.Permutation() - 1); err == nil {
				s.m.grid.Rule = nr
			}
			return nil
		})
	}
	if perm < (1 << 18) {
		s.clickPts.add(rgn.Text(9, 15, "▲", settingsTextStyle), func() tea.Cmd {
			if nr, err := logic.NewRuleFromPermutation(s.m.grid.Rule.Permutation() + 1); err == nil {
				s.m.grid.Rule = nr
			}
			return nil
		})
	}
	return nil
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
		s.m.cellStyle = s.m.cellStyle.Foreground(lipgloss.RGBColor{R: uint8(r), G: uint8(g), B: uint8(b)})
	} else {
		s.m.cellStyle = s.m.cellStyle.Background(lipgloss.RGBColor{R: uint8(r), G: uint8(g), B: uint8(b)})
	}
	s.m.gridSurface.ClearStyle(0, 0, -1, -1, &s.m.cellStyle)
}

func (s *settings) renderColorsSettings(rgn layout.Surface) *tea.Cursor {
	fgR, fgG, fgB := rgb(s.m.cellStyle.GetForeground())
	bgR, bgG, bgB := rgb(s.m.cellStyle.GetBackground())
	rgn.Text(3, 2, "Foreground  R: ▲   ▼   G: ▲   ▼   B: ▲   ▼", settingsTextStyle)
	rgn.Text(4, 2, "Background  R: ▲   ▼   G: ▲   ▼   B: ▲   ▼", settingsTextStyle)
	rgn.TextRight(3, 18, 3, strconv.Itoa(fgR), settingsTextStyle)
	rgn.TextRight(3, 29, 3, strconv.Itoa(fgG), settingsTextStyle)
	rgn.TextRight(3, 40, 3, strconv.Itoa(fgB), settingsTextStyle)
	rgn.TextRight(4, 18, 3, strconv.Itoa(bgR), settingsTextStyle)
	rgn.TextRight(4, 29, 3, strconv.Itoa(bgG), settingsTextStyle)
	rgn.TextRight(4, 40, 3, strconv.Itoa(bgB), settingsTextStyle)
	if fgR < 255 {
		s.clickPts.add(rgn.Text(3, 17, "▲", settingsTextStyle), func() tea.Cmd {
			s.adjustCellColor(true, fgR+1, fgG, fgB)
			return nil
		})
	}
	if fgR > 0 {
		s.clickPts.add(rgn.Text(3, 21, "▼", settingsTextStyle), func() tea.Cmd {
			s.adjustCellColor(true, fgR-1, fgG, fgB)
			return nil
		})
	}
	if fgG < 255 {
		s.clickPts.add(rgn.Text(3, 28, "▲", settingsTextStyle), func() tea.Cmd {
			s.adjustCellColor(true, fgR, fgG+1, fgB)
			return nil
		})
	}
	if fgG > 0 {
		s.clickPts.add(rgn.Text(3, 32, "▼", settingsTextStyle), func() tea.Cmd {
			s.adjustCellColor(true, fgR, fgG-1, fgB)
			return nil
		})
	}
	if fgB < 255 {
		s.clickPts.add(rgn.Text(3, 39, "▲", settingsTextStyle), func() tea.Cmd {
			s.adjustCellColor(true, fgR, fgG, fgB+1)
			return nil
		})
	}
	if fgB > 0 {
		s.clickPts.add(rgn.Text(3, 43, "▼", settingsTextStyle), func() tea.Cmd {
			s.adjustCellColor(true, fgR, fgG, fgB-1)
			return nil
		})
	}
	if bgR < 255 {
		s.clickPts.add(rgn.Text(4, 17, "▲", settingsTextStyle), func() tea.Cmd {
			s.adjustCellColor(false, bgR+1, bgG, bgB)
			return nil
		})
	}
	if bgR > 0 {
		s.clickPts.add(rgn.Text(4, 21, "▼", settingsTextStyle), func() tea.Cmd {
			s.adjustCellColor(false, bgR-1, bgG, bgB)
			return nil
		})
	}
	if bgG < 255 {
		s.clickPts.add(rgn.Text(4, 28, "▲", settingsTextStyle), func() tea.Cmd {
			s.adjustCellColor(false, bgR, bgG+1, bgB)
			return nil
		})
	}
	if bgG > 0 {
		s.clickPts.add(rgn.Text(4, 32, "▼", settingsTextStyle), func() tea.Cmd {
			s.adjustCellColor(false, bgR, bgG-1, bgB)
			return nil
		})
	}
	if bgB < 255 {
		s.clickPts.add(rgn.Text(4, 39, "▲", settingsTextStyle), func() tea.Cmd {
			s.adjustCellColor(false, bgR, bgG, bgB+1)
			return nil
		})
	}
	if bgB > 0 {
		s.clickPts.add(rgn.Text(4, 43, "▼", settingsTextStyle), func() tea.Cmd {
			s.adjustCellColor(false, bgR, bgG, bgB-1)
			return nil
		})
	}
	return nil
}

func (s *settings) renderPatternSettings(rgn layout.Surface) *tea.Cursor {
	if s.currentPattern == nil {
		if p, ok := patterns.PatternLibrary[sortedPatterns[0]]; ok {
			s.currentPattern = &p
		}
	}
	if s.currentPattern != nil {
		rgn.Text(2, 1, "Name:", settingsTextStyle)
		s.clickPts.add(rgn.Text(2, 7, "▲", settingsTextStyle), func() tea.Cmd {
			s.patternOffX, s.patternOffY = 0, 0
			if s.currentPattern != nil {
				idx := slices.Index(sortedPatterns, s.currentPattern.Name)
				if idx <= 0 {
					idx = len(sortedPatterns) - 1
				} else {
					idx--
				}
				if p, ok := patterns.PatternLibrary[sortedPatterns[idx]]; ok {
					s.currentPattern = &p
				}
			}
			return nil
		})
		s.clickPts.add(rgn.Text(2, 8, "▼", settingsTextStyle), func() tea.Cmd {
			s.patternOffX, s.patternOffY = 0, 0
			if s.currentPattern != nil {
				idx := slices.Index(sortedPatterns, s.currentPattern.Name)
				if idx == -1 || idx >= len(sortedPatterns)-1 {
					idx = 0
				} else {
					idx++
				}
				if p, ok := patterns.PatternLibrary[sortedPatterns[idx]]; ok {
					s.currentPattern = &p
				}
			}
			return nil
		})
		rgn.Text(2, 10, s.currentPattern.Name, settingsTextStyle)
		wd := s.currentPattern.Width - s.patternOffX
		if wd > rgn.Width()-2 {
			wd = rgn.Width() - 2
		}
		ht := s.currentPattern.Height - s.patternOffY
		if ht > rgn.Height()-5 {
			ht = rgn.Height() - 5
		}
		if preview := rgn.Region(3, 1, ht, wd); preview != nil {
			preview.FillWith(0, 0, preview.Height(), preview.Width(), '\u00A0', settingsPreviewStyle)
			s.currentPattern.DrawTo(func(row, col int, alive bool) {
				if alive {
					ar, ac := row-s.patternOffY, col-s.patternOffX
					if ar >= 0 && ac >= 0 {
						preview.Text(ar, ac, "\u00A0", settingsPreviewAliveStyle)
					}
				}
			})
		}
		s.clickPts.add(rgn.Text(rgn.Height()-2, 1, "Place", settingsTabStyle), func() tea.Cmd {
			if s.currentPattern != nil {
				s.currentPattern.Draw(s.m.grid, s.patternPlaceY, s.patternPlaceX)
			}
			return nil
		})
		rgn.Text(rgn.Height()-2, 8, "At Y: ??      X: ??", settingsTextStyle)
		s.clickPts.add(rgn.Text(rgn.Height()-2, 14, "▲", settingsTextStyle), func() tea.Cmd {
			s.patternPlaceY++
			return nil
		})
		s.clickPts.add(rgn.Text(rgn.Height()-2, 15, "▼", settingsTextStyle), func() tea.Cmd {
			if s.patternPlaceY > 0 {
				s.patternPlaceY--
			}
			return nil
		})
		s.clickPts.add(rgn.Text(rgn.Height()-2, 25, "▲", settingsTextStyle), func() tea.Cmd {
			s.patternPlaceX++
			return nil
		})
		s.clickPts.add(rgn.Text(rgn.Height()-2, 26, "▼", settingsTextStyle), func() tea.Cmd {
			if s.patternPlaceX > 0 {
				s.patternPlaceX--
			}
			return nil
		})
		rgn.Text(rgn.Height()-2, 16, strconv.Itoa(s.patternPlaceY), settingsTextStyle)
		rgn.Text(rgn.Height()-2, 27, strconv.Itoa(s.patternPlaceX), settingsTextStyle)
	}
	return nil
}

var sortedPatterns = sortPatterns()

func sortPatterns() []string {
	names := make([]string, 0, len(patterns.PatternLibrary))
	for name := range patterns.PatternLibrary {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (s *settings) renderLoadPatternsSettings(rgn layout.Surface) *tea.Cursor {
	rgn.Text(3, 2, "From:", settingsTextStyle)
	maxWd := rgn.Width() - 8
	show := s.loadFrom
	if len(show) > maxWd {
		show = show[len(show)-maxWd:]
	}
	rgn.TextFixed(3, 7, maxWd, show, settingsTabStyle)
	rgn.TextCenter(5, 2, rgn.Width()-4, "Enter path to directory/file and press...", settingsTextStyle)
	s.clickPts.add(rgn.Text(6, (rgn.Width()/2)-2, "Load", settingsTabStyle), func() tea.Cmd {
		s.loadPatternsResult = nil
		if s.loadFrom != "" {
			return func() tea.Msg {
				fs, err := os.Stat(s.loadFrom)
				if err != nil {
					return loadPatternsResult{error: "Invalid filepath"}
				}
				if fs.IsDir() {
					count := 0
					filepath.Walk(s.loadFrom, func(path string, info os.FileInfo, err error) error {
						if !info.IsDir() && strings.HasSuffix(info.Name(), ".rle") {
							func(path string) {
								if f, err := os.Open(path); err == nil {
									defer f.Close()
									if p, err := patterns.NewPatternFromRle(f); err == nil {
										patterns.PatternLibrary[p.Name] = p
										count++
									}
								}
							}(path)
						}
						return nil
					})
					if count > 0 {
						sortedPatterns = sortPatterns()
						return loadPatternsResult{loaded: count}
					} else {
						return loadPatternsResult{error: "No .rle files found"}
					}
				} else {
					f, err := os.Open(s.loadFrom)
					if err != nil {
						return loadPatternsResult{error: "Unable to open file"}
					}
					defer f.Close()
					if p, err := patterns.NewPatternFromRle(f); err != nil {
						return loadPatternsResult{error: err.Error()}
					} else {
						patterns.PatternLibrary[p.Name] = p
						sortedPatterns = sortPatterns()
						return loadPatternsResult{loaded: 1}
					}
				}
				return nil
			}
		}
		return nil
	})
	if s.loadPatternsResult != nil {
		if s.loadPatternsResult.error != "" {
			rgn.TextWrapped(7, 1, rgn.Width()-2, "Error: "+s.loadPatternsResult.error, settingsTextStyle)
		} else {
			rgn.TextCenter(7, 1, rgn.Width()-2, fmt.Sprintf("Successfully loaded %d pattern(s)", s.loadPatternsResult.loaded), settingsTextStyle)
		}
	}
	l := len(show)
	if l == maxWd {
		l--
	}
	return tea.NewCursor(rgn.AbsoluteLeft()+7+l, rgn.AbsoluteTop()+3)
}

type loadPatternsResult struct {
	error  string
	loaded int
}

const (
	tabGrid = iota
	tabRule
	tabColors
	tabPatterns
	tabLoad
)

var tabs = []struct {
	title string
	ul    int
	tabNo int
}{
	{"Grid", 0, tabGrid},
	{"Rule", 0, tabRule},
	{"Colors", 1, tabColors},
	{"Patterns", 0, tabPatterns},
	{"Load", -1, tabLoad},
}

func (s *settings) renderTabs(rgn layout.Surface) {
	x := 3
	for _, tab := range tabs {
		if tab.tabNo == s.tab {
			rgn.Text(1, x-1, " "+tab.title+" ", settingsTabStyle)
		} else {
			s.clickPts.add(rgn.Text(1, x, tab.title, settingsTextStyle), func() tea.Cmd {
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
			s.tab = tabGrid
		case "ctrl+r":
			s.tab = tabRule
		case "ctrl+o":
			s.tab = tabColors
		case "ctrl+p":
			s.tab = tabPatterns
		case "backspace":
			if s.tab == tabLoad && len(s.loadFrom) > 0 {
				s.loadFrom = s.loadFrom[:len(s.loadFrom)-1]
			}
		case "up":
			if s.tab == tabPatterns && s.patternOffY > 0 {
				s.patternOffY--
			}
		case "down":
			if s.tab == tabPatterns {
				s.patternOffY++
			}
		case "left":
			if s.tab == tabPatterns && s.patternOffX > 0 {
				s.patternOffX--
			}
		case "right":
			if s.tab == tabPatterns {
				s.patternOffX++
			}
		case "pgup":
			if s.tab == tabPatterns {
				s.patternOffX, s.patternOffY = 0, 0
				idx := 0
				if s.currentPattern != nil {
					idx = slices.Index(sortedPatterns, s.currentPattern.Name)
					if idx <= 0 {
						idx = len(sortedPatterns) - 1
					} else {
						idx--
					}
				}
				if p, ok := patterns.PatternLibrary[sortedPatterns[idx]]; ok {
					s.currentPattern = &p
				}
			}
		case "pgdown":
			if s.tab == tabPatterns {
				s.patternOffX, s.patternOffY = 0, 0
				idx := 0
				if s.currentPattern != nil {
					idx = slices.Index(sortedPatterns, s.currentPattern.Name)
					if idx == -1 || idx >= len(sortedPatterns)-1 {
						idx = 0
					} else {
						idx++
					}
				}
				if p, ok := patterns.PatternLibrary[sortedPatterns[idx]]; ok {
					s.currentPattern = &p
				}
			}
		default:
			if len(mt.String()) == 1 {
				if s.tab == tabLoad {
					s.loadFrom += mt.String()
				} else if s.tab == tabPatterns {
					str := strings.ToLower(mt.String())
					idx := slices.IndexFunc(sortedPatterns, func(s string) bool {
						return strings.HasPrefix(strings.ToLower(s), str)
					})
					if idx != -1 {
						if p, ok := patterns.PatternLibrary[sortedPatterns[idx]]; ok {
							s.currentPattern = &p
						}
					}
				}
			}
		}
	case tea.MouseClickMsg:
		clickX, clickY := mt.X-s.absLeft, mt.Y-s.absTop
		if clickX < 0 || clickX >= s.width || clickY < 0 || clickY >= s.height {
			s.m.settingsShowing = false
		} else if fn, ok := s.clickPts[clickPt{mt.Y, mt.X}]; ok {
			return fn()
		}
	case tea.PasteMsg:
		switch s.tab {
		case tabRule:
			if r, err := logic.NewRuleRle("", mt.Content); err == nil {
				s.m.grid.Rule = r
			}
		case tabLoad:
			s.loadFrom = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(mt.Content, "\n", ""), "\r", ""), "\t", "")
		}
	case loadPatternsResult:
		s.loadPatternsResult = &mt
	}
	return nil
}

type clickPt [2]int //Y,X

type clickPts map[clickPt]func() tea.Cmd

func (cp clickPts) add(pl layout.Placement, fn func() tea.Cmd) {
	for l := 0; l < pl.Extent; l++ {
		cp[clickPt{pl.Row, pl.Col + l}] = fn
	}
}

/*

Rule ctrl+r
* rule name
* rle
* perm

Colors ctrl+o
* foreground
* background

Patterns ctrl+p
* pattern select
* position
* place button
*/
