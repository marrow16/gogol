package logic

import (
	"iter"
	"strings"
)

type HeatLocation struct {
	Row   int
	Col   int
	Value float64
}

type HeatMap interface {
	HeatMap() iter.Seq[HeatLocation]
	StepsCount() uint64
	Type() HeatMapperType
}

type HeatMapperType int

const (
	NoHeatMapper HeatMapperType = iota
	AllHeatMapper
	ActivityHeatMapper
	AgeHeatMapper
	BirthsHeatMapper
	FreshnessHeatMapper
	LongevityHeatMapper
	OccupancyHeatMapper
	PhaseParityHeatMapper
)

func (hmt HeatMapperType) String() string {
	switch hmt {
	case AllHeatMapper:
		return "All"
	case ActivityHeatMapper:
		return "Activity"
	case AgeHeatMapper:
		return "Age"
	case BirthsHeatMapper:
		return "Births"
	case FreshnessHeatMapper:
		return "Freshness"
	case LongevityHeatMapper:
		return "Longevity"
	case OccupancyHeatMapper:
		return "Occupancy"
	case PhaseParityHeatMapper:
		return "Phase Parity"
	default:
		return "None"
	}
}

func (hmt HeatMapperType) Token() string {
	switch hmt {
	case NoHeatMapper:
		return "X"
	case AllHeatMapper:
		return "AGBFLOP"
	case ActivityHeatMapper:
		return "A"
	case AgeHeatMapper:
		return "G"
	case BirthsHeatMapper:
		return "B"
	case FreshnessHeatMapper:
		return "F"
	case LongevityHeatMapper:
		return "L"
	case OccupancyHeatMapper:
		return "O"
	case PhaseParityHeatMapper:
		return "P"
	default:
		return "U"
	}
}

func (hmt HeatMapperType) New(g *Grid, halfLife float32) HeatMap {
	switch hmt {
	case ActivityHeatMapper:
		return NewActivityHeatMapInstrument(g)
	case AgeHeatMapper:
		return NewAgeHeatMapInstrument(g)
	case BirthsHeatMapper:
		return NewBirthsHeatMapInstrument(g, halfLife)
	case FreshnessHeatMapper:
		return NewFreshnessHeatMapInstrument(g, halfLife)
	case LongevityHeatMapper:
		return NewLongevityHeatMapInstrument(g)
	case OccupancyHeatMapper:
		return NewOccupancyHeatMapInstrument(g)
	case PhaseParityHeatMapper:
		return NewPhaseHeatMapInstrument(g)
	case AllHeatMapper:
		return NewAllHeatMapInstrument(g, halfLife)
	default:
		return nil
	}
}

func HeatMapperTypeFrom(s string) HeatMapperType {
	switch strings.ToLower(s) {
	case "activity":
		return ActivityHeatMapper
	case "age":
		return AgeHeatMapper
	case "births":
		return BirthsHeatMapper
	case "freshness":
		return FreshnessHeatMapper
	case "longevity":
		return LongevityHeatMapper
	case "occupancy":
		return OccupancyHeatMapper
	case "phase parity":
		return PhaseParityHeatMapper
	case "all":
		return AllHeatMapper
	default:
		return NoHeatMapper
	}
}

func NewAllHeatMapInstrument(g *Grid, decay float32) *AllHeatMapInstrument {
	l := g.height * g.width
	step := g.StepCount.Load()
	result := &AllHeatMapInstrument{
		grid:            g,
		step:            step,
		startStep:       step,
		decay:           max(0, min(decay, 1)),
		aliveSince:      make([]uint64, l),
		activityCounts:  make([]uint64, l),
		birthValues:     make([]float32, l),
		freshnessValues: make([]float32, l),
		occupancyCounts: make([]uint64, l),
		phaseCounts:     make([]uint64, l),
		phaseSince:      make([]uint64, l),
		longest:         make([]uint64, l),
	}
	for idx, c := range g.cells {
		if c == aliveCell {
			result.aliveSince[idx] = step
		}
		result.phaseSince[idx] = step + 1
	}
	return result
}

type AllHeatMapInstrument struct {
	grid      *Grid
	steps     uint64
	step      uint64
	startStep uint64
	decay     float32

	aliveSince      []uint64
	activityCounts  []uint64
	birthValues     []float32
	freshnessValues []float32
	occupancyCounts []uint64
	phaseCounts     []uint64
	phaseSince      []uint64
	longest         []uint64
}

