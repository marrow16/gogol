package gui

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
)

const separatorOption = "--"

func newShortcutPromptMenu(c *Core, prompt string, selected int, options ...string) *shortcutPromptMenu {
	if len(prompt) == 0 {
		prompt = "Prompt"
	}
	options = cleanOptions(options)
	actualOptions := make([]string, 0, len(options))
	indices := make(map[int]int, len(options))
	length := 0
	for i, option := range options {
		if option != separatorOption {
			indices[i] = len(actualOptions)
			actualOptions = append(actualOptions, option)
			length++
		} else {
			indices[i] = -1
			length++
		}
	}
	if len(actualOptions) == 0 {
		// no items...
		return nil
	}
	result := &shortcutPromptMenu{
		core:         c,
		reply:        make(chan int),
		selected:     max(0, min(selected, len(actualOptions)-1)),
		prompt:       prompt,
		list:         widget.List{Axis: layout.Vertical},
		options:      actualOptions,
		optionClicks: make([]widget.Clickable, len(actualOptions)),
		itemIndices:  indices,
		length:       length,
	}
	result.listStyle = material.List(theme, &result.list)
	result.listStyle.AnchorStrategy = material.Overlay
	return result
}

func cleanOptions(s []string) []string {
	result := make([]string, 0, len(s))
	for _, v := range s {
		if v != separatorOption || len(result) == 0 || result[len(result)-1] != separatorOption {
			result = append(result, v)
		}
	}
	for len(result) > 0 && result[0] == separatorOption {
		result = result[1:]
	}
	for len(result) > 0 && result[len(result)-1] == separatorOption {
		result = result[:len(result)-1]
	}
	return result
}

type shortcutPromptMenu struct {
	core         *Core
	reply        chan int
	prompt       string
	selected     int
	list         widget.List
	listStyle    material.ListStyle
	options      []string
	optionClicks []widget.Clickable
	length       int
	itemIndices  map[int]int
}

func (p *shortcutPromptMenu) wait() (int, bool) {
	for {
		select {
		case reply, ok := <-p.reply:
			p.core.shortcutPromptMenu = nil
			return reply, ok
		}
	}
}

func (p *shortcutPromptMenu) Layout(gtx layout.Context) layout.Dimensions {
	for {
		ev, ok := gtx.Event(
			key.Filter{Name: key.NameReturn},
			key.Filter{Name: key.NameEnter},
			key.Filter{Name: key.NameSpace},
			key.Filter{Name: key.NameUpArrow},
			key.Filter{Name: key.NameDownArrow},
			key.Filter{Name: key.NameHome},
			key.Filter{Name: key.NameEnd})
		if !ok {
			break
		}
		if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
			switch kev.Name {
			case key.NameEnter, key.NameReturn, key.NameSpace:
				p.reply <- p.selected
			case key.NameUpArrow:
				p.selected--
				if p.selected < 0 {
					p.selected = len(p.options) - 1
				}
			case key.NameDownArrow:
				p.selected++
				if p.selected >= len(p.options) {
					p.selected = 0
				}
			case key.NameHome:
				p.selected = 0
			case key.NameEnd:
				p.selected = len(p.options) - 1
			}
		}
	}
	macro := op.Record(gtx.Ops)
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	itemHt := measureText(gtx, "Xy").Size.Y
	dims := p.listStyle.Layout(gtx, p.length, func(gtx layout.Context, index int) layout.Dimensions {
		idx := p.itemIndices[index]
		if idx == -1 {
			// separator...
			dims := measureText(gtx, "-")
			dims.Size.X = gtx.Constraints.Max.X
			dims.Size.Y = (dims.Size.Y * 2) / 3
			paint.FillShape(gtx.Ops, borderColor, clip.Rect{Min: image.Point{X: 0, Y: dims.Size.Y / 2}, Max: image.Point{X: dims.Size.X, Y: (dims.Size.Y / 2) + 1}}.Op())
			return dims
		}
		if p.optionClicks[idx].Clicked(gtx) {
			p.reply <- idx
		}
		if idx == p.selected {
			paint.FillShape(gtx.Ops, selectedFocusedColor, clip.Rect{Max: image.Point{X: gtx.Constraints.Max.X, Y: itemHt}}.Op())
		}
		return material.Clickable(gtx, &p.optionClicks[idx], inset(0, 0, 3, 3, label(p.options[idx])))
	})
	call := macro.Stop()
	paint.FillShape(gtx.Ops, backgroundColor, clip.Rect{Max: dims.Size}.Op())
	borders(gtx, dims, borderColor, true, true, false, true)
	call.Add(gtx.Ops)
	return dims
}
