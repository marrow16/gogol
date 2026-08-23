package imaging

import (
	"github.com/marrow16/gogol/patterns"
	"image"
	"image/color"
)

func Pattern(p patterns.Pattern, cfg Config) image.Image {
	if cfg.Paletted {
		return PatternImagePaletted(p, cfg)
	}
	return PatternImage(p, cfg)
}

func PatternImage(p patterns.Pattern, cfg Config) *image.NRGBA {
	wd := p.Width * cfg.CellSize
	ht := p.Height * cfg.CellSize
	offset := 0
	if cfg.Borders && cfg.CellSize > 2 {
		offset++
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
	r := cfg.AliveColor
	cellWidth := cellSize - offset
	p.DrawTo(patterns.Rotate0, func(row, col int, a bool) {
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
				pix[i] = r.R
				pix[i+1] = r.G
				pix[i+2] = r.B
				pix[i+3] = r.A
				i += 4
			}
			line += stride
		}
	})
	return img
}

func PatternImagePaletted(p patterns.Pattern, cfg Config) *image.Paletted {
	wd := p.Width * cfg.CellSize
	ht := p.Height * cfg.CellSize
	offset := 0
	if cfg.Borders && cfg.CellSize > 2 {
		offset++
		wd++
		ht++
	}
	img := image.NewPaletted(image.Rectangle{Max: image.Point{wd, ht}}, color.Palette{
		dead:   cfg.DeadColor,
		alive:  cfg.AliveColor,
		border: cfg.BorderColor,
	})
	if cfg.Borders && cfg.CellSize > 2 {
		DrawCellBordersPaletted(img, wd, ht, cfg.CellSize, border)
	}
	if cfg.CellSize == 1 {
		p.DrawTo(patterns.Rotate0, func(row, col int, a bool) {
			if a {
				img.Pix[img.PixOffset(col, row)] = alive
			}
		})
	} else {
		p.DrawTo(patterns.Rotate0, func(row, col int, a bool) {
			if a {
				x0 := (col * cfg.CellSize) + offset
				y0 := (row * cfg.CellSize) + offset
				x1 := x0 + cfg.CellSize - offset
				y1 := y0 + cfg.CellSize - offset
				for y := y0; y < y1; y++ {
					off := img.PixOffset(x0, y)
					r := img.Pix[off : off+(x1-x0)]
					for i := range r {
						r[i] = alive
					}
				}
			}
		})
	}
	return img
}
