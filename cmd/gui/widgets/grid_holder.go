package widgets

import (
	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strconv"
	"time"
)

func newGridHolder(c *Core) (*gridHolder, error) {
	var lg *logic.Grid
	fromSaved := false
	if c.settings.SavedGrid != nil {
		fromSaved = true
		lg = c.settings.SavedGrid
		c.settings.Height, c.settings.Width, c.settings.WrapMode, c.settings.BoundaryMode = lg.Height, lg.Width, lg.WrapMode, lg.BoundaryMode
	} else {
		var err error
		lg, err = logic.NewGrid(c.settings.Height, c.settings.Width, c.settings.WrapMode, c.settings.BoundaryMode)
		if err != nil {
			return nil, err
		}
		if c.settings.Rule != "" {
			if r, ok := logic.Rules[c.settings.Rule]; ok {
				lg.SetRule(r)
			} else if r, err = logic.NewRuleRle("", c.settings.Rule); err == nil {
				lg.SetRule(r)
			} else {
				lg.Rule = logic.StandardRule
			}
		} else {
			lg.Rule = logic.StandardRule
		}
	}
	g := &gridHolder{
		core: c,
		grid: lg,
		zoom: c.settings.Zoom,
	}
	g.editor = &editor{g: g}
	g.rebuild()
	lg.Render = g.renderCell
	if fromSaved {
		lg.Draw()
	} else {
		lg.Randomize(c.settings.Randomization)
	}
	return g, nil
}

type gridHolder struct {
	core      *Core
	grid      *logic.Grid
	canvas    *image.NRGBA
	imgOp     paint.ImageOp
	dirty     bool
	clickable widget.Clickable
	zoom      float32
	pan       f32.Point
	lastPos   f32.Point

	// underlays/overlays (for pattern placing)...
	underlay *underlay
	overlay  *overlay
	editor   *editor
	// heat mapping...
	heatMapCanvas *image.NRGBA
	heatMapImgOp  paint.ImageOp
}

func (g *gridHolder) replaceGrid(grid *logic.Grid) {
	g.grid = grid
	g.rebuild()
	g.grid.Render = g.renderCell
	g.grid.Draw()
}

func (g *gridHolder) layout(gtx layout.Context) layout.Dimensions {
	size := gtx.Constraints.Max
	viewport := image.Rectangle{Max: size}
	canvasSize := image.Pt(
		int(float32(g.canvas.Bounds().Dx())*g.zoom),
		int(float32(g.canvas.Bounds().Dy())*g.zoom),
	)
	g.pan = clampPan(g.pan, canvasSize, size)
	canvasRect := image.Rectangle{
		Min: image.Pt(int(g.pan.X), int(g.pan.Y)),
		Max: image.Pt(int(g.pan.X)+canvasSize.X, int(g.pan.Y)+canvasSize.Y),
	}
	eventFilters := []event.Filter{
		pointer.Filter{
			Target:  &g.clickable,
			Kinds:   pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel | pointer.Scroll,
			ScrollX: pointer.ScrollRange{Min: -canvasSize.X, Max: canvasSize.X},
			ScrollY: pointer.ScrollRange{Min: -canvasSize.Y, Max: canvasSize.Y},
		},
	}
	if g.overlay != nil {
		gtx.Execute(key.FocusCmd{Tag: &g.clickable})
		eventFilters = append(eventFilters,
			key.Filter{Focus: &g.clickable, Optional: key.ModShift | key.ModCtrl | key.ModAlt | key.ModShortcut},
		)

	} else if g.editor.active {
		gtx.Execute(key.FocusCmd{Tag: &g.clickable})
		eventFilters = append(eventFilters,
			key.Filter{Focus: &g.clickable, Optional: key.ModShift | key.ModCtrl | key.ModAlt | key.ModShortcut},
			transfer.TargetFilter{
				Target: &g.editor.clipboardTag,
				Type:   clipboardReadType,
			},
		)
	}
	for {
		ev, ok := gtx.Event(eventFilters...)
		if !ok {
			break
		}
		switch evt := ev.(type) {
		case transfer.DataEvent:
			g.editor.handlePaste(evt)
		case key.Event:
			g.handleKeys(gtx, evt)
		case pointer.Event:
			switch evt.Kind {
			case pointer.Scroll:
				if canvasSize.X > size.X {
					g.pan.X += evt.Scroll.X
				}
				if canvasSize.Y > size.Y {
					g.pan.Y += evt.Scroll.Y
				}
				g.pan = clampPan(g.pan, canvasSize, size)
				gtx.Execute(op.InvalidateCmd{})
			case pointer.Press:
				gtx.Execute(key.FocusCmd{Tag: &g.clickable})
				if g.overlay != nil && evt.Buttons == pointer.ButtonPrimary {
					if r, c, ok := g.mouseToGrid(evt.Position); ok {
						g.overlay.row, g.overlay.col = r, c
						g.overlay.moved = true
					}
				}
				if g.editor.active && evt.Buttons == pointer.ButtonPrimary {
					if r, c, ok := g.mouseToGrid(evt.Position); ok {
						g.editor.setPosition(r, c)
					}
				}
			case pointer.Drag:
				if g.editor.active {
					if r, c, ok := g.mouseToGrid(evt.Position); ok {
						g.editor.markArea(r, c)
					}
				}
			case pointer.Release, pointer.Cancel:
				// anything to do?
			}
		}
	}
	gridClip := clip.Rect(viewport).Push(gtx.Ops)
	defer gridClip.Pop()
	defer op.Offset(canvasRect.Min).Push(gtx.Ops).Pop()
	gtx.Constraints = layout.Exact(canvasRect.Size())
	g.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		imgOp := g.imageOp()
		zoom := op.Affine(
			f32.Affine2D{}.
				Scale(f32.Pt(0, 0), f32.Pt(g.zoom, g.zoom)),
		).Push(gtx.Ops)
		imgOp.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		zoom.Pop()
		return layout.Dimensions{Size: canvasRect.Size()}
	})
	return layout.Dimensions{Size: size}
}

