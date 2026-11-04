package controls

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"reflect"
	"unsafe"
)

func NewRadioButton(theme *material.Theme, enum *widget.Enum, key string, label string) *RadioButton {
	s := material.RadioButton(theme, enum, key, label)
	s.Size = 18
	s.TextSize = theme.TextSize
	return &RadioButton{
		Key:   key,
		style: s,
	}
}

type RadioButton struct {
	Key   string
	style material.RadioButtonStyle
}

func (b *RadioButton) Layout(gtx layout.Context) layout.Dimensions {
	return b.style.Layout(gtx)
}

func (b *RadioButton) SetFocused(gtx layout.Context) {
	if tag := enumTag(b.style.Group, b.Key); tag != nil {
		gtx.Execute(key.FocusCmd{Tag: tag})
	}
}

// HandleRadioKeys adds left/right/up/down key support to radio buttons
//
// the radios are specified by group and in logical order
func HandleRadioKeys(gtx layout.Context, radios ...*RadioButton) {
	if len(radios) < 2 {
		return
	}
	for i, radio := range radios {
		rn := i
		if tag := enumTag(radio.style.Group, radio.Key); tag != nil {
			if gtx.Focused(tag) {
				for {
					ev, ok := gtx.Event(
						key.Filter{Focus: tag, Name: key.NameLeftArrow},
						key.Filter{Focus: tag, Name: key.NameUpArrow},
						key.Filter{Focus: tag, Name: key.NameRightArrow},
						key.Filter{Focus: tag, Name: key.NameDownArrow},
					)
					if !ok {
						break
					}
					if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
						switch kev.Name {
						case key.NameLeftArrow, key.NameUpArrow:
							var prev *RadioButton
							if rn == 0 {
								prev = radios[len(radios)-1]
							} else {
								prev = radios[rn-1]
							}
							radio.style.Group.Value = prev.Key
							prev.SetFocused(gtx)
						case key.NameRightArrow, key.NameDownArrow:
							var next *RadioButton
							if rn < len(radios)-1 {
								next = radios[rn+1]
							} else {
								next = radios[0]
							}
							radio.style.Group.Value = next.Key
							next.SetFocused(gtx)
						}
					}
				}
			}
		}
	}
}

func enumTag(e *widget.Enum, k string) (result any) {
	defer func() {
		if r := recover(); r != nil {
			result = nil
		}
	}()
	v := reflect.ValueOf(e).Elem()
	keys := v.FieldByName("keys")
	for i := 0; i < keys.Len(); i++ {
		entry := keys.Index(i).Elem()
		if entry.FieldByName("key").String() == k {
			tag := entry.FieldByName("tag")
			result = reflect.NewAt(tag.Type(), unsafe.Pointer(tag.UnsafeAddr())).Elem().Addr().Interface()
			break
		}
	}
	return result
}
