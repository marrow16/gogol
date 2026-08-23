package imaging

import (
	"github.com/marrow16/gogol/logic"
	"image"
	"image/color"
	"math"
)

// HeatMap creates a heat map image for the given heat mapper
//
// the colorizer arg is an optional func for converting heat map values to a color -
// if this is nil, the default colorizer is used
//
// Note: logic.HeatMap does not know about original grid dimensions, so rows and cols size have to be supplied and
// **must** match the original grid size!
func HeatMap(heatMap logic.HeatMap, rows, cols int, cfg Config, colorizer func(v float64) color.NRGBA) *image.NRGBA {
	if colorizer == nil {
		colorizer = heatColor
	}
	wd := cols * cfg.CellSize
	ht := rows * cfg.CellSize
	offset := 0
	if cfg.Borders && cfg.CellSize > 2 {
		offset = 1
		wd++
		ht++
	}
	img := image.NewNRGBA(image.Rectangle{
		Max: image.Point{X: wd, Y: ht},
	})
	pix := img.Pix
	stride := img.Stride
	// Dead background.
	dc := cfg.DeadColor
	for i := 0; i < len(pix); i += 4 {
		pix[i] = dc.R
		pix[i+1] = dc.G
		pix[i+2] = dc.B
		pix[i+3] = dc.A
	}
	if cfg.Borders && cfg.CellSize > 2 {
		DrawCellBorders(img, wd, ht, cfg.CellSize, cfg.BorderColor)
	}
	cellSize := cfg.CellSize
	cellWidth := cellSize - offset
	colors := make(map[float64]color.NRGBA)
	for pt := range heatMap.HeatMap() {
		clr, ok := colors[pt.Value]
		if !ok {
			clr = colorizer(pt.Value)
			colors[pt.Value] = clr
		}
		xMin := pt.Col*cellSize + offset
		yMin := pt.Row*cellSize + offset
		yMax := yMin + cellWidth
		line := yMin*stride + xMin*4
		lineBytes := cellWidth * 4
		for y := yMin; y < yMax; y++ {
			i := line
			end := line + lineBytes
			for i < end {
				pix[i] = clr.R
				pix[i+1] = clr.G
				pix[i+2] = clr.B
				pix[i+3] = clr.A
				i += 4
			}
			line += stride
		}
	}
	return img
}

var heatColors = []color.NRGBA{
	{R: 15, G: 24, B: 180, A: 255},   // deep blue
	{R: 30, G: 90, B: 255, A: 255},   // blue
	{R: 60, G: 200, B: 255, A: 255},  // cyan
	{R: 245, G: 245, B: 235, A: 255}, // near white
	{R: 255, G: 220, B: 80, A: 255},  // yellow
	{R: 255, G: 140, B: 20, A: 255},  // orange
	{R: 220, G: 40, B: 10, A: 255},   // red
	{R: 110, G: 0, B: 0, A: 255},     // dark red
}

// heatColor returns a color for v in the range [0,1]
func heatColor(v float64) color.NRGBA {
	if v <= 0 {
		return heatColors[0]
	}
	if v >= 1 {
		return heatColors[len(heatColors)-1]
	}
	v *= float64(len(heatColors) - 1)
	i := int(math.Floor(v))
	t := v - float64(i)
	c0 := heatColors[i]
	c1 := heatColors[i+1]
	return color.NRGBA{
		R: uint8(float64(c0.R) + (float64(c1.R)-float64(c0.R))*t + 0.5),
		G: uint8(float64(c0.G) + (float64(c1.G)-float64(c0.G))*t + 0.5),
		B: uint8(float64(c0.B) + (float64(c1.B)-float64(c0.B))*t + 0.5),
		A: 255,
	}
}
