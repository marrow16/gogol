package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"errors"
	"fmt"
	"github.com/marrow16/gogol/cmd/tui/layout"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	patternsTabPatterns = iota
	patternsTabLoad
)

var patternsTabs = tabs{
	{"Patterns", 0, patternsTabPatterns},
	{"Load", 1, patternsTabLoad},
}

type patternsDialog struct {
	m                      *model
	tab                    int
	currentForm            *layout.Form[*patternsDialog]
	clickPts               layout.ClickPoints[*patternsDialog]
	currentPattern         *patterns.Pattern
	patternPlaceY          int
	patternPlaceX          int
	patternRotate          patterns.Rotation
	patternOffY            int
	patternOffX            int
	patternInfo            bool
	loadFrom               string
	loadPatternsResult     *loadPatternsResult
	sortedPatterns         []string
	patternRuleFilter      string
	pastedRle              *string
	pastedPattern          *patterns.Pattern
	pastedRleError         error
	pastedOffY, pastedOffX int
	pastedPreview          bool
}

func (d *patternsDialog) title() string {
	return "[patterns]"
}

func (d *patternsDialog) render(rgn layout.Surface) *tea.Cursor {
	d.clickPts = layout.ClickPoints[*patternsDialog]{}
	// outer draw...
	rgn.FillWith(0, 0, rgn.Height(), rgn.Width(), '\u00A0', dialogBgStyle)
	rgn.BoxRounded(0, 0, rgn.Height(), rgn.Width(), dialogTextStyle)
	rgn.TextCenter(0, 0, rgn.Width(), "Patterns", dialogTextStyle)
	renderTabs(rgn, d.clickPts, patternsTabs, d.tab, func(t int) {
		d.tab = t
	})
	d.currentForm = nil
	var csr *tea.Cursor
	switch d.tab {
	case patternsTabPatterns:
		d.currentForm = patternsForm
	case patternsTabLoad:
		d.currentForm = loadPatternsForm
	}
	if d.currentForm != nil {
		csr = d.currentForm.Render(d, d.clickPts, rgn)
	}
	return csr
}

var patternsForm = &layout.Form[*patternsDialog]{
	Style:        dialogTextStyle,
	FocusedStyle: dialogTabStyle,
	FormRows: layout.FormRows[*patternsDialog]{
		3: {
			2: {Item: "Name:"},
			8: {
				Item: layout.NewDropdownSelect(func(d *patternsDialog) []string {
					return d.getPatterns()
				}, -1, func(d *patternsDialog) string {
					if d.currentPattern == nil {
						ps := d.getPatterns()
						if p, ok := patterns.PatternLibrary[ps[0]]; ok {
							d.currentPattern = &p
						} else {
							return ""
						}
					}
					return d.currentPattern.Name
				}, func(d *patternsDialog, value string) tea.Cmd {
					d.patternOffY, d.patternOffX = 0, 0
					if p, ok := patterns.PatternLibrary[value]; ok {
						d.currentPattern = &p
					}
					return nil
				}),
			},
		},
		4: {
			0: {Item: &patternPreview[*patternsDialog]{}},
		},
		18: {
			1: {Item: "At  Y:      X:       Rotate:"},
			7: {
				Item: layout.NewNumberInput(4, 0,
					func(d *patternsDialog) int {
						return d.m.grid.Height - 1
					},
					func(d *patternsDialog) int {
						return d.patternPlaceY
					},
					func(d *patternsDialog, value int) tea.Cmd {
						if value < d.m.grid.Height {
							d.patternPlaceY = value
						}
						return nil
					}),
			},
			15: {
				Item: layout.NewNumberInput(4, 0,
					func(d *patternsDialog) int {
						return d.m.grid.Width - 1
					},
					func(d *patternsDialog) int {
						return d.patternPlaceX
					},
					func(d *patternsDialog, value int) tea.Cmd {
						if value < d.m.grid.Width {
							d.patternPlaceX = value
						}
						return nil
					}),
			},
			30: {
				Item: layout.NewRadio([]string{"0°", "90°", "180°", "270°"}, func(d *patternsDialog) int {
					return int(d.patternRotate)
				}, func(d *patternsDialog, value int) tea.Cmd {
					d.patternRotate = patterns.Rotation(value)
					return nil
				}),
			},
			52: {
				Item: layout.NewButton(" Place ", func(d *patternsDialog) tea.Cmd {
					if d.currentPattern != nil {
						d.currentPattern.Draw(d.m.grid, d.patternPlaceY, d.patternPlaceX, d.patternRotate)
					}
					return nil
				}, false),
			},
		},
	},
}

