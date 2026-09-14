package logic

import (
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func nullRender(row, col int, alive, changed bool) {}

type RenderCell func(row, col int, alive, changed bool)

var ErrInvalidGridDimension = errors.New("invalid grid dimension")

func NewGrid(height, width int, rule Rule, wrapMode WrapMode, boundaryMode BoundaryMode) (*Grid, error) {
	if height < 2 || width < 2 {
		return nil, ErrInvalidGridDimension
	}
	size := height * width
	result := &Grid{
		rule:          rule,
		height:        height,
		width:         width,
		cells:         make([]uint8, size),
		wrapMode:      wrapMode,
		boundaryMode:  boundaryMode,
		changesBuffer: make([]int, 0, size),
		renderer:      nullRender,
	}
	result.buildRuleCache()
	result.buildChangeFuncs()
	return result, nil
}

type Grid struct {
	rule          Rule
	width         int
	height        int
	wrapMode      WrapMode
	boundaryMode  BoundaryMode
	cells         []uint8
	changesBuffer []int
	renderer      RenderCell
	ruleCache     [2][9]bool
	changeFns     [16]func(idx int) bool
	mutex         sync.Mutex
	StepCount     atomic.Uint64
}

const (
	deadIndex  = 0
	aliveIndex = 1
	deadCell   = uint8(0)
	aliveCell  = uint8(1)
)

func (g *Grid) buildRuleCache() {
	for i := range 9 {
		g.ruleCache[deadIndex][i] = g.rule.StateChanged(false, uint8(i))
		g.ruleCache[aliveIndex][i] = g.rule.StateChanged(true, uint8(i))
	}
}

const (
	fnIdxInner  = 0
	fnIdxTop    = 1
	fnIdxBottom = 2
	fnIdxLeft   = 4
	fnIdxRight  = 8
)

func (g *Grid) buildChangeFuncs() {
	w := g.width
	h := g.height
	n := w * h
	wrapH := g.wrapMode == WrapHorizontal || g.wrapMode == WrapAll
	wrapV := g.wrapMode == WrapVertical || g.wrapMode == WrapAll
	var boundary uint8
	if g.boundaryMode == AliveBoundary {
		boundary = 1
	}
	// inner cells
	g.changeFns[fnIdxInner] = func(idx int) bool {
		alives := g.cells[idx-w-1] +
			g.cells[idx-w] +
			g.cells[idx-w+1] +
			g.cells[idx-1] +
			g.cells[idx+1] +
			g.cells[idx+w-1] +
			g.cells[idx+w] +
			g.cells[idx+w+1]
		return g.ruleCache[g.cells[idx]][alives]
	}
	lastRow := (h - 1) * w
	// top cells
	if wrapV {
		g.changeFns[fnIdxTop] = func(idx int) bool {
			alives := g.cells[idx+lastRow-1] +
				g.cells[idx+lastRow] +
				g.cells[idx+lastRow+1] +
				g.cells[idx-1] +
				g.cells[idx+1] +
				g.cells[idx+w-1] +
				g.cells[idx+w] +
				g.cells[idx+w+1]
			return g.ruleCache[g.cells[idx]][alives]
		}
	} else {
		g.changeFns[fnIdxTop] = func(idx int) bool {
			alives := boundary*3 +
				g.cells[idx-1] +
				g.cells[idx+1] +
				g.cells[idx+w-1] +
				g.cells[idx+w] +
				g.cells[idx+w+1]
			return g.ruleCache[g.cells[idx]][alives]
		}
	}
	// bottom cells
	if wrapV {
		g.changeFns[fnIdxBottom] = func(idx int) bool {
			alives := g.cells[idx-w-1] +
				g.cells[idx-w] +
				g.cells[idx-w+1] +
				g.cells[idx-1] +
				g.cells[idx+1] +
				g.cells[idx-lastRow-1] +
				g.cells[idx-lastRow] +
				g.cells[idx-lastRow+1]
			return g.ruleCache[g.cells[idx]][alives]
		}
	} else {
		g.changeFns[fnIdxBottom] = func(idx int) bool {
			alives := g.cells[idx-w-1] +
				g.cells[idx-w] +
				g.cells[idx-w+1] +
				g.cells[idx-1] +
				g.cells[idx+1] +
				boundary*3
			return g.ruleCache[g.cells[idx]][alives]
		}
	}
	// left cells
	if wrapH {
		g.changeFns[fnIdxLeft] = func(idx int) bool {
			alives := g.cells[idx-1] + // top-left, wrapped
				g.cells[idx-w] +
				g.cells[idx-w+1] +
				g.cells[idx+w-1] + // left, wrapped
				g.cells[idx+1] +
				g.cells[idx+2*w-1] + // bottom-left, wrapped
				g.cells[idx+w] +
				g.cells[idx+w+1]
			return g.ruleCache[g.cells[idx]][alives]
		}
	} else {
		g.changeFns[fnIdxLeft] = func(idx int) bool {
			alives := boundary*3 +
				g.cells[idx-w] +
				g.cells[idx-w+1] +
				g.cells[idx+1] +
				g.cells[idx+w] +
				g.cells[idx+w+1]
			return g.ruleCache[g.cells[idx]][alives]
		}
	}
	// right cells
	if wrapH {
		g.changeFns[fnIdxRight] = func(idx int) bool {
			alives := g.cells[idx-w-1] +
				g.cells[idx-w] +
				g.cells[idx-2*w+1] + // top-right, wrapped
				g.cells[idx-1] +
				g.cells[idx-w+1] + // right, wrapped
				g.cells[idx+w-1] +
				g.cells[idx+w] +
				g.cells[idx+1] // bottom-right, wrapped
			return g.ruleCache[g.cells[idx]][alives]
		}
	} else {
		g.changeFns[fnIdxRight] = func(idx int) bool {
			alives := g.cells[idx-w-1] +
				g.cells[idx-w] +
				g.cells[idx-1] +
				g.cells[idx+w-1] +
				g.cells[idx+w] +
				boundary*3
			return g.ruleCache[g.cells[idx]][alives]
		}
	}
	// top-left cells
	switch {
	case wrapH && wrapV:
		g.changeFns[fnIdxTop|fnIdxLeft] = func(idx int) bool {
			alives := g.cells[n-1] +
				g.cells[lastRow] +
				g.cells[lastRow+1] +
				g.cells[w-1] +
				g.cells[1] +
				g.cells[2*w-1] +
				g.cells[w] +
				g.cells[w+1]
			return g.ruleCache[g.cells[idx]][alives]
		}
	case wrapH:
		g.changeFns[fnIdxTop|fnIdxLeft] = func(idx int) bool {
			alives := boundary*3 +
				g.cells[w-1] +
				g.cells[1] +
				g.cells[2*w-1] +
				g.cells[w] +
				g.cells[w+1]
			return g.ruleCache[g.cells[idx]][alives]
		}
	case wrapV:
		g.changeFns[fnIdxTop|fnIdxLeft] = func(idx int) bool {
			alives := boundary*3 +
				g.cells[lastRow] +
				g.cells[lastRow+1] +
				g.cells[1] +
				g.cells[w] +
				g.cells[w+1]
			return g.ruleCache[g.cells[idx]][alives]
		}
	default:
		g.changeFns[fnIdxTop|fnIdxLeft] = func(idx int) bool {
			alives := boundary*5 +
				g.cells[1] +
				g.cells[w] +
				g.cells[w+1]
			return g.ruleCache[g.cells[idx]][alives]
		}
	}
	// top-right cells
	switch {
	case wrapH && wrapV:
		g.changeFns[fnIdxTop|fnIdxRight] = func(idx int) bool {
			alives := g.cells[n-2] +
				g.cells[n-1] +
				g.cells[lastRow] +
				g.cells[idx-1] +
				g.cells[0] +
				g.cells[idx+w-1] +
				g.cells[idx+w] +
				g.cells[w]
			return g.ruleCache[g.cells[idx]][alives]
		}
	case wrapH:
		g.changeFns[fnIdxTop|fnIdxRight] = func(idx int) bool {
			alives := boundary*3 +
				g.cells[idx-1] +
				g.cells[0] +
				g.cells[idx+w-1] +
				g.cells[idx+w] +
				g.cells[w]
			return g.ruleCache[g.cells[idx]][alives]
		}
	case wrapV:
		g.changeFns[fnIdxTop|fnIdxRight] = func(idx int) bool {
			alives := boundary*3 +
				g.cells[n-2] +
				g.cells[n-1] +
				g.cells[idx-1] +
				g.cells[idx+w-1] +
				g.cells[idx+w]
			return g.ruleCache[g.cells[idx]][alives]
		}
	default:
		g.changeFns[fnIdxTop|fnIdxRight] = func(idx int) bool {
			alives := boundary*5 +
				g.cells[idx-1] +
				g.cells[idx+w-1] +
				g.cells[idx+w]
			return g.ruleCache[g.cells[idx]][alives]
		}
	}
	// bottom-left cells
	switch {
	case wrapH && wrapV:
		g.changeFns[fnIdxBottom|fnIdxLeft] = func(idx int) bool {
			alives := g.cells[idx-1] +
				g.cells[idx-w] +
				g.cells[idx-w+1] +
				g.cells[n-1] +
				g.cells[idx+1] +
				g.cells[w-1] +
				g.cells[0] +
				g.cells[1]
			return g.ruleCache[g.cells[idx]][alives]
		}
	case wrapH:
		g.changeFns[fnIdxBottom|fnIdxLeft] = func(idx int) bool {
			alives := g.cells[idx-1] +
				g.cells[idx-w] +
				g.cells[idx-w+1] +
				g.cells[n-1] +
				g.cells[idx+1] +
				boundary*3
			return g.ruleCache[g.cells[idx]][alives]
		}
	case wrapV:
		g.changeFns[fnIdxBottom|fnIdxLeft] = func(idx int) bool {
			alives := boundary*3 +
				g.cells[idx-w] +
				g.cells[idx-w+1] +
				g.cells[idx+1] +
				g.cells[0] +
				g.cells[1]
			return g.ruleCache[g.cells[idx]][alives]
		}
	default:
		g.changeFns[fnIdxBottom|fnIdxLeft] = func(idx int) bool {
			alives := boundary*5 +
				g.cells[idx-w] +
				g.cells[idx-w+1] +
				g.cells[idx+1]
			return g.ruleCache[g.cells[idx]][alives]
		}
	}
	// bottom-right cells
	switch {
	case wrapH && wrapV:
		g.changeFns[fnIdxBottom|fnIdxRight] = func(idx int) bool {
			alives := g.cells[idx-w-1] +
				g.cells[idx-w] +
				g.cells[idx-2*w+1] +
				g.cells[idx-1] +
				g.cells[lastRow] +
				g.cells[w-2] +
				g.cells[w-1] +
				g.cells[0]
			return g.ruleCache[g.cells[idx]][alives]
		}
	case wrapH:
		g.changeFns[fnIdxBottom|fnIdxRight] = func(idx int) bool {
			alives := g.cells[idx-w-1] +
				g.cells[idx-w] +
				g.cells[idx-2*w+1] +
				g.cells[idx-1] +
				g.cells[lastRow] +
				boundary*3
			return g.ruleCache[g.cells[idx]][alives]
		}
	case wrapV:
		g.changeFns[fnIdxBottom|fnIdxRight] = func(idx int) bool {
			alives := g.cells[idx-w-1] +
				g.cells[idx-w] +
				g.cells[idx-1] +
				g.cells[w-2] +
				g.cells[w-1] +
				boundary*3
			return g.ruleCache[g.cells[idx]][alives]
		}
	default:
		g.changeFns[fnIdxBottom|fnIdxRight] = func(idx int) bool {
			alives := g.cells[idx-w-1] +
				g.cells[idx-w] +
				g.cells[idx-1] +
				boundary*5
			return g.ruleCache[g.cells[idx]][alives]
		}
	}
}

func (g *Grid) Rule() Rule {
	return g.rule
}

func (g *Grid) SetRule(r Rule) *Grid {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.rule = r
	g.buildRuleCache()
	g.buildChangeFuncs()
	return g
}

func (g *Grid) SetRenderer(r RenderCell) *Grid {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if r == nil {
		g.renderer = nullRender
	} else {
		g.renderer = r
	}
	return g
}

func (g *Grid) Width() int {
	return g.width
}

func (g *Grid) Height() int {
	return g.height
}

func (g *Grid) WrapMode() WrapMode {
	return g.wrapMode
}

func (g *Grid) SetWrapMode(m WrapMode) *Grid {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.wrapMode = m
	g.buildChangeFuncs()
	return g
}

func (g *Grid) BoundaryMode() BoundaryMode {
	return g.boundaryMode
}

func (g *Grid) SetBoundaryMode(m BoundaryMode) *Grid {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.boundaryMode = m
	g.buildChangeFuncs()
	return g
}

func (g *Grid) SetAll(r Rule, wm WrapMode, bm BoundaryMode) *Grid {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.rule = r
	g.wrapMode = wm
	g.boundaryMode = bm
	g.buildRuleCache()
	g.buildChangeFuncs()
	return g
}

func (g *Grid) GetCell(row, col int) (alive bool) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if row >= 0 && row < g.height && col >= 0 && col < g.width {
		return g.cells[(row*g.width)+col] == aliveCell
	}
	return false
}

