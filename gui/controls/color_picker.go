package controls

import (
	"fmt"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
	"image/color"
	"image/draw"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type ColorPicker struct {
	theme           *material.Theme
	dims            layout.Dimensions
	onChangeFn      func(value color.NRGBA)
	editor          widget.Editor
	focus           bool
	editorStyle     material.EditorStyle
	current         color.NRGBA
	button          Clickable
	squareClickable Clickable
	bandClickable   Clickable
	opened          bool
	dropdownUp      bool
	colorBand       *image.RGBA
	redLabel        material.LabelStyle
	redInput        *NumberInput[int]
	greenLabel      material.LabelStyle
	greenInput      *NumberInput[int]
	blueLabel       material.LabelStyle
	blueInput       *NumberInput[int]
}

type colorComponent int

const (
	red colorComponent = iota
	blue
	green
)

func NewColorPicker(theme *material.Theme, clr color.NRGBA, fn func(color.NRGBA)) *ColorPicker {
	result := &ColorPicker{
		theme: theme,
		editor: widget.Editor{
			Alignment:  text.Start,
			MaxLen:     7,
			Filter:     "#0123456789ABCDEFabcdef",
			SingleLine: true,
			Submit:     true,
		},
		current:    clr,
		onChangeFn: fn,
		redLabel:   material.Label(theme, theme.TextSize, "R:"),
		greenLabel: material.Label(theme, theme.TextSize, "G:"),
		blueLabel:  material.Label(theme, theme.TextSize, "B:"),
	}
	result.editorStyle = material.Editor(theme, &result.editor, "")
	result.editor.SetText(fmt.Sprintf("#%02X%02X%02X", clr.R, clr.G, clr.B))
	result.redInput = NewNumberInput(theme, 3, 0, 255, 16, func(v int) {
		if v != int(result.current.R) {
			result.componentChanged(red, v)
		}
	}).SetValue(int(clr.R)).WithSpinner(true).OnSubmit(func(_ string) {
		result.focus = true
		result.opened = false
	})
	result.greenInput = NewNumberInput(theme, 3, 0, 255, 16, func(v int) {
		if v != int(result.current.R) {
			result.componentChanged(green, v)
		}
	}).SetValue(int(clr.G)).WithSpinner(true).OnSubmit(func(_ string) {
		result.focus = true
		result.opened = false
	})
	result.blueInput = NewNumberInput(theme, 3, 0, 255, 16, func(v int) {
		if v != int(result.current.B) {
			result.componentChanged(blue, v)
		}
	}).SetValue(int(clr.B)).WithSpinner(true).OnSubmit(func(_ string) {
		result.focus = true
		result.opened = false
	})
	return result
}

func (i *ColorPicker) Direction(up bool) *ColorPicker {
	i.dropdownUp = up
	return i
}

func (i *ColorPicker) Layout(gtx layout.Context) layout.Dimensions {
	i.update(gtx)
	if i.focus {
		i.focus = false
		gtx.Execute(key.FocusCmd{Tag: &i.editor})
	}
	bc, bt := Style.bordering(gtx.Focused(&i.editor) && !i.isInnerFocused(gtx))
	txtDims := measureText(gtx, i.theme, "#CCCCCC")
	i.dims = widget.Border{Color: bc, CornerRadius: 3, Width: bt}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Gap: 10, Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return i.button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						dims := layout.Dimensions{Size: image.Point{X: txtDims.Size.Y, Y: txtDims.Size.Y}}
						paint.FillShape(gtx.Ops, i.current, clip.Rect{Max: dims.Size}.Op())
						borders(gtx, dims, Style.ForegroundColor, true, true, true, true)
						return dims
					})
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X, gtx.Constraints.Max.X = txtDims.Size.X, txtDims.Size.X
					return i.editorStyle.Layout(gtx)
				}),
			)
		})
	})
	return i.dims
}

func (i *ColorPicker) LayoutDropdown(gtx layout.Context) {
	focused := i.isFocused(gtx)
	if !focused && i.opened {
		i.opened = false
	}
	if i.opened && focused {
		i.layoutDropdown(gtx)
	}
}

