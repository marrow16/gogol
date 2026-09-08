package help

import (
	"fmt"
	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
	"image/color"
	"io"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"sync"
)

func Show(topic Topic) {
	helpHolder.show(topic)
}

func HelpWindow() *app.Window {
	return helpHolder.window
}

var (
	linkColor = color.NRGBA{R: 102, G: 128, B: 230, A: 255}
	lineColor = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
	isMac     = runtime.GOOS == "darwin"
)

const (
	altMac = "⌥"
	cmdMac = "⌘"
)

var helpHolder = &holder{
	list:        widget.List{Axis: layout.Vertical},
	resultsList: widget.List{Axis: layout.Vertical},
	contents:    make(map[Topic]layout.Widget),
}

type holder struct {
	mutex    sync.Mutex
	window   *app.Window
	theme    *material.Theme
	fonts    []font.FontFace
	mono     font.Font
	topic    Topic
	contents map[Topic]layout.Widget
	list     widget.List
	index    widget.Clickable
	back     widget.Clickable
	backs    []Topic
	// searching...
	searching         bool
	search            widget.Clickable
	exit              widget.Clickable
	searchEdit        widget.Editor
	results           []Topic
	resultsList       widget.List
	resultsClickables []widget.Clickable
	contentsIndex     map[Topic]string
}

func (h *holder) show(topic Topic) {
	h.mutex.Lock()
	if topic > -1 {
		if topic != h.topic && h.topic != Index {
			h.backs = append(h.backs, h.topic)
		}
		h.topic = topic
		h.list.ScrollTo(0)
	}
	h.searching = false
	if h.window == nil {
		h.window = new(app.Window)
		h.fonts = gofont.Collection()
		h.theme = material.NewTheme()
		h.theme.Shaper = text.NewShaper(text.WithCollection(h.fonts))
		h.mono = h.fonts[0].Font
		for _, f := range h.fonts {
			if f.Font.Typeface == "Go Mono" && f.Font.Weight == font.Normal && f.Font.Style == font.Regular {
				h.mono = f.Font
				break
			}
		}
		h.window.Option(
			app.Title("GoGoL Help"),
			app.Size(800, 600),
		)
		h.mutex.Unlock()
		go h.run()
		return
	}
	w := h.window
	h.mutex.Unlock()
	w.Invalidate()
}

func (h *holder) run() {
	defer func() {
		h.mutex.Lock()
		h.window = nil
		h.mutex.Unlock()
	}()
	for {
		switch e := h.window.Event().(type) {
		case app.DestroyEvent:
			return
		case app.FrameEvent:
			var ops op.Ops
			gtx := app.NewContext(&ops, e)
			h.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (h *holder) layout(gtx layout.Context) {
	layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		h.header(gtx),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if h.searching {
				return layout.Inset{Top: 8, Bottom: 8, Left: 8, Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return material.List(h.theme, &h.resultsList).Layout(gtx, len(h.results), func(gtx layout.Context, index int) layout.Dimensions {
						btn := &h.resultsClickables[index]
						if btn.Clicked(gtx) {
							newTopic := h.results[index]
							if newTopic != h.topic {
								h.backs = append(h.backs, h.topic)
							}
							h.topic = newTopic
							h.searching = false
							h.window.Invalidate()
						}
						return linkLabel(btn, h.theme, h.results[index].String())(gtx)
					})
				})
			}
			return material.List(h.theme, &h.list).Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
				return layout.Inset{Top: 8, Bottom: 8, Left: 8, Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return h.content()(gtx)
				})
			})
		}),
	)
}

func (h *holder) content() layout.Widget {
	if c, ok := h.contents[h.topic]; ok {
		return c
	}
	if src, ok := contents[h.topic]; ok && len(src) > 0 {
		c := buildContent(h, 0, src)
		h.contents[h.topic] = c
		return c
	}
	ts := HelpText(h.theme, 0, SpanStyle{
		Content: "No help available",
		Font:    font.Font{Style: font.Italic},
	})
	return ts.Layout
}

