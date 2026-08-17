package patterns

import (
	"github.com/marrow16/gogol/logic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestNewPattern(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		p, err := NewPattern("Test", 3, []bool{true, false, true})
		require.NoError(t, err)
		assert.Equal(t, "Test", p.Name)
	})
	t.Run("errors", func(t *testing.T) {
		_, err := NewPattern("Test", 3, []bool{true, false})
		require.Error(t, err)
	})
	t.Run("must", func(t *testing.T) {
		require.NotPanics(t, func() {
			_ = MustNewPattern("Test", 3, []bool{true, false, true})
		})
	})
	t.Run("must panics", func(t *testing.T) {
		require.Panics(t, func() {
			_ = MustNewPattern("Test", 3, []bool{true, false})
		})
	})
}

func TestNewPatternFromRle(t *testing.T) {
	const data = `#N Glider
x = 3, y = 3, rule = B3/S23
bob$2bo$3o!`
	t.Run("ok", func(t *testing.T) {
		p, err := NewPatternFromRle(strings.NewReader(data))
		require.NoError(t, err)
		assert.Equal(t, "Glider", p.Name)
		assert.Equal(t, 3, p.Height)
		assert.Equal(t, 3, p.Width)
		require.NotNil(t, p.Rule)
		assert.Equal(t, "B3/S23", p.Rule.Rle())
	})
	t.Run("errors", func(t *testing.T) {
		_, err := NewPatternFromRle(strings.NewReader(""))
		require.Error(t, err)
	})
	t.Run("must", func(t *testing.T) {
		require.NotPanics(t, func() {
			_ = MustNewPatternFromRle(strings.NewReader(data))
		})
	})
	t.Run("must panics", func(t *testing.T) {
		require.Panics(t, func() {
			_ = MustNewPatternFromRle(strings.NewReader(""))
		})
	})
}

func TestPattern_Draw(t *testing.T) {
	p, err := NewPattern("Glider", 5, []bool{
		false, false, false, false, false,
		false, false, false, true, false,
		false, true, false, true, false,
		false, false, true, true, false,
		false, false, false, false, false})
	require.NoError(t, err)

	g, err := logic.NewGrid(5, 5, 0, 0)
	require.NoError(t, err)

	p.Draw(g, 0, 0, Rotate0)
	assert.True(t, g.GetCell(1, 3).Alive)
	assert.True(t, g.GetCell(2, 1).Alive)
	assert.True(t, g.GetCell(2, 3).Alive)
	assert.True(t, g.GetCell(3, 2).Alive)
	assert.True(t, g.GetCell(3, 3).Alive)
}

func TestPattern_DrawTo(t *testing.T) {
	p, err := NewPattern("Glider", 3, []bool{
		false, false, true,
		true, false, true,
		false, true, true})
	require.NoError(t, err)

	positions := make([][]int, 0)
	fn := func(row, col int, alive bool) {
		positions = append(positions, []int{row, col})
	}

	p.DrawTo(Rotate0, fn)
	assert.Equal(t, [][]int{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}, {1, 2}, {2, 0}, {2, 1}, {2, 2}}, positions)

	positions = make([][]int, 0)
	p.DrawTo(Rotate90, fn)
	assert.Equal(t, [][]int{{0, 2}, {1, 2}, {2, 2}, {0, 1}, {1, 1}, {2, 1}, {0, 0}, {1, 0}, {2, 0}}, positions)

	positions = make([][]int, 0)
	p.DrawTo(Rotate180, fn)
	assert.Equal(t, [][]int{{2, 2}, {2, 1}, {2, 0}, {1, 2}, {1, 1}, {1, 0}, {0, 2}, {0, 1}, {0, 0}}, positions)

	positions = make([][]int, 0)
	p.DrawTo(Rotate270, fn)
	assert.Equal(t, [][]int{{2, 0}, {1, 0}, {0, 0}, {2, 1}, {1, 1}, {0, 1}, {2, 2}, {1, 2}, {0, 2}}, positions)

	p.DrawTo(Rotate0, nil)
}

func TestNewPatternFromGrid_Errors(t *testing.T) {
	_, err := NewPatternFromGrid(nil)
	require.Error(t, err)
}

