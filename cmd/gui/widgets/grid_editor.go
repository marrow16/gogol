package widgets

import (
	"bytes"
	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"github.com/marrow16/gogol/patterns"
	"io"
	"slices"
	"strings"
	"sync"
	"time"
)

type editor struct {
	g                *gridHolder
	active           bool
	blink            bool
	dirty            bool // blink state changed since last render
	stop             chan struct{}
	row, col         int
	undos            []undo
	wasUnder         [][2]int
	patternRotation  patterns.Rotation
	patternInterlace bool
	mutex            sync.Mutex

	marking          bool
	markingUnderlays []underlay
	markingDirty     bool
	markStartRow     int
	markStartCol     int

	clipboardTag struct{}
}

type undoKind int

const (
	undoCell undoKind = iota
	undoPattern
)

type undo struct {
	kind                   undoKind
	restoreRow, restoreCol int
	row, col               int
	wasAlive               bool
	wasGrid                patterns.Pattern
	keyName                key.Name
	keyMods                key.Modifiers
	keyTimestamp           time.Time
}

func (e *editor) handleKeys(gtx layout.Context, kev key.Event) (handled bool) {
	if !e.active {
		return false
	}
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if kev.Name == key.NameCtrl || kev.Name == key.NameCommand {
		return false
	}
	if kev.Modifiers == key.ModShortcut {
		if e.handleShortcutKeys(gtx, kev) {
			return true
		}
	}
	handled = true
	switch kev.Name {
	case key.NameCtrl, key.NameAlt, key.NameShift, key.NameCommand:
		//do nothing
	case key.NameEnter, key.NameReturn:
		e.endMarking(true)
	case key.NameSpace:
		e.endMarking(false)
		e.setCell(kev.Modifiers&key.ModShift != 0, kev.Name, kev.Modifiers)
		e.adjustColumn(1)
	case key.NameDeleteBackward:
		e.endMarking(false)
		e.setCell(kev.Modifiers == key.ModShift, kev.Name, kev.Modifiers)
		e.adjustColumn(-1)
	case key.NameLeftArrow:
		if kev.Modifiers&key.ModAlt == key.ModAlt {
			e.endMarking(false)
			e.setCell(kev.Modifiers&key.ModShift == 0, kev.Name, kev.Modifiers)
			e.adjustColumn(-1)
		} else if kev.Modifiers == key.ModShift {
			e.markingAdjust(0, -1)
		} else {
			e.endMarking(false)
			e.adjustColumn(-1)
		}
	case key.NameRightArrow:
		if kev.Modifiers&key.ModAlt == key.ModAlt {
			e.endMarking(false)
			e.setCell(kev.Modifiers&key.ModShift == 0, kev.Name, kev.Modifiers)
			e.adjustColumn(1)
		} else if kev.Modifiers == key.ModShift {
			e.markingAdjust(0, 1)
		} else {
			e.endMarking(false)
			e.adjustColumn(1)
		}
	case key.NameUpArrow:
		if kev.Modifiers&key.ModAlt == key.ModAlt {
			e.endMarking(false)
			e.setCell(kev.Modifiers&key.ModShift == 0, kev.Name, kev.Modifiers)
			e.adjustRow(-1)
		} else if kev.Modifiers == key.ModShift {
			e.markingAdjust(-1, 0)
		} else {
			e.endMarking(false)
			e.adjustRow(-1)
		}
	case key.NameDownArrow:
		if kev.Modifiers&key.ModAlt == key.ModAlt {
			e.endMarking(false)
			e.setCell(kev.Modifiers&key.ModShift == 0, kev.Name, kev.Modifiers)
			e.adjustRow(1)
		} else if kev.Modifiers == key.ModShift {
			e.markingAdjust(1, 0)
		} else {
			e.endMarking(false)
			e.adjustRow(1)
		}
	case key.NameHome:
		e.endMarking(false)
		e.adjustColumn(-e.col)
	case key.NameEnd:
		e.endMarking(false)
		e.adjustColumn(e.g.grid.Width - e.col - 1)
	case key.NamePageUp:
		e.endMarking(false)
		e.adjustRow(-e.row)
	case key.NamePageDown:
		e.endMarking(false)
		e.adjustRow(e.g.grid.Height - e.row - 1)
	default:
		if kev.Modifiers&key.ModAlt == key.ModAlt {
			e.handleSpecialKeys(gtx, kev)
		} else {
			e.endMarking(false)
			e.drawLetter(gtx, kev)
		}
	}
	return handled
}

