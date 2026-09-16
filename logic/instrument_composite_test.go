package logic

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCompositeInstrument_Step(t *testing.T) {
	g, err := NewGrid(5, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)

	g.SetCell(0, 2, true)
	g.SetCell(1, 0, true)
	g.SetCell(1, 2, true)
	g.SetCell(2, 1, true)
	g.SetCell(2, 2, true)
	assert.Equal(t, uint64(0), g.StepCount.Load())
	i := &testInstrument{}
	ci := CompositeInstrument{i, i}
	changed, changes := g.StepWithInstrumentation(ci)
	assert.True(t, changed)
	assert.Equal(t, 4, changes)
	assert.Equal(t, uint64(1), g.StepCount.Load())
	assert.Equal(t, 2, i.called)
}

func TestCompositeInstrument_StepAhead(t *testing.T) {
	g, err := NewGrid(5, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)

	g.SetCell(0, 2, true)
	g.SetCell(1, 0, true)
	g.SetCell(1, 2, true)
	g.SetCell(2, 1, true)
	g.SetCell(2, 2, true)
	i := &testInstrument{}
	ci := CompositeInstrument{i, i}
	g.StepAheadWithInstrumentation(30, ci)
	assert.Equal(t, 60, i.called)

	i.stop = true
	g.StepAheadWithInstrumentation(10, ci)
	assert.Equal(t, 62, i.called)
}
