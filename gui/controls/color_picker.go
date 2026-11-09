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
	theme            *material.Theme
	dims             layout.Dimensions
	onChangeFn       func(value color.NRGBA)
	editor           widget.Editor
	focus            bool
	editorStyle      material.EditorStyle
	current          color.NRGBA
	undo             color.NRGBA
	hue              float32
	saturation       float32
	brightness       float32
	dropdown         Clickable
	button           Clickable
	squareClickable  Clickable
	opened           bool
	opening          bool
	dropdownUp       bool
	dropdownFocus    func(layout.Context)
	redLabel         material.LabelStyle
	redInput         *NumberInput[int]
	greenLabel       material.LabelStyle
	greenInput       *NumberInput[int]
	blueLabel        material.LabelStyle
	blueInput        *NumberInput[int]
	hueSlider        *Slider
	hueLabel         material.LabelStyle
	saturationSlider *Slider
	saturationLabel  material.LabelStyle
	brightnessSlider *Slider
	brightnessLabel  material.LabelStyle
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
		current:         clr,
		undo:            clr,
		onChangeFn:      fn,
		redLabel:        material.Label(theme, theme.TextSize, "R:"),
		greenLabel:      material.Label(theme, theme.TextSize, "G:"),
		blueLabel:       material.Label(theme, theme.TextSize, "B:"),
		hueLabel:        material.Label(theme, theme.TextSize, "Hue:"),
		saturationLabel: material.Label(theme, theme.TextSize, "S%:"),
		brightnessLabel: material.Label(theme, theme.TextSize, "V%:"),
	}
	result.editorStyle = material.Editor(theme, &result.editor, "")
	result.editorStyle.Font = *Style.MonoFont
	result.editorStyle.TextSize = (result.editorStyle.TextSize * 14) / 15
	result.editor.SetText(hexColor(clr))
	result.redInput = NewNumberInput(theme, 3, 0, 255, 16, func(v int) {
		if v != int(result.current.R) {
			result.componentChanged(red, v)
		}
	}).SetValue(int(clr.R)).WithSpinner(true).OnSubmit(func(_ string) {
		result.dropdownFocus = result.redInput.SetFocused
		result.focus = true
		result.opened = false
	})
	result.greenInput = NewNumberInput(theme, 3, 0, 255, 16, func(v int) {
		if v != int(result.current.G) {
			result.componentChanged(green, v)
		}
	}).SetValue(int(clr.G)).WithSpinner(true).OnSubmit(func(_ string) {
		result.dropdownFocus = result.greenInput.SetFocused
		result.focus = true
		result.opened = false
	})
	result.blueInput = NewNumberInput(theme, 3, 0, 255, 16, func(v int) {
		if v != int(result.current.B) {
			result.componentChanged(blue, v)
		}
	}).SetValue(int(clr.B)).WithSpinner(true).OnSubmit(func(_ string) {
		result.dropdownFocus = result.blueInput.SetFocused
		result.focus = true
		result.opened = false
	})
	result.dropdownFocus = result.redInput.SetFocused
	h, s, v := toHsv(clr)
	result.hue, result.saturation, result.brightness = h, s, v
	result.hueLabel.Alignment = text.End
	result.saturationLabel.Alignment = text.End
	result.brightnessLabel.Alignment = text.End
	result.hueSlider = NewSlider(theme, layout.Horizontal, h/360).
		OnSubmit(func(f float32) {
			result.dropdownFocus = result.saturationSlider.SetFocused
			result.focus = true
			result.opened = false
		}).Square().Increments(0.011, 0.088)
	result.hueSlider.bgDrawer = func(gtx layout.Context, dims image.Rectangle) image.Image {
		img := image.NewRGBA(image.Rectangle{Max: dims.Max})
		margin := gtx.Dp(4)
		for x := range dims.Max.X - (margin * 2) {
			h := float32(x) * 360 / float32(dims.Max.X-1)
			c := toRgb(h, 1, 1)
			for y := range dims.Max.Y - (margin * 2) {
				img.Set(x+margin, y+margin, c)
			}
		}
		return img
	}
	result.saturationSlider = NewSlider(theme, layout.Horizontal, s).
		OnSubmit(func(_ float32) {
			result.dropdownFocus = result.saturationSlider.SetFocused
			result.focus = true
			result.opened = false
		}).Square().Increments(0.005, 0.0624)
	result.brightnessSlider = NewSlider(theme, layout.Horizontal, v).
		OnSubmit(func(_ float32) {
			result.dropdownFocus = result.brightnessSlider.SetFocused
			result.focus = true
			result.opened = false
		}).Square().Increments(0.0039, 0.0624)
	result.brightnessSlider.next = &result.editor
	return result
}