func (g *Grid) SetCell(row, col int, alive bool) (changed bool) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if row >= 0 && row < g.height && col >= 0 && col < g.width {
		g.StepCount.Store(0)
		idx := (row * g.width) + col
		was := g.cells[idx]
		set := deadCell
		if alive {
			set |= aliveCell
		}
		changed = was != set
		g.cells[idx] = set
		g.renderer(row, col, alive, changed)
	}
	return changed
}

func (g *Grid) Clear() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.StepCount.Store(0)
	clear(g.cells)
	for row := range g.height {
		for col := range g.width {
			g.renderer(row, col, false, true)
		}
	}
}

func (g *Grid) Draw() {
	idx := 0
	for row := range g.height {
		for col := range g.width {
			g.renderer(row, col, g.cells[idx] == aliveCell, true)
			idx++
		}
	}
}

func (g *Grid) DrawTo(render func(row, col int, alive bool)) {
	idx := 0
	for row := range g.height {
		for col := range g.width {
			render(row, col, g.cells[idx] == aliveCell)
			idx++
		}
	}
}

func (g *Grid) Population() int {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	result := 0
	for _, c := range g.cells {
		if c == aliveCell {
			result++
		}
	}
	return result
}

func (g *Grid) Randomize(rf int) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.StepCount.Store(0)
	idx := 0
	for row := range g.height {
		for col := range g.width {
			if rng.Intn(100) < rf {
				g.renderer(row, col, true, true)
				g.cells[idx] = aliveCell
			} else {
				g.renderer(row, col, false, true)
				g.cells[idx] = deadCell
			}
			idx++
		}
	}
}

