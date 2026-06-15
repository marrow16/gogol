package layout

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

type ResetInput[T any] interface {
	Reset(parent T)
}

type Input[T any] interface {
	ResetInput[T]
	Align(a Alignment) Input[T]
	Update(parent T, msg tea.Msg) tea.Cmd
	Render(parent T, form *Form[T], inputNo int, sf Surface, clickPts ClickPoints[T], row, col int, focused bool, style lipgloss.Style, focusedStyle lipgloss.Style) (*tea.Cursor, Surface)
}

type CustomInput[T any] interface {
	ResetInput[T]
	Update(parent T, msg tea.Msg, focused bool) tea.Cmd
	Render(parent T, form *Form[T], inputNo int, sf Surface, clickPts ClickPoints[T], row, col int, focused bool, style lipgloss.Style, focusedStyle lipgloss.Style) *tea.Cursor
}

func NewNumberInput[T any](width int, min, max any, getFn func(parent T) int, setFn func(parent T, value int) tea.Cmd) Input[T] {
	return &numberInput[T]{
		width:    width,
		min:      min,
		max:      max,
		getValue: getFn,
		setValue: setFn,
	}
}

type numberInput[T any] struct {
	alignment Alignment
	width     int
	max       any
	min       any
	value     string
	got       bool
	getValue  func(parent T) int
	setValue  func(parent T, value int) tea.Cmd
}

func (n *numberInput[T]) Reset(parent T) {
	n.got = false
}

func (n *numberInput[T]) getMin(parent T) int {
	switch mt := n.min.(type) {
	case int:
		return mt
	case func(parent T) int:
		return mt(parent)
	}
	return 0
}

func (n *numberInput[T]) getMax(parent T) int {
	switch mt := n.max.(type) {
	case int:
		return mt
	case func(parent T) int:
		return mt(parent)
	}
	return 0
}

func (n *numberInput[T]) Align(a Alignment) Input[T] {
	n.alignment = a
	return n
}

func (n *numberInput[T]) Update(parent T, msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		return n.key(parent, mt)
	case tea.PasteMsg:
		return n.paste(parent, mt)
	case tea.MouseWheelMsg:
		switch mt.Mouse().Button {
		case tea.MouseWheelDown:
			return n.key(parent, tea.KeyPressMsg{Text: "down"})
		case tea.MouseWheelUp:
			return n.key(parent, tea.KeyPressMsg{Text: "up"})
		}
	}
	return nil
}

func (n *numberInput[T]) key(parent T, msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "up":
		if n.value == "" {
			v := n.getMax(parent)
			n.value = strconv.Itoa(v)
			return n.setValue(parent, v)
		} else if v, err := strconv.Atoi(n.value); err == nil && v < n.getMax(parent) {
			v++
			n.value = strconv.Itoa(v)
			return n.setValue(parent, v)
		}
	case "down":
		if n.value == "" {
			v := n.getMin(parent)
			n.value = strconv.Itoa(v)
			return n.setValue(parent, v)
		} else if v, err := strconv.Atoi(n.value); err == nil && v > n.getMin(parent) {
			v--
			n.value = strconv.Itoa(v)
			return n.setValue(parent, v)
		}
	case "backspace":
		if len(n.value) > 0 {
			n.value = n.value[:len(n.value)-1]
			if v, err := strconv.Atoi(n.value); err == nil && v >= n.getMin(parent) && v <= n.getMax(parent) {
				n.value = strconv.Itoa(v)
				return n.setValue(parent, v)
			}
		}
	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if v, err := strconv.Atoi(n.value + msg.String()); err == nil && v >= n.getMin(parent) && v <= n.getMax(parent) {
			n.value = strconv.Itoa(v)
			return n.setValue(parent, v)
		}
	}
	return nil
}

func (n *numberInput[T]) paste(parent T, msg tea.PasteMsg) tea.Cmd {
	var sb strings.Builder
	for _, r := range msg.Content {
		if r >= '0' && r <= '9' {
			sb.WriteRune(r)
		}
	}
	if v, err := strconv.Atoi(sb.String()); err == nil && v >= n.getMin(parent) && v <= n.getMax(parent) {
		n.value = strconv.Itoa(v)
		return n.setValue(parent, v)
	}
	return nil
}

