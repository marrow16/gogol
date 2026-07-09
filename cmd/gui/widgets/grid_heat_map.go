package widgets

import (
	"gioui.org/op/paint"
	"github.com/marrow16/gogol/logic"
	"image"
	"image/color"
	"image/draw"
	"math"
)

func (g *gridHolder) buildHeatMap(heatMap logic.HeatMap) {
	g.drawHeatMap(g.heatMapCanvas, heatMap)
	// build the pre-prepared image op...
	g.heatMapImgOp = paint.NewImageOp(g.heatMapCanvas)
}

func (g *gridHolder) drawHeatMap(img *image.NRGBA, heatMap logic.HeatMap) {
	if heatMap == nil {
		return
	}
	draw.Draw(img, image.Rect(0, 0, g.core.settings.Width*g.core.settings.CellSize, g.core.settings.Height*g.core.settings.CellSize), &image.Uniform{g.core.settings.CellDeadColor}, image.Point{}, draw.Src)
	g.drawCellBorders(img)
	off := 0
	if g.core.settings.CellBorders {
		off = 1
	}
	for pt := range heatMap.HeatMap() {
		clr := heatColor(pt.Value)
		draw.Draw(img, image.Rect(
			(pt.Col*g.core.settings.CellSize)+off,
			(pt.Row*g.core.settings.CellSize)+off,
			(pt.Col+1)*g.core.settings.CellSize,
			(pt.Row+1)*g.core.settings.CellSize),
			&image.Uniform{clr}, image.Point{}, draw.Src)
	}
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

// heatColor returns a colour for v in the range [0,1]
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