func (g *Grid) RandomChanges(rf int) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.StepCount.Store(0)
	idx := 0
	for row := range g.height {
		for col := range g.width {
			if rng.Intn(100) < rf {
				g.cells[idx] ^= aliveCell
				g.renderer(row, col, g.cells[idx] == aliveCell, true)
			}
			idx++
		}
	}
}

func (g *Grid) RandomizePopulation(rf int) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.StepCount.Store(0)
	switch {
	case rf == 0:
		// all dead
		idx := 0
		for row := range g.height {
			for col := range g.width {
				g.renderer(row, col, false, true)
				g.cells[idx] = deadCell
				idx++
			}
		}
		return
	case rf >= 100:
		// all alive
		idx := 0
		for row := range g.height {
			for col := range g.width {
				g.renderer(row, col, true, true)
				g.cells[idx] = aliveCell
				idx++
			}
		}
		return
	}
	total := g.width * g.height
	alive := (total*rf + 50) / 100
	selected := make([]bool, total)
	// random ordering of every cell position...
	order := rng.Perm(total)
	// first `alive` positions in shuffled order...
	for _, index := range order[:alive] {
		selected[index] = true
	}
	idx := 0
	for row := range g.height {
		for col := range g.width {
			if selected[idx] {
				g.renderer(row, col, true, true)
				g.cells[idx] = aliveCell
			} else {
				g.renderer(row, col, false, true)
				g.cells[idx] = deadCell
			}
			idx++
		}
	}
}