func TestNewPatternFromGridPortion(t *testing.T) {
	g, err := logic.NewGrid(3, 3, 0, 0)
	require.NoError(t, err)
	g.SetCell(1, 1, true)
	g.SetCell(1, 2, true)
	g.SetCell(2, 1, true)
	g.SetCell(2, 2, true)
	p := NewPatternFromGridPortion(g, 1, 1, 3, 3)
	require.Equal(t, 2, p.Height)
	require.Equal(t, 2, p.Width)
	assert.Equal(t, []bool{true, true, true, true}, p.Cells)
}

func TestPattern_String(t *testing.T) {
	p := Pattern{Filename: "test"}
	assert.Equal(t, "test", p.String())
	p.Name = "Test"
	assert.Equal(t, "Test", p.String())
}

func TestPattern_Trimmed(t *testing.T) {
	p, err := NewPattern("Glider", 5, []bool{
		false, false, false, false, false,
		false, false, false, true, false,
		false, true, false, true, false,
		false, false, true, true, false,
		false, false, false, false, false})
	require.NoError(t, err)
	pt := p.Trimmed()
	require.Equal(t, 3, pt.Height)
	require.Equal(t, 3, pt.Width)
	assert.Equal(t, []bool{
		false, false, true,
		true, false, true,
		false, true, true,
	}, pt.Cells)
	pt = pt.Trimmed()
	require.Equal(t, 3, pt.Height)
	require.Equal(t, 3, pt.Width)
	assert.Equal(t, []bool{
		false, false, true,
		true, false, true,
		false, true, true,
	}, pt.Cells)

	p = Pattern{
		Width:  1,
		Height: 1,
		Cells:  []bool{false},
	}
	pt = p.Trimmed()
	require.Equal(t, 0, pt.Height)
	require.Equal(t, 0, pt.Width)
	require.Len(t, pt.Cells, 0)

	p = Pattern{}
	pt = p.Trimmed()
	require.Equal(t, 0, pt.Height)
	require.Equal(t, 0, pt.Width)
	require.Len(t, pt.Cells, 0)
}

func TestPattern_Reflected(t *testing.T) {
	p, err := NewPattern("Glider", 5, []bool{
		false, false, false, false, false,
		false, false, false, true, false,
		false, true, false, true, false,
		false, false, true, true, false,
		false, false, false, false, false})
	require.NoError(t, err)
	pr := p.Trimmed().Reflected()
	require.Equal(t, 3, pr.Height)
	require.Equal(t, 3, pr.Width)
	assert.Equal(t, []bool{
		true, false, false,
		true, false, true,
		true, true, false,
	}, pr.Cells)
}

func TestPattern_Rotated(t *testing.T) {
	p, err := NewPattern("", 5, []bool{
		false, false, false, false, false,
		false, false, false, true, true,
		false, true, false, true, false,
		false, false, true, true, false,
		false, false, false, false, false})
	require.NoError(t, err)
	p = p.Trimmed()
	require.Equal(t, 3, p.Height)
	require.Equal(t, 4, p.Width)

	pr := p.Rotated(Rotate0)
	require.Equal(t, 3, pr.Height)
	require.Equal(t, 4, pr.Width)
	require.Equal(t, pr.Cells, p.Cells)

	pr = p.Rotated(Rotate90)
	require.Equal(t, 4, pr.Height)
	require.Equal(t, 3, pr.Width)
	require.Equal(t, []bool{
		false, true, false,
		true, false, false,
		true, true, true,
		false, false, true}, pr.Cells)

	pr = p.Rotated(Rotate180)
	require.Equal(t, 3, pr.Height)
	require.Equal(t, 4, pr.Width)
	require.Equal(t, []bool{
		false, true, true, false,
		false, true, false, true,
		true, true, false, false}, pr.Cells)

	pr = p.Rotated(Rotate270)
	require.Equal(t, 4, pr.Height)
	require.Equal(t, 3, pr.Width)
	require.Equal(t, []bool{
		true, false, false,
		true, true, true,
		false, false, true,
		false, true, false}, pr.Cells)

	require.Panics(t, func() {
		p.Rotated(-1)
	})
}

