package widgets

import (
	"errors"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"
)

func (g *Core) start() {
	g.mutex.Lock()
	g.clearMode()
	if g.running {
		g.mutex.Unlock()
		return
	}
	g.running = true
	g.stopRun = make(chan struct{})
	stop := g.stopRun
	delay := time.Duration(g.settings.StepDelay) * time.Millisecond
	g.mutex.Unlock()
	go func() {
		defer func() {
			g.mutex.Lock()
			if g.stopRun == stop {
				g.running = false
				g.stopRun = nil
			}
			g.mutex.Unlock()
		}()
		for {
			select {
			case <-stop:
				return
			default:
			}
			start := time.Now()
			if !g.gridHolder.grid.StepWithInstrumentation(g.instrumentation) {
				time.Sleep(50 * time.Millisecond)
				g.gridHolder.dirty = true
				g.window.Invalidate()
				return
			}
			g.window.Invalidate()
			if sleep := delay - time.Since(start); sleep > 0 {
				timer := time.NewTimer(sleep)
				select {
				case <-stop:
					timer.Stop()
					return
				case <-timer.C:
				}
			}
		}
	}()
}

func (g *Core) stop() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.clearMode()
	if g.running && g.stopRun != nil {
		close(g.stopRun)
		g.stopRun = nil
	}
	g.running = false
	g.window.Invalidate()
}

func (g *Core) stopRunning() {
	if g.running && g.stopRun != nil {
		close(g.stopRun)
		g.stopRun = nil
	}
	g.running = false
	g.window.Invalidate()
}

func (g *Core) step() {
	g.mutex.Lock()
	g.clearMode()
	defer g.mutex.Unlock()
	g.stopRunning()
	g.gridHolder.grid.StepWithInstrumentation(g.instrumentation)
	g.window.Invalidate()
}

func (g *Core) stepAhead() {
	g.mutex.Lock()
	g.clearMode()
	defer g.mutex.Unlock()
	g.stopRunning()
	if g.stepAheadQueued {
		return
	}
	g.stepAheadQueued = true
	g.status = "Stepping ahead " + strconv.Itoa(g.settings.StepAheadBy)
	if g.settings.StepAheadSnapshot {
		if pattern, err := g.settings.PatternFromGrid(g.gridHolder.grid); err == nil {
			g.snapshotsStep = append(g.snapshotsStep, g.gridHolder.grid.StepCount.Load())
			g.snapshots = append(g.snapshots, pattern)
		}
	}
	go func() {
		g.gridHolder.grid.StepAheadWithInstrumentation(g.settings.StepAheadBy, nil, g.instrumentation)
		g.gridHolder.grid.Draw()
		g.window.Invalidate()
		g.stepAheadQueued = false
		g.status = ""
	}()
}

func (g *Core) clear() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.clearMode()
	g.stopRunning()
	g.gridHolder.grid.Clear()
	g.resetInstrumentation()
}

func (g *Core) permutationIncrement() {
	g.stop()
	n := g.gridHolder.grid.Rule.Permutation()
	if r, err := logic.NewRuleFromPermutation(n + 1); err == nil {
		g.gridHolder.grid.SetRule(r)
	}
}

func (g *Core) permutationDecrement() {
	g.stop()
	if n := g.gridHolder.grid.Rule.Permutation(); n > 0 {
		if r, err := logic.NewRuleFromPermutation(n - 1); err == nil {
			g.gridHolder.grid.SetRule(r)
		}
	}
}

func (g *Core) standardRule() {
	g.stop()
	g.gridHolder.grid.SetRule(logic.StandardRule)
	g.statusBar.rulesPopup.updateInputs()
}

func (g *Core) bornChange(w string) {
	g.stop()
	bw, sw := g.gridHolder.grid.Rule.BornWith(), g.gridHolder.grid.Rule.SurvivesWith()
	if strings.Contains(bw, w) {
		bw = strings.Replace(bw, w, "", 1)
	} else {
		bw += w
	}
	if r, err := logic.NewRuleRle("", "B"+bw+"/S"+sw); err == nil {
		g.gridHolder.grid.SetRule(r)
		g.statusBar.rulesPopup.updateInputs()
	}
}

func (g *Core) survivesChange(w string) {
	g.stop()
	bw, sw := g.gridHolder.grid.Rule.BornWith(), g.gridHolder.grid.Rule.SurvivesWith()
	if strings.Contains(sw, w) {
		sw = strings.Replace(sw, w, "", 1)
	} else {
		sw += w
	}
	if r, err := logic.NewRuleRle("", "B"+bw+"/S"+sw); err == nil {
		g.gridHolder.grid.SetRule(r)
		g.statusBar.rulesPopup.updateInputs()
	}
}

func (g *Core) zoomIn() {
	g.gridHolder.zoom *= 1.1
}

func (g *Core) zoomOut() {
	newZoom := g.gridHolder.zoom / 1.1
	if float32(g.settings.CellSize)*newZoom >= 1.0 {
		g.gridHolder.zoom = newZoom
	}
}

func (g *Core) gridResize(height, width int) {
	g.stop()
	if g.settings.Height != height || g.settings.Width != width {
		g.settings.Width = width
		g.settings.Height = height
		g.gridHolder.resize()
		g.resetInstrumentation()
	}
}

func (g *Core) decreaseGridWidth() {
	g.stop()
	if g.settings.Width > 2 {
		g.settings.Width--
		g.gridHolder.resize()
		g.resetInstrumentation()
	}
}

