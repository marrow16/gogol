package settings

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"image/color"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func NewSettings() *Settings {
	s := &Settings{
		ScreenHeight:    600,
		ScreenWidth:     900,
		StepDelay:       25,
		StepAheadBy:     2000,
		SkipBackBy:      100,
		Randomization:   15,
		Height:          100,
		Width:           100,
		Zoom:            1.0,
		WrapMode:        logic.WrapAll,
		BoundaryMode:    logic.DeadBoundary,
		Rule:            "B3/S23",
		CellSize:        10,
		CellBorders:     true,
		CellAliveColor:  color.NRGBA{R: 0, G: 0, B: 0, A: 0xff},
		CellDeadColor:   color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
		CellBorderColor: color.NRGBA{R: 240, G: 240, B: 239, A: 255},
	}
	if path, err := settingsPath(false); err == nil {
		if f, err := os.Open(path); err == nil {
			defer func() {
				_ = f.Close()
			}()
			p := prefs{}
			if err = json.NewDecoder(f).Decode(&p); err == nil {
				s.fromPrefs(p)
			}
		}
	}
	return s
}

func settingsPath(create bool) (string, error) {
	const prefsFilename = "prefs.gui.json"
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "GoGoL")
	if create {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
	}
	return filepath.Join(dir, prefsFilename), nil
}

type Settings struct {
	ScreenHeight      int
	ScreenWidth       int
	StepDelay         int
	StepAheadBy       int
	StepAheadSnapshot bool
	SkipBackBy        int
	Randomization     int
	Height            int
	Width             int
	Zoom              float32
	WrapMode          logic.WrapMode
	BoundaryMode      logic.BoundaryMode
	Rule              string
	CellSize          int
	CellBorders       bool
	CellAliveColor    color.NRGBA
	CellDeadColor     color.NRGBA
	CellBorderColor   color.NRGBA
	Originator        string
	Rules             map[string]string
	Patterns          []string
	PatternLibraries  []string
	Recipes           []string
	SavedGrid         *logic.Grid
	Recording         bool
	RepeatDetection   bool
}

func (s *Settings) Save(grid *logic.Grid, zoom float32) {
	if path, err := settingsPath(true); err == nil {
		if f, err := os.Create(path); err == nil {
			defer func() {
				_ = f.Close()
			}()
			p := prefs{
				ScreenHeight:      s.ScreenHeight,
				ScreenWidth:       s.ScreenWidth,
				Height:            s.Height,
				Width:             s.Width,
				Zoom:              zoom,
				StepDelay:         s.StepDelay,
				StepAheadBy:       s.StepAheadBy,
				StepAheadSnapshot: s.StepAheadSnapshot,
				SkipBackBy:        s.SkipBackBy,
				Randomization:     s.Randomization,
				WrapMode:          grid.WrapMode.String(),
				BoundaryMode:      grid.BoundaryMode.String(),
				CellAliveColor:    fmt.Sprintf("#%02X%02X%02X", s.CellAliveColor.R, s.CellAliveColor.G, s.CellAliveColor.B),
				CellDeadColor:     fmt.Sprintf("#%02X%02X%02X", s.CellDeadColor.R, s.CellDeadColor.G, s.CellDeadColor.B),
				CellBorderColor:   fmt.Sprintf("#%02X%02X%02X", s.CellBorderColor.R, s.CellBorderColor.G, s.CellBorderColor.B),
				CellBorders:       s.CellBorders,
				CellSize:          s.CellSize,
				Rule:              grid.Rule.Rle(),
				Rules:             s.Rules,
				Patterns:          s.Patterns,
				PatternLibraries:  s.PatternLibraries,
				Originator:        s.Originator,
				Recipes:           s.Recipes,
				Recording:         s.Recording,
				RepeatDetection:   s.RepeatDetection,
			}
			if pattern, err := s.PatternFromGrid(grid); err == nil {
				var buf bytes.Buffer
				if err = patterns.PatternRleEncode(pattern, &buf); err == nil {
					p.Grid = buf.String()
				}
			}
			enc := json.NewEncoder(f)
			enc.SetIndent("", "  ")
			_ = enc.Encode(p)
		}
	}
}

func (s *Settings) PatternFromGrid(grid *logic.Grid) (patterns.Pattern, error) {
	p, err := patterns.NewPatternFromGrid(grid)
	if err == nil {
		p.Rule = grid.Rule
		p.Comments = []string{"Exported from GoGoL (https://github.com/marrow16/gogol)",
			"Wrap mode: " + grid.WrapMode.String(),
			"Boundary mode: " + grid.BoundaryMode.String(),
			"Step: " + strconv.FormatUint(grid.StepCount.Load(), 10),
		}
		p.Origination = s.Originator
	}
	return p, err
}

func LoadPatternsLibrary(path string) (*int, error) {
	count := 0
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".rle") {
			if err := LoadPattern(path); err == nil {
				count++
			}
		}
		return nil
	})
	if count > 0 {
		return &count, nil
	} else {
		return nil, errors.New("No .rle files found")
	}
}

func LoadPattern(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New("Unable to open file")
	}
	defer func() {
		_ = f.Close()
	}()
	if p, err := patterns.NewPatternFromRle(f); err != nil {
		return err
	} else {
		if p.Name == "" {
			p.Name = filepath.Base(path)
		}
		p.Filename = filepath.Base(path)
		patterns.PatternLibrary[p.Name] = p
		return nil
	}
}

