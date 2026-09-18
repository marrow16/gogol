package logic

import "iter"

func NewOccupancyHeatMapInstrument(g *Grid) *OccupancyHeatMapInstrument {
	l := g.height * g.width
	step := g.StepCount.Load()
	result := &OccupancyHeatMapInstrument{
		grid:       g,
		counts:     make([]uint64, l),
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

type OccupancyHeatMapInstrument struct {
	grid       *Grid
	counts     []uint64
	aliveSince []uint64
	steps      uint64
	step       uint64
}

var _ HeatMap = (*OccupancyHeatMapInstrument)(nil)
var _ StepInstrumentation = (*OccupancyHeatMapInstrument)(nil)
var _ StepStopInstrumentation = (*OccupancyHeatMapInstrument)(nil)
var _ DualUseInstrumentation = (*OccupancyHeatMapInstrument)(nil)

func (h *OccupancyHeatMapInstrument) InstrumentStop(step uint64, locations []int) bool {
	h.Instrument(step, locations)
	return false
}

func (h *OccupancyHeatMapInstrument) Instrument(step uint64, locations []int) {
	if step > h.step {
		h.step = step
		h.steps++
		for _, idx := range locations {
			if h.grid.cells[idx] == aliveCell {
				// became alive on this step...
				h.aliveSince[idx] = step
			} else {
				// became dead on this step...
				h.counts[idx] += step - h.aliveSince[idx]
			}
		}
	}
}

func (h *OccupancyHeatMapInstrument) HeatMap() iter.Seq[HeatLocation] {
	return func(yield func(HeatLocation) bool) {
		var hmax uint64
		for idx, count := range h.counts {
			if h.grid.cells[idx] == aliveCell {
				count += h.step - h.aliveSince[idx] + 1
			}
			hmax = max(hmax, count)
		}
		width := h.grid.width
		for idx, count := range h.counts {
			if h.grid.cells[idx] == aliveCell {
				count += h.step - h.aliveSince[idx] + 1
			}
			value := 0.0
			if hmax > 0 {
				value = float64(count) / float64(hmax)
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

func (h *OccupancyHeatMapInstrument) StepsCount() uint64 {
	return h.steps
}
