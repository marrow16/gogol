package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"errors"
	"github.com/marrow16/gogol/cmd/tui/layout"
	"github.com/marrow16/gogol/patterns"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type captureStage int

const (
	captureStartMark captureStage = iota
	captureEndMark
	captureEditPattern
)

func (c captureStage) String() string {
	switch c {
	case captureStartMark:
		return "start-mark"
	case captureEndMark:
		return "end-mark"
	case captureEditPattern:
		return "edit-pattern"
	}
	return ""
}

type capture struct {
	m                        *model
	stage                    captureStage
	tab                      int
	currentForm              *layout.Form[*capture]
	clickPts                 layout.ClickPoints[*capture]
	startRow, startCol       int
	endRow, endCol           int
	resetStyle               *lipgloss.Style
	markedStyle              *lipgloss.Style
	cropTop, cropBottom      int
	cropLeft, cropRight      int
	pattern                  *patterns.Pattern
	patternOffY, patternOffX int
	cursorY, cursorX         int
	filename                 string
	addLibrary               int
	saveResult               *patternSaveResult
}

func (c *capture) start() {
	c.m.capturing = true
	c.stage = captureStartMark
	c.tab = captureTabDetails
	c.startRow, c.startCol = 0, 0
	c.endRow, c.endCol = 0, 0
	c.resetStyle = &c.m.cellStyle
	ms := lipgloss.NewStyle().Foreground(c.m.cellStyle.GetForeground()).Background(lipgloss.Color("#D00000"))
	c.markedStyle = &ms
	c.cropTop, c.cropBottom, c.cropLeft, c.cropRight = 0, 0, 0, 0
	c.patternOffY, c.patternOffX = 0, 0
	c.cursorY, c.cursorX = 0, 0
	c.pattern = nil
	c.saveResult = nil
	// anything else to reset???
}

const (
	captureDialogWidth  = 50
	captureDialogHeight = 22
)

func (c *capture) render(sf layout.Surface) *tea.Cursor {
	c.clickPts = layout.ClickPoints[*capture]{}
	var csr *tea.Cursor
	switch c.stage {
	case captureStartMark:
		csr = tea.NewCursor(c.startCol, c.startRow)
		csr.Color = lipgloss.Color("#00D000")
	case captureEndMark:
		csr = tea.NewCursor(c.endCol, c.endRow)
		csr.Color = lipgloss.Color("#00D000")
	default:
		rgn := sf.Region(1, (sf.Width()-captureDialogWidth)/2, captureDialogHeight, captureDialogWidth)
		rgn.FillWith(0, 0, rgn.Height(), rgn.Width(), '\u00A0', settingsBgStyle)
		rgn.BoxRounded(0, 0, rgn.Height(), rgn.Width(), settingsTextStyle)
		rgn.TextCenter(0, 0, rgn.Width(), "Pattern Capture", settingsTextStyle)
		c.renderTabs(rgn)
		c.currentForm = nil
		switch c.tab {
		case captureTabDetails:
			c.currentForm = captureDetailsForm
		case captureTabModify:
			c.currentForm = captureModifyForm
		}
		if c.currentForm != nil {
			csr = c.currentForm.Render(c, c.clickPts, rgn)
		}
	}
	return csr
}

