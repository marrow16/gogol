package logic

import (
	"iter"
	"slices"
)

type RepeatInstrument struct {
	hash       uint64
	zobrist    [][]uint64
	seen       map[uint64]uint64
	step       uint64
	FirstStep  uint64
	RepeatStep uint64
	Period     uint64
	Found      bool
}

var _ StepInstrumentation = (*RepeatInstrument)(nil)
var _ StepStopInstrumentation = (*RepeatInstrument)(nil)
var _ DualUseInstrumentation = (*RepeatInstrument)(nil)

func NewRepeatInstrument(g *Grid) *RepeatInstrument {
	zobrist := make([][]uint64, g.Height)
	for r := range zobrist {
		zobrist[r] = make([]uint64, g.Width)
		for c := range zobrist[r] {
			zobrist[r][c] = rng.Uint64()
		}
	}
	hash := uint64(0)
	for r, row := range g.Rows {
		for c, cell := range row {
			if cell.Alive {
				hash ^= zobrist[r][c]
			}
		}
	}
	step := g.StepCount.Load()
	return &RepeatInstrument{
		hash:    hash,
		zobrist: zobrist,
		seen: map[uint64]uint64{
			hash: step,
		},
		step: step,
	}
}

func (r *RepeatInstrument) InstrumentStop(step uint64, _ []*Cell, locations [][2]int) bool {
	if step > r.step {
		r.step = step
		for _, loc := range locations {
			r.hash ^= r.zobrist[loc[0]][loc[1]]
		}
		if first, ok := r.seen[r.hash]; ok {
			if !r.Found {
				r.FirstStep = first
				r.RepeatStep = step
				r.Period = step - first
			}
			r.Found = true
			return true
		}
		r.seen[r.hash] = step
	}
	return false
}

func (r *RepeatInstrument) Instrument(step uint64, _ []*Cell, locations [][2]int) {
	r.InstrumentStop(step, nil, locations)
}

type frame struct {
	step      uint64
	locations [][2]int
}

func NewRecordInstrument(g *Grid) *RecordInstrument {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return &RecordInstrument{
		grid:    g,
		frames:  make([]frame, 0),
		initial: initialGridState(g),
	}
}

func initialGridState(g *Grid) [][]bool {
	result := make([][]bool, g.Height)
	for r, row := range g.Rows {
		result[r] = make([]bool, g.Width)
		for c, cell := range row {
			result[r][c] = cell.Alive
		}
	}
	return result
}

type RecordInstrument struct {
	grid    *Grid
	frames  []frame
	initial [][]bool
}

var _ StepInstrumentation = (*RecordInstrument)(nil)
var _ StepStopInstrumentation = (*RecordInstrument)(nil)
var _ DualUseInstrumentation = (*RecordInstrument)(nil)

func (r *RecordInstrument) StepsCount() int {
	return len(r.frames)
}

func (r *RecordInstrument) InitialGrid() [][]bool {
	result := make([][]bool, len(r.initial))
	for rn, row := range r.initial {
		result[rn] = slices.Clone(row)
	}
	return result
}

func (r *RecordInstrument) StepChangeLocations() iter.Seq[[][2]int] {
	return func(yield func([][2]int) bool) {
		for _, f := range r.frames {
			if !yield(f.locations) {
				return
			}
		}
	}
}

func (r *RecordInstrument) InstrumentStop(step uint64, _ []*Cell, locations [][2]int) bool {
	r.Instrument(step, nil, locations)
	return false
}

func (r *RecordInstrument) Instrument(step uint64, _ []*Cell, locations [][2]int) {
	r.frames = append(r.frames, frame{
		step:      step,
		locations: append([][2]int(nil), locations...),
	})
}

func (r *RecordInstrument) Undo() bool {
	r.grid.mutex.Lock()
	defer r.grid.mutex.Unlock()
	if len(r.frames) == 0 {
		return false
	}
	f := r.frames[len(r.frames)-1]
	r.frames = r.frames[:len(r.frames)-1]
	for _, loc := range f.locations {
		r.grid.Rows[loc[0]][loc[1]].flip()
	}
	r.grid.StepCount.Store(f.step - 1)
	return true
}

func (r *RecordInstrument) Undos(n int) int {
	r.grid.mutex.Lock()
	defer r.grid.mutex.Unlock()
	undone := 0
	for undone < n && len(r.frames) > 0 {
		f := r.frames[len(r.frames)-1]
		r.frames = r.frames[:len(r.frames)-1]
		for _, loc := range f.locations {
			r.grid.Rows[loc[0]][loc[1]].flip()
		}
		r.grid.StepCount.Store(f.step - 1)
		undone++
	}
	return undone
}

func (r *RecordInstrument) UndoTo(step uint64) bool {
	r.grid.mutex.Lock()
	defer r.grid.mutex.Unlock()
	changed := false
	for len(r.frames) > 0 {
		f := r.frames[len(r.frames)-1]
		if f.step <= step {
			break
		}
		r.frames = r.frames[:len(r.frames)-1]
		for _, loc := range f.locations {
			r.grid.Rows[loc[0]][loc[1]].flip()
		}
		changed = true
	}
	if changed {
		r.grid.StepCount.Store(step)
	}
	return changed
}

