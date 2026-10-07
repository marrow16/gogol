package help

import (
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
	"strings"
)

type expandable struct {
	title           any
	content         any
	initialExpanded bool
	border          bool
	indent          unit.Dp
	rightMargin     unit.Dp
	padding         unit.Dp
	spacing
}

func (e expandable) String() string {
	var sb strings.Builder
	switch ct := e.title.(type) {
	case string:
		sb.WriteString(ct)
	case content:
		sb.WriteString(ct.String())
	case []any:
		sb.WriteString(content(ct).String())
	}
	sb.WriteString(" ")
	switch ct := e.content.(type) {
	case string:
		sb.WriteString(ct)
	case content:
		sb.WriteString(ct.String())
	case []any:
		sb.WriteString(content(ct).String())
	}
	return sb.String()
}

type expandableState struct {
	expanded  bool
	clickable *widget.Clickable
}

func (e expandable) widget(h *holder) layout.Widget {
	state := &expandableState{
		clickable: &widget.Clickable{},
		expanded:  e.initialExpanded,
	}
	var expContent layout.Widget
	switch ct := e.content.(type) {
	case string:
		expContent = buildContent(h, 0, content{ct})
	case content:
		expContent = buildContent(h, 0, ct)
	case []any:
		expContent = buildContent(h, 0, ct)
	}
	return func(gtx layout.Context) layout.Dimensions {
		if state.clickable.Clicked(gtx) {
			state.expanded = !state.expanded
		}
		return layout.Inset{Left: e.indent, Right: e.rightMargin, Top: e.spaceBefore, Bottom: e.spaceAfter}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return clickable(gtx, state.clickable, func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Top: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal, Gap: 8}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									if state.expanded {
										return material.Label(h.theme, h.theme.TextSize, keyUpTriangle).Layout(gtx)
									}
									return material.Label(h.theme, h.theme.TextSize, keyDownTriangle).Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									switch ct := e.title.(type) {
									case string:
										return buildContent(h, 0, content{ct})(gtx)
									case content:
										return buildContent(h, 0, ct)(gtx)
									case []any:
										return buildContent(h, 0, ct)(gtx)
									}
									return layout.Dimensions{}
								}),
							)
						})
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if state.expanded && expContent != nil {
						return layout.Inset{Top: e.padding, Bottom: e.padding, Left: e.padding, Right: e.padding}.Layout(gtx, expContent)
					}
					return layout.Dimensions{}
				}),
			)
			if e.border {
				expandableBorder(gtx, h.theme, dims)
			}
			return dims
		})
	}
}

func expandableBorder(gtx layout.Context, theme *material.Theme, dims layout.Dimensions) {
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
	defer clip.Stroke{
		Path:  rr.Path(gtx.Ops),
		Width: width,
	}.Op().Push(gtx.Ops).Pop()
	paint.ColorOp{Color: theme.Fg}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}
