package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/marrow16/gogol/cmd/tui/layout"
)

const (
	dialogHeight = 20
	dialogWidth  = 60
)

var (
	bgColor       = lipgloss.Color("#eeeeee")
	dialogBgStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#eeeeee"))
	dialogTextStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#eeeeee")).
			Foreground(lipgloss.Color("#6680e6"))
	dialogTextUlStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#eeeeee")).
				Foreground(lipgloss.Color("#6680e6")).
				Underline(true)
	dialogTabStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#6680e6")).
			Foreground(lipgloss.Color("#ffffff"))
	dialogPreviewStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#ffffff")).
				Foreground(lipgloss.Color("#000000"))
	dialogPreviewFocusedStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("#ffffff")).
					Foreground(lipgloss.Color("#6680e6"))
	cursorColor = lipgloss.Color("#00D000")
)

func clampOffsets(offsetY, offsetX int, availableHeight, availableWidth int, itemHeight, itemWidth int) (int, int) {
	maxY := itemHeight - availableHeight
	if maxY < 0 {
		maxY = 0
	}
	maxX := itemWidth - availableWidth
	if maxX < 0 {
		maxX = 0
	}
	if offsetY < 0 {
		offsetY = 0
	} else if offsetY > maxY {
		offsetY = maxY
	}
	if offsetX < 0 {
		offsetX = 0
	} else if offsetX > maxX {
		offsetX = maxX
	}
	return offsetY, offsetX
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

type tabs []tab

func renderTabs[T any](rgn layout.Surface, clickPts layout.ClickPoints[T], tbs tabs, current int, set func(t int)) {
	x := 3
	for _, t := range tbs {
		if t.tabNo == current {
			rgn.Text(1, x-1, " "+t.title+" ", dialogTabStyle)
		} else {
			clickPts.Add(rgn.Text(1, x, t.title, dialogTextStyle), func(d T) tea.Cmd {
				set(t.tabNo)
				return nil
			})
			if t.ul != -1 {
				rgn.Text(1, x+t.ul, t.title[t.ul:t.ul+1], dialogTextUlStyle)
			}
		}
		x += len(t.title) + 3
	}
}

type tab struct {
	title string
	ul    int
	tabNo int
}