func (n *numberInput[T]) Render(parent T, form *Form[T], inputNo int, sf Surface, clickPts ClickPoints[T], row, col int, focused bool, style lipgloss.Style, focusedStyle lipgloss.Style) (*tea.Cursor, Surface) {
	if focused {
		if !n.got {
			n.value = strconv.Itoa(n.getValue(parent))
			n.got = true
		}
		sf.Text(row, col, strings.Repeat(" ", n.width), focusedStyle)
		csrX := len(n.value)
		if csrX >= n.width {
			csrX = n.width - 1
		}
		sf.Text(row, col, n.value, focusedStyle)
		return tea.NewCursor(sf.AbsoluteLeft()+col+csrX, sf.AbsoluteTop()+row), nil
	} else {
		n.got = false
		n.value = strconv.Itoa(n.getValue(parent))
		pl := sf.TextRightPad(row, col, n.width, n.value, style)
		clickPts.Add(pl, func(parent T) tea.Cmd {
			form.SetFocusedInput(inputNo)
			return nil
		})
		return nil, nil
	}
}

func NewTextInput[T any](width int, validChars string, getFn func(parent T) string, setFn func(parent T, value string) tea.Cmd, opts ...bool) Input[T] {
	unlimited := len(opts) > 0 && opts[0]
	openSupport := len(opts) > 1 && opts[1]
	return &textInput[T]{
		width:       width,
		unlimited:   unlimited,
		validChars:  validChars,
		getValue:    getFn,
		setValue:    setFn,
		openSupport: openSupport,
	}
}

type textInput[T any] struct {
	alignment   Alignment
	width       int
	unlimited   bool
	validChars  string
	value       string
	got         bool
	getValue    func(parent T) string
	setValue    func(parent T, value string) tea.Cmd
	openSupport bool
}

func (t *textInput[T]) Reset(parent T) {
	t.got = false
}

func (t *textInput[T]) Align(a Alignment) Input[T] {
	t.alignment = a
	return t
}

func (t *textInput[T]) Update(parent T, msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		return t.key(parent, mt)
	case tea.PasteMsg:
		return t.paste(parent, mt)
	}
	return nil
}

func (t *textInput[T]) open(parent T) tea.Cmd {
	type dummyCmd struct{}
	if runtime.GOOS == "darwin" {
		return func() tea.Msg {
			out, err := exec.Command(
				"osascript",
				"-e",
				`POSIX path of (choose file)`,
			).Output()
			if err == nil {
				t.value = strings.TrimSpace(string(out))
				t.setValue(parent, t.value)
			}
			return dummyCmd{}
		}
	}
	return nil
}

func (t *textInput[T]) key(parent T, msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "backspace":
		if len(t.value) > 0 {
			t.value = t.value[:len(t.value)-1]
			return t.setValue(parent, t.value)
		}
	case "ctrl+f":
		if t.openSupport {
			return t.open(parent)
		}
	default:
		ch := msg.String()
		if ch == "space" {
			ch = " "
		}
		if len(ch) == 1 && (t.unlimited || len(t.value) < t.width) {
			if t.validChars == "" || strings.Contains(t.validChars, ch) {
				t.value += ch
				return t.setValue(parent, t.value)
			}
		}
	}
	return nil
}

func (t *textInput[T]) paste(parent T, msg tea.PasteMsg) tea.Cmd {
	var sb strings.Builder
	if t.validChars == "" {
		sb.WriteString(strings.ReplaceAll(strings.ReplaceAll(msg.Content, "\r", ""), "\n", ""))
	} else {
		for _, r := range msg.Content {
			if strings.ContainsRune(t.validChars, r) {
				sb.WriteRune(r)
			}
		}
	}
	t.value += sb.String()
	return t.setValue(parent, t.value)
}

func (t *textInput[T]) Render(parent T, form *Form[T], inputNo int, sf Surface, clickPts ClickPoints[T], row, col int, focused bool, style lipgloss.Style, focusedStyle lipgloss.Style) (*tea.Cursor, Surface) {
	if t.width < 1 {
		t.width = sf.Width() - col - 1
	}
	if focused {
		if !t.got {
			t.value = t.getValue(parent)
			t.got = true
		}
		csrX := len(t.value)
		if csrX >= t.width {
			csrX = t.width - 1
		}
		sf.TextFixed(row, col, t.width, t.value, focusedStyle)
		return tea.NewCursor(sf.AbsoluteLeft()+col+csrX, sf.AbsoluteTop()+row), nil
	} else {
		t.got = false
		t.value = t.getValue(parent)
		pl := sf.TextFixed(row, col, t.width, MaxWidth(t.value, t.width), style)
		pl.Extent = t.width
		clickPts.Add(pl, func(parent T) tea.Cmd {
			form.SetFocusedInput(inputNo)
			return nil
		})
		return nil, nil
	}
}

