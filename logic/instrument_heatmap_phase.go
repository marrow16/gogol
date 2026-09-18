package logic

import "iter"

func NewPhaseHeatMapInstrument(g *Grid) *PhaseHeatMapInstrument {
	step := g.StepCount.Load()
	l := g.height * g.width
	result := &PhaseHeatMapInstrument{
		grid:      g,
		counts:    make([]uint64, l),
		since:     make([]uint64, l),
		step:      step,
		startStep: step,
	}
	for idx := range l {
		result.since[idx] = step + 1
	}
	return result
}

type PhaseHeatMapInstrument struct {
	grid      *Grid
	counts    []uint64
	since     []uint64
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

func (h *PhaseHeatMapInstrument) Instrument(step uint64, locations []int) {
	if step > h.step {
		h.step = step
		h.steps++
		for _, idx := range locations {
			// grid has already changed, so the opposite of the current state was in effect up to step-1...
			previous := h.grid.cells[idx] ^ 1
			h.counts[idx] += h.phaseCount(previous, h.since[idx], step-1)
			// current state starts at this step...
			h.since[idx] = step
		}
	}
}

func (h *PhaseHeatMapInstrument) HeatMap() iter.Seq[HeatLocation] {
	return func(yield func(HeatLocation) bool) {
		var hmax uint64
		for idx := range h.counts {
			hmax = max(hmax, h.count(idx))
		}
		width := h.grid.width
		for idx := range h.counts {
			count := h.count(idx)
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

func (h *PhaseHeatMapInstrument) StepsCount() uint64 {
	return h.steps
}

func (h *PhaseHeatMapInstrument) count(idx int) uint64 {
	return h.counts[idx] + h.phaseCount(h.grid.cells[idx], h.since[idx], h.step)
}

func (h *PhaseHeatMapInstrument) phaseCount(alive uint8, from, to uint64) uint64 {
	if from > to {
		return 0
	}
	n := to - from + 1
	count := n / 2
	expected := uint8((h.startStep - from) & 1)
	if n&1 != 0 && alive != expected {
		count++
	}
	return count
}
