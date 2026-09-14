package logic

type RepeatInstrument struct {
	hash       uint64
	zobrist    []uint64
	seen       map[uint64]uint64
	Step       uint64
	Steps      uint64
	FirstStep  uint64
	RepeatStep uint64
	Period     uint64
	Found      bool
}

var _ StepInstrumentation = (*RepeatInstrument)(nil)
var _ StepStopInstrumentation = (*RepeatInstrument)(nil)
var _ DualUseInstrumentation = (*RepeatInstrument)(nil)

func NewRepeatInstrument(g *Grid) *RepeatInstrument {
	zobrist := make([]uint64, len(g.cells))
	var hash uint64
	for idx, cell := range g.cells {
		z := rng.Uint64()
		zobrist[idx] = z
		if cell == aliveCell {
			hash ^= z
		}
	}
	step := g.StepCount.Load()
	return &RepeatInstrument{
		hash:    hash,
		zobrist: zobrist,
		seen: map[uint64]uint64{
			hash: step,
		},
		Step: step,
	}
}

func (r *RepeatInstrument) InstrumentStop(step uint64, locations []int) bool {
	if step <= r.Step {
		return false
	}
	r.Step = step
	r.Steps++
	for _, idx := range locations {
		r.hash ^= r.zobrist[idx]
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

func (r *RepeatInstrument) Instrument(step uint64, locations []int) {
	r.InstrumentStop(step, locations)
}
