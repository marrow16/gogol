package layout

import (
	"charm.land/lipgloss/v2"
)

type region struct {
	textSurface
	parent               Surface
	offsetRow, offsetCol int
}

func newRegion(parent Surface, offsetRow, offsetCol, height, width int) Surface {
	r := &region{
		parent:    parent,
		offsetRow: offsetRow,
		offsetCol: offsetCol,
	}
	r.textSurface = textSurface{
		height: height,
		width:  width,
		placer: r,
	}
	return r
}

func (r *region) place(row int, col int, text string, extent int, styles ...lipgloss.Style) (result Placement) {
	if row >= 0 && row < r.height && col < r.width && extent > 0 {
		style := inheritStyles(styles...)
		seg := &surfaceSegment{
			text:   text,
			extent: extent,
			style:  style,
		}
		if col < 0 {
			crop := -col
			if crop >= seg.extent {
				return
			}
			seg.text = string([]rune(seg.text)[crop:])
			seg.extent -= crop
			col = 0
		}
		if col+seg.extent > r.width {
			seg.extent = r.width - col
			seg.text = string([]rune(seg.text)[:seg.extent])
		}
		if seg.extent <= 0 {
			return
		}
		result = r.parent.place(r.offsetRow+row, r.offsetCol+col, seg.text, seg.extent, styles...)
	}
	return result
}

func (r *region) placeRune(row, col int, pr rune, style lipgloss.Style) {
	if row >= 0 && col >= 0 && row < r.height && col < r.width {
		r.parent.placeRune(r.offsetRow+row, r.offsetCol+col, pr, style)
	}
}

func (r *region) rows_() rows {
	parentRows := r.parent.rows_()[r.offsetRow : r.offsetRow+r.height]
	result := make(rows, len(parentRows))
	for y, pr := range parentRows {
		result[y] = pr[r.offsetCol : r.offsetCol+r.width]
	}
	return result
}

func (r *region) Render() string {
	return r.rows_().render()
}

func (r *region) Region(row, col, height, width int) Surface {
	if row >= 0 && col >= 0 && height > 0 && width > 0 && row < r.height && col < r.width {
		if row+height > r.height {
			height = r.height - row
		}
		if col+width > r.width {
			width = r.width - col
		}
		return newRegion(r, row, col, height, width)
	}
	return nil
}

func (r *region) AbsoluteTop() int {
	return r.parent.AbsoluteTop() + r.offsetRow
}

func (r *region) AbsoluteLeft() int {
	return r.parent.AbsoluteLeft() + r.offsetCol
}

func (r *region) Clear() {
	r.parent.ClearArea(r.offsetRow, r.offsetCol, r.height, r.width)
}

func (r *region) ClearArea(row, col, height, width int) {
	r.parent.ClearArea(r.offsetRow+row, r.offsetCol+col, height, width)
}

func (r *region) Get(row, col int) string {
	if row >= 0 && row < r.height && col < r.width {
		return r.parent.Get(r.offsetRow+row, r.offsetCol+col)
	}
	return ""
}

func (r *region) ClearStyle(row, col, height, width int, style *lipgloss.Style) {
	if row >= 0 && row < r.height && col >= 0 && col < r.width {
		if height < 0 {
			height = r.height
		}
		if height > r.height {
			height = r.height
		}
		if width < 0 {
			width = r.width
		}
		if width > r.width {
			width = r.width
		}
		r.parent.ClearStyle(r.offsetRow+row, r.offsetCol+col, height, width, style)
	}
}

func (r *region) GetStyle(row, col int) *lipgloss.Style {
	if row >= 0 && row < r.height && col >= 0 && col < r.width {
		return r.parent.GetStyle(r.offsetRow+row, r.offsetCol+col)
	}
	return nil
}

func (r *region) SetStyle(row, col int, style *lipgloss.Style) {
	if row >= 0 && row < r.height && col >= 0 && col < r.width {
		r.parent.SetStyle(r.offsetRow+row, r.offsetCol+col, style)
	}
}
