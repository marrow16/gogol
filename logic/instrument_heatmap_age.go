package logic

import "iter"

func NewAgeHeatMapInstrument(g *Grid) *AgeHeatMapInstrument {
	l := g.height * g.width
	step := g.StepCount.Load()
	result := &AgeHeatMapInstrument{
		grid:       g,
		aliveSince: make([]uint64, l),
		step:       step,
	}
	for idx, c := range g.cells {
		if c == aliveCell {
			result.aliveSince[idx] = step
		}
	}
	return result
}

type AgeHeatMapInstrument struct {
	grid       *Grid
	aliveSince []uint64
	steps      uint64
	step       uint64
}

var _ HeatMap = (*AgeHeatMapInstrument)(nil)
var _ StepInstrumentation = (*AgeHeatMapInstrument)(nil)
var _ StepStopInstrumentation = (*AgeHeatMapInstrument)(nil)
var _ DualUseInstrumentation = (*AgeHeatMapInstrument)(nil)

func (h *AgeHeatMapInstrument) Type() HeatMapperType {
	return AgeHeatMapper
}

func (h *AgeHeatMapInstrument) InstrumentStop(step uint64, locations []int) bool {
	h.Instrument(step, locations)
	return false
}

func (h *AgeHeatMapInstrument) Instrument(step uint64, locations []int) {
	if step > h.step {
		h.step = step
		h.steps++
		for _, idx := range locations {
			if h.grid.cells[idx] == aliveCell {
				// became alive on this step...
				h.aliveSince[idx] = step
			}
		}
	}
}

func (h *AgeHeatMapInstrument) HeatMap() iter.Seq[HeatLocation] {
	return func(yield func(HeatLocation) bool) {
		var hmax uint64
		for idx, since := range h.aliveSince {
			if h.grid.cells[idx] == aliveCell {
				hmax = max(hmax, h.step-since+1)
			}
		}
		width := h.grid.width
		for idx, since := range h.aliveSince {
			var age uint64
			if h.grid.cells[idx] == aliveCell {
				age = h.step - since + 1
			}
			value := 0.0
			if hmax > 0 {
				value = float64(age) / float64(hmax)
			}
			if !yield(HeatLocation{
				Row:   idx / width,
				Col:   idx % width,
				Value: value,
			}) {
				return
			}
		}
	}
}

func (h *AgeHeatMapInstrument) StepsCount() uint64 {
	return h.steps
}