type HeatLocation struct {
	Row   int
	Col   int
	Value float64
}

type ActivityLocation struct {
	Row   int
	Col   int
	Value uint64
}

type HeatMap interface {
	HeatMap() iter.Seq[HeatLocation]
	Maximum() uint64
	StepsCount() uint64
}

func NewActivityHeatMapInstrument(g *Grid) *ActivityHeatMapInstrument {
	return &ActivityHeatMapInstrument{
		grid:   g,
		counts: make([]uint64, g.Height*g.Width),
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

func (h *ActivityHeatMapInstrument) InstrumentStop(step uint64, _ []*Cell, locations [][2]int) bool {
	h.Instrument(step, nil, locations)
	return false
}

func (h *ActivityHeatMapInstrument) Instrument(step uint64, _ []*Cell, locations [][2]int) {
	if step > h.step {
		h.step = step
		h.steps++
		for _, loc := range locations {
			i := loc[0]*h.grid.Width + loc[1]
			h.counts[i]++
			if h.counts[i] > h.max {
				h.max = h.counts[i]
			}
		}
	}
}

func (h *ActivityHeatMapInstrument) Activity() iter.Seq[ActivityLocation] {
	return func(yield func(ActivityLocation) bool) {
		for i, v := range h.counts {
			if !yield(ActivityLocation{
				Row:   i / h.grid.Width,
				Col:   i % h.grid.Width,
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
				Row:   i / h.grid.Width,
				Col:   i % h.grid.Width,
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

func NewOccupancyHeatMapInstrument(g *Grid) *OccupancyHeatMapInstrument {
	result := &OccupancyHeatMapInstrument{
		grid:   g,
		counts: make([]uint64, g.Height*g.Width),
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
	wd := h.grid.Width
	for r, row := range h.grid.Rows {
		for c, cell := range row {
			if cell.Alive {
				h.max = 1
				h.counts[(r*wd)+c] = 1
			}
		}
	}
}

func (h *OccupancyHeatMapInstrument) InstrumentStop(step uint64, _ []*Cell, _ [][2]int) bool {
	h.Instrument(step, nil, nil)
	return false
}

func (h *OccupancyHeatMapInstrument) Instrument(step uint64, _ []*Cell, _ [][2]int) {
	if step > h.step {
		h.step = step
		h.steps++
		wd := h.grid.Width
		for r, row := range h.grid.Rows {
			for c, cell := range row {
				if cell.Alive {
					i := (r * wd) + c
					h.counts[i]++
					if h.counts[i] > h.max {
						h.max = h.counts[i]
					}
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
				Row:   i / h.grid.Width,
				Col:   i % h.grid.Width,
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

func NewFreshnessHeatMapInstrument(g *Grid, decay float32) *FreshnessHeatMapInstrument {
	if decay < 0 {
		decay = 0
	}
	if decay > 1 {
		decay = 1
	}
	return &FreshnessHeatMapInstrument{
		grid:   g,
		values: make([]float32, g.Height*g.Width),
		step:   g.StepCount.Load(),
		steps:  0,
		decay:  decay,
	}
}

type FreshnessHeatMapInstrument struct {
	grid   *Grid
	values []float32
	step   uint64
	steps  uint64
	decay  float32
}

var _ HeatMap = (*FreshnessHeatMapInstrument)(nil)
var _ StepInstrumentation = (*FreshnessHeatMapInstrument)(nil)
var _ StepStopInstrumentation = (*FreshnessHeatMapInstrument)(nil)
var _ DualUseInstrumentation = (*FreshnessHeatMapInstrument)(nil)

func (h *FreshnessHeatMapInstrument) InstrumentStop(step uint64, _ []*Cell, locations [][2]int) bool {
	h.Instrument(step, nil, locations)
	return false
}

func (h *FreshnessHeatMapInstrument) Instrument(step uint64, _ []*Cell, locations [][2]int) {
	if step > h.step {
		h.step = step
		h.steps++
		// fade existing heat...
		for i := range h.values {
			h.values[i] *= h.decay
		}
		// changed cells become "hot"...
		width := h.grid.Width
		for _, loc := range locations {
			cp := loc[0]*width + loc[1]
			h.values[cp] = 1.0
		}
	}
}

func (h *FreshnessHeatMapInstrument) HeatMap() iter.Seq[HeatLocation] {
	return func(yield func(HeatLocation) bool) {
		width := h.grid.Width
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

func (h *FreshnessHeatMapInstrument) Maximum() uint64 {
	return 1
}

func (h *FreshnessHeatMapInstrument) StepsCount() uint64 {
	return h.steps
}

type CompositeInstrument []DualUseInstrumentation

var _ StepInstrumentation = (CompositeInstrument)(nil)
var _ StepStopInstrumentation = (CompositeInstrument)(nil)
var _ DualUseInstrumentation = (CompositeInstrument)(nil)

func (c CompositeInstrument) Instrument(step uint64, changes []*Cell, locations [][2]int) {
	for _, inst := range c {
		inst.Instrument(step, changes, locations)
	}
}

func (c CompositeInstrument) InstrumentStop(step uint64, changes []*Cell, locations [][2]int) bool {
	for _, inst := range c {
		if inst.InstrumentStop(step, changes, locations) {
			return true
		}
	}
	return false
}