func (s *Settings) AddPattern(filename string) {
	if ap, err := filepath.Abs(filename); err == nil {
		filename = ap
	}
	if !slices.Contains(s.Patterns, filename) {
		s.Patterns = append(s.Patterns, filename)
	}
}

func (s *Settings) AddPatternLibrary(path string) {
	if ap, err := filepath.Abs(path); err == nil {
		path = ap
	}
	if !slices.Contains(s.PatternLibraries, path) {
		s.PatternLibraries = append(s.PatternLibraries, path)
	}
}

func (s *Settings) AddRecipe(filename string) {
	if ap, err := filepath.Abs(filename); err == nil {
		filename = ap
	}
	if !slices.Contains(s.Recipes, filename) {
		s.Recipes = append(s.Recipes, filename)
	}
}

func (s *Settings) fromPrefs(p prefs) {
	if p.ScreenWidth >= 100 {
		s.ScreenWidth = p.ScreenWidth
	}
	if p.ScreenHeight >= 100 {
		s.ScreenHeight = p.ScreenHeight
	}
	if p.Height > 2 {
		s.Height = p.Height
	}
	if p.Width > 2 {
		s.Width = p.Width
	}
	if p.StepDelay > 0 {
		s.StepDelay = p.StepDelay
	}
	if p.StepAheadBy > 0 {
		s.StepAheadBy = p.StepAheadBy
	}
	s.StepAheadSnapshot = p.StepAheadSnapshot
	if p.SkipBackBy > 0 {
		s.SkipBackBy = p.SkipBackBy
	}
	if p.Randomization >= 0 && p.Randomization <= 100 {
		s.Randomization = p.Randomization
	}
	s.WrapMode = logic.WrapModeFromString(p.WrapMode, s.WrapMode)
	s.BoundaryMode = logic.BoundaryModeFromString(p.BoundaryMode, s.BoundaryMode)
	if c, ok := parseColor(p.CellAliveColor); ok {
		s.CellAliveColor = c
	}
	if c, ok := parseColor(p.CellDeadColor); ok {
		s.CellDeadColor = c
	}
	if c, ok := parseColor(p.CellBorderColor); ok {
		s.CellBorderColor = c
	}
	s.CellBorders = p.CellBorders
	if p.CellSize > 1 {
		s.CellSize = p.CellSize
	}
	if zf := p.Zoom * float32(s.CellSize); zf >= 1.0 {
		s.Zoom = p.Zoom
	}
	if r, err := logic.NewRuleRle("", p.Rule); err == nil {
		s.Rule = r.Rle()
	}
	s.Rules = p.Rules
	s.Patterns = p.Patterns
	s.PatternLibraries = p.PatternLibraries
	s.Originator = p.Originator
	s.Recipes = p.Recipes
	s.Recording = p.Recording
	s.RepeatDetection = p.RepeatDetection
	if len(p.Grid) > 0 {
		if pattern, err := patterns.NewPatternFromRle(strings.NewReader(p.Grid)); err == nil {
			if g, err := logic.NewGrid(pattern.Height, pattern.Width, s.WrapMode, s.BoundaryMode); err == nil {
				s.SavedGrid = g
				s.SavedGrid.Rule = pattern.Rule
				pattern.Draw(s.SavedGrid, 0, 0, patterns.Rotate0)
				for _, c := range pattern.Comments {
					if after, ok := strings.CutPrefix(c, "Step: "); ok {
						if n, err := strconv.Atoi(after); err == nil {
							s.SavedGrid.StepCount.Store(uint64(n))
						}
					}
				}
			}
		}
	}
	for _, path := range s.PatternLibraries {
		_, _ = LoadPatternsLibrary(path)
	}
	for _, path := range s.Patterns {
		_ = LoadPattern(path)
	}
}

type prefs struct {
	ScreenHeight      int               `json:"screen_height"`
	ScreenWidth       int               `json:"screen_width"`
	Height            int               `json:"height"`
	Width             int               `json:"width"`
	Zoom              float32           `json:"zoom"`
	StepDelay         int               `json:"step_delay"`
	StepAheadBy       int               `json:"step_ahead_by"`
	StepAheadSnapshot bool              `json:"snapshot_before_step_ahead"`
	SkipBackBy        int               `json:"skip_back_by"`
	Randomization     int               `json:"randomization"`
	WrapMode          string            `json:"wrap_mode"`
	BoundaryMode      string            `json:"boundary_mode"`
	CellAliveColor    string            `json:"cell_alive_color"`
	CellDeadColor     string            `json:"cell_dead_color"`
	CellBorderColor   string            `json:"cell_border_color"`
	CellBorders       bool              `json:"cell_borders"`
	CellSize          int               `json:"cell_size"`
	Rule              string            `json:"rule"`
	Rules             map[string]string `json:"rules,omitempty"`
	Patterns          []string          `json:"patterns,omitempty"`
	PatternLibraries  []string          `json:"pattern_libraries,omitempty"`
	Originator        string            `json:"originator,omitempty"`
	Grid              string            `json:"grid,omitempty"`
	Recipes           []string          `json:"recipes,omitempty"`
	Recording         bool              `json:"recording"`
	RepeatDetection   bool              `json:"repeat_detection"`
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
