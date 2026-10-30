package gui

import (
	"os/exec"
	"runtime"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
)

const (
	gogolVersion = "1.6.77"
	gogolRepo    = "https://github.com/marrow16/gogol"
)

func newAboutPopout(p *menuPopup) *aboutPopout {
	result := &aboutPopout{
		parent:  p,
		version: "Version: " + gogolVersion + " (Go: " + strings.TrimPrefix(runtime.Version(), "go") + ")",
	}
	result.links = map[*widget.Clickable]string{
		&result.linkRepo: gogolRepo,
	}
	return result
}

type aboutPopout struct {
	parent   *menuPopup
	linkRepo widget.Clickable
	links    map[*widget.Clickable]string
	version  string
}

func openURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default: // Linux, BSD, etc.
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func (p *aboutPopout) layout(gtx layout.Context) layout.Dimensions {
	m := measureText(gtx, "M")
	minX := m.Size.X * 30
	for c, l := range p.links {
		if c.Clicked(gtx) {
			_ = openURL(l)
		}
	}
	return popoutLayout(gtx, flexVertical(0,
		rigidLabel("GoGoL", text.Middle, font.Bold, minX),
		rigidLabel(p.version, text.Middle, 0, minX),
		rigidSpacerVertical(m.Size.Y/2),
		rigidLabel(`Author: Martin "Marrow" Rowlinson`, text.Middle, 0, minX),
		rigidFixedWidth(linkLabel(&p.linkRepo, gogolRepo), minX, layout.Center),
		rigidSpacerVertical(m.Size.Y/2),
	))
}

func (p *aboutPopout) hasFocus(_ layout.Context) bool {
	return false
}

func (p *aboutPopout) reset() {
	// nothing to reset
}
