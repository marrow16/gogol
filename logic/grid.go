package logic

import (
	"errors"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type WrapMode int

const (
	WrapNone WrapMode = iota
	WrapHorizontal
	WrapVertical
	WrapAll
)

func (m WrapMode) String() string {
	switch m {
	case WrapNone:
		return "None"
	case WrapHorizontal:
		return "Horizontal"
	case WrapVertical:
		return "Vertical"
	case WrapAll:
		return "Toroidal"
	}
	return "Unknown"
}

func WrapModeFromString(s string, def WrapMode) WrapMode {
	switch strings.ToLower(s) {
	case "none":
		return WrapNone
	case "horizontal":
		return WrapHorizontal
	case "vertical":
		return WrapVertical
	case "all", "toroidal":
		return WrapAll
	}
	return def
}

type BoundaryMode int

const (
	DeadBoundary BoundaryMode = iota
	AliveBoundary
)

func (m BoundaryMode) String() string {
	switch m {
	case DeadBoundary:
		return "Dead"
	case AliveBoundary:
		return "Alive"
	}
	return "Unknown"
}

func BoundaryModeFromString(s string, def BoundaryMode) BoundaryMode {
	switch strings.ToLower(s) {
	case "dead":
		return DeadBoundary
	case "alive":
		return AliveBoundary
	}
	return def
}

type RenderCell func(row, col int, alive, changed bool)

type GridRow []*Cell

type Grid struct {
	Rule             Rule
	Rows             []GridRow
	Height           int
	Width            int
	WrapMode         WrapMode
	BoundaryMode     BoundaryMode
	boundarySentinel *Cell
	changesBuffer    []*Cell
	Render           RenderCell
	StepCount        atomic.Uint64
	mutex            sync.Mutex
}

var ErrInvalidGridDimension = errors.New("invalid grid dimension")

func NewGrid(height int, width int, wrapMode WrapMode, boundaryMode BoundaryMode) (*Grid, error) {
	if height < 2 || width < 2 {
		return nil, ErrInvalidGridDimension
	}
	result := &Grid{
		Rule:             StandardRule,
		Rows:             makeRows(height, width),
		Height:           height,
		Width:            width,
		WrapMode:         wrapMode,
		BoundaryMode:     boundaryMode,
		boundarySentinel: &Cell{Alive: boundaryMode == AliveBoundary},
		changesBuffer:    make([]*Cell, 0, height*width),
	}
	result.joinAdjacents()
	return result, nil
}

func makeRows(height int, width int) []GridRow {
	result := make([]GridRow, height)
	for r := range result {
		result[r] = makeCols(width)
	}
	return result
}

func makeCols(width int) []*Cell {
	result := make([]*Cell, width)
	for c := range result {
		result[c] = &Cell{}
	}
	return result
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func (g *Grid) Randomize(rf int) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.StepCount.Store(0)
	for row := 0; row < g.Height; row++ {
		for col := 0; col < g.Width; col++ {
			cell := g.Rows[row][col]
			cell.Alive = rng.Intn(100) < rf
			if g.Render != nil {
				g.Render(row, col, cell.Alive, true)
			}
		}
	}
}

func (g *Grid) Clear() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.StepCount.Store(0)
	for row := 0; row < g.Height; row++ {
		for col := 0; col < g.Width; col++ {
			cell := g.Rows[row][col]
			cell.Alive = false
			if g.Render != nil {
				g.Render(row, col, cell.Alive, true)
			}
		}
	}
}

func (g *Grid) Draw() {
	for row := 0; row < g.Height; row++ {
		for col := 0; col < g.Width; col++ {
			cell := g.Rows[row][col]
			if g.Render != nil {
				g.Render(row, col, cell.Alive, true)
			}
		}
	}
}

func (g *Grid) joinAdjacents() {
	for r := 0; r < g.Height; r++ {
		for c := 0; c < g.Width; c++ {
			cell := g.Rows[r][c]
			cell.Adjacents = Adjacents{
				g.cellAt(r-1, c-1),
				g.cellAt(r-1, c),
				g.cellAt(r-1, c+1),
				g.cellAt(r, c-1),
				g.cellAt(r, c+1),
				g.cellAt(r+1, c-1),
				g.cellAt(r+1, c),
				g.cellAt(r+1, c+1),
			}
		}
	}
}

func (g *Grid) cellAt(r, c int) *Cell {
	if g.WrapMode == WrapVertical || g.WrapMode == WrapAll {
		if r < 0 {
			r = g.Height - 1
		} else if r >= g.Height {
			r = 0
		}
	}
	if g.WrapMode == WrapHorizontal || g.WrapMode == WrapAll {
		if c < 0 {
			c = g.Width - 1
		} else if c >= g.Width {
			c = 0
		}
	}
	if r >= 0 && r < g.Height && c >= 0 && c < g.Width {
		return g.Rows[r][c]
	}
	return g.boundarySentinel
}

func (g *Grid) SetBoundaryMode(m BoundaryMode) {
	g.BoundaryMode = m
	g.boundarySentinel.Alive = m == AliveBoundary
}

func (g *Grid) SetWrapMode(m WrapMode) {
	if g.WrapMode != m {
		g.WrapMode = m
		g.joinAdjacents()
	}
}

func (g *Grid) SetCell(row, col int, alive bool) (changed bool) {
	if row >= 0 && row < g.Height && col >= 0 && col < g.Width {
		g.mutex.Lock()
		defer g.mutex.Unlock()
		cell := g.Rows[row][col]
		was := cell.Alive
		cell.Alive = alive
		changed = was != alive
		if g.Render != nil {
			g.Render(row, col, alive, changed)
		}
	}
	return changed
}

func (g *Grid) GetCell(row, col int) *Cell {
	if row >= 0 && row < g.Height && col >= 0 && col < g.Width {
		return g.Rows[row][col]
	}
	return nil
}

func nullRender(row, col int, alive, changed bool) {}

func (g *Grid) Step() (gridChanged bool) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if g.Rule == nil {
		g.Rule = StandardRule
	}
	render := g.Render
	if render == nil {
		render = nullRender
	}
	g.changesBuffer = g.changesBuffer[:0]
	for r, row := range g.Rows {
		for c, cell := range row {
			if g.Rule.StateChanged(cell) {
				g.changesBuffer = append(g.changesBuffer, cell)
				render(r, c, !cell.Alive, true)
			} else {
				render(r, c, cell.Alive, false)
			}
		}
	}
	gridChanged = len(g.changesBuffer) > 0
	for _, cell := range g.changesBuffer {
		cell.flip()
	}
	if gridChanged {
		g.StepCount.Add(1)
	}
	return gridChanged
}

func (g *Grid) StepAhead(by int) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if g.Rule == nil {
		g.Rule = StandardRule
	}
	count := uint64(0)
	for i := 0; i < by; i++ {
		g.changesBuffer = g.changesBuffer[:0]
		for _, row := range g.Rows {
			for _, cell := range row {
				if g.Rule.StateChanged(cell) {
					g.changesBuffer = append(g.changesBuffer, cell)
				}
			}
		}
		if len(g.changesBuffer) == 0 {
			break
		}
		count++
		for _, cell := range g.changesBuffer {
			cell.flip()
		}
	}
	g.StepCount.Add(count)
}