type patternPreview[T any] struct{}

func (p *patternPreview[T]) Update(parent T, msg tea.Msg, focused bool) tea.Cmd {
	if focused {
		d := asPatternsDialog(parent)
		switch mt := msg.(type) {
		case tea.KeyPressMsg:
			switch mt.String() {
			case "home":
				d.patternOffY, d.patternOffX = 0, 0
			case "end":
				if d.currentPattern != nil {
					d.patternOffY, d.patternOffX = d.currentPattern.Height, d.currentPattern.Width
				}
			case "up":
				if d.patternOffY > 0 {
					d.patternOffY--
				}
			case "down":
				d.patternOffY++
			case "left":
				if d.patternOffX > 0 {
					d.patternOffX--
				}
			case "right":
				d.patternOffX++
			case "ctrl+k":
				d.patternOffY, d.patternOffX = 0, 0
				d.patternInfo = !d.patternInfo
			}
		case tea.MouseWheelMsg:
			switch mt.Mouse().Button {
			case tea.MouseWheelUp:
				if d.patternOffY > 0 {
					d.patternOffY--
				}
			case tea.MouseWheelDown:
				d.patternOffY++
			}
		}
	}
	return nil
}

func (p *patternPreview[T]) Reset(parent T) {
}

func (p *patternPreview[T]) Render(parent T, form *layout.Form[T], inputNo int, sf layout.Surface, clickPts layout.ClickPoints[T], row, col int, focused bool, style lipgloss.Style, focusedStyle lipgloss.Style) *tea.Cursor {
	rgn := sf.Region(row, col+1, 14, sf.Width()-2)
	d := asPatternsDialog(parent)
	if d.currentPattern == nil {
		rgn.TextCenter(7, 0, rgn.Width(), "No preview/info available", dialogTextStyle)
	} else if d.patternInfo {
		rgn.TextCenter(1, 0, rgn.Width(), "Pattern Information", dialogTextStyle)
		rgn.Text(2, 3, "Height: "+strconv.Itoa(d.currentPattern.Height), dialogTextStyle)
		rgn.Text(2, 17, "Width: "+strconv.Itoa(d.currentPattern.Width), dialogTextStyle)
		clickPts.Add(rgn.Text(2, rgn.Width()-17, "ctrl+k", dialogTextUlStyle), func(parent T) tea.Cmd {
			d.patternOffY, d.patternOffX = 0, 0
			d.patternInfo = false
			return nil
		})
		rgn.Text(2, rgn.Width()-10, "- Preview", dialogTextStyle)
		if d.currentPattern.Rule != nil {
			rgn.Text(3, 5, "Rule:", dialogTextStyle)
			rn := d.currentPattern.Rule.Rle()
			if n, ok := logic.RleToName(rn); ok {
				rn = n
			}
			clickPts.Add(rgn.Text(3, 11, rn, dialogTextUlStyle), func(parent T) tea.Cmd {
				d.m.grid.Rule = d.currentPattern.Rule
				d.m.prefs.setRule(d.m.grid.Rule)
				return d.m.savePrefs()
			})

			if len(d.patternRuleFilter) > 0 && d.patternRuleFilter == d.currentPattern.Rule.Rle() {
				clickPts.Add(rgn.Text(3, rgn.Width()-9, "Filtered", dialogTextUlStyle), func(parent T) tea.Cmd {
					d.patternRuleFilter = ""
					return nil
				})
			} else {
				clickPts.Add(rgn.Text(3, rgn.Width()-7, "Filter", dialogTextUlStyle), func(parent T) tea.Cmd {
					d.patternRuleFilter = d.currentPattern.Rule.Rle()
					return nil
				})
			}
		}
		rgn.Text(4, 1, "Filename: "+d.currentPattern.Filename, dialogTextStyle)
		rgn.Text(5, 3, "Origin: "+d.currentPattern.Origination, dialogTextStyle)
		if len(d.currentPattern.Comments) > 0 {
			rgn.Text(6, 2, "Comment:", dialogTextStyle)
			y := 6
			for _, comment := range d.currentPattern.Comments {
				y += rgn.TextWrapped(y, 11, rgn.Width()-11, comment, dialogTextStyle)
			}
		}
	} else {
		d.patternOffY, d.patternOffX = clampOffsets(d.patternOffY, d.patternOffX, rgn.Height(), rgn.Width(), d.currentPattern.Height, d.currentPattern.Width)
		if preview := rgn.Region(0, 0, min(rgn.Height(), d.currentPattern.Height), min(rgn.Width(), d.currentPattern.Width)); preview != nil {
			clickPts.AddRegion(preview, func(parent T) tea.Cmd {
				d.patternInfo = true
				return nil
			})
			useStyle := dialogPreviewStyle
			if focused {
				useStyle = dialogPreviewFocusedStyle
			}
			preview.FillWith(0, 0, preview.Height(), preview.Width(), '\u00A0', useStyle)
			aliveStyle := lipgloss.NewStyle().Foreground(useStyle.GetBackground()).Background(useStyle.GetForeground())
			d.currentPattern.DrawTo(patterns.Rotate0, func(row, col int, alive bool) {
				if alive {
					ar, ac := row-d.patternOffY, col-d.patternOffX
					if ar >= 0 && ac >= 0 {
						preview.Text(ar, ac, "\u00A0", aliveStyle)
					}
				}
			})
		}
	}
	return nil
}