func (i *ColorPicker) Direction(up bool) *ColorPicker {
	i.dropdownUp = up
	return i
}

func (i *ColorPicker) measureEditor(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min = image.Point{}
	gtx.Constraints.Max = image.Point{X: 1e6, Y: 1e6}
	macro := op.Record(gtx.Ops)
	lbl := material.Label(i.theme, i.theme.TextSize, "#CCCCCC")
	lbl.Font = *Style.MonoFont
	dims := lbl.Layout(gtx)
	_ = macro.Stop()
	return dims
}

func (i *ColorPicker) Layout(gtx layout.Context) layout.Dimensions {
	i.update(gtx)
	if i.focus {
		i.focus = false
		gtx.Execute(key.FocusCmd{Tag: &i.editor})
	}
	focused := gtx.Focused(&i.editor)
	dropFocused := i.isDropdownFocused(gtx)
	bc, bt := Style.bordering(focused && !dropFocused)
	txtDims := i.measureEditor(gtx)
	i.dims = widget.Border{Color: bc, CornerRadius: 3, Width: bt}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: 2, Bottom: 2, Left: 4, Right: 6}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Gap: 0, Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return i.button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						dims := layout.Dimensions{Size: image.Point{X: txtDims.Size.Y, Y: txtDims.Size.Y}}
						paint.FillShape(gtx.Ops, i.current, clip.Rect{Max: dims.Size}.Op())
						borders(gtx, dims, Style.ForegroundColor, true, true, true, true)
						if focused || dropFocused {
							pointer.CursorPointer.Add(gtx.Ops)
						} else {
							pointer.CursorText.Add(gtx.Ops)
						}
						dims.Size.X += 10
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
	if i.opening {
		if i.dropdownFocus != nil {
			i.dropdownFocus(gtx)
		}
		i.opening = false
	}
	i.handleUndoKeys(gtx, i.redInput.tag(), i.greenInput.tag(), i.blueInput.tag(), i.hueSlider, i.saturationSlider, i.brightnessSlider)
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
	gtx.Constraints.Max.Y = 1e6 // no effective max Y
	txtDims := measureText(gtx, i.theme, "\u2003")
	maxLabel := measureMaxText(gtx, i.theme, "S%:", "V%:", "Hue:")
	width := txtDims.Size.X * 20
	if i.hueSlider.Update(gtx) {
		if i.hue <= 0 {
			i.hue = i.hueSlider.Value() * 360
			_, _, v := toHsv(i.current)
			i.current = toRgb(i.hue, 1-v, v)
			_, i.saturation, i.brightness = toHsv(i.current)
		} else {
			i.hue = i.hueSlider.Value() * 360
			i.current = toRgb(i.hue, i.saturation, i.brightness)
		}
		if max(i.current.R, i.current.G, i.current.B) == 0 {
			// black - set max saturation & min brightness...
			i.current = toRgb(i.hue, 1, 0.005)
			_, i.saturation, i.brightness = toHsv(i.current)
		} else if min(i.current.R, i.current.G, i.current.B) == 255 {
			i.current = toRgb(i.hue, 1, 0.995)
			_, i.saturation, i.brightness = toHsv(i.current)
		}
		i.editor.SetText(hexColor(i.current))
		i.saturationSlider.SetValue(i.saturation)
		i.brightnessSlider.SetValue(i.brightness)
		i.redInput.SetValue(int(i.current.R))
		i.greenInput.SetValue(int(i.current.G))
		i.blueInput.SetValue(int(i.current.B))
		if i.onChangeFn != nil {
			i.onChangeFn(i.current)
		}
	}
	if i.saturationSlider.Update(gtx) {
		i.saturation = i.saturationSlider.Value()
		i.current = toRgb(i.hue, i.saturation, i.brightness)
		i.editor.SetText(hexColor(i.current))
		i.redInput.SetValue(int(i.current.R))
		i.greenInput.SetValue(int(i.current.G))
		i.blueInput.SetValue(int(i.current.B))
		h, _, v := toHsv(i.current)
		i.hueSlider.SetValue(h / 360)
		i.brightnessSlider.SetValue(v)
		if i.onChangeFn != nil {
			i.onChangeFn(i.current)
		}
	}
	if i.brightnessSlider.Update(gtx) {
		v := i.brightnessSlider.Value()
		if i.hue < 0 {
			// retain greyscale...
			level := uint8(math.Round(float64(v) * 255))
			i.current = color.NRGBA{R: level, G: level, B: level, A: 255}
		} else {
			i.brightness = v
			i.current = toRgb(i.hue, i.saturation, i.brightness)
		}
		i.editor.SetText(hexColor(i.current))
		i.redInput.SetValue(int(i.current.R))
		i.greenInput.SetValue(int(i.current.G))
		i.blueInput.SetValue(int(i.current.B))
		h, _, _ := toHsv(i.current)
		i.hueSlider.SetValue(h / 360)
		if i.onChangeFn != nil {
			i.onChangeFn(i.current)
		}
	}
	return i.dropdown.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
							return layout.Flex{Gap: 10, Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Gap: 8, Axis: layout.Horizontal}.Layout(gtx,
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return layout.Inset{Top: 2}.Layout(gtx, i.redLabel.Layout)
										}),
										layout.Rigid(i.redInput.Layout),
									)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Gap: 8, Axis: layout.Horizontal}.Layout(gtx,
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return layout.Inset{Top: 2}.Layout(gtx, i.greenLabel.Layout)
										}),
										layout.Rigid(i.greenInput.Layout),
									)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Gap: 8, Axis: layout.Horizontal}.Layout(gtx,
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return layout.Inset{Top: 2}.Layout(gtx, i.blueLabel.Layout)
										}),
										layout.Rigid(i.blueInput.Layout),
									)
								}),
							)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Gap: 8, Axis: layout.Horizontal}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Min.X, gtx.Constraints.Max.X = maxLabel, maxLabel
									return i.hueLabel.Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									wd := innerWidth - 8 - maxLabel
									gtx.Constraints.Min.X, gtx.Constraints.Max.X = wd, wd
									gtx.Constraints.Min.Y, gtx.Constraints.Max.Y = txtDims.Size.Y, txtDims.Size.Y
									return i.hueSlider.Layout(gtx)
								}),
							)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Gap: 8, Axis: layout.Horizontal}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Min.X, gtx.Constraints.Max.X = maxLabel, maxLabel
									return i.saturationLabel.Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									wd := innerWidth - 8 - maxLabel
									gtx.Constraints.Min.X, gtx.Constraints.Max.X = wd, wd
									gtx.Constraints.Min.Y, gtx.Constraints.Max.Y = txtDims.Size.Y, txtDims.Size.Y
									return i.saturationSlider.Layout(gtx)
								}),
							)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Gap: 8, Axis: layout.Horizontal}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Min.X, gtx.Constraints.Max.X = maxLabel, maxLabel
									return i.brightnessLabel.Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									wd := innerWidth - 8 - maxLabel
									gtx.Constraints.Min.X, gtx.Constraints.Max.X = wd, wd
									gtx.Constraints.Min.Y, gtx.Constraints.Max.Y = txtDims.Size.Y, txtDims.Size.Y
									return i.brightnessSlider.Layout(gtx)
								}),
							)
						}),
					)
				})
			}),
		)
	})
}

