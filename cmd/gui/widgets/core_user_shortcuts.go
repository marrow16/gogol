package widgets

import (
	"fmt"
	"gioui.org/io/key"
	"github.com/marrow16/gogol/logic"
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
	c.window.Invalidate()
	go func() {
		defer func() {
			c.shortcutRunning = false
			c.window.Invalidate()
		}()
		c.runUserShortcut(shortcut)
	}()
	return true
}

func (c *Core) runUserShortcut(shortcut []string) {
	for _, token := range shortcut {
		if !c.shortcutRunning {
			break
		}
		if after, ok := strings.CutPrefix(token, shortcutRepeat); ok {
			if parts := strings.Split(after, ","); len(parts) >= 2 {
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
						c.runUserShortcut(useParts)
					}
				}
			}
			continue
		}
		now := time.Now()
		c.shortcutCurrent = now.Format("2006-01-02 15-04-05") + fmt.Sprintf("-%03d", now.Nanosecond()/1e6)
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
		default:
			if parts := strings.SplitN(token, ":", 2); len(parts) == 2 {
				switch parts[0] {
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
				}
			}
		}
	}
}

const (
	shortcutRepeat           = "repeat:"
	shortcutRun              = "run"
	shortcutStop             = "stop"
	shortcutExport           = "export"
	shortcutClear            = "clear"
	shortcutSnapshot         = "snapshot"
	shortcutUndoToSnapshot   = "undo-to-snapshot"
	shortcutStep             = "step"
	shortcutStepAhead        = "step-ahead"
	shortcutStepAheadDec     = "step-ahead--"
	shortcutStepAheadInc     = "step-ahead++"
	shortcutRandomize        = "randomize"
	shortcutRandomizationDec = "randomization--"
	shortcutRandomizationInc = "randomization++"
	shortcutStepDelayDec     = "step-delay--"
	shortcutStepDelayInc     = "step-delay++"
	shortcutRulePermDec      = "rule-perm--"
	shortcutRulePermInc      = "rule-perm++"
	shortcutSleep            = "sleep"
	shortcutRunRecipe        = "run-recipe"
	shortcutWrapMode         = "wrap-mode"
	shortcutBoundaryMode     = "boundary-mode"
	shortcutStepAheadBy      = "step-ahead-by"
	shortcutRandomization    = "randomization"
	shortcutRandomChanges    = "random-changes"
	shortcutStepDelay        = "step-delay"
	shortcutRulePerm         = "rule-perm"
	shortcutRule             = "rule"
	shortcutBornWith         = "rule-born-with"
	shortcutSurvivesWith     = "rule-survives-with"
	shortcutGridWidth        = "grid-width"
	shortcutGridHeight       = "grid-height"
	shortcutGridSize         = "grid-size" // "widthXheight"
	shortcutRecord           = "record"
	shortcutRepeatDetect     = "repeat-detect"
	shortcutHeatMap          = "heat-map"
	shortcutHeatMapSave      = "heat-map-save"
)