func TestPattern_Hash(t *testing.T) {
	p, err := NewPattern("", 5, []bool{
		false, false, false, false, false,
		false, false, false, true, true,
		false, true, false, true, false,
		false, false, true, true, false,
		false, false, false, false, false})
	require.NoError(t, err)
	p = p.Trimmed()
	require.Equal(t, 3, p.Height)
	require.Equal(t, 4, p.Width)
	hash := p.Hash()
	require.Len(t, hash, 32)
	require.Equal(t, [32]byte{0x7f, 0x29, 0x13, 0xa2, 0x76, 0xaa, 0xa4, 0xf, 0x75, 0x4b, 0x88, 0x5c, 0x67, 0xb4, 0x51, 0xfb, 0xc3, 0x59, 0x48, 0xd1, 0xd, 0x97, 0x6d, 0xbe, 0x9e, 0xf3, 0x4, 0x1b, 0x0, 0x8b, 0x1d, 0xad}, hash)
}

func TestPattern_AllHashes(t *testing.T) {
	p, err := NewPattern("", 5, []bool{
		false, false, false, false, false,
		false, false, false, true, true,
		false, true, false, true, false,
		false, false, true, true, false,
		false, false, false, false, false})
	require.NoError(t, err)
	p = p.Trimmed()
	require.Equal(t, 3, p.Height)
	require.Equal(t, 4, p.Width)

	ah := p.AllHashes()
	require.Len(t, ah, 8)
	require.Equal(t, p.Hash(), ah[0])

	p = Pattern{
		Width:  2,
		Height: 1,
		Cells:  []bool{false, false},
	}
	ah = p.AllHashes()
	require.Len(t, ah, 2)

	p = Pattern{
		Width:  2,
		Height: 1,
		Cells:  []bool{true, false},
	}
	ah = p.AllHashes()
	require.Len(t, ah, 4)

	p = Pattern{
		Width:  1,
		Height: 1,
		Cells:  []bool{false},
	}
	ah = p.AllHashes()
	require.Len(t, ah, 1)
}

func TestPattern_RotatedUp(t *testing.T) {
	p, err := NewPattern("", 3, []bool{
		false, false, true,
		true, false, true,
		false, true, true})
	require.NoError(t, err)
	pr := p.RotatedUp()
	require.Equal(t, 3, pr.Height)
	require.Equal(t, 3, pr.Width)
	require.Equal(t, []bool{
		true, false, true,
		false, true, true,
		false, false, true}, pr.Cells)

	p = Pattern{Width: 1, Height: 1, Cells: []bool{true}}
	pr = p.RotatedUp()
	require.True(t, pr.Cells[0])
}

func TestPattern_RotatedDown(t *testing.T) {
	p, err := NewPattern("", 3, []bool{
		false, false, true,
		true, false, true,
		false, true, true})
	require.NoError(t, err)
	pr := p.RotatedDown()
	require.Equal(t, 3, pr.Height)
	require.Equal(t, 3, pr.Width)
	require.Equal(t, []bool{
		false, true, true,
		false, false, true,
		true, false, true}, pr.Cells)

	p = Pattern{Width: 1, Height: 1, Cells: []bool{true}}
	pr = p.RotatedDown()
	require.True(t, pr.Cells[0])
}

func TestPattern_RotatedLeft(t *testing.T) {
	p, err := NewPattern("", 3, []bool{
		false, false, true,
		true, false, true,
		false, true, true})
	require.NoError(t, err)
	pr := p.RotatedLeft()
	require.Equal(t, 3, pr.Height)
	require.Equal(t, 3, pr.Width)
	require.Equal(t, []bool{
		false, true, false,
		false, true, true,
		true, true, false}, pr.Cells)

	p = Pattern{Width: 1, Height: 1, Cells: []bool{true}}
	pr = p.RotatedLeft()
	require.True(t, pr.Cells[0])
}

