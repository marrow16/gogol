package main

import (
	"bytes"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/marrow16/gogol/cmd/tui/layout"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
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
		prefs:       prfs,
		grid:        grid,
		gridSurface: newGridSurface(grid, cs),
		cellStyle:   cs,
		stepDelay:   prfs.StepDelay,
		random:      prfs.Random,
		gridHeight:  prfs.Height,
		gridWidth:   prfs.Width,
		stepAheadBy: prfs.StepAheadBy,
	}
	grid.Render = m.renderCell
	m.settings = &settings{m: m}
	m.capture = &capture{m: m}
	if len(prfs.Grid) > 0 {
		if p, err := patterns.NewPatternFromRle(strings.NewReader(prfs.Grid)); err == nil {
			p.Draw(grid, 0, 0, patterns.Rotate0)
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
	// settings...
	stepDelay   int
	stepAheadBy int
	random      int
	gridHeight  int
	gridWidth   int
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
		return m, m.savePrefs()
	case tea.KeyPressMsg:
		switch m.mode {
		case modeSettings:
			return m, m.settings.update(msg)
		case modeCapture:
			return m, m.capture.update(msg)
		default:
			return m, m.key(mt)
		}
	case steppedAhead:
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
		}
	}
	return m, nil
}

func (m *model) key(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c":
		return tea.Quit
	case "esc":
		if m.mode == modeSplash {
			m.mode = modeStopped
		} else {
			m.save()
			return tea.Quit
		}
	case "ctrl+s":
		m.mode = modeSettings
	case "ctrl+k":
		m.capture.start()
		m.mode = modeCapture
	case "end":
		m.mode = modeStopped
		return m.stepAhead()
	case "home":
		m.mode = modeStopped
		m.grid.Randomize(m.random)
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
	// save current grid as rle
	cells := make([]bool, m.grid.Height*m.grid.Width)
	idx := 0
	for y := 0; y < m.grid.Height; y++ {
		for x := 0; x < m.grid.Width; x++ {
			if cell := m.grid.GetCell(y, x); cell != nil {
				cells[idx] = cell.Alive
			}
			idx++
		}
	}
	m.prefs.Grid = ""
	if p, err := patterns.NewPattern("Grid save", m.grid.Width, cells); err == nil {
		var buf bytes.Buffer
		if err = patterns.PatternRleEncode(p, &buf); err == nil {
			m.prefs.Grid = buf.String()
		}
	}
	m.prefs.save()
}

func (m *model) stopped() {
	m.mode = modeStopped
}

var bgColor = lipgloss.Color("#eeeeee")

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
		renderSplash(sf2)
		sf = sf2
	case modeSettings:
		title = "[settings]"
		sf2 := layout.NewSurface(m.height, m.width)
		sf2.Draw(0, 0, sf)
		rgn := sf2.Region(2, 5, 20, 60)
		csr = m.settings.render(rgn)
		sf = sf2
	case modeCapture:
		title = "[capture-" + m.capture.stage.String() + "]"
		sf2 := layout.NewSurface(m.height, m.width)
		sf2.Draw(0, 0, sf)
		csr = m.capture.render(sf2)
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
		m.grid.StepAhead(m.stepAheadBy)
		return steppedAhead{}
	}
}
