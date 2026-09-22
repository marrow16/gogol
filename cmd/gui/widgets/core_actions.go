package widgets

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/marrow16/gogol/imaging"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"github.com/marrow16/gogol/recipes"
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
	c.hertz.Store(0)
	c.changes.Store(0)
	fps := int64(time.Second / time.Duration(c.settings.Fps))
	if len(c.instrumentation) == 0 {
		go func() {
			defer func() {
				c.hertz.Store(0)
				c.mutex.Lock()
				if c.stopRun == stop {
					c.running = false
					c.stopRun = nil
				}
				c.mutex.Unlock()
			}()
			rateStart := time.Now()
			var rateSteps uint64
			var lastInvalidate int64
			for {
				select {
				case <-stop:
					return
				default:
				}
				start := time.Now()
				stepped, changes := c.gridHolder.grid.Step()
				if !stepped {
					c.hertz.Store(0)
					c.changes.Store(0)
					c.gridHolder.invalidate()
					window.Invalidate()
					return
				}
				rateSteps++
				if elapsed := time.Since(rateStart); elapsed >= time.Second {
					c.changes.Store(int64(changes))
					c.hertz.Store(uint64(float64(rateSteps) / elapsed.Seconds()))
					rateSteps = 0
					rateStart = time.Now()
				}
				if now := time.Now().UnixNano(); now-lastInvalidate >= fps {
					window.Invalidate()
					lastInvalidate = now
				}
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
	} else {
		ignoreRepeat := false
		if c.instrumentRepeat != nil && c.instrumentRepeat.Found {
			ignoreRepeat = true
		}
		go func() {
			defer func() {
				c.hertz.Store(0)
				c.mutex.Lock()
				if c.stopRun == stop {
					c.running = false
					c.stopRun = nil
				}
				c.mutex.Unlock()
			}()
			rateStart := time.Now()
			var rateSteps uint64
			var lastInvalidate int64
			for {
				select {
				case <-stop:
					return
				default:
				}
				start := time.Now()
				stepped, changes := c.gridHolder.grid.StepWithInstrumentation(c.instrumentation)
				if !stepped {
					c.hertz.Store(0)
					c.changes.Store(0)
					c.gridHolder.invalidate()
					window.Invalidate()
					return
				}
				rateSteps++
				if elapsed := time.Since(rateStart); elapsed >= time.Second {
					c.changes.Store(int64(changes))
					c.hertz.Store(uint64(float64(rateSteps) / elapsed.Seconds()))
					rateSteps = 0
					rateStart = time.Now()
				}
				if !ignoreRepeat && c.instrumentRepeat != nil && c.instrumentRepeat.Found {
					c.gridHolder.invalidate()
					window.Invalidate()
					return
				}
				if now := time.Now().UnixNano(); now-lastInvalidate >= fps {
					window.Invalidate()
					lastInvalidate = now
				}
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
	window.Invalidate()
}

func (c *Core) stopRunning() {
	if c.running && c.stopRun != nil {
		close(c.stopRun)
		c.stopRun = nil
	}
	c.running = false
	window.Invalidate()
}

func (c *Core) step() {
	c.mutex.Lock()
	c.clearMode()
	defer c.mutex.Unlock()
	c.stopRunning()
	_, changes := c.gridHolder.grid.StepWithInstrumentation(c.instrumentation)
	c.changes.Store(int64(changes))
	window.Invalidate()
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
		_, changes := c.gridHolder.grid.StepAheadWithInstrumentation(c.settings.StepAheadBy, c.instrumentation)
		c.changes.Store(int64(changes))
		c.gridHolder.grid.Draw()
		window.Invalidate()
		c.stepAheadQueued = false
		c.status = ""
	}()
}

func (c *Core) stepAheadBy(n int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clearMode()
	c.stopRunning()
	_, changes := c.gridHolder.grid.StepAheadWithInstrumentation(n, c.instrumentation)
	c.changes.Store(int64(changes))
	c.gridHolder.grid.Draw()
	window.Invalidate()
}

func (c *Core) stepBack() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clearMode()
	c.stopRunning()
	if c.instrumentRecord != nil {
		c.instrumentRecord.Undo()
		c.gridHolder.grid.Draw()
		window.Invalidate()
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
			window.Invalidate()
			c.skipBackQueued = false
			c.status = ""
		}()
	}
}

func (c *Core) skipBackBy(n int) {
	c.stop()
	if c.instrumentRecord != nil {
		c.instrumentRecord.Undos(n)
		c.gridHolder.grid.Draw()
		window.Invalidate()
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
	c.resetInstrumentation()
	c.statusBar.rulesPopup.updateInputs()
}

func (c *Core) permutationIncrement() {
	c.stop()
	n := c.gridHolder.grid.Rule().Permutation()
	if n+1 < 1<<18 {
		n++
	} else {
		n = 0
	}
	if r, err := logic.NewRuleFromPermutation(n); err == nil {
		c.gridHolder.grid.SetRule(r)
		c.statusBar.rulesPopup.updateInputs()
	}
}

func (c *Core) permutationDecrement() {
	c.stop()
	n := c.gridHolder.grid.Rule().Permutation()
	if n == 0 {
		n = (1 << 18) - 1
	} else {
		n--
	}
	if r, err := logic.NewRuleFromPermutation(n); err == nil {
		c.gridHolder.grid.SetRule(r)
		c.statusBar.rulesPopup.updateInputs()
	}
}

func (c *Core) integerIncrement() {
	c.stop()
	n := c.gridHolder.grid.Rule().Integer()
	if n+1 < 1<<18 {
		n++
	} else {
		n = 0
	}
	if r, err := logic.NewRuleFromInteger(n); err == nil {
		c.gridHolder.grid.SetRule(r)
		c.statusBar.rulesPopup.updateInputs()
	}
}

func (c *Core) integerDecrement() {
	c.stop()
	n := c.gridHolder.grid.Rule().Integer()
	if n == 0 {
		n = (1 << 18) - 1
	} else {
		n--
	}
	if r, err := logic.NewRuleFromInteger(n); err == nil {
		c.gridHolder.grid.SetRule(r)
		c.statusBar.rulesPopup.updateInputs()
	}
}

func (c *Core) permutationIncrementBorn() {
	c.stop()
	perm := c.gridHolder.grid.Rule().Permutation()
	b := (perm >> 9) & 0x1FF
	s := perm & 0x1FF
	b = (b + 1) & 0x1FF
	if r, err := logic.NewRuleFromPermutation((b << 9) | s); err == nil {
		c.gridHolder.grid.SetRule(r)
		c.statusBar.rulesPopup.updateInputs()
	}
}

func (c *Core) permutationDecrementBorn() {
	c.stop()
	perm := c.gridHolder.grid.Rule().Permutation()
	b := (perm >> 9) & 0x1FF
	s := perm & 0x1FF
	if b == 0 {
		b = 0x1FF
	} else {
		b = (b - 1) & 0x1FF
	}
	if r, err := logic.NewRuleFromPermutation((b << 9) | s); err == nil {
		c.gridHolder.grid.SetRule(r)
		c.statusBar.rulesPopup.updateInputs()
	}
}

func (c *Core) permutationIncrementSurvives() {
	c.stop()
	perm := c.gridHolder.grid.Rule().Permutation()
	b := (perm >> 9) & 0x1FF
	s := perm & 0x1FF
	s = (s + 1) & 0x1FF
	if r, err := logic.NewRuleFromPermutation((b << 9) | s); err == nil {
		c.gridHolder.grid.SetRule(r)
		c.statusBar.rulesPopup.updateInputs()
	}
}

func (c *Core) permutationDecrementSurvives() {
	c.stop()
	perm := c.gridHolder.grid.Rule().Permutation()
	b := (perm >> 9) & 0x1FF
	s := perm & 0x1FF
	if s == 0 {
		s = 0x1FF
	} else {
		s = (s - 1) & 0x1FF
	}
	if r, err := logic.NewRuleFromPermutation((b << 9) | s); err == nil {
		c.gridHolder.grid.SetRule(r)
		c.statusBar.rulesPopup.updateInputs()
	}
}

func (c *Core) standardRule() {
	c.stop()
	c.gridHolder.grid.SetRule(logic.StandardRule)
	c.statusBar.rulesPopup.updateInputs()
}

func (c *Core) bornChange(w string) {
	c.stop()
	bw, sw := c.gridHolder.grid.Rule().BornWith(), c.gridHolder.grid.Rule().SurvivesWith()
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
	bw, sw := c.gridHolder.grid.Rule().BornWith(), c.gridHolder.grid.Rule().SurvivesWith()
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
		c.settingsChanged()
	}
}

func (c *Core) decreaseGridWidth() {
	c.stop()
	if c.settings.Width > 2 {
		c.settings.Width--
		c.gridHolder.resize()
		c.resetInstrumentation()
		c.settingsChanged()
	}
}

func (c *Core) increaseGridWidth() {
	c.stop()
	c.settings.Width++
	c.gridHolder.resize()
	c.resetInstrumentation()
	c.settingsChanged()
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

func (c *Core) randomize(rf ...int) {
	c.stop()
	if len(rf) > 0 {
		c.gridHolder.grid.Randomize(rf[0])
	} else {
		c.gridHolder.grid.Randomize(c.settings.Randomization)
	}
	c.resetInstrumentation()
}

func (c *Core) randomizePopulation(rf ...int) {
	c.stop()
	if len(rf) > 0 {
		c.gridHolder.grid.RandomizePopulation(rf[0])
	} else {
		c.gridHolder.grid.RandomizePopulation(c.settings.Randomization)
	}
	c.resetInstrumentation()
}

func (c *Core) population() int {
	c.stop()
	return c.gridHolder.grid.Population()
}

func (c *Core) maximumAdjacents(mx int) {
	c.stop()
	c.gridHolder.grid.LimitAliveAdjacents(mx)
	c.resetInstrumentation()
}

func (c *Core) randomChanges(rf ...int) {
	c.stop()
	if len(rf) > 0 {
		c.gridHolder.grid.RandomChanges(rf[0])
	} else {
		c.gridHolder.grid.RandomChanges(c.settings.Randomization)
	}
	c.resetInstrumentation()
}

func (c *Core) randomAdditions(rf ...int) {
	c.stop()
	if len(rf) > 0 {
		c.gridHolder.grid.RandomAdditions(rf[0])
	} else {
		c.gridHolder.grid.RandomAdditions(c.settings.Randomization)
	}
	c.resetInstrumentation()
}

func (c *Core) randomCull(rf ...int) {
	c.stop()
	if len(rf) > 0 {
		c.gridHolder.grid.RandomCull(rf[0])
	} else {
		c.gridHolder.grid.RandomCull(c.settings.Randomization)
	}
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
	if size != c.settings.CellSize && size > 0 {
		c.settings.CellSize = size
		c.gridHolder.rebuild()
	}
}

func (c *Core) setCellBorders(on bool) {
	c.stop()
	if on != c.settings.CellBorders {
		c.settings.CellBorders = on
		c.gridHolder.rebuild()
		c.settingsChanged()
		window.Invalidate()
	}
}

func (c *Core) toggleCellBorders() {
	c.stop()
	c.settings.CellBorders = !c.settings.CellBorders
	c.gridHolder.rebuild()
	c.settingsChanged()
	window.Invalidate()
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

func (c *Core) replaySnapshot() {
	c.stop()
	if len(c.snapshots) > 0 {
		pattern := c.snapshots[len(c.snapshots)-1]
		step := c.snapshotsStep[len(c.snapshotsStep)-1]
		c.gridHolder.grid.StepCount.Store(step)
		pattern.Draw(c.gridHolder.grid, 0, 0, patterns.Rotate0)
		c.resetInstrumentation()
	} else {
		c.gridHolder.grid.Randomize(c.settings.Randomization)
		c.resetInstrumentation()
		if pattern, err := c.settings.PatternFromGrid(c.gridHolder.grid); err == nil {
			c.snapshotsStep = append(c.snapshotsStep, c.gridHolder.grid.StepCount.Load())
			c.snapshots = append(c.snapshots, pattern)
		}
	}
}

func (c *Core) export() (err error) {
	c.stop()
	var p patterns.Pattern
	if p, err = c.settings.PatternFromGrid(c.gridHolder.grid); err == nil {
		filename := c.nowFilename("Grid Export", ".rle")
		p.Name = filename
		var f *os.File
		if f, err = saveFile(filename, false); err == nil {
			defer func() {
				_ = f.Close()
			}()
			err = patterns.PatternRleEncode(p, f)
		}
		if c.settings.ExportImage {
			_ = c.exportImage(nil)
		}
	}
	return err
}

func (c *Core) exportImage(metadata [][2]string) (err error) {
	c.stop()
	filename := c.nowFilename("Grid Export", ".png")
	var f *os.File
	if f, err = saveFile(filename, false); err == nil {
		defer func() {
			_ = f.Close()
		}()
		img := imaging.GridImagePaletted(c.gridHolder.grid, imaging.Config{
			CellSize:    c.settings.CellSize,
			Borders:     c.settings.CellBorders,
			AliveColor:  c.settings.CellAliveColor,
			DeadColor:   c.settings.CellDeadColor,
			BorderColor: c.settings.CellBorderColor,
		})
		err = imaging.PngEncode(f, img, metadata)
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
			c.settings.Height, c.settings.Width, c.settings.WrapMode, c.settings.BoundaryMode = grid.Height(), grid.Width(), grid.WrapMode(), grid.BoundaryMode()
			c.gridHolder.replaceGrid(grid)
			c.resetInstrumentation()
			window.Invalidate()
		} else {
			c.resetInstrumentation()
			c.gridHolder.grid.Draw()
			window.Invalidate()
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
	if c.heatMapperType != logic.NoHeatMapper && c.instrumentHeatMap != nil {
		c.statusBar.showHidePopup(popupNone)
		if c.heatMapperType == logic.AllHeatMapper {
			if all, ok := c.instrumentHeatMap.(*logic.AllHeatMapInstrument); ok {
				c.gridHolder.buildHeatMap(all.Specific(c.showingHeatMapType()))
			} else {
				c.gridHolder.buildHeatMap(c.instrumentHeatMap)
			}
		} else {
			c.gridHolder.buildHeatMap(c.instrumentHeatMap)
		}
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

func (c *Core) setInstrumentationHeatMapper(hmt logic.HeatMapperType) {
	c.stop()
	c.heatMapperType = hmt
	c.instrumentHeatMap = c.heatMapperType.New(c.gridHolder.grid, c.settings.HeatMappingHalfLife)
	c.updateInstrumentation()
}

func (c *Core) saveHeatMapImage(metadata [][2]string) {
	c.stop()
	if c.instrumentHeatMap != nil {
		if all, ok := c.instrumentHeatMap.(*logic.AllHeatMapInstrument); ok {
			fs := make([]*os.File, 0)
			defer func() {
				for _, f := range fs {
					_ = f.Close()
				}
			}()
			for hm := range all.List() {
				filename := c.heatMapFilename(hm.Type(), ".png")
				if f, err := saveFile(filename, false); err == nil {
					fs = append(fs, f)
					img := imaging.HeatMap(hm, c.settings.Height, c.settings.Width, imaging.Config{
						CellSize:    c.settings.CellSize,
						Borders:     c.settings.CellBorders,
						AliveColor:  c.settings.CellAliveColor,
						DeadColor:   c.settings.CellDeadColor,
						BorderColor: c.settings.CellBorderColor,
					}, c.settings.HeatMapColors)
					_ = imaging.PngEncode(f, img, formatHeatmapMetadata(metadata, hm.Type()))
				}
			}
		} else {
			filename := c.heatMapFilename(c.heatMapperType, ".png")
			if f, err := saveFile(filename, false); err == nil {
				defer func() {
					_ = f.Close()
				}()
				img := imaging.HeatMap(c.instrumentHeatMap, c.settings.Height, c.settings.Width, imaging.Config{
					CellSize:    c.settings.CellSize,
					Borders:     c.settings.CellBorders,
					AliveColor:  c.settings.CellAliveColor,
					DeadColor:   c.settings.CellDeadColor,
					BorderColor: c.settings.CellBorderColor,
				}, c.settings.HeatMapColors)
				_ = imaging.PngEncode(f, img, formatHeatmapMetadata(metadata, c.heatMapperType))
			}
		}
	}
}

const hmtToken = "%hmt"

func formatHeatmapMetadata(metadata [][2]string, hmt logic.HeatMapperType) [][2]string {
	result := make([][2]string, 0)
	for _, part := range metadata {
		if len(part[0]) > 0 && len(part[1]) > 0 {
			result = append(result, [2]string{part[0], strings.Replace(part[1], hmtToken, hmt.Token(), 1)})
		}
	}
	return result
}

func (c *Core) heatMapFilename(hmt logic.HeatMapperType, extension string) string {
	const pfx = "Heat Map "
	if !c.isShortcutsRunning() {
		return c.nowFilename(pfx+hmt.String(), extension)
	}
	var filename string
	shortcutCurrent := c.getShortcutsCurrent()
	if strings.Contains(shortcutCurrent, hmtToken) {
		filename = strings.ReplaceAll(strings.Replace(shortcutCurrent, hmtToken, hmt.Token(), 1), "%", "") + extension
	} else if strings.HasSuffix(shortcutCurrent, "/") {
		filename = strings.ReplaceAll(shortcutCurrent, "%", "") + pfx + hmt.String() + extension
	} else {
		filename = strings.ReplaceAll(shortcutCurrent, "%", "") + " " + pfx + hmt.String() + extension
	}
	c.addShortcutsCollectFile(filename)
	return filename
}

func (c *Core) saveRepeatDetect() {
	c.stop()
	if c.instrumentRepeat != nil {
		filename := c.nowFilename("Repeat detection", ".json")
		if f, err := saveFile(filename, false); err == nil {
			defer func() {
				_ = f.Close()
			}()
			_ = json.NewEncoder(f).Encode(map[string]any{
				"step":   c.instrumentRepeat.Step,
				"found":  c.instrumentRepeat.Found,
				"first":  c.instrumentRepeat.FirstStep,
				"repeat": c.instrumentRepeat.RepeatStep,
				"period": c.instrumentRepeat.Period,
			})
		}
	}
}

func (c *Core) isRecording() bool {
	return c.instrumentRecord != nil
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
	c.instrumentHeatMap = c.heatMapperType.New(c.gridHolder.grid, c.settings.HeatMappingHalfLife)
	c.updateInstrumentation()
}

func (c *Core) placePattern(gtx layout.Context) {
	if c.gridHolder.overlay != nil {
		c.placePatternRow, c.placePatternCol, c.placePatternRotation = c.gridHolder.overlay.row, c.gridHolder.overlay.col, c.gridHolder.overlay.rotation
	}
	c.gridHolder.placeOverlay()
	c.clearMode()
	c.resetInstrumentation()
	gtx.Execute(op.InvalidateCmd{})
}
