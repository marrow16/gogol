package logic

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewGrid(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		g, err := NewGrid(10, 10, WrapAll, DeadBoundary)
		require.NoError(t, err)
		require.NotNil(t, g)
	})
	t.Run("bad height", func(t *testing.T) {
		_, err := NewGrid(1, 10, WrapAll, DeadBoundary)
		require.Error(t, err)
		require.Equal(t, ErrInvalidGridDimension, err)
	})
	t.Run("bad width", func(t *testing.T) {
		_, err := NewGrid(10, 1, WrapAll, DeadBoundary)
		require.Error(t, err)
		require.Equal(t, ErrInvalidGridDimension, err)
	})
}

type cellCoord [2]int

const (
	ROW = 0
	COL = 1
)

func TestGrid_AdjacencyAliveCounts(t *testing.T) {
	testCases := []struct {
		name         string
		wrapMode     WrapMode
		boundaryMode BoundaryMode
		alives       []cellCoord
		cell         cellCoord
		expectAlives int
	}{
		{
			name:         "corner no wrap dead boundary",
			wrapMode:     WrapNone,
			boundaryMode: DeadBoundary,
			cell:         cellCoord{0, 0},
			expectAlives: 0,
		},
		{
			name:         "corner no wrap alive boundary",
			wrapMode:     WrapNone,
			boundaryMode: AliveBoundary,
			cell:         cellCoord{0, 0},
			expectAlives: 5,
		},
		{
			name:         "corner sees three real neighbours",
			wrapMode:     WrapNone,
			boundaryMode: DeadBoundary,
			alives:       []cellCoord{{0, 1}, {1, 0}, {1, 1}},
			cell:         cellCoord{0, 0},
			expectAlives: 3,
		},
		{
			name:         "corner wraps horizontally",
			wrapMode:     WrapHorizontal,
			boundaryMode: DeadBoundary,
			alives:       []cellCoord{{0, 9}, {1, 9}},
			cell:         cellCoord{0, 0},
			expectAlives: 2,
		},
		{
			name:         "corner wraps vertically",
			wrapMode:     WrapVertical,
			boundaryMode: DeadBoundary,
			alives:       []cellCoord{{9, 0}, {9, 1}},
			cell:         cellCoord{0, 0},
			expectAlives: 2,
		},
		{
			name:         "corner wraps all",
			wrapMode:     WrapAll,
			boundaryMode: DeadBoundary,
			alives:       []cellCoord{{0, 9}, {9, 0}, {9, 9}},
			cell:         cellCoord{0, 0},
			expectAlives: 3,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := NewGrid(10, 10, tc.wrapMode, tc.boundaryMode)
			require.NoError(t, err)
			for _, a := range tc.alives {
				g.SetCell(a[ROW], a[COL], true)
			}
			c := g.GetCell(tc.cell[ROW], tc.cell[COL])
			require.NotNil(t, c)
			assert.Equal(t, tc.expectAlives, c.AdjacentsAlive())
		})
	}
}

func TestGrid_AdjacencyIdentity_WrapAllCorner(t *testing.T) {
	g, err := NewGrid(10, 10, WrapAll, DeadBoundary)
	require.NoError(t, err)
	c := g.GetCell(0, 0)
	assert.Same(t, g.GetCell(9, 9), c.Adjacents[0])
	assert.Same(t, g.GetCell(9, 0), c.Adjacents[1])
	assert.Same(t, g.GetCell(9, 1), c.Adjacents[2])
	assert.Same(t, g.GetCell(0, 9), c.Adjacents[3])
	assert.Same(t, g.GetCell(0, 1), c.Adjacents[4])
	assert.Same(t, g.GetCell(1, 9), c.Adjacents[5])
	assert.Same(t, g.GetCell(1, 0), c.Adjacents[6])
	assert.Same(t, g.GetCell(1, 1), c.Adjacents[7])
}

func TestGrid_SetBoundaryMode(t *testing.T) {
	g, err := NewGrid(2, 2, WrapNone, DeadBoundary)
	require.NoError(t, err)
	g.SetCell(0, 0, true)
	g.SetCell(0, 1, true)
	g.SetCell(1, 0, true)
	g.SetCell(1, 1, true)

	assert.Equal(t, 3, g.GetCell(0, 0).AdjacentsAlive())
	assert.Equal(t, 3, g.GetCell(0, 1).AdjacentsAlive())
	assert.Equal(t, 3, g.GetCell(1, 0).AdjacentsAlive())
	assert.Equal(t, 3, g.GetCell(1, 1).AdjacentsAlive())

	g.SetBoundaryMode(AliveBoundary)
	assert.Equal(t, 8, g.GetCell(0, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(0, 1).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 1).AdjacentsAlive())
}