func (g *gridHolder) mouseToGrid(pos f32.Point) (row, col int, ok bool) {
	x := pos.X
	y := pos.Y
	if x < 0 || y < 0 {
		return 0, 0, false
	}
	cellSize := float32(g.core.settings.CellSize) * g.zoom
	col = int(x / cellSize)
	row = int(y / cellSize)
	if row < 0 || row >= g.grid.Height ||
		col < 0 || col >= g.grid.Width {
		return row, col, false
	}
	return row, col, true
}

func (g *gridHolder) handleKeys(gtx layout.Context, kev key.Event) {
	if kev.State != key.Press {
		return
	}
	if g.editor.handleKeys(gtx, kev) {
		return
	} else if g.overlay == nil {
		return
	}
	// we're only handling overlay keys from here!
	switch kev.Name {
	case key.NameEnter, key.NameReturn:
		g.core.placePattern(gtx)
	case key.NameLeftArrow:
		if g.overlay.col > 0 {
			g.overlay.col--
			g.overlay.moved = true
		}
	case key.NameRightArrow:
		if g.overlay.col < g.grid.Width-1 {
			g.overlay.col++
			g.overlay.moved = true
		}
	case key.NameUpArrow:
		if g.overlay.row > 0 {
			g.overlay.row--
			g.overlay.moved = true
		}
	case key.NameDownArrow:
		if g.overlay.row < g.grid.Height-1 {
			g.overlay.row++
			g.overlay.moved = true
		}
	case "R":
		rot := g.overlay.rotation + 1
		if rot > patterns.Rotate270 {
			rot = patterns.Rotate0
		}
		g.overlay.rotation = rot
		g.overlay.moved = true
	}
}

func clampPan(pan f32.Point, canvasSize, viewportSize image.Point) f32.Point {
	minX := float32(viewportSize.X - canvasSize.X)
	minY := float32(viewportSize.Y - canvasSize.Y)
	if minX >= 0 {
		pan.X = 0
	} else {
		pan.X = min(0, max(minX, pan.X))
	}
	if minY >= 0 {
		pan.Y = 0
	} else {
		pan.Y = min(0, max(minY, pan.Y))
	}
	return pan
}

func (g *gridHolder) placeOverlay() {
	g.restoreUnderlay()
	if o := g.overlay; o != nil {
		o.pattern.Draw(g.grid, o.row, o.col, o.rotation, o.interlaced)
		g.overlay = nil
	}
}

