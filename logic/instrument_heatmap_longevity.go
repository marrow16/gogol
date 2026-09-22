package logic

import "iter"

func NewLongevityHeatMapInstrument(g *Grid) *LongevityHeatMapInstrument {
	l := g.height * g.width
	step := g.StepCount.Load()
	result := &LongevityHeatMapInstrument{
		grid:       g,
		aliveSince: make([]uint64, l),
		longest:    make([]uint64, l),
		step:       step,
	}
	for idx, c := range g.cells {
		if c == aliveCell {
			result.aliveSince[idx] = step
		}
	}
	return result
}

type LongevityHeatMapInstrument struct {
	grid       *Grid
	aliveSince []uint64
	longest    []uint64
	steps      uint64
	step       uint64
}

var _ HeatMap = (*LongevityHeatMapInstrument)(nil)
var _ StepInstrumentation = (*LongevityHeatMapInstrument)(nil)
var _ StepStopInstrumentation = (*LongevityHeatMapInstrument)(nil)
var _ DualUseInstrumentation = (*LongevityHeatMapInstrument)(nil)

func (h *LongevityHeatMapInstrument) Type() HeatMapperType {
	return LongevityHeatMapper
}

func (h *LongevityHeatMapInstrument) InstrumentStop(step uint64, locations []int) bool {
	h.Instrument(step, locations)
	return false
}

func (h *LongevityHeatMapInstrument) Instrument(step uint64, locations []int) {
	if step > h.step {
		h.step = step
		h.steps++
		for _, idx := range locations {
			if h.grid.cells[idx] == aliveCell {
				// became alive on this step...
				h.aliveSince[idx] = step
			} else {
				// became dead on this step - record the completed lifespan...
				age := step - h.aliveSince[idx]
				h.longest[idx] = max(h.longest[idx], age)
			}
		}
	}
}

func (h *LongevityHeatMapInstrument) HeatMap() iter.Seq[HeatLocation] {
	return func(yield func(HeatLocation) bool) {
		var hmax uint64
		for idx, longest := range h.longest {
			if h.grid.cells[idx] == aliveCell {
				longest = max(longest, h.step-h.aliveSince[idx]+1)
			}
			hmax = max(hmax, longest)
		}
		width := h.grid.width
		for idx, longest := range h.longest {
			if h.grid.cells[idx] == aliveCell {
				longest = max(longest, h.step-h.aliveSince[idx]+1)
			}
			value := 0.0
			if hmax > 0 {
				value = float64(longest) / float64(hmax)
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

func (h *LongevityHeatMapInstrument) StepsCount() uint64 {
	return h.steps
}
