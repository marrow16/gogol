package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/marrow16/gogol/cmd/tui/layout"
	"github.com/marrow16/gogol/logic"
	"strconv"
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
		prefs:         prfs,
		grid:          grid,
		gridSurface:   newGridSurface(grid, cs),
		cellStyle:     cs,
		stepDelay:     prfs.StepDelay,
		random:        prfs.Random,
		gridHeight:    prfs.Height,
		gridWidth:     prfs.Width,
		stepAheadBy:   prfs.StepAheadBy,
		splashShowing: true,
	}
	grid.Render = m.renderCell
	m.settings = &settings{m: m}
	m.capture = &capture{m: m}
	grid.Randomize(m.random)
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

type model struct {
	prefs           *prefs
	height          int
	width           int
	grid            *logic.Grid
	gridSurface     layout.Surface
	cellStyle       lipgloss.Style
	running         bool
	settingsShowing bool
	capturing       bool
	settings        *settings
	capture         *capture
	splashShowing   bool
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
		m.splashShowing = false
		if m.settingsShowing {
			return m, m.settings.update(msg)
		} else if m.capturing {
			return m, m.capture.update(msg)
		} else {
			switch mt.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			case "ctrl+s":
				m.running = false
				m.settingsShowing = true
			case "ctrl+k":
				m.running = false
				m.capture.start()
			case "end":
				m.running = false
				return m, m.stepAhead()
			case "home":
				m.running = false
				m.grid.Randomize(m.random)
			case "f1":
				m.running = false
				m.splashShowing = true
			case "space":
				m.grid.Step()
			case "enter":
				if m.running {
					m.running = false
				} else {
					m.running = true
					return m, m.tick()
				}
			}
		}
	case steppedAhead:
		m.grid.Draw()
	case tickMsg:
		if m.running && !m.settingsShowing && !m.capturing {
			if m.grid.Step() {
				return m, m.tick()
			} else {
				m.running = false
			}
		}
	default:
		if m.settingsShowing {
			return m, m.settings.update(msg)
		} else if m.capturing {
			return m, m.capture.update(msg)
		}
	}
	return m, nil
}

var bgColor = lipgloss.Color("#eeeeee")

func (m *model) View() tea.View {
	sf := m.gridSurface
	title := "[stopped]"
	var csr *tea.Cursor
	overlayed := false
	if m.splashShowing {
		overlayed = true
		sf2 := layout.NewSurface(m.height, m.width)
		sf2.Draw(0, 0, sf)
		renderSplash(sf2)
		sf = sf2
	} else if m.settingsShowing {
		overlayed = true
		title = "[settings]"
		sf2 := layout.NewSurface(m.height, m.width)
		sf2.Draw(0, 0, sf)
		rgn := sf2.Region(2, 5, 20, 60)
		csr = m.settings.render(rgn)
		sf = sf2
	} else if m.capturing {
		overlayed = true
		title = "[capture-" + m.capture.stage.String() + "]"
		sf2 := layout.NewSurface(m.height, m.width)
		sf2.Draw(0, 0, sf)
		csr = m.capture.render(sf2)
		sf = sf2
	} else if m.running {
		title = "[running " + strconv.FormatUint(m.grid.StepCount.Load(), 10) + " - " + m.grid.Rule.Name() + "]"
	} else {
		title = "[stopped " + strconv.FormatUint(m.grid.StepCount.Load(), 10) + " - " + m.grid.Rule.Name() + "]"
	}
	gsf, isGsf := sf.(layout.GridSurface)
	if overlayed || !isGsf {
		return tea.View{
			WindowTitle:     title,
			Content:         sf.Render(),
			AltScreen:       true,
			MouseMode:       tea.MouseModeCellMotion,
			Cursor:          csr,
			BackgroundColor: bgColor,
		}
	} else {
		return tea.View{
			WindowTitle:     title,
			Content:         gsf.RenderGrid(m.cellStyle),
			AltScreen:       true,
			MouseMode:       tea.MouseModeCellMotion,
			Cursor:          csr,
			BackgroundColor: bgColor,
		}
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
