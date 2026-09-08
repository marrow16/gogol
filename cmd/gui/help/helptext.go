package help

import (
	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"gioui.org/x/styledtext"
	"image"
	"image/color"
	"time"
)

type InteractiveSpan struct {
	click gesture.Click
	span  *SpanStyle
}

func (i *InteractiveSpan) Clicked(gtx layout.Context) bool {
	if i == nil {
		return false
	}
	for {
		e, ok := i.click.Update(gtx.Source)
		if !ok {
			break
		}
		if e.Kind == gesture.KindClick {
			return true
		}
	}
	return false
}

func (i *InteractiveSpan) Layout(gtx layout.Context) layout.Dimensions {
	for {
		ok := i.Clicked(gtx)
		if !ok {
			break
		}
	}
	defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
	pointer.CursorPointer.Add(gtx.Ops)
	i.click.Add(gtx.Ops)
	return layout.Dimensions{}
}

type InteractiveState struct {
	Spans       []InteractiveSpan
	lastUpdate  time.Time
	updateIndex int
}

func (i *InteractiveState) resize(n int) {
	if n == 0 && i == nil {
		return
	}
	if cap(i.Spans) >= n {
		i.Spans = i.Spans[:n]
	} else {
		i.Spans = make([]InteractiveSpan, n)
	}
}

func (i *InteractiveState) Update(gtx layout.Context) (*SpanStyle, bool) {
	if i == nil {
		return nil, false
	}
	if i.lastUpdate != gtx.Now {
		i.lastUpdate = gtx.Now
		i.updateIndex = 0
	}
	for k := i.updateIndex; k < len(i.Spans); k++ {
		i.updateIndex = k
		span := &i.Spans[k]
		for {
			ok := span.Clicked(gtx)
			if !ok {
				break
			}
			return span.span, true
		}
	}
	return nil, false
}

type SpanStyle struct {
	Font           font.Font
	Size           unit.Sp
	Color          *color.NRGBA
	Content        string
	Interactive    bool
	Decorator      func(gtx layout.Context, theme *material.Theme, span *SpanStyle, dims layout.Dimensions)
	OnClick        func(gtx layout.Context)
	interactiveIdx int
}

type HelpTextStyle struct {
	State           *InteractiveState
	Styles          []SpanStyle
	Alignment       text.Alignment
	WrapPolicy      styledtext.WrapPolicy
	LineHeight      unit.Sp
	LineHeightScale float32
	Theme           *material.Theme
	*text.Shaper
}

func HelpText(theme *material.Theme, alignment text.Alignment, styles ...SpanStyle) HelpTextStyle {
	return HelpTextStyle{
		State:     &InteractiveState{},
		Styles:    styles,
		Shaper:    theme.Shaper,
		Theme:     theme,
		Alignment: alignment,
	}
}

func (t HelpTextStyle) Layout(gtx layout.Context) layout.Dimensions {
	for {
		span, ok := t.State.Update(gtx)
		if !ok {
			break
		}
		if span.OnClick != nil {
			span.OnClick(gtx)
		}
	}
	styles := make([]styledtext.SpanStyle, len(t.Styles))
	numInteractive := 0
	for i := range t.Styles {
		st := &t.Styles[i]
		if st.Interactive {
			st.interactiveIdx = numInteractive
			numInteractive++
		}
		sz := st.Size
		if sz == 0 {
			sz = t.Theme.TextSize
		}
		clr := t.Theme.Fg
		if st.Color != nil {
			clr = *st.Color
		}
		styles[i] = styledtext.SpanStyle{
			Font:    st.Font,
			Size:    sz,
			Color:   clr,
			Content: st.Content,
		}
	}
	t.State.resize(numInteractive)
	txt := styledtext.Text(t.Shaper, styles...)
	txt.WrapPolicy = t.WrapPolicy
	txt.Alignment = t.Alignment
	txt.LineHeight = 24       //t.LineHeight
	txt.LineHeightScale = 1.0 //t.LineHeightScale
	return txt.Layout(gtx, func(gtx layout.Context, i int, dims layout.Dimensions) {
		span := &t.Styles[i]
		if span.Decorator != nil {
			span.Decorator(gtx, t.Theme, span, dims)
		}
		if !span.Interactive {
			return
		}
		state := &t.State.Spans[span.interactiveIdx]
		state.span = span
		state.Layout(gtx)
	})
}

