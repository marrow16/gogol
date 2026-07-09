package logic

type StepInstrumentation interface {
	Instrument(step uint64, changes []*Cell, locations [][2]int)
}

type StepStopInstrumentation interface {
	InstrumentStop(step uint64, changes []*Cell, locations [][2]int) bool
}

type DualUseInstrumentation interface {
	StepInstrumentation
	StepStopInstrumentation
}

type StopReason int

const (
	StepCompleted StopReason = iota
	NoChangesDetected
	InstrumentStoppedOnBefore
	InstrumentStoppedOnAfter
)

func (r StopReason) String() string {
	switch r {
	case StepCompleted:
		return "Step Completed"
	case NoChangesDetected:
		return "No Changes Detected"
	case InstrumentStoppedOnBefore:
		return "Instrument Stopped On Before"
	case InstrumentStoppedOnAfter:
		return "Instrument Stopped On After"
	}
	return "Unknown"
}

// StepWithInstrumentation
// the after StepInstrumentation is called synchronously while the grid is locked.
// Do not call back into Grid from an instrumentation implementation.
// changes and locations are only valid for the duration of the call.
func (g *Grid) StepWithInstrumentation(after StepInstrumentation) (gridChanged bool) {
	if after == nil {
		return g.Step()
	}
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if g.Rule == nil {
		g.Rule = StandardRule
	}
	render := g.Render
	if render == nil {
		render = nullRender
	}
	g.changesBuffer = g.changesBuffer[:0]
	g.locationsBuffer = g.locationsBuffer[:0]
	for r, row := range g.Rows {
		for c, cell := range row {
			if g.Rule.StateChanged(cell) {
				g.changesBuffer = append(g.changesBuffer, cell)
				g.locationsBuffer = append(g.locationsBuffer, [2]int{r, c})
				render(r, c, !cell.Alive, true)
			} else {
				render(r, c, cell.Alive, false)
			}
		}
	}
	if len(g.changesBuffer) == 0 {
		return false
	}
	for _, cell := range g.changesBuffer {
		cell.flip()
	}
	step := g.StepCount.Add(1)
	after.Instrument(step, g.changesBuffer, g.locationsBuffer)
	return true
}

// StepAheadWithInstrumentation
// steps ahead a given number of steps (without rendering calls)
//
// before receives the pending changes for a step that has not yet been applied.
// returning true prevents the step from being executed.
//
// after receives the same changes after they have been applied.
// returning true stops further execution.
func (g *Grid) StepAheadWithInstrumentation(by int, before, after StepStopInstrumentation) StopReason {
	if before == nil && after == nil {
		g.StepAhead(by)
		return StepCompleted
	}
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if g.Rule == nil {
		g.Rule = StandardRule
	}
	reason := StepCompleted
	count := uint64(0)
	step := g.StepCount.Load() + 1
	for i := 0; i < by; i++ {
		g.changesBuffer = g.changesBuffer[:0]
		g.locationsBuffer = g.locationsBuffer[:0]
		for r, row := range g.Rows {
			for c, cell := range row {
				if g.Rule.StateChanged(cell) {
					g.changesBuffer = append(g.changesBuffer, cell)
					g.locationsBuffer = append(g.locationsBuffer, [2]int{r, c})
				}
			}
		}
		if len(g.changesBuffer) == 0 {
			reason = NoChangesDetected
			break
		}
		if before != nil && before.InstrumentStop(step, g.changesBuffer, g.locationsBuffer) {
			reason = InstrumentStoppedOnBefore
			break
		}
		for _, cell := range g.changesBuffer {
			cell.flip()
		}
		count++
		if after != nil && after.InstrumentStop(step, g.changesBuffer, g.locationsBuffer) {
			reason = InstrumentStoppedOnAfter
			break
		}
		step++
	}
	g.StepCount.Add(count)
	return reason
}
