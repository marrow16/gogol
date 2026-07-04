package logic

import (
	"iter"
	"slices"
)

type RepeatInstrument struct {
	hash       uint64
	zobrist    [][]uint64
	seen       map[uint64]uint64
	FirstStep  uint64
	RepeatStep uint64
	Period     uint64
	Found      bool
}

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
	}
}

func (r *RepeatInstrument) InstrumentStop(step uint64, _ []*Cell, locations [][2]int) bool {
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

type CompositeInstrument []any

func (c CompositeInstrument) Instrument(step uint64, changes []*Cell, locations [][2]int) {
	for _, inst := range c {
		if i, ok := inst.(StepInstrumentation); ok {
			i.Instrument(step, changes, locations)
		}
	}
}

func (c CompositeInstrument) InstrumentStop(step uint64, changes []*Cell, locations [][2]int) bool {
	for _, inst := range c {
		if i, ok := inst.(StepStopInstrumentation); ok {
			if i.InstrumentStop(step, changes, locations) {
				return true
			}
		} else if i, ok := inst.(StepInstrumentation); ok {
			i.Instrument(step, changes, locations)
		}
	}
	return false
}