var captureDetailsForm = &layout.Form[*capture]{
	Style:        settingsTextStyle,
	FocusedStyle: settingsTabStyle,
	FormRows: layout.FormRows[*capture]{
		3: {
			3: {Item: "File:"},
			9: {
				Item: layout.NewTextInput(captureDialogWidth-10, "",
					func(c *capture) string {
						return c.filename
					}, func(c *capture, value string) tea.Cmd {
						c.filename = value
						return nil
					}, true),
			},
		},
		4: {
			3: {Item: "Name:"},
			9: {
				Item: layout.NewTextInput(captureDialogWidth-10, "",
					func(c *capture) string {
						return c.pattern.Name
					}, func(c *capture, value string) tea.Cmd {
						c.pattern.Name = value
						return nil
					}, true),
			},
		},
		5: {
			1: {Item: "Origin:"},
			9: {
				Item: layout.NewTextInput(captureDialogWidth-10, "",
					func(c *capture) string {
						return c.pattern.Origination
					}, func(c *capture, value string) tea.Cmd {
						c.pattern.Origination = value
						return nil
					}, true),
			},
		},
		7: {
			1: {Item: "Comments:"},
		},
		8: {
			1: {
				Item: layout.NewTextInput(captureDialogWidth-2, "",
					func(c *capture) string {
						if len(c.pattern.Comments) > 0 {
							return c.pattern.Comments[0]
						}
						return ""
					}, func(c *capture, value string) tea.Cmd {
						c.setComment(0, value)
						return nil
					}, true),
			},
		},
		9: {
			1: {
				Item: layout.NewTextInput(captureDialogWidth-2, "",
					func(c *capture) string {
						if len(c.pattern.Comments) > 1 {
							return c.pattern.Comments[1]
						}
						return ""
					}, func(c *capture, value string) tea.Cmd {
						c.setComment(1, value)
						return nil
					}, true),
			},
		},
		10: {
			1: {
				Item: layout.NewTextInput(captureDialogWidth-2, "",
					func(c *capture) string {
						if len(c.pattern.Comments) > 2 {
							return c.pattern.Comments[2]
						}
						return ""
					}, func(c *capture, value string) tea.Cmd {
						c.setComment(2, value)
						return nil
					}, true),
			},
		},
		11: {
			1: {
				Item: layout.NewTextInput(captureDialogWidth-2, "",
					func(c *capture) string {
						if len(c.pattern.Comments) > 3 {
							return c.pattern.Comments[3]
						}
						return ""
					}, func(c *capture, value string) tea.Cmd {
						c.setComment(3, value)
						return nil
					}, true),
			},
		},
		captureDialogHeight - 6: {
			1: {
				Item: func(c *capture) any {
					if c.saveResult != nil && c.saveResult.error != nil {
						return c.saveResult.error.Error()
					}
					return ""
				},
				Style: &errorStyle,
			},
		},
		captureDialogHeight - 5: {
			captureDialogWidth - 21: {Item: "Add pattern:"},
			captureDialogWidth - 8: {
				Item: layout.NewRadio([]string{"Yes", "No"}, func(c *capture) int {
					return c.addLibrary
				}, func(c *capture, value int) tea.Cmd {
					c.addLibrary = value
					return nil
				}),
			},
		},
		captureDialogHeight - 3: {
			captureDialogWidth - 6: {
				Item: layout.NewButton("Save", func(c *capture) tea.Cmd {
					c.saveResult = nil
					return c.save()
				}),
			},
		},
	},
}

var errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))

func (c *capture) save() tea.Cmd {
	return func() tea.Msg {
		fn := c.filename
		if !strings.HasSuffix(fn, ".rle") {
			fn += ".rle"
		}
		f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			if errors.Is(err, fs.ErrExist) {
				return patternSaveResult{
					error: errors.New("File already exists"),
				}
			} else {
				return patternSaveResult{
					error: err,
				}
			}
		}
		defer func() {
			_ = f.Close()
		}()
		p := c.croppedPattern()
		err = patterns.PatternRleEncode(*p, f)
		return patternSaveResult{
			error:    err,
			pattern:  p,
			filename: fn,
		}
	}
}

func (c *capture) setComment(n int, value string) {
	for len(c.pattern.Comments) < n+1 {
		c.pattern.Comments = append(c.pattern.Comments, "")
	}
	c.pattern.Comments[n] = value
}