func (i *ColorPicker) layoutColorSquare(gtx layout.Context) layout.Dimensions {
	img := i.colorSquareImage(gtx)
	size := img.Bounds().Size()
	dims := image.Point{X: size.X, Y: size.Y}
	margin := gtx.Dp(squareMargin)
	for {
		ev, ok := gtx.Event(pointer.Filter{
			Target:  &i.squareClickable,
			Kinds:   pointer.Press | pointer.Drag | pointer.Scroll,
			ScrollX: pointer.ScrollRange{Min: -img.Bounds().Dx(), Max: img.Bounds().Dx()},
			ScrollY: pointer.ScrollRange{Min: -img.Bounds().Dy(), Max: img.Bounds().Dy()},
		})
		if !ok {
			break
		}
		if pev, ok := ev.(pointer.Event); ok {
			switch pev.Kind {
			case pointer.Press, pointer.Drag:
				px := int(pev.Position.X - float32(margin) + squareMarkerRadius)
				py := int(pev.Position.Y - float32(margin) + squareMarkerRadius)
				maxW := size.X - margin - 1
				maxH := size.Y - margin - 1
				px = max(margin, min(px, maxW))
				py = max(margin, min(py, maxH))
				c := img.RGBAAt(px, py)
				// check we're not into margin...
				if c.A != 0 {
					i.current = color.NRGBA{R: c.R, G: c.G, B: c.B, A: 255}
					i.hue, i.saturation, i.brightness = toHsv(i.current)
					i.editor.SetText(hexColor(i.current))
					i.updateControls()
					if i.onChangeFn != nil {
						i.onChangeFn(i.current)
					}
					gtx.Execute(op.InvalidateCmd{})
				}
			case pointer.Scroll:
				switch {
				case pev.Scroll.X < 0:
					i.saturation = max(0, i.saturation-0.005)
					i.current = toRgb(i.hue, i.saturation, i.brightness)
					i.updateControls()
					i.editor.SetText(hexColor(i.current))
					if i.onChangeFn != nil {
						i.onChangeFn(i.current)
					}
				case pev.Scroll.X > 0:
					i.saturation = min(1, i.saturation+0.005)
					i.current = toRgb(i.hue, i.saturation, i.brightness)
					i.updateControls()
					i.editor.SetText(hexColor(i.current))
					if i.onChangeFn != nil {
						i.onChangeFn(i.current)
					}
				case pev.Scroll.Y < 0:
					i.brightness = min(1, i.brightness+0.005)
					i.current = toRgb(i.hue, i.saturation, i.brightness)
					i.updateControls()
					i.editor.SetText(hexColor(i.current))
					if i.onChangeFn != nil {
						i.onChangeFn(i.current)
					}
				case pev.Scroll.Y > 0:
					i.brightness = max(0, i.brightness-0.005)
					i.current = toRgb(i.hue, i.saturation, i.brightness)
					i.updateControls()
					i.editor.SetText(hexColor(i.current))
					if i.onChangeFn != nil {
						i.onChangeFn(i.current)
					}
				}
			}
		}
	}
	return i.squareClickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
		paint.NewImageOp(img).Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		// marker position in displayed coordinates
		x, y := i.colorSquarePosition(gtx)
		drawMarker(gtx, x, y, perceptualLightness(i.current) < 0.45)
		pointer.CursorPointer.Add(gtx.Ops)
		return layout.Dimensions{Size: dims}
	})
}