var loadPatternsForm = &layout.Form[*patternsDialog]{
	Style:        dialogTextStyle,
	FocusedStyle: dialogTabStyle,
	FormRows: layout.FormRows[*patternsDialog]{
		3: {
			2: {Item: "From:"},
			8: {Item: layout.NewTextInput(-1, "", func(d *patternsDialog) string {
				return d.loadFrom
			}, func(d *patternsDialog, value string) tea.Cmd {
				d.loadFrom = value
				return nil
			}, true, true)},
		},
		5: {
			11: {
				Item: func(d *patternsDialog) any {
					if d.loadPatternsResult != nil {
						if d.loadPatternsResult.error != "" {
							return "Error: " + d.loadPatternsResult.error
						} else {
							return fmt.Sprintf("Successfully loaded %d pattern(s)", d.loadPatternsResult.loaded)
						}
					}
					return "Enter path to directory or file and press"
				},
				Alignment: layout.AlignRight,
				Width:     dialogWidth - 19,
			},
			dialogWidth - 6: {
				Item: layout.NewButton("Load", func(d *patternsDialog) tea.Cmd {
					d.loadPatternsResult = nil
					if d.loadFrom != "" {
						return func() tea.Msg {
							fs, err := os.Stat(d.loadFrom)
							if err != nil {
								return loadPatternsResult{error: "Invalid filepath"}
							}
							if fs.IsDir() {
								return loadPatternsLibrary(d.loadFrom)
							} else if err = loadPattern(d.loadFrom); err == nil {
								return loadPatternsResult{loaded: 1, filename: d.loadFrom}
							} else {
								return loadPatternsResult{error: err.Error()}
							}
						}
					}
					return nil
				}),
			},
		},
		7: {
			1: {Item: strings.Repeat("─", dialogWidth-2)},
		},
		8: {
			1: {Item: &patternRleInput[*patternsDialog]{}},
		},
		17: {
			1: {Item: strings.Repeat("─", dialogWidth-2)},
		},
		18: {
			3: {
				Item: "To clear all loaded patterns and libraries press ",
			},
			dialogWidth - 8: {
				Item: "Clear",
				OnClick: func(d *patternsDialog) tea.Cmd {
					d.loadPatternsResult = nil
					patterns.ResetLibrary()
					d.m.prefs.clearPatterns()
					return d.m.savePrefs()
				},
				Style: &dialogTextUlStyle,
			},
		},
	},
}