var captureModifyForm = &layout.Form[*capture]{
	Style:        settingsTextStyle,
	FocusedStyle: settingsTabStyle,
	FormRows: layout.FormRows[*capture]{
		2: {
			1: {
				Item: &capturePatternPreview[*capture]{},
			},
		},
		captureDialogHeight - 2: {
			2: {Item: "Crop  Top:     Left:     Bottom:     Right:"},
			12: {
				Item: layout.NewNumberInput(3, 0,
					func(c *capture) int {
						return c.pattern.Height - c.cropBottom - 1
					},
					func(c *capture) int {
						return c.cropTop
					}, func(c *capture, value int) tea.Cmd {
						if c.pattern.Height-value-c.cropBottom >= 1 {
							c.patternOffY, c.patternOffX = 0, 0
							c.cursorY, c.cursorX = 0, 0
							c.cropTop = value
						}
						return nil
					}),
			},
			22: {
				Item: layout.NewNumberInput(3, 0,
					func(c *capture) int {
						return c.pattern.Width - c.cropRight - 1
					},
					func(c *capture) int {
						return c.cropLeft
					}, func(c *capture, value int) tea.Cmd {
						if c.pattern.Width-value-c.cropRight >= 1 {
							c.patternOffY, c.patternOffX = 0, 0
							c.cursorY, c.cursorX = 0, 0
							c.cropLeft = value
						}
						return nil
					}),
			},
			34: {
				Item: layout.NewNumberInput(3, 0,
					func(c *capture) int {
						return c.pattern.Height - c.cropTop - 1
					},
					func(c *capture) int {
						return c.cropBottom
					}, func(c *capture, value int) tea.Cmd {
						if c.pattern.Height-c.cropTop-value >= 1 {
							c.patternOffY, c.patternOffX = 0, 0
							c.cursorY, c.cursorX = 0, 0
							c.cropBottom = value
						}
						return nil
					}),
			},
			45: {
				Item: layout.NewNumberInput(3, 0,
					func(c *capture) int {
						return c.pattern.Width - c.cropLeft - 1
					},
					func(c *capture) int {
						return c.cropRight
					}, func(c *capture, value int) tea.Cmd {
						if c.pattern.Width-c.cropLeft-value >= 1 {
							c.patternOffY, c.patternOffX = 0, 0
							c.cursorY, c.cursorX = 0, 0
							c.cropRight = value
						}
						return nil
					}),
			},
		},
	},
}

type capturePatternPreview[T any] struct {
	preview layout.Surface
}

func (p *capturePatternPreview[T]) Update(parent T, msg tea.Msg, focused bool) tea.Cmd {
	if focused && p.preview != nil {
		c := asCapture(parent)
		switch mt := msg.(type) {
		case tea.KeyPressMsg:
			switch mt.String() {
			case "up":
				if c.cursorY > 0 {
					c.cursorY--
				} else if c.patternOffY > 0 {
					c.patternOffY--
				}
			case "down":
				if c.cursorY+1 < p.preview.Height() {
					c.cursorY++
				} else if maxOffY := c.pattern.Height - c.cropTop - c.cropBottom - p.preview.Height(); maxOffY > 0 && c.patternOffY < maxOffY {
					c.patternOffY++
				}
			case "left":
				if c.cursorX > 0 {
					c.cursorX--
				} else if c.patternOffX > 0 {
					c.patternOffX--
				}
			case "right":
				if c.cursorX+1 < p.preview.Width() {
					c.cursorX++
				} else if maxOffX := c.pattern.Width - c.cropLeft - c.cropRight - p.preview.Width(); maxOffX > 0 && c.patternOffX < maxOffX {
					c.patternOffX++
				}
			case "home":
				c.cursorX = 0
			case "end":
				c.cursorX = c.pattern.Width
			case "backspace":
				y, x := c.cursorY+c.patternOffY+c.cropTop, c.cursorX+c.patternOffX+c.cropLeft
				if idx := (y * c.pattern.Width) + x; idx < len(c.pattern.Cells) {
					c.pattern.Cells[idx] = false
					if c.cursorX > 0 {
						c.cursorX--
					} else if c.patternOffX > 0 {
						c.patternOffX--
					}
				}
			case "ctrl+space":
				y, x := c.cursorY+c.patternOffY+c.cropTop, c.cursorX+c.patternOffX+c.cropLeft
				if idx := (y * c.pattern.Width) + x; idx < len(c.pattern.Cells) {
					c.pattern.Cells[idx] = true
					if c.cursorX+1 < p.preview.Width() {
						c.cursorX++
					} else if maxOffX := c.pattern.Width - c.cropLeft - c.cropRight - p.preview.Width(); maxOffX > 0 && c.patternOffX < maxOffX {
						c.patternOffX++
					}
				}
			case "space":
				y, x := c.cursorY+c.patternOffY+c.cropTop, c.cursorX+c.patternOffX+c.cropLeft
				if idx := (y * c.pattern.Width) + x; idx < len(c.pattern.Cells) {
					c.pattern.Cells[idx] = false
					if c.cursorX+1 < p.preview.Width() {
						c.cursorX++
					} else if maxOffX := c.pattern.Width - c.cropLeft - c.cropRight - p.preview.Width(); maxOffX > 0 && c.patternOffX < maxOffX {
						c.patternOffX++
					}
				}
			}
		}
	}
	return nil
}

