package main

import (
	tea "charm.land/bubbletea/v2"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/recipes"
	"slices"
	"strconv"
	"strings"
	"time"
)

type shortcutResult struct {
	redraw      bool
	displayMode displayMode
	savePrefs   bool
}

func (m *model) runShortcut(shortcut []string) tea.Cmd {
	if len(shortcut) == 0 {
		return nil
	}
	tokens := slices.Clone(shortcut)
	m.mode = modeShortcut
	return func() tea.Msg {
		result := shortcutResult{displayMode: -1}
		for _, token := range tokens {
			switch token {
			case "snapshot":
				m.takeSnapshot()
			case "step":
				m.grid.Step()
			case "step-ahead":
				m.grid.StepAhead(m.stepAheadBy)
			case "step-ahead--":
				if m.stepAheadBy > 2 {
					result.savePrefs = true
					m.stepAheadBy--
					m.prefs.StepAheadBy = m.stepAheadBy
				}
			case "step-ahead++":
				if m.stepAheadBy < 9999 {
					result.savePrefs = true
					m.stepAheadBy++
					m.prefs.StepAheadBy = m.stepAheadBy
				}
			case "randomize":
				m.grid.Randomize(m.random)
			case "randomization--":
				if m.random > 0 {
					result.savePrefs = true
					m.random--
					m.prefs.Random = m.random
				}
			case "randomization++":
				if m.random < 100 {
					result.savePrefs = true
					m.random++
					m.prefs.Random = m.random
				}
			case "step-delay--":
				if m.stepDelay > 1 {
					result.savePrefs = true
					m.stepDelay--
					m.prefs.StepDelay = m.stepDelay
				}
			case "step-delay++":
				if m.stepDelay < 2000 {
					result.savePrefs = true
					m.stepDelay++
					m.prefs.StepDelay = m.stepDelay
				}
			case "rule-perm--":
				if n := m.grid.Rule.Permutation(); n > 0 {
					if r, err := logic.NewRuleFromPermutation(n - 1); err == nil {
						result.savePrefs = true
						m.grid.Rule = r
						m.prefs.Rule = r.Rle()
					}
				}
			case "rule-perm++":
				if n := m.grid.Rule.Permutation(); n < (1<<18)-1 {
					if r, err := logic.NewRuleFromPermutation(n + 1); err == nil {
						result.savePrefs = true
						m.grid.Rule = r
						m.prefs.Rule = r.Rle()
					}
				}
			case "export":
				_, _ = m.exportGrid()
			case "clear":
				m.grid.Clear()
			case "run":
				result.displayMode = modeRunning
			case "settings":
				result.displayMode = modeSettings
			case "patterns":
				result.displayMode = modePatterns
			case "recipes":
				result.displayMode = modeRecipes
			case "capture":
				result.displayMode = modeCapture
			case "edit":
				result.displayMode = modeEditing
			case "help":
				result.displayMode = modeSplash
			case "place":
				if m.patterns.currentPattern != nil {
					m.patterns.currentPattern.Draw(m.grid, m.patterns.patternPlaceY, m.patterns.patternPlaceX, m.patterns.patternRotate)
				}
			case "run-recipe":
				_ = m.recipes.runSelected()
			default:
				if parts := strings.SplitN(token, ":", 2); len(parts) == 2 {
					switch parts[0] {
					case "sleep":
						if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 && n <= 2000 {
							time.Sleep(time.Duration(n) * time.Millisecond)
						}
					case "wrap-mode":
						result.savePrefs = true
						m.grid.WrapMode = logic.WrapModeFromString(parts[1], m.grid.WrapMode)
						m.prefs.WrapMode = m.grid.WrapMode.String()
					case "boundary-mode":
						result.savePrefs = true
						m.grid.BoundaryMode = logic.BoundaryModeFromString(parts[1], m.grid.BoundaryMode)
						m.prefs.BoundaryMode = m.grid.BoundaryMode.String()
					case "step-ahead":
						if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 && n <= 9999 {
							result.savePrefs = true
							m.stepAheadBy = n
							m.prefs.StepAheadBy = m.stepAheadBy
						}
					case "step-ahead-by":
						if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 && n <= 9999 {
							m.grid.StepAhead(n)
						}
					case "randomization":
						if n, err := strconv.Atoi(parts[1]); err == nil && n >= 0 && n <= 100 {
							result.savePrefs = true
							m.random = n
							m.prefs.Random = m.random
						}
					case "random-changes":
						if n, err := strconv.Atoi(parts[1]); err == nil && n >= 0 && n <= 100 {
							result.savePrefs = true
							m.grid.RandomChanges(n)
							m.prefs.RandomChanges = n
						}
					case "step-delay":
						if n, err := strconv.Atoi(parts[1]); err == nil && n >= 1 && n <= 2000 {
							result.savePrefs = true
							m.stepDelay = n
							m.prefs.StepDelay = m.stepDelay
						}
					case "run-recipe":
						if fn := parts[1]; fn != "" {
							if recipe, err := recipes.Load(fn); err == nil {
								_, _, _ = recipe.Run(m.grid, false)
							}
						}
					case "rule-perm":
						if n, err := strconv.Atoi(parts[1]); err == nil && n >= 0 && n <= 1<<18 {
							if r, err := logic.NewRuleFromPermutation(n); err == nil {
								result.savePrefs = true
								m.grid.Rule = r
								m.prefs.Rule = r.Rle()
							}
						}
					case "rule":
						if r, ok := logic.Rules[parts[1]]; ok {
							result.savePrefs = true
							m.grid.Rule = r
							m.prefs.Rule = r.Rle()
						} else if r, err := logic.NewRuleRle("", parts[1]); err == nil {
							result.savePrefs = true
							m.grid.Rule = r
							m.prefs.Rule = r.Rle()
						}
					case "rule-born-with":
						bw, sw := m.grid.Rule.BornWith(), m.grid.Rule.SurvivesWith()
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
							result.savePrefs = true
							m.grid.Rule = r
							m.prefs.Rule = r.Rle()
						}
					case "rule-survives-with":
						bw, sw := m.grid.Rule.BornWith(), m.grid.Rule.SurvivesWith()
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
							result.savePrefs = true
							m.grid.Rule = r
							m.prefs.Rule = r.Rle()
						}
					}
				}
			}
		}
		return result
	}
}