func (d *patternsDialog) loadPastedRle() {
	d.pastedRleError = nil
	d.pastedOffY, d.pastedOffX = 0, 0
	if d.pastedPattern != nil {
		if d.pastedPattern.Name == "" {
			d.pastedPattern.Name = "(unknown)"
		}
		if _, ok := patterns.PatternLibrary[d.pastedPattern.Name]; ok {
			d.sortedPatterns = nil
		}
		patterns.PatternLibrary[d.pastedPattern.Name] = *d.pastedPattern
		d.currentPattern = d.pastedPattern
		d.tab = patternsTabPatterns
		d.pastedRle = nil
		d.pastedPattern = nil
		patternsForm.Reset(d)
	}
}

type patternRleInput[T any] struct{}

func (p *patternRleInput[T]) Update(parent T, msg tea.Msg, focused bool) tea.Cmd {
	if focused {
		d := asPatternsDialog(parent)
		switch mt := msg.(type) {
		case tea.KeyPressMsg:
			switch mt.String() {
			case "home":
				d.pastedOffY, d.pastedOffX = 0, 0
			case "end":
				if d.pastedPattern != nil {
					d.pastedOffY, d.pastedOffX = d.pastedPattern.Height, d.pastedPattern.Width
				}
			case "up":
				if d.pastedOffY > 0 {
					d.pastedOffY--
				}
			case "down":
				d.pastedOffY++
			case "left":
				if d.pastedOffX > 0 {
					d.pastedOffX--
				}
			case "right":
				d.pastedOffX++
			case "enter":
				if d.pastedRle != nil {
					d.loadPastedRle()
				}
			case "ctrl+k":
				d.pastedOffY, d.pastedOffX = 0, 0
				d.pastedPreview = !d.pastedPreview
			}
		case tea.MouseWheelMsg:
			switch mt.Mouse().Button {
			case tea.MouseWheelUp:
				if d.pastedOffY > 0 {
					d.pastedOffY--
				}
			case tea.MouseWheelDown:
				d.pastedOffY++
			}
		case tea.PasteMsg:
			s := strings.ReplaceAll(mt.Content, "\r", "\n")
			d.pastedRleError = nil
			d.pastedPattern = nil
			d.pastedOffY, d.pastedOffX = 0, 0
			d.pastedPreview = false
			if pattern, err := patterns.NewPatternFromRle(strings.NewReader(s)); err == nil {
				d.pastedPreview = true
				d.pastedRle = &s
				d.pastedPattern = &pattern
			} else {
				d.pastedRleError = err
			}
		}
	}
	return nil
}

func (p *patternRleInput[T]) Reset(parent T) {
}

