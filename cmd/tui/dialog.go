package main

import (
	tea "charm.land/bubbletea/v2"
	"github.com/marrow16/gogol/cmd/tui/layout"
)

type dialog interface {
	render(rgn layout.Surface) *tea.Cursor
	update(msg tea.Msg) tea.Cmd
	title() string
}
