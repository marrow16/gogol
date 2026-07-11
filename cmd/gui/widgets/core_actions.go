package widgets

import (
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"github.com/marrow16/gogol/recipes"
	"image"
	"image/png"
	"os"
	"strconv"
	"strings"
	"time"
)

func (c *Core) start() {
	c.mutex.Lock()
	c.clearMode()
	if c.running {
		c.mutex.Unlock()
		return
	}
	c.running = true
	c.stopRun = make(chan struct{})
	stop := c.stopRun
	delay := time.Duration(c.settings.StepDelay) * time.Millisecond
	c.mutex.Unlock()
	go func() {
		defer func() {
			c.mutex.Lock()
			if c.stopRun == stop {
				c.running = false
				c.stopRun = nil
			}
			c.mutex.Unlock()
		}()
		for {
			select {
			case <-stop:
				return
			default:
			}
			start := time.Now()
			if !c.gridHolder.grid.StepWithInstrumentation(c.instrumentation) {
				time.Sleep(50 * time.Millisecond)
				c.gridHolder.dirty = true
				c.window.Invalidate()
				return
			}
			c.window.Invalidate()
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

func (c *Core) stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clearMode()
	if c.running && c.stopRun != nil {
		close(c.stopRun)
		c.stopRun = nil
	}
	c.running = false
	c.window.Invalidate()
}

func (c *Core) stopRunning() {
	if c.running && c.stopRun != nil {
		close(c.stopRun)
		c.stopRun = nil
	}
	c.running = false
	c.window.Invalidate()
}

func (c *Core) step() {
	c.mutex.Lock()
	c.clearMode()
	defer c.mutex.Unlock()
	c.stopRunning()
	c.gridHolder.grid.StepWithInstrumentation(c.instrumentation)
	c.window.Invalidate()
}

func (c *Core) stepAhead() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clearMode()
	c.stopRunning()
	if c.stepAheadQueued {
		return
	}
	c.stepAheadQueued = true
	c.status = "Stepping ahead " + strconv.Itoa(c.settings.StepAheadBy)
	if c.settings.StepAheadSnapshot {
		if pattern, err := c.settings.PatternFromGrid(c.gridHolder.grid); err == nil {
			c.snapshotsStep = append(c.snapshotsStep, c.gridHolder.grid.StepCount.Load())
			c.snapshots = append(c.snapshots, pattern)
		}
	}
	go func() {
		c.gridHolder.grid.StepAheadWithInstrumentation(c.settings.StepAheadBy, nil, c.instrumentation)
		c.gridHolder.grid.Draw()
		c.window.Invalidate()
		c.stepAheadQueued = false
		c.status = ""
	}()
}

func (c *Core) stepAheadBy(n int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clearMode()
	c.stopRunning()
	c.gridHolder.grid.StepAheadWithInstrumentation(n, nil, c.instrumentation)
	c.gridHolder.grid.Draw()
	c.window.Invalidate()
}

func (c *Core) stepBack() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clearMode()
	c.stopRunning()
	if c.instrumentRecord != nil {
		c.instrumentRecord.Undo()
		c.gridHolder.grid.Draw()
		c.window.Invalidate()
	}
}

func (c *Core) skipBack() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clearMode()
	c.stopRunning()
	if c.skipBackQueued {
		return
	}
	if c.instrumentRecord != nil {
		c.skipBackQueued = true
		c.status = "Skipping back " + strconv.Itoa(c.settings.SkipBackBy)
		go func() {
			c.instrumentRecord.Undos(c.settings.SkipBackBy)
			c.gridHolder.grid.Draw()
			c.window.Invalidate()
			c.skipBackQueued = false
			c.status = ""
		}()
	}
}

func (c *Core) clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clearMode()
	c.stopRunning()
	c.gridHolder.grid.Clear()
	c.resetInstrumentation()
}

func (c *Core) setRule(r logic.Rule) {
	c.stop()
	c.gridHolder.grid.SetRule(r)
}

func (c *Core) permutationIncrement() {
	c.stop()
	n := c.gridHolder.grid.Rule.Permutation()
	if r, err := logic.NewRuleFromPermutation(n + 1); err == nil {
		c.gridHolder.grid.SetRule(r)
	}
}

func (c *Core) permutationDecrement() {
	c.stop()
	if n := c.gridHolder.grid.Rule.Permutation(); n > 0 {
		if r, err := logic.NewRuleFromPermutation(n - 1); err == nil {
			c.gridHolder.grid.SetRule(r)
		}
	}
}

func (c *Core) standardRule() {
	c.stop()
	c.gridHolder.grid.SetRule(logic.StandardRule)
	c.statusBar.rulesPopup.updateInputs()
}

