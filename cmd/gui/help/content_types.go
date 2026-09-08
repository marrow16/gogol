package help

import (
	"fmt"
	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/clipboard"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/cmd/gui/help/images"
	"image"
	"image/color"
	"io"
	"strings"
)

type content []any

func (c content) String() string {
	var sb strings.Builder
	for _, i := range c {
		switch itm := i.(type) {
		case string:
			sb.WriteString(itm)
			sb.WriteString(" ")
		case fmt.Stringer:
			sb.WriteString(itm.String())
			sb.WriteString(" ")
		default:
			sb.WriteString(fmt.Sprintf("%v", i))
			sb.WriteString(" ")
		}
	}
	return sb.String()
}

type topicLink struct {
	Text   string
	Topic  Topic
	Italic bool
}

func (t topicLink) String() string {
	return t.Text + " " + t.Topic.String()
}

func url(url string, s ...string) externalLink {
	if len(s) > 0 && s[0] != "" {
		return externalLink{
			Url:  url,
			Text: s[0],
		}
	}
	return externalLink{Url: url}
}

type externalLink struct {
	Text string
	Url  string
}

func (e externalLink) String() string {
	return e.Text + " " + e.Url
}

type iconText struct {
	image image.Image
}

func (i iconText) String() string {
	return ""
}

type bold string
type italic string
type boldItalic string
type boxed string
type button string
type code string
type codeCopyable string

type keys []string

func (k keys) String() string {
	var b strings.Builder
	if isMac {
		for _, s := range k {
			b.WriteString(s)
		}
	} else {
		lastPlus := false
		for _, s := range k {
			if b.Len() > 0 && !lastPlus {
				b.WriteString("+")
			}
			lastPlus = strings.HasSuffix(s, "+")
			switch s {
			case altMac:
				b.WriteString("Alt")
			case cmdMac:
				b.WriteString("Ctrl")
			default:
				b.WriteString(s)
			}
		}
	}
	return b.String()
}

type hanging struct {
	indent    unit.Dp
	gap       int
	prefix    any
	content   any
	alignment text.Alignment
	spacing
}

func (a hanging) widget(h *holder) layout.Widget {
	var left layout.Widget
	var right layout.Widget
	switch ct := a.prefix.(type) {
	case string:
		left = buildContent(h, a.alignment, content{ct})
	case content:
		left = buildContent(h, a.alignment, ct)
	case []any:
		left = buildContent(h, a.alignment, ct)
	default:
		left = func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{}
		}
	}
	switch ct := a.content.(type) {
	case string:
		right = buildContent(h, 0, content{ct})
	case content:
		right = buildContent(h, 0, ct)
	case []any:
		right = buildContent(h, 0, ct)
	default:
		right = func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{}
		}
	}
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: a.indent, Top: a.spaceBefore, Bottom: a.spaceAfter}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Gap: a.gap}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return left(gtx)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return right(gtx)
				}),
			)
		})
	}
}

func (a hanging) String() string {
	var sb strings.Builder
	switch ct := a.prefix.(type) {
	case string:
		sb.WriteString(ct)
	case content:
		sb.WriteString(ct.String())
	case []any:
		sb.WriteString(content(ct).String())
	}
	sb.WriteString(" ")
	switch ct := a.content.(type) {
	case string:
		sb.WriteString(ct)
	case content:
		sb.WriteString(ct.String())
	case []any:
		sb.WriteString(content(ct).String())
	}
	return sb.String()
}

type indent struct {
	content any
	indent  unit.Dp
	spacing
}

func (i indent) widget(h *holder) layout.Widget {
	var c layout.Widget
	switch ct := i.content.(type) {
	case string:
		c = buildContent(h, 0, content{ct})
	case content:
		c = buildContent(h, 0, ct)
	case []any:
		c = buildContent(h, 0, ct)
	default:
		c = func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{}
		}
	}
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: i.indent, Top: i.spaceBefore, Bottom: i.spaceAfter}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return c(gtx)
		})
	}
}

