package controls

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget/material"
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

func (n *NumberInput[T]) SetValue(v T) *NumberInput[T] {
	n.input.SetText(n.toString(v))
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