func (g *Grid) Step() (bool, int) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.changesBuffer = g.changesBuffer[:0]
	idx := 0
	rowFn := fnIdxTop
	for row := 0; row < g.height; row++ {
		leftFn := g.changeFns[rowFn|fnIdxLeft]
		innerFn := g.changeFns[rowFn]
		rightFn := g.changeFns[rowFn|fnIdxRight]
		// left
		if leftFn(idx) {
			g.changesBuffer = append(g.changesBuffer, idx)
			g.renderer(row, 0, g.cells[idx] == deadCell, true)
		}
		idx++
		// inner
		for col := 1; col < g.width-1; col++ {
			if innerFn(idx) {
				g.changesBuffer = append(g.changesBuffer, idx)
				g.renderer(row, col, g.cells[idx] == deadCell, true)
			}
			idx++
		}
		// right
		if rightFn(idx) {
			g.changesBuffer = append(g.changesBuffer, idx)
			g.renderer(row, g.width-1, g.cells[idx] == deadCell, true)
		}
		idx++
		if row == 0 {
			rowFn = fnIdxInner
		} else if row == g.height-2 {
			rowFn = fnIdxBottom
		}
	}
	l := len(g.changesBuffer)
	if l == 0 {
		return false, 0
	}
	for _, idx = range g.changesBuffer {
		g.cells[idx] ^= 1
	}
	g.StepCount.Add(1)
	return true, l
}

