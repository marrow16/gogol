package imaging

import (
	"github.com/marrow16/gogol/logic"
	"image"
	"image/color"
)

type Config struct {
	CellSize    int
	Borders     bool
	AliveColor  color.NRGBA
	DeadColor   color.NRGBA
	BorderColor color.NRGBA
	Paletted    bool
}

const (
	dead   = uint8(0)
	alive  = uint8(1)
	border = uint8(2)
)

func Grid(g *logic.Grid, cfg Config) image.Image {
	if cfg.Paletted {
		return GridImagePaletted(g, cfg)
	}
	return GridImage(g, cfg)
}

func GridImage(g *logic.Grid, cfg Config) *image.NRGBA {
	wd := g.Width * cfg.CellSize
	ht := g.Height * cfg.CellSize
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
	g.DrawTo(func(row, col int, a bool) {
		if !a {
			return
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
	})
	return img
}

func GridImagePaletted(g *logic.Grid, cfg Config) *image.Paletted {
	wd := g.Width * cfg.CellSize
	ht := g.Height * cfg.CellSize
	offset := 0
	if cfg.Borders && cfg.CellSize > 2 {
		offset = 1
		wd++
		ht++
	}
	img := image.NewPaletted(
		image.Rectangle{Max: image.Point{X: wd, Y: ht}},
		color.Palette{
			cfg.DeadColor,
			cfg.AliveColor,
			cfg.BorderColor,
		},
	)
	if cfg.Borders && cfg.CellSize > 2 {
		DrawCellBordersPaletted(img, wd, ht, cfg.CellSize, border)
	}
	pix := img.Pix
	stride := img.Stride
	cellSize := cfg.CellSize
	cellWidth := cellSize - offset
	g.DrawTo(func(row, col int, a bool) {
		if !a {
			return
		}
		xMin := col*cellSize + offset
		yMin := row*cellSize + offset
		yMax := yMin + cellWidth
		line := yMin*stride + xMin
		for y := yMin; y < yMax; y++ {
			end := line + cellWidth
			for i := line; i < end; i++ {
				pix[i] = alive
			}
			line += stride
		}
	})
	return img
}

func GridSliceImage(g [][]bool, cfg Config) *image.NRGBA {
	height := len(g)
	width := len(g[0])
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
			if !g[row][col] {
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

func GridImageSlicePaletted(g [][]bool, cfg Config) *image.Paletted {
	height := len(g)
	width := len(g[0])
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
			cfg.DeadColor,
			cfg.AliveColor,
			cfg.BorderColor,
		},
	)
	if cfg.Borders && cfg.CellSize > 2 {
		DrawCellBordersPaletted(img, wd, ht, cfg.CellSize, border)
	}
	pix := img.Pix
	stride := img.Stride
	cellSize := cfg.CellSize
	cellWidth := cellSize - offset
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			if !g[row][col] {
				continue
			}
			xMin := col*cellSize + offset
			yMin := row*cellSize + offset
			yMax := yMin + cellWidth
			line := yMin*stride + xMin
			for y := yMin; y < yMax; y++ {
				end := line + cellWidth
				for i := line; i < end; i++ {
					pix[i] = alive
				}
				line += stride
			}
		}
	}
	return img
}
