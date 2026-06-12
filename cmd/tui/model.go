package main

import (
	"bytes"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"errors"
	"github.com/marrow16/gogol/cmd/tui/layout"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func newModel() *model {
	prfs := loadPrefs()
	for n, rle := range prfs.Rules {
		if nr, err := logic.NewRuleRle(n, rle); err == nil {
			logic.AddRule(n, nr)
		}
	}
	grid, err := logic.NewGrid(prfs.Height, prfs.Width, prfs.wrapMode(), prfs.boundaryMode())
	grid.Rule = prfs.rule()
	if err != nil {
		panic(err)
	}
	cs := prfs.cellStyle()
	m := &model{
		prefs:          prfs,
		grid:           grid,
		gridSurface:    newGridSurface(grid, cs),
		cellStyle:      cs,
		stepDelay:      prfs.StepDelay,
		random:         prfs.Random,
		gridHeight:     prfs.Height,
		gridWidth:      prfs.Width,
		stepAheadBy:    prfs.StepAheadBy,
		snapshotBefore: prfs.SnapshotBefore,
	}
	grid.Render = m.renderCell
	m.settings = &settings{m: m}
	m.capture = &capture{m: m}
	m.recipes = &recipesDialog{m: m}
	if len(prfs.Grid) > 0 {
		if p, err := patterns.NewPatternFromRle(strings.NewReader(prfs.Grid)); err == nil {
			p.Draw(grid, 0, 0, patterns.Rotate0)
			for _, c := range p.Comments {
				if after, ok := strings.CutPrefix(c, "Step: "); ok {
					if n, err := strconv.ParseUint(after, 10, 64); err == nil {
						m.grid.StepCount.Store(n)
					}
					break
				}
			}
		} else {
			grid.Randomize(m.random)
		}
		prfs.Grid = ""
	} else {
		grid.Randomize(m.random)
	}
	return m
}

func (m *model) renderCell(row int, col int, alive bool, changed bool) {
	qRow, qCol := row>>1, col>>1
	curr := m.gridSurface.Get(qRow, qCol)
	if len(curr) < 1 {
		curr = " "
	}
	r, _ := utf8.DecodeRuneInString(curr)
	m.gridSurface.Rune(qRow, qCol, rune(quadRune(r).update(col%2, row%2, alive, changed)), m.cellStyle)
}

func newGridSurface(g *logic.Grid, cellStyle lipgloss.Style) layout.Surface {
	sf := layout.NewSurface(g.Height/2, g.Width/2)
	for y := 0; y < sf.Height(); y++ {
		for x := 0; x < sf.Width(); x++ {
			sf.Text(y, x, " ", cellStyle)
		}
	}
	return sf
}

type displayMode int

const (
	modeSplash displayMode = iota
	modeStopped
	modeRunning
	modeSettings
	modeCapture
	modeRecipes
)

type model struct {
	mode        displayMode
	prefs       *prefs
	height      int
	width       int
	grid        *logic.Grid
	gridSurface layout.Surface
	cellStyle   lipgloss.Style
	settings    *settings
	capture     *capture
	recipes     *recipesDialog
	// settings...
	stepDelay      int
	stepAheadBy    int
	snapshotBefore bool
	snapshot       *patterns.Pattern
	snapshotSteps  uint64
	random         int
	gridHeight     int
	gridWidth      int
	// step ahead de-queuing...
	stepAheadActive bool
	stepAheadQueued bool
}

func (m *model) Init() tea.Cmd {
	return nil
}

type gridResizeResult struct {
	grid        *logic.Grid
	surface     layout.Surface
	noRandomize bool
}

func (m *model) savePrefs() tea.Cmd {
	return func() tea.Msg {
		m.prefs.save()
		return nil
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch mt := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = mt.Width
		m.height = mt.Height
	case gridResizeResult:
		m.gridSurface = mt.surface
		m.grid = mt.grid
		m.grid.Render = m.renderCell
		if mt.noRandomize {
			m.grid.Draw()
		} else {
			m.grid.Randomize(m.random)
		}
		m.prefs.Height, m.prefs.Width = m.grid.Height, m.grid.Width
		m.gridHeight, m.gridWidth = m.grid.Height, m.grid.Width
		m.prefs.WrapMode, m.prefs.BoundaryMode, m.prefs.Rule = m.grid.WrapMode.String(), m.grid.BoundaryMode.String(), m.grid.Rule.Rle()
		return m, m.savePrefs()
	case tea.KeyPressMsg:
		switch m.mode {
		case modeSettings:
			return m, m.settings.update(msg)
		case modeCapture:
			return m, m.capture.update(msg)
		case modeRecipes:
			return m, m.recipes.update(msg)
		default:
			return m, m.key(mt)
		}
	case steppedAhead:
		m.stepAheadActive = false
		m.stepAheadQueued = false
		m.grid.Draw()
	case tickMsg:
		if m.mode == modeRunning {
			if m.grid.Step() {
				return m, m.tick()
			} else {
				m.mode = modeStopped
			}
		}
	default:
		switch m.mode {
		case modeSettings:
			return m, m.settings.update(msg)
		case modeCapture:
			return m, m.capture.update(msg)
		case modeRecipes:
			return m, m.recipes.update(msg)
		}
	}
	return m, nil
}

func (m *model) key(msg tea.KeyPressMsg) tea.Cmd {
	if msg.String() != "tab" {
		m.stepAheadActive = false
	}
	switch msg.String() {
	case "ctrl+c":
		return tea.Quit
	case "esc":
		if m.mode == modeSplash {
			m.mode = modeStopped
		} else {
			m.mode = modeStopped
			m.save()
			return tea.Quit
		}
	case "ctrl+s":
		m.mode = modeSettings
	case "ctrl+k":
		m.capture.start()
		m.mode = modeCapture
	case "ctrl+g":
		m.mode = modeRecipes
	case "tab":
		m.mode = modeStopped
		m.stepAheadActive = true
		if !m.stepAheadQueued {
			m.stepAheadQueued = true
			return m.stepAhead()
		}
	case "home":
		m.mode = modeStopped
		m.grid.Randomize(m.random)
	case "backspace":
		m.mode = modeStopped
		if m.snapshot != nil {
			m.grid.StepCount.Store(m.snapshotSteps)
			m.snapshot.Draw(m.grid, 0, 0, patterns.Rotate0)
		}
	case "ctrl+o":
		m.mode = modeStopped
		if p, err := m.patternFromGrid(); err == nil {
			m.snapshotSteps = m.grid.StepCount.Load()
			m.snapshot = &p
		}
	case "ctrl+x":
		return func() tea.Msg {
			_, _ = m.exportGrid()
			return nil
		}
	case "f1":
		m.mode = modeSplash
	case "space":
		m.mode = modeStopped
		m.grid.Step()
	case "enter":
		if m.mode != modeRunning {
			m.mode = modeRunning
			return m.tick()
		} else {
			m.mode = modeStopped
		}
	}
	return nil
}

func (m *model) save() {
	m.prefs.Grid = ""
	if p, err := m.patternFromGrid(); err == nil {
		var buf bytes.Buffer
		if err = patterns.PatternRleEncode(p, &buf); err == nil {
			m.prefs.Grid = buf.String()
		}
	}
	m.prefs.save()
}

func (m *model) patternFromGrid() (patterns.Pattern, error) {
	p, err := patterns.NewPatternFromGrid(m.grid)
	if err == nil {
		p.Rule = m.grid.Rule
		p.Comments = []string{"Exported from GoGoL (https://github.com/marrow16/gogol)",
			"Wrap mode: " + m.grid.WrapMode.String(),
			"Boundary mode: " + m.grid.BoundaryMode.String(),
			"Step: " + strconv.FormatUint(m.grid.StepCount.Load(), 10),
		}
		p.Origination = m.prefs.Originator
	}
	return p, err
}

func (m *model) exportGrid() (filename string, err error) {
	var p patterns.Pattern
	if p, err = m.patternFromGrid(); err == nil {
		now := time.Now()
		p.Name = "Grid " + now.Format("2006-01-02 15:04:05")
		filename = "Grid " + now.Format("2006-01-02T150405") + ".rle"
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
	return filename, err
}

func (m *model) stopped() {
	m.mode = modeStopped
}

var bgColor = lipgloss.Color("#eeeeee")

const (
	dialogHeight = 20
	dialogWidth  = 60
)

func (m *model) dialogPosition(height, width int) (top, left int) {
	top = (m.height - height) / 2
	if top < 0 {
		top = 0
	}
	left = (m.width - width) / 2
	if left < 0 {
		left = 0
	}
	return
}

func (m *model) View() tea.View {
	sf := m.gridSurface
	title := "[stopped]"
	var csr *tea.Cursor
	switch m.mode {
	case modeRunning:
		return tea.View{
			WindowTitle:     "[running " + strconv.FormatUint(m.grid.StepCount.Load(), 10) + " - " + m.grid.Rule.Name() + "]",
			Content:         sf.(layout.GridSurface).RenderGrid(m.cellStyle),
			AltScreen:       true,
			MouseMode:       tea.MouseModeCellMotion,
			BackgroundColor: bgColor,
		}
	case modeStopped:
		if m.stepAheadActive {
			return tea.View{
				WindowTitle:     "[stepping ahead " + strconv.Itoa(m.stepAheadBy) + " - " + m.grid.Rule.Name() + "]",
				Content:         sf.(layout.GridSurface).RenderGrid(m.cellStyle),
				AltScreen:       true,
				MouseMode:       tea.MouseModeCellMotion,
				BackgroundColor: bgColor,
			}
		}
		return tea.View{
			WindowTitle:     "[stopped " + strconv.FormatUint(m.grid.StepCount.Load(), 10) + " - " + m.grid.Rule.Name() + "]",
			Content:         sf.(layout.GridSurface).RenderGrid(m.cellStyle),
			AltScreen:       true,
			MouseMode:       tea.MouseModeCellMotion,
			BackgroundColor: bgColor,
		}
	case modeSplash:
		sf2 := layout.NewSurface(m.height, m.width)
		sf2.Draw(0, 0, sf)
		renderSplash(m, sf2)
		sf = sf2
	case modeSettings:
		title = "[settings]"
		sf2 := layout.NewSurface(m.height, m.width)
		sf2.Draw(0, 0, sf)
		t, l := m.dialogPosition(dialogHeight, dialogWidth)
		rgn := sf2.Region(t, l, dialogHeight, dialogWidth)
		csr = m.settings.render(rgn)
		sf = sf2
	case modeCapture:
		title = "[" + m.capture.stage.String() + "]"
		sf2 := layout.NewSurface(m.height, m.width)
		sf2.Draw(0, 0, sf)
		csr = m.capture.render(sf2)
		sf = sf2
	case modeRecipes:
		title = "[recipes]"
		sf2 := layout.NewSurface(m.height, m.width)
		sf2.Draw(0, 0, sf)
		t, l := m.dialogPosition(dialogHeight, dialogWidth)
		rgn := sf2.Region(t, l, dialogHeight, dialogWidth)
		csr = m.recipes.render(rgn)
		sf = sf2
	}
	return tea.View{
		WindowTitle:     title,
		Content:         sf.Render(),
		AltScreen:       true,
		MouseMode:       tea.MouseModeCellMotion,
		Cursor:          csr,
		BackgroundColor: bgColor,
	}
}

type tickMsg time.Time

func (m *model) tick() tea.Cmd {
	return tea.Tick(time.Duration(m.stepDelay)*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type steppedAhead struct{}

func (m *model) stepAhead() tea.Cmd {
	return func() tea.Msg {
		if m.snapshotBefore {
			m.snapshotSteps = m.grid.StepCount.Load()
			if p, err := m.patternFromGrid(); err == nil {
				m.snapshot = &p
			} else {
				m.snapshot = nil
			}
		}
		m.grid.StepAhead(m.stepAheadBy)
		return steppedAhead{}
	}
}