func TestGrid_SetWrapMode(t *testing.T) {
	g, err := NewGrid(3, 3, WrapNone, DeadBoundary)
	require.NoError(t, err)
	g.SetCell(0, 0, true)
	g.SetCell(0, 1, true)
	g.SetCell(0, 2, true)
	g.SetCell(1, 0, true)
	g.SetCell(1, 1, true)
	g.SetCell(1, 2, true)
	g.SetCell(2, 0, true)
	g.SetCell(2, 1, true)
	g.SetCell(2, 2, true)

	assert.Equal(t, 3, g.GetCell(0, 0).AdjacentsAlive())
	assert.Equal(t, 5, g.GetCell(0, 1).AdjacentsAlive())
	assert.Equal(t, 3, g.GetCell(0, 2).AdjacentsAlive())
	assert.Equal(t, 5, g.GetCell(1, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 1).AdjacentsAlive())
	assert.Equal(t, 5, g.GetCell(1, 2).AdjacentsAlive())
	assert.Equal(t, 3, g.GetCell(2, 0).AdjacentsAlive())
	assert.Equal(t, 5, g.GetCell(2, 1).AdjacentsAlive())
	assert.Equal(t, 3, g.GetCell(2, 2).AdjacentsAlive())

	g.SetWrapMode(WrapAll)
	assert.Equal(t, 8, g.GetCell(0, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(0, 1).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(0, 2).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 1).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 2).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(2, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(2, 1).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(2, 2).AdjacentsAlive())

	g.SetWrapMode(WrapHorizontal)
	assert.Equal(t, 5, g.GetCell(0, 0).AdjacentsAlive())
	assert.Equal(t, 5, g.GetCell(0, 1).AdjacentsAlive())
	assert.Equal(t, 5, g.GetCell(0, 2).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 1).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 2).AdjacentsAlive())
	assert.Equal(t, 5, g.GetCell(2, 0).AdjacentsAlive())
	assert.Equal(t, 5, g.GetCell(2, 1).AdjacentsAlive())
	assert.Equal(t, 5, g.GetCell(2, 2).AdjacentsAlive())

	g.SetWrapMode(WrapVertical)
	assert.Equal(t, 5, g.GetCell(0, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(0, 1).AdjacentsAlive())
	assert.Equal(t, 5, g.GetCell(0, 2).AdjacentsAlive())
	assert.Equal(t, 5, g.GetCell(1, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 1).AdjacentsAlive())
	assert.Equal(t, 5, g.GetCell(1, 2).AdjacentsAlive())
	assert.Equal(t, 5, g.GetCell(2, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(2, 1).AdjacentsAlive())
	assert.Equal(t, 5, g.GetCell(2, 2).AdjacentsAlive())

	g.SetWrapMode(WrapHorizontal)
	g.SetBoundaryMode(AliveBoundary)
	assert.Equal(t, 8, g.GetCell(0, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(0, 1).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(0, 2).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 1).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 2).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(2, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(2, 1).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(2, 2).AdjacentsAlive())

	g.SetWrapMode(WrapVertical)
	assert.Equal(t, 8, g.GetCell(0, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(0, 1).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(0, 2).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 1).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(1, 2).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(2, 0).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(2, 1).AdjacentsAlive())
	assert.Equal(t, 8, g.GetCell(2, 2).AdjacentsAlive())
}

func TestGrid_GetCell(t *testing.T) {
	g, err := NewGrid(2, 2, WrapNone, DeadBoundary)
	require.NoError(t, err)
	assert.NotNil(t, g.GetCell(0, 0))
	assert.NotNil(t, g.GetCell(0, 1))
	assert.NotNil(t, g.GetCell(1, 0))
	assert.NotNil(t, g.GetCell(1, 1))
	assert.Nil(t, g.GetCell(-1, -1))
	assert.Nil(t, g.GetCell(0, 2))
	assert.Nil(t, g.GetCell(2, 0))
	assert.Nil(t, g.GetCell(2, 2))
}

func TestGrid_Step(t *testing.T) {
	t.Run("Beacon", func(t *testing.T) {
		g, err := NewGrid(6, 6, WrapAll, DeadBoundary)
		require.NoError(t, err)

		cellsChanged := 0
		g.Render = func(row, col int, alive, changed bool) {
			if changed {
				cellsChanged++
			}
		}

		g.SetCell(1, 1, true)
		g.SetCell(1, 2, true)
		g.SetCell(2, 1, true)
		g.SetCell(2, 2, true)
		g.SetCell(3, 3, true)
		g.SetCell(3, 4, true)
		g.SetCell(4, 3, true)
		g.SetCell(4, 4, true)
		testDisplay(g)

		cellsChanged = 0
		gridChanged := g.Step()
		assert.True(t, gridChanged)
		assert.Equal(t, 2, cellsChanged)
		assert.False(t, g.GetCell(2, 2).Alive)
		assert.False(t, g.GetCell(3, 3).Alive)
		testDisplay(g)

		cellsChanged = 0
		gridChanged = g.Step()
		assert.True(t, gridChanged)
		assert.Equal(t, 2, cellsChanged)
		assert.True(t, g.GetCell(2, 2).Alive)
		assert.True(t, g.GetCell(3, 3).Alive)
		testDisplay(g)
	})
	t.Run("Toad", func(t *testing.T) {
		g, err := NewGrid(6, 6, WrapAll, DeadBoundary)
		require.NoError(t, err)

		cellsChanged := 0
		g.Render = func(row, col int, alive, changed bool) {
			if changed {
				cellsChanged++
			}
		}

		g.SetCell(2, 1, true)
		g.SetCell(2, 2, true)
		g.SetCell(2, 3, true)
		g.SetCell(3, 2, true)
		g.SetCell(3, 3, true)
		g.SetCell(3, 4, true)
		testDisplay(g)

		cellsChanged = 0
		gridChanged := g.Step()
		assert.True(t, gridChanged)
		assert.Equal(t, 8, cellsChanged)
		testDisplay(g)

		cellsChanged = 0
		gridChanged = g.Step()
		assert.True(t, gridChanged)
		assert.Equal(t, 8, cellsChanged)
		testDisplay(g)
	})
	t.Run("Glider", func(t *testing.T) {
		g, err := NewGrid(5, 5, WrapAll, DeadBoundary)
		require.NoError(t, err)

		cellsChanged := 0
		g.Render = func(row, col int, alive, changed bool) {
			if changed {
				cellsChanged++
			}
		}

		g.SetCell(1, 3, true)
		g.SetCell(2, 1, true)
		g.SetCell(2, 3, true)
		g.SetCell(3, 2, true)
		g.SetCell(3, 3, true)
		testDisplay(g)

		for s := 0; s < 20; s++ {
			cellsChanged = 0
			gridChanged := g.Step()
			assert.True(t, gridChanged)
			assert.Equal(t, 4, cellsChanged)
		}
		// glider should be back where it started...
		testDisplay(g)
		assert.True(t, g.GetCell(1, 3).Alive)
		assert.True(t, g.GetCell(2, 1).Alive)
		assert.True(t, g.GetCell(2, 3).Alive)
		assert.True(t, g.GetCell(3, 2).Alive)
		assert.True(t, g.GetCell(3, 3).Alive)
	})
	t.Run("no rule, no render func", func(t *testing.T) {
		g, err := NewGrid(2, 2, WrapAll, DeadBoundary)
		require.NoError(t, err)
		g.SetCell(0, 0, true)
		g.SetCell(0, 1, true)
		g.SetCell(1, 0, true)
		g.SetCell(1, 1, true)
		g.Rule = nil
		gridChanged := g.Step()
		assert.True(t, gridChanged)
		assert.False(t, g.GetCell(0, 0).Alive)
		assert.False(t, g.GetCell(0, 1).Alive)
		assert.False(t, g.GetCell(1, 0).Alive)
		assert.False(t, g.GetCell(1, 1).Alive)
	})
}

func testDisplay(g *Grid) {
	fmt.Println("")
	for _, row := range g.Rows {
		for _, cell := range row {
			if cell.Alive {
				fmt.Print("█")
			} else {
				fmt.Print("░")
			}
		}
		fmt.Println()
	}
}