type asWidget interface {
	widget(h *holder) layout.Widget
}

func buildContent(h *holder, alignment text.Alignment, src content) layout.Widget {
	type contentRow struct {
		widget layout.Widget
		spans  []SpanStyle
	}
	var rows []*contentRow
	var currentRow *contentRow
	newRow := func() {
		currentRow = &contentRow{}
		rows = append(rows, currentRow)
	}
	newRow()
	for _, s := range src {
		switch st := s.(type) {
		case separator:
			rows = append(rows, &contentRow{
				widget: func(gtx layout.Context) layout.Dimensions {
					overallHt := gtx.Sp(8) + gtx.Dp(st.spaceBefore+st.spaceAfter)
					lt := gtx.Sp(4) + gtx.Dp(st.spaceBefore)
					wd := gtx.Constraints.Max.X
					paint.FillShape(gtx.Ops, lineColor, clip.Rect{
						Min: image.Point{X: 0, Y: lt},
						Max: image.Point{X: wd, Y: lt + 1},
					}.Op())
					return layout.Dimensions{Size: image.Point{X: wd, Y: overallHt}}
				},
			})
			newRow()
		case asWidget:
			rows = append(rows, &contentRow{
				widget: st.widget(h),
			})
			newRow()
		case bold:
			currentRow.spans = append(currentRow.spans, SpanStyle{Content: string(st), Font: font.Font{Weight: font.Bold}})
		case italic:
			currentRow.spans = append(currentRow.spans, SpanStyle{Content: string(st), Font: font.Font{Style: font.Italic}})
		case boldItalic:
			currentRow.spans = append(currentRow.spans, SpanStyle{Content: string(st), Font: font.Font{Weight: font.Bold, Style: font.Italic}})
		case boxed:
			currentRow.spans = append(currentRow.spans,
				SpanStyle{Content: " ", Size: 24},
				SpanStyle{Content: strings.TrimSpace(string(st)), Decorator: roundedBox()},
				SpanStyle{Content: " ", Size: 24},
			)
		case button:
			currentRow.spans = append(currentRow.spans,
				SpanStyle{Content: " ", Size: 24},
				SpanStyle{Content: strings.TrimSpace(string(st)), Decorator: buttonDecorator()},
				SpanStyle{Content: " ", Size: 24},
			)
		case code:
			currentRow.spans = append(currentRow.spans, SpanStyle{
				Content:   string(st),
				Font:      h.mono,
				Size:      15,
				Decorator: codeBackground(),
			})
		case codeCopyable:
			currentRow.spans = append(currentRow.spans, SpanStyle{
				Content:     string(st),
				Font:        h.mono,
				Size:        15,
				Decorator:   codeBackground(),
				Interactive: true,
				OnClick: func(gtx layout.Context) {
					gtx.Execute(clipboard.WriteCmd{
						Type: "text/plain",
						Data: io.NopCloser(strings.NewReader(string(st))),
					})
				},
			})
		case string:
			currentRow.spans = append(currentRow.spans, SpanStyle{Content: st})
		case keys:
			currentRow.spans = append(currentRow.spans,
				SpanStyle{Content: " ", Size: 24},
				SpanStyle{Content: st.String(), Decorator: roundedBox()},
				SpanStyle{Content: " ", Size: 24},
			)
		case Topic:
			currentRow.spans = append(currentRow.spans, SpanStyle{
				Content:     st.String(),
				Color:       &linkColor,
				Interactive: true,
				Decorator:   underline(),
				OnClick: func(gtx layout.Context) {
					if h.topic != st {
						h.backs = append(h.backs, h.topic)
					}
					h.list.ScrollTo(0)
					h.topic = st
					h.window.Invalidate()
				},
			})
		case topicLink:
			txt := st.Text
			if txt == "" {
				txt = st.Topic.String()
			}
			ss := SpanStyle{
				Content:     txt,
				Color:       &linkColor,
				Interactive: true,
				Decorator:   underline(),
				OnClick: func(gtx layout.Context) {
					if h.topic != st.Topic {
						h.backs = append(h.backs, h.topic)
					}
					h.list.ScrollTo(0)
					h.topic = st.Topic
					h.window.Invalidate()
				},
			}
			if st.Italic {
				ss.Font = font.Font{Style: font.Italic}
			}
			currentRow.spans = append(currentRow.spans, ss)
		case externalLink:
			txt := st.Text
			if txt == "" {
				txt = st.Url
			}
			currentRow.spans = append(currentRow.spans, SpanStyle{
				Content:     txt,
				Color:       &linkColor,
				Interactive: true,
				Decorator:   underline(),
				OnClick: func(gtx layout.Context) {
					_ = openURL(st.Url)
				},
			})
		case iconText:
			currentRow.spans = append(currentRow.spans, SpanStyle{
				Content:   "  ",
				Size:      32,
				Decorator: imageDecorator(st.image),
			})
		}
	}
	rigidRows := make([]layout.FlexChild, 0, len(rows))
	for _, r := range rows {
		switch {
		case r.widget != nil:
			rigidRows = append(rigidRows, layout.Rigid(r.widget))
		case len(r.spans) > 0:
			ht := HelpText(h.theme, alignment, r.spans...)
			htw := ht.Layout
			rigidRows = append(rigidRows, layout.Rigid(htw))
		}
	}
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Gap: 4}.Layout(gtx,
			rigidRows...,
		)
	}
}

