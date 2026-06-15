package layout

import tea "charm.land/bubbletea/v2"

type ClickPoint [2]int //Y,X

type ClickPoints[T any] map[ClickPoint]func(T) tea.Cmd

func (cp ClickPoints[T]) Add(pl Placement, fn func(parent T) tea.Cmd) {
	for l := 0; l < pl.Extent; l++ {
		cp[ClickPoint{pl.Row, pl.Col + l}] = fn
	}
}

func (cp ClickPoints[T]) AddRegion(rgn Surface, fn func(parent T) tea.Cmd) {
	for r := 0; r < rgn.Height(); r++ {
		for c := 0; c < rgn.Width(); c++ {
			cp[ClickPoint{r + rgn.AbsoluteTop(), c + rgn.AbsoluteLeft()}] = fn
		}
	}
}
