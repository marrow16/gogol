package logic

import "iter"

func NewActivityHeatMapInstrument(g *Grid) *ActivityHeatMapInstrument {
	return &ActivityHeatMapInstrument{
		grid:   g,
		counts: make([]uint64, g.height*g.width),
		step:   g.StepCount.Load(),
	}
}

type ActivityHeatMapInstrument struct {
	grid   *Grid
	counts []uint64
	max    uint64
	steps  uint64
	step   uint64
}

var _ HeatMap = (*ActivityHeatMapInstrument)(nil)
var _ StepInstrumentation = (*ActivityHeatMapInstrument)(nil)
var _ StepStopInstrumentation = (*ActivityHeatMapInstrument)(nil)
var _ DualUseInstrumentation = (*ActivityHeatMapInstrument)(nil)

func (h *ActivityHeatMapInstrument) InstrumentStop(step uint64, locations []int) bool {
	h.Instrument(step, locations)
	return false
}

func (h *ActivityHeatMapInstrument) Instrument(step uint64, locations []int) {
	if step > h.step {
		h.step = step
		h.steps++
		for _, idx := range locations {
			h.counts[idx]++
			if m := h.counts[idx]; m > h.max {
				h.max = m
			}
		}
	}
}

func (h *ActivityHeatMapInstrument) Activity() iter.Seq[ActivityLocation] {
	return func(yield func(ActivityLocation) bool) {
		for i, v := range h.counts {
			if !yield(ActivityLocation{
				Row:   i / h.grid.width,
				Col:   i % h.grid.width,
				Value: v,
			}) {
				return
			}
		}
	}
}

func (h *ActivityHeatMapInstrument) HeatMap() iter.Seq[HeatLocation] {
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

func (h *ActivityHeatMapInstrument) Maximum() uint64 {
	return h.max
}

func (h *ActivityHeatMapInstrument) StepsCount() uint64 {
	return h.steps
}