func (i *ColorPicker) colorSquarePosition(gtx layout.Context) (x, y int) {
	h, s, v := toHsv(i.current)
	margin := gtx.Dp(squareMargin)
	wd := gtx.Constraints.Max.X - margin*2
	if h < 0 {
		x = margin + wd/2
	} else {
		x = margin + int(math.Round(float64(s)*float64(wd-1)))
	}
	ht := float64(gtx.Constraints.Max.X/2-margin*2) - 1
	y = margin + int(math.Round((1-float64(v))*ht))
	return x, y
}

const (
	squareMargin       = 5
	squareMarkerRadius = 4
)

func drawMarker(gtx layout.Context, x, y int, light bool) {
	radius := gtx.Dp(squareMarkerRadius)
	circle := clip.Ellipse{
		Min: image.Pt(x-radius, y-radius),
		Max: image.Pt(x+radius, y+radius),
	}
	c := color.NRGBA{A: 255}
	if light {
		c = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	}
	paint.FillShape(gtx.Ops, c,
		clip.Stroke{
			Path:  circle.Path(gtx.Ops),
			Width: 2,
		}.Op(),
	)
}

func (i *ColorPicker) colorSquareImage(gtx layout.Context) *image.RGBA {
	size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Max.X / 2}
	r := image.Rectangle{Max: size}
	img := image.NewRGBA(r)
	draw.Draw(img, r, &image.Uniform{C: color.NRGBA{}}, image.Point{}, draw.Src)
	margin := gtx.Dp(squareMargin)
	wd := size.X - (margin * 2)
	ht := size.Y - (margin * 2)
	h, _, _ := toHsv(i.current)
	for y := range ht {
		v := 1 - float32(y)/float32(ht-1)
		for x := range wd {
			s := float32(x) / float32(wd-1)
			var c color.NRGBA
			if h < 0 {
				// achromatic - saturation has no effect...
				c = toRgb(0, 0, v)
			} else {
				c = toRgb(h, s, v)
			}
			img.Set(x+margin, y+margin, c)
		}
	}
	return img
}

