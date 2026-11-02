package controls

import (
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"image"
)

// Clickable same as widget.Clickable except it doesn't receive keyboard focus
type Clickable struct {
	click gesture.Click
}

// Layout and update the button state.
func (b *Clickable) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	return b.layout(b, gtx, w)
}

func (b *Clickable) layout(t event.Tag, gtx layout.Context, w layout.Widget) layout.Dimensions {
	for {
		_, ok := b.update(t, gtx)
		if !ok {
			break
		}
	}
	m := op.Record(gtx.Ops)
	dims := w(gtx)
	c := m.Stop()
	defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
	semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
	b.click.Add(gtx.Ops)
	event.Op(gtx.Ops, t)
	c.Add(gtx.Ops)
	return dims
}

// Update the button state by processing events, and return the next
// click, if any.
func (b *Clickable) Update(gtx layout.Context) (widget.Click, bool) {
	return b.update(b, gtx)
}

func (b *Clickable) update(t event.Tag, gtx layout.Context) (widget.Click, bool) {
	for {
		e, ok := b.click.Update(gtx.Source)
		if !ok {
			break
		}
		switch e.Kind {
		case gesture.KindClick:
			return widget.Click{
				Modifiers: e.Modifiers,
				NumClicks: e.NumClicks,
			}, true
		}
	}
	return widget.Click{}, false
}

// Clicked calls Update and reports whether a click was registered.
func (b *Clickable) Clicked(gtx layout.Context) bool {
	return b.clicked(b, gtx)
}

func (b *Clickable) clicked(t event.Tag, gtx layout.Context) bool {
	_, clicked := b.update(t, gtx)
	return clicked
}

// Hovered reports whether a pointer is over the element.
func (b *Clickable) Hovered() bool {
	return b.click.Hovered()
}