func (i *ColorPicker) layoutDropdown(gtx layout.Context) {
	macro := op.Record(gtx.Ops)
	dims := i.layoutDropdownControls(gtx)
	call := macro.Stop()
	y := i.dims.Size.Y
	if i.dropdownUp {
		y = -dims.Size.Y
	}
	stack := op.Offset(image.Point{X: 0, Y: y}).Push(gtx.Ops)
	defer stack.Pop()
	paint.FillShape(gtx.Ops, Style.BackgroundColor, clip.Rect{Max: dims.Size}.Op())
	borders(gtx, dims, Style.BorderColor, true, true, true, true)
	call.Add(gtx.Ops)
}

func (i *ColorPicker) layoutDropdownControls(gtx layout.Context) layout.Dimensions {
	txtDims := measureText(gtx, i.theme, "\u2003")
	width := txtDims.Size.X * 20
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			pgtx := gtx
			pgtx.Constraints.Min.X, pgtx.Constraints.Max.X = width, width
			innerWidth := width - gtx.Dp(16)
			return layout.UniformInset(8).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Gap: 10, Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X, gtx.Constraints.Max.X = innerWidth, innerWidth
						return i.layoutColorSquare(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X, gtx.Constraints.Max.X = innerWidth, innerWidth
						return i.layoutColorsBand(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X, gtx.Constraints.Max.X = innerWidth, innerWidth
						return layout.Flex{Gap: 10, Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{Gap: 8, Axis: layout.Horizontal}.Layout(gtx,
									layout.Rigid(i.redLabel.Layout),
									layout.Rigid(i.redInput.Layout),
								)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{Gap: 8, Axis: layout.Horizontal}.Layout(gtx,
									layout.Rigid(i.greenLabel.Layout),
									layout.Rigid(i.greenInput.Layout),
								)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{Gap: 8, Axis: layout.Horizontal}.Layout(gtx,
									layout.Rigid(i.blueLabel.Layout),
									layout.Rigid(i.blueInput.Layout),
								)
							}),
						)
					}),
				)
			})
		}),
	)
}

func (i *ColorPicker) layoutColorSquare(gtx layout.Context) layout.Dimensions {
	img := i.colorSquare(gtx)
	size := img.Bounds().Size()
	dims := image.Point{X: size.X, Y: size.Y}
	margin := gtx.Dp(squareMargin)
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: &i.squareClickable, Kinds: pointer.Press})
		if !ok {
			break
		}
		if pev, ok := ev.(pointer.Event); ok && pev.Kind == pointer.Press {
			px := int(pev.Position.X - float32(margin))
			py := int(pev.Position.Y - float32(margin))
			maxW := size.X - (margin * 2)
			maxH := size.Y - (margin * 2)
			px = max(0, min(px, maxW))
			py = max(0, min(py, maxH))
			c := img.RGBAAt(px, py)
			i.current = color.NRGBA{R: c.R, G: c.G, B: c.B, A: 255}
			i.updateControls()
			if i.onChangeFn != nil {
				i.onChangeFn(i.current)
			}
			gtx.Execute(op.InvalidateCmd{})
		}
	}
	return i.squareClickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
		paint.NewImageOp(img).Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)

		// marker position in displayed coordinates
		h, s, v := rgbToHSV(i.current)
		c := color.NRGBA{A: 255}
		if v < 0.5 {
			c = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		var x int
		wd := size.X - (margin * 2)
		ht := size.Y - (margin * 2)
		if h == -1 {
			half := float32(gtx.Constraints.Max.X) / 2.0
			x = int(half) - margin
		} else {
			x = int(s*float64(wd)) + margin
		}
		y := int((1-v)*float64(ht-1)) + margin
		radius := gtx.Dp(squareMarkerRadius)
		circle := clip.Ellipse{
			Min: image.Pt(x-radius, y-radius),
			Max: image.Pt(x+radius, y+radius),
		}
		paint.FillShape(gtx.Ops, c,
			clip.Stroke{
				Path:  circle.Path(gtx.Ops),
				Width: float32(gtx.Dp(2)),
			}.Op(),
		)
		return layout.Dimensions{Size: dims}
	})
}