var _ HeatMap = (*AllHeatMapInstrument)(nil)
var _ StepInstrumentation = (*AllHeatMapInstrument)(nil)
var _ StepStopInstrumentation = (*AllHeatMapInstrument)(nil)
var _ DualUseInstrumentation = (*AllHeatMapInstrument)(nil)

func (h *AllHeatMapInstrument) Type() HeatMapperType {
	return AllHeatMapper
}

func (h *AllHeatMapInstrument) InstrumentStop(step uint64, locations []int) bool {
	h.Instrument(step, locations)
	return false
}

func (h *AllHeatMapInstrument) Instrument(step uint64, locations []int) {
	if step > h.step {
		h.step = step
		h.steps++
		// pre changes - births...
		for i := range h.birthValues {
			h.birthValues[i] *= h.decay
		}
		// pre changes - freshness - fade existing heat...
		for i := range h.freshnessValues {
			h.freshnessValues[i] *= h.decay
		}
		// actual changes...
		for _, idx := range locations {
			// activity...
			h.activityCounts[idx]++
			if h.grid.cells[idx] == aliveCell {
				// became alive on this step...
				h.aliveSince[idx] = step
				// births...
				h.birthValues[idx] = 1.0
			} else {
				// became dead on this step...
				// occupancy -
				h.occupancyCounts[idx] += step - h.aliveSince[idx]
				// longevity...
				age := step - h.aliveSince[idx]
				h.longest[idx] = max(h.longest[idx], age)
			}
			// freshness...
			h.freshnessValues[idx] = 1.0
			// phase parity...
			// grid has already changed, so the opposite of the current state was in effect up to step-1...
			previous := h.grid.cells[idx] ^ 1
			h.phaseCounts[idx] += phaseCount(previous, h.startStep, h.phaseSince[idx], step-1)
			// current state starts at this step...
			h.phaseSince[idx] = step
		}
	}
}

// HeatMap by default, returns activity heat mapping info
func (h *AllHeatMapInstrument) HeatMap() iter.Seq[HeatLocation] {
	return h.Specific(ActivityHeatMapper).HeatMap()
}

func (h *AllHeatMapInstrument) StepsCount() uint64 {
	return h.steps
}

func (h *AllHeatMapInstrument) Specific(hmt HeatMapperType) HeatMap {
	switch hmt {
	case ActivityHeatMapper:
		return &ActivityHeatMapInstrument{
			grid:   h.grid,
			counts: h.activityCounts,
			steps:  h.steps,
			step:   h.step,
		}
	case AgeHeatMapper:
		return &AgeHeatMapInstrument{
			grid:       h.grid,
			aliveSince: h.aliveSince,
			steps:      h.steps,
			step:       h.step,
		}
	case BirthsHeatMapper:
		return &BirthsHeatMapInstrument{
			grid:   h.grid,
			values: h.birthValues,
			step:   h.step,
			steps:  h.steps,
			decay:  h.decay,
		}
	case FreshnessHeatMapper:
		return &FreshnessHeatMapInstrument{
			grid:   h.grid,
			values: h.freshnessValues,
			step:   h.step,
			steps:  h.steps,
			decay:  h.decay,
		}
	case LongevityHeatMapper:
		return &LongevityHeatMapInstrument{
			grid:       h.grid,
			aliveSince: h.aliveSince,
			longest:    h.longest,
			steps:      h.steps,
			step:       h.step,
		}
	case OccupancyHeatMapper:
		return &OccupancyHeatMapInstrument{
			grid:       h.grid,
			counts:     h.occupancyCounts,
			aliveSince: h.aliveSince,
			steps:      h.steps,
			step:       h.step,
		}
	case PhaseParityHeatMapper:
		return &PhaseHeatMapInstrument{
			grid:      h.grid,
			counts:    h.phaseCounts,
			since:     h.phaseSince,
			steps:     h.steps,
			step:      h.step,
			startStep: h.startStep,
		}
	}
	return nil
}

func (h *AllHeatMapInstrument) List() iter.Seq[HeatMap] {
	return func(yield func(HeatMap) bool) {
		for _, hmt := range []HeatMapperType{ActivityHeatMapper, AgeHeatMapper, BirthsHeatMapper, FreshnessHeatMapper, LongevityHeatMapper, OccupancyHeatMapper, PhaseParityHeatMapper} {
			if !yield(h.Specific(hmt)) {
				return
			}
		}
	}
}
