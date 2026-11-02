package controls

import (
	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
	"image"
	"strconv"
)

type NumberInput[T int | uint | float64] struct {
	input      *Input
	onChangeFn func(v T)
	minValue   T
	maxValue   T
	incSize    T
	pagingSize T
	precision  int
	btnUp      Clickable
	btnDown    Clickable
}

func NewNumberInput[T int | uint | float64](theme *material.Theme, maxLen int, minValue, maxValue T, pagingSize T, onChangeFn func(v T)) *NumberInput[T] {
	result := &NumberInput[T]{
		onChangeFn: onChangeFn,
		minValue:   minValue,
		maxValue:   maxValue,
		incSize:    1,
		pagingSize: pagingSize,
		precision:  -1,
	}
	result.input = NewInput(theme, filterFromType(minValue), maxLen, func(text string) {
		if v, err := result.fromString(text); err == nil {
			if result.onChangeFn != nil {
				result.onChangeFn(v)
			}
		}
	}).UpDownSupport(result.upDown)
	return result
}

func (n *NumberInput[T]) WithSpinner(on bool) *NumberInput[T] {
	if on {
		n.input.decorator = n.upDownSpinner
	} else {
		n.input.decorator = nil
	}
	return n
}

func (n *NumberInput[T]) upDownSpinner(gtx layout.Context) layout.Dimensions {
	changed := false
	var newValue T
	if n.btnUp.Clicked(gtx) {
		if v, err := n.fromString(n.input.Text()); err == nil {
			if v < n.maxValue {
				changed = true
				newValue = v + n.incSize
			}
		} else {
			changed = true
			newValue = n.maxValue
		}
		n.input.SetFocused(gtx)
	}
	if n.btnDown.Clicked(gtx) {
		if v, err := n.fromString(n.input.Text()); err == nil {
			if v > n.minValue {
				changed = true
				newValue = v - n.incSize
			}
		} else {
			changed = true
			newValue = n.minValue
		}
		n.input.SetFocused(gtx)
	}
	if changed {
		n.input.SetText(n.toString(newValue))
		if n.onChangeFn != nil {
			n.onChangeFn(newValue)
		}
	}
	// width is half the height plus the inset...
	halfHt := (gtx.Constraints.Max.Y + 1) / 2
	dims := layout.Dimensions{Size: gtx.Constraints.Max}
	inset := gtx.Dp(2)
	dims.Size.X = halfHt + (inset * 2)
	size := image.Point{X: halfHt, Y: halfHt}
	al := float32(inset)
	ar := al + float32(halfHt-inset)
	ah := ((ar - al) / 2) + float32(inset)
	// shift everything across to allow for border
	dp1 := float32(gtx.Dp(1))
	al, ar, ah = al+dp1, ar+dp1, ah+dp1
	layout.Flex{Gap: 0, Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max, gtx.Constraints.Min = size, size
			if n.btnUp.Hovered() {
				paint.FillShape(gtx.Ops, Style.SelectedColor, clip.Rect{Max: image.Point{X: size.X + inset*2, Y: size.Y}}.Op())
			}
			return n.btnUp.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				var p clip.Path
				p.Begin(gtx.Ops)
				p.MoveTo(f32.Point{X: ah, Y: float32(inset)})
				p.LineTo(f32.Point{X: al, Y: float32(size.Y - inset)})
				p.LineTo(f32.Point{X: ar, Y: float32(size.Y - inset)})
				p.Close()
				paint.FillShape(gtx.Ops, Style.ForegroundColor, clip.Outline{Path: p.End()}.Op())
				return layout.Dimensions{Size: size}
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max, gtx.Constraints.Min = size, size
			if n.btnDown.Hovered() {
				paint.FillShape(gtx.Ops, Style.SelectedColor, clip.Rect{Max: image.Point{X: size.X + inset*2, Y: size.Y}}.Op())
			}
			return n.btnDown.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				var p clip.Path
				p.Begin(gtx.Ops)
				p.MoveTo(f32.Point{X: al, Y: float32(inset)})
				p.LineTo(f32.Point{X: ar, Y: float32(inset)})
				p.LineTo(f32.Point{X: ah, Y: float32(size.Y - inset)})
				p.Close()
				paint.FillShape(gtx.Ops, Style.ForegroundColor, clip.Outline{Path: p.End()}.Op())
				return layout.Dimensions{Size: size}
			})
		}),
	)
	return dims
}

func (n *NumberInput[T]) SetValue(v T) *NumberInput[T] {
	n.input.SetText(n.toString(v))
	return n
}

func (n *NumberInput[T]) SetText(v string) *NumberInput[T] {
	n.input.SetText(v)
	return n
}

func (n *NumberInput[T]) SetFocused(gtx layout.Context) {
	n.input.SetFocused(gtx)
}

func (n *NumberInput[T]) IsFocused(gtx layout.Context) bool {
	return n.input.IsFocused(gtx)
}

func (n *NumberInput[T]) Layout(gtx layout.Context) layout.Dimensions {
	return n.input.Layout(gtx)
}

func (n *NumberInput[T]) Update(gtx layout.Context) {
	n.input.Update(gtx)
}

func (n *NumberInput[T]) upDown(k key.Name, text string) (string, bool) {
	if v, err := n.fromString(text); err == nil {
		switch k {
		case key.NameDownArrow:
			if v-n.incSize >= n.minValue {
				v -= n.incSize
				n.input.SetText(n.toString(v))
			}
		case key.NameUpArrow:
			if v+n.incSize <= n.maxValue {
				v += n.incSize
				n.input.SetText(n.toString(v))
			}
		case key.NamePageDown:
			if n.pagingSize > 0 {
				if v-n.pagingSize >= n.minValue {
					v -= n.pagingSize
					n.input.SetText(n.toString(v))
				} else if v > n.minValue {
					v = n.minValue
					n.input.SetText(n.toString(v))
				}
			}
		case key.NamePageUp:
			if v+n.pagingSize > n.maxValue {
				v = n.maxValue
				n.input.SetText(n.toString(v))
			} else if v < n.maxValue {
				v += n.pagingSize
				n.input.SetText(n.toString(v))
			}
		}
	}
	return "", false
}

func (n *NumberInput[T]) Text() string {
	return n.input.Text()
}

func (n *NumberInput[T]) Current() T {
	v, _ := n.fromString(n.input.editor.Text())
	return v
}

func (n *NumberInput[T]) fromString(s string) (T, error) {
	var zero T
	switch any(zero).(type) {
	case int:
		v, err := strconv.Atoi(s)
		return any(v).(T), err
	case uint:
		v, err := strconv.ParseUint(s, 10, 0)
		return any(uint(v)).(T), err

	case float64:
		v, err := strconv.ParseFloat(s, 64)
		return any(v).(T), err
	}
	panic("unreachable")
}

func (n *NumberInput[T]) toString(v T) string {
	switch x := any(v).(type) {
	case int:
		return strconv.Itoa(x)
	case uint:
		return strconv.FormatUint(uint64(x), 10)
	case float64:
		return strconv.FormatFloat(x, 'f', n.precision, 64)
	}
	panic("unreachable")
}