func (g *Grid) StepAhead(by int) int {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	count := uint64(0)
	changes := 0
	for range by {
		g.changesBuffer = g.changesBuffer[:0]
		idx := 0
		rowFn := fnIdxTop
		for row := 0; row < g.height; row++ {
			leftFn := g.changeFns[rowFn|fnIdxLeft]
			innerFn := g.changeFns[rowFn]
			rightFn := g.changeFns[rowFn|fnIdxRight]
			// left
			if leftFn(idx) {
				g.changesBuffer = append(g.changesBuffer, idx)
			}
			idx++
			// inner
			for col := 1; col < g.width-1; col++ {
				if innerFn(idx) {
					g.changesBuffer = append(g.changesBuffer, idx)
				}
				idx++
			}
			// right
			if rightFn(idx) {
				g.changesBuffer = append(g.changesBuffer, idx)
			}
			idx++
			if row == 0 {
				rowFn = fnIdxInner
			} else if row == g.height-2 {
				rowFn = fnIdxBottom
			}
		}
		changes = len(g.changesBuffer)
		if changes == 0 {
			break
		}
		count++
		for _, idx = range g.changesBuffer {
			g.cells[idx] ^= aliveCell
		}
	}
	g.StepCount.Add(count)
	return changes
}

// StepWithInstrumentation
// the after StepInstrumentation is called synchronously while the grid is locked.
// Do not call back into Grid from an instrumentation implementation.
// changes and locations are only valid for the duration of the call.
func (g *Grid) StepWithInstrumentation(instrument StepInstrumentation) (bool, int) {
	if instrument == nil {
		return g.Step()
	}
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.changesBuffer = g.changesBuffer[:0]
	idx := 0
	rowFn := fnIdxTop
	for row := 0; row < g.height; row++ {
		leftFn := g.changeFns[rowFn|fnIdxLeft]
		innerFn := g.changeFns[rowFn]
		rightFn := g.changeFns[rowFn|fnIdxRight]
		// left
		if leftFn(idx) {
			g.changesBuffer = append(g.changesBuffer, idx)
			g.renderer(row, 0, g.cells[idx] == deadCell, true)
		}
		idx++
		// inner
		for col := 1; col < g.width-1; col++ {
			if innerFn(idx) {
				g.changesBuffer = append(g.changesBuffer, idx)
				g.renderer(row, col, g.cells[idx] == deadCell, true)
			}
			idx++
		}
		// right
		if rightFn(idx) {
			g.changesBuffer = append(g.changesBuffer, idx)
			g.renderer(row, g.width-1, g.cells[idx] == deadCell, true)
		}
		idx++
		if row == 0 {
			rowFn = fnIdxInner
		} else if row == g.height-2 {
			rowFn = fnIdxBottom
		}
	}
	l := len(g.changesBuffer)
	if l == 0 {
		return false, 0
	}
	for _, idx = range g.changesBuffer {
		g.cells[idx] ^= 1
	}
	step := g.StepCount.Add(1)
	instrument.Instrument(step, g.changesBuffer)
	return true, l
}