func (h *holder) header(gtx layout.Context) layout.FlexChild {
	if h.searching {
		return h.searchHeader(gtx)
	}
	if h.index.Clicked(gtx) {
		h.backs = append(h.backs, h.topic)
		h.topic = Index
		h.window.Invalidate()
	}
	if h.back.Clicked(gtx) && len(h.backs) > 0 {
		h.topic = h.backs[len(h.backs)-1]
		h.backs = h.backs[:len(h.backs)-1]
		h.window.Invalidate()
	}
	if h.search.Clicked(gtx) {
		h.searching = true
		h.window.Invalidate()
	}
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		dims := layout.Inset{Top: 8, Bottom: 8, Left: 8, Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(h.theme, (h.theme.TextSize*5)/4, h.topic.String())
					lbl.MaxLines = 1
					lbl.Font.Weight = font.Bold
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if h.topic != Index {
								return linkLabel(&h.index, h.theme, "Index")(gtx)
							}
							return layout.Dimensions{}
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if len(h.backs) == 0 {
								return layout.Dimensions{}
							}
							return linkLabel(&h.back, h.theme, "Back")(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return linkLabel(&h.search, h.theme, "Search")(gtx)
						}),
					)
				}),
			)
		})
		y := dims.Size.Y - 2
		defer clip.Rect{
			Min: image.Point{X: 0, Y: y},
			Max: image.Point{X: dims.Size.X, Y: y + 2},
		}.Push(gtx.Ops).Pop()
		paint.ColorOp{Color: h.theme.Palette.Fg}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		return dims
	})
}

func (h *holder) doSearch(gtx layout.Context, s string) {
	h.resultsList.ScrollTo(0)
	if s = strings.TrimSpace(s); len(s) > 1 {
		h.buildSearchIndex()
		i := 0
		candidates := make(map[Topic]struct{})
		for wd := range strings.SplitSeq(s, " ") {
			if wd = strings.ToLower(strings.TrimSpace(wd)); len(wd) > 1 {
				for t, tstr := range h.contentsIndex {
					if strings.Contains(tstr, wd) {
						if i == 0 {
							candidates[t] = struct{}{}
						}
					} else if i > 0 {
						delete(candidates, t)
					}
				}
				i++
			}
		}
		h.results = make([]Topic, 0, len(candidates))
		for t := range candidates {
			h.results = append(h.results, t)
		}
		slices.SortFunc(h.results, func(a, b Topic) int {
			return strings.Compare(a.String(), b.String())
		})
		h.resultsClickables = make([]widget.Clickable, len(h.results))
	}
	h.window.Invalidate()
}

