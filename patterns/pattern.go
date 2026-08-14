package patterns

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"github.com/marrow16/gogol/logic"
	"io"
	"slices"
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

func (p Pattern) Trimmed() Pattern {
	if p.Width <= 0 || p.Height <= 0 || len(p.Cells) != p.Width*p.Height {
		return Pattern{}
	}
	minRow, minCol := p.Height, p.Width
	maxRow, maxCol := -1, -1
	for row := 0; row < p.Height; row++ {
		for col := 0; col < p.Width; col++ {
			if !p.Cells[row*p.Width+col] {
				continue
			}
			if row < minRow {
				minRow = row
			}
			if row > maxRow {
				maxRow = row
			}
			if col < minCol {
				minCol = col
			}
			if col > maxCol {
				maxCol = col
			}
		}
	}
	if maxRow == -1 {
		// no live cells...
		return Pattern{}
	}
	width := maxCol - minCol + 1
	height := maxRow - minRow + 1
	if width == p.Width && height == p.Height {
		return p
	}
	cells := make([]bool, width*height)
	for row := 0; row < height; row++ {
		src := (minRow+row)*p.Width + minCol
		dst := row * width
		copy(cells[dst:dst+width], p.Cells[src:src+width])
	}
	return Pattern{
		Width:  width,
		Height: height,
		Cells:  cells,
	}
}

func (p Pattern) Rotated(rotation Rotation) Pattern {
	switch rotation {
	case Rotate0:
		return p
	case Rotate90:
		cells := make([]bool, len(p.Cells))
		width, height := p.Height, p.Width
		for row := 0; row < p.Height; row++ {
			for col := 0; col < p.Width; col++ {
				newRow := col
				newCol := p.Height - 1 - row
				cells[newRow*width+newCol] =
					p.Cells[row*p.Width+col]
			}
		}
		return Pattern{
			Width:  width,
			Height: height,
			Cells:  cells,
		}
	case Rotate180:
		cells := make([]bool, len(p.Cells))
		for i := range p.Cells {
			cells[len(p.Cells)-1-i] = p.Cells[i]
		}
		return Pattern{
			Width:  p.Width,
			Height: p.Height,
			Cells:  cells,
		}
	case Rotate270:
		cells := make([]bool, len(p.Cells))
		width, height := p.Height, p.Width
		for row := 0; row < p.Height; row++ {
			for col := 0; col < p.Width; col++ {
				newRow := p.Width - 1 - col
				newCol := row
				cells[newRow*width+newCol] =
					p.Cells[row*p.Width+col]
			}
		}
		return Pattern{
			Width:  width,
			Height: height,
			Cells:  cells,
		}
	default:
		panic("invalid rotation")
	}
}

func (p Pattern) Reflected() Pattern {
	cells := make([]bool, len(p.Cells))
	for row := 0; row < p.Height; row++ {
		for col := 0; col < p.Width; col++ {
			cells[row*p.Width+(p.Width-1-col)] = p.Cells[row*p.Width+col]
		}
	}
	return Pattern{
		Width:  p.Width,
		Height: p.Height,
		Cells:  cells,
	}
}

func (p Pattern) Hash() [32]byte {
	// 8 bytes for dimensions + packed alive cell bits...
	data := make([]byte, 8+(len(p.Cells)+7)/8)
	binary.LittleEndian.PutUint32(data[0:4], uint32(p.Width))
	binary.LittleEndian.PutUint32(data[4:8], uint32(p.Height))
	for i, alive := range p.Cells {
		if alive {
			data[8+i/8] |= 1 << (i % 8)
		}
	}
	return sha256.Sum256(data)
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

func (p Pattern) RotatedUp() Pattern {
	result := Pattern{
		Width:  p.Width,
		Height: p.Height,
		Rule:   p.Rule,
		Cells:  slices.Clone(p.Cells),
	}
	if p.Width <= 0 || p.Height <= 1 || len(p.Cells) != p.Width*p.Height {
		return result
	}
	width := result.Width
	firstRow := append([]bool(nil), result.Cells[:width]...)
	copy(result.Cells, result.Cells[width:])
	copy(result.Cells[len(result.Cells)-width:], firstRow)
	return result
}

func (p Pattern) RotatedDown() Pattern {
	result := Pattern{
		Width:  p.Width,
		Height: p.Height,
		Rule:   p.Rule,
		Cells:  slices.Clone(p.Cells),
	}
	if p.Width <= 0 || p.Height <= 1 || len(p.Cells) != p.Width*p.Height {
		return result
	}
	width := result.Width
	lastRowStart := len(result.Cells) - width
	lastRow := append([]bool(nil), result.Cells[lastRowStart:]...)
	copy(result.Cells[width:], result.Cells[:lastRowStart])
	copy(result.Cells[:width], lastRow)
	return result
}

func (p Pattern) RotatedLeft() Pattern {
	result := Pattern{
		Width:  p.Width,
		Height: p.Height,
		Rule:   p.Rule,
		Cells:  slices.Clone(p.Cells),
	}
	if p.Width <= 1 || p.Height <= 0 || len(p.Cells) != p.Width*p.Height {
		return result
	}
	for row := 0; row < result.Height; row++ {
		start := row * result.Width
		end := start + result.Width
		first := result.Cells[start]
		copy(result.Cells[start:end-1], result.Cells[start+1:end])
		result.Cells[end-1] = first
	}
	return result
}

func (p Pattern) RotatedRight() Pattern {
	result := Pattern{
		Width:  p.Width,
		Height: p.Height,
		Rule:   p.Rule,
		Cells:  slices.Clone(p.Cells),
	}
	if p.Width <= 1 || p.Height <= 0 || len(p.Cells) != p.Width*p.Height {
		return result
	}
	for row := 0; row < result.Height; row++ {
		start := row * result.Width
		end := start + result.Width
		last := result.Cells[end-1]
		copy(result.Cells[start+1:end], result.Cells[start:end-1])
		result.Cells[start] = last
	}
	return result
}