// StepAheadWithInstrumentation
// steps ahead a given number of steps (without rendering calls)
//
// before receives the pending changes for a step that has not yet been applied.
// returning true prevents the step from being executed.
//
// after receives the same changes after they have been applied.
// returning true stops further execution.
func (g *Grid) StepAheadWithInstrumentation(by int, instrument StepStopInstrumentation) (StopReason, int) {
	if instrument == nil {
		return StepCompleted, g.StepAhead(by)
	}
	g.mutex.Lock()
	defer g.mutex.Unlock()
	reason := StepCompleted
	count := uint64(0)
	step := g.StepCount.Load() + 1
	changes := 0
	for range by {
		g.changesBuffer = g.changesBuffer[:0]
		idx := 0
		rowFn := fnIdxTop
		for row := 0; row < g.height; row++ {
			leftFn := g.changeFns[rowFn|fnIdxLeft]
			innerFn := g.changeFns[rowFn]
			rightFn := g.changeFns[rowFn|fnIdxRight]
			// left
			if leftFn(idx) {
				g.changesBuffer = append(g.changesBuffer, idx)
			}
			idx++
			// inner
			for col := 1; col < g.width-1; col++ {
				if innerFn(idx) {
					g.changesBuffer = append(g.changesBuffer, idx)
				}
				idx++
			}
			// right
			if rightFn(idx) {
				g.changesBuffer = append(g.changesBuffer, idx)
			}
			idx++
			if row == 0 {
				rowFn = fnIdxInner
			} else if row == g.height-2 {
				rowFn = fnIdxBottom
			}
		}
		changes = len(g.changesBuffer)
		if changes == 0 {
			reason = NoChangesDetected
			break
		}
		for _, idx = range g.changesBuffer {
			g.cells[idx] ^= 1
		}
		count++
		if instrument.InstrumentStop(step, g.changesBuffer) {
			reason = InstrumentStopped
			break
		}
		step++
	}
	g.StepCount.Add(count)
	return reason, changes
}

