package controls

import (
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"image"
)

func NewIconButton(img image.Image, size unit.Sp) *IconButton {
	return &IconButton{
		image: img,
		size:  size,
	}
}

type IconButton struct {
	clickable widget.Clickable
	image     image.Image
	size      unit.Sp
}

func (b *IconButton) Layout(gtx layout.Context) layout.Dimensions {
	return b.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		var wh int
		switch {
		case b.size == 0:
			wh = gtx.Constraints.Max.Y
		case b.size < 0:
			wh = gtx.Constraints.Max.Y - gtx.Sp(0-b.size)
		default:
			wh = gtx.Sp(b.size)
		}
		size := image.Point{X: wh, Y: wh}
		r := b.image.Bounds()
		defer op.Affine(
			f32.Affine2D{}.Scale(
				f32.Point{X: 0, Y: 0},
				f32.Point{X: float32(size.X) / float32(r.Dx()), Y: float32(size.Y) / float32(r.Dy())}),
		).Push(gtx.Ops).Pop()
		paint.NewImageOp(b.image).Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		return layout.Dimensions{Size: size}
	})
}

func (b *IconButton) Clicked(gtx layout.Context) bool {
	return b.clickable.Clicked(gtx)
}