func (c *Core) bornChange(w string) {
	c.stop()
	bw, sw := c.gridHolder.grid.Rule.BornWith(), c.gridHolder.grid.Rule.SurvivesWith()
	if strings.Contains(bw, w) {
		bw = strings.Replace(bw, w, "", 1)
	} else {
		bw += w
	}
	if r, err := logic.NewRuleRle("", "B"+bw+"/S"+sw); err == nil {
		c.gridHolder.grid.SetRule(r)
		c.statusBar.rulesPopup.updateInputs()
	}
}

func (c *Core) survivesChange(w string) {
	c.stop()
	bw, sw := c.gridHolder.grid.Rule.BornWith(), c.gridHolder.grid.Rule.SurvivesWith()
	if strings.Contains(sw, w) {
		sw = strings.Replace(sw, w, "", 1)
	} else {
		sw += w
	}
	if r, err := logic.NewRuleRle("", "B"+bw+"/S"+sw); err == nil {
		c.gridHolder.grid.SetRule(r)
		c.statusBar.rulesPopup.updateInputs()
	}
}

func (c *Core) zoomIn() {
	c.gridHolder.zoom *= 1.1
}

func (c *Core) zoomOut() {
	newZoom := c.gridHolder.zoom / 1.1
	if float32(c.settings.CellSize)*newZoom >= 1.0 {
		c.gridHolder.zoom = newZoom
	}
}

func (c *Core) gridResize(height, width int) {
	c.stop()
	if c.settings.Height != height || c.settings.Width != width {
		c.settings.Width = width
		c.settings.Height = height
		c.gridHolder.resize()
		c.resetInstrumentation()
	}
}

func (c *Core) decreaseGridWidth() {
	c.stop()
	if c.settings.Width > 2 {
		c.settings.Width--
		c.gridHolder.resize()
		c.resetInstrumentation()
	}
}

func (c *Core) increaseGridWidth() {
	c.stop()
	c.settings.Width++
	c.gridHolder.resize()
	c.resetInstrumentation()
}

func (c *Core) decreaseGridHeight() {
	c.stop()
	if c.settings.Height > 2 {
		c.settings.Height--
		c.gridHolder.resize()
		c.resetInstrumentation()
	}
}

func (c *Core) increaseGridHeight() {
	c.stop()
	c.settings.Height++
	c.gridHolder.resize()
	c.resetInstrumentation()
}

func (c *Core) randomize() {
	c.stop()
	c.gridHolder.grid.Randomize(c.settings.Randomization)
	c.resetInstrumentation()
}

func (c *Core) randomChanges() {
	c.stop()
	c.gridHolder.grid.RandomChanges(c.settings.Randomization)
	c.resetInstrumentation()
}

func (c *Core) setWrapMode(m logic.WrapMode) {
	c.stop()
	c.gridHolder.grid.SetWrapMode(m)
}

func (c *Core) setBoundaryMode(m logic.BoundaryMode) {
	c.stop()
	c.gridHolder.grid.SetBoundaryMode(m)
}

func (c *Core) setRandomization(v int) {
	c.settings.Randomization = v
}

func (c *Core) setCellSize(size int) {
	c.stop()
	if size != c.settings.CellSize && size >= 3 {
		c.settings.CellSize = size
		c.gridHolder.rebuild()
		c.gridHolder.grid.Draw()
	}
}

func (c *Core) setCellBorders(on bool) {
	c.stop()
	if on != c.settings.CellBorders {
		c.settings.CellBorders = on
		c.gridHolder.rebuild()
		c.gridHolder.grid.Draw()
		c.window.Invalidate()
	}
}

func (c *Core) toggleCellBorders() {
	c.stop()
	c.settings.CellBorders = !c.settings.CellBorders
	c.gridHolder.rebuild()
	c.gridHolder.grid.Draw()
	c.window.Invalidate()
}

func (c *Core) showMenu() {
	c.statusBar.showHidePopup(popupMenu)
}

func (c *Core) showLifeRules() {
	c.statusBar.showHidePopup(popupRule)
}

func (c *Core) snapshot() {
	c.stop()
	if pattern, err := c.settings.PatternFromGrid(c.gridHolder.grid); err == nil {
		c.snapshotsStep = append(c.snapshotsStep, c.gridHolder.grid.StepCount.Load())
		c.snapshots = append(c.snapshots, pattern)
	}
}

func (c *Core) undoToSnapshot() {
	c.stop()
	if len(c.snapshots) > 0 {
		pattern := c.snapshots[len(c.snapshots)-1]
		step := c.snapshotsStep[len(c.snapshotsStep)-1]
		c.snapshots = c.snapshots[:len(c.snapshots)-1]
		c.snapshotsStep = c.snapshotsStep[:len(c.snapshotsStep)-1]
		c.gridHolder.grid.StepCount.Store(step)
		pattern.Draw(c.gridHolder.grid, 0, 0, patterns.Rotate0)
		c.resetInstrumentation()
	}
}