func (i *ColorPicker) update(gtx layout.Context) {
	if i.button.Clicked(gtx) {
		if i.opened || !gtx.Focused(&i.editor) {
			i.opened = false
			gtx.Execute(key.FocusCmd{Tag: &i.editor})
		} else {
			i.opened = true
			i.opening = true
		}
	}
	i.handleUndoKeys(gtx, &i.editor)
	i.handleUpDownKeys(gtx)
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
				if i.dropdownFocus != nil {
					i.dropdownFocus(gtx)
				}
			} else {
				i.opened = false
			}
		}
	}
}

func (i *ColorPicker) handleUndoKeys(gtx layout.Context, tags ...any) {
	seen := false
	for _, tag := range tags {
		for {
			ev, ok := gtx.Event(
				key.Filter{Focus: tag, Name: "Z", Required: key.ModCommand},
				key.Filter{Focus: tag, Name: "Z", Required: key.ModCtrl},
			)
			if !ok {
				break
			}
			if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
				i.current = i.undo
				i.updateControls()
				i.editor.SetText(hexColor(i.current))
				gtx.Execute(op.InvalidateCmd{})
				seen = true
				break
			}
		}
		if seen {
			break
		}
	}
}

func (i *ColorPicker) handleUpDownKeys(gtx layout.Context) {
	if !gtx.Focused(&i.editor) {
		return
	}
	curr := i.editor.Text()
	if !hashColorRegex.MatchString(curr) {
		return
	}
	selStart, selEnd := i.editor.Selection()
	start, end := min(selStart, selEnd), max(selStart, selEnd)
	if start == end {
		end++
	}
	if start < 1 || end > 7 {
		return
	}
	for {
		ev, ok := gtx.Event(
			key.Filter{Focus: &i.editor, Name: key.NameUpArrow},
			key.Filter{Focus: &i.editor, Name: key.NameDownArrow})
		if !ok {
			break
		}
		if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
			switch kev.Name {
			case key.NameUpArrow:
				v, _ := strconv.ParseInt(curr[start:end], 16, 64)
				cv := int(v)
				if cv < (1<<((end-start)*4))-1 {
					cv++
				} else {
					cv = 0
				}
				nv := curr[:start] + fmt.Sprintf("%0"+strconv.Itoa(end-start)+"X", cv) + curr[end:]
				i.editor.SetText(nv)
				i.editor.SetCaret(selStart, selEnd)
				r, _ := strconv.ParseUint(nv[1:3], 16, 8)
				g, _ := strconv.ParseUint(nv[3:5], 16, 8)
				b, _ := strconv.ParseUint(nv[5:7], 16, 8)
				i.current = color.NRGBA{
					R: uint8(r),
					G: uint8(g),
					B: uint8(b),
					A: 0xff,
				}
				i.hue, i.saturation, i.brightness = toHsv(i.current)
				i.updateControls()
				if i.onChangeFn != nil {
					i.onChangeFn(i.current)
				}
			case key.NameDownArrow:
				v, _ := strconv.ParseInt(curr[start:end], 16, 64)
				cv := int(v)
				if cv > 0 {
					cv--
				} else {
					cv = (1 << ((end - start) * 4)) - 1
				}
				nv := curr[:start] + fmt.Sprintf("%0"+strconv.Itoa(end-start)+"X", cv) + curr[end:]
				i.editor.SetText(nv)
				i.editor.SetCaret(selStart, selEnd)
				r, _ := strconv.ParseUint(nv[1:3], 16, 8)
				g, _ := strconv.ParseUint(nv[3:5], 16, 8)
				b, _ := strconv.ParseUint(nv[5:7], 16, 8)
				i.current = color.NRGBA{
					R: uint8(r),
					G: uint8(g),
					B: uint8(b),
					A: 0xff,
				}
				i.hue, i.saturation, i.brightness = toHsv(i.current)
				i.updateControls()
				if i.onChangeFn != nil {
					i.onChangeFn(i.current)
				}
			}
		}
	}

}

var (
	hashColorRegex = regexp.MustCompile("^#[0-9a-fA-F]{6}$")
	colorRegex     = regexp.MustCompile("^[0-9a-fA-F]{6}$")
)