func MaxWidth(s string, width int) string {
	if l := utf8.RuneCountInString(s); l > width {
		return string([]rune(s)[:width])
	}
	return s
}

func NewRadio[T any](options []string, getFn func(parent T) int, setFn func(parent T, value int) tea.Cmd) Input[T] {
	return &radio[T]{
		options:  options,
		getValue: getFn,
		setValue: setFn,
	}
}

type radio[T any] struct {
	options  []string
	getValue func(parent T) int
	setValue func(parent T, value int) tea.Cmd
}

func (r *radio[T]) Reset(parent T) {
}

func (r *radio[T]) Align(a Alignment) Input[T] {
	return r
}

func (r *radio[T]) Update(parent T, msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		return r.key(parent, mt)
	case tea.PasteMsg:
		return r.paste(parent, mt)
	}
	return nil
}

func (r *radio[T]) key(parent T, msg tea.KeyPressMsg) tea.Cmd {
	v := r.getValue(parent)
	switch msg.String() {
	case "up", "left":
		if v > 0 {
			v--
		} else {
			v = len(r.options) - 1
		}
		r.setValue(parent, v)
	case "down", "right":
		if v < len(r.options)-1 {
			v++
		} else {
			v = 0
		}
		r.setValue(parent, v)
	case "home":
		r.setValue(parent, 0)
	case "end":
		r.setValue(parent, len(r.options)-1)
	default:
		if ch := strings.ToLower(msg.String()); len(ch) == 1 {
			for i, option := range r.options {
				if strings.HasPrefix(strings.ToLower(option), ch) {
					r.setValue(parent, i)
					break
				}
			}
		}
	}
	return nil
}

func (r *radio[T]) paste(parent T, msg tea.PasteMsg) tea.Cmd {
	return nil
}

func (r *radio[T]) Render(parent T, form *Form[T], inputNo int, sf Surface, clickPts ClickPoints[T], row, col int, focused bool, style lipgloss.Style, focusedStyle lipgloss.Style) (*tea.Cursor, Surface) {
	x := col
	v := r.getValue(parent)
	for i, option := range r.options {
		switch {
		case focused && i == v:
			sf.Text(row, x-1, " "+option+" ", focusedStyle)
		case i == v:
			s := style.Inherit(lipgloss.NewStyle().Underline(true))
			pl := sf.Text(row, x, option, s)
			clickPts.Add(pl, func(parent T) tea.Cmd {
				form.SetFocusedInput(inputNo)
				return nil
			})
		default:
			on := i
			clickPts.Add(sf.Text(row, x, option, style), func(parent T) tea.Cmd {
				form.SetFocusedInput(inputNo)
				return r.setValue(parent, on)
			})
		}
		x += len(option) + 1
	}
	return nil, nil
}

func NewButton[T any](text string, onClick func(parent T) tea.Cmd, opts ...bool) Input[T] {
	big := true
	if len(opts) > 0 {
		big = opts[0]
	}
	return &button[T]{
		big:     big,
		text:    text,
		top:     rndBoxTL + strings.Repeat(boxH, len(text)) + rndBoxTR,
		bottom:  rndBoxBL + strings.Repeat(boxH, len(text)) + rndBoxBR,
		onClick: onClick,
	}
}

type button[T any] struct {
	alignment Alignment
	big       bool
	text      string
	top       string
	bottom    string
	onClick   func(parent T) tea.Cmd
}

func (b *button[T]) Reset(parent T) {
}

func (b *button[T]) Align(a Alignment) Input[T] {
	b.alignment = a
	return b
}

func (b *button[T]) Update(parent T, msg tea.Msg) tea.Cmd {
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		if mt.String() == "space" {
			return b.onClick(parent)
		}
	}
	return nil
}

