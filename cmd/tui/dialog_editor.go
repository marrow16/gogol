package main

import (
	tea "charm.land/bubbletea/v2"
	"github.com/marrow16/gogol/cmd/tui/layout"
	"github.com/marrow16/gogol/patterns"
	"strconv"
)

type editor struct {
	m                *model
	row, col         int
	offsetY, offsetX int
	undo             *patterns.Pattern
}

func (d *editor) title() string {
	return "[editing " + strconv.Itoa(d.row*2) + "/" + strconv.Itoa(d.col*2) + "]"
}

func (d *editor) render(sf layout.Surface) *tea.Cursor {
	if d.row < 0 {
		d.row = 0
	} else if d.row >= d.m.gridSurface.Height() {
		d.row = d.m.gridSurface.Height() - 1
	}
	if d.col < 0 {
		d.col = 0
	} else if d.col >= d.m.gridSurface.Width() {
		d.col = d.m.gridSurface.Width() - 1
	}
	csr := tea.NewCursor(d.col, d.row)
	csr.Color = cursorColor
	return csr
}

func (d *editor) update(msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		switch mt.String() {
		case "esc":
			d.m.stopped()
		case "up":
			if d.row > 0 {
				d.offsetY, d.offsetX = 0, 0
				d.row--
			}
		case "down":
			if d.row < d.m.gridSurface.Height()-1 {
				d.offsetY, d.offsetX = 0, 0
				d.row++
			}
		case "left":
			if d.col > 0 {
				d.offsetY, d.offsetX = 0, 0
				d.col--
			}
		case "right":
			if d.col < d.m.gridSurface.Width()-1 {
				d.offsetY, d.offsetX = 0, 0
				d.col++
			}
		case "home":
			d.offsetY, d.offsetX = 0, 0
			d.col = 0
		case "end":
			d.offsetY, d.offsetX = 0, 0
			d.col = d.m.gridSurface.Width() - 1
		case "pgup":
			d.offsetY, d.offsetX = 0, 0
			d.row = 0
		case "pgdown":
			d.offsetY, d.offsetX = 0, 0
			d.row = d.m.gridSurface.Height() - 1
		case "backspace":
			d.offsetY, d.offsetX = 0, 0
			cy, cx := d.row*2, d.col*2
			d.m.grid.SetCell(cy, cx, false)
			d.m.grid.SetCell(cy, cx+1, false)
			d.m.grid.SetCell(cy+1, cx, false)
			d.m.grid.SetCell(cy+1, cx+1, false)
			if d.col > 0 {
				d.col--
			} else if d.row > 0 {
				d.row--
				d.col = d.m.gridSurface.Width() - 1
			}
		case "space":
			d.offsetY, d.offsetX = 0, 0
			cy, cx := d.row*2, d.col*2
			d.m.grid.SetCell(cy, cx, false)
			d.m.grid.SetCell(cy, cx+1, false)
			d.m.grid.SetCell(cy+1, cx, false)
			d.m.grid.SetCell(cy+1, cx+1, false)
			if d.col < d.m.gridSurface.Width()-1 {
				d.col++
			} else if d.row < d.m.gridSurface.Height()-1 {
				d.row++
				d.col = 0
			}
		case "ctrl+space":
			d.offsetY, d.offsetX = 0, 0
			cy, cx := d.row*2, d.col*2
			d.m.grid.SetCell(cy, cx, true)
			d.m.grid.SetCell(cy, cx+1, true)
			d.m.grid.SetCell(cy+1, cx, true)
			d.m.grid.SetCell(cy+1, cx+1, true)
			if d.col < d.m.gridSurface.Width()-1 {
				d.col++
			} else if d.row < d.m.gridSurface.Height()-1 {
				d.row++
				d.col = 0
			}
		case "1":
			d.offsetY, d.offsetX = 0, 0
			cy, cx := d.row*2, d.col*2
			if cell := d.m.grid.GetCell(cy, cx); cell != nil {
				d.m.grid.SetCell(cy, cx, !cell.Alive)
			}
		case "2":
			d.offsetY, d.offsetX = 0, 1
			cy, cx := d.row*2, (d.col*2)+1
			if cell := d.m.grid.GetCell(cy, cx); cell != nil {
				d.m.grid.SetCell(cy, cx, !cell.Alive)
			}
		case "3":
			d.offsetY, d.offsetX = 1, 0
			cy, cx := (d.row*2)+1, d.col*2
			if cell := d.m.grid.GetCell(cy, cx); cell != nil {
				d.m.grid.SetCell(cy, cx, !cell.Alive)
			}
		case "4":
			d.offsetY, d.offsetX = 1, 1
			cy, cx := (d.row*2)+1, (d.col*2)+1
			if cell := d.m.grid.GetCell(cy, cx); cell != nil {
				d.m.grid.SetCell(cy, cx, !cell.Alive)
			}
		case "5":
			d.offsetY, d.offsetX = 0, 0
		case "6":
			d.offsetY, d.offsetX = 0, 1
		case "7":
			d.offsetY, d.offsetX = 1, 0
		case "8":
			d.offsetY, d.offsetX = 1, 1
		case "ctrl+a":
			d.m.grid.Clear()
		case "ctrl+p":
			if d.m.patterns.currentPattern != nil {
				if undo, err := patterns.NewPatternFromGrid(d.m.grid); err == nil {
					d.undo = &undo
				}
				r, c := (d.row*2)+d.offsetY, (d.col*2)+d.offsetX
				d.m.patterns.currentPattern.Draw(d.m.grid, r, c, d.m.patterns.patternRotate)
			}
		case "ctrl+b":
			if d.undo != nil {
				d.undo.Draw(d.m.grid, 0, 0, patterns.Rotate0)
			}
		default:
			if chPattern, ok := alphabet[mt.String()]; ok {
				r, c := ((d.row*2)+d.offsetY)-5, (d.col*2)+d.offsetX
				chPattern.Draw(d.m.grid, r, c, patterns.Rotate0)
				d.col += (chPattern.Width / 2) + 1
				if d.col >= d.m.gridSurface.Width() {
					d.col, d.offsetY, d.offsetX = 0, 0, 0
					d.row += 5
					if d.row >= d.m.gridSurface.Height() {
						d.row = 0
					}
				}
			}
		}
	case tea.MouseClickMsg:
		mmsg := mt.Mouse()
		if mmsg.Y >= 0 && mmsg.X >= 0 && mmsg.Y < d.m.gridSurface.Height() && mmsg.X < d.m.gridSurface.Width() {
			d.row, d.col = mmsg.Y, mmsg.X
		}
	}
	return nil
}