const (
	clipboardWriteType = "text/plain"
	clipboardReadType  = "application/text" // Gio paste DataEvent type
)

func (e *editor) handlePaste(evt transfer.DataEvent) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if e.active {
		r := evt.Open()
		defer func() {
			_ = r.Close()
		}()
		if pattern, err := patterns.NewPatternFromRle(r); err == nil {
			e.endMarking(false)
			e.addPatternUndo(pattern, e.row, e.col, e.patternRotation)
			pattern.Draw(e.g.grid, e.row, e.col, e.patternRotation, e.patternInterlace)
		}
	}
}

func (e *editor) handleShortcutKeys(gtx layout.Context, kev key.Event) (handled bool) {
	switch kev.Name {
	case "V":
		//paste
		handled = true
		gtx.Execute(clipboard.ReadCmd{Tag: &e.clipboardTag})
	case "C":
		//copy
		handled = true
		if pattern := e.markedPattern(); pattern != nil {
			pattern.Name = time.Now().Format("2006-01-02 15:04:05")
			pattern.Origination = e.g.core.settings.Originator
			pattern.Comments = []string{"Copied from GoGoL"}
			var buf bytes.Buffer
			if err := patterns.PatternRleEncode(*pattern, &buf); err == nil {
				gtx.Execute(clipboard.WriteCmd{
					Type: clipboardWriteType,
					Data: io.NopCloser(&buf),
				})
			}
		}
	case "X":
		//cut
		handled = true
		if pattern := e.markedPattern(); pattern != nil {
			e.endMarking(false)
			pattern.Name = time.Now().Format("2006-01-02 15:04:05")
			pattern.Origination = e.g.core.settings.Originator
			pattern.Comments = []string{"Copied from GoGoL"}
			var buf bytes.Buffer
			if err := patterns.PatternRleEncode(*pattern, &buf); err == nil {
				gtx.Execute(clipboard.WriteCmd{
					Type: clipboardWriteType,
					Data: io.NopCloser(&buf),
				})
				sr := min(e.row, e.markStartRow)
				sc := min(e.col, e.markStartCol)
				e.addPatternUndo(*pattern, sr, sc, patterns.Rotate0)
				cp := patterns.Pattern{
					Width:  pattern.Width,
					Height: pattern.Height,
					Cells:  slices.Clone(pattern.Cells),
				}
				for i := 0; i < len(cp.Cells); i++ {
					cp.Cells[i] = false
				}
				e.clearMarking()
				cp.Draw(e.g.grid, sr, sc, patterns.Rotate0)
			}
		}
	case "Z":
		//undo
		handled = true
		e.endMarking(false)
		e.doUndo()
	case "A":
		//all
		handled = true
		e.markStartRow = e.g.grid.Height - 1
		e.markStartCol = e.g.grid.Width - 1
		e.markArea(0, 0)
		e.blink = true
		e.wasUnder = append(e.wasUnder, [2]int{e.row, e.col})
		e.row, e.col = e.markStartRow, e.markStartCol
		e.dirty = true
	}
	return handled
}

func (e *editor) drawLetter(gtx layout.Context, kev key.Event) {
	ch := string(kev.Name)
	if (ch >= "A" && ch <= "Z") && kev.Modifiers != key.ModShift {
		ch = strings.ToLower(ch)
	}
	if chpatt, ok := alphabet[ch]; ok {
		r, c := e.row-chpatt.Height+3, e.col
		ajr, ajc := 0, chpatt.Width+1
		switch e.patternRotation {
		case patterns.Rotate90:
			r, c = e.row, e.col-2
			ajr, ajc = chpatt.Width+1, 0
		case patterns.Rotate180:
			r, c = e.row-2, e.col-chpatt.Width+1
			ajr, ajc = 0, -(chpatt.Width + 1)
		case patterns.Rotate270:
			r, c = e.row-chpatt.Width+1, e.col-chpatt.Height+3
			ajr, ajc = -(chpatt.Width + 1), 0
		}
		e.addPatternUndo(chpatt, r, c, e.patternRotation)
		chpatt.Draw(e.g.grid, r, c, e.patternRotation, e.patternInterlace)
		e.adjustRow(ajr)
		e.adjustColumn(ajc)
	}
}

