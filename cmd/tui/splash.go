package main

import (
	"github.com/marrow16/gogol/cmd/tui/layout"
	"strings"
)

func renderSplash(sf layout.Surface) {
	const (
		height = 11
		width  = 40
	)
	rgn := sf.Region((sf.Height()/2)-(height/2), (sf.Width()/2)-(width/2), height, width)
	if rgn == nil {
		return
	}
	rgn.FillWith(0, 0, rgn.Height(), rgn.Width(), '\u00A0', settingsBgStyle)
	rgn.BoxRounded(0, 0, rgn.Height(), rgn.Width(), settingsTextStyle)
	rgn.TextCenter(0, 0, rgn.Width(), "Welcome to GoGoL", settingsTextStyle)
	rgn.TextCenter(1, 1, rgn.Width()-2, "Game Of Life", settingsTextStyle)
	rgn.TextCenter(2, 2, rgn.Width()-2, "implemented in Go & BubbleTea", settingsTextStyle)
	rgn.TextCenter(3, 2, rgn.Width()-2, "written by", settingsTextStyle)
	rgn.TextCenter(4, 2, rgn.Width()-2, "Martin \"Marrow\" Rowlinson", settingsTextStyle)
	rgn.Text(5, 1, strings.Repeat("─", rgn.Width()-2), settingsTextStyle)
	rgn.Text(6, 9, "enter", settingsTextUlStyle)
	rgn.Text(6, 16, "- to start/stop", settingsTextStyle)
	rgn.Text(7, 9, "space", settingsTextUlStyle)
	rgn.Text(7, 16, "- to step", settingsTextStyle)
	rgn.Text(8, 9, "ctrl+s", settingsTextUlStyle)
	rgn.Text(8, 16, "- for settings", settingsTextStyle)
	rgn.Text(9, 9, "esc", settingsTextUlStyle)
	rgn.Text(9, 13, "/", settingsTextStyle)
	rgn.Text(9, 15, "ctrl+c", settingsTextUlStyle)
	rgn.Text(9, 22, "- to quit", settingsTextStyle)
}
