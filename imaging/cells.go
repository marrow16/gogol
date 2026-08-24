package imaging

import (
	"image"
	"image/color"
)

func CellsImage(cells [][]bool, cfg Config) *image.NRGBA {
	height := len(cells)
	width := len(cells[0])
	ht := height * cfg.CellSize
	wd := width * cfg.CellSize
	offset := 0
	if cfg.Borders && cfg.CellSize > 2 {
		offset = 1
		wd++
		ht++
	}
	img := image.NewNRGBA(image.Rectangle{Max: image.Point{X: wd, Y: ht}})
	pix := img.Pix
	stride := img.Stride
	// fill dead background...
	for i := 0; i < len(pix); i += 4 {
		pix[i] = cfg.DeadColor.R
		pix[i+1] = cfg.DeadColor.G
		pix[i+2] = cfg.DeadColor.B
		pix[i+3] = cfg.DeadColor.A
	}
	if cfg.Borders && cfg.CellSize > 2 {
		DrawCellBorders(img, wd, ht, cfg.CellSize, cfg.BorderColor)
	}
	cellSize := cfg.CellSize
	c := cfg.AliveColor
	cellWidth := cellSize - offset
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			if !cells[row][col] {
				continue
			}
			xMin := col*cellSize + offset
			yMin := row*cellSize + offset
			yMax := yMin + cellWidth
			line := yMin*stride + xMin*4
			lineBytes := cellWidth * 4
			for y := yMin; y < yMax; y++ {
				i := line
				end := line + lineBytes
				for i < end {
					pix[i] = c.R
					pix[i+1] = c.G
					pix[i+2] = c.B
					pix[i+3] = c.A
					i += 4
				}
				line += stride
			}
		}
	}
	return img
}

func CellsImagePaletted(cells [][]bool, cfg Config) *image.Paletted {
	height := len(cells)
	width := len(cells[0])
	wd := width * cfg.CellSize
	ht := height * cfg.CellSize
	offset := 0
	if cfg.Borders && cfg.CellSize > 2 {
		offset = 1
		wd++
		ht++
	}
	img := image.NewPaletted(
		image.Rectangle{Max: image.Point{X: wd, Y: ht}},
		color.Palette{
			DeadIndex:   cfg.DeadColor,
			AliveIndex:  cfg.AliveColor,
			BorderIndex: cfg.BorderColor,
		},
	)
	if cfg.Borders && cfg.CellSize > 2 {
		DrawCellBordersPaletted(img, wd, ht, cfg.CellSize, BorderIndex)
	}
	pix := img.Pix
	stride := img.Stride
	cellSize := cfg.CellSize
	cellWidth := cellSize - offset
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			if !cells[row][col] {
				continue
			}
			xMin := col*cellSize + offset
			yMin := row*cellSize + offset
			yMax := yMin + cellWidth
			line := yMin*stride + xMin
			for y := yMin; y < yMax; y++ {
				end := line + cellWidth
				for i := line; i < end; i++ {
					pix[i] = AliveIndex
				}
				line += stride
			}
		}
	}
	return img
}
