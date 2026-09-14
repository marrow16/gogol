package logic

import (
	"iter"
	"slices"
)

func NewRecordInstrument(g *Grid) *RecordInstrument {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return &RecordInstrument{
		grid:    g,
		frames:  make([]frame, 0),
		initial: slices.Clone(g.cells),
	}
}

type RecordInstrument struct {
	grid    *Grid
	frames  []frame
	initial []uint8
	step    uint64
}

type frame struct {
	step      uint64
	locations []int
}

var _ StepInstrumentation = (*RecordInstrument)(nil)
var _ StepStopInstrumentation = (*RecordInstrument)(nil)
var _ DualUseInstrumentation = (*RecordInstrument)(nil)

func (r *RecordInstrument) FramesCount() int {
	return len(r.frames)
}

func (r *RecordInstrument) InitialGrid() [][]bool {
	result := make([][]bool, r.grid.height)
	idx := 0
	for row := range r.grid.height {
		result[row] = make([]bool, r.grid.width)
		for col := range r.grid.width {
			result[row][col] = r.initial[idx] == aliveCell
			idx++
		}
	}
	return result
}

func (r *RecordInstrument) StepChangeLocations() iter.Seq[[][2]int] {
	return func(yield func([][2]int) bool) {
		for _, f := range r.frames {
			locations := make([][2]int, len(f.locations))
			for i, idx := range f.locations {
				locations[i] = [2]int{
					idx / r.grid.width,
					idx % r.grid.width,
				}
			}
			if !yield(locations) {
				return
			}
		}
	}
}

func (r *RecordInstrument) InstrumentStop(step uint64, locations []int) bool {
	r.Instrument(step, locations)
	return false
}

func (r *RecordInstrument) Instrument(step uint64, locations []int) {
	if step > r.step {
		r.step = step
		r.frames = append(r.frames, frame{
			step:      step,
			locations: append([]int(nil), locations...),
		})
	}
}

func (r *RecordInstrument) restoreInitial() {
	r.grid.StepCount.Store(0)
	r.step = 0
	r.frames = r.frames[:0]
	for idx, c := range r.initial {
		r.grid.cells[idx] = c
	}
}

func (r *RecordInstrument) Undo() {
	r.grid.mutex.Lock()
	defer r.grid.mutex.Unlock()
	if len(r.frames) == 0 {
		// restore original grid state
		r.restoreInitial()
		return
	}
	f := r.frames[len(r.frames)-1]
	r.frames = r.frames[:len(r.frames)-1]
	for _, idx := range f.locations {
		r.grid.cells[idx] ^= aliveCell
	}
	r.grid.StepCount.Store(f.step - 1)
	r.step = f.step - 1
}

func (r *RecordInstrument) Undos(n int) {
	r.grid.mutex.Lock()
	defer r.grid.mutex.Unlock()
	if n > len(r.frames) {
		r.restoreInitial()
		return
	}
	undone := 0
	for undone < n && len(r.frames) > 0 {
		f := r.frames[len(r.frames)-1]
		r.frames = r.frames[:len(r.frames)-1]
		for _, idx := range f.locations {
			r.grid.cells[idx] ^= aliveCell
		}
		r.grid.StepCount.Store(f.step - 1)
		r.step = f.step - 1
		undone++
	}
}
