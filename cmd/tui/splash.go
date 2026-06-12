package main

import (
	"github.com/marrow16/gogol/cmd/tui/layout"
	"strings"
)

func renderSplash(m *model, sf layout.Surface) {
	const (
		height = 18
		width  = 40
	)
	t, l := m.dialogPosition(height, width)
	rgn := sf.Region(t, l, height, width)
	if rgn == nil {
		return
	}
	rgn.FillWith(0, 0, rgn.Height(), rgn.Width(), '\u00A0', dialogBgStyle)
	rgn.BoxRounded(0, 0, rgn.Height(), rgn.Width(), dialogTextStyle)
	rgn.TextCenter(0, 0, rgn.Width(), "Welcome to GoGoL", dialogTextStyle)
	rgn.TextCenter(1, 1, rgn.Width()-2, "Game Of Life", dialogTextStyle)
	rgn.TextCenter(2, 2, rgn.Width()-2, "implemented in Go & BubbleTea", dialogTextStyle)
	rgn.TextCenter(3, 2, rgn.Width()-2, "written by ", dialogTextStyle)
	rgn.TextCenter(4, 2, rgn.Width()-2, "Martin \"Marrow\" Rowlinson", dialogTextStyle)
	rgn.Text(5, 1, strings.Repeat("─", rgn.Width()-2), dialogTextStyle)
	rgn.Text(6, 7, "enter", dialogTextUlStyle)
	rgn.Text(6, 16, "- to start/stop", dialogTextStyle)
	rgn.Text(7, 7, "space", dialogTextUlStyle)
	rgn.Text(7, 16, "- to step", dialogTextStyle)
	rgn.Text(8, 7, "tab", dialogTextUlStyle)
	rgn.Text(8, 16, "- to step ahead", dialogTextStyle)
	rgn.Text(9, 7, "home", dialogTextUlStyle)
	rgn.Text(9, 16, "- randomize", dialogTextStyle)
	rgn.Text(10, 7, "ctrl+s", dialogTextUlStyle)
	rgn.Text(10, 16, "- settings", dialogTextStyle)
	rgn.Text(11, 7, "ctrl+k", dialogTextUlStyle)
	rgn.Text(11, 16, "- capture mode", dialogTextStyle)
	rgn.Text(12, 7, "ctrl+g", dialogTextUlStyle)
	rgn.Text(12, 16, "- grid recipes", dialogTextStyle)
	rgn.Text(13, 7, "ctrl+o", dialogTextUlStyle)
	rgn.Text(13, 16, "- snapshot", dialogTextStyle)
	rgn.Text(14, 6, "backspace", dialogTextUlStyle)
	rgn.Text(14, 16, "- restore to snapshot", dialogTextStyle)
	rgn.Text(15, 7, "ctrl+x", dialogTextUlStyle)
	rgn.Text(15, 16, "- export grid", dialogTextStyle)
	rgn.Text(16, 9, "esc", dialogTextUlStyle)
	rgn.Text(16, 13, "/", dialogTextStyle)
	rgn.Text(16, 15, "ctrl+c", dialogTextUlStyle)
	rgn.Text(16, 22, "- to quit", dialogTextStyle)
}
