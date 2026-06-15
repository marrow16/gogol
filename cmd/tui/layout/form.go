package layout

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"strconv"
)

type Form[T any] struct {
	FormRows[T]
	focusedInput int
	maxInput     int
	resetInputs  []ResetInput[T]
	currentInput Input[T]
	customInputs map[int]CustomInput[T]
	Style        lipgloss.Style
	FocusedStyle lipgloss.Style
}

func (f *Form[T]) Render(parent T, clickPts ClickPoints[T], sf Surface) *tea.Cursor {
	var csr *tea.Cursor
	f.customInputs = make(map[int]CustomInput[T])
	onInput := 0
	f.currentInput = nil
	f.resetInputs = make([]ResetInput[T], 0)
	delayedSurfaces := map[[2]int]Surface{}
	for r, fr := range f.FormRows {
		for c, el := range fr {
			if el != nil {
				if el.Condition != nil && !el.Condition(parent) {
					continue
				}
				switch it := el.Item.(type) {
				case Input[T]:
					f.resetInputs = append(f.resetInputs, it)
					fcsr, ds := it.Render(parent, f, onInput, sf, clickPts, r, c, onInput == f.focusedInput, f.Style, f.FocusedStyle)
					if onInput == f.focusedInput {
						if ds != nil {
							delayedSurfaces[[2]int{r, c - 1}] = ds
						}
						f.currentInput = it
						csr = fcsr
					}
					onInput++
				case CustomInput[T]:
					f.resetInputs = append(f.resetInputs, it)
					f.customInputs[onInput] = it
					fcsr := it.Render(parent, f, onInput, sf, clickPts, r, c, onInput == f.focusedInput, f.Style, f.FocusedStyle)
					if onInput == f.focusedInput {
						csr = fcsr
					}
					onInput++
				default:
					styles := []lipgloss.Style{f.Style}
					if el.Style != nil {
						styles[0] = *el.Style
					}
					if el.StyleFn != nil {
						if s := el.StyleFn(parent, el); s != nil {
							styles[0] = *s
						}
					}
					var pl Placement
					switch el.Alignment {
					case AlignRight:
						width := el.Width
						if width < 1 {
							width = sf.Width() - c - 2
						}
						if el.OnClick != nil {
							pl = sf.TextRightPad(r, c, width, el.text(parent), styles...)
						} else {
							pl = sf.TextRight(r, c, width, el.text(parent), styles...)
						}
					case AlignCenter:
						width := el.Width
						if width < 1 {
							width = sf.Width() - c - 2
						}
						pl = sf.TextCenter(r, c, width, el.text(parent), styles...)
					default:
						width := el.Width
						if width > 0 {
							pl = sf.TextFixed(r, c, width, el.text(parent), styles...)
						} else {
							pl = sf.Text(r, c, el.text(parent), styles...)
						}
					}
					if el.OnClick != nil {
						clickPts.Add(pl, el.OnClick)
					}
				}
			}
		}
	}
	for pt, rgn := range delayedSurfaces {
		sf.Draw(pt[0], pt[1], rgn)
	}
	f.maxInput = onInput
	return csr
}

func (f *Form[T]) Reset(parent T) {
	f.focusedInput = 0
	f.currentInput = nil
	for _, ri := range f.resetInputs {
		ri.Reset(parent)
	}
}

func (f *Form[T]) SetFocusedInput(input int) {
	f.focusedInput = input
}

func (f *Form[T]) Update(parent T, msg tea.Msg) tea.Cmd {
	for inputNo, custom := range f.customInputs {
		custom.Update(parent, msg, inputNo == f.focusedInput)
	}
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		switch mt.String() {
		case "tab":
			if f.focusedInput < f.maxInput-1 {
				f.focusedInput++
			} else {
				f.focusedInput = 0
			}
		case "shift+tab":
			if f.focusedInput > 0 {
				f.focusedInput--
			} else {
				f.focusedInput = f.maxInput - 1
			}
		default:
			if f.currentInput != nil {
				return f.currentInput.Update(parent, mt)
			}
		}
	default:
		if f.currentInput != nil {
			return f.currentInput.Update(parent, msg)
		}
	}
	return nil
}

type FormRows[T any] []FormRow[T]

type FormRow[T any] []*FormElement[T]

type Alignment int

const (
	AlignLeft = iota
	AlignRight
	AlignCenter
)

type FormElement[T any] struct {
	Item      any // could be text, a number or an input
	OnClick   func(parent T) tea.Cmd
	Style     *lipgloss.Style // could be a fixed style or a func that returns a style based on the item (and some state cargo)
	StyleFn   func(parent T, element *FormElement[T]) *lipgloss.Style
	Width     int // only used when Alignment is center or right
	Alignment Alignment
	Condition func(parent T) bool
}

func (e FormElement[T]) text(parent T) string {
	it := e.Item
	if f, ok := e.Item.(func(parent T) any); ok {
		it = f(parent)
	}
	switch et := it.(type) {
	case func(parent T) string:
	case string:
		return et
	case int:
		return strconv.Itoa(et)
	case float64:
		return strconv.FormatFloat(et, 'g', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(et), 'g', -1, 64)
	case bool:
		if et {
			return "Y"
		}
		return "N"
	}
	return ""
}
