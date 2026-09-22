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
	steps  uint64
	step   uint64
}

var _ HeatMap = (*ActivityHeatMapInstrument)(nil)
var _ StepInstrumentation = (*ActivityHeatMapInstrument)(nil)
var _ StepStopInstrumentation = (*ActivityHeatMapInstrument)(nil)
var _ DualUseInstrumentation = (*ActivityHeatMapInstrument)(nil)

func (h *ActivityHeatMapInstrument) Type() HeatMapperType {
	return ActivityHeatMapper
}

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
		}
	}
}

func (h *ActivityHeatMapInstrument) HeatMap() iter.Seq[HeatLocation] {
	return func(yield func(HeatLocation) bool) {
		hmax := float64(0)
		for _, count := range h.counts {
			hmax = max(hmax, float64(count))
		}
		width := h.grid.width
		for i, v := range h.counts {
			value := 0.0
			if hmax > 0 {
				value = float64(v) / hmax
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

func (h *ActivityHeatMapInstrument) StepsCount() uint64 {
	return h.steps
}
