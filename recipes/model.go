package recipes

import (
	"encoding/json"
	"fmt"
	"github.com/marrow16/gogol/patterns"
	"slices"
	"strconv"
	"strings"
)

type Recipe struct {
	Name         *string
	GridSettings *GridSettings `json:"grid,omitempty"`
	Patterns     Patterns      `json:"patterns"`
	Vars         Vars          `json:"vars"`
	Do           []Do          `json:"do"`
}

type GridSettings struct {
	Width        *int    `json:"width"`
	Height       *int    `json:"height"`
	Rule         *string `json:"rule"`
	WrapMode     *string `json:"wrap_mode"`
	BoundaryMode *string `json:"boundary_mode"`
}

type Vars map[string]*Number

type Patterns map[string]*Pattern

type Pattern struct {
	Name     *string `json:"name"`
	Filename *string `json:"filename"`
	Width    *int    `json:"width"`
	Height   *int    `json:"height"`
	Rle      *string `json:"rle"`
	pattern  *patterns.Pattern
}

func (p *Pattern) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		p.Name = &name
		return nil
	}
	var raw struct {
		Filename *string `json:"filename"`
		Width    *int    `json:"width"`
		Height   *int    `json:"height"`
		Rle      *string `json:"rle"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch {
	case raw.Filename != nil:
		p.Filename = raw.Filename
		return nil
	case raw.Rle != nil:
		if raw.Width == nil {
			return fmt.Errorf("expected 'width' property with 'rle'")
		} else if *raw.Width < 1 {
			return fmt.Errorf("expected 'width' property to be greater than zero")
		}
		if raw.Height == nil {
			return fmt.Errorf("expected 'height' property with 'rle'")
		} else if *raw.Height < 1 {
			return fmt.Errorf("expected 'height' property to be greater than zero")
		}
		p.Width = raw.Width
		p.Height = raw.Height
		p.Rle = raw.Rle
		return nil
	default:
		return fmt.Errorf("expected string pattern name, or object with 'filename' or 'rle' property")
	}
}

type Do struct {
	Place  *string   `json:"place"`
	At     *Position `json:"at"`
	Move   *Move     `json:"move"`
	Repeat *Number   `json:"repeat"`
	Rotate *int      `json:"rotate"`
	Do     []Do      `json:"do"`
	VarOps Ops       `json:"var_operations"`
}

type Number string

func (n *Number) UnmarshalJSON(data []byte) error {
	*n = Number(data)
	return nil
}

func (n *Number) raw() string {
	if n == nil {
		return ""
	}
	if strings.HasPrefix(string(*n), "\"") && strings.HasPrefix(string(*n), "\"") {
		s := string(*n)
		return s[1 : len(s)-1]
	}
	return string(*n)
}

func (n *Number) int() int {
	if nn, err := strconv.Atoi(string(*n)); err == nil {
		return nn
	} else {
		s := strings.TrimSuffix(strings.TrimPrefix(string(*n), "\""), "\"")
		if nn, err := strconv.ParseInt(s, 0, 64); err == nil {
			return int(nn)
		}
	}
	return 0
}

func (n *Number) stripPrefixes() (value string, prefixes map[string]bool) {
	value = n.raw()
	prefixes = make(map[string]bool)
	for {
		if cAt := strings.Index(value, ":"); cAt != -1 {
			prefixes[strings.ToLower(value[:cAt])] = true
			value = value[cAt+1:]
		} else {
			break
		}
	}
	return
}

func (n *Number) toPattern(gridHeight, gridWidth int, y, x int) patterns.Pattern {
	v, prefixes := n.stripPrefixes()
	if nn, err := strconv.Atoi(v); err == nil && len(prefixes) == 0 {
		if nn < 2 {
			return patterns.Pattern{Height: 1, Width: 1, Cells: []bool{nn == 1}}
		} else {
			cells := make([]bool, 0, 16)
			for nn > 0 {
				cells = append(cells, nn%2 == 1)
				nn = nn >> 1
			}
			slices.Reverse(cells)
			return patterns.Pattern{Height: 1, Width: len(cells), Cells: cells}
		}
	} else if nn, err := strconv.ParseInt(v, 0, 64); err == nil {
		wd := len(v)
		switch {
		case strings.HasPrefix(v, "0b"):
			wd -= 2
		case strings.HasPrefix(v, "0x"):
			wd -= 2
			wd *= 4
		case strings.HasPrefix(v, "0o"):
			wd -= 2
			wd *= 3
		default:
			wd = len(fmt.Sprintf("%b", nn))
		}
		cells := make([]bool, 0, wd)
		for i := wd; i > 0; i-- {
			cells = append(cells, nn%2 == 1)
			nn = nn >> 1
		}
		slices.Reverse(cells)
		if prefixes["alt"] || prefixes["alternate"] {
			for i := 0; i < len(cells); i++ {
				cells[i] = !cells[i]
			}
		}
		if prefixes["rot"] || prefixes["rotate"] {
			l := len(cells)
			if rotate := y % l; rotate != 0 {
				tmp := append([]bool(nil), cells[:rotate]...)
				copy(cells, cells[rotate:])
				copy(cells[l-rotate:], tmp)
			}
		}
		if prefixes["fw"] || prefixes["fillwidth"] || prefixes["fill-width"] {
			nw := ((gridWidth - x) / wd) + 1
			newCells := make([]bool, 0, nw*wd)
			for i := 0; i < nw; i++ {
				newCells = append(newCells, cells...)
			}
			wd = len(newCells)
			cells = newCells
		}
		ht := 1
		if prefixes["fh"] || prefixes["fillheight"] || prefixes["fill-height"] {
			ht = (gridHeight - y) + 2
			newCells := make([]bool, 0, ht*wd)
			for i := 0; i < ht; i++ {
				newCells = append(newCells, cells...)
			}
			cells = newCells
		}
		return patterns.Pattern{Height: ht, Width: wd, Cells: cells}
	}
	return patterns.Pattern{Height: 1, Width: 1, Cells: []bool{false}}
}

type Ops []Op

func (o Ops) performOnVar(num *Number, after bool) {
	for _, op := range o {
		if (after && (op.When == nil || *op.When == "after")) || (!after && op.When != nil && *op.When == "before") {
			op.performOnVar(num)
		}
	}
}

type Op struct {
	When        *string `json:"when"` // defaults to after
	ShiftLeft   *int    `json:"shift_left"`
	ShiftRight  *int    `json:"shift_right"`
	RotateLeft  *int    `json:"rotate-left"`
	RotateRight *int    `json:"rotate-right"`
	Increment   *int    `json:"increment"`
	Decrement   *int    `json:"decrement"`
	// todo more?
}

func (o Op) performOnVar(num *Number) {
	v, prefixes := num.stripPrefixes()
	if nn, err := strconv.ParseInt(v, 0, 64); err == nil {
		_ = prefixes
		wd := len(v)
		switch {
		case strings.HasPrefix(v, "0b"):
			wd -= 2
		case strings.HasPrefix(v, "0x"):
			wd -= 2
			wd *= 4
		case strings.HasPrefix(v, "0o"):
			wd -= 2
			wd *= 3
		default:
			wd = len(fmt.Sprintf("%b", nn))
		}
		if o.ShiftLeft != nil && *o.ShiftLeft > 0 {
			nn = nn << uint(*o.ShiftLeft)
		}
		if o.ShiftRight != nil && *o.ShiftRight > 0 {
			nn = nn >> uint(*o.ShiftRight)
		}
		if o.Increment != nil && *o.Increment != 0 {
			nn += int64(*o.Increment)
		}
		if o.Decrement != nil && *o.Decrement != 0 {
			nn -= int64(*o.Decrement)
		}
		if wd > 0 {
			mask := int64((uint64(1) << uint(wd)) - 1)
			if o.RotateLeft != nil && *o.RotateLeft > 0 {
				n := *o.RotateLeft % wd
				if n < 0 {
					n += wd
				}
				nn = ((nn << uint(n)) | (nn >> uint(wd-n))) & mask
			}
			if o.RotateRight != nil && *o.RotateRight > 0 {
				n := *o.RotateRight % wd
				if n < 0 {
					n += wd
				}
				nn = ((nn >> uint(n)) | (nn << uint(wd-n))) & mask
			}
		}
		nv := fmt.Sprintf("%0"+strconv.Itoa(wd)+"b", nn)
		if l := len(nv); l > wd {
			nv = nv[l-wd:]
		}
		var sb strings.Builder
		for pfx := range prefixes {
			sb.WriteString(pfx + ":")
		}
		sb.WriteString("0b" + nv)
		newNum := Number(sb.String())
		*num = newNum
	}
}

type Move struct {
	X    *Number `json:"x"`
	Y    *Number `json:"y"`
	When *string `json:"when"`
}

func (m *Move) isAfter() bool {
	return m.When != nil && strings.EqualFold(*m.When, "after")
}

func (m *Move) move(cy, cx int, lh, lw int) (int, int) {
	if m.Y != nil {
		if n, ok := resolveOffset(m.Y.raw(), lh, lw); ok {
			cy += n
		} else {
			cy += m.Y.int()
		}
	}
	if m.X != nil {
		if n, ok := resolveOffset(m.X.raw(), lh, lw); ok {
			cx += n
		} else {
			cx += m.X.int()
		}
	}
	return cy, cx
}

func resolveOffset(s string, lh, lw int) (int, bool) {
	if after, ok := strings.CutPrefix(s, "lh"); ok {
		return afterAdj(lh, after), true
	} else if after, ok = strings.CutPrefix(s, "lw"); ok {
		return afterAdj(lw, after), true
	} else if after, ok = strings.CutPrefix(s, "-lh"); ok {
		return afterAdj(-lh, after), true
	} else if after, ok = strings.CutPrefix(s, "-lw"); ok {
		return afterAdj(-lw, after), true
	}
	return 0, false
}

func afterAdj(n int, after string) int {
	switch {
	case strings.HasPrefix(after, "++"):
		n++
	case strings.HasPrefix(after, "--"):
		n--
	case strings.HasPrefix(after, "+") && len(after) > 1:
		if na, err := strconv.Atoi(after[1:]); err == nil {
			n += na
		}
	case strings.HasPrefix(after, "-") && len(after) > 1:
		if na, err := strconv.Atoi(after[1:]); err == nil {
			n -= na
		}
	}
	return n
}

type Position struct {
	X *Number `json:"x"`
	Y *Number `json:"y"`
}

func (p *Position) move(cy, cx int) (int, int) {
	if p.Y != nil {
		cy = p.Y.int()
	}
	if p.X != nil {
		cx = p.X.int()
	}
	return cy, cx
}