func (i indent) String() string {
	switch ct := i.content.(type) {
	case string:
		return ct
	case content:
		return ct.String()
	case []any:
		return content(ct).String()
	}
	return ""
}

type Image struct {
	name          string
	width, height int
	indent        unit.Dp
	spacing
}

func (i Image) widget(h *holder) layout.Widget {
	img, err := images.LoadImage(i.name)
	if err != nil {
		return func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: i.indent, Top: i.spaceBefore, Bottom: i.spaceAfter}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(h.theme, h.theme.TextSize, fmt.Sprintf("Image %q error: %s", i.name, err.Error()))
				return lbl.Layout(gtx)
			})
		}
	}
	imgOp := paint.NewImageOp(img)
	src := img.Bounds().Size()
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: i.indent, Top: i.spaceBefore, Bottom: i.spaceAfter}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			size := src
			switch {
			case i.width != 0 && i.height != 0:
				size = image.Point{
					X: i.width,
					Y: i.height,
				}
			case i.width != 0 && i.width != src.X:
				size.X = i.width
				size.Y = src.Y * i.width / src.X
			case i.height != 0 && i.height != src.Y:
				size.Y = i.height
				size.X = src.X * i.height / src.Y
			}
			sx := float32(size.X) / float32(src.X)
			sy := float32(size.Y) / float32(src.Y)
			stack := clip.Rect{Max: size}.Push(gtx.Ops)
			defer stack.Pop()
			transform := op.Affine(
				f32.Affine2D{}.
					Scale(f32.Point{}, f32.Pt(sx, sy)),
			).Push(gtx.Ops)
			defer transform.Pop()
			imgOp.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			return layout.Dimensions{Size: size}
		})
	}
}

type codeBlock struct {
	code   string
	indent unit.Dp
	spacing
}

func (cb codeBlock) widget(h *holder) layout.Widget {
	btn := widget.Clickable{}
	return func(gtx layout.Context) layout.Dimensions {
		if btn.Clicked(gtx) {
			gtx.Execute(clipboard.WriteCmd{
				Type: "text/plain",
				Data: io.NopCloser(strings.NewReader(cb.code)),
			})
		}
		return layout.Inset{Left: cb.indent, Top: cb.spaceBefore, Bottom: cb.spaceAfter}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return material.Clickable(gtx, &btn, func(gtx layout.Context) layout.Dimensions {
				// Record the complete padded contents.
				macro := op.Record(gtx.Ops)
				dims := layout.UniformInset(4).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(
						h.theme,
						(h.theme.TextSize*4)/5,
						strings.ReplaceAll(cb.code, "\t", "    "),
					)
					lbl.Font = h.mono
					return lbl.Layout(gtx)
				})
				call := macro.Stop()
				radius := gtx.Dp(4)
				rr := clip.RRect{
					Rect: image.Rectangle{
						Max: dims.Size,
					},
					NW: radius,
					NE: radius,
					SW: radius,
					SE: radius,
				}
				// background
				{
					stack := rr.Push(gtx.Ops)
					paint.ColorOp{
						Color: color.NRGBA{R: 128, G: 128, B: 128, A: 40},
					}.Add(gtx.Ops)
					paint.PaintOp{}.Add(gtx.Ops)
					stack.Pop()
				}
				// border
				{
					stack := clip.Stroke{
						Path:  rr.Path(gtx.Ops),
						Width: float32(max(gtx.Dp(1), 1)),
					}.Op().Push(gtx.Ops)
					paint.ColorOp{Color: lineColor}.Add(gtx.Ops)
					paint.PaintOp{}.Add(gtx.Ops)
					stack.Pop()
				}
				// padded code contents
				call.Add(gtx.Ops)
				// pointer cursor over the whole code block
				{
					area := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
					pointer.CursorPointer.Add(gtx.Ops)
					area.Pop()
				}
				return dims
			})
		})
	}
}

func (cb codeBlock) String() string {
	return cb.code
}

type header struct {
	level int
	text  string
	mono  bool
	spacing
}