func (h *holder) buildSearchIndex() {
	if h.contentsIndex == nil {
		h.contentsIndex = make(map[Topic]string, len(contents))
		for t, c := range contents {
			var sb strings.Builder
			sb.WriteString(" ")
			for _, i := range c {
				switch itm := i.(type) {
				case string:
					sb.WriteString(strings.ToLower(itm))
					sb.WriteString(" ")
				case fmt.Stringer:
					sb.WriteString(strings.ToLower(itm.String()))
					sb.WriteString(" ")
				default:
					sb.WriteString(strings.ToLower(fmt.Sprintf("%v", itm)))
					sb.WriteString(" ")
				}
			}
			h.contentsIndex[t] = sb.String()
		}
	}
}

func (h *holder) searchHeader(gtx layout.Context) layout.FlexChild {
	gtx.Execute(key.FocusCmd{Tag: &h.searchEdit})
	if h.exit.Clicked(gtx) {
		h.searching = false
		h.window.Invalidate()
	}
	for {
		ev, ok := h.searchEdit.Update(gtx)
		if !ok {
			break
		}
		if _, ok = ev.(widget.ChangeEvent); ok {
			h.doSearch(gtx, h.searchEdit.Text())
		}
	}
	for {
		ev, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		if evt, ok := ev.(key.Event); ok && evt.State == key.Press && evt.Name == key.NameEscape {
			h.searching = false
			h.window.Invalidate()
		}
	}
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		dims := layout.Inset{Top: 8, Bottom: 8, Left: 8, Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 10}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: 3}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(h.theme, h.theme.TextSize, "Search:")
								lbl.MaxLines = 1
								lbl.Font.Weight = font.Bold
								return lbl.Layout(gtx)
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = gtx.Constraints.Max.X - gtx.Sp(40)
							return widget.Border{Color: h.theme.Fg, CornerRadius: 3, Width: 1}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									editStyle := material.Editor(h.theme, &h.searchEdit, "")
									return editStyle.Layout(gtx)
								})
							})
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Clickable(gtx, &h.exit, func(gtx layout.Context) layout.Dimensions {
						return material.Label(h.theme, (h.theme.TextSize*5)/4, "ⓧ").Layout(gtx)
					})
				}),
			)
		})
		y := dims.Size.Y - 2
		defer clip.Rect{
			Min: image.Point{X: 0, Y: y},
			Max: image.Point{X: dims.Size.X, Y: y + 2},
		}.Push(gtx.Ops).Pop()
		paint.ColorOp{Color: h.theme.Palette.Fg}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		return dims
	})
}

func linkLabel(btn *widget.Clickable, theme *material.Theme, s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return material.Clickable(gtx, btn, func(gtx layout.Context) layout.Dimensions {
			return layout.Stack{}.Layout(gtx,
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(theme, theme.TextSize, s)
					lbl.MaxLines = 1
					lbl.Color = linkColor
					pointer.CursorPointer.Add(gtx.Ops)
					return lbl.Layout(gtx)
				}),
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					thickness := gtx.Dp(1)
					rect := clip.Rect{
						Min: image.Point{X: 0, Y: gtx.Constraints.Min.Y - thickness},
						Max: image.Point{X: gtx.Constraints.Min.X, Y: gtx.Constraints.Min.Y},
					}
					defer rect.Push(gtx.Ops).Pop()
					paint.ColorOp{Color: linkColor}.Add(gtx.Ops)
					paint.PaintOp{}.Add(gtx.Ops)
					return layout.Dimensions{Size: gtx.Constraints.Min}
				}),
			)
		})
	}
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
