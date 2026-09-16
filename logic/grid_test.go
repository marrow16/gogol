package logic

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewGrid(t *testing.T) {
	testCases := []struct {
		wrapMode          WrapMode
		boundaryMode      BoundaryMode
		expectRenderCalls int
	}{
		{
			wrapMode:     WrapNone,
			boundaryMode: DeadBoundary,
		},
		{
			wrapMode:          WrapNone,
			boundaryMode:      AliveBoundary,
			expectRenderCalls: 4,
		},
		{
			wrapMode:     WrapHorizontal,
			boundaryMode: DeadBoundary,
		},
		{
			wrapMode:          WrapHorizontal,
			boundaryMode:      AliveBoundary,
			expectRenderCalls: 6,
		},
		{
			wrapMode:     WrapVertical,
			boundaryMode: DeadBoundary,
		},
		{
			wrapMode:          WrapVertical,
			boundaryMode:      AliveBoundary,
			expectRenderCalls: 6,
		},
		{
			wrapMode: WrapAll,
		},
		{
			wrapMode: WrapAll,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			g, err := NewGrid(3, 3, StandardRule, tc.wrapMode, tc.boundaryMode)
			require.NoError(t, err)
			require.Equal(t, tc.wrapMode, g.WrapMode())
			require.Equal(t, tc.boundaryMode, g.BoundaryMode())
			require.Equal(t, 3, g.Height())
			require.Equal(t, 3, g.Width())
			require.Equal(t, StandardRule, g.Rule())
			require.Equal(t, 9, len(g.cells))
			assert.Equal(t, [2][9]bool{{false, false, false, true, false, false, false, false, false}, {true, true, false, false, true, true, true, true, true}}, g.ruleCache)
			assert.NotNil(t, g.changeFns[0])
			assert.NotNil(t, g.changeFns[1])
			assert.NotNil(t, g.changeFns[2])
			assert.Nil(t, g.changeFns[3])
			assert.NotNil(t, g.changeFns[4])
			assert.NotNil(t, g.changeFns[5])
			assert.NotNil(t, g.changeFns[6])
			assert.Nil(t, g.changeFns[7])
			assert.NotNil(t, g.changeFns[8])
			assert.NotNil(t, g.changeFns[9])
			assert.NotNil(t, g.changeFns[10])
			assert.Nil(t, g.changeFns[11])
			assert.Nil(t, g.changeFns[12])
			assert.Nil(t, g.changeFns[13])
			assert.Nil(t, g.changeFns[14])
			assert.Nil(t, g.changeFns[15])
			called := 0
			g.SetRenderer(func(row, col int, alive, changed bool) {
				called++
			})
			// check coverage on change funcs...
			g.Step()
			assert.Equal(t, tc.expectRenderCalls, called)
		})
	}
}

func TestNewGrid_Errors(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		g, err := NewGrid(10, 10, StandardRule, WrapAll, DeadBoundary)
		require.NoError(t, err)
		require.NotNil(t, g)
	})
	t.Run("bad height", func(t *testing.T) {
		_, err := NewGrid(1, 10, StandardRule, WrapAll, DeadBoundary)
		require.Error(t, err)
		require.Equal(t, ErrInvalidGridDimension, err)
	})
	t.Run("bad width", func(t *testing.T) {
		_, err := NewGrid(10, 1, StandardRule, WrapAll, DeadBoundary)
		require.Error(t, err)
		require.Equal(t, ErrInvalidGridDimension, err)
	})
}

func TestGrid_SetRenderer(t *testing.T) {
	g, err := NewGrid(3, 3, StandardRule, WrapAll, AliveBoundary)
	require.NoError(t, err)
	require.NotNil(t, g.renderer)
	g.SetRenderer(nil)
	require.NotNil(t, g.renderer)
	g.SetRenderer(func(row, col int, alive, changed bool) {})
	require.NotNil(t, g.renderer)
}