func (g *gridHolder) restoreUnderlay() {
	if u := g.underlay; u != nil {
		u.pattern.DrawTo(patterns.Rotate0, func(row, col int, alive bool) {
			g.renderCell(row+u.row, col+u.col, alive, true)
		})
		g.underlay = nil
	}
}

func (g *gridHolder) imageOp() paint.ImageOp {
	if g.core.mode == heatMapMode {
		return g.heatMapImgOp
	}
	o := g.overlay
	if g.underlay != nil && (o == nil || o.moved) {
		g.restoreUnderlay()
	}
	if o != nil && o.moved {
		// remember underlay...
		height, width := o.pattern.Height, o.pattern.Width
		if o.rotation == patterns.Rotate90 || o.rotation == patterns.Rotate270 {
			height, width = o.pattern.Width, o.pattern.Height
		}
		g.underlay = &underlay{
			pattern: patterns.NewPatternFromGridPortion(g.grid, o.row, o.col, height, width),
			row:     o.row,
			col:     o.col,
		}
		// draw overlay...
		aliveColor, deadColor := placementColors(g.core.settings.CellAliveColor, g.core.settings.CellDeadColor)
		o.pattern.DrawTo(o.rotation, func(row, col int, alive bool) {
			if !o.interlaced || (o.interlaced && alive) {
				g.renderCellWithColors(row+o.row, col+o.col, alive, aliveColor, deadColor)
			}
		})
		o.moved = false
	}
	g.editor.imageOps()
	if g.dirty {
		g.imgOp = paint.NewImageOp(g.canvas)
		g.dirty = false
	}
	return g.imgOp
}

type underlay struct {
	pattern  patterns.Pattern
	row, col int
}

type overlay struct {
	pattern    patterns.Pattern
	row, col   int
	rotation   patterns.Rotation
	interlaced bool
	moved      bool
}

func (g *gridHolder) resize() {
	if lg, err := logic.NewGrid(g.core.settings.Height, g.core.settings.Width, g.grid.WrapMode, g.grid.BoundaryMode); err == nil {
		lg.Rule = g.grid.Rule
		lg.Render = g.renderCell
		g.grid = lg
		g.rebuild()
		g.core.status = "Grid " + strconv.Itoa(g.core.settings.Width) + " x " + strconv.Itoa(g.core.settings.Height)
		go func() {
			time.Sleep(5 * time.Second)
			g.core.status = ""
		}()
	}
}

func (g *gridHolder) rebuild() {
	g.canvas = image.NewNRGBA(image.Rect(0, 0, g.core.settings.Width*g.core.settings.CellSize, g.core.settings.Height*g.core.settings.CellSize))
	g.heatMapCanvas = image.NewNRGBA(image.Rect(0, 0, g.core.settings.Width*g.core.settings.CellSize, g.core.settings.Height*g.core.settings.CellSize))
	draw.Draw(g.canvas, image.Rect(0, 0, g.core.settings.Width*g.core.settings.CellSize, g.core.settings.Height*g.core.settings.CellSize), &image.Uniform{g.core.settings.CellDeadColor}, image.Point{}, draw.Src)
	g.drawCellBorders(g.canvas)
	g.dirty = true
}

func (g *gridHolder) drawCellBorders(img *image.NRGBA) {
	if g.core.settings.CellBorders {
		for y := 0; y <= g.core.settings.Height; y++ {
			yy := y * g.core.settings.CellSize
			draw.Draw(
				img,
				image.Rect(0, yy, g.core.settings.Width*g.core.settings.CellSize, yy+1),
				&image.Uniform{g.core.settings.CellBorderColor},
				image.Point{},
				draw.Src,
			)
		}
		for x := 0; x <= g.core.settings.Width; x++ {
			xx := x * g.core.settings.CellSize
			draw.Draw(
				img,
				image.Rect(xx, 0, xx+1, g.core.settings.Height*g.core.settings.CellSize),
				&image.Uniform{g.core.settings.CellBorderColor},
				image.Point{},
				draw.Src,
			)
		}
	}
}

