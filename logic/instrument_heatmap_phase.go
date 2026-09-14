package logic

import "iter"

func NewPhaseHeatMapInstrument(g *Grid) *PhaseHeatMapInstrument {
	return &PhaseHeatMapInstrument{
		grid:      g,
		counts:    make([]uint64, g.height*g.width),
		step:      g.StepCount.Load(),
		startStep: g.StepCount.Load(),
	}
}

type PhaseHeatMapInstrument struct {
	grid      *Grid
	counts    []uint64
	max       uint64
	steps     uint64
	step      uint64
	startStep uint64
}

var _ HeatMap = (*PhaseHeatMapInstrument)(nil)
var _ StepInstrumentation = (*PhaseHeatMapInstrument)(nil)
var _ StepStopInstrumentation = (*PhaseHeatMapInstrument)(nil)
var _ DualUseInstrumentation = (*PhaseHeatMapInstrument)(nil)

func (h *PhaseHeatMapInstrument) InstrumentStop(step uint64, locations []int) bool {
	h.Instrument(step, locations)
	return false
}

func (h *PhaseHeatMapInstrument) Instrument(step uint64, _ []int) {
	if step > h.step {
		h.step = step
		h.steps++
		expected := uint8((h.startStep - step) & 1)
		for idx, c := range h.grid.cells {
			if c != expected {
				h.counts[idx]++
				if m := h.counts[idx]; m > h.max {
					h.max = m
				}
			}
		}
	}
}

func (h *PhaseHeatMapInstrument) HeatMap() iter.Seq[HeatLocation] {
	return func(yield func(HeatLocation) bool) {
		width := h.grid.width
		for i, v := range h.counts {
			value := 0.0
			if h.max > 0 {
				value = float64(v) / float64(h.max)
			}
			if !yield(HeatLocation{
				Row:   i / width,
				Col:   i % width,
				Value: value,
			}) {
				return
			}
		}
	}
}

func (h *PhaseHeatMapInstrument) Maximum() uint64 {
	return h.max
}

func (h *PhaseHeatMapInstrument) StepsCount() uint64 {
	return h.steps
}