func (hdr header) widget(h *holder) layout.Widget {
	sz := h.theme.TextSize
	switch hdr.level {
	case 1:
		sz = sz * 1.75
	case 2:
		sz = sz * 1.5
	case 3:
		sz = sz * 1.3
	case 4:
		sz = sz * 1.15
	case 5:
	default:
		sz = sz * 0.9
	}
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: hdr.spaceBefore, Bottom: hdr.spaceAfter}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(h.theme, sz, hdr.text)
			if hdr.mono {
				lbl.Font = h.mono
			}
			lbl.Font.Weight = font.Bold
			lbl.MaxLines = 1
			return lbl.Layout(gtx)
		})
	}
}

func (hdr header) String() string {
	return fmt.Sprintf("%v", hdr.text)
}

func h(s string, sp ...unit.Dp) header {
	if len(sp) > 1 {
		return header{
			level:       5,
			text:        s,
			spaceBefore: sp[0],
			spaceAfter:  sp[1],
		}
	} else if len(sp) > 0 {
		return header{
			level:       5,
			text:        s,
			spaceBefore: sp[0],
		}
	}
	return header{
		level: 5,
		text:  s,
	}
}

func h1(s string, sp ...unit.Dp) header {
	if len(sp) > 1 {
		return header{
			level:       1,
			text:        s,
			spaceBefore: sp[0],
			spaceAfter:  sp[1],
		}
	} else if len(sp) > 0 {
		return header{
			level:       1,
			text:        s,
			spaceBefore: sp[0],
		}
	}
	return header{
		level: 1,
		text:  s,
	}
}

func h2(s string, sp ...unit.Dp) header {
	if len(sp) > 1 {
		return header{
			level:       2,
			text:        s,
			spaceBefore: sp[0],
			spaceAfter:  sp[1],
		}
	} else if len(sp) > 0 {
		return header{
			level:       2,
			text:        s,
			spaceBefore: sp[0],
		}
	}
	return header{
		level: 2,
		text:  s,
	}
}

func h3(s string, sp ...unit.Dp) header {
	if len(sp) > 1 {
		return header{
			level:       3,
			text:        s,
			spaceBefore: sp[0],
			spaceAfter:  sp[1],
		}
	} else if len(sp) > 0 {
		return header{
			level:       3,
			text:        s,
			spaceBefore: sp[0],
		}
	}
	return header{
		level: 3,
		text:  s,
	}
}

func h4(s string, sp ...unit.Dp) header {
	if len(sp) > 1 {
		return header{
			level:       4,
			text:        s,
			spaceBefore: sp[0],
			spaceAfter:  sp[1],
		}
	} else if len(sp) > 0 {
		return header{
			level:       4,
			text:        s,
			spaceBefore: sp[0],
		}
	}
	return header{
		level: 4,
		text:  s,
	}
}

func h5(s string, sp ...unit.Dp) header {
	if len(sp) > 1 {
		return header{
			level:       5,
			text:        s,
			spaceBefore: sp[0],
			spaceAfter:  sp[1],
		}
	} else if len(sp) > 0 {
		return header{
			level:       5,
			text:        s,
			spaceBefore: sp[0],
		}
	}
	return header{
		level: 5,
		text:  s,
	}
}

func h6(s string, sp ...unit.Dp) header {
	if len(sp) > 1 {
		return header{
			level:       6,
			text:        s,
			spaceBefore: sp[0],
			spaceAfter:  sp[1],
		}
	} else if len(sp) > 0 {
		return header{
			level:       6,
			text:        s,
			spaceBefore: sp[0],
		}
	}
	return header{
		level: 6,
		text:  s,
	}
}

type separator struct {
	spacing
}

func (s separator) String() string {
	return ""
}

type spacing struct {
	spaceBefore unit.Dp
	spaceAfter  unit.Dp
}

const (
	keyLeft         = "←"
	keyRight        = "→"
	keyUp           = " ↑ "
	keyDown         = " ↓ "
	keyUpTriangle   = "▲"
	keyDownTriangle = "▼"
	keyTab          = "⇥"
	keyEnter        = "↩"
	keyBack         = "⌫"
)