func (e *editor) setCell(alive bool, keyName key.Name, keyMods key.Modifiers) {
	if cell := e.g.grid.GetCell(e.row, e.col); cell != nil {
		e.undos = append(e.undos, undo{
			kind:         undoCell,
			restoreRow:   e.row,
			restoreCol:   e.col,
			row:          e.row,
			col:          e.col,
			wasAlive:     cell.Alive,
			keyName:      keyName,
			keyMods:      keyMods,
			keyTimestamp: time.Now(),
		})
	}
	e.g.grid.SetCell(e.row, e.col, alive)
}

func (e *editor) addPatternUndo(pattern patterns.Pattern, row, col int, rotation patterns.Rotation) {
	u := undo{
		kind:       undoPattern,
		restoreRow: e.row,
		restoreCol: e.col,
	}
	height, width := pattern.Height, pattern.Width
	if rotation == patterns.Rotate90 || rotation == patterns.Rotate270 {
		height, width = width, height
	}
	startRow, startCol := row, col
	endRow, endCol := row+height, col+width
	if startRow < 0 {
		startRow = 0
	}
	if startCol < 0 {
		startCol = 0
	}
	if endRow > e.g.grid.Height {
		endRow = e.g.grid.Height
	}
	if endCol > e.g.grid.Width {
		endCol = e.g.grid.Width
	}
	// make sure pattern isn't entirely outside the grid...
	if startRow >= endRow || startCol >= endCol {
		return
	}
	u.row, u.col = startRow, startCol
	u.wasGrid = patterns.NewPatternFromGridPortion(e.g.grid, startRow, startCol, endRow-startRow, endCol-startCol)
	e.undos = append(e.undos, u)
}

func (e *editor) handleSpecialKeys(gtx layout.Context, kev key.Event) {
	switch kev.Name {
	case "C":
		e.endMarking(false)
		if pattern, err := patterns.NewPatternFromGrid(e.g.grid); err == nil {
			e.undos = append(e.undos, undo{
				kind:       undoPattern,
				restoreRow: e.row,
				restoreCol: e.col,
				row:        0,
				col:        0,
				wasGrid:    pattern,
			})
		}
		e.g.grid.Clear()
	case "I":
		e.patternInterlace = !e.patternInterlace
	case "R":
		e.patternRotation++
		if e.patternRotation > patterns.Rotate270 {
			e.patternRotation = patterns.Rotate0
		}
	case "P":
		e.endMarking(false)
		if pattern, _ := e.g.core.statusBar.menuPopup.patternsPopout.currentPattern(); pattern != nil {
			e.addPatternUndo(*pattern, e.row, e.col, e.patternRotation)
			pattern.Draw(e.g.grid, e.row, e.col, e.patternRotation)
		}
	case "F":
		//fill
		if pattern := e.markedPattern(); pattern != nil {
			e.endMarking(false)
			sr := min(e.row, e.markStartRow)
			sc := min(e.col, e.markStartCol)
			e.addPatternUndo(*pattern, sr, sc, patterns.Rotate0)
			np := patterns.Pattern{
				Width:  pattern.Width,
				Height: pattern.Height,
				Cells:  slices.Clone(pattern.Cells),
			}
			fill := kev.Modifiers&key.ModShift == 0
			for i := 0; i < len(np.Cells); i++ {
				np.Cells[i] = fill
			}
			e.clearMarking()
			np.Draw(e.g.grid, sr, sc, patterns.Rotate0)
		}
	case "Z":
		e.endMarking(false)
		e.doUndo()
	case "U":
		// rotate up
		e.endMarking(false)
		if pattern, err := patterns.NewPatternFromGrid(e.g.grid); err == nil {
			e.addPatternUndo(pattern, 0, 0, patterns.Rotate0)
			pattern = pattern.RotatedUp()
			pattern.Draw(e.g.grid, 0, 0, patterns.Rotate0)
		}
	case "D":
		// rotate down
		e.endMarking(false)
		if pattern, err := patterns.NewPatternFromGrid(e.g.grid); err == nil {
			e.addPatternUndo(pattern, 0, 0, patterns.Rotate0)
			pattern = pattern.RotatedDown()
			pattern.Draw(e.g.grid, 0, 0, patterns.Rotate0)
		}
	case ",":
		// rotate left
		e.endMarking(false)
		if pattern, err := patterns.NewPatternFromGrid(e.g.grid); err == nil {
			e.addPatternUndo(pattern, 0, 0, patterns.Rotate0)
			pattern = pattern.RotatedLeft()
			pattern.Draw(e.g.grid, 0, 0, patterns.Rotate0)
		}
	case ".":
		// rotate right
		e.endMarking(false)
		if pattern, err := patterns.NewPatternFromGrid(e.g.grid); err == nil {
			e.addPatternUndo(pattern, 0, 0, patterns.Rotate0)
			pattern = pattern.RotatedRight()
			pattern.Draw(e.g.grid, 0, 0, patterns.Rotate0)
		}
	}
}

