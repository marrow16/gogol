package imaging

import (
	"image"
	"image/color"
)

// ScaleSparse converts a sparse paletted image to an NRGBA image
//
// Especially useful for very large pattern images.
func ScaleSparse(src *image.Paletted, scale float32) *image.NRGBA {
	sb := src.Bounds()
	w := max(1, int(float32(sb.Dx())*scale))
	h := max(1, int(float32(sb.Dy())*scale))
	dst := image.NewNRGBA(image.Rectangle{Max: image.Point{X: w, Y: h}})
	dstPix := dst.Pix
	// background fill...
	bg := src.Palette[DeadIndex].(color.NRGBA)
	for i := 0; i < len(dstPix); i += 4 {
		dstPix[i] = bg.R
		dstPix[i+1] = bg.G
		dstPix[i+2] = bg.B
		dstPix[i+3] = bg.A
	}
	srcPix := src.Pix
	srcStride := src.Stride
	dstStride := dst.Stride
	srcW := sb.Dx()
	srcH := sb.Dy()
	for sy := 0; sy < srcH; sy++ {
		srcOff := sy * srcStride
		dy := int(float32(sy) * scale)
		if dy >= h {
			dy = h - 1
		}
		dstRow := dy * dstStride
		for sx := 0; sx < srcW; sx++ {
			pi := srcPix[srcOff+sx]
			if pi == DeadIndex {
				continue
			}
			dx := int(float32(sx) * scale)
			if dx >= w {
				dx = w - 1
			}
			clr := src.Palette[pi].(color.NRGBA)
			i := dstRow + dx*4
			dstPix[i] = clr.R
			dstPix[i+1] = clr.G
			dstPix[i+2] = clr.B
			dstPix[i+3] = clr.A
		}
	}
	return dst
}
