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
		for range width {
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
		for range height {
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
		for range height {
			pix[i] = c
			i += stride
		}
	}
}