func (b *button[T]) Render(parent T, form *Form[T], inputNo int, sf Surface, clickPts ClickPoints[T], row, col int, focused bool, style lipgloss.Style, focusedStyle lipgloss.Style) (*tea.Cursor, Surface) {
	if b.alignment == AlignCenter {
		col += (sf.Width() - 2 - len(b.text)) / 2
	}
	if b.big {
		clickPts.Add(sf.Text(row-1, col-1, b.top, style), func(parent T) tea.Cmd {
			form.SetFocusedInput(inputNo)
			return b.onClick(parent)
		})
		if focused {
			clickPts.Add(sf.Text(row, col-1, boxV, style), b.onClick)
			clickPts.Add(sf.Text(row, col, b.text, focusedStyle), b.onClick)
			clickPts.Add(sf.Text(row, col+len(b.text), boxV, style), b.onClick)
		} else {
			clickPts.Add(sf.Text(row, col-1, boxV+b.text+boxV, style), func(parent T) tea.Cmd {
				form.SetFocusedInput(inputNo)
				return b.onClick(parent)
			})
		}
		clickPts.Add(sf.Text(row+1, col-1, b.bottom, style), func(parent T) tea.Cmd {
			form.SetFocusedInput(inputNo)
			return b.onClick(parent)
		})
	} else if focused {
		clickPts.Add(sf.Text(row, col, b.text, focusedStyle), b.onClick)
	} else {
		clickPts.Add(sf.Text(row, col, b.text, style), func(parent T) tea.Cmd {
			form.SetFocusedInput(inputNo)
			return b.onClick(parent)
		})
	}
	return nil, nil
}

func NewDropdownSelect[T any, V []string | func(parent T) []string](values V, maxWidth int, getFn func(parent T) string, setFn func(parent T, value string) tea.Cmd, opts ...bool) Input[T] {
	vs, vsup := getValuesOrSupplier[T](values)
	return &dropdownSelect[T]{
		values:         vs,
		maxWidth:       maxWidth,
		valuesSupplier: vsup,
		getValue:       getFn,
		setValue:       setFn,
		startEllipses:  len(opts) > 0 && opts[0],
	}
}

func getValuesOrSupplier[T any](v any) ([]string, func(parent T) []string) {
	switch vt := v.(type) {
	case []string:
		return vt, nil
	case func(parent T) []string:
		return nil, vt
	}
	panic("should never get here")
}

type dropdownSelect[T any] struct {
	values         []string
	droppedDown    bool
	scrollbar      Scrollbar
	pageSize       int
	maxWidth       int
	valuesSupplier func(parent T) []string
	getValue       func(parent T) string
	setValue       func(parent T, value string) tea.Cmd
	startEllipses  bool
}

func (d *dropdownSelect[T]) Reset(parent T) {
}

func (d *dropdownSelect[T]) Align(a Alignment) Input[T] {
	return d
}

func (d *dropdownSelect[T]) Update(parent T, msg tea.Msg) tea.Cmd {
	if d.scrollbar != nil {
		if d.scrollbar.Update(msg) {
			return nil
		}
	}
	switch mt := msg.(type) {
	case tea.KeyPressMsg:
		return d.key(parent, mt)
	case tea.MouseWheelMsg:
		switch mt.Mouse().Button {
		case tea.MouseWheelDown:
			return d.scroll(parent, ScrollDown)
		case tea.MouseWheelUp:
			return d.scroll(parent, ScrollUp)
		}
	case tea.PasteMsg:
		items := d.getItems(parent)
		if search := strings.ToLower(strings.TrimSpace(mt.Content)); len(search) > 0 {
			idx := slices.IndexFunc(items, func(item string) bool {
				return strings.HasPrefix(strings.ToLower(item), search)
			})
			if idx == -1 {
				idx = slices.IndexFunc(items, func(item string) bool {
					return strings.Contains(strings.ToLower(item), search)
				})
			}
			if idx != -1 {
				return d.setValue(parent, items[idx])
			}
		}
	}
	return nil
}

func (d *dropdownSelect[T]) key(parent T, msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "up":
		return d.scroll(parent, ScrollUp)
	case "down":
		return d.scroll(parent, ScrollDown)
	case "pgup":
		return d.scroll(parent, ScrollPageUp)
	case "pgdown":
		return d.scroll(parent, ScrollPageDown)
	case "home":
		return d.scroll(parent, ScrollHome)
	case "end":
		return d.scroll(parent, ScrollEnd)
	case "enter":
		if d.droppedDown {
			d.droppedDown = false
		} else if len(d.getItems(parent)) > 1 {
			d.droppedDown = true
		}
	default:
		if ch := strings.ToUpper(msg.String()); len(ch) == 1 {
			items := d.getItems(parent)
			current := d.getValue(parent)
			if strings.ToUpper(current[:1]) == ch {
				// current one starts with letter
				idx := slices.Index(items, current)
				if idx+1 < len(items) && strings.ToUpper(items[idx+1][:1]) == ch {
					return d.setValue(parent, items[idx+1])
				}
			}
			for _, item := range items {
				if ch == strings.ToUpper(item[:1]) {
					return d.setValue(parent, item)
				}
			}
		}
	}
	return nil
}