func (g *Core) increaseGridWidth() {
	g.stop()
	g.settings.Width++
	g.gridHolder.resize()
	g.resetInstrumentation()
}

func (g *Core) decreaseGridHeight() {
	g.stop()
	if g.settings.Height > 2 {
		g.settings.Height--
		g.gridHolder.resize()
		g.resetInstrumentation()
	}
}

func (g *Core) increaseGridHeight() {
	g.stop()
	g.settings.Height++
	g.gridHolder.resize()
	g.resetInstrumentation()
}

func (g *Core) randomize() {
	g.stop()
	g.gridHolder.grid.Randomize(g.settings.Randomization)
	g.resetInstrumentation()
}

func (g *Core) randomChanges() {
	g.stop()
	g.gridHolder.grid.RandomChanges(g.settings.Randomization)
	g.resetInstrumentation()
}

func (g *Core) setWrapMode(m logic.WrapMode) {
	g.stop()
	g.gridHolder.grid.SetWrapMode(m)
}

func (g *Core) setBoundaryMode(m logic.BoundaryMode) {
	g.stop()
	g.gridHolder.grid.SetBoundaryMode(m)
}

func (g *Core) setRandomization(v int) {
	g.settings.Randomization = v
}

func (g *Core) setCellSize(size int) {
	g.stop()
	if size != g.settings.CellSize && size >= 3 {
		g.settings.CellSize = size
		g.gridHolder.rebuild()
		g.gridHolder.grid.Draw()
	}
}

func (g *Core) setCellBorders(on bool) {
	g.stop()
	if on != g.settings.CellBorders {
		g.settings.CellBorders = on
		g.gridHolder.rebuild()
		g.gridHolder.grid.Draw()
		g.window.Invalidate()
	}
}

func (g *Core) toggleCellBorders() {
	g.stop()
	g.settings.CellBorders = !g.settings.CellBorders
	g.gridHolder.rebuild()
	g.gridHolder.grid.Draw()
	g.window.Invalidate()
}

func (g *Core) showMenu() {
	g.statusBar.showHidePopup(popupMenu)
}

func (g *Core) showLifeRules() {
	g.statusBar.showHidePopup(popupRule)
}

func (g *Core) snapshot() {
	g.stop()
	if pattern, err := g.settings.PatternFromGrid(g.gridHolder.grid); err == nil {
		g.snapshotsStep = append(g.snapshotsStep, g.gridHolder.grid.StepCount.Load())
		g.snapshots = append(g.snapshots, pattern)
	}
}

func (g *Core) undoToSnapshot() {
	g.stop()
	if len(g.snapshots) > 0 {
		pattern := g.snapshots[len(g.snapshots)-1]
		step := g.snapshotsStep[len(g.snapshotsStep)-1]
		g.snapshots = g.snapshots[:len(g.snapshots)-1]
		g.snapshotsStep = g.snapshotsStep[:len(g.snapshotsStep)-1]
		g.gridHolder.grid.StepCount.Store(step)
		pattern.Draw(g.gridHolder.grid, 0, 0, patterns.Rotate0)
		g.resetInstrumentation()
	}
}

func (g *Core) export() (err error) {
	g.stop()
	var p patterns.Pattern
	if p, err = g.settings.PatternFromGrid(g.gridHolder.grid); err == nil {
		now := time.Now()
		p.Name = "Grid " + now.Format("2006-01-02 15:04:05")
		filename := "Grid " + now.Format("2006-01-02T150405") + ".rle"
		var f *os.File
		if f, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644); err != nil {
			if errors.Is(err, fs.ErrExist) {
				err = errors.New("File already exists")
			}
			return
		}
		defer func() {
			_ = f.Close()
		}()
		err = patterns.PatternRleEncode(p, f)
	}
	return err
}

func (g *Core) editMode() {
	g.stop()
	g.clearMode()
	g.statusBar.showHidePopup(popupNone)
	g.gridHolder.startEditing()
	g.mode = editMode
}

func (g *Core) setInstrumentationRepeat(on bool) {
	g.stop()
	if on {
		g.instrumentRepeat = logic.NewRepeatInstrument(g.gridHolder.grid)
	} else {
		g.instrumentRepeat = nil
	}
	g.updateInstrumentation()
}

func (g *Core) setInstrumentationRecord(on bool) {
	g.stop()
	if on {
		g.instrumentRecord = &logic.RecordInstrument{Grid: g.gridHolder.grid}
	} else {
		g.instrumentRecord = nil
	}
	g.updateInstrumentation()
}

func (g *Core) updateInstrumentation() {
	g.instrumentation = nil
	if g.instrumentRepeat != nil {
		g.instrumentation = append(g.instrumentation, g.instrumentRepeat)
	}
	if g.instrumentRecord != nil {
		g.instrumentation = append(g.instrumentation, g.instrumentRecord)
	}
}

func (g *Core) resetInstrumentation() {
	if g.instrumentRepeat != nil {
		g.instrumentRepeat = logic.NewRepeatInstrument(g.gridHolder.grid)
	}
	if g.instrumentRecord != nil {
		g.instrumentRecord = &logic.RecordInstrument{Grid: g.gridHolder.grid}
	}
	g.updateInstrumentation()
}

func (g *Core) placePattern(gtx layout.Context) {
	if g.gridHolder.overlay != nil {
		g.placePatternRow, g.placePatternCol, g.placePatternRotation = g.gridHolder.overlay.row, g.gridHolder.overlay.col, g.gridHolder.overlay.rotation
	}
	g.gridHolder.placeOverlay()
	g.clearMode()
	gtx.Execute(op.InvalidateCmd{})
}