func (g *gridHolder) renderCell(row, col int, alive, changed bool) {
	if changed {
		c := g.core.settings.CellDeadColor
		if alive {
			c = g.core.settings.CellAliveColor
		}
		off := 0
		if g.core.settings.CellBorders {
			off = 1
		}
		draw.Draw(g.canvas, image.Rect(
			(col*g.core.settings.CellSize)+off,
			(row*g.core.settings.CellSize)+off,
			(col+1)*g.core.settings.CellSize,
			(row+1)*g.core.settings.CellSize),
			&image.Uniform{c}, image.Point{}, draw.Src)
		g.dirty = true
	}
}

func (g *gridHolder) renderCellWithColors(row, col int, alive bool, aliveColor, deadColor color.Color) {
	c := deadColor
	if alive {
		c = aliveColor
	}
	off := 0
	if g.core.settings.CellBorders {
		off = 1
	}
	draw.Draw(g.canvas, image.Rect(
		(col*g.core.settings.CellSize)+off,
		(row*g.core.settings.CellSize)+off,
		(col+1)*g.core.settings.CellSize,
		(row+1)*g.core.settings.CellSize),
		&image.Uniform{c}, image.Point{}, draw.Src)
	g.dirty = true
}

func (g *gridHolder) startEditing() {
	g.editor.start()
}

func (g *gridHolder) stopEditing() {
	g.editor.end()
}

func placementColors(alive, dead color.NRGBA) (placeAlive, placeDead color.NRGBA) {
	return placementAliveColor(alive), dead
}

func placementAliveColor(c color.NRGBA) color.NRGBA {
	h, s, v := rgbToHSV(c.R, c.G, c.B)
	if s < 0.01 {
		// Greys have no useful hue, so force green.
		h = 120
		s = 1
		v = maxFloat(v, 0.85)
	} else {
		h += 180
		if h >= 360 {
			h -= 360
		}
		s = maxFloat(s, 0.75)
		v = maxFloat(v, 0.75)
	}
	r, g, b := hsvToRGB(h, s, v)
	return color.NRGBA{R: r, G: g, B: b, A: 255}
}

func placementDeadColor(c color.NRGBA) color.NRGBA {
	h, s, v := rgbToHSV(c.R, c.G, c.B)
	if s < 0.01 {
		// Greys have no useful hue, so force green.
		h = 0
		s = 1
		v = maxFloat(v, 0.85)
	} else {
		h += 180
		if h >= 360 {
			h -= 360
		}
		s = maxFloat(s, 0.75)
		v = maxFloat(v, 0.75)
	}
	r, g, b := hsvToRGB(h, s, v)
	return color.NRGBA{R: r, G: g, B: b, A: 255}
}

func rgbToHSV(r8, g8, b8 uint8) (h, s, v float64) {
	r := float64(r8) / 255
	g := float64(g8) / 255
	b := float64(b8) / 255
	maxC := maxFloat(r, maxFloat(g, b))
	minC := minFloat(r, minFloat(g, b))
	delta := maxC - minC
	v = maxC
	if maxC == 0 {
		return 0, 0, v
	}
	s = delta / maxC
	if delta == 0 {
		return 0, s, v
	}
	switch maxC {
	case r:
		h = 60 * ((g - b) / delta)
		if h < 0 {
			h += 360
		}
	case g:
		h = 60 * (((b - r) / delta) + 2)
	case b:
		h = 60 * (((r - g) / delta) + 4)
	}
	return h, s, v
}

func hsvToRGB(h, s, v float64) (uint8, uint8, uint8) {
	if s <= 0 {
		x := uint8(v*255 + 0.5)
		return x, x, x
	}
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c
	var rp, gp, bp float64
	switch {
	case h < 60:
		rp, gp, bp = c, x, 0
	case h < 120:
		rp, gp, bp = x, c, 0
	case h < 180:
		rp, gp, bp = 0, c, x
	case h < 240:
		rp, gp, bp = 0, x, c
	case h < 300:
		rp, gp, bp = x, 0, c
	default:
		rp, gp, bp = c, 0, x
	}
	return floatToByte(rp + m), floatToByte(gp + m), floatToByte(bp + m)
}

func floatToByte(v float64) uint8 {
	v *= 255
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v + 0.5)
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
