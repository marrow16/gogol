package gui

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"strings"
)

func newShortcutPrompt(c *Core, prompt string, replies ...string) *shortcutPrompt {
	if len(prompt) == 0 {
		prompt = "Prompt"
	}
	maxLen := 0
	for _, reply := range replies {
		if len(reply) > maxLen {
			maxLen = len(reply)
		}
	}
	result := &shortcutPrompt{
		core:    c,
		reply:   make(chan string),
		prompt:  strings.TrimSuffix(prompt, ":") + ":",
		replies: replies,
		input: widget.Editor{
			MaxLen:     maxLen,
			SingleLine: true,
			Submit:     true,
		},
	}
	result.style = material.Editor(theme, &result.input, "")
	return result
}

type shortcutPrompt struct {
	core    *Core
	reply   chan string
	prompt  string
	replies []string
	input   widget.Editor
	style   material.EditorStyle
}

func (p *shortcutPrompt) wait() (string, bool) {
	for {
		select {
		case reply, ok := <-p.reply:
			p.core.shortcutPrompt = nil
			return reply, ok
		}
	}
}

func (p *shortcutPrompt) Layout(gtx layout.Context) layout.Dimensions {
	if !gtx.Focused(&p.input) {
		gtx.Execute(key.FocusCmd{Tag: &p.input})
	}
	for {
		ev, ok := gtx.Event(key.Filter{Focus: &p.input, Name: key.NameEscape})
		if !ok {
			break
		}
		if kev, ok := ev.(*key.Event); ok && kev.State == key.Press && kev.Name == key.NameEscape {
			p.core.shortcutRunning = false
			close(p.reply)
		}
	}
	for {
		ev, ok := p.input.Update(gtx)
		if !ok {
			break
		}
		switch ev.(type) {
		case widget.ChangeEvent:
			s := strings.ToLower(p.input.Text())
			for _, reply := range p.replies {
				if s == strings.ToLower(reply) {
					p.reply <- p.input.Text()
				}
			}
		case widget.SubmitEvent:
			if len(p.replies) == 0 {
				p.reply <- p.input.Text()
			}
		}
	}
	pw := gtx.Constraints.Max.X - (measureText(gtx, p.prompt).Size.X - 4)
	return flexHorizontal(8,
		rigid(label(p.prompt)),
		rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.X, gtx.Constraints.Min.X = pw, pw
			return p.style.Layout(gtx)
		}),
	)(gtx)
}
