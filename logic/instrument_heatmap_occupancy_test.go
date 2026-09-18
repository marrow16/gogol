package logic

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOccupancyHeatMapInstrument_Step(t *testing.T) {
	g, err := NewGrid(5, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	g.SetCell(0, 2, true)
	g.SetCell(1, 0, true)
	g.SetCell(1, 2, true)
	g.SetCell(2, 1, true)
	g.SetCell(2, 2, true)
	i := NewOccupancyHeatMapInstrument(g)
	for range 100 {
		g.StepWithInstrumentation(i)
	}
	assert.Equal(t, uint64(100), i.StepsCount())
	count := 0
	for hl := range i.HeatMap() {
		count++
		assert.True(t, hl.Value >= 0 && hl.Value <= 1, hl.Value)
	}
	assert.Equal(t, 50, count)
}

func TestOccupancyHeatMapInstrument_StepAhead(t *testing.T) {
	g, err := NewGrid(5, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	g.SetCell(0, 2, true)
	g.SetCell(1, 0, true)
	g.SetCell(1, 2, true)
	g.SetCell(2, 1, true)
	g.SetCell(2, 2, true)
	i := NewOccupancyHeatMapInstrument(g)
	g.StepAheadWithInstrumentation(100, i)
	assert.Equal(t, uint64(100), i.StepsCount())
	count := 0
	for hl := range i.HeatMap() {
		count++
		assert.True(t, hl.Value >= 0 && hl.Value <= 1, hl.Value)
	}
	assert.Equal(t, 50, count)
	// coverage on break...
	count = 0
	for range i.HeatMap() {
		count++
		break
	}
	assert.Equal(t, 1, count)
}
