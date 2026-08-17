package patterns

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/marrow16/gogol/logic"
	"strconv"
	"strings"
)

type CellsEncoding int

const (
	RleCellsEncoding CellsEncoding = iota
	Base64CellsEncoding
)

type JsonPattern struct {
	Name        string    `json:"name,omitempty"`
	Width       uint      `json:"width"`
	Height      uint      `json:"height"`
	Origination string    `json:"origination,omitempty"`
	Rule        string    `json:"rule,omitempty"`
	Comments    []string  `json:"comments,omitempty"`
	Cells       JsonCells `json:"cells"`
}

type JsonCells struct {
	Rle    *string `json:"rle,omitempty"`
	Base64 []byte  `json:"base64,omitempty"`
}

var _ json.Unmarshaler = (*Pattern)(nil)
var _ json.Marshaler = (*Pattern)(nil)

func (p Pattern) MarshalJSON() ([]byte, error) {
	cells := JsonCells{}
	switch p.CellsEncoding {
	case Base64CellsEncoding:
		cells.Base64 = PackCells(p.Cells)
	default:
		cs, err := RleCells(p.Width, p.Width, p.Cells)
		if err != nil {
			return nil, err
		}
		cells.Rle = &cs
	}
	r := ""
	if p.Rule != nil {
		r = p.Rule.Rle()
	}
	jp := JsonPattern{
		Name:        p.Name,
		Width:       uint(p.Width),
		Height:      uint(p.Height),
		Origination: p.Origination,
		Rule:        r,
		Comments:    p.Comments,
		Cells:       cells,
	}
	return json.Marshal(jp)
}

func (p *Pattern) UnmarshalJSON(data []byte) error {
	var jp JsonPattern
	if err := json.Unmarshal(data, &jp); err != nil {
		return err
	}
	rp := Pattern{
		Name:        jp.Name,
		Width:       int(jp.Width),
		Height:      int(jp.Height),
		Origination: jp.Origination,
		Comments:    jp.Comments,
	}
	if jp.Rule != "" {
		r, err := logic.NewRuleRle("", jp.Rule)
		if err != nil {
			return err
		}
		rp.Rule = r
	}
	switch {
	case jp.Cells.Base64 != nil:
		cells, err := UnpackCells(jp.Cells.Base64, jp.Width, jp.Height)
		if err != nil {
			return err
		}
		rp.Cells = cells
	case jp.Cells.Rle != nil:
		cells, err := ParseRLEData(rp.Width, rp.Height, *jp.Cells.Rle)
		if err != nil {
			return err
		}
		rp.Cells = cells
	default:
		return errors.New("invalid pattern - no cells defined")
	}
	*p = rp
	return nil
}

func RleCells(width, height int, cells []bool) (string, error) {
	if width <= 0 || height <= 0 {
		return "", errors.New("invalid pattern cells - no cells defined")
	} else if len(cells) != width*height {
		return "", errors.New("invalid pattern cells - incorrect cells size")
	}
	var lb strings.Builder
	pendingRows := 0
	written := false
	for r := 0; r < height; r++ {
		row := cells[r*width : (r+1)*width]
		// trim trailing dead cells...
		for len(row) > 0 && !row[len(row)-1] {
			row = row[:len(row)-1]
		}
		if len(row) == 0 {
			pendingRows++
			continue
		}
		if pendingRows > 0 || written {
			n := pendingRows
			if written {
				n++
			}
			if n > 1 {
				lb.WriteString(strconv.Itoa(n))
			}
			lb.WriteByte('$')
		}
		for _, run := range runs(row) {
			lb.WriteString(run)
		}
		written = true
		pendingRows = 0
	}
	return lb.String(), nil
}

func PackCells(cells []bool) []byte {
	out := make([]byte, (len(cells)+7)/8)
	for i, alive := range cells {
		if alive {
			out[i>>3] |= 1 << (7 - (i & 7))
		}
	}
	return out
}

func UnpackCells(data []byte, width, height uint) ([]bool, error) {
	count := width * height
	required := (count + 7) / 8
	if len(data) != int(required) {
		return nil, fmt.Errorf(
			"invalid cell data length: got %d bytes, expected %d",
			len(data), required,
		)
	}
	if rem := count & 7; rem != 0 && len(data) > 0 {
		mask := byte((1 << (8 - rem)) - 1)
		if data[len(data)-1]&mask != 0 {
			return nil, errors.New("non-zero padding bits")
		}
	}
	cells := make([]bool, count)
	for i := range cells {
		cells[i] = data[i>>3]&(1<<(7-(i&7))) != 0
	}
	return cells, nil
}