const (
	squareMargin       = 5
	squareMarkerRadius = 4
)

func (i *ColorPicker) colorSquare(gtx layout.Context) *image.RGBA {
	size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Max.X / 2}
	r := image.Rectangle{Max: size}
	img := image.NewRGBA(r)
	draw.Draw(img, r, &image.Uniform{C: Style.BackgroundColor}, image.Point{}, draw.Src)
	margin := gtx.Dp(squareMargin)
	wd := size.X - (margin * 2)
	ht := size.Y - (margin * 2)
	h := hue(i.current)
	for y := range ht {
		v := 1 - float64(y)/float64(ht-1)
		for x := range wd {
			s := float64(x) / float64(wd-1)
			var c color.NRGBA
			if h < 0 {
				// achromatic - saturation has no effect...
				c = hsv(0, 0, v)
			} else {
				c = hsv(h, s, v)
			}
			img.Set(x+margin, y+margin, c)
		}
	}
	return img
}

func (i *ColorPicker) layoutColorsBand(gtx layout.Context) layout.Dimensions {
	img := i.colorBandImage(gtx)
	size := img.Bounds().Size()
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: &i.bandClickable, Kinds: pointer.Press})
		if !ok {
			break
		}
		if pev, ok := ev.(pointer.Event); ok && pev.Kind == pointer.Press {
			px := int(pev.Position.X)
			py := int(pev.Position.Y)
			px = max(0, min(px, size.X))
			py = max(0, min(py, size.Y))
			c := img.RGBAAt(px, py)
			i.current = color.NRGBA{R: c.R, G: c.G, B: c.B, A: 255}
			i.updateControls()
			if i.onChangeFn != nil {
				i.onChangeFn(i.current)
			}
			gtx.Execute(op.InvalidateCmd{})
		}
	}
	return i.bandClickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
		paint.NewImageOp(img).Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		return layout.Dimensions{Size: size}
	})
}

func (i *ColorPicker) colorBandImage(gtx layout.Context) *image.RGBA {
	wd := gtx.Constraints.Max.X
	if i.colorBand == nil || i.colorBand.Bounds().Size().X != wd {
		size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Max.X / 20}
		i.colorBand = image.NewRGBA(image.Rectangle{Max: size})
		for x := range size.X {
			h := float64(x) * 360 / float64(size.X-1)
			c := hsv(h, 1, 1)
			for y := range size.Y {
				i.colorBand.Set(x, y, c)
			}
		}
	}
	return i.colorBand
}

func (i *ColorPicker) update(gtx layout.Context) {
	if i.button.Clicked(gtx) {
		i.opened = !i.opened
		i.updateControls()
		if i.opened {
			i.redInput.SetFocused(gtx)
		} else {
			gtx.Execute(key.FocusCmd{Tag: &i.editor})
		}
	}
	for {
		ev, ok := i.editor.Update(gtx)
		if !ok {
			break
		}
		switch ev.(type) {
		case widget.ChangeEvent:
			i.changed()
		case widget.SubmitEvent:
			if !i.opened {
				i.opened = true
				i.redInput.SetFocused(gtx)
			} else {
				i.opened = false
			}
		}
	}
}

var colorRegex = regexp.MustCompile("^[0-9a-fA-F]{6}$")

func (i *ColorPicker) changed() {
	s := strings.TrimPrefix(i.editor.Text(), "#")
	if len(s) == 3 {
		s = s[0:1] + s[0:1] + s[1:2] + s[1:2] + s[2:3] + s[2:3]
	}
	if len(s) < 6 {
		s = s + strings.Repeat("0", 6-len(s))
	}
	if colorRegex.MatchString(s) {
		r, _ := strconv.ParseUint(s[0:2], 16, 8)
		g, _ := strconv.ParseUint(s[2:4], 16, 8)
		b, _ := strconv.ParseUint(s[4:6], 16, 8)
		i.current = color.NRGBA{
			R: uint8(r),
			G: uint8(g),
			B: uint8(b),
			A: 0xff,
		}
	} else {
		i.current = color.NRGBA{A: 255}
	}
	i.updateControls()
	if i.onChangeFn != nil {
		i.onChangeFn(i.current)
	}
}