func (d *dropdownSelect[T]) getItems(parent T) []string {
	if d.valuesSupplier != nil {
		return d.valuesSupplier(parent)
	}
	return d.values
}

func (d *dropdownSelect[T]) Render(parent T, form *Form[T], inputNo int, sf Surface, clickPts ClickPoints[T], row, col int, focused bool, style lipgloss.Style, focusedStyle lipgloss.Style) (*tea.Cursor, Surface) {
	d.scrollbar = nil
	maxWd := d.maxWidth
	if maxWd <= 0 {
		maxWd = sf.Width() - col - 1
	}
	current := d.getValue(parent)
	show := current
	if len(show) > maxWd {
		if d.startEllipses {
			show = "..." + show[len(show)-(maxWd-4):]
		} else {
			show = current[:maxWd]
		}
	}
	if focused {
		var ddRgn Surface
		if d.droppedDown {
			items := d.getItems(parent)
			idx := slices.Index(items, current)
			if idx == -1 {
				// current must be a custom value...
				items = append([]string{current}, items...)
				idx = 0
			}
			d.pageSize = sf.Height() - row - 1
			if d.pageSize > len(items) {
				d.pageSize = len(items)
			}
			ddRgn = newSurface(d.pageSize, maxWd+1, sf.AbsoluteTop()+row, sf.AbsoluteLeft()+col)
			ddRgn.Fill(0, 0, ddRgn.Height(), ddRgn.Width(), style)

			topIdx := idx
			if adj := topIdx - len(items) + d.pageSize; adj > 0 {
				topIdx -= adj
			}
			for i := 0; i < d.pageSize && i+topIdx < len(items); i++ {
				on := items[i+topIdx]
				s := style
				if i+topIdx == idx {
					s = focusedStyle
				}
				clickPts.Add(ddRgn.TextFixed(i, 1, ddRgn.Width(), on, s), func(parent T) tea.Cmd {
					d.droppedDown = false
					return d.setValue(parent, on)
				})
			}
			if ddRgn.Height() < len(items) {
				d.pageSize--
				d.scrollbar = NewVerticalScrollbar(func(evt ScrollEvent) tea.Cmd {
					return d.scroll(parent, evt)
				})
				d.scrollbar.Draw(ddRgn, len(items)-1, topIdx)
			}
		} else {
			sf.Fill(row, col, 1, maxWd, focusedStyle)
			clickPts.Add(sf.TextFixed(row, col, maxWd, show, focusedStyle), func(parent T) tea.Cmd {
				d.droppedDown = true
				return nil
			})
			if len(d.getItems(parent)) > 1 {
				clickPts.Add(sf.Text(row, col+maxWd-1, "▼", focusedStyle), func(parent T) tea.Cmd {
					d.droppedDown = true
					return nil
				})
			}
		}
		return nil, ddRgn
	} else {
		d.droppedDown = false
		clickPts.Add(sf.TextFixed(row, col, maxWd, current, style), func(parent T) tea.Cmd {
			form.SetFocusedInput(inputNo)
			return nil
		})
	}
	return nil, nil
}

func (d *dropdownSelect[T]) scroll(parent T, evt ScrollEvent) tea.Cmd {
	switch evt {
	case ScrollUp:
		items := d.getItems(parent)
		if idx := slices.Index(items, d.getValue(parent)); idx > 0 {
			return d.setValue(parent, items[idx-1])
		} else {
			return d.setValue(parent, items[len(items)-1])
		}
	case ScrollDown:
		items := d.getItems(parent)
		if idx := slices.Index(items, d.getValue(parent)); idx < len(items)-1 {
			return d.setValue(parent, items[idx+1])
		} else {
			return d.setValue(parent, items[0])
		}
	case ScrollPageUp:
		items := d.getItems(parent)
		idx := slices.Index(items, d.getValue(parent)) - d.pageSize
		if idx < 0 {
			idx = 0
		}
		return d.setValue(parent, items[idx])
	case ScrollPageDown:
		items := d.getItems(parent)
		idx := slices.Index(items, d.getValue(parent)) + d.pageSize
		if idx >= len(items) {
			idx = len(items) - 1
		}
		return d.setValue(parent, items[idx])
	case ScrollHome:
		items := d.getItems(parent)
		return d.setValue(parent, items[0])
	case ScrollEnd:
		items := d.getItems(parent)
		return d.setValue(parent, items[len(items)-1])
	}
	return nil
}
