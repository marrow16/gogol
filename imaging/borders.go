package imaging

import (
	"image"
	"image/color"
)

func DrawCellBorders(img *image.NRGBA, width, height, cellSize int, c color.NRGBA) {
	pix := img.Pix
	stride := img.Stride
	// horizontal borders...
	for y := 0; y < height; y += cellSize {
		i := y * stride
		for x := 0; x < width; x++ {
			pix[i] = c.R
			pix[i+1] = c.G
			pix[i+2] = c.B
			pix[i+3] = c.A
			i += 4
		}
	}
	// vertical borders...
	for x := 0; x < width; x += cellSize {
		i := x * 4
		for y := 0; y < height; y++ {
			pix[i] = c.R
			pix[i+1] = c.G
			pix[i+2] = c.B
			pix[i+3] = c.A
			i += stride
		}
	}
}

func DrawCellBordersPaletted(img *image.Paletted, width, height, cellSize int, c uint8) {
	pix := img.Pix
	stride := img.Stride
	// horizontal borders...
	for y := 0; y < height; y += cellSize {
		i := y * stride
		row := pix[i : i+width]
		for x := range row {
			row[x] = c
		}
	}
	// vertical borders...
	for x := 0; x < width; x += cellSize {
		i := x
		for y := 0; y < height; y++ {
			pix[i] = c
			i += stride
		}
	}
}
