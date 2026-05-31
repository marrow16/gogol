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

func (p Pattern) Draw(grid *logic.Grid, row, col int) {
	idx := 0
	for y := 0; y < p.Height; y++ {
		for x := 0; x < p.Width; x++ {
			grid.SetCell(row+y, col+x, p.Cells[idx])
			idx++
		}
	}
}
