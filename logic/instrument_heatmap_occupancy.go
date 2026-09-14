package logic

import "iter"

func NewOccupancyHeatMapInstrument(g *Grid) *OccupancyHeatMapInstrument {
	result := &OccupancyHeatMapInstrument{
		grid:   g,
		counts: make([]uint64, g.height*g.width),
		step:   g.StepCount.Load(),
	}
	result.initialise()
	return result
}

type OccupancyHeatMapInstrument struct {
	grid   *Grid
	counts []uint64
	max    uint64
	steps  uint64
	step   uint64
}

var _ HeatMap = (*OccupancyHeatMapInstrument)(nil)
var _ StepInstrumentation = (*OccupancyHeatMapInstrument)(nil)
var _ StepStopInstrumentation = (*OccupancyHeatMapInstrument)(nil)
var _ DualUseInstrumentation = (*OccupancyHeatMapInstrument)(nil)

func (h *OccupancyHeatMapInstrument) initialise() {
	for idx, c := range h.grid.cells {
		if c == aliveCell {
			h.counts[idx] = 1
			h.max = 1
		}
	}
}

func (h *OccupancyHeatMapInstrument) InstrumentStop(step uint64, _ []int) bool {
	h.Instrument(step, nil)
	return false
}

func (h *OccupancyHeatMapInstrument) Instrument(step uint64, _ []int) {
	if step > h.step {
		h.step = step
		h.steps++
		for idx, c := range h.grid.cells {
			if c == aliveCell {
				h.counts[idx]++
				if m := h.counts[idx]; m > h.max {
					h.max = m
				}
			}
		}
	}
}

func (h *OccupancyHeatMapInstrument) HeatMap() iter.Seq[HeatLocation] {
	return func(yield func(HeatLocation) bool) {
		for i, v := range h.counts {
			value := 0.0
			if h.max > 0 {
				value = float64(v) / float64(h.max)
			}
			if !yield(HeatLocation{
				Row:   i / h.grid.width,
				Col:   i % h.grid.width,
				Value: value,
			}) {
				return
			}
		}
	}
}

func (h *OccupancyHeatMapInstrument) Maximum() uint64 {
	return h.max
}

func (h *OccupancyHeatMapInstrument) StepsCount() uint64 {
	return h.steps
}
