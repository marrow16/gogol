package controls

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
)

func NewSlider(theme *material.Theme, axis layout.Axis, initial float32) *Slider {
	float := &widget.Float{Value: initial}
	return &Slider{
		theme:      theme,
		value:      float,
		Axis:       axis,
		fingerSize: theme.FingerSize,
	}
}

type Slider struct {
	theme      *material.Theme
	value      *widget.Float
	Axis       layout.Axis
	fingerSize unit.Dp
	focused    bool
}

func (s *Slider) Layout(gtx layout.Context) layout.Dimensions {
	const thumbRadius unit.Dp = 6
	tr := gtx.Dp(thumbRadius)
	trackWidth := gtx.Dp(2)
	axis := s.Axis
	minLength := tr + 3*tr + tr // keep a minimum length so that the track is always visible
	// try to expand to finger size, but only if the constraints allow for it...
	touchSizePx := min(gtx.Dp(s.fingerSize), axis.Convert(gtx.Constraints.Max).Y)
	sizeMain := max(axis.Convert(gtx.Constraints.Min).X, minLength)
	sizeCross := max(2*tr, touchSizePx)
	size := axis.Convert(image.Pt(sizeMain, sizeCross))

	// define the clipping area and register the tag
	stack := clip.Rect(image.Rectangle{Max: size}).Push(gtx.Ops)
	event.Op(gtx.Ops, s)

	o := axis.Convert(image.Pt(tr, 0))
	trans := op.Offset(o).Push(gtx.Ops)
	gtx.Constraints.Min = axis.Convert(image.Pt(sizeMain-2*tr, sizeCross))
	dims := s.value.Layout(gtx, axis, thumbRadius)
	gtx.Constraints.Min = gtx.Constraints.Min.Add(axis.Convert(image.Pt(0, sizeCross)))
	thumbPos := tr + int(s.value.Value*float32(axis.Convert(dims.Size).X))
	trans.Pop()

	if s.focused {
		// draw focus rect...
		bc, bt := Style.bordering(true)
		widget.Border{Color: bc, CornerRadius: 3, Width: bt}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: size}
		})
	}

	rect := func(minx, miny, maxx, maxy int) image.Rectangle {
		r := image.Rect(minx, miny, maxx, maxy)
		if axis == layout.Vertical {
			r.Max.X, r.Min.X = sizeMain-r.Min.X, sizeMain-r.Max.X
		}
		r.Min = axis.Convert(r.Min)
		r.Max = axis.Convert(r.Max)
		return r
	}
	// draw track before thumb...
	track := rect(
		tr, sizeCross/2-trackWidth/2,
		thumbPos, sizeCross/2+trackWidth/2,
	)
	clr := Style.ButtonColor
	if s.focused {
		clr = s.theme.Palette.ContrastBg
	}
	paint.FillShape(gtx.Ops, clr, clip.Rect(track).Op())
	// draw track after thumb....
	track = rect(
		thumbPos, axis.Convert(track.Min).Y,
		sizeMain-tr, axis.Convert(track.Max).Y,
	)
	paint.FillShape(gtx.Ops, Style.BorderColor, clip.Rect(track).Op())
	// draw thumb...
	pt := image.Pt(thumbPos, sizeCross/2)
	thumb := rect(
		pt.X-tr, pt.Y-tr,
		pt.X+tr, pt.Y+tr,
	)
	paint.FillShape(gtx.Ops, clr, clip.Ellipse(thumb).Op(gtx.Ops))
	// end...
	stack.Pop()
	return layout.Dimensions{Size: size}
}

func (s *Slider) Update(gtx layout.Context) bool {
	changed := s.value.Update(gtx)
	minus := key.NameLeftArrow
	plus := key.NameRightArrow
	if s.Axis == layout.Vertical {
		minus = key.NameDownArrow
		plus = key.NameUpArrow
	}
	for {
		ev, ok := gtx.Event(
			pointer.Filter{Target: s, Kinds: pointer.Press},
			key.FocusFilter{Target: s},
			key.Filter{Focus: s, Name: key.NameHome},
			key.Filter{Focus: s, Name: key.NameEnd},
			key.Filter{Focus: s, Name: key.NamePageUp},
			key.Filter{Focus: s, Name: key.NamePageDown},
			key.Filter{Focus: s, Name: minus},
			key.Filter{Focus: s, Name: plus},
		)
		if !ok {
			break
		}
		switch ev := ev.(type) {
		case pointer.Event:
			if ev.Kind == pointer.Press {
				gtx.Execute(key.FocusCmd{Tag: s})
			}
		case key.FocusEvent:
			s.focused = ev.Focus
		case key.Event:
			if ev.State != key.Press {
				continue
			}
			switch ev.Name {
			case minus:
				s.value.Value = max(0, s.value.Value-0.01)
				changed = true
			case plus:
				s.value.Value = min(1, s.value.Value+0.01)
				changed = true
			case key.NamePageUp:
				s.value.Value = min(1, s.value.Value+0.1)
				changed = true
			case key.NamePageDown:
				s.value.Value = max(0, s.value.Value-0.1)
				changed = true
			case key.NameHome:
				s.value.Value = 0
				changed = true
			case key.NameEnd:
				s.value.Value = 1
				changed = true
			}
		}
	}
	return changed
}

func (s *Slider) Value() float32 {
	return s.value.Value
}

func (s *Slider) SetValue(v float32) {
	s.value.Value = max(0, min(1, v))
}

func (s *Slider) IsFocused(gtx layout.Context) bool {
	return gtx.Focused(s)
}

func (s *Slider) SetFocused(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: s})
}
