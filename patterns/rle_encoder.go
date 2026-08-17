package patterns

import (
	"errors"
	"github.com/marrow16/gogol/logic"
	"io"
	"strconv"
	"strings"
)

const (
	maxLineLength  = 70
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
	if p.Width > 0 && p.Height > 0 && len(p.Cells) == p.Width*p.Height {
		rw.writeData(p.Width, p.Height, p.Cells)
	} else {
		rw.err = errors.New("invalid pattern cells or pattern size")
	}
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

func (w *rleWriter) writeData(width, height int, cells []bool) {
	var lb strings.Builder
	write := func(s string) {
		if lb.Len()+len(s) > maxLineLength {
			w.writeLine(lb.String())
			lb.Reset()
		}
		lb.WriteString(s)
	}
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
				write(strconv.Itoa(n) + "$")
			} else {
				write("$")
			}
		}
		for _, run := range runs(row) {
			write(run)
		}
		written = true
		pendingRows = 0
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