func (p *capturePatternPreview[T]) Render(parent T, form *layout.Form[T], inputNo int, sf layout.Surface, clickPts layout.ClickPoints[T], row, col int, focused bool, style lipgloss.Style, focusedStyle lipgloss.Style) *tea.Cursor {
	rgn := sf.Region(row, col, captureDialogHeight-4, sf.Width()-2)
	c := asCapture(parent)
	wd := c.pattern.Width - c.cropLeft - c.cropRight - c.patternOffX
	if wd > rgn.Width() {
		wd = rgn.Width()
	}
	ht := c.pattern.Height - c.cropTop - c.cropBottom - c.patternOffY
	if ht > rgn.Height() {
		ht = rgn.Height()
	}
	var csr *tea.Cursor
	if p.preview = rgn.Region(0, 0, ht, wd); p.preview != nil {
		if !focused {
			for y := 0; y < ht; y++ {
				clickPts.Add(layout.Placement{
					Extent: wd,
					Row:    p.preview.AbsoluteTop() + y,
					Col:    p.preview.AbsoluteLeft(),
				}, func(parent T) tea.Cmd {
					form.SetFocusedInput(inputNo)
					return nil
				})
			}
		}
		useStyle := settingsPreviewStyle
		if focused {
			useStyle = settingsPreviewFocusedStyle
		}
		p.preview.FillWith(0, 0, p.preview.Height(), p.preview.Width(), '\u00A0', useStyle)
		aliveStyle := lipgloss.NewStyle().Foreground(useStyle.GetBackground()).Background(useStyle.GetForeground())
		c.pattern.DrawTo(patterns.Rotate0, func(row, col int, alive bool) {
			if alive && row >= c.cropTop && col >= c.cropLeft {
				ar, ac := row-c.cropTop-c.patternOffY, col-c.cropLeft-c.patternOffX
				if ar >= 0 && ac >= 0 {
					p.preview.Text(ar, ac, "\u00A0", aliveStyle)
				}
			}
		})
		if focused {
			csr = tea.NewCursor(p.preview.AbsoluteLeft()+c.cursorX, p.preview.AbsoluteTop()+c.cursorY)
			csr.Color = lipgloss.Color("#00D000")
		}
	}
	return csr
}

func asCapture(parent any) *capture {
	return parent.(*capture)
}

func (c *capture) renderTabs(rgn layout.Surface) {
	x := 3
	for _, tab := range captureTabs {
		if tab.tabNo == c.tab {
			rgn.Text(1, x-1, " "+tab.title+" ", settingsTabStyle)
		} else {
			c.clickPts.Add(rgn.Text(1, x, tab.title, settingsTextStyle), func(c *capture) tea.Cmd {
				c.tab = tab.tabNo
				return nil
			})
			if tab.ul != -1 {
				rgn.Text(1, x+tab.ul, tab.title[tab.ul:tab.ul+1], settingsTextUlStyle)
			}
		}
		x += len(tab.title) + 3
	}
}

