package patterns

import (
	"errors"
	"github.com/marrow16/gogol/logic"
	"io"
)

type Pattern struct {
	Name        string
	Width       int
	Height      int
	Cells       []bool // flat slice of cells - rows x cols
	Comments    []string
	Origination string
	Coordinates string
	Rule        logic.Rule
	Filename    string
}

func (p Pattern) String() string {
	if len(p.Name) > 0 {
		return p.Name
	}
	return p.Filename
}

func NewPattern(name string, width int, cells []bool) (Pattern, error) {
	if len(cells)%width != 0 {
		return Pattern{}, errors.New("pattern cells must be a multiple of width")
	}
	return Pattern{
		Name:   name,
		Width:  width,
		Height: len(cells) / width,
		Cells:  cells,
		Rule:   logic.StandardRule,
	}, nil
}

func MustNewPattern(name string, width int, cells []bool) Pattern {
	if p, err := NewPattern(name, width, cells); err != nil {
		panic(err)
	} else {
		return p
	}
}

func NewPatternFromRle(r io.Reader) (Pattern, error) {
	return PatternRleDecoder(r)
}

func MustNewPatternFromRle(r io.Reader) Pattern {
	if p, err := NewPatternFromRle(r); err != nil {
		panic(err)
	} else {
		return p
	}
}

func NewPatternFromGrid(grid *logic.Grid) (result Pattern, err error) {
	if grid == nil {
		return Pattern{}, errors.New("grid must not be nil")
	}
	result = Pattern{
		Name:   "Grid",
		Width:  grid.Width,
		Height: grid.Height,
		Rule:   grid.Rule,
		Cells:  make([]bool, grid.Width*grid.Height),
	}
	for r := 0; r < grid.Height; r++ {
		for c := 0; c < grid.Width; c++ {
			result.Cells[r*grid.Width+c] = grid.GetCell(r, c).Alive
		}
	}
	return result, nil
}

func NewPatternFromGridPortion(grid *logic.Grid, startRow, startCol, height, width int) Pattern {
	startRow = max(0, min(startRow, grid.Height-1))
	startCol = max(0, min(startCol, grid.Width-1))
	height = max(1, height)
	width = max(1, width)
	if startRow+height > grid.Height {
		height = grid.Height - startRow
	}
	if startCol+width > grid.Width {
		width = grid.Width - startCol
	}
	result := Pattern{
		Width:  width,
		Height: height,
		Rule:   grid.Rule,
		Cells:  make([]bool, width*height),
	}
	for r := 0; r < height; r++ {
		for c := 0; c < width; c++ {
			result.Cells[r*width+c] =
				grid.GetCell(startRow+r, startCol+c).Alive
		}
	}
	return result
}

type Rotation int

const (
	Rotate0 Rotation = iota
	Rotate90
	Rotate180
	Rotate270
)

func (r Rotation) String() string {
	switch r {
	case Rotate90:
		return "90°"
	case Rotate180:
		return "180°"
	case Rotate270:
		return "270°"
	default:
		return "0°"
	}
}

func (p Pattern) Draw(grid *logic.Grid, row, col int, rot Rotation, flags ...bool) {
	interlaced := len(flags) > 0 && flags[0]
	p.DrawTo(rot, func(y, x int, alive bool) {
		if !interlaced || (interlaced && alive) {
			grid.SetCell(row+y, col+x, alive)
		}
	})
}

func (p Pattern) DrawTo(rot Rotation, fn func(row, col int, alive bool)) {
	if fn == nil {
		return
	}
	for y := 0; y < p.Height; y++ {
		for x := 0; x < p.Width; x++ {
			alive := p.Cells[y*p.Width+x]
			var row, col int
			switch rot {
			case Rotate90:
				row, col = x, p.Height-1-y
			case Rotate180:
				row, col = p.Height-1-y, p.Width-1-x
			case Rotate270:
				row, col = p.Width-1-x, y
			default:
				row, col = y, x
			}
			fn(row, col, alive)
		}
	}
}