func (i *ColorPicker) changed() {
	s := strings.TrimPrefix(i.editor.Text(), "#")
	if len(s) == 3 {
		s = s[0:1] + s[0:1] + s[1:2] + s[1:2] + s[2:3] + s[2:3]
	}
	if len(s) < 6 {
		s = s + strings.Repeat("0", 6-len(s))
	}
	actualChange := true
	if colorRegex.MatchString(s) {
		r, _ := strconv.ParseUint(s[0:2], 16, 8)
		g, _ := strconv.ParseUint(s[2:4], 16, 8)
		b, _ := strconv.ParseUint(s[4:6], 16, 8)
		if i.current.R == uint8(r) && i.current.G == uint8(g) && i.current.B == uint8(b) {
			actualChange = false
		} else {
			i.current = color.NRGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
				A: 0xff,
			}
		}
	} else {
		i.current = color.NRGBA{A: 255}
	}
	if actualChange {
		i.hue, i.saturation, i.brightness = toHsv(i.current)
		i.updateControls()
		if i.onChangeFn != nil {
			i.onChangeFn(i.current)
		}
	}
}

func (i *ColorPicker) updateControls() {
	if i.opened {
		i.redInput.SetValue(int(i.current.R))
		i.greenInput.SetValue(int(i.current.G))
		i.blueInput.SetValue(int(i.current.B))
		h, s, v := toHsv(i.current)
		i.hueSlider.SetValue(h / 360)
		i.saturationSlider.SetValue(s)
		i.brightnessSlider.SetValue(v)
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
		i.hue, i.saturation, i.brightness = toHsv(i.current)
		i.hueSlider.SetValue(i.hue / 360)
		i.saturationSlider.SetValue(i.saturation)
		i.brightnessSlider.SetValue(i.brightness)
		i.editor.SetText(hexColor(i.current))
		if i.onChangeFn != nil {
			i.onChangeFn(i.current)
		}
	}
}

func (i *ColorPicker) isFocused(gtx layout.Context) bool {
	return gtx.Focused(&i.editor) || i.isDropdownFocused(gtx)
}

func (i *ColorPicker) isDropdownFocused(gtx layout.Context) bool {
	return i.redInput.IsFocused(gtx) || i.greenInput.IsFocused(gtx) || i.blueInput.IsFocused(gtx) ||
		i.hueSlider.IsFocused(gtx) || i.saturationSlider.IsFocused(gtx) || i.brightnessSlider.IsFocused(gtx)
}

func (i *ColorPicker) SetFocused(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: &i.editor})
}

func (i *ColorPicker) IsFocused(gtx layout.Context) bool {
	return gtx.Focused(&i.editor) || i.opened
}

func (i *ColorPicker) SetValue(c color.NRGBA) {
	i.current = c
	i.undo = c
	i.hue, i.saturation, i.brightness = toHsv(i.current)
	i.editor.SetText(hexColor(c))
	i.updateControls()
}

func (i *ColorPicker) Value() color.NRGBA {
	return i.current
}

func toRgb(hue, saturation, brightness float32) color.NRGBA {
	hue = float32(math.Mod(float64(hue), 360))
	if hue < 0 {
		hue += 360
	}
	c := brightness * saturation
	x := c * (1 - float32(math.Abs(math.Mod(float64(hue)/60, 2)-1)))
	m := brightness - c
	var r, g, b float32
	switch {
	case hue < 60:
		r, g, b = c, x, 0
	case hue < 120:
		r, g, b = x, c, 0
	case hue < 180:
		r, g, b = 0, c, x
	case hue < 240:
		r, g, b = 0, x, c
	case hue < 300:
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

func toHsv(c color.NRGBA) (hue, saturation, brightness float32) {
	r := float32(c.R) / 255
	g := float32(c.G) / 255
	b := float32(c.B) / 255
	mx := max(r, g, b)
	mn := min(r, g, b)
	delta := mx - mn
	brightness = mx
	if c.R == c.G && c.G == c.B {
		return -1, 0, brightness
	}
	saturation = delta / mx
	if delta == 0 {
		return 0, 0, brightness
	}
	switch mx {
	case r:
		hue = 60 * float32(math.Mod(float64((g-b)/delta), 6))
	case g:
		hue = 60 * ((b-r)/delta + 2)
	default:
		hue = 60 * ((r-g)/delta + 4)
	}
	if hue < 0 {
		hue += 360
	}
	return hue, saturation, brightness
}

func perceptualLightness(c color.NRGBA) float32 {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	linear := func(v float64) float64 {
		if v <= 0.04045 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return float32(
		0.2126*linear(r) +
			0.7152*linear(g) +
			0.0722*linear(b),
	)
}

func hexColor(c color.NRGBA) string {
	const hex = "0123456789ABCDEF"
	return string([]byte{
		'#',
		hex[c.R>>4], hex[c.R&0x0F],
		hex[c.G>>4], hex[c.G&0x0F],
		hex[c.B>>4], hex[c.B&0x0F],
	})
}
