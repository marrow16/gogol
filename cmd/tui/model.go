package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/marrow16/gogol/cmd/tui/layout"
	"github.com/marrow16/gogol/logic"
	"time"
	"unicode/utf8"
)

func newModel() *model {
	prfs := loadPrefs()
	grid, err := logic.NewGrid(prfs.Height, prfs.Width, prfs.wrapMode(), prfs.boundaryMode())
	if err != nil {
		panic(err)
	}
	cs := prfs.cellStyle()
	m := &model{
		prefs:         prfs,
		grid:          grid,
		gridSurface:   newGridSurface(grid, cs),
		cellStyle:     cs,
		stepDelay:     50,
		random:        30,
		gridHeight:    prfs.Height,
		gridWidth:     prfs.Width,
		splashShowing: true,
	}
	grid.Render = m.renderCell
	m.settings = &settings{m: m}
	grid.Randomize(m.random)
	return m
}

func (m *model) renderCell(row int, col int, alive bool, changed bool) {
	tRow, tCol := row/2, col/2
	curr := m.gridSurface.Get(tRow, tCol)
	if len(curr) < 1 {
		curr = " "
	}
	r, _ := utf8.DecodeRuneInString(curr)
	qr := quadRune(r)
	qr = qr.update(col%2, row%2, alive, changed)
	m.gridSurface.Text(tRow, tCol, string(qr), m.cellStyle)
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
	settings        *settings
	splashShowing   bool
	// settings...
	stepDelay  int
	random     int
	gridHeight int
	gridWidth  int
}

func (m *model) Init() tea.Cmd {
	return nil
}

type gridResizeResult struct {
	grid    *logic.Grid
	surface layout.Surface
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
		m.grid.Randomize(m.random)
		m.prefs.Height, m.prefs.Width = m.grid.Height, m.grid.Width
		return m, m.savePrefs()
	case tea.KeyPressMsg:
		m.splashShowing = false
		if m.settingsShowing {
			return m, m.settings.update(msg)
		} else {
			switch mt.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			case "ctrl+s":
				m.running = false
				m.settingsShowing = true
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
	case tickMsg:
		if m.running && !m.settingsShowing {
			if m.grid.Step() {
				return m, m.tick()
			} else {
				m.running = false
			}
		}
	default:
		if m.settingsShowing {
			return m, m.settings.update(msg)
		}
	}
	return m, nil
}

var bgColor = lipgloss.Color("#eeeeee")

func (m *model) View() tea.View {
	sf := m.gridSurface
	title := "[stopped]"
	var csr *tea.Cursor
	if m.splashShowing {
		sf2 := layout.NewSurface(m.height, m.width)
		sf2.Draw(0, 0, sf)
		renderSplash(sf2)
		sf = sf2
	} else if m.settingsShowing {
		title = "[settings]"
		sf2 := layout.NewSurface(m.height, m.width)
		sf2.Draw(0, 0, sf)
		rgn := sf2.Region(2, 5, 20, 60)
		csr = m.settings.render(rgn)
		sf = sf2
	} else if m.running {
		title = "[running]"
	}
	return tea.View{
		WindowTitle:     title,
		Content:         sf.Render(),
		AltScreen:       true,
		MouseMode:       tea.MouseModeCellMotion,
		Cursor:          csr,
		BackgroundColor: bgColor,
	}
	/*
		v := tea.NewView(m.gridSurface.Render())
		//v.Cursor = csr
		v.AltScreen = true
		v.MouseMode = tea.MouseModeCellMotion
		if m.running {
			v.WindowTitle = "[running]"
		} else {
			v.WindowTitle = "[paused]"
		}
		return v
	*/
}

type tickMsg time.Time

func (m *model) tick() tea.Cmd {
	return tea.Tick(time.Duration(m.stepDelay)*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
