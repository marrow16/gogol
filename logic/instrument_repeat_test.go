package logic

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRepeatInstrument(t *testing.T) {
	t.Run("stable empty - step ahead", func(t *testing.T) {
		g, _ := NewGrid(10, 10, StandardRule, WrapAll, DeadBoundary)
		i := NewRepeatInstrument(g)
		result, _ := g.StepAheadWithInstrumentation(10, i)
		assert.Equal(t, NoChangesDetected, result)
		assert.Equal(t, uint64(0), g.StepCount.Load())
		assert.False(t, i.Found)
	})
	t.Run("stable block - step ahead", func(t *testing.T) {
		g, _ := NewGrid(10, 10, StandardRule, WrapAll, DeadBoundary)
		g.SetCell(0, 0, true)
		g.SetCell(0, 1, true)
		g.SetCell(1, 0, true)
		g.SetCell(1, 1, true)
		i := NewRepeatInstrument(g)
		result, _ := g.StepAheadWithInstrumentation(10, i)
		assert.Equal(t, NoChangesDetected, result)
		assert.Equal(t, uint64(0), g.StepCount.Load())
		assert.False(t, i.Found)
	})
	t.Run("blinker - step ahead", func(t *testing.T) {
		g, _ := NewGrid(10, 10, StandardRule, WrapAll, DeadBoundary)
		g.SetCell(0, 0, true)
		g.SetCell(0, 1, true)
		g.SetCell(0, 2, true)
		i := NewRepeatInstrument(g)
		result, _ := g.StepAheadWithInstrumentation(10, i)
		assert.Equal(t, InstrumentStopped, result)
		assert.Equal(t, uint64(2), g.StepCount.Load())
		assert.True(t, i.Found)
		assert.Equal(t, uint64(0), i.FirstStep)
		assert.Equal(t, uint64(2), i.RepeatStep)
		assert.Equal(t, uint64(2), i.Period)
		// grid reset ignored
		steps := i.Steps
		g.StepCount.Store(0)
		g.StepWithInstrumentation(i)
		assert.Equal(t, steps, i.Steps)
	})
	t.Run("glider - step ahead", func(t *testing.T) {
		g, _ := NewGrid(10, 10, StandardRule, WrapAll, DeadBoundary)
		g.SetCell(0, 2, true)
		g.SetCell(1, 0, true)
		g.SetCell(1, 2, true)
		g.SetCell(2, 1, true)
		g.SetCell(2, 2, true)
		i := NewRepeatInstrument(g)
		result, _ := g.StepAheadWithInstrumentation(100, i)
		assert.Equal(t, InstrumentStopped, result)
		assert.Equal(t, uint64(40), g.StepCount.Load())
		assert.True(t, i.Found)
		assert.Equal(t, uint64(0), i.FirstStep)
		assert.Equal(t, uint64(40), i.RepeatStep)
		assert.Equal(t, uint64(40), i.Period)
	})
}
