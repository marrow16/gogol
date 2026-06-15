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

const (
	captureTabDetails = iota
	captureTabModify
)

var captureTabs = tabs{
	{"Details", 1, captureTabDetails},
	{"Modify", 1, captureTabModify},
}

func (c captureStage) String() string {
	switch c {
	case captureStartMark:
		return "capture start - space/enter"
	case captureEndMark:
		return "capture end - space/enter"
	case captureEditPattern:
		return "capture"
	}
	return ""
}

type captureDialog struct {
	m                        *model
	stage                    captureStage
	tab                      int
	currentForm              *layout.Form[*captureDialog]
	clickPts                 layout.ClickPoints[*captureDialog]
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

func (d *captureDialog) title() string {
	return "[" + d.stage.String() + "]"
}

func (d *captureDialog) start() {
	d.stage = captureStartMark
	d.tab = captureTabDetails
	d.startRow, d.startCol = 0, 0
	d.endRow, d.endCol = 0, 0
	d.resetStyle = &d.m.cellStyle
	ms := lipgloss.NewStyle().Foreground(d.m.cellStyle.GetForeground()).Background(lipgloss.Color("#D00000"))
	d.markedStyle = &ms
	d.cropTop, d.cropBottom, d.cropLeft, d.cropRight = 0, 0, 0, 0
	d.patternOffY, d.patternOffX = 0, 0
	d.cursorY, d.cursorX = 0, 0
	d.pattern = nil
	d.saveResult = nil
	// anything else to reset???
}

func (d *captureDialog) render(sf layout.Surface) *tea.Cursor {
	d.clickPts = layout.ClickPoints[*captureDialog]{}
	var csr *tea.Cursor
	switch d.stage {
	case captureStartMark:
		csr = tea.NewCursor(d.startCol, d.startRow)
		csr.Color = cursorColor
	case captureEndMark:
		csr = tea.NewCursor(d.endCol, d.endRow)
		csr.Color = cursorColor
	default:
		topPos, leftPos := d.positionDialog(sf)
		rgn := sf.Region(topPos, leftPos, dialogHeight, dialogWidth)
		rgn.FillWith(0, 0, rgn.Height(), rgn.Width(), '\u00A0', dialogBgStyle)
		rgn.BoxRounded(0, 0, rgn.Height(), rgn.Width(), dialogTextStyle)
		rgn.TextCenter(0, 0, rgn.Width(), "Pattern Capture", dialogTextStyle)
		renderTabs(rgn, d.clickPts, captureTabs, d.tab, func(t int) {
			d.tab = t
		})
		d.currentForm = nil
		switch d.tab {
		case captureTabDetails:
			d.currentForm = captureDetailsForm
		case captureTabModify:
			d.currentForm = captureModifyForm
		}
		if d.currentForm != nil {
			csr = d.currentForm.Render(d, d.clickPts, rgn)
		}
	}
	return csr
}

func (d *captureDialog) positionDialog(sf layout.Surface) (topPos, leftPos int) {
	sr, sc, er, ec := d.normalizeDimensions()
	maxLeft := max(0, sf.Width()-dialogWidth)
	topPos = clamp(sr, 0, max(0, sf.Height()-dialogHeight))
	leftPos = clamp((sf.Width()-dialogWidth)/2, 0, maxLeft)
	switch {
	case ec+2+dialogWidth <= sf.Width():
		// right of captureDialog
		leftPos = ec + 2
	case sc-2-dialogWidth >= 0:
		// left of captureDialog
		leftPos = sc - 2 - dialogWidth
	case er+1+dialogHeight <= sf.Height():
		// below captureDialog
		topPos = er + 1
		leftPos = clamp(sc, 0, maxLeft)
	case sr-1-dialogHeight >= 0:
		// above captureDialog
		topPos = sr - 1 - dialogHeight
		leftPos = clamp(sc, 0, maxLeft)
	default:
		// nowhere ideal - fallback centered
		topPos, leftPos = d.m.dialogPosition(dialogHeight, dialogWidth)
	}
	return topPos, leftPos
}

var captureDetailsForm = &layout.Form[*captureDialog]{
	Style:        dialogTextStyle,
	FocusedStyle: dialogTabStyle,
	FormRows: layout.FormRows[*captureDialog]{
		3: {
			3: {Item: "File:"},
			9: {
				Item: layout.NewTextInput(dialogWidth-10, "",
					func(c *captureDialog) string {
						return c.filename
					}, func(c *captureDialog, value string) tea.Cmd {
						c.filename = value
						return nil
					}, true),
			},
		},
		4: {
			3: {Item: "Name:"},
			9: {
				Item: layout.NewTextInput(dialogWidth-10, "",
					func(c *captureDialog) string {
						return c.pattern.Name
					}, func(c *captureDialog, value string) tea.Cmd {
						c.pattern.Name = value
						return nil
					}, true),
			},
		},
		5: {
			1: {Item: "Origin:"},
			9: {
				Item: layout.NewTextInput(dialogWidth-10, "",
					func(c *captureDialog) string {
						return c.pattern.Origination
					}, func(c *captureDialog, value string) tea.Cmd {
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
				Item: layout.NewTextInput(dialogWidth-2, "",
					func(c *captureDialog) string {
						if len(c.pattern.Comments) > 0 {
							return c.pattern.Comments[0]
						}
						return ""
					}, func(c *captureDialog, value string) tea.Cmd {
						c.setComment(0, value)
						return nil
					}, true),
			},
		},
		9: {
			1: {
				Item: layout.NewTextInput(dialogWidth-2, "",
					func(c *captureDialog) string {
						if len(c.pattern.Comments) > 1 {
							return c.pattern.Comments[1]
						}
						return ""
					}, func(c *captureDialog, value string) tea.Cmd {
						c.setComment(1, value)
						return nil
					}, true),
			},
		},
		10: {
			1: {
				Item: layout.NewTextInput(dialogWidth-2, "",
					func(c *captureDialog) string {
						if len(c.pattern.Comments) > 2 {
							return c.pattern.Comments[2]
						}
						return ""
					}, func(c *captureDialog, value string) tea.Cmd {
						c.setComment(2, value)
						return nil
					}, true),
			},
		},
		11: {
			1: {
				Item: layout.NewTextInput(dialogWidth-2, "",
					func(c *captureDialog) string {
						if len(c.pattern.Comments) > 3 {
							return c.pattern.Comments[3]
						}
						return ""
					}, func(c *captureDialog, value string) tea.Cmd {
						c.setComment(3, value)
						return nil
					}, true),
			},
		},
		dialogHeight - 6: {
			1: {
				Item: func(c *captureDialog) any {
					if c.saveResult != nil && c.saveResult.error != nil {
						return c.saveResult.error.Error()
					}
					return ""
				},
				Style: &errorStyle,
			},
		},
		dialogHeight - 5: {
			dialogWidth - 21: {Item: "Add pattern:"},
			dialogWidth - 8: {
				Item: layout.NewRadio([]string{"Yes", "No"}, func(c *captureDialog) int {
					return c.addLibrary
				}, func(c *captureDialog, value int) tea.Cmd {
					c.addLibrary = value
					return nil
				}),
			},
		},
		dialogHeight - 3: {
			dialogWidth - 6: {
				Item: layout.NewButton("Save", func(c *captureDialog) tea.Cmd {
					c.saveResult = nil
					return c.save()
				}),
			},
		},
	},
}

var errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))

func (d *captureDialog) save() tea.Cmd {
	return func() tea.Msg {
		fn := d.filename
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
		p := d.croppedPattern()
		err = patterns.PatternRleEncode(*p, f)
		return patternSaveResult{
			error:    err,
			pattern:  p,
			filename: fn,
		}
	}
}

func (d *captureDialog) setComment(n int, value string) {
	for len(d.pattern.Comments) < n+1 {
		d.pattern.Comments = append(d.pattern.Comments, "")
	}
	d.pattern.Comments[n] = value
}

var captureModifyForm = &layout.Form[*captureDialog]{
	Style:        dialogTextStyle,
	FocusedStyle: dialogTabStyle,
	FormRows: layout.FormRows[*captureDialog]{
		2: {
			1: {
				Item: &capturePatternPreview[*captureDialog]{},
			},
		},
		dialogHeight - 2: {
			2: {Item: "Crop  Top:     Left:     Bottom:     Right:"},
			12: {
				Item: layout.NewNumberInput(3, 0,
					func(c *captureDialog) int {
						return c.pattern.Height - c.cropBottom - 1
					},
					func(c *captureDialog) int {
						return c.cropTop
					}, func(c *captureDialog, value int) tea.Cmd {
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
					func(c *captureDialog) int {
						return c.pattern.Width - c.cropRight - 1
					},
					func(c *captureDialog) int {
						return c.cropLeft
					}, func(c *captureDialog, value int) tea.Cmd {
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
					func(c *captureDialog) int {
						return c.pattern.Height - c.cropTop - 1
					},
					func(c *captureDialog) int {
						return c.cropBottom
					}, func(c *captureDialog, value int) tea.Cmd {
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
					func(c *captureDialog) int {
						return c.pattern.Width - c.cropLeft - 1
					},
					func(c *captureDialog) int {
						return c.cropRight
					}, func(c *captureDialog, value int) tea.Cmd {
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

func (p *capturePatternPreview[T]) Reset(parent T) {
}

func (p *capturePatternPreview[T]) Update(parent T, msg tea.Msg, focused bool) tea.Cmd {
	if focused && p.preview != nil {
		d := asCaptureDialog(parent)
		switch mt := msg.(type) {
		case tea.KeyPressMsg:
			switch mt.String() {
			case "up":
				if d.cursorY > 0 {
					d.cursorY--
				} else if d.patternOffY > 0 {
					d.patternOffY--
				}
			case "down":
				if d.cursorY+1 < p.preview.Height() {
					d.cursorY++
				} else if maxOffY := d.pattern.Height - d.cropTop - d.cropBottom - p.preview.Height(); maxOffY > 0 && d.patternOffY < maxOffY {
					d.patternOffY++
				}
			case "left":
				if d.cursorX > 0 {
					d.cursorX--
				} else if d.patternOffX > 0 {
					d.patternOffX--
				}
			case "right":
				if d.cursorX+1 < p.preview.Width() {
					d.cursorX++
				} else if maxOffX := d.pattern.Width - d.cropLeft - d.cropRight - p.preview.Width(); maxOffX > 0 && d.patternOffX < maxOffX {
					d.patternOffX++
				}
			case "home":
				d.cursorX = 0
			case "end":
				d.cursorX = d.pattern.Width
			case "backspace":
				y, x := d.cursorY+d.patternOffY+d.cropTop, d.cursorX+d.patternOffX+d.cropLeft
				if idx := (y * d.pattern.Width) + x; idx < len(d.pattern.Cells) {
					d.pattern.Cells[idx] = false
					if d.cursorX > 0 {
						d.cursorX--
					} else if d.patternOffX > 0 {
						d.patternOffX--
					}
				}
			case "ctrl+space":
				y, x := d.cursorY+d.patternOffY+d.cropTop, d.cursorX+d.patternOffX+d.cropLeft
				if idx := (y * d.pattern.Width) + x; idx < len(d.pattern.Cells) {
					d.pattern.Cells[idx] = true
					if d.cursorX+1 < p.preview.Width() {
						d.cursorX++
					} else if maxOffX := d.pattern.Width - d.cropLeft - d.cropRight - p.preview.Width(); maxOffX > 0 && d.patternOffX < maxOffX {
						d.patternOffX++
					}
				}
			case "space":
				y, x := d.cursorY+d.patternOffY+d.cropTop, d.cursorX+d.patternOffX+d.cropLeft
				if idx := (y * d.pattern.Width) + x; idx < len(d.pattern.Cells) {
					d.pattern.Cells[idx] = false
					if d.cursorX+1 < p.preview.Width() {
						d.cursorX++
					} else if maxOffX := d.pattern.Width - d.cropLeft - d.cropRight - p.preview.Width(); maxOffX > 0 && d.patternOffX < maxOffX {
						d.patternOffX++
					}
				}
			}
		}
	}
	return nil
}

func (p *capturePatternPreview[T]) Render(parent T, form *layout.Form[T], inputNo int, sf layout.Surface, clickPts layout.ClickPoints[T], row, col int, focused bool, style lipgloss.Style, focusedStyle lipgloss.Style) *tea.Cursor {
	rgn := sf.Region(row, col, dialogHeight-4, sf.Width()-2)
	c := asCaptureDialog(parent)
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
		useStyle := dialogPreviewStyle
		if focused {
			useStyle = dialogPreviewFocusedStyle
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

func asCaptureDialog(parent any) *captureDialog {
	return parent.(*captureDialog)
}

func (d *captureDialog) normalizeDimensions() (startRow, startCol, endRow, endCol int) {
	startRow, startCol, endRow, endCol = d.startRow, d.startCol, d.endRow, d.endCol
	if endRow < startRow {
		startRow, endRow = endRow, startRow
	}
	if endCol < startCol {
		startCol, endCol = endCol, startCol
	}
	return startRow, startCol, endRow, endCol
}

func (d *captureDialog) resetArea() {
	sr, sc, er, ec := d.normalizeDimensions()
	for row := sr; row <= er; row++ {
		for col := sc; col <= ec; col++ {
			d.m.gridSurface.SetStyle(row, col, d.resetStyle)
		}
	}
}

func (d *captureDialog) markArea() {
	sr, sc, er, ec := d.normalizeDimensions()
	for row := sr; row <= er; row++ {
		for col := sc; col <= ec; col++ {
			d.m.gridSurface.SetStyle(row, col, d.markedStyle)
		}
	}
}

func (d *captureDialog) scrapePattern() {
	sr, sc, er, ec := d.normalizeDimensions()
	sr, sc, er, ec = sr*2, sc*2, (er*2)+1, (ec*2)+1
	ht, wd := er-sr+1, ec-sc+1
	idx := 0
	cells := make([]bool, ht*wd)
	for y := sr; y <= er; y++ {
		for x := sc; x <= ec; x++ {
			if cell := d.m.grid.GetCell(y, x); cell != nil {
				cells[idx] = cell.Alive
			} else {
				panic("Cell not found - shouldn't happen")
			}
			idx++
		}
	}
	now := time.Now().Format("2006-01-02T1504")
	if d.filename == "" {
		d.filename = now
		if path := d.m.prefs.SavePath; path != "" {
			d.filename = filepath.Join(path, d.filename)
		}
	}
	pt := patterns.MustNewPattern(now, wd, cells)
	d.pattern = &pt
	d.pattern.Origination = d.m.prefs.Originator
	if d.pattern.Origination == "" {
		d.pattern.Origination = "(your name)"
	}
	d.pattern.Rule = d.m.grid.Rule
	d.pattern.Comments = []string{"Captured from GoGoL (https://github.com/marrow16/gogol)"}
}

func (d *captureDialog) update(msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case patternSaveResult:
		if mt.error != nil {
			d.saveResult = &mt
		} else {
			if d.stage > captureStartMark {
				d.resetArea()
			}
			d.m.stopped()
			d.m.prefs.Originator = mt.pattern.Origination
			d.m.prefs.SavePath = filepath.Dir(mt.filename)
			d.m.prefs.addPattern(mt.filename)
			if d.addLibrary == 0 {
				mt.pattern.Filename = filepath.Base(mt.filename)
				patterns.PatternLibrary[mt.pattern.Name] = *mt.pattern
			}
			return d.m.savePrefs()
		}
	case tea.KeyPressMsg:
		switch mt.String() {
		case "ctrl+d":
			return tea.Quit
		case "esc", "ctrl+k":
			if d.stage == captureEndMark {
				d.resetArea()
				d.startRow, d.startCol = d.endRow, d.endCol
				d.stage = captureStartMark
				return nil
			}
			if d.stage > captureStartMark {
				d.resetArea()
			}
			d.m.stopped()
			return nil
		}
	}
	switch d.stage {
	case captureStartMark:
		return d.updateStart(msg)
	case captureEndMark:
		return d.updateEnd(msg)
	case captureEditPattern:
		return d.updateEdit(msg)
	}
	return nil
}

func (d *captureDialog) updateStart(msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		switch mt.String() {
		case "up":
			if d.startRow > 0 {
				d.startRow--
			}
		case "down":
			if d.startRow*2 < d.m.grid.Height-2 {
				d.startRow++
			}
		case "left":
			if d.startCol > 0 {
				d.startCol--
			}
		case "right":
			if d.startCol*2 < d.m.grid.Width-2 {
				d.startCol++
			}
		case "home":
			d.startCol = 0
		case "end":
			d.startCol = (d.m.grid.Width / 2) - 1
		case "pgup":
			d.startRow = 0
		case "pgdown":
			d.startRow = (d.m.grid.Height / 2) - 1
		case "space", "enter":
			d.endRow, d.endCol = d.startRow, d.startCol
			d.markArea()
			d.stage = captureEndMark
		}
	case tea.MouseClickMsg:
		mmsg := mt.Mouse()
		if mmsg.Y*2 < d.m.grid.Height && mmsg.X*2 < d.m.grid.Width {
			d.startRow, d.startCol = mmsg.Y, mmsg.X
		}
	}
	return nil
}

func (d *captureDialog) updateEnd(msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		switch mt.String() {
		case "up":
			if d.endRow > 0 {
				d.resetArea()
				d.endRow--
				d.markArea()
			}
		case "down":
			if d.endRow*2 < d.m.grid.Height-2 {
				d.resetArea()
				d.endRow++
				d.markArea()
			}
		case "left":
			if d.endCol > 0 {
				d.resetArea()
				d.endCol--
				d.markArea()
			}
		case "right":
			if d.endCol*2 < d.m.grid.Width-2 {
				d.resetArea()
				d.endCol++
				d.markArea()
			}
		case "home":
			d.resetArea()
			d.endCol = 0
			d.markArea()
		case "end":
			d.resetArea()
			d.endCol = (d.m.grid.Width / 2) - 1
			d.markArea()
		case "pgup":
			d.resetArea()
			d.endRow = 0
			d.markArea()
		case "pgdown":
			d.resetArea()
			d.endRow = (d.m.grid.Height / 2) - 1
			d.markArea()
		case "space", "enter":
			d.scrapePattern()
			captureDetailsForm.Reset(d)
			captureModifyForm.Reset(d)
			d.stage = captureEditPattern
		}
	case tea.MouseClickMsg:
		mmsg := mt.Mouse()
		if mmsg.Y*2 < d.m.grid.Height && mmsg.X*2 < d.m.grid.Width {
			d.resetArea()
			d.endRow, d.endCol = mmsg.Y, mmsg.X
			d.markArea()
		}
	}
	return nil
}

func (d *captureDialog) updateEdit(msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		switch mt.String() {
		case "ctrl+e":
			d.tab = captureTabDetails
		case "ctrl+o":
			d.tab = captureTabModify
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
	default:
		if d.currentForm != nil {
			return d.currentForm.Update(d, msg)
		}
	}
	return nil
}

func (d *captureDialog) croppedPattern() *patterns.Pattern {
	newWidth := d.pattern.Width - d.cropLeft - d.cropRight
	newHeight := d.pattern.Height - d.cropTop - d.cropBottom
	cells := make([]bool, 0, newWidth*newHeight)
	for y := d.cropTop; y < d.pattern.Height-d.cropBottom; y++ {
		from := y*d.pattern.Width + d.cropLeft
		to := from + newWidth
		cells = append(cells, d.pattern.Cells[from:to]...)
	}
	return &patterns.Pattern{
		Name:        d.pattern.Name,
		Width:       newWidth,
		Height:      newHeight,
		Cells:       cells,
		Comments:    d.pattern.Comments,
		Origination: d.pattern.Origination,
		Coordinates: d.pattern.Coordinates,
		Rule:        d.pattern.Rule,
	}
}

type patternSaveResult struct {
	error    error
	pattern  *patterns.Pattern
	filename string
}