const (
	captureTabDetails = iota
	captureTabModify
)

var captureTabs = []struct {
	title string
	ul    int
	tabNo int
}{
	{"Details", 1, captureTabDetails},
	{"Modify", 1, captureTabModify},
}

func (c *capture) resetArea() {
	sr, sc, er, ec := c.startRow, c.startCol, c.endRow, c.endCol
	if er < sr {
		sr, er = er, sr
	}
	if ec < sc {
		sc, ec = ec, sc
	}
	for row := sr; row <= er; row++ {
		for col := sc; col <= ec; col++ {
			c.m.gridSurface.SetStyle(row, col, c.resetStyle)
		}
	}
}

func (c *capture) markArea() {
	sr, sc, er, ec := c.startRow, c.startCol, c.endRow, c.endCol
	if er < sr {
		sr, er = er, sr
	}
	if ec < sc {
		sc, ec = ec, sc
	}
	for row := sr; row <= er; row++ {
		for col := sc; col <= ec; col++ {
			c.m.gridSurface.SetStyle(row, col, c.markedStyle)
		}
	}
}

func (c *capture) scrapePattern() {
	sr, sc, er, ec := c.startRow, c.startCol, c.endRow, c.endCol
	if er < sr {
		sr, er = er, sr
	}
	if ec < sc {
		sc, ec = ec, sc
	}
	sr, sc, er, ec = sr*2, sc*2, (er*2)+1, (ec*2)+1
	ht, wd := er-sr+1, ec-sc+1
	idx := 0
	cells := make([]bool, ht*wd)
	for y := sr; y <= er; y++ {
		for x := sc; x <= ec; x++ {
			if cell := c.m.grid.GetCell(y, x); cell != nil {
				cells[idx] = cell.Alive
			} else {
				panic("Cell not found - shouldn't happen")
			}
			idx++
		}
	}
	now := time.Now().Format("2006-01-02T1504")
	if c.filename == "" {
		c.filename = now
		if path := c.m.prefs.SavePath; path != "" {
			c.filename = filepath.Join(path, c.filename)
		}
	}
	pt := patterns.MustNewPattern(now, wd, cells)
	c.pattern = &pt
	c.pattern.Origination = c.m.prefs.Originator
	if c.pattern.Origination == "" {
		c.pattern.Origination = "(your name)"
	}
	c.pattern.Rule = c.m.grid.Rule
	c.pattern.Comments = []string{"Captured from GoGoL (https://github.com/marrow16/gogol)"}
}

func (c *capture) update(msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case patternSaveResult:
		if mt.error != nil {
			c.saveResult = &mt
		} else {
			if c.stage > captureStartMark {
				c.resetArea()
			}
			c.m.capturing = false
			c.m.prefs.Originator = mt.pattern.Origination
			c.m.prefs.SavePath = filepath.Dir(mt.filename)
			c.m.prefs.addPattern(mt.filename)
			if c.addLibrary == 0 {
				patterns.PatternLibrary[mt.pattern.Name] = *mt.pattern
				sortedPatterns = sortPatterns()
			}
			return c.m.savePrefs()
		}
	case tea.KeyPressMsg:
		switch mt.String() {
		case "ctrl+c":
			return tea.Quit
		case "esc", "ctrl+k":
			if c.stage > captureStartMark {
				c.resetArea()
			}
			c.m.capturing = false
			return nil
		}
	}
	switch c.stage {
	case captureStartMark:
		return c.updateStart(msg)
	case captureEndMark:
		return c.updateEnd(msg)
	case captureEditPattern:
		return c.updateEdit(msg)
	}
	return nil
}