func (i *ColorPicker) updateControls() {
	if i.opened {
		i.redInput.SetValue(int(i.current.R))
		i.greenInput.SetValue(int(i.current.G))
		i.blueInput.SetValue(int(i.current.B))
	}
}

func (i *ColorPicker) componentChanged(c colorComponent, v int) {
	if v >= 0 && v < 256 {
		switch c {
		case red:
			i.current.R = uint8(v)
		case green:
			i.current.G = uint8(v)
		case blue:
			i.current.B = uint8(v)
		}
		i.editor.SetText(fmt.Sprintf("#%02X%02X%02X", i.current.R, i.current.G, i.current.B))
	}
}

func (i *ColorPicker) isFocused(gtx layout.Context) bool {
	return gtx.Focused(&i.editor) || i.isInnerFocused(gtx)
}

func (i *ColorPicker) isInnerFocused(gtx layout.Context) bool {
	return i.redInput.IsFocused(gtx) || i.greenInput.IsFocused(gtx) || i.blueInput.IsFocused(gtx)
}

func (i *ColorPicker) SetFocused(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: &i.editor})
}

func (i *ColorPicker) IsFocused(gtx layout.Context) bool {
	return gtx.Focused(&i.editor) || i.opened
}

func (i *ColorPicker) SetValue(c color.NRGBA) {
	i.current = c
	i.editor.SetText(fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B))
	i.updateControls()
}

func (i *ColorPicker) Value() color.NRGBA {
	return i.current
}

func drawImage(gtx layout.Context, img image.Image) layout.Dimensions {
	size := img.Bounds().Size()

	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()

	paint.NewImageOp(img).Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	return layout.Dimensions{Size: size}
}

func colorSquare(hue float64, size image.Point) *image.RGBA {
	img := image.NewRGBA(image.Rectangle{Max: size})
	for y := range size.Y {
		v := 1 - float64(y)/float64(size.Y-1)
		for x := range size.X {
			s := float64(x) / float64(size.X-1)
			var c color.NRGBA
			if hue < 0 {
				// achromatic - saturation has no effect...
				c = hsv(0, 0, v)
			} else {
				c = hsv(hue, s, v)
			}
			img.Set(x, y, c)
		}
	}
	return img
}

func hsv(h, s, v float64) color.NRGBA {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return color.NRGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}

func hue(c color.NRGBA) float64 {
	if c.R == c.G && c.G == c.B {
		return -1
	}
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	mx := math.Max(r, math.Max(g, b))
	mn := math.Min(r, math.Min(g, b))
	delta := mx - mn
	if delta == 0 {
		return 0
	}
	var h float64
	switch mx {
	case r:
		h = 60 * math.Mod((g-b)/delta, 6)
	case g:
		h = 60 * ((b-r)/delta + 2)
	default:
		h = 60 * ((r-g)/delta + 4)
	}
	if h < 0 {
		h += 360
	}
	return h
}

func rgbToHSV(c color.NRGBA) (h, s, v float64) {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	mx := math.Max(r, math.Max(g, b))
	mn := math.Min(r, math.Min(g, b))
	delta := mx - mn
	v = mx
	if c.R == c.G && c.G == c.B {
		return -1, 0, v
	}
	s = delta / mx
	if delta == 0 {
		return 0, 0, v
	}
	switch mx {
	case r:
		h = 60 * math.Mod((g-b)/delta, 6)
	case g:
		h = 60 * ((b-r)/delta + 2)
	default:
		h = 60 * ((r-g)/delta + 4)
	}
	if h < 0 {
		h += 360
	}
	return h, s, v
}
