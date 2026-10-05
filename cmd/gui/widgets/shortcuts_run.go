package widgets

import (
	"errors"
	"fmt"
	"github.com/marrow16/gogol/animator"
	"github.com/marrow16/gogol/cmd/gui/settings"
	"github.com/marrow16/gogol/cmd/gui/shortcuts"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/logic/meta"
	"image/color"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

func (c *Core) startShortcuts(name string) {
	c.shortcutMutex.Lock()
	defer c.shortcutMutex.Unlock()
	c.shortcutStatus = ""
	c.shortcutCollectFiles = false
	c.shortcutRunning = true
	c.shortcutLogger = c.setupShortcutLogger()
	c.shortcutErrorLogger = c.setupShortcutErrorLogger()
}

func (c *Core) stopShortcuts() {
	c.shortcutRunning = false
}

func (c *Core) isShortcutsRunning() bool {
	c.shortcutMutex.RLock()
	defer c.shortcutMutex.RUnlock()
	return c.shortcutRunning
}

func (c *Core) setShortcutsStatus(s string, repeats []int, override bool) {
	c.shortcutMutex.Lock()
	defer c.shortcutMutex.Unlock()
	if override {
		if len(s) > 0 {
			c.shortcutStatus = " " + s
		} else if len(repeats) > 0 {
			c.shortcutStatus = fmt.Sprintf("%v", repeats)
		} else {
			c.shortcutStatus = ""
		}
	} else if len(c.shortcutStatus) == 0 && len(repeats) > 0 {
		c.shortcutStatus = fmt.Sprintf("%v", repeats)
	}
}

func (c *Core) getShortcutsStatus() string {
	c.shortcutMutex.RLock()
	defer c.shortcutMutex.RUnlock()
	return c.shortcutStatus
}

func (c *Core) setShortcutsCurrent(s string) {
	c.shortcutMutex.Lock()
	defer c.shortcutMutex.Unlock()
	c.shortcutCurrent = s
}

func (c *Core) getShortcutsCurrent() string {
	c.shortcutMutex.RLock()
	defer c.shortcutMutex.RUnlock()
	return c.shortcutCurrent
}

func (c *Core) startShortcutsCollectFiles() {
	c.shortcutMutex.Lock()
	defer c.shortcutMutex.Unlock()
	c.shortcutCollectFiles = true
	c.shortcutFiles = make([]string, 0)
	c.shortcutFilesName = c.shortcutCurrent
}

func (c *Core) runShortcut(name string) bool {
	if c.isShortcutsRunning() {
		return false
	}
	lines, ok := c.settings.Shortcuts[name]
	if !ok {
		return false
	}
	c.statusBar.showHidePopup(popupNone)
	c.startShortcuts(name)
	shortcut, errs := shortcuts.ParseShortcut(name, lines)
	window.Invalidate()
	for _, err := range errs {
		c.shortcutErrorLogger.Warn("shortcut parse error", nil, "error", err.Error())
	}
	if len(shortcut.Actions) == 0 {
		c.shortcutErrorLogger.Warn("no actions on shortcut", nil)
	}
	callStack := []string{name}
	go func() {
		defer func() {
			st := c.getShortcutsStatus()
			c.stopShortcuts()
			c.gridHolder.invalidate()
			if st != "" && !strings.HasPrefix(st, "[") {
				c.tempStatus(strings.TrimSpace(st))
			}
			window.Invalidate()
		}()
		c.executeShortcut(callStack, shortcut.Actions, nil, "")
		c.saveShortcutsCollectedFiles()
	}()
	return true
}

func (c *Core) executeShortcut(callStack []string, actions []shortcuts.ActionHolder, repeats []int, nameFmt string) {
	c.setShortcutsStatus("", repeats, false)
	c.shortcutErrorLogger.callStack = callStack
	for _, ah := range actions {
		if !c.isShortcutsRunning() {
			break
		}
		if len(nameFmt) == 0 {
			now := time.Now()
			c.setShortcutsCurrent(now.Format("2006-01-02 15-04-05") + fmt.Sprintf("-%03d", now.Nanosecond()/1e6))
		} else {
			c.setShortcutsCurrent(c.shortcutFormat(nameFmt, repeats))
		}
		switch ah.Action {
		case shortcuts.Repeat:
			if len(ah.Subs) > 0 {
				if nTimes, err := c.shortcutVarIntMin(ah, 0); err == nil {
					for i := range nTimes {
						c.executeShortcut(callStack, ah.Subs, append(append([]int{}, repeats...), i), nameFmt)
					}
				}
			}
		case shortcuts.IterateMetaRule:
			if len(ah.Subs) > 0 {
				name := strings.TrimSpace(c.shortcutVarString(ah))
				if (strings.HasPrefix(name, `"`) && strings.HasSuffix(name, `"`)) || (strings.HasPrefix(name, `'`) && strings.HasSuffix(name, `'`)) {
					name = name[1 : len(name)-1]
				}
				mrs := name
				if named, ok := c.settings.MetaRules[name]; ok {
					mrs = named
				}
				if ev, err := meta.ParseRule(mrs); err == nil {
					for r := range ev.MatchingRules() {
						c.setRule(r)
						c.executeShortcut(callStack, ah.Subs, append(append([]int{}, repeats...), r.Permutation()), nameFmt)
					}
				} else {
					c.shortcutErrorLogger.Error(fmt.Errorf("parse meta rule: %w", err), &ah)
				}
			}
		case shortcuts.IterateCollectedRules:
			perms := make([]int, 0, len(c.settings.CollectedRules))
			for p := range c.settings.CollectedRules {
				perms = append(perms, p)
			}
			slices.Sort(perms)
			for i, perm := range perms {
				if r, err := logic.NewRuleFromPermutation(perm); err == nil {
					c.setRule(r)
					c.executeShortcut(callStack, ah.Subs, append(append([]int{}, repeats...), i), nameFmt)
				}
			}
		case shortcuts.Run:
			c.start()
		case shortcuts.Stop:
			c.stop()
		case shortcuts.Export:
			c.shortcutErrorLogger.Error(c.export(), &ah)
		case shortcuts.ExportImage:
			c.stop()
			if ah.HasOperand {
				c.shortcutErrorLogger.Error(c.exportImage(c.shortcutMetadata(c.shortcutVarString(ah))), &ah)
			} else {
				c.shortcutErrorLogger.Error(c.exportImage(nil), &ah)
			}
		case shortcuts.Clear:
			c.clear()
		case shortcuts.Snapshot:
			c.snapshot()
		case shortcuts.UndoToSnapshot:
			c.undoToSnapshot()
		case shortcuts.ReplaySnapshot:
			c.replaySnapshot()
		case shortcuts.Step:
			c.step()
		case shortcuts.StepBack:
			c.stepBack()
		case shortcuts.StepAhead:
			switch {
			case ah.IsInc:
				c.shortcutSettingChange(func(settings *settings.Settings) {
					if settings.StepAheadBy < 9999 {
						settings.StepAheadBy++
					}
				})
			case ah.IsDec:
				c.shortcutSettingChange(func(settings *settings.Settings) {
					if settings.StepAheadBy > 1 {
						settings.StepAheadBy--
					}
				})
			case ah.HasOperand:
				if n, err := c.shortcutVarIntRange(ah, 1, 9999); err == nil {
					c.shortcutSettingChange(func(settings *settings.Settings) {
						settings.StepAheadBy = n
					})
				}
			default:
				c.stepAhead()
			}
		case shortcuts.Randomize:
			if ah.HasOperand {
				if n, err := c.shortcutVarIntRange(ah, 0, 100); err == nil {
					c.randomize(n)
				}
			} else {
				c.randomize()
			}
		case shortcuts.RandomizePopulation:
			if ah.HasOperand {
				if n, err := c.shortcutVarIntRange(ah, 0, 100); err == nil {
					c.randomizePopulation(n)
				}
			} else {
				c.randomizePopulation()
			}
		case shortcuts.Randomization:
			switch {
			case ah.IsInc:
				c.shortcutSettingChange(func(settings *settings.Settings) {
					if settings.Randomization < 100 {
						settings.Randomization++
					}
				})
			case ah.IsDec:
				c.shortcutSettingChange(func(settings *settings.Settings) {
					if settings.Randomization > 0 {
						settings.Randomization--
					}
				})
			case ah.HasOperand:
				if n, err := c.shortcutVarIntRange(ah, 0, 100); err == nil {
					c.shortcutSettingChange(func(settings *settings.Settings) {
						settings.Randomization = n
					})
				}
			}
		case shortcuts.RandomChanges:
			if ah.HasOperand {
				if n, err := c.shortcutVarIntRange(ah, 0, 100); err == nil {
					c.randomChanges(n)
				}
			} else {
				c.randomChanges()
			}
		case shortcuts.RandomAdditions:
			if ah.HasOperand {
				if n, err := c.shortcutVarIntRange(ah, 0, 100); err == nil {
					c.randomAdditions(n)
				}
			} else {
				c.randomAdditions()
			}
		case shortcuts.RandomCull:
			if ah.HasOperand {
				if n, err := c.shortcutVarIntRange(ah, 0, 100); err == nil {
					c.randomCull(n)
				}
			} else {
				c.randomCull()
			}
		case shortcuts.StepDelay:
			switch {
			case ah.IsInc:
				c.shortcutSettingChange(func(settings *settings.Settings) {
					if settings.StepDelay < 2000 {
						settings.StepDelay++
					}
				})
			case ah.IsDec:
				c.shortcutSettingChange(func(settings *settings.Settings) {
					if settings.StepDelay > 0 {
						settings.StepDelay--
					}
				})
			case ah.HasOperand:
				if n, err := c.shortcutVarIntRange(ah, 0, 2000); err == nil {
					c.shortcutSettingChange(func(settings *settings.Settings) {
						settings.StepDelay = n
					})
				}
			}
		case shortcuts.RulePerm:
			switch {
			case ah.IsInc:
				c.permutationIncrement()
			case ah.IsDec:
				c.permutationDecrement()
			case ah.HasOperand:
				if n, err := c.shortcutVarInt(ah); err == nil {
					if r, err := logic.NewRuleFromPermutation(n); err == nil {
						c.setRule(r)
					} else {
						c.shortcutErrorLogger.Error(err, &ah)
					}
				}
			}
		case shortcuts.RuleInt:
			switch {
			case ah.IsInc:
				c.integerIncrement()
			case ah.IsDec:
				c.integerDecrement()
			case ah.HasOperand:
				if n, err := c.shortcutVarInt(ah); err == nil {
					if r, err := logic.NewRuleFromInteger(n); err == nil {
						c.setRule(r)
					} else {
						c.shortcutErrorLogger.Error(err, &ah)
					}
				}
			}
		case shortcuts.BornWith:
			switch {
			case ah.IsInc:
				c.permutationIncrementBorn()
			case ah.IsDec:
				c.permutationDecrementBorn()
			case ah.HasOperand:
				c.shortcutBornWith(ah)
			}
		case shortcuts.SurvivesWith:
			switch {
			case ah.IsInc:
				c.permutationIncrementSurvives()
			case ah.IsDec:
				c.permutationDecrementSurvives()
			case ah.HasOperand:
				c.shortcutSurvivesWith(ah)
			}
		case shortcuts.RunRecipe:
			if ah.HasOperand {
				c.runRecipe(c.shortcutVarString(ah))
			} else if c.gridRecipes != nil {
				c.stop()
				c.gridRecipes.runRecipe()
			}
		case shortcuts.Record:
			if ah.HasOperand {
				if b, err := c.shortcutVarBool(ah); err == nil {
					c.setInstrumentationRecord(b)
				}
			} else {
				c.setInstrumentationRecord(true)
			}
		case shortcuts.RepeatDetect:
			if ah.HasOperand {
				if b, err := c.shortcutVarBool(ah); err == nil {
					c.setInstrumentationRepeat(b)
				}
			} else {
				c.setInstrumentationRepeat(true)
			}
		case shortcuts.HeatMap:
			if ah.HasOperand {
				hmt := logic.HeatMapperTypeFrom(c.shortcutVarString(ah))
				c.setInstrumentationHeatMapper(hmt)
			} else {
				c.setInstrumentationHeatMapper(logic.ActivityHeatMapper)
			}
		case shortcuts.HeatMapSave:
			c.stop()
			var errs []error
			if ah.HasOperand {
				errs = c.saveHeatMapImage(c.shortcutMetadata(c.shortcutVarString(ah)))
			} else {
				errs = c.saveHeatMapImage(nil)
			}
			for _, err := range errs {
				c.shortcutErrorLogger.Error(err, &ah)
			}
		case shortcuts.HeatMapReveal:
			c.stop()
			if c.heatMapperType != logic.NoHeatMapper && c.instrumentHeatMap != nil {
				c.gridHolder.buildHeatMap(c.instrumentHeatMap)
				c.mode = heatMapMode
				window.Invalidate()
			}
		case shortcuts.RepeatDetectSave:
			c.stop()
			c.shortcutErrorLogger.Error(c.saveRepeatDetect(), &ah)
		case shortcuts.AnimationSave:
			c.stop()
			recorder := c.instrumentRecord
			if recorder != nil && recorder.FramesCount() > 1 {
				var filename string
				var err error
				if c.settings.AnimationFormat != "mp4" {
					filename, err = resolveSavePath(c.nowFilename("Grid", ".gif"))
				} else {
					filename, err = resolveSavePath(c.nowFilename("Grid", ".mp4"))
				}
				if err == nil {
					ani := animator.NewAnimator(c.settings.CellSize, c.settings.CellAliveColor, c.settings.CellDeadColor, c.settings.CellBorderColor, c.settings.CellBorders, c.settings.AnimationFormat)
					_ = ani.Animate(filename, recorder)
				}
			}
		case shortcuts.CollectFiles:
			c.startShortcutsCollectFiles()
		case shortcuts.AddCollectedRule:
			c.shortcutSettingChange(func(settings *settings.Settings) {
				settings.CollectedRules[c.gridHolder.grid.Rule().Permutation()] = true
			})
		case shortcuts.RemoveCollectedRule:
			c.shortcutSettingChange(func(settings *settings.Settings) {
				delete(settings.CollectedRules, c.gridHolder.grid.Rule().Permutation())
			})
		case shortcuts.PreviousCollectedRule:
			c.shortcutCollectedRuleMove(false)
		case shortcuts.NextCollectedRule:
			c.shortcutCollectedRuleMove(true)
		case shortcuts.Call:
			subName := c.shortcutVarString(ah)
			if lines, ok := c.settings.Shortcuts[subName]; ok {
				sub, errs := shortcuts.ParseShortcut(subName, lines)
				for _, err := range errs {
					c.shortcutErrorLogger.Warn("shortcut parse error", &ah, "call", subName, "error", err.Error())
				}
				if len(sub.Actions) == 0 {
					c.shortcutErrorLogger.Warn("no actions on called shortcut", &ah, "call", subName)
				} else {
					ncs := append([]string{}, callStack...)
					c.executeShortcut(append(ncs, subName), sub.Actions, repeats, nameFmt)
				}
			} else {
				c.shortcutErrorLogger.Error(fmt.Errorf("unknown shortcut %q", subName), &ah)
			}
		case shortcuts.If:
			if c.shortcutConditionEvaluate(ah.Operand, repeats) {
				if len(ah.Subs) > 0 {
					c.executeShortcut(callStack, ah.Subs, repeats, nameFmt)
				}
			}
		case shortcuts.While:
			if len(ah.Subs) > 0 {
				i := 0
				for c.shortcutConditionEvaluate(ah.Operand, repeats) {
					c.executeShortcut(callStack, ah.Subs, append(append([]int{}, repeats...), i), nameFmt)
					i++
				}
			}
		case shortcuts.BreakIf:
			if c.shortcutConditionEvaluate(ah.Operand, repeats) {
				c.stop()
				return
			}
		case shortcuts.StopIf:
			if c.shortcutConditionEvaluate(ah.Operand, repeats) {
				c.stop()
				c.stopShortcuts()
				return
			}
		case shortcuts.Variable:
			c.shortcutVariable(ah, repeats)
		case shortcuts.ClearVariable:
			c.shortcutSettingChange(func(settings *settings.Settings) {
				c.settings.ShortcutVariables.Delete(ah.Operand)
			})
		case shortcuts.ClearVariables:
			if ah.HasOperand {
				c.shortcutSettingChange(func(settings *settings.Settings) {
					for name := range strings.SplitSeq(ah.Operand, ";") {
						c.settings.ShortcutVariables.Delete(strings.TrimPrefix(name, "$"))
					}
				})
			} else {
				c.shortcutSettingChange(func(settings *settings.Settings) {
					c.settings.ShortcutVariables.DeleteAll()
				})
			}
		case shortcuts.Name:
			nameFmt += ah.Operand
		case shortcuts.NameReset:
			nameFmt = ah.Operand
		case shortcuts.StepAheadBy:
			if n, err := c.shortcutVarIntMin(ah, 0); err == nil {
				c.stepAheadBy(n)
			}
		case shortcuts.StepBackBy:
			if n, err := c.shortcutVarIntMin(ah, 0); err == nil {
				c.skipBackBy(n)
			}
		case shortcuts.MaxAdjacents:
			if n, err := c.shortcutVarIntRange(ah, 0, 8); err == nil {
				c.maximumAdjacents(n)
			}
		case shortcuts.Rule:
			if r, ok := logic.Rules[c.shortcutVarString(ah)]; ok {
				c.setRule(r)
			} else if r, err := logic.NewRuleRle("", c.shortcutVarString(ah)); err == nil {
				c.setRule(r)
			} else {
				c.shortcutErrorLogger.Error(err, &ah)
			}
		case shortcuts.WrapMode:
			if wm := logic.WrapModeFromString(c.shortcutVarString(ah), -1); wm != -1 {
				c.setWrapMode(wm)
			} else {
				c.shortcutErrorLogger.Error(errors.New("invalid/unkown wrap mode"), &ah)
			}
		case shortcuts.BoundaryMode:
			if bm := logic.BoundaryModeFromString(c.shortcutVarString(ah), -1); bm != -1 {
				c.setBoundaryMode(bm)
			} else {
				c.shortcutErrorLogger.Error(errors.New("invalid/unkown boundary mode"), &ah)
			}
		case shortcuts.Sleep:
			if ah.HasOperand {
				if n, err := c.shortcutVarIntMin(ah, 0); err == nil && n > 0 {
					time.Sleep(time.Duration(n) * time.Millisecond)
				}
			} else {
				time.Sleep(1 * time.Second)
			}
		case shortcuts.GridSize:
			if dims := strings.Split(strings.ToLower(c.shortcutVarString(ah)), "x"); len(dims) == 2 {
				if wd, err := strconv.Atoi(dims[0]); err == nil && wd > 2 && wd <= 99999 {
					if ht, err := strconv.Atoi(dims[1]); err == nil && ht > 2 && ht <= 99999 {
						c.gridResize(ht, wd)
					} else if err != nil {
						c.shortcutErrorLogger.Error(err, &ah)
					} else {
						c.shortcutErrorLogger.Error(errors.New("value is out of range (3-99999)"), &ah)
					}
				} else if err != nil {
					c.shortcutErrorLogger.Error(err, &ah)
				} else {
					c.shortcutErrorLogger.Error(errors.New("value is out of range (3-99999)"), &ah)
				}
			} else {
				c.shortcutErrorLogger.Error(errors.New("invalid dimensions (must be wXh)"), &ah)
			}
		case shortcuts.GridHeight:
			if n, err := c.shortcutVarIntRange(ah, 3, 99999); err == nil {
				c.gridResize(n, c.settings.Width)
			}
		case shortcuts.GridWidth:
			if n, err := c.shortcutVarIntRange(ah, 3, 99999); err == nil {
				c.gridResize(c.settings.Height, n)
			}
		case shortcuts.HeatMapColors:
			clrs := make([]color.NRGBA, 0)
			for s := range strings.SplitSeq(ah.Operand, ";") {
				if clr, ok := parseColor(c.shortcutResolveVar(s, ah)); ok {
					clrs = append(clrs, clr)
				}
			}
			if len(clrs) >= 2 {
				c.shortcutSettingChange(func(settings *settings.Settings) {
					settings.HeatMapColors = clrs
				})
			} else {
				c.shortcutSettingChange(func(settings *settings.Settings) {
					settings.HeatMapColors = nil
				})
			}
		case shortcuts.NextMetaRule:
			mrs := c.shortcutVarString(ah)
			if strings.HasPrefix(mrs, `"`) && strings.HasSuffix(mrs, `"`) && len(mrs) > 2 {
				name := mrs[1 : len(mrs)-1]
				if named, ok := c.settings.MetaRules[name]; ok {
					mrs = named
				}
			}
			if mr, err := meta.ParseRule(mrs); err == nil {
				curr := c.gridHolder.grid.Rule().Permutation()
				if next := mr.Next(curr); next != curr {
					if r, err := logic.NewRuleFromPermutation(next); err == nil {
						c.setRule(r)
					}
				} else if next = mr.Next(-1); next != -1 {
					if r, err := logic.NewRuleFromPermutation(next); err == nil {
						c.setRule(r)
					}
				}
			} else {
				c.shortcutErrorLogger.Error(fmt.Errorf("parse meta rule: %w", err), &ah)
			}
		case shortcuts.PreviousMetaRule:
			mrs := c.shortcutVarString(ah)
			if strings.HasPrefix(mrs, `"`) && strings.HasSuffix(mrs, `"`) && len(mrs) > 2 {
				name := mrs[1 : len(mrs)-1]
				if named, ok := c.settings.MetaRules[name]; ok {
					mrs = named
				}
			}
			if mr, err := meta.ParseRule(mrs); err == nil {
				curr := c.gridHolder.grid.Rule().Permutation()
				if prev := mr.Previous(curr); prev != curr {
					if r, err := logic.NewRuleFromPermutation(prev); err == nil {
						c.setRule(r)
					}
				} else if prev = mr.Previous(-1); prev != -1 {
					if r, err := logic.NewRuleFromPermutation(prev); err == nil {
						c.setRule(r)
					}
				}
			} else {
				c.shortcutErrorLogger.Error(fmt.Errorf("parse meta rule: %w", err), &ah)
			}
		case shortcuts.Log:
			c.shortcutLog(ah.Operand, repeats)
		case shortcuts.Status:
			if ah.HasOperand {
				c.setShortcutsStatus(c.shortcutFormat(ah.Operand, repeats), nil, true)
			} else {
				c.setShortcutsStatus("", repeats, true)
			}
		case shortcuts.Borders:
			if b, err := c.shortcutVarBool(ah); err == nil {
				c.setCellBorders(b)
			}
		case shortcuts.CellSize:
			if n, err := c.shortcutVarIntRange(ah, 1, 100); err == nil {
				c.setCellSize(n)
			}
		case shortcuts.CellColorAlive:
			if clr, ok := parseColor(c.shortcutVarString(ah)); ok {
				c.stop()
				c.shortcutSettingChange(func(settings *settings.Settings) {
					settings.CellAliveColor = clr
				})
				c.gridHolder.grid.Draw()
			}
		case shortcuts.CellColorDead:
			if clr, ok := parseColor(c.shortcutVarString(ah)); ok {
				c.stop()
				c.shortcutSettingChange(func(settings *settings.Settings) {
					settings.CellDeadColor = clr
				})
				c.gridHolder.grid.Draw()
			}
		case shortcuts.CellColorBorder:
			if clr, ok := parseColor(c.shortcutVarString(ah)); ok {
				c.stop()
				c.shortcutSettingChange(func(settings *settings.Settings) {
					settings.CellBorderColor = clr
				})
				c.gridHolder.grid.Draw()
			}
		case shortcuts.AnimationFormat:
			if strings.ToLower(c.shortcutVarString(ah)) == "mp4" && animator.Mp4Available() {
				c.shortcutSettingChange(func(settings *settings.Settings) {
					settings.AnimationFormat = "mp4"
				})
			} else {
				c.shortcutSettingChange(func(settings *settings.Settings) {
					settings.AnimationFormat = "gif"
				})
			}
		}
	}
}

func (c *Core) addShortcutsCollectFile(filename string) {
	c.shortcutMutex.Lock()
	defer c.shortcutMutex.Unlock()
	if c.shortcutCollectFiles {
		c.shortcutFiles = append(c.shortcutFiles, filename)
	}
}

func (c *Core) saveShortcutsCollectedFiles() {
	c.shortcutMutex.Lock()
	defer c.shortcutMutex.Unlock()
	if c.shortcutCollectFiles && len(c.shortcutFiles) > 0 {
		curr := c.shortcutFilesName
		c.shortcutCurrent = curr
		if f, err := saveFile(c.nowFilename("files", ".txt"), true); err == nil {
			for _, filename := range c.shortcutFiles {
				_, _ = f.WriteString(strings.TrimPrefix(filename, curr) + "\n")
			}
			_ = f.Close()
		}
	}
}

func (c *Core) shortcutSettingChange(fn func(settings *settings.Settings)) {
	c.shortcutMutex.Lock()
	defer c.shortcutMutex.Unlock()
	fn(c.settings)
}

func (c *Core) shortcutBornWith(ah shortcuts.ActionHolder) {
	bw, sw := c.gridHolder.grid.Rule().BornWith(), c.gridHolder.grid.Rule().SurvivesWith()
	s := c.shortcutVarString(ah)
	if after, ok := strings.CutPrefix(s, "|"); ok {
		var sb strings.Builder
		for i := range 9 {
			ch := rune(i + 48)
			if strings.ContainsRune(after, ch) || strings.ContainsRune(bw, ch) {
				sb.WriteRune(ch)
			}
		}
		bw = sb.String()
	} else if after, ok = strings.CutPrefix(s, "&"); ok {
		var sb strings.Builder
		for i := range 9 {
			ch := rune(i + 48)
			if strings.ContainsRune(after, ch) && strings.ContainsRune(bw, ch) {
				sb.WriteRune(ch)
			}
		}
		bw = sb.String()
	} else if after, ok = strings.CutPrefix(s, "!"); ok {
		var sb strings.Builder
		if len(after) == 0 {
			for i := range 9 {
				ch := rune(i + 48)
				if !strings.ContainsRune(bw, ch) {
					sb.WriteRune(ch)
				}
			}
		} else {
			for i := range 9 {
				ch := rune(i + 48)
				if strings.ContainsRune(after, ch) {
					if !strings.ContainsRune(bw, ch) {
						sb.WriteRune(ch)
					}
				} else if strings.ContainsRune(bw, ch) {
					sb.WriteRune(ch)
				}
			}
		}
		bw = sb.String()
	} else {
		bw = s
	}
	if r, err := logic.NewRuleRle("", "B"+bw+"/S"+sw); err == nil {
		c.setRule(r)
	} else {
		c.shortcutErrorLogger.Error(err, &ah)
	}
}

func (c *Core) shortcutSurvivesWith(ah shortcuts.ActionHolder) {
	bw, sw := c.gridHolder.grid.Rule().BornWith(), c.gridHolder.grid.Rule().SurvivesWith()
	s := c.shortcutVarString(ah)
	if after, ok := strings.CutPrefix(s, "|"); ok {
		var sb strings.Builder
		for i := range 9 {
			ch := rune(i + 48)
			if strings.ContainsRune(after, ch) || strings.ContainsRune(sw, ch) {
				sb.WriteRune(ch)
			}
		}
		sw = sb.String()
	} else if after, ok = strings.CutPrefix(s, "&"); ok {
		var sb strings.Builder
		for i := range 9 {
			ch := rune(i + 48)
			if strings.ContainsRune(after, ch) && strings.ContainsRune(sw, ch) {
				sb.WriteRune(ch)
			}
		}
		sw = sb.String()
	} else if after, ok = strings.CutPrefix(s, "!"); ok {
		var sb strings.Builder
		if len(after) == 0 {
			for i := range 9 {
				ch := rune(i + 48)
				if !strings.ContainsRune(sw, ch) {
					sb.WriteRune(ch)
				}
			}
		} else {
			for i := range 9 {
				ch := rune(i + 48)
				if strings.ContainsRune(after, ch) {
					if !strings.ContainsRune(sw, ch) {
						sb.WriteRune(ch)
					}
				} else if strings.ContainsRune(sw, ch) {
					sb.WriteRune(ch)
				}
			}
		}
		sw = sb.String()
	} else {
		sw = s
	}
	if r, err := logic.NewRuleRle("", "B"+bw+"/S"+sw); err == nil {
		c.setRule(r)
	} else {
		c.shortcutErrorLogger.Error(err, &ah)
	}
}

func (c *Core) shortcutCollectedRuleMove(inc bool) {
	if len(c.settings.CollectedRules) == 0 {
		return
	}
	fr := make([]int, 0, len(c.settings.CollectedRules))
	for i := range c.settings.CollectedRules {
		fr = append(fr, i)
	}
	if len(fr) == 1 {
		if r, err := logic.NewRuleFromPermutation(fr[0]); err == nil {
			c.setRule(r)
		}
	}
	slices.Sort(fr)
	idx, found := slices.BinarySearch(fr, c.gridHolder.grid.Rule().Permutation())
	if !found && idx == 0 {
		idx = -1
	}
	if inc {
		idx++
	} else {
		idx--
	}
	if idx >= len(fr) {
		idx = 0
	} else if idx < 0 {
		idx = len(fr) - 1
	}
	if r, err := logic.NewRuleFromPermutation(fr[idx]); err == nil {
		c.setRule(r)
	}
}

func (c *Core) shortcutMetadata(s string) [][2]string {
	result := make([][2]string, 0)
	for item := range strings.SplitSeq(s, ";") {
		if parts := strings.SplitN(item, "=", 2); len(parts) == 2 && len(parts[0]) > 0 && len(parts[1]) > 0 {
			result = append(result, [2]string{parts[0], c.shortcutFormat(parts[1], nil)})
		}
	}
	return result
}

var colorRegex = regexp.MustCompile("^#[0-9a-fA-F]{6}$")

func parseColor(s string) (c color.NRGBA, ok bool) {
	if !colorRegex.MatchString(s) {
		return c, false
	}
	r, _ := strconv.ParseUint(s[1:3], 16, 8)
	g, _ := strconv.ParseUint(s[3:5], 16, 8)
	b, _ := strconv.ParseUint(s[5:7], 16, 8)
	return color.NRGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 0xff,
	}, true
}
