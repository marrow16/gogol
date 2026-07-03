package logic

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

type Frame struct {
	Step      uint64
	Locations [][2]int
}

type RecordInstrument struct {
	Grid   *Grid
	Frames []Frame
}

func (r *RecordInstrument) InstrumentStop(step uint64, _ []*Cell, locations [][2]int) bool {
	r.Instrument(step, nil, locations)
	return false
}

func (r *RecordInstrument) Instrument(step uint64, _ []*Cell, locations [][2]int) {
	r.Frames = append(r.Frames, Frame{
		Step:      step,
		Locations: append([][2]int(nil), locations...),
	})
}

func (r *RecordInstrument) Undo() bool {
	r.Grid.mutex.Lock()
	defer r.Grid.mutex.Unlock()
	if len(r.Frames) == 0 {
		return false
	}
	frame := r.Frames[len(r.Frames)-1]
	r.Frames = r.Frames[:len(r.Frames)-1]
	for _, loc := range frame.Locations {
		r.Grid.Rows[loc[0]][loc[1]].flip()
	}
	r.Grid.StepCount.Store(frame.Step - 1)
	return true
}

func (r *RecordInstrument) Undos(n int) int {
	r.Grid.mutex.Lock()
	defer r.Grid.mutex.Unlock()
	undone := 0
	for undone < n && len(r.Frames) > 0 {
		frame := r.Frames[len(r.Frames)-1]
		r.Frames = r.Frames[:len(r.Frames)-1]
		for _, loc := range frame.Locations {
			r.Grid.Rows[loc[0]][loc[1]].flip()
		}
		r.Grid.StepCount.Store(frame.Step - 1)
		undone++
	}
	return undone
}

func (r *RecordInstrument) UndoTo(step uint64) bool {
	r.Grid.mutex.Lock()
	defer r.Grid.mutex.Unlock()
	changed := false
	for len(r.Frames) > 0 {
		frame := r.Frames[len(r.Frames)-1]
		if frame.Step <= step {
			break
		}
		r.Frames = r.Frames[:len(r.Frames)-1]
		for _, loc := range frame.Locations {
			r.Grid.Rows[loc[0]][loc[1]].flip()
		}
		changed = true
	}
	if changed {
		r.Grid.StepCount.Store(step)
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
