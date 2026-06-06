package layout

import (
	"charm.land/lipgloss/v2"
)

type surfaceInternal interface {
	place(row int, col int, text string, extent int, styles ...lipgloss.Style) Placement
	rows_() rows
}

type Surface interface {
	surfaceInternal
	surfaceText
	Render() string
	Region(row, col, height, width int) Surface
	AbsoluteTop() int
	AbsoluteLeft() int
	Clear()
	ClearArea(row, col, height, width int)
	Get(row, col int) string
	ClearStyle(row, col, height, width int, style *lipgloss.Style)
	GetStyle(row, col int) *lipgloss.Style
	SetStyle(row, col int, style *lipgloss.Style)
}

type surface struct {
	textSurface
	rows                 rows
	offsetRow, offsetCol int
}

func NewSurface(height, width int) Surface {
	return newSurface(height, width, 0, 0)
}

func newSurface(height, width int, offsetRow, offsetCol int) Surface {
	sf := &surface{
		rows:      newRows(height, width),
		offsetRow: offsetRow,
		offsetCol: offsetCol,
	}
	sf.textSurface = textSurface{
		height: height,
		width:  width,
		placer: sf,
	}
	return sf
}

func (s *surface) Render() string {
	return s.rows.render()
}

func (s *surface) rows_() rows {
	return s.rows
}

func (s *surface) place(row, col int, text string, extent int, styles ...lipgloss.Style) (result Placement) {
	if row >= 0 && row < s.height && col < s.width && extent > 0 {
		style := inheritStyles(styles...)
		seg := &surfaceSegment{
			text:   text,
			extent: extent,
			style:  style,
		}
		s.rows[row].place(col, seg)
		result = Placement{
			Text:   text,
			Extent: extent,
			Row:    row + s.offsetRow,
			Col:    col + s.offsetCol,
		}
		result.Text = text
		result.Extent = extent
	}
	return result
}

func (s *surface) Region(row, col, height, width int) Surface {
	if row >= 0 && col >= 0 && height > 0 && width > 0 && row < s.height && col < s.width {
		if row+height > s.height {
			height = s.height - row
		}
		if col+width > s.width {
			width = s.width - col
		}
		return newRegion(s, row, col, height, width)
	}
	return nil
}

func (s *surface) AbsoluteTop() int {
	return s.offsetRow
}

func (s *surface) AbsoluteLeft() int {
	return s.offsetCol
}

func (s *surface) Clear() {
	s.rows = newRows(s.height, s.width)
}

func (s *surface) ClearArea(row, col, height, width int) {
	s.rows.clear(row, col, height, width)
}

func (s *surface) Get(row, col int) string {
	if row >= 0 && row < s.height && col < s.width {
		if seg := s.rows[row][col]; seg != nil {
			return seg.text
		}
	}
	return ""
}

func (s *surface) ClearStyle(row, col, height, width int, style *lipgloss.Style) {
	if row >= 0 && row < s.height && col >= 0 && col < s.width {
		if height < 0 {
			height = s.height
		}
		if width < 0 {
			width = s.width
		}
		for y := 0; y < height && y+row < s.height; y++ {
			for x := 0; x < width && x+col < s.width; x++ {
				if seg := s.rows[y][x]; seg != nil {
					seg.style = style
				}
			}
		}
	}
}

func (s *surface) GetStyle(row, col int) *lipgloss.Style {
	if row >= 0 && row < s.height && col >= 0 && col < s.width {
		if seg := s.rows[row][col]; seg != nil {
			return seg.style
		}
	}
	return nil
}

func (s *surface) SetStyle(row, col int, style *lipgloss.Style) {
	if row >= 0 && row < s.height && col >= 0 && col < s.width {
		if seg := s.rows[row][col]; seg != nil {
			seg.style = style
		}
	}
}

func styleValue(style *lipgloss.Style) []lipgloss.Style {
	if style == nil {
		return nil
	}
	return []lipgloss.Style{*style}
}

func inheritStyles(styles ...lipgloss.Style) (result *lipgloss.Style) {
	if len(styles) > 0 {
		st := styles[0]
		for i := 1; i < len(styles); i++ {
			st = st.Inherit(styles[i])
		}
		result = &st
	}
	return result
}
