package logic

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"slices"
	"testing"
)

func TestRecordInstrument_Step(t *testing.T) {
	g, err := NewGrid(5, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	g.SetCell(0, 2, true)
	g.SetCell(1, 0, true)
	g.SetCell(1, 2, true)
	g.SetCell(2, 1, true)
	g.SetCell(2, 2, true)
	initial := slices.Clone(g.cells)
	_ = initial
	i := NewRecordInstrument(g)
	assert.Equal(t, g.cells, i.initial)
	assert.Equal(t, [][]bool{
		{false, false, true, false, false, false, false, false, false, false},
		{true, false, true, false, false, false, false, false, false, false},
		{false, true, true, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false}}, i.InitialGrid())
	for range 30 {
		g.StepWithInstrumentation(i)
	}
	assert.Equal(t, 30, i.FramesCount())

	i.Undos(29)
	assert.Equal(t, 1, i.FramesCount())
	locs := 0
	for cl := range i.StepChangeLocations() {
		locs++
		assert.Equal(t, [][2]int{{0, 1}, {0, 2}, {1, 0}, {1, 3}}, cl)
		break
	}
	assert.Equal(t, locs, 1)
	i.Undo()
	assert.Equal(t, 0, i.FramesCount())
	assert.Equal(t, initial, g.cells)
	i.Undo()
	assert.Equal(t, 0, i.FramesCount())
	assert.Equal(t, initial, g.cells)
	i.Undos(10)
	assert.Equal(t, 0, i.FramesCount())
	assert.Equal(t, initial, g.cells)
}

func TestRecordInstrument_StepAhead(t *testing.T) {
	g, err := NewGrid(5, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	g.SetCell(0, 2, true)
	g.SetCell(1, 0, true)
	g.SetCell(1, 2, true)
	g.SetCell(2, 1, true)
	g.SetCell(2, 2, true)
	initial := slices.Clone(g.cells)
	_ = initial
	i := NewRecordInstrument(g)
	assert.Equal(t, g.cells, i.initial)
	assert.Equal(t, [][]bool{
		{false, false, true, false, false, false, false, false, false, false},
		{true, false, true, false, false, false, false, false, false, false},
		{false, true, true, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false}}, i.InitialGrid())
	g.StepAheadWithInstrumentation(30, i)
	assert.Equal(t, 30, i.FramesCount())

	i.Undos(29)
	assert.Equal(t, 1, i.FramesCount())
	locs := 0
	for cl := range i.StepChangeLocations() {
		locs++
		assert.Equal(t, [][2]int{{0, 1}, {0, 2}, {1, 0}, {1, 3}}, cl)
		break
	}
	assert.Equal(t, locs, 1)
	i.Undo()
	assert.Equal(t, 0, i.FramesCount())
	assert.Equal(t, initial, g.cells)
	i.Undo()
	assert.Equal(t, 0, i.FramesCount())
	assert.Equal(t, initial, g.cells)
	i.Undos(10)
	assert.Equal(t, 0, i.FramesCount())
	assert.Equal(t, initial, g.cells)
}