func underline() func(gtx layout.Context, theme *material.Theme, span *SpanStyle, dims layout.Dimensions) {
	return func(gtx layout.Context, theme *material.Theme, span *SpanStyle, dims layout.Dimensions) {
		h := max(gtx.Dp(1), 1)
		y := dims.Size.Y - h
		defer clip.Rect{
			Min: image.Point{X: 0, Y: y},
			Max: image.Point{X: dims.Size.X, Y: dims.Size.Y},
		}.Push(gtx.Ops).Pop()
		clr := theme.Fg
		if span.Color != nil {
			clr = *span.Color
		}
		paint.ColorOp{Color: clr}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	}
}

func roundedBox() func(gtx layout.Context, theme *material.Theme, span *SpanStyle, dims layout.Dimensions) {
	return func(gtx layout.Context, theme *material.Theme, span *SpanStyle, dims layout.Dimensions) {
		width := float32(max(gtx.Dp(1), 1))
		radius := gtx.Dp(4)
		rr := clip.RRect{
			Rect: image.Rectangle{
				Min: image.Point{X: -radius, Y: 0},
				Max: image.Point{X: dims.Size.X + radius, Y: dims.Size.Y},
			},
			NE: radius,
			NW: radius,
			SE: radius,
			SW: radius,
		}
		clr := theme.Fg
		if span.Color != nil {
			clr = *span.Color
		}
		defer clip.Stroke{
			Path:  rr.Path(gtx.Ops),
			Width: width,
		}.Op().Push(gtx.Ops).Pop()
		paint.ColorOp{Color: clr}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	}
}

func codeBackground() func(gtx layout.Context, theme *material.Theme, span *SpanStyle, dims layout.Dimensions) {
	return func(gtx layout.Context, theme *material.Theme, span *SpanStyle, dims layout.Dimensions) {
		defer clip.Rect{
			Min: image.Point{},
			Max: dims.Size,
		}.Push(gtx.Ops).Pop()
		paint.ColorOp{
			Color: color.NRGBA{R: 128, G: 128, B: 128, A: 80},
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	}
}

func buttonDecorator() func(gtx layout.Context, theme *material.Theme, span *SpanStyle, dims layout.Dimensions) {
	return func(gtx layout.Context, theme *material.Theme, span *SpanStyle, dims layout.Dimensions) {
		radius := gtx.Dp(4)
		width := float32(max(gtx.Dp(1), 1))
		rr := clip.RRect{
			Rect: image.Rectangle{
				Min: image.Point{X: -radius, Y: 0},
				Max: image.Point{
					X: dims.Size.X + radius,
					Y: dims.Size.Y,
				},
			},
			NE: radius,
			NW: radius,
			SE: radius,
			SW: radius,
		}
		// background
		defer rr.Push(gtx.Ops).Pop()
		paint.ColorOp{
			Color: color.NRGBA{R: 128, G: 128, B: 128, A: 80},
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		// border
		clr := theme.Fg
		if span.Color != nil {
			clr = *span.Color
		}
		defer clip.Stroke{
			Path:  rr.Path(gtx.Ops),
			Width: width,
		}.Op().Push(gtx.Ops).Pop()
		paint.ColorOp{Color: clr}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	}
}

func imageDecorator(img image.Image) func(layout.Context, *material.Theme, *SpanStyle, layout.Dimensions) {
	imgOp := paint.NewImageOp(img)
	return func(gtx layout.Context, _ *material.Theme, _ *SpanStyle, dims layout.Dimensions) {
		src := img.Bounds().Size()
		if src.X == 0 || src.Y == 0 {
			return
		}
		sx := float32(dims.Size.X) / float32(src.X)
		sy := float32(dims.Size.Y) / float32(src.Y)
		scale := min(sx, sy)
		w := float32(src.X) * scale
		h := float32(src.Y) * scale
		x := (float32(dims.Size.X) - w) / 2
		y := (float32(dims.Size.Y) - h) / 2
		defer op.Affine(f32.Affine2D{}.
			Offset(f32.Pt(x, y)).
			Scale(f32.Point{}, f32.Pt(scale, scale)),
		).Push(gtx.Ops).Pop()
		imgOp.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	}
}