func TestPattern_RotatedRight(t *testing.T) {
	p, err := NewPattern("", 3, []bool{
		false, false, true,
		true, false, true,
		false, true, true})
	require.NoError(t, err)
	pr := p.RotatedRight()
	require.Equal(t, 3, pr.Height)
	require.Equal(t, 3, pr.Width)
	require.Equal(t, []bool{
		true, false, false,
		true, true, false,
		true, false, true}, pr.Cells)

	p = Pattern{Width: 1, Height: 1, Cells: []bool{true}}
	pr = p.RotatedRight()
	require.True(t, pr.Cells[0])
}

func TestPattern_Phases(t *testing.T) {
	p, err := NewPattern("", 3, []bool{
		false, false, true,
		true, false, true,
		false, true, true})
	require.NoError(t, err)

	phases, reason := p.Phases(50, 4, 10)
	require.Len(t, phases, 4)
	require.Equal(t, PhaseRepeat, reason)
	assert.Equal(t, []bool{
		false, false, true,
		true, false, true,
		false, true, true}, phases[0].Cells)
	assert.Equal(t, []bool{
		true, false, false,
		false, true, true,
		true, true, false}, phases[1].Cells)
	assert.Equal(t, []bool{
		false, true, false,
		false, false, true,
		true, true, true}, phases[2].Cells)
	assert.Equal(t, []bool{
		true, false, true,
		false, true, true,
		false, true, false}, phases[3].Cells)

	_, reason = p.Phases(0, 0, 0)
	require.Equal(t, PhaseLimit, reason)

	p = MustNewPattern("Copperhead spaceship", 10, []bool{
		false, false, false, false, false, false, false, false, false, false,
		false, false, true, true, false, false, true, true, false, false,
		false, false, false, false, true, true, false, false, false, false,
		false, false, false, false, true, true, false, false, false, false,
		false, true, false, true, false, false, true, false, true, false,
		false, true, false, false, false, false, false, false, true, false,
		false, false, false, false, false, false, false, false, false, false,
		false, true, false, false, false, false, false, false, true, false,
		false, false, true, true, false, false, true, true, false, false,
		false, false, false, true, true, true, true, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, true, true, false, false, false, false,
		false, false, false, false, true, true, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
	})
	phases, reason = p.Phases(50, 4, 10)
	require.Len(t, phases, 10)
	require.Equal(t, PhaseRepeat, reason)

	p = MustNewPattern("Beacon", 6, []bool{
		false, false, false, false, false, false,
		false, true, true, false, false, false,
		false, true, true, false, false, false,
		false, false, false, true, true, false,
		false, false, false, true, true, false,
		false, false, false, false, false, false,
	})
	phases, reason = p.Phases(50, 4, 10)
	require.Len(t, phases, 2)
	require.Equal(t, PhaseRepeat, reason)

	p, _ = NewPattern("", 3, []bool{
		false, true, false,
		true, false, true,
		true, false, true,
		false, true, false})
	phases, reason = p.Phases(50, 4, 10)
	require.Len(t, phases, 1)
	require.Equal(t, PhaseStopped, reason)

	p, _ = NewPattern("", 1, []bool{true})
	phases, reason = p.Phases(50, 4, 10)
	require.Len(t, phases, 1)
	require.Equal(t, PhaseExtinct, reason)

	p, err = PatternRleDecoder(strings.NewReader(`x = 3, y = 4, rule = B3/S23
2bo$b2o$2o$o!`))
	require.NoError(t, err)
	phases, reason = p.Phases(50, 2, 10)
	require.Len(t, phases, 3)
	require.Equal(t, PhaseGrowth, reason)

	p, err = PatternRleDecoder(strings.NewReader(`x = 7, y = 7, rule = B3/S23
o2bo2bo$7b$7b$2b3o$7b$7b$o2bo2bo!`))
	require.NoError(t, err)
	phases, reason = p.Phases(50, 2, 10)
	require.Len(t, phases, 1)
	require.Equal(t, PhaseDecay, reason)
}

func TestRotation_String(t *testing.T) {
	assert.Equal(t, "0°", Rotate0.String())
	assert.Equal(t, "0°", Rotation(-1).String())
	assert.Equal(t, "90°", Rotate90.String())
	assert.Equal(t, "180°", Rotate180.String())
	assert.Equal(t, "270°", Rotate270.String())
}