// LimitAliveAdjacents limits the number of alive neighbours across the entire grid
//
// Note: this does not render changed cells!
func (g *Grid) LimitAliveAdjacents(maximum int) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	switch {
	case maximum >= 8:
		g.StepCount.Store(0)
		return
	case maximum <= 0:
		g.StepCount.Store(0)
		clear(g.cells)
		return
	}
	total := len(g.cells)
	counts := make([]uint8, total)
	var buckets [9][]int
	for idx := range g.cells {
		n := g.aliveAdjacentCount(idx)
		counts[idx] = n
		if int(n) > maximum {
			buckets[n] = append(buckets[n], idx)
		}
	}
	for highest := 8; highest > maximum; {
		if len(buckets[highest]) == 0 {
			highest--
			continue
		}
		candidates := buckets[highest]
		n := rng.Intn(len(candidates))
		idx := candidates[n]
		last := len(candidates) - 1
		candidates[n] = candidates[last]
		buckets[highest] = candidates[:last]
		// stale entry?
		if int(counts[idx]) != highest {
			continue
		}
		// pick a live neighbour to kill...
		var live [8]int
		liveCount := 0
		g.forEachAdjacentIndex(idx, func(adjIdx int) {
			if g.cells[adjIdx] == aliveCell {
				live[liveCount] = adjIdx
				liveCount++
			}
		})
		// counts[idx] says this should be impossible unless the count includes live boundary cells...
		if liveCount == 0 {
			continue
		}
		killIdx := live[rng.Intn(liveCount)]
		g.cells[killIdx] = deadCell
		// killing this cell reduces the alive-neighbour count of EVERY real neighbour - alive or dead...
		g.forEachAdjacentIndex(killIdx, func(adjIdx int) {
			old := counts[adjIdx]
			counts[adjIdx] = old - 1
			if int(old-1) > maximum {
				buckets[old-1] = append(buckets[old-1], adjIdx)
			}
		})
	}
	g.StepCount.Store(0)
}

func (g *Grid) aliveAdjacentCount(idx int) uint8 {
	row := idx / g.width
	col := idx - row*g.width
	var count uint8
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			if dr == 0 && dc == 0 {
				continue
			}
			r := row + dr
			c := col + dc
			if r < 0 {
				if g.wrapMode == WrapVertical || g.wrapMode == WrapAll {
					r = g.height - 1
				} else {
					if g.boundaryMode == AliveBoundary {
						count++
					}
					continue
				}
			} else if r >= g.height {
				if g.wrapMode == WrapVertical || g.wrapMode == WrapAll {
					r = 0
				} else {
					if g.boundaryMode == AliveBoundary {
						count++
					}
					continue
				}
			}
			if c < 0 {
				if g.wrapMode == WrapHorizontal || g.wrapMode == WrapAll {
					c = g.width - 1
				} else {
					if g.boundaryMode == AliveBoundary {
						count++
					}
					continue
				}
			} else if c >= g.width {
				if g.wrapMode == WrapHorizontal || g.wrapMode == WrapAll {
					c = 0
				} else {
					if g.boundaryMode == AliveBoundary {
						count++
					}
					continue
				}
			}
			count += g.cells[r*g.width+c]
		}
	}
	return count
}

func (g *Grid) forEachAdjacentIndex(idx int, fn func(int)) {
	row := idx / g.width
	col := idx - row*g.width
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			if dr == 0 && dc == 0 {
				continue
			}
			r := row + dr
			c := col + dc
			if r < 0 {
				if g.wrapMode == WrapVertical || g.wrapMode == WrapAll {
					r = g.height - 1
				} else {
					continue
				}
			} else if r >= g.height {
				if g.wrapMode == WrapVertical || g.wrapMode == WrapAll {
					r = 0
				} else {
					continue
				}
			}
			if c < 0 {
				if g.wrapMode == WrapHorizontal || g.wrapMode == WrapAll {
					c = g.width - 1
				} else {
					continue
				}
			} else if c >= g.width {
				if g.wrapMode == WrapHorizontal || g.wrapMode == WrapAll {
					c = 0
				} else {
					continue
				}
			}
			fn(r*g.width + c)
		}
	}
}