func (c *capture) updateStart(msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		switch mt.String() {
		case "up":
			if c.startRow > 0 {
				c.startRow--
			}
		case "down":
			if c.startRow*2 < c.m.grid.Height-2 {
				c.startRow++
			}
		case "left":
			if c.startCol > 0 {
				c.startCol--
			}
		case "right":
			if c.startCol*2 < c.m.grid.Width-2 {
				c.startCol++
			}
		case "home":
			c.startRow, c.startCol = 0, 0
		case "end":
			c.startRow, c.startCol = (c.m.grid.Height/2)-1, (c.m.grid.Width/2)-1
		case "space", "enter":
			c.endRow, c.endCol = c.startRow, c.startCol
			c.markArea()
			c.stage = captureEndMark
		}
	case tea.MouseClickMsg:
		mmsg := mt.Mouse()
		if mmsg.Y*2 < c.m.grid.Height && mmsg.X*2 < c.m.grid.Width {
			c.startRow, c.startCol = mmsg.Y, mmsg.X
		}
	}
	return nil
}

func (c *capture) updateEnd(msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		switch mt.String() {
		case "up":
			if c.endRow > 0 {
				c.resetArea()
				c.endRow--
				c.markArea()
			}
		case "down":
			if c.endRow*2 < c.m.grid.Height-2 {
				c.resetArea()
				c.endRow++
				c.markArea()
			}
		case "left":
			if c.endCol > 0 {
				c.resetArea()
				c.endCol--
				c.markArea()
			}
		case "right":
			if c.endCol*2 < c.m.grid.Width-2 {
				c.resetArea()
				c.endCol++
				c.markArea()
			}
		case "home":
			c.resetArea()
			c.endRow, c.endCol = 0, 0
			c.markArea()
		case "end":
			c.resetArea()
			c.endRow, c.endCol = (c.m.grid.Height/2)-1, (c.m.grid.Width/2)-1
			c.markArea()
		case "space", "enter":
			c.scrapePattern()
			c.stage = captureEditPattern
		}
	case tea.MouseClickMsg:
		mmsg := mt.Mouse()
		if mmsg.Y*2 < c.m.grid.Height && mmsg.X*2 < c.m.grid.Width {
			c.resetArea()
			c.endRow, c.endCol = mmsg.Y, mmsg.X
			c.markArea()
		}
	}
	return nil
}

func (c *capture) updateEdit(msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		switch mt.String() {
		case "ctrl+e":
			c.tab = captureTabDetails
		case "ctrl+o":
			c.tab = captureTabModify
		default:
			if c.currentForm != nil {
				return c.currentForm.Update(c, msg)
			}
		}
	case tea.MouseClickMsg:
		if fn, ok := c.clickPts[layout.ClickPoint{mt.Y, mt.X}]; ok {
			return fn(c)
		} else if c.currentForm != nil {
			return c.currentForm.Update(c, msg)
		}
	default:
		if c.currentForm != nil {
			return c.currentForm.Update(c, msg)
		}
	}
	return nil
}

func (c *capture) croppedPattern() *patterns.Pattern {
	newWidth := c.pattern.Width - c.cropLeft - c.cropRight
	newHeight := c.pattern.Height - c.cropTop - c.cropBottom
	cells := make([]bool, 0, newWidth*newHeight)
	for y := c.cropTop; y < c.pattern.Height-c.cropBottom; y++ {
		from := y*c.pattern.Width + c.cropLeft
		to := from + newWidth
		cells = append(cells, c.pattern.Cells[from:to]...)
	}
	return &patterns.Pattern{
		Name:        c.pattern.Name,
		Width:       newWidth,
		Height:      newHeight,
		Cells:       cells,
		Comments:    c.pattern.Comments,
		Origination: c.pattern.Origination,
		Coordinates: c.pattern.Coordinates,
		Rule:        c.pattern.Rule,
	}
}

type patternSaveResult struct {
	error    error
	pattern  *patterns.Pattern
	filename string
}
