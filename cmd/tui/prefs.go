package main

import (
	"charm.land/lipgloss/v2"
	"encoding/json"
	"fmt"
	"github.com/marrow16/gogol/logic"
	"os"
	"path/filepath"
	"regexp"
	"slices"
)

type prefs struct {
	Height           int                 `json:"height"`
	Width            int                 `json:"width"`
	StepDelay        int                 `json:"step_delay"`
	StepAheadBy      int                 `json:"step_ahead_by"`
	SnapshotBefore   bool                `json:"snapshot_before_step_ahead"`
	Random           int                 `json:"random"`
	RandomChanges    int                 `json:"random_changes"`
	WrapMode         string              `json:"wrap_mode"`
	BoundaryMode     string              `json:"boundary_mode"`
	CellFG           string              `json:"cell_foreground_color"`
	CellBG           string              `json:"cell_background_color"`
	Rule             string              `json:"rule"`
	Rules            map[string]string   `json:"rules,omitempty"`
	Patterns         []string            `json:"patterns,omitempty"`
	PatternLibraries []string            `json:"pattern_libraries,omitempty"`
	Originator       string              `json:"originator,omitempty"`
	SavePath         string              `json:"save_path,omitempty"`
	Grid             string              `json:"grid,omitempty"`
	Recipes          []string            `json:"recipes,omitempty"`
	Shortcuts        map[string][]string `json:"shortcuts"`
}

const (
	prefsFilename       = "prefs.json"
	defaultHeight       = 100
	defaultWidth        = 200
	defaultStep         = 50
	defaultStepAhead    = 1000
	defaultRandom       = 30
	defaultWrapMode     = "toroidal"
	defaultBoundaryMode = "dead"
	defaultFg           = "#6680e6"
	defaultBg           = "#ffffff"
	defaultRule         = "B3/S23"
)

func loadPrefs() *prefs {
	if f, err := os.Open(prefsFilename); err == nil {
		defer f.Close()
		result := &prefs{}
		if err = json.NewDecoder(f).Decode(result); err == nil {
			result.validate()
			go result.loadPatterns()
			return result
		}
	}
	return &prefs{
		Height:         defaultHeight,
		Width:          defaultWidth,
		StepDelay:      defaultStep,
		StepAheadBy:    defaultStepAhead,
		SnapshotBefore: true,
		Random:         defaultRandom,
		WrapMode:       defaultWrapMode,
		BoundaryMode:   defaultBoundaryMode,
		CellFG:         defaultFg,
		CellBG:         defaultBg,
		Rule:           defaultRule,
		Rules:          make(map[string]string),
		Shortcuts:      make(map[string][]string),
	}
}

func (p *prefs) validate() {
	if p.Height < 2 {
		p.Height = 2
	}
	if p.Height%2 != 0 {
		p.Height++
	}
	if p.Width < 2 {
		p.Width = 2
	}
	if p.Width%2 != 0 {
		p.Width++
	}
	if p.StepDelay < 1 {
		p.StepDelay = defaultStep
	}
	if p.StepAheadBy < 1 {
		p.StepAheadBy = defaultStepAhead
	}
	if p.Random < 0 || p.Random > 100 {
		p.Random = defaultRandom
	}
	colorRegex := regexp.MustCompile("^#[0-9a-fA-F]{6}$")
	if !colorRegex.MatchString(p.CellFG) {
		p.CellFG = defaultFg
	}
	if !colorRegex.MatchString(p.CellBG) {
		p.CellBG = defaultBg
	}
	if p.Rules == nil {
		p.Rules = make(map[string]string)
	}
}

func (p *prefs) wrapMode() logic.WrapMode {
	return logic.WrapModeFromString(p.WrapMode, logic.WrapNone)
}

func (p *prefs) setWrapMode(m logic.WrapMode) *prefs {
	p.WrapMode = m.String()
	return p
}

func (p *prefs) boundaryMode() logic.BoundaryMode {
	return logic.BoundaryModeFromString(p.BoundaryMode, logic.DeadBoundary)
}

func (p *prefs) setBoundaryMode(m logic.BoundaryMode) *prefs {
	p.BoundaryMode = m.String()
	return p
}

func (p *prefs) rule() logic.Rule {
	if r, err := logic.NewRuleRle("", p.Rule); err == nil {
		return r
	}
	return logic.StandardRule
}

func (p *prefs) setRule(r logic.Rule) *prefs {
	p.Rule = r.Rle()
	return p
}

func (p *prefs) cellStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(p.CellFG)).Background(lipgloss.Color(p.CellBG))
}

func (p *prefs) setCellStyle(s lipgloss.Style) *prefs {
	fgR, fgG, fgB := rgb(s.GetForeground())
	bgR, bgG, bgB := rgb(s.GetBackground())
	p.CellFG = fmt.Sprintf("#%02X%02X%02X", fgR, fgG, fgB)
	p.CellBG = fmt.Sprintf("#%02X%02X%02X", bgR, bgG, bgB)
	return p
}

func (p *prefs) loadPatterns() {
	count := 0
	for _, pl := range p.PatternLibraries {
		lr := loadPatternsLibrary(pl)
		count += lr.loaded
	}
	for _, pt := range p.Patterns {
		if err := loadPattern(pt); err == nil {
			count++
		}
	}
}

func (p *prefs) addPattern(filename string) {
	if ap, err := filepath.Abs(filename); err == nil {
		filename = ap
	}
	if !slices.Contains(p.Patterns, filename) {
		p.Patterns = append(p.Patterns, filename)
	}
}

func (p *prefs) addPatternLibrary(path string) {
	if ap, err := filepath.Abs(path); err == nil {
		path = ap
	}
	if !slices.Contains(p.PatternLibraries, path) {
		p.PatternLibraries = append(p.PatternLibraries, path)
	}
}

func (p *prefs) addRule(name string, rle string) {
	p.Rules[name] = rle
}

func (p *prefs) addRecipe(filename string) {
	if ap, err := filepath.Abs(filename); err == nil {
		filename = ap
	}
	if !slices.Contains(p.Recipes, filename) {
		p.Recipes = append(p.Recipes, filename)
	}
}

func (p *prefs) clearPatterns() {
	p.PatternLibraries = []string{}
	p.Patterns = []string{}
}

func (p *prefs) save() {
	if f, err := os.Create(prefsFilename); err == nil {
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		_ = enc.Encode(p)
	}
}
