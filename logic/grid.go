package logic

import (
	"errors"
	"math/rand"
	"time"
)

type WrapMode int

const (
	WrapNone WrapMode = iota
	WrapHorizontal
	WrapVertical
	WrapAll
)

type BoundaryMode int

const (
	DeadBoundary BoundaryMode = iota
	AliveBoundary
)

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
	return gridChanged
}