func TestGrid_SetRule(t *testing.T) {
	g, err := NewGrid(3, 3, StandardRule, WrapAll, AliveBoundary)
	require.NoError(t, err)
	cache := g.ruleCache
	g.SetRule(Rules["AntiLife"])
	assert.NotEqual(t, cache, g.ruleCache)
	g.SetRule(StandardRule)
	assert.Equal(t, cache, g.ruleCache)
}

func TestGrid_SetWrapMode(t *testing.T) {
	g, err := NewGrid(3, 3, StandardRule, WrapAll, AliveBoundary)
	require.NoError(t, err)
	assert.Equal(t, WrapAll, g.WrapMode())
	g.SetWrapMode(WrapHorizontal)
	assert.Equal(t, WrapHorizontal, g.WrapMode())
}

func TestGrid_SetBoundaryMode(t *testing.T) {
	g, err := NewGrid(3, 3, StandardRule, WrapAll, AliveBoundary)
	require.NoError(t, err)
	assert.Equal(t, AliveBoundary, g.BoundaryMode())
	g.SetBoundaryMode(DeadBoundary)
	assert.Equal(t, DeadBoundary, g.BoundaryMode())
}

func TestGrid_SetAll(t *testing.T) {
	g, err := NewGrid(3, 3, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	assert.Equal(t, StandardRule, g.Rule())
	assert.Equal(t, WrapAll, g.WrapMode())
	assert.Equal(t, DeadBoundary, g.BoundaryMode())
	cache := g.ruleCache
	g.SetAll(Rules["AntiLife"], WrapNone, AliveBoundary)
	assert.NotEqual(t, StandardRule, g.Rule())
	assert.Equal(t, WrapNone, g.WrapMode())
	assert.Equal(t, AliveBoundary, g.BoundaryMode())
	assert.NotEqual(t, cache, g.ruleCache)
}

func TestGrid_GetCell(t *testing.T) {
	g, err := NewGrid(3, 3, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	assert.False(t, g.GetCell(0, 0))
	g.SetCell(0, 0, true)
	assert.True(t, g.GetCell(0, 0))

	assert.False(t, g.GetCell(10, 10))
}

func TestGrid_SetCell(t *testing.T) {
	g, err := NewGrid(3, 3, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	g.StepCount.Store(1)
	tr := &testRenderer{}
	g.SetRenderer(tr.render)
	assert.True(t, g.SetCell(0, 0, true))
	assert.Equal(t, 1, tr.called)
	assert.Equal(t, 1, tr.changes)
	assert.Equal(t, 1, tr.alives)
	assert.Equal(t, uint64(0), g.StepCount.Load())
	assert.False(t, g.SetCell(0, 0, true))
	assert.Equal(t, 2, tr.called)
	assert.Equal(t, 1, tr.changes)
	assert.Equal(t, 2, tr.alives)
	assert.True(t, g.SetCell(0, 0, false))
	assert.Equal(t, 3, tr.called)
	assert.Equal(t, 2, tr.changes)
	assert.Equal(t, 2, tr.alives)
	assert.Equal(t, 1, tr.deads)
	assert.False(t, g.SetCell(0, 0, false))
	assert.Equal(t, 4, tr.called)
	assert.Equal(t, 2, tr.changes)
	assert.Equal(t, 2, tr.alives)
	assert.Equal(t, 2, tr.deads)
}

type testRenderer struct {
	called  int
	changes int
	alives  int
	deads   int
}

func (t *testRenderer) render(_, _ int, alive, changed bool) {
	t.called++
	if changed {
		t.changes++
	}
	if alive {
		t.alives++
	} else {
		t.deads++
	}
}

func (t *testRenderer) to(row, col int, alive bool) {
	t.render(row, col, alive, true)
}

func TestGrid_Clear(t *testing.T) {
	g, err := NewGrid(3, 3, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	g.RandomizePopulation(100)
	assert.Equal(t, 9, g.Population())
	g.StepCount.Store(1)

	tr := &testRenderer{}
	g.SetRenderer(tr.render)
	g.Clear()
	assert.Equal(t, uint64(0), g.StepCount.Load())
	assert.Equal(t, 0, g.Population())
	assert.Equal(t, 9, tr.called)
	assert.Equal(t, 9, tr.changes)
	assert.Equal(t, 0, tr.alives)
	assert.Equal(t, 9, tr.deads)
}

func TestGrid_Randomize(t *testing.T) {
	g, err := NewGrid(10, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	tr := &testRenderer{}
	g.SetRenderer(tr.render)

	g.StepCount.Store(1)
	g.Randomize(50)
	assert.Greater(t, g.Population(), 30)
	assert.Less(t, g.Population(), 70)
	assert.Equal(t, 100, tr.called)
	assert.Equal(t, 100, tr.changes)
	assert.Equal(t, g.Population(), tr.alives)
	assert.Equal(t, 100-g.Population(), tr.deads)
	assert.Equal(t, uint64(0), g.StepCount.Load())
}

func TestGrid_RandomChanges(t *testing.T) {
	g, err := NewGrid(10, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	tr := &testRenderer{}
	g.SetRenderer(tr.render)

	g.StepCount.Store(1)
	g.RandomChanges(50)
	assert.Equal(t, tr.changes, tr.called)
}

func TestGrid_RandomAdditions(t *testing.T) {
	g, err := NewGrid(100, 100, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	tr := &testRenderer{}
	g.SetRenderer(tr.render)

	g.RandomizePopulation(10)
	assert.Equal(t, 1000, g.Population())
	assert.Equal(t, 10000, tr.called)
	assert.Equal(t, 10000, tr.changes)
	assert.Equal(t, 1000, tr.alives)
	assert.Equal(t, 9000, tr.deads)

	g.RandomAdditions(10)
	assert.Equal(t, 1900, g.Population())
	assert.Equal(t, 10900, tr.called)
	assert.Equal(t, 10900, tr.changes)
	assert.Equal(t, 1900, tr.alives)
	assert.Equal(t, 9000, tr.deads)

	g.RandomAdditions(0)
	assert.Equal(t, 1900, g.Population())
	assert.Equal(t, 10900, tr.called)
	assert.Equal(t, 10900, tr.changes)
	assert.Equal(t, 1900, tr.alives)
	assert.Equal(t, 9000, tr.deads)
}

func TestGrid_RandomCull(t *testing.T) {
	g, err := NewGrid(100, 100, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	tr := &testRenderer{}
	g.SetRenderer(tr.render)

	g.RandomizePopulation(10)
	assert.Equal(t, 1000, g.Population())
	assert.Equal(t, 10000, tr.called)
	assert.Equal(t, 10000, tr.changes)
	assert.Equal(t, 1000, tr.alives)
	assert.Equal(t, 9000, tr.deads)

	g.RandomCull(50)
	assert.Equal(t, 500, g.Population())
	assert.Equal(t, 10500, tr.called)
	assert.Equal(t, 10500, tr.changes)
	assert.Equal(t, 1000, tr.alives)
	assert.Equal(t, 9500, tr.deads)

	g.RandomCull(0)
	assert.Equal(t, 500, g.Population())
	assert.Equal(t, 10500, tr.called)
	assert.Equal(t, 10500, tr.changes)
	assert.Equal(t, 1000, tr.alives)
	assert.Equal(t, 9500, tr.deads)
}

func TestGrid_RandomizePopulation(t *testing.T) {
	g, err := NewGrid(10, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	tr := &testRenderer{}
	g.SetRenderer(tr.render)

	g.StepCount.Store(1)
	g.RandomizePopulation(100)
	assert.Equal(t, 100, g.Population())
	assert.Equal(t, 100, tr.called)
	assert.Equal(t, 100, tr.changes)
	assert.Equal(t, 100, tr.alives)
	assert.Equal(t, 0, tr.deads)
	assert.Equal(t, uint64(0), g.StepCount.Load())

	g.StepCount.Store(1)
	g.RandomizePopulation(0)
	assert.Equal(t, 0, g.Population())
	assert.Equal(t, 200, tr.called)
	assert.Equal(t, 200, tr.changes)
	assert.Equal(t, 100, tr.alives)
	assert.Equal(t, 100, tr.deads)
	assert.Equal(t, uint64(0), g.StepCount.Load())

	g.StepCount.Store(1)
	g.RandomizePopulation(50)
	assert.Equal(t, 50, g.Population())
	assert.Equal(t, 300, tr.called)
	assert.Equal(t, 300, tr.changes)
	assert.Equal(t, 150, tr.alives)
	assert.Equal(t, 150, tr.deads)
	assert.Equal(t, uint64(0), g.StepCount.Load())
}

func TestGrid_Draw(t *testing.T) {
	g, err := NewGrid(10, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	g.RandomizePopulation(50)

	tr := &testRenderer{}
	g.SetRenderer(tr.render)
	g.Draw()
	assert.Equal(t, 100, tr.called)
	assert.Equal(t, 100, tr.changes)
	assert.Equal(t, 50, tr.alives)
	assert.Equal(t, 50, tr.deads)
}

func TestGrid_DrawTo(t *testing.T) {
	g, err := NewGrid(10, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	g.RandomizePopulation(50)

	tr := &testRenderer{}
	g.DrawTo(tr.to)
	assert.Equal(t, 100, tr.called)
	assert.Equal(t, 100, tr.changes)
	assert.Equal(t, 50, tr.alives)
	assert.Equal(t, 50, tr.deads)
}

func TestGrid_LimitAliveAdjacents(t *testing.T) {
	t.Run("toroidal", func(t *testing.T) {
		g, err := NewGrid(10, 10, StandardRule, WrapAll, DeadBoundary)
		require.NoError(t, err)
		g.RandomizePopulation(100)
		require.Equal(t, 100, g.Population())
		tr := &testRenderer{}
		g.SetRenderer(tr.render)
		g.LimitAliveAdjacents(2)
		for idx := range g.cells {
			assert.LessOrEqual(t, g.aliveAdjacentCount(idx), uint8(2))
		}
		assert.Greater(t, tr.called, 0)
		g.RandomizePopulation(100)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(8)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(0)
		require.Equal(t, 0, g.Population())
	})
	t.Run("wrap vertical, dead boundary", func(t *testing.T) {
		g, err := NewGrid(10, 10, StandardRule, WrapVertical, DeadBoundary)
		require.NoError(t, err)
		g.RandomizePopulation(100)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(2)
		for idx := range g.cells {
			assert.LessOrEqual(t, g.aliveAdjacentCount(idx), uint8(2))
		}

		g.RandomizePopulation(100)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(8)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(0)
		require.Equal(t, 0, g.Population())
	})
	t.Run("wrap vertical, alive boundary", func(t *testing.T) {
		g, err := NewGrid(10, 10, StandardRule, WrapVertical, AliveBoundary)
		require.NoError(t, err)
		g.RandomizePopulation(100)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(3)
		for idx := range g.cells {
			assert.LessOrEqual(t, g.aliveAdjacentCount(idx), uint8(3))
		}

		g.RandomizePopulation(100)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(8)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(0)
		require.Equal(t, 0, g.Population())
	})
	t.Run("wrap horizontal, dead boundary", func(t *testing.T) {
		g, err := NewGrid(10, 10, StandardRule, WrapHorizontal, DeadBoundary)
		require.NoError(t, err)
		g.RandomizePopulation(100)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(2)
		for idx := range g.cells {
			assert.LessOrEqual(t, g.aliveAdjacentCount(idx), uint8(2))
		}

		g.RandomizePopulation(100)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(8)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(0)
		require.Equal(t, 0, g.Population())
	})
	t.Run("wrap horizontal, alive boundary", func(t *testing.T) {
		g, err := NewGrid(10, 10, StandardRule, WrapHorizontal, AliveBoundary)
		require.NoError(t, err)
		g.RandomizePopulation(100)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(3)
		for idx := range g.cells {
			assert.LessOrEqual(t, g.aliveAdjacentCount(idx), uint8(3))
		}

		g.RandomizePopulation(100)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(8)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(0)
		require.Equal(t, 0, g.Population())
	})
	t.Run("wrap none, dead boundary", func(t *testing.T) {
		g, err := NewGrid(10, 10, StandardRule, WrapNone, DeadBoundary)
		require.NoError(t, err)
		g.RandomizePopulation(100)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(2)
		for idx := range g.cells {
			assert.LessOrEqual(t, g.aliveAdjacentCount(idx), uint8(2))
		}

		g.RandomizePopulation(100)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(8)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(0)
		require.Equal(t, 0, g.Population())
	})
	t.Run("wrap none, alive boundary", func(t *testing.T) {
		g, err := NewGrid(10, 10, StandardRule, WrapNone, AliveBoundary)
		require.NoError(t, err)
		g.RandomizePopulation(100)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(5)
		for idx := range g.cells {
			assert.LessOrEqual(t, g.aliveAdjacentCount(idx), uint8(5))
		}

		g.RandomizePopulation(100)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(8)
		require.Equal(t, 100, g.Population())
		g.LimitAliveAdjacents(0)
		require.Equal(t, 0, g.Population())
	})
}

func TestGrid_Step(t *testing.T) {
	g, err := NewGrid(10, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	g.StepCount.Store(1)
	tr := &testRenderer{}
	g.SetRenderer(tr.render)

	changed, changes := g.Step()
	assert.False(t, changed)
	assert.Equal(t, 0, changes)
	assert.Equal(t, uint64(1), g.StepCount.Load())
	assert.Equal(t, 0, tr.called)
	assert.Equal(t, 0, tr.changes)
	assert.Equal(t, 0, tr.alives)
	assert.Equal(t, 0, tr.deads)

	// glider...
	g.SetCell(0, 2, true)
	g.SetCell(1, 0, true)
	g.SetCell(1, 2, true)
	g.SetCell(2, 1, true)
	g.SetCell(2, 2, true)
	assert.Equal(t, uint64(0), g.StepCount.Load())
	changed, changes = g.Step()
	assert.True(t, changed)
	assert.Equal(t, 4, changes)
	assert.Equal(t, uint64(1), g.StepCount.Load())
	assert.Equal(t, 9, tr.called)
	assert.Equal(t, 9, tr.changes)
	assert.Equal(t, 7, tr.alives)
	assert.Equal(t, 2, tr.deads)
}

func TestGrid_StepAhead(t *testing.T) {
	g, err := NewGrid(5, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	g.StepCount.Store(1)
	tr := &testRenderer{}
	g.SetRenderer(tr.render)

	changes := g.StepAhead(10)
	assert.Equal(t, 0, changes)
	assert.Equal(t, uint64(1), g.StepCount.Load())
	assert.Equal(t, 0, tr.called)

	// glider...
	g.SetCell(0, 2, true)
	g.SetCell(1, 0, true)
	g.SetCell(1, 2, true)
	g.SetCell(2, 1, true)
	g.SetCell(2, 2, true)
	assert.Equal(t, uint64(0), g.StepCount.Load())
	tr = &testRenderer{}
	g.SetRenderer(tr.render)
	changes = g.StepAhead(30)
	assert.Equal(t, 4, changes)
	assert.Equal(t, uint64(30), g.StepCount.Load())
	assert.Equal(t, 0, tr.called)
}

func TestGrid_StepWithInstrumentation(t *testing.T) {
	g, err := NewGrid(5, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	g.StepCount.Store(1)
	tr := &testRenderer{}
	g.SetRenderer(tr.render)

	i := &testInstrument{}
	changed, changes := g.StepWithInstrumentation(i)
	assert.False(t, changed)
	assert.Equal(t, 0, changes)
	assert.Equal(t, uint64(1), g.StepCount.Load())
	assert.Equal(t, 0, tr.called)
	assert.Equal(t, 0, tr.changes)
	assert.Equal(t, 0, tr.alives)
	assert.Equal(t, 0, tr.deads)

	// glider...
	g.SetCell(0, 2, true)
	g.SetCell(1, 0, true)
	g.SetCell(1, 2, true)
	g.SetCell(2, 1, true)
	g.SetCell(2, 2, true)
	assert.Equal(t, uint64(0), g.StepCount.Load())
	i = &testInstrument{}
	changed, changes = g.StepWithInstrumentation(i)
	assert.True(t, changed)
	assert.Equal(t, 4, changes)
	assert.Equal(t, uint64(1), g.StepCount.Load())
	assert.Equal(t, 9, tr.called)
	assert.Equal(t, 9, tr.changes)
	assert.Equal(t, 7, tr.alives)
	assert.Equal(t, 2, tr.deads)
	assert.Equal(t, 1, i.called)
	for range 30 {
		g.StepWithInstrumentation(i)
	}
	assert.Equal(t, uint64(31), g.StepCount.Load())
	assert.Equal(t, 129, tr.called)
	assert.Equal(t, 129, tr.changes)
	assert.Equal(t, 31, i.called)

	g.StepWithInstrumentation(nil)
	assert.Equal(t, 133, tr.called)
	assert.Equal(t, 133, tr.changes)
}

func TestGrid_StepAheadWithInstrumentation(t *testing.T) {
	g, err := NewGrid(5, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	g.StepCount.Store(1)
	tr := &testRenderer{}
	g.SetRenderer(tr.render)

	i := &testInstrument{}
	reason, changes := g.StepAheadWithInstrumentation(10, i)
	assert.Equal(t, 0, changes)
	assert.Equal(t, uint64(1), g.StepCount.Load())
	assert.Equal(t, 0, tr.called)
	assert.Equal(t, NoChangesDetected, reason)
	assert.Equal(t, 0, i.called)

	// glider...
	g.SetCell(0, 2, true)
	g.SetCell(1, 0, true)
	g.SetCell(1, 2, true)
	g.SetCell(2, 1, true)
	g.SetCell(2, 2, true)
	assert.Equal(t, uint64(0), g.StepCount.Load())
	tr = &testRenderer{}
	g.SetRenderer(tr.render)
	i = &testInstrument{}
	reason, changes = g.StepAheadWithInstrumentation(30, i)
	assert.Equal(t, 4, changes)
	assert.Equal(t, uint64(30), g.StepCount.Load())
	assert.Equal(t, 0, tr.called)
	assert.Equal(t, StepCompleted, reason)
	assert.Equal(t, 30, i.called)

	i.stop = true
	reason, changes = g.StepAheadWithInstrumentation(10, i)
	assert.Equal(t, 4, changes)
	assert.Equal(t, uint64(31), g.StepCount.Load())
	assert.Equal(t, 0, tr.called)
	assert.Equal(t, InstrumentStopped, reason)
	assert.Equal(t, 31, i.called)

	g.StepAheadWithInstrumentation(1, nil)
}

type testInstrument struct {
	called int
	stop   bool
}

var _ StepInstrumentation = (*testInstrument)(nil)
var _ StepStopInstrumentation = (*testInstrument)(nil)
var _ DualUseInstrumentation = (*testInstrument)(nil)

func (t *testInstrument) Instrument(_ uint64, _ []int) {
	t.called++
}

func (t *testInstrument) InstrumentStop(_ uint64, _ []int) bool {
	t.called++
	return t.stop
}
