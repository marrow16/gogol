package widgets

import (
	"fmt"
	"github.com/go-andiamo/splitter"
	"github.com/marrow16/gogol/animator"
	"github.com/marrow16/gogol/cmd/gui/settings"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/logic/meta"
	"image/color"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

func (c *Core) startShortcuts() {
	c.shortcutMutex.Lock()
	defer c.shortcutMutex.Unlock()
	c.shortcutStatus = ""
	c.shortcutCollectFiles = false
	c.shortcutRunning = true
}

func (c *Core) stopShortcuts() {
	c.shortcutRunning = false
}

func (c *Core) isShortcutsRunning() bool {
	c.shortcutMutex.RLock()
	defer c.shortcutMutex.RUnlock()
	return c.shortcutRunning
}

func (c *Core) setShortcutsStatus(s string) {
	c.shortcutMutex.Lock()
	defer c.shortcutMutex.Unlock()
	c.shortcutStatus = s
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

func (c *Core) runShortcut(name string) bool {
	if c.isShortcutsRunning() {
		return false
	}
	shortcut, ok := c.settings.Shortcuts[name]
	if !ok {
		return false
	}
	c.statusBar.showHidePopup(popupNone)
	c.startShortcuts()
	window.Invalidate()
	go func() {
		defer func() {
			c.stopShortcuts()
			c.gridHolder.invalidate()
			window.Invalidate()
		}()
		c.runUserShortcut(shortcut, nil, "")
		c.saveShortcutsCollectedFiles()
	}()
	return true
}

var shortcutCommasSplitter = splitter.MustCreateSplitter(',', splitter.DoubleQuotes, splitter.SingleQuotes).
	AddDefaultOptions(splitter.IgnoreEmpties)

func (c *Core) runUserShortcut(shortcut []string, repeats []int, nameFmt string) {
	if len(repeats) > 0 {
		c.setShortcutsStatus(fmt.Sprintf("%v", repeats))
	}
	for _, token := range shortcut {
		if !c.isShortcutsRunning() {
			break
		}
		if after, ok := strings.CutPrefix(token, shortcutRepeat); ok {
			if parts, err := shortcutCommasSplitter.Split(after); err == nil && len(parts) >= 2 {
				if nTimes, err := c.shortcutVarInt(parts[0]); err == nil && nTimes > 0 {
					parts = parts[1:]
					useParts := make([]string, 0, len(parts))
					for i := 0; i < len(parts); i++ {
						if strings.HasPrefix(parts[i], shortcutRepeat) {
							useParts = append(useParts, strings.Join(parts[i:], ","))
							break
						} else {
							useParts = append(useParts, parts[i])
						}
					}
					for i := range nTimes {
						c.runUserShortcut(useParts, append(repeats, i), nameFmt)
					}
				}
			}
			continue
		} else if after, ok = strings.CutPrefix(token, shortcutIterateMetaRule); ok {
			if parts, err := shortcutCommasSplitter.Split(after); err == nil && len(parts) >= 2 {
				name := strings.TrimSpace(c.shortcutVarString(parts[0]))
				if (strings.HasPrefix(name, `"`) && strings.HasSuffix(name, `"`)) || (strings.HasPrefix(name, `'`) && strings.HasSuffix(name, `'`)) {
					name = name[1 : len(name)-1]
				}
				mrs := name
				if named, ok := c.settings.MetaRules[name]; ok {
					mrs = named
				}
				if ev, err := meta.ParseRule(mrs); err == nil {
					parts = parts[1:]
					useParts := make([]string, 0, len(parts))
					for i := 0; i < len(parts); i++ {
						if strings.HasPrefix(parts[i], shortcutRepeat) {
							useParts = append(useParts, strings.Join(parts[i:], ","))
							break
						} else {
							useParts = append(useParts, parts[i])
						}
					}
					i := 0
					for r := range ev.MatchingRules() {
						c.setRule(r)
						c.runUserShortcut(useParts, append(repeats, i), nameFmt)
						i++
					}
				} else {
					return
				}
			} else {
				return
			}
			continue
		} else if after, ok = strings.CutPrefix(token, shortcutIterateCollectedRules); ok {
			if parts, err := shortcutCommasSplitter.Split(after); err == nil && len(parts) >= 1 {
				useParts := make([]string, 0, len(parts))
				for i := range parts {
					if strings.HasPrefix(parts[i], shortcutRepeat) {
						useParts = append(useParts, strings.Join(parts[i:], ","))
						break
					} else {
						useParts = append(useParts, parts[i])
					}
				}
				perms := make([]int, 0, len(c.settings.CollectedRules))
				for p := range c.settings.CollectedRules {
					perms = append(perms, p)
				}
				slices.Sort(perms)
				for i, perm := range perms {
					if r, err := logic.NewRuleFromPermutation(perm); err == nil {
						c.setRule(r)
						c.runUserShortcut(useParts, append(repeats, i), nameFmt)
					}
				}
			}
		}
		if len(nameFmt) == 0 {
			now := time.Now()
			c.setShortcutsCurrent(now.Format("2006-01-02 15-04-05") + fmt.Sprintf("-%03d", now.Nanosecond()/1e6))
		} else {
			c.setShortcutsCurrent(c.shortcutFormat(nameFmt, repeats))
		}
		switch token {
		case shortcutRun:
			c.start()
		case shortcutStop:
			c.stop()
		case shortcutExport:
			_ = c.export()
		case shortcutExportImage:
			c.stop()
			_ = c.exportImage(nil)
		case shortcutClear:
			c.clear()
		case shortcutSnapshot:
			c.snapshot()
		case shortcutUndoToSnapshot:
			c.undoToSnapshot()
		case shortcutReplaySnapshot:
			c.replaySnapshot()
		case shortcutStep:
			c.step()
		case shortcutStepAhead:
			c.stepAhead()
		case shortcutStepAheadDec:
			c.shortcutSettingChange(func(settings *settings.Settings) {
				if settings.StepAheadBy > 1 {
					settings.StepAheadBy--
				}
			})
		case shortcutStepAheadInc:
			c.shortcutSettingChange(func(settings *settings.Settings) {
				if settings.StepAheadBy < 9999 {
					settings.StepAheadBy++
				}
			})
		case shortcutRandomize:
			c.randomize()
		case shortcutRandomizePopulation:
			c.randomizePopulation()
		case shortcutRandomizationDec:
			c.shortcutSettingChange(func(settings *settings.Settings) {
				if settings.Randomization > 0 {
					settings.Randomization--
				}
			})
		case shortcutRandomizationInc:
			c.shortcutSettingChange(func(settings *settings.Settings) {
				if settings.Randomization < 100 {
					settings.Randomization++
				}
			})
		case shortcutRandomChanges:
			c.randomChanges()
		case shortcutRandomAdditions:
			c.randomAdditions()
		case shortcutRandomCull:
			c.randomCull()
		case shortcutStepDelayDec:
			c.shortcutSettingChange(func(settings *settings.Settings) {
				if settings.StepDelay > 0 {
					settings.StepDelay--
				}
			})
		case shortcutStepDelayInc:
			c.shortcutSettingChange(func(settings *settings.Settings) {
				if settings.StepDelay < 2000 {
					settings.StepDelay++
				}
			})
		case shortcutRulePermDec:
			c.permutationDecrement()
		case shortcutRulePermInc:
			c.permutationIncrement()
		case shortcutRuleIntDec:
			c.integerDecrement()
		case shortcutRuleIntInc:
			c.integerIncrement()
		case shortcutBornWithInc:
			c.permutationIncrementBorn()
		case shortcutBornWithDec:
			c.permutationDecrementBorn()
		case shortcutSurvivesWithInc:
			c.permutationIncrementSurvives()
		case shortcutSurvivesWithDec:
			c.permutationDecrementSurvives()
		case shortcutRunRecipe:
			if c.gridRecipes != nil {
				c.stop()
				c.gridRecipes.runRecipe()
			}
		case shortcutRecord:
			c.setInstrumentationRecord(true)
		case shortcutRepeatDetect:
			c.setInstrumentationRepeat(true)
		case shortcutHeatMap:
			c.setInstrumentationHeatMapper(logic.ActivityHeatMapper)
		case shortcutHeatMapSave:
			c.stop()
			c.saveHeatMapImage(nil)
		case shortcutHeatMapReveal:
			c.stop()
			if c.heatMapperType != logic.NoHeatMapper && c.instrumentHeatMap != nil {
				c.gridHolder.buildHeatMap(c.instrumentHeatMap)
				c.mode = heatMapMode
				window.Invalidate()
			}
		case shortcutRepeatDetectSave:
			c.stop()
			c.saveRepeatDetect()
		case shortcutAnimationSave:
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
		case shortcutFiles:
			c.startShortcutsCollectFiles()
		case shortcutAddCollectedRule:
			c.shortcutSettingChange(func(settings *settings.Settings) {
				settings.CollectedRules[c.gridHolder.grid.Rule().Permutation()] = true
			})
		case shortcutRemoveCollectedRule:
			c.shortcutSettingChange(func(settings *settings.Settings) {
				delete(settings.CollectedRules, c.gridHolder.grid.Rule().Permutation())
			})
		case shortcutPreviousCollectedRule:
			c.shortcutCollectedRuleMove(false)
		case shortcutNextCollectedRule:
			c.shortcutCollectedRuleMove(true)
		default:
			if parts := strings.SplitN(token, ":", 2); len(parts) == 2 {
				switch parts[0] {
				case shortcutBreakIf:
					if c.shortcutStopIf(parts[1], repeats) {
						c.stop()
						return
					}
				case shortcutStopIf:
					if c.shortcutStopIf(parts[1], repeats) {
						c.stop()
						c.stopShortcuts()
						return
					}
				case shortcutCall, shortcutCallShort:
					if sub, ok := c.settings.Shortcuts[c.shortcutVarString(parts[1])]; ok {
						c.runUserShortcut(sub, repeats, nameFmt)
					}
				case shortcutVariable:
					if kv := strings.SplitN(parts[1], "=", 2); len(kv) == 2 {
						switch {
						case strings.HasSuffix(kv[0], "+"):
							vn := strings.TrimRight(strings.TrimSuffix(kv[0], "+"), " ")
							vv := kv[1]
							c.shortcutSettingChange(func(settings *settings.Settings) {
								if v, ok := c.settings.ShortcutVariables[vn]; ok {
									done := false
									if nv, err := strconv.Atoi(v); err == nil {
										if opand, err := strconv.Atoi(c.shortcutFormat(vv, repeats)); err == nil {
											c.settings.ShortcutVariables[vn] = strconv.Itoa(nv + opand)
											done = true
										}
									}
									if !done {
										// string concat...
										c.settings.ShortcutVariables[vn] = v + c.shortcutFormat(vv, repeats)
									}
								}
							})
						case strings.HasSuffix(kv[0], "-"):
							vn := strings.TrimRight(strings.TrimSuffix(kv[0], "-"), " ")
							vv := strings.TrimSpace(kv[1])
							c.shortcutSettingChange(func(settings *settings.Settings) {
								if v, ok := c.settings.ShortcutVariables[vn]; ok {
									if nv, err := strconv.Atoi(v); err == nil {
										if opand, err := strconv.Atoi(c.shortcutFormat(vv, repeats)); err == nil {
											c.settings.ShortcutVariables[vn] = strconv.Itoa(nv + opand)
										}
									}
								}
							})
						case strings.HasSuffix(kv[0], "*"):
							vn := strings.TrimRight(strings.TrimSuffix(kv[0], "*"), " ")
							vv := strings.TrimSpace(kv[1])
							c.shortcutSettingChange(func(settings *settings.Settings) {
								if v, ok := c.settings.ShortcutVariables[vn]; ok {
									if nv, err := strconv.Atoi(v); err == nil {
										if opand, err := strconv.Atoi(c.shortcutFormat(vv, repeats)); err == nil {
											c.settings.ShortcutVariables[vn] = strconv.Itoa(nv * opand)
										}
									}
								}
							})
						case strings.HasSuffix(kv[0], "/"):
							vn := strings.TrimRight(strings.TrimSuffix(kv[0], "/"), " ")
							vv := strings.TrimSpace(kv[1])
							c.shortcutSettingChange(func(settings *settings.Settings) {
								if v, ok := c.settings.ShortcutVariables[vn]; ok {
									if nv, err := strconv.Atoi(v); err == nil {
										if opand, err := strconv.Atoi(c.shortcutFormat(vv, repeats)); err == nil {
											c.settings.ShortcutVariables[vn] = strconv.Itoa(nv / opand)
										}
									}
								}
							})
						case strings.HasSuffix(kv[0], "&"):
							vn := strings.TrimRight(strings.TrimSuffix(kv[0], "&"), " ")
							vv := strings.TrimSpace(kv[1])
							c.shortcutSettingChange(func(settings *settings.Settings) {
								if v, ok := c.settings.ShortcutVariables[vn]; ok {
									if nv, err := strconv.Atoi(v); err == nil {
										if opand, err := strconv.Atoi(c.shortcutFormat(vv, repeats)); err == nil {
											c.settings.ShortcutVariables[vn] = strconv.Itoa(nv & opand)
										}
									}
								}
							})
						case strings.HasSuffix(kv[0], "|"):
							vn := strings.TrimRight(strings.TrimSuffix(kv[0], "|"), " ")
							vv := strings.TrimSpace(kv[1])
							c.shortcutSettingChange(func(settings *settings.Settings) {
								if v, ok := c.settings.ShortcutVariables[vn]; ok {
									if nv, err := strconv.Atoi(v); err == nil {
										if opand, err := strconv.Atoi(c.shortcutFormat(vv, repeats)); err == nil {
											c.settings.ShortcutVariables[vn] = strconv.Itoa(nv | opand)
										}
									}
								}
							})
						case strings.HasSuffix(kv[0], "^"):
							vn := strings.TrimRight(strings.TrimSuffix(kv[0], "^"), " ")
							vv := strings.TrimSpace(kv[1])
							c.shortcutSettingChange(func(settings *settings.Settings) {
								if v, ok := c.settings.ShortcutVariables[vn]; ok {
									if nv, err := strconv.Atoi(v); err == nil {
										if opand, err := strconv.Atoi(c.shortcutFormat(vv, repeats)); err == nil {
											c.settings.ShortcutVariables[vn] = strconv.Itoa(nv ^ opand)
										}
									}
								}
							})
						default:
							vn := strings.TrimRight(kv[0], " ")
							vv := kv[1]
							c.shortcutSettingChange(func(settings *settings.Settings) {
								c.settings.ShortcutVariables[vn] = c.shortcutFormat(vv, repeats)
							})
						}
					} else if vn, ok := strings.CutSuffix(parts[1], "++"); ok {
						c.shortcutSettingChange(func(settings *settings.Settings) {
							if v, ok := c.settings.ShortcutVariables[vn]; ok {
								if nv, err := strconv.Atoi(v); err == nil {
									c.settings.ShortcutVariables[vn] = strconv.Itoa(nv + 1)
								}
							}
						})
					} else if vn, ok = strings.CutSuffix(parts[1], "--"); ok {
						c.shortcutSettingChange(func(settings *settings.Settings) {
							if v, ok := c.settings.ShortcutVariables[vn]; ok {
								if nv, err := strconv.Atoi(v); err == nil {
									c.settings.ShortcutVariables[vn] = strconv.Itoa(nv - 1)
								}
							}
						})
					} else {
						c.shortcutSettingChange(func(settings *settings.Settings) {
							delete(c.settings.ShortcutVariables, parts[1])
						})
					}
				case shortcutName:
					nameFmt += parts[1]
				case "-" + shortcutName:
					nameFmt = parts[1]
				case shortcutStepAhead:
					if n, err := c.shortcutVarInt(parts[1]); err == nil && n > 0 && n <= 9999 {
						c.shortcutSettingChange(func(settings *settings.Settings) {
							settings.StepAheadBy = n
						})
					}
				case shortcutStepAheadBy:
					if n, err := c.shortcutVarInt(parts[1]); err == nil && n > 0 {
						c.stepAheadBy(n)
					}
				case shortcutStepBackBy:
					if n, err := c.shortcutVarInt(parts[1]); err == nil && n > 0 {
						c.skipBackBy(n)
					}
				case shortcutRandomization:
					if n, err := c.shortcutVarInt(parts[1]); err == nil && n >= 0 && n <= 100 {
						c.shortcutSettingChange(func(settings *settings.Settings) {
							settings.Randomization = n
						})
					}
				case shortcutRandomize:
					if n, err := c.shortcutVarInt(parts[1]); err == nil && n >= 0 && n <= 100 {
						c.randomize(n)
					}
				case shortcutRandomizePopulation:
					if n, err := c.shortcutVarInt(parts[1]); err == nil && n >= 0 && n <= 100 {
						c.randomizePopulation(n)
					}
				case shortcutRandomChanges:
					if n, err := c.shortcutVarInt(parts[1]); err == nil && n >= 0 && n <= 100 {
						c.randomChanges(n)
					}
				case shortcutRandomAdditions:
					if n, err := c.shortcutVarInt(parts[1]); err == nil && n >= 0 && n <= 100 {
						c.randomAdditions(n)
					}
				case shortcutRandomCull:
					if n, err := c.shortcutVarInt(parts[1]); err == nil && n >= 0 && n <= 100 {
						c.randomCull(n)
					}
				case shortcutMaxAdjacents:
					if n, err := c.shortcutVarInt(parts[1]); err == nil && n >= 0 && n <= 8 {
						c.maximumAdjacents(n)
					}
				case shortcutStepDelay:
					if n, err := c.shortcutVarInt(parts[1]); err == nil && n >= 0 && n <= 2000 {
						c.shortcutSettingChange(func(settings *settings.Settings) {
							settings.StepDelay = n
						})
					}
				case shortcutRule:
					if r, ok := logic.Rules[c.shortcutVarString(parts[1])]; ok {
						c.setRule(r)
					} else if r, err := logic.NewRuleRle("", c.shortcutVarString(parts[1])); err == nil {
						c.setRule(r)
					}
				case shortcutRulePerm:
					if n, err := c.shortcutVarInt(parts[1]); err == nil {
						if r, err := logic.NewRuleFromPermutation(n); err == nil {
							c.setRule(r)
						}
					}
				case shortcutRuleInt:
					if n, err := c.shortcutVarInt(parts[1]); err == nil {
						if r, err := logic.NewRuleFromInteger(n); err == nil {
							c.setRule(r)
						}
					}
				case shortcutWrapMode:
					if wm := logic.WrapModeFromString(c.shortcutVarString(parts[1]), -1); wm != -1 {
						c.setWrapMode(wm)
					}
				case shortcutBoundaryMode:
					if bm := logic.BoundaryModeFromString(c.shortcutVarString(parts[1]), -1); bm != -1 {
						c.setBoundaryMode(bm)
					}
				case shortcutSleep:
					if n, err := c.shortcutVarInt(parts[1]); err == nil {
						time.Sleep(time.Duration(n) * time.Millisecond)
					} else {
						time.Sleep(1 * time.Second)
					}
				case shortcutRunRecipe:
					c.runRecipe(c.shortcutVarString(parts[1]))
				case shortcutGridSize:
					if dims := strings.Split(strings.ToLower(c.shortcutVarString(parts[1])), "x"); len(dims) == 2 {
						if wd, err := strconv.Atoi(dims[0]); err == nil && wd > 0 && wd <= 1000 {
							if ht, err := strconv.Atoi(dims[1]); err == nil && ht > 0 && ht <= 1000 {
								c.gridResize(ht, wd)
							}
						}
					}
				case shortcutGridHeight:
					if n, err := c.shortcutVarInt(parts[1]); err == nil && n > 2 && n <= 1000 {
						c.gridResize(n, c.settings.Width)
					}
				case shortcutGridWidth:
					if n, err := c.shortcutVarInt(parts[1]); err == nil && n > 2 && n <= 1000 {
						c.gridResize(c.settings.Height, n)
					}
				case shortcutRecord:
					if b, err := c.shortcutVarBool(parts[1]); err == nil {
						c.setInstrumentationRecord(b)
					}
				case shortcutRepeatDetect:
					if b, err := c.shortcutVarBool(parts[1]); err == nil {
						c.setInstrumentationRepeat(b)
					}
				case shortcutHeatMapSave:
					c.stop()
					c.saveHeatMapImage(c.shortcutMetadata(c.shortcutVarString(parts[1])))
				case shortcutExportImage:
					c.stop()
					_ = c.exportImage(c.shortcutMetadata(c.shortcutVarString(parts[1])))
				case shortcutHeatMap:
					hmt := logic.HeatMapperTypeFrom(c.shortcutVarString(parts[1]))
					c.setInstrumentationHeatMapper(hmt)
				case shortcutHeatMapColors:
					clrs := make([]color.NRGBA, 0)
					for s := range strings.SplitSeq(c.shortcutVarString(parts[1]), ";") {
						if clr, ok := parseColor(c.shortcutVarString(s)); ok {
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
				case shortcutBornWith:
					bw, sw := c.gridHolder.grid.Rule().BornWith(), c.gridHolder.grid.Rule().SurvivesWith()
					if after, ok := strings.CutPrefix(c.shortcutVarString(parts[1]), "|"); ok {
						var sb strings.Builder
						for i := range 9 {
							ch := rune(i + 48)
							if strings.ContainsRune(after, ch) || strings.ContainsRune(bw, ch) {
								sb.WriteRune(ch)
							}
						}
						bw = sb.String()
					} else if after, ok = strings.CutPrefix(c.shortcutVarString(parts[1]), "&"); ok {
						var sb strings.Builder
						for i := range 9 {
							ch := rune(i + 48)
							if strings.ContainsRune(after, ch) && strings.ContainsRune(bw, ch) {
								sb.WriteRune(ch)
							}
						}
						bw = sb.String()
					} else if after, ok = strings.CutPrefix(c.shortcutVarString(parts[1]), "!"); ok {
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
						bw = c.shortcutVarString(parts[1])
					}
					if r, err := logic.NewRuleRle("", "B"+bw+"/S"+sw); err == nil {
						c.setRule(r)
					}
				case shortcutSurvivesWith:
					bw, sw := c.gridHolder.grid.Rule().BornWith(), c.gridHolder.grid.Rule().SurvivesWith()
					if after, ok := strings.CutPrefix(c.shortcutVarString(parts[1]), "|"); ok {
						var sb strings.Builder
						for i := range 9 {
							ch := rune(i + 48)
							if strings.ContainsRune(after, ch) || strings.ContainsRune(sw, ch) {
								sb.WriteRune(ch)
							}
						}
						sw = sb.String()
					} else if after, ok = strings.CutPrefix(c.shortcutVarString(parts[1]), "&"); ok {
						var sb strings.Builder
						for i := range 9 {
							ch := rune(i + 48)
							if strings.ContainsRune(after, ch) && strings.ContainsRune(sw, ch) {
								sb.WriteRune(ch)
							}
						}
						sw = sb.String()
					} else if after, ok = strings.CutPrefix(c.shortcutVarString(parts[1]), "!"); ok {
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
						sw = c.shortcutVarString(parts[1])
					}
					if r, err := logic.NewRuleRle("", "B"+bw+"/S"+sw); err == nil {
						c.setRule(r)
					}
				case shortcutNextMetaRule:
					mrs := c.shortcutVarString(parts[1])
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
					}
				case shortcutPreviousMetaRule:
					mrs := c.shortcutVarString(parts[1])
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
					}
				case shortcutLog:
					c.shortcutLog(parts[1], repeats)
				case shortcutBorders:
					if b, err := c.shortcutVarBool(parts[1]); err == nil {
						c.setCellBorders(b)
					}
				case shortcutCellSize:
					if n, err := c.shortcutVarInt(parts[1]); err == nil && n > 0 && n <= 100 {
						c.setCellSize(n)
					}
				case shortcutCellColorAlive:
					if clr, ok := parseColor(c.shortcutVarString(parts[1])); ok {
						c.stop()
						c.shortcutSettingChange(func(settings *settings.Settings) {
							settings.CellAliveColor = clr
						})
						c.gridHolder.grid.Draw()
					}
				case shortcutCellColorDead:
					if clr, ok := parseColor(c.shortcutVarString(parts[1])); ok {
						c.stop()
						c.shortcutSettingChange(func(settings *settings.Settings) {
							settings.CellDeadColor = clr
						})
						c.gridHolder.grid.Draw()
					}
				case shortcutCellColorBorder:
					if clr, ok := parseColor(c.shortcutVarString(parts[1])); ok {
						c.stop()
						c.shortcutSettingChange(func(settings *settings.Settings) {
							settings.CellBorderColor = clr
						})
						c.gridHolder.grid.Draw()
					}
				case shortcutAnimationFormat:
					if strings.ToLower(c.shortcutVarString(parts[1])) == "mp4" && animator.Mp4Available() {
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
	}
}

func (c *Core) shortcutVarString(s string) string {
	if name, ok := strings.CutPrefix(s, "$"); ok {
		if v, ok := c.settings.ShortcutVariables[name]; ok {
			return v
		}
	}
	return s
}

func (c *Core) shortcutVarInt(s string) (int, error) {
	if name, ok := strings.CutPrefix(s, "$"); ok {
		if v, ok := c.settings.ShortcutVariables[name]; ok {
			s = v
		}
	}
	return strconv.Atoi(s)
}

func (c *Core) shortcutVarBool(s string) (bool, error) {
	if name, ok := strings.CutPrefix(s, "$"); ok {
		if v, ok := c.settings.ShortcutVariables[name]; ok {
			s = v
		}
	}
	return strconv.ParseBool(s)
}

func (c *Core) shortcutStopIf(condition string, repeats []int) bool {
	switch {
	case condition == "rule-collected":
		return c.settings.CollectedRules[c.gridHolder.grid.Rule().Permutation()]
	case strings.Contains(condition, "=="):
		return c.shortcutCompare(equals, strings.SplitN(condition, "==", 2), repeats)
	case strings.Contains(condition, "!="):
		return c.shortcutCompare(notEquals, strings.SplitN(condition, "!=", 2), repeats)
	case strings.Contains(condition, ">="):
		return c.shortcutCompare(greaterThanOrEqual, strings.SplitN(condition, ">=", 2), repeats)
	case strings.Contains(condition, "<="):
		return c.shortcutCompare(lessThanOrEqual, strings.SplitN(condition, "<=", 2), repeats)
	case strings.Contains(condition, ">"):
		return c.shortcutCompare(greaterThan, strings.SplitN(condition, ">", 2), repeats)
	case strings.Contains(condition, "<"):
		return c.shortcutCompare(lessThan, strings.SplitN(condition, "<", 2), repeats)
	}
	return false
}

type comparator int

const (
	equals comparator = iota
	notEquals
	greaterThan
	lessThan
	greaterThanOrEqual
	lessThanOrEqual
)

func (c *Core) shortcutCompare(comp comparator, parts []string, repeats []int) bool {
	if len(parts) != 2 {
		return false
	}
	vl := c.shortcutFormat(strings.TrimSpace(parts[0]), repeats)
	vr := c.shortcutFormat(strings.TrimSpace(parts[1]), repeats)
	if nl, err := strconv.Atoi(vl); err == nil {
		if nr, err := strconv.Atoi(vr); err == nil {
			switch comp {
			case equals:
				return nl == nr
			case notEquals:
				return nl != nr
			case greaterThan:
				return nl > nr
			case lessThan:
				return nl < nr
			case greaterThanOrEqual:
				return nl >= nr
			case lessThanOrEqual:
				return nl <= nr
			}
		}
	}
	sc := strings.Compare(strings.ToLower(vl), strings.ToLower(vr))
	switch comp {
	case equals:
		return sc == 0
	case notEquals:
		return sc != 0
	case greaterThan:
		return sc > 0
	case lessThan:
		return sc < 0
	case greaterThanOrEqual:
		return sc >= 0
	case lessThanOrEqual:
		return sc <= 0
	default:
		return false
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

func (c *Core) shortcutLog(msgf string, repeats []int) {
	if fp, err := resolveSavePath("./output.log"); err == nil {
		if f, err := os.OpenFile(fp, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			defer func() {
				_ = f.Close()
			}()
			if (strings.HasPrefix(msgf, `"`) && strings.HasSuffix(msgf, `"`)) || (strings.HasPrefix(msgf, `'`) && strings.HasSuffix(msgf, `'`)) {
				msgf = msgf[1 : len(msgf)-1]
			}
			msg := c.shortcutFormat(msgf, repeats) + "\n"
			_, _ = f.WriteString(msg)
		}
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

func (c *Core) shortcutFormat(s string, repeats []int) string {
	rIndex := 0
	runes := []rune(s)
	l := len(runes)
	isEscaped := func(i int, r rune) bool {
		if i+1 < l {
			return runes[i+1] == r
		}
		return true
	}
	hasPrefix := func(i int, prefix string) bool {
		p := []rune(prefix)
		return l >= i+1+len(p) && slices.Equal(runes[i+1:i+1+len(p)], p)
	}
	_ = hasPrefix
	var b strings.Builder
	b.Grow(l)
	// sort current vars - longest names first...
	chkVars := make([][2]string, 0)
	for k, v := range c.settings.ShortcutVariables {
		chkVars = append(chkVars, [2]string{k, v})
	}
	slices.SortFunc(chkVars, func(a, b [2]string) int {
		return len(b[0]) - len(a[0])
	})
	now := time.Now()
	for i := 0; i < l; {
		char := runes[i]
		switch char {
		case '%':
			if isEscaped(i, char) {
				b.WriteRune(char)
				i += 2
				continue
			}
			token := shortcutToken(-1)
			tokenLen := 0
			for _, tp := range shortcutFormatTokens {
				if hasPrefix(i, tp.string) {
					token = tp.token
					tokenLen = len(tp.string)
					break
				}
			}
			if tokenLen == 0 {
				b.WriteRune(char)
				i++
				continue
			}
			i += tokenLen
			b.WriteString(c.shortcutToken(token, repeats, rIndex, now))
			if token == tokenIteration {
				rIndex++
			}
		case '$':
			if isEscaped(i, char) {
				b.WriteRune(char)
				i += 2
				continue
			}
			// check for $var-name...
			found := false
			for _, v := range chkVars {
				if hasPrefix(i, v[0]) {
					i += len(v[0])
					b.WriteString(v[1])
					found = true
					break
				}
			}
			if !found {
				b.WriteRune(char)
			}
		default:
			b.WriteRune(char)
		}
		i++
	}
	return b.String()
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

const (
	shortcutRepeat                = "repeat:"
	shortcutName                  = "name"
	shortcutFiles                 = "collect-files"
	shortcutRun                   = "run"
	shortcutStop                  = "stop"
	shortcutExport                = "export"
	shortcutExportImage           = "export-image"
	shortcutClear                 = "clear"
	shortcutSnapshot              = "snapshot"
	shortcutUndoToSnapshot        = "undo-to-snapshot"
	shortcutReplaySnapshot        = "replay-snapshot"
	shortcutStep                  = "step"
	shortcutStepAhead             = "step-ahead"
	shortcutStepAheadDec          = "step-ahead--"
	shortcutStepAheadInc          = "step-ahead++"
	shortcutRandomize             = "randomize"
	shortcutRandomizationDec      = "randomization--"
	shortcutRandomizationInc      = "randomization++"
	shortcutRandomization         = "randomization"
	shortcutRandomChanges         = "random-changes"
	shortcutRandomAdditions       = "random-additions"
	shortcutRandomCull            = "random-cull"
	shortcutRandomizePopulation   = "randomize-population"
	shortcutMaxAdjacents          = "max-adjacents"
	shortcutStepDelayDec          = "step-delay--"
	shortcutStepDelayInc          = "step-delay++"
	shortcutRulePermDec           = "rule-perm--"
	shortcutRulePermInc           = "rule-perm++"
	shortcutRuleIntDec            = "rule-int--"
	shortcutRuleIntInc            = "rule-int++"
	shortcutSleep                 = "sleep"
	shortcutRunRecipe             = "run-recipe"
	shortcutWrapMode              = "wrap-mode"
	shortcutBoundaryMode          = "boundary-mode"
	shortcutStepAheadBy           = "step-ahead-by"
	shortcutStepBackBy            = "step-back-by"
	shortcutStepDelay             = "step-delay"
	shortcutRulePerm              = "rule-perm"
	shortcutRuleInt               = "rule-int"
	shortcutRule                  = "rule"
	shortcutBornWith              = "rule-born-with"
	shortcutBornWithInc           = "rule-born-with++"
	shortcutBornWithDec           = "rule-born-with--"
	shortcutSurvivesWith          = "rule-survives-with"
	shortcutSurvivesWithInc       = "rule-survives-with++"
	shortcutSurvivesWithDec       = "rule-survives-with--"
	shortcutGridWidth             = "grid-width"
	shortcutGridHeight            = "grid-height"
	shortcutGridSize              = "grid-size" // "widthXheight"
	shortcutRecord                = "record"
	shortcutRepeatDetect          = "repeat-detect"
	shortcutRepeatDetectSave      = "repeat-detect-save"
	shortcutHeatMap               = "heat-map"
	shortcutHeatMapColors         = "heat-map-colors"
	shortcutHeatMapSave           = "heat-map-save"
	shortcutHeatMapReveal         = "heat-map-reveal"
	shortcutNextMetaRule          = "next-meta-rule"
	shortcutPreviousMetaRule      = "previous-meta-rule"
	shortcutIterateMetaRule       = "iterate-meta-rule:"
	shortcutLog                   = "log"
	shortcutAddCollectedRule      = "add-collected-rule"
	shortcutRemoveCollectedRule   = "remove-collected-rule"
	shortcutPreviousCollectedRule = "previous-collected-rule"
	shortcutNextCollectedRule     = "next-collected-rule"
	shortcutIterateCollectedRules = "iterate-collected-rules"
	shortcutBorders               = "borders"
	shortcutCellSize              = "cell-size"
	shortcutCellColorAlive        = "cell-color-alive"
	shortcutCellColorDead         = "cell-color-dead"
	shortcutCellColorBorder       = "cell-color-border"
	shortcutAnimationSave         = "record-animation-save"
	shortcutAnimationFormat       = "record-animation-format"
	shortcutBreakIf               = "break-if"
	shortcutStopIf                = "stop-if"
	shortcutCall                  = "call-shortcut"
	shortcutCallShort             = "call"
	shortcutVariable              = "variable"
)
