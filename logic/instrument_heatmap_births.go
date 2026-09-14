package logic

import "iter"

func NewBirthsHeatMapInstrument(g *Grid, decay float32) *BirthsHeatMapInstrument {
	if decay < 0 {
		decay = 0
	}
	if decay > 1 {
		decay = 1
	}
	return &BirthsHeatMapInstrument{
		grid:   g,
		values: make([]float32, g.height*g.width),
		step:   g.StepCount.Load(),
		steps:  0,
		decay:  decay,
	}
}

type BirthsHeatMapInstrument struct {
	grid   *Grid
	values []float32
	step   uint64
	steps  uint64
	decay  float32
}

var _ HeatMap = (*BirthsHeatMapInstrument)(nil)
var _ StepInstrumentation = (*BirthsHeatMapInstrument)(nil)
var _ StepStopInstrumentation = (*BirthsHeatMapInstrument)(nil)
var _ DualUseInstrumentation = (*BirthsHeatMapInstrument)(nil)

func (h *BirthsHeatMapInstrument) InstrumentStop(step uint64, locations []int) bool {
	h.Instrument(step, locations)
	return false
}

func (h *BirthsHeatMapInstrument) Instrument(step uint64, locations []int) {
	if step > h.step {
		h.step = step
		h.steps++
		// fade existing heat
		for i := range h.values {
			h.values[i] *= h.decay
		}
		// only births become "hot"
		for _, idx := range locations {
			if h.grid.cells[idx] == aliveCell {
				h.values[idx] = 1.0
			}
		}
	}
}

func (h *BirthsHeatMapInstrument) HeatMap() iter.Seq[HeatLocation] {
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

func (h *BirthsHeatMapInstrument) Maximum() uint64 {
	return 1
}

func (h *BirthsHeatMapInstrument) StepsCount() uint64 {
	return h.steps
}
