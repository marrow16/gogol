package widgets

import (
	"fmt"
	"gioui.org/io/key"
	"github.com/go-andiamo/splitter"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/logic/meta"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

func (c *Core) stopShortcuts() {
	c.shortcutRunning = false
}

func (c *Core) userShortcutKeys(kn key.Name) bool {
	if c.shortcutRunning {
		return false
	}
	shortcut, handled := c.settings.Shortcuts[string(kn)]
	if !handled {
		return false
	}
	c.shortcutRunning = true
	c.shortcutCollectFiles = false
	c.shortcutStatus.Store(nil)
	c.window.Invalidate()
	go func() {
		defer func() {
			c.shortcutRunning = false
			c.window.Invalidate()
		}()
		c.runUserShortcut(shortcut, nil, "")
		if c.shortcutCollectFiles && len(c.shortcutFiles) > 0 {
			c.shortcutCurrent = c.shortcutFilesName
			if f, err := saveFile(c.nowFilename("files", ".txt"), true); err == nil {
				for _, filename := range c.shortcutFiles {
					_, _ = f.WriteString(strings.TrimPrefix(filename, c.shortcutCurrent) + "\n")
				}
				_ = f.Close()
			}
		}
	}()
	return true
}

var shortcutCommasSplitter = splitter.MustCreateSplitter(',',
	splitter.DoubleQuotes, splitter.SingleQuotes)

func (c *Core) runUserShortcut(shortcut []string, repeats []int, nameFmt string) {
	if len(repeats) > 0 {
		st := fmt.Sprintf("%v", repeats)
		c.shortcutStatus.Store(&st)
	}
	for _, token := range shortcut {
		if !c.shortcutRunning {
			break
		}
		if after, ok := strings.CutPrefix(token, shortcutRepeat); ok {
			if parts, err := shortcutCommasSplitter.Split(after); err == nil && len(parts) >= 2 {
				if nTimes, err := strconv.Atoi(parts[0]); err == nil && nTimes > 0 {
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
					for i := 0; i < nTimes; i++ {
						c.runUserShortcut(useParts, append(repeats, i), nameFmt)
					}
				}
			}
			continue
		} else if after, ok = strings.CutPrefix(token, shortcutIterateMetaRule); ok {
			if parts, err := shortcutCommasSplitter.Split(after); err == nil && len(parts) >= 2 {
				name := strings.TrimSpace(parts[0])
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
		}
		if len(nameFmt) == 0 {
			now := time.Now()
			c.shortcutCurrent = now.Format("2006-01-02 15-04-05") + fmt.Sprintf("-%03d", now.Nanosecond()/1e6)
		} else {
			c.shortcutCurrent = c.shortcutFormatName(nameFmt, repeats)
		}
		switch token {
		case shortcutRun:
			c.start()
		case shortcutStop:
			c.stop()
		case shortcutExport:
			_ = c.export()
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
			if c.settings.StepAheadBy > 1 {
				c.settings.StepAheadBy--
			}
		case shortcutStepAheadInc:
			if c.settings.StepAheadBy < 9999 {
				c.settings.StepAheadBy++
			}
		case shortcutRandomize:
			c.randomize()
		case shortcutRandomizePopulation:
			c.randomizePopulation()
		case shortcutRandomizationDec:
			if c.settings.Randomization > 0 {
				c.settings.Randomization--
			}
		case shortcutRandomizationInc:
			if c.settings.Randomization < 100 {
				c.settings.Randomization++
			}
		case shortcutRandomChanges:
			c.randomChanges()
		case shortcutStepDelayDec:
			if c.settings.StepDelay > 0 {
				c.settings.StepDelay--
			}
		case shortcutStepDelayInc:
			if c.settings.StepDelay < 2000 {
				c.settings.StepDelay++
			}
		case shortcutRulePermDec:
			c.permutationDecrement()
		case shortcutRulePermInc:
			c.permutationIncrement()
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
			c.setInstrumentationHeatMapper(activityHeatMapper)
		case shortcutHeatMapSave:
			c.stop()
			c.saveHeatMapImage()
		case shortcutHeatMapReveal:
			c.stop()
			if c.heatMapperType != noHeatMapper && c.instrumentHeatMap != nil {
				c.gridHolder.buildHeatMap(c.instrumentHeatMap)
				c.mode = heatMapMode
				c.window.Invalidate()
			}
		case shortcutRepeatDetectSave:
			c.stop()
			c.saveRepeatDetect()
		case shortcutFiles:
			c.shortcutCollectFiles = true
			c.shortcutFiles = make([]string, 0)
			c.shortcutFilesName = c.shortcutCurrent
		case shortcutAddCollectedRule:
			c.settings.CollectedRules[c.gridHolder.grid.Rule.Permutation()] = true
		case shortcutRemoveCollectedRule:
			delete(c.settings.CollectedRules, c.gridHolder.grid.Rule.Permutation())
		case shortcutPreviousCollectedRule:
			c.shortcutCollectedRuleMove(false)
		case shortcutNextCollectedRule:
			c.shortcutCollectedRuleMove(true)
		default:
			if parts := strings.SplitN(token, ":", 2); len(parts) == 2 {
				switch parts[0] {
				case shortcutName:
					nameFmt += parts[1]
				case shortcutStepAhead:
					if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 && n <= 9999 {
						c.settings.StepAheadBy = n
					}
				case shortcutStepAheadBy:
					if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 {
						c.stepAheadBy(n)
					}
				case shortcutRandomization:
					if n, err := strconv.Atoi(parts[1]); err == nil && n >= 0 && n <= 100 {
						c.settings.Randomization = n
					}
				case shortcutMaxAdjacents:
					if n, err := strconv.Atoi(parts[1]); err == nil && n >= 0 && n <= 8 {
						c.maximumAdjacents(n)
					}
				case shortcutStepDelay:
					if n, err := strconv.Atoi(parts[1]); err == nil && n >= 0 && n <= 2000 {
						c.settings.StepDelay = n
					}
				case shortcutRule:
					if r, ok := logic.Rules[parts[1]]; ok {
						c.setRule(r)
					} else if r, err := logic.NewRuleRle("", parts[1]); err == nil {
						c.setRule(r)
					}
				case shortcutRulePerm:
					if n, err := strconv.Atoi(parts[1]); err == nil {
						if r, err := logic.NewRuleFromPermutation(n); err == nil {
							c.setRule(r)
						}
					}
				case shortcutWrapMode:
					if wm := logic.WrapModeFromString(parts[1], -1); wm != -1 {
						c.settings.WrapMode = wm
					}
				case shortcutBoundaryMode:
					if bm := logic.BoundaryModeFromString(parts[1], -1); bm != -1 {
						c.settings.BoundaryMode = bm
					}
				case shortcutSleep:
					if n, err := strconv.Atoi(parts[1]); err == nil {
						time.Sleep(time.Duration(n) * time.Millisecond)
					} else {
						time.Sleep(1 * time.Second)
					}
				case shortcutRunRecipe:
					c.runRecipe(parts[1])
				case shortcutGridSize:
					if dims := strings.Split(strings.ToLower(parts[1]), "x"); len(dims) == 2 {
						if wd, err := strconv.Atoi(dims[0]); err == nil && wd > 0 && wd <= 1000 {
							if ht, err := strconv.Atoi(dims[1]); err == nil && ht > 0 && ht <= 1000 {
								c.gridResize(ht, wd)
							}
						}
					}
				case shortcutGridHeight:
					if n, err := strconv.Atoi(parts[1]); err == nil && n > 2 && n <= 1000 {
						c.gridResize(n, c.settings.Width)
					}
				case shortcutGridWidth:
					if n, err := strconv.Atoi(parts[1]); err == nil && n > 2 && n <= 1000 {
						c.gridResize(c.settings.Height, n)
					}
				case shortcutRecord:
					if b, err := strconv.ParseBool(parts[1]); err == nil {
						c.setInstrumentationRecord(b)
					}
				case shortcutRepeatDetect:
					if b, err := strconv.ParseBool(parts[1]); err == nil {
						c.setInstrumentationRepeat(b)
					}
				case shortcutHeatMap:
					hmt := heatMapperTypeFrom(parts[1])
					c.setInstrumentationHeatMapper(hmt)
				case shortcutBornWith:
					bw, sw := c.gridHolder.grid.Rule.BornWith(), c.gridHolder.grid.Rule.SurvivesWith()
					if after, ok := strings.CutPrefix(parts[1], "|"); ok {
						var sb strings.Builder
						for i := 0; i < 9; i++ {
							ch := rune(i + 48)
							if strings.ContainsRune(after, ch) || strings.ContainsRune(bw, ch) {
								sb.WriteRune(ch)
							}
						}
						bw = sb.String()
					} else if after, ok = strings.CutPrefix(parts[1], "&"); ok {
						var sb strings.Builder
						for i := 0; i < 9; i++ {
							ch := rune(i + 48)
							if strings.ContainsRune(after, ch) && strings.ContainsRune(bw, ch) {
								sb.WriteRune(ch)
							}
						}
						bw = sb.String()
					} else if after, ok = strings.CutPrefix(parts[1], "!"); ok {
						var sb strings.Builder
						if len(after) == 0 {
							for i := 0; i < 9; i++ {
								ch := rune(i + 48)
								if !strings.ContainsRune(bw, ch) {
									sb.WriteRune(ch)
								}
							}
						} else {
							for i := 0; i < 9; i++ {
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
						bw = parts[1]
					}
					if r, err := logic.NewRuleRle("", "B"+bw+"/S"+sw); err == nil {
						c.setRule(r)
					}
				case shortcutSurvivesWith:
					bw, sw := c.gridHolder.grid.Rule.BornWith(), c.gridHolder.grid.Rule.SurvivesWith()
					if after, ok := strings.CutPrefix(parts[1], "|"); ok {
						var sb strings.Builder
						for i := 0; i < 9; i++ {
							ch := rune(i + 48)
							if strings.ContainsRune(after, ch) || strings.ContainsRune(sw, ch) {
								sb.WriteRune(ch)
							}
						}
						sw = sb.String()
					} else if after, ok = strings.CutPrefix(parts[1], "&"); ok {
						var sb strings.Builder
						for i := 0; i < 9; i++ {
							ch := rune(i + 48)
							if strings.ContainsRune(after, ch) && strings.ContainsRune(sw, ch) {
								sb.WriteRune(ch)
							}
						}
						sw = sb.String()
					} else if after, ok = strings.CutPrefix(parts[1], "!"); ok {
						var sb strings.Builder
						if len(after) == 0 {
							for i := 0; i < 9; i++ {
								ch := rune(i + 48)
								if !strings.ContainsRune(sw, ch) {
									sb.WriteRune(ch)
								}
							}
						} else {
							for i := 0; i < 9; i++ {
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
						sw = parts[1]
					}
					if r, err := logic.NewRuleRle("", "B"+bw+"/S"+sw); err == nil {
						c.setRule(r)
					}
				case shortcutNextMetaRule:
					mrs := parts[1]
					if strings.HasPrefix(mrs, `"`) && strings.HasSuffix(mrs, `"`) && len(mrs) > 2 {
						name := mrs[1 : len(mrs)-1]
						if named, ok := c.settings.MetaRules[name]; ok {
							mrs = named
						}
					}
					if mr, err := meta.ParseRule(mrs); err == nil {
						curr := c.gridHolder.grid.Rule.Permutation()
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
					mrs := parts[1]
					if strings.HasPrefix(mrs, `"`) && strings.HasSuffix(mrs, `"`) && len(mrs) > 2 {
						name := mrs[1 : len(mrs)-1]
						if named, ok := c.settings.MetaRules[name]; ok {
							mrs = named
						}
					}
					if mr, err := meta.ParseRule(mrs); err == nil {
						curr := c.gridHolder.grid.Rule.Permutation()
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
					c.shortcutLog(parts[1])
				}
			}
		}
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
	idx, found := slices.BinarySearch(fr, c.gridHolder.grid.Rule.Permutation())
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

func (c *Core) shortcutLog(msgf string) {
	if fp, err := resolveSavePath("./output.log"); err == nil {
		if f, err := os.OpenFile(fp, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			defer f.Close()
			if (strings.HasPrefix(msgf, `"`) && strings.HasSuffix(msgf, `"`)) || (strings.HasPrefix(msgf, `'`) && strings.HasSuffix(msgf, `'`)) {
				msgf = msgf[1 : len(msgf)-1]
			}
			msg := c.shortcutFormatName(msgf, nil) + "\n"
			_, _ = f.WriteString(msg)
		}
	}
}

func (c *Core) shortcutFormatName(s string, repeats []int) string {
	rIndex := 0
	var b strings.Builder
	rule := c.gridHolder.grid.Rule
	for i := 0; i < len(s); {
		switch {
		case strings.HasPrefix(s[i:], "%rule"):
			b.WriteString(rule.Rle())
			i += 5
		case strings.HasPrefix(s[i:], "%born"):
			b.WriteString("B" + rule.BornWith())
			i += 5
		case strings.HasPrefix(s[i:], "%survives"):
			b.WriteString("S" + rule.SurvivesWith())
			i += 9
		case strings.HasPrefix(s[i:], "%perm"):
			b.WriteString(strconv.Itoa(rule.Permutation()))
			i += 5
		case strings.HasPrefix(s[i:], "%rand"):
			b.WriteString(strconv.Itoa(c.settings.Randomization))
			i += 5
		case strings.HasPrefix(s[i:], "%now"):
			now := time.Now()
			b.WriteString(now.Format("2006-01-02 15-04-05") + fmt.Sprintf("-%03d", now.Nanosecond()/1e6))
			i += 4
		case strings.HasPrefix(s[i:], "%timestamp"):
			b.WriteString(time.Now().Format(time.RFC3339))
			i += 10
		case strings.HasPrefix(s[i:], "%step"):
			step := c.gridHolder.grid.StepCount.Load()
			b.WriteString(strconv.FormatUint(step, 10))
			i += 5
		case strings.HasPrefix(s[i:], "%repeat-found"):
			if c.instrumentRepeat != nil {
				b.WriteString(strconv.FormatBool(c.instrumentRepeat.Found))
			} else {
				b.WriteString("_")
			}
			i += 13
		case strings.HasPrefix(s[i:], "%repeat-first"):
			if c.instrumentRepeat != nil {
				b.WriteString(strconv.FormatUint(c.instrumentRepeat.FirstStep, 10))
			} else {
				b.WriteString("_")
			}
			i += 13
		case strings.HasPrefix(s[i:], "%repeat-at"):
			if c.instrumentRepeat != nil {
				b.WriteString(strconv.FormatUint(c.instrumentRepeat.RepeatStep, 10))
			} else {
				b.WriteString("_")
			}
			i += 10
		case strings.HasPrefix(s[i:], "%repeat-period"):
			if c.instrumentRepeat != nil {
				b.WriteString(strconv.FormatUint(c.instrumentRepeat.Period, 10))
			} else {
				b.WriteString("_")
			}
			i += 14
		case strings.HasPrefix(s[i:], "%population"):
			b.WriteString(strconv.Itoa(c.population()))
			i += 11
		case strings.HasPrefix(s[i:], "%R"):
			if len(repeats) > 0 {
				b.WriteString(strconv.Itoa(repeats[len(repeats)-1]))
			}
			i += 2
		case strings.HasPrefix(s[i:], "%r"):
			if rIndex < len(repeats) {
				b.WriteString(strconv.Itoa(repeats[rIndex]))
			} else {
				b.WriteString("_r_")
			}
			rIndex++
			i += 2
		case s[i] == '%':
			i++
		default:
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
}

const (
	shortcutRepeat                = "repeat:"
	shortcutName                  = "name"
	shortcutFiles                 = "collect-files"
	shortcutRun                   = "run"
	shortcutStop                  = "stop"
	shortcutExport                = "export"
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
	shortcutRandomizePopulation   = "randomize-population"
	shortcutMaxAdjacents          = "max-adjacents"
	shortcutStepDelayDec          = "step-delay--"
	shortcutStepDelayInc          = "step-delay++"
	shortcutRulePermDec           = "rule-perm--"
	shortcutRulePermInc           = "rule-perm++"
	shortcutSleep                 = "sleep"
	shortcutRunRecipe             = "run-recipe"
	shortcutWrapMode              = "wrap-mode"
	shortcutBoundaryMode          = "boundary-mode"
	shortcutStepAheadBy           = "step-ahead-by"
	shortcutStepDelay             = "step-delay"
	shortcutRulePerm              = "rule-perm"
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
)