func (e *editor) doUndo() {
	if len(e.undos) > 0 {
		u := e.undos[len(e.undos)-1]
		e.undos = e.undos[:len(e.undos)-1]
		if u.kind == undoCell {
			e.g.grid.SetCell(u.row, u.col, u.wasAlive)
			u = e.groupedKeyUndos(u)
		} else {
			u.wasGrid.Draw(e.g.grid, u.row, u.col, patterns.Rotate0)
		}
		e.wasUnder = append(e.wasUnder, [2]int{e.row, e.col})
		e.row, e.col = u.restoreRow, u.restoreCol
		e.blink = true
		e.dirty = true
	}
}

func (e *editor) groupedKeyUndos(u undo) undo {
	threshold := 500 * time.Millisecond
	for len(e.undos) > 0 {
		gu := e.undos[len(e.undos)-1]
		if gu.kind == undoCell && gu.keyName == u.keyName && gu.keyMods == u.keyMods && u.keyTimestamp.Sub(gu.keyTimestamp) < threshold {
			e.g.grid.SetCell(gu.row, gu.col, gu.wasAlive)
			e.undos = e.undos[:len(e.undos)-1]
			u = gu
		} else {
			break
		}
	}
	return u
}

func (e *editor) setPosition(row, col int) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if row >= 0 && row < e.g.grid.Height && col >= 0 && col < e.g.grid.Width && row != e.row && col != e.col {
		e.blink = true
		e.wasUnder = append(e.wasUnder, [2]int{e.row, e.col})
		e.row, e.col = row, col
		e.dirty = true
		if e.marking {
			e.markingDirty = true
			e.marking = false
		}
	}
}

func (e *editor) adjustRow(adj int) {
	e.blink = true
	e.wasUnder = append(e.wasUnder, [2]int{e.row, e.col})
	e.dirty = true
	e.row += adj
	switch {
	case e.row < 0:
		e.row = e.g.grid.Height - 1
	case e.row >= e.g.grid.Height:
		e.row = 0
	}
}

func (e *editor) markedPattern() (pattern *patterns.Pattern) {
	if e.marking && len(e.markingUnderlays) > 0 {
		pattern = &e.markingUnderlays[len(e.markingUnderlays)-1].pattern
	}
	return pattern
}

func (e *editor) endMarking(capture bool) {
	if e.marking {
		e.markingDirty = true
		if pattern := e.markedPattern(); pattern != nil && capture {
			e.g.core.statusBar.menuPopup.capturedPatternsPopout.addCapturedPattern(*pattern)
		}
	}
	e.marking = false
}

func (e *editor) markArea(toRow, toCol int) {
	if toRow < 0 || toRow >= e.g.grid.Height || toCol < 0 || toCol >= e.g.grid.Width {
		return
	}
	if !e.marking {
		e.marking = true
		e.markStartRow, e.markStartCol = e.row, e.col
	}
	sr := min(toRow, e.markStartRow)
	sc := min(toCol, e.markStartCol)
	er := max(toRow, e.markStartRow)
	ec := max(toCol, e.markStartCol)
	ht := er - sr + 1
	wd := ec - sc + 1
	e.markingUnderlays = append(e.markingUnderlays, underlay{
		row:     sr,
		col:     sc,
		pattern: patterns.NewPatternFromGridPortion(e.g.grid, sr, sc, ht, wd),
	})
	e.markingDirty = true
}

func (e *editor) markingAdjust(rowAdj, colAdj int) {
	if !e.marking {
		e.marking = true
		e.markStartRow, e.markStartCol = e.row, e.col
	}
	e.blink = true
	nr := e.row + rowAdj
	nc := e.col + colAdj
	if nr < 0 || nr >= e.g.grid.Height || nc < 0 || nc >= e.g.grid.Width {
		return
	}
	sr := min(nr, e.markStartRow)
	sc := min(nc, e.markStartCol)
	er := max(nr, e.markStartRow)
	ec := max(nc, e.markStartCol)
	ht := er - sr + 1
	wd := ec - sc + 1
	e.markingUnderlays = append(e.markingUnderlays, underlay{
		row:     sr,
		col:     sc,
		pattern: patterns.NewPatternFromGridPortion(e.g.grid, sr, sc, ht, wd),
	})
	e.markingDirty = true
	e.row = nr
	e.col = nc
}