func (c *Core) export() (err error) {
	c.stop()
	var p patterns.Pattern
	if p, err = c.settings.PatternFromGrid(c.gridHolder.grid); err == nil {
		now := time.Now()
		p.Name = "Grid " + now.Format("2006-01-02 15:04:05")
		filename := "Grid " + now.Format("2006-01-02T150405") + ".rle"
		var f *os.File
		if f, err = saveFile(filename, false); err == nil {
			defer func() {
				_ = f.Close()
			}()
			err = patterns.PatternRleEncode(p, f)
		}
	}
	return err
}

func (c *Core) runRecipe(filename string) {
	c.stop()
	if recipe, err := recipes.Load(filename); err == nil {
		c.resetInstrumentation()
		grid, resized, err := recipe.Run(c.gridHolder.grid, true)
		if err != nil {
			return
		}
		if resized {
			c.settings.Height, c.settings.Width, c.settings.WrapMode, c.settings.BoundaryMode = grid.Height, grid.Width, grid.WrapMode, grid.BoundaryMode
			c.gridHolder.replaceGrid(grid)
			c.resetInstrumentation()
			c.window.Invalidate()
		} else {
			c.resetInstrumentation()
		}
	}
}

func (c *Core) startEditMode() {
	c.stop()
	c.clearMode()
	c.statusBar.showHidePopup(popupNone)
	c.gridHolder.startEditing()
	c.mode = editMode
}

func (c *Core) showHeatMap() {
	c.stop()
	c.clearMode()
	if c.heatMapperType != noHeatMapper && c.instrumentHeatMap != nil {
		c.statusBar.showHidePopup(popupNone)
		c.gridHolder.buildHeatMap(c.instrumentHeatMap)
		c.mode = heatMapMode
	}
}

func (c *Core) setInstrumentationRepeat(on bool) {
	c.stop()
	if on {
		c.instrumentRepeat = logic.NewRepeatInstrument(c.gridHolder.grid)
	} else {
		c.instrumentRepeat = nil
	}
	c.updateInstrumentation()
}

func (c *Core) setInstrumentationRecord(on bool) {
	c.stop()
	if on {
		c.instrumentRecord = logic.NewRecordInstrument(c.gridHolder.grid)
	} else {
		c.instrumentRecord = nil
	}
	c.updateInstrumentation()
}

func (c *Core) setInstrumentationHeatMapper(hmt heatMapperType) {
	c.stop()
	c.heatMapperType = hmt
	c.instrumentHeatMap = c.heatMapperType.newHeatMapper(c.gridHolder.grid, c.settings.HeatMappingHalfLife)
	c.updateInstrumentation()
}

func (c *Core) saveHeatMapImage() {
	if c.instrumentHeatMap != nil {
		filename := "Heat Map " + c.heatMapperType.String() + " " + time.Now().Format("2006-01-02T15-04-05.999") + ".png"
		if f, err := saveFile(filename, false); err == nil {
			defer func() {
				_ = f.Close()
			}()
			img := image.NewNRGBA(image.Rect(0, 0, c.settings.Width*c.settings.CellSize, c.settings.Height*c.settings.CellSize))
			c.gridHolder.drawHeatMap(img, c.instrumentHeatMap)
			_ = png.Encode(f, img)
		}
	}
}

func (c *Core) isRecording() bool {
	if c.instrumentRecord != nil {
		return c.instrumentRecord.StepsCount() > 0
	}
	return false
}

func (c *Core) updateInstrumentation() {
	c.instrumentation = nil
	if c.instrumentRepeat != nil {
		c.instrumentation = append(c.instrumentation, c.instrumentRepeat)
	}
	if c.instrumentRecord != nil {
		c.instrumentation = append(c.instrumentation, c.instrumentRecord)
	}
	if c.instrumentHeatMap != nil {
		c.instrumentation = append(c.instrumentation, c.instrumentHeatMap.(logic.DualUseInstrumentation))
	}
}

func (c *Core) resetInstrumentation() {
	if c.instrumentRepeat != nil {
		c.instrumentRepeat = logic.NewRepeatInstrument(c.gridHolder.grid)
	}
	if c.instrumentRecord != nil {
		c.instrumentRecord = logic.NewRecordInstrument(c.gridHolder.grid)
	}
	c.instrumentHeatMap = c.heatMapperType.newHeatMapper(c.gridHolder.grid, c.settings.HeatMappingHalfLife)
	c.updateInstrumentation()
}

func (c *Core) placePattern(gtx layout.Context) {
	if c.gridHolder.overlay != nil {
		c.placePatternRow, c.placePatternCol, c.placePatternRotation = c.gridHolder.overlay.row, c.gridHolder.overlay.col, c.gridHolder.overlay.rotation
	}
	c.gridHolder.placeOverlay()
	c.clearMode()
	gtx.Execute(op.InvalidateCmd{})
}
