package logic

import "iter"

func NewFreshnessHeatMapInstrument(g *Grid, decay float32) *FreshnessHeatMapInstrument {
	if decay < 0 {
		decay = 0
	}
	if decay > 1 {
		decay = 1
	}
	return &FreshnessHeatMapInstrument{
		grid:   g,
		values: make([]float32, g.height*g.width),
		step:   g.StepCount.Load(),
		steps:  0,
		decay:  decay,
	}
}

type FreshnessHeatMapInstrument struct {
	grid   *Grid
	values []float32
	step   uint64
	steps  uint64
	decay  float32
}

var _ HeatMap = (*FreshnessHeatMapInstrument)(nil)
var _ StepInstrumentation = (*FreshnessHeatMapInstrument)(nil)
var _ StepStopInstrumentation = (*FreshnessHeatMapInstrument)(nil)
var _ DualUseInstrumentation = (*FreshnessHeatMapInstrument)(nil)

func (h *FreshnessHeatMapInstrument) InstrumentStop(step uint64, locations []int) bool {
	h.Instrument(step, locations)
	return false
}

func (h *FreshnessHeatMapInstrument) Instrument(step uint64, locations []int) {
	if step > h.step {
		h.step = step
		h.steps++
		// fade existing heat...
		for i := range h.values {
			h.values[i] *= h.decay
		}
		// changed cells become "hot"...
		for _, idx := range locations {
			h.values[idx] = 1.0
		}
	}
}

func (h *FreshnessHeatMapInstrument) HeatMap() iter.Seq[HeatLocation] {
	return func(yield func(HeatLocation) bool) {
		width := h.grid.width
		for i, v := range h.values {
			if !yield(HeatLocation{
				Row:   i / width,
				Col:   i % width,
				Value: float64(v),
			}) {
				return
			}
		}
	}
}

func (h *FreshnessHeatMapInstrument) Maximum() uint64 {
	return 1
}

func (h *FreshnessHeatMapInstrument) StepsCount() uint64 {
	return h.steps
}