func (e *editor) adjustColumn(adj int) {
	e.blink = true
	e.wasUnder = append(e.wasUnder, [2]int{e.row, e.col})
	e.dirty = true
	e.col += adj
	switch {
	case e.col < 0:
		e.col = e.g.grid.Width - 1
		e.row--
		if e.row < 0 {
			e.row = e.g.grid.Height - 1
		}
	case e.col >= e.g.grid.Width:
		e.col = 0
		e.row++
		if e.row >= e.g.grid.Height {
			e.row = 0
		}
	}
}

func (e *editor) imageOps() {
	if e.markingDirty {
		if !e.marking {
			e.clearMarking()
		} else {
			// restore underlays (apart from the last one)
			for i := 0; i < len(e.markingUnderlays)-1; i++ {
				e.restoreUnderlay(e.markingUnderlays[i])
			}
			if l := len(e.markingUnderlays); l > 0 {
				// now draw the last overlay as highlighted...
				e.markingUnderlays = e.markingUnderlays[l-1:]
				ul := e.markingUnderlays[0]
				aliveColor := placementAliveColor(e.g.core.settings.CellAliveColor)
				deadColor := placementDeadColor(e.g.core.settings.CellDeadColor)
				ul.pattern.DrawTo(patterns.Rotate0, func(row, col int, alive bool) {
					e.g.renderCellWithColors(row+ul.row, col+ul.col, alive, deadColor, aliveColor)
				})
			}
		}
		e.markingDirty = false
	}
	if e.dirty {
		e.drain()
		if e.active {
			aliveColor := placementAliveColor(e.g.core.settings.CellAliveColor)
			deadColor := e.g.core.settings.CellDeadColor
			if cell := e.g.grid.GetCell(e.row, e.col); cell != nil && cell.Alive {
				deadColor = e.g.core.settings.CellAliveColor
			}
			e.g.renderCellWithColors(e.row, e.col, e.blink, aliveColor, deadColor)
		}
	}
}

func (e *editor) restoreUnderlay(u underlay) {
	u.pattern.DrawTo(patterns.Rotate0, func(row, col int, alive bool) {
		e.g.renderCell(row+u.row, col+u.col, alive, true)
	})
}

func (e *editor) drain() {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	for _, u := range e.wasUnder {
		if cell := e.g.grid.GetCell(u[0], u[1]); cell != nil {
			e.g.renderCell(u[0], u[1], cell.Alive, true)
		}
	}
	e.wasUnder = e.wasUnder[:0]
	e.dirty = false
}

func (e *editor) start() {
	if e.stop != nil {
		return
	}
	if e.row < 0 {
		e.row = 0
	}
	if e.row >= e.g.grid.Height-1 {
		e.row = e.g.grid.Height - 1
	}
	if e.col < 0 {
		e.col = 0
	}
	if e.col >= e.g.grid.Width-1 {
		e.col = e.g.grid.Width - 1
	}
	e.active = true
	e.blink = true
	e.dirty = true
	e.patternRotation = patterns.Rotate0
	e.patternInterlace = false
	e.wasUnder = make([][2]int, 0)
	e.undos = make([]undo, 0)
	e.marking = false
	e.markingUnderlays = make([]underlay, 0)
	e.stop = make(chan struct{})
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				e.blink = !e.blink
				e.dirty = true
				e.g.core.window.Invalidate()
			case <-e.stop:
				return
			}
		}
	}()
}

func (e *editor) end() {
	if e.stop == nil {
		return
	}
	e.mutex.Lock()
	defer e.mutex.Unlock()
	close(e.stop)
	e.stop = nil
	e.blink = false
	e.dirty = true
	e.wasUnder = append(e.wasUnder, [2]int{e.row, e.col})
	e.active = false
	if e.marking {
		e.marking = false
		e.clearMarking()
	}
	e.g.grid.StepCount.Store(0)
	e.g.core.resetInstrumentation()
	e.g.core.window.Invalidate()
}

func (e *editor) clearMarking() {
	for _, ul := range e.markingUnderlays {
		e.restoreUnderlay(ul)
	}
	e.markingUnderlays = e.markingUnderlays[:0]
}