func (p *patternRleInput[T]) Render(parent T, form *layout.Form[T], inputNo int, sf layout.Surface, clickPts layout.ClickPoints[T], row, col int, focused bool, style lipgloss.Style, focusedStyle lipgloss.Style) *tea.Cursor {
	rgn := sf.Region(row, col, 9, sf.Width()-2)
	if !focused {
		clickPts.AddRegion(rgn, func(d T) tea.Cmd {
			form.SetFocusedInput(inputNo)
			return nil
		})
	}
	d := asPatternsDialog(parent)
	s := style
	if focused {
		s = focusedStyle
	}
	if d.pastedRleError != nil {
		rgn.TextCenter(3, 0, rgn.Width(), " Error ", s)
		rgn.TextWrapped(4, 1, rgn.Width()-1, d.pastedRleError.Error(), style)
		return nil
	} else if d.pastedRle == nil {
		rgn.TextCenter(4, 0, rgn.Width(), "Paste RLE here", s)
		return nil
	}
	ht := rgn.Height()
	if focused {
		ht--
	}
	if d.pastedPreview && d.pastedPattern != nil {
		d.pastedOffY, d.pastedOffX = clampOffsets(d.pastedOffY, d.pastedOffX, min(ht, d.pastedPattern.Height), rgn.Width(), d.pastedPattern.Height, d.pastedPattern.Width)
		if rgn2 := rgn.Region(0, 0, min(ht, d.pastedPattern.Height), d.pastedPattern.Width); rgn2 != nil {
			useStyle := dialogPreviewStyle
			if focused {
				useStyle = dialogPreviewFocusedStyle
			}
			rgn2.FillWith(0, 0, rgn2.Height(), rgn2.Width(), '\u00A0', useStyle)
			aliveStyle := lipgloss.NewStyle().Foreground(useStyle.GetBackground()).Background(useStyle.GetForeground())
			d.pastedPattern.DrawTo(patterns.Rotate0, func(row, col int, alive bool) {
				if alive {
					ar, ac := row-d.pastedOffY, col-d.pastedOffX
					if ar >= 0 && ac >= 0 {
						rgn2.Text(ar, ac, "\u00A0", aliveStyle)
					}
				}
			})
		}
	} else if d.pastedRle != nil {
		lines := strings.Split(*d.pastedRle, "\n")
		maxLineWd := 0
		for _, line := range lines {
			if l := utf8.RuneCountInString(line); l > maxLineWd {
				maxLineWd = len(line)
			}
		}
		d.pastedOffY, d.pastedOffX = clampOffsets(d.pastedOffY, d.pastedOffX, min(ht, len(lines)), rgn.Width(), len(lines), maxLineWd)
		rgn2 := rgn.Region(0, 0, min(ht, len(lines)), rgn.Width())
		rgn2.Fill(0, 0, rgn2.Height(), rgn2.Width(), s)
		for i, line := range lines {
			rgn2.Text(i-d.pastedOffY, 0-d.pastedOffX, line, s)
		}
	}
	if focused {
		rgn.TextRight(8, 0, rgn.Width(), "Press enter to load ", style)
		clickPts.Add(rgn.Text(8, rgn.Width()-14, "enter", dialogTextUlStyle), func(t T) tea.Cmd {
			d.loadPastedRle()
			return nil
		})
		clickPts.Add(rgn.Text(8, 0, "ctrl+k", dialogTextUlStyle), func(t T) tea.Cmd {
			d.pastedOffY, d.pastedOffX = 0, 0
			d.pastedPreview = !d.pastedPreview
			return nil
		})
		show := "show preview"
		if d.pastedPreview {
			show = "show RLE"
		}
		rgn.Text(8, 7, show, dialogTextStyle)
	}
	return nil
}

func (d *patternsDialog) getPatterns() []string {
	if len(d.sortedPatterns) != len(patterns.PatternLibrary) {
		d.sortedPatterns = make([]string, 0, len(patterns.PatternLibrary))
		for name := range patterns.PatternLibrary {
			d.sortedPatterns = append(d.sortedPatterns, name)
		}
		slices.SortStableFunc(d.sortedPatterns, func(a, b string) int {
			return strings.Compare(strings.ToLower(a), strings.ToLower(b))
		})
	}
	if len(d.patternRuleFilter) == 0 {
		return d.sortedPatterns
	}
	result := make([]string, 0, len(d.sortedPatterns))
	for _, n := range d.sortedPatterns {
		if p, ok := patterns.PatternLibrary[n]; ok && p.Rule != nil && p.Rule.Rle() == d.patternRuleFilter {
			result = append(result, n)
		}
	}
	return result
}

func (d *patternsDialog) update(msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		switch mt.String() {
		case "ctrl+c":
			return tea.Quit
		case "esc":
			d.m.stopped()
		case "ctrl+p":
			d.tab = patternsTabPatterns
		case "ctrl+o":
			d.tab = patternsTabLoad
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
	case loadPatternsResult:
		d.loadPatternsResult = &mt
		if mt.error == "" && mt.loaded > 0 {
			if mt.filename != "" {
				d.m.prefs.addPattern(mt.filename)
			} else {
				d.m.prefs.addPatternLibrary(mt.directory)
			}
			return d.m.savePrefs()
		}
	default:
		if d.currentForm != nil {
			return d.currentForm.Update(d, msg)
		}
	}
	return nil
}

func asPatternsDialog(parent any) *patternsDialog {
	return parent.(*patternsDialog)
}

type loadPatternsResult struct {
	error     string
	loaded    int
	directory string
	filename  string
}

func loadPatternsLibrary(fp string) loadPatternsResult {
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
		p.Filename = filepath.Base(fp)
		patterns.PatternLibrary[p.Name] = p
		return nil
	}
}
