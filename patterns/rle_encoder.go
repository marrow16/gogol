package patterns

import (
	"github.com/marrow16/gogol/logic"
	"io"
	"strconv"
	"strings"
)

const (
	maxLineLength  = 71
	tagName        = 'N'
	tagOrigination = 'O'
	tagComment     = 'C'
)

var (
	nl = []byte{'\n'}
)

func PatternRleEncode(p Pattern, w io.Writer) (err error) {
	rw := &rleWriter{w: w}
	rw.writeTag(tagName, p.Name)
	rw.writeTag(tagOrigination, p.Origination)
	for _, line := range p.Comments {
		rw.writeTag(tagComment, line)
	}
	rw.writeDimensions(p.Width, p.Height, p.Rule)
	rw.writeData(p.Width, p.Cells)
	rw.write([]byte{'!'})
	return rw.err
}

type rleWriter struct {
	w   io.Writer
	err error
}

func (w *rleWriter) write(p []byte) {
	if w.err == nil {
		_, w.err = w.w.Write(p)
	}
}

func (w *rleWriter) writeString(s string) {
	w.write([]byte(s))
}

func (w *rleWriter) writeLine(s string) {
	w.write([]byte(s))
	w.write(nl)
}

func (w *rleWriter) writeTag(t byte, s string) {
	if len(s) > 0 {
		w.write([]byte{'#', t, ' '})
		w.writeString(s)
		w.write(nl)
	}
}

func (w *rleWriter) writeDimensions(x, y int, r logic.Rule) {
	w.writeString("x = " + strconv.Itoa(x) + ", y = " + strconv.Itoa(y))
	if r != nil {
		w.writeString(", rule = " + r.Rle())
	}
	w.write(nl)
}

func (w *rleWriter) writeData(width int, cells []bool) {
	ht := len(cells) / width
	var lb strings.Builder
	checkLineLen := func() {
		if lb.Len()+1 >= maxLineLength {
			w.writeLine(lb.String())
			lb.Reset()
		}
	}
	for r := 0; r < ht; r++ {
		if r > 0 {
			lb.WriteRune('$')
		}
		row := cells[r*width : (r+1)*width]

		for len(row) > 0 && !row[len(row)-1] {
			row = row[:len(row)-1]
		}
		if len(row) == 0 {
			row = make([]bool, width)
		}
		for _, run := range runs(row) {
			checkLineLen()
			lb.WriteString(run)
		}
	}
	w.writeString(lb.String())
}

func runs(cells []bool) []string {
	result := make([]string, 0, len(cells))
	for c := 0; c < len(cells); c++ {
		ob := cells[c]
		n := 1
		for i := c + 1; i < len(cells); i++ {
			if cells[i] == ob {
				n++
				c++
			} else {
				break
			}
		}
		switch {
		case n == 1 && ob:
			result = append(result, "o")
		case n == 1:
			result = append(result, "b")
		case ob:
			result = append(result, strconv.Itoa(n)+"o")
		default:
			result = append(result, strconv.Itoa(n)+"b")
		}
	}
	return result
}
