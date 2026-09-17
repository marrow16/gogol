package help

import (
	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"image"
	"strings"
)

type table struct {
	borders bool
	indent  unit.Dp
	padding unit.Dp
	columns []tableColumn
	rows    []tableRow
	spacing
}

type tableColumn struct {
	width  int //percent
	header string
	align  text.Alignment
}

type tableRow []content

func (t table) String() string {
	var sb strings.Builder
	for _, col := range t.columns {
		sb.WriteString(col.header)
		sb.WriteString(" ")
	}
	for _, row := range t.rows {
		for _, col := range row {
			sb.WriteString(col.String())
			sb.WriteString(" ")
		}
	}
	return sb.String()
}

func (t table) widget(h *holder) layout.Widget {
	numCols := len(t.columns)
	for _, row := range t.rows {
		if l := len(row); l > numCols {
			numCols = l
		}
	}
	showHeaders := false
	columnWidths := make([]float32, numCols)
	availableWd := float32(1)
	usedCols := 0
	for i, col := range t.columns {
		if col.header != "" {
			showHeaders = true
		}
		if col.width > 0 {
			usedCols++
			wd := float32(col.width) / 100.0
			availableWd -= wd
			columnWidths[i] = wd
		}
	}
	if availableWd < 0 || availableWd == 1.0 {
		aw := float32(1)
		c := float32(numCols)
		for i := range columnWidths {
			w := aw / c
			columnWidths[i] = w
			aw -= w
			c--
		}
	} else if availableWd > 0 && usedCols < numCols {
		aw := availableWd
		c := float32(numCols - usedCols)
		for i := range columnWidths {
			if columnWidths[i] == 0 {
				w := aw / c
				columnWidths[i] = w
				aw -= w
				c--
			}
		}
	}
	padding := t.padding
	if padding < 0 {
		padding = 0
	} else if padding == 0 {
		padding = 4
	}
	rows := t.buildRows(h, showHeaders, columnWidths, padding)
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: t.indent, Top: t.spaceBefore, Bottom: t.spaceAfter}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
			if t.borders {
				tableBorders(gtx, columnWidths, dims)
			}
			return dims
		})
	}
}

func (t table) buildRows(h *holder, showHeaders bool, columnWidths []float32, padding unit.Dp) []layout.FlexChild {
	rows := make([]layout.FlexChild, 0, len(t.rows)+1)
	if showHeaders {
		rows = append(rows, t.buildHeaderRow(h, columnWidths, padding))
	}
	for _, row := range t.rows {
		rows = append(rows, t.buildRow(row, h, columnWidths, padding))
	}
	return rows
}

func (t table) buildRow(row tableRow, h *holder, columnWidths []float32, padding unit.Dp) layout.FlexChild {
	cellWidgets := make([]layout.Widget, len(columnWidths))
	for col := range columnWidths {
		align := text.Start
		if col < len(t.columns) {
			align = t.columns[col].align
		}
		if col < len(row) {
			cellWidgets[col] = buildContent(h, align, row[col])
		} else {
			cellWidgets[col] = buildContent(h, align, nil)
		}
	}
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		width := float32(gtx.Constraints.Max.X)
		cols := make([]layout.FlexChild, 0, len(columnWidths))
		for i, w := range columnWidths {
			col := i
			colWd := int(width * w)
			cols = append(cols, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Max.X, gtx.Constraints.Min.X = colWd, colWd
				return layout.Inset{Top: padding, Bottom: padding, Left: padding, Right: padding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return cellWidgets[col](gtx)
				})
			}))
		}
		dims := layout.Flex{Axis: layout.Horizontal}.Layout(gtx, cols...)
		if !t.borders {
			return dims
		}
		// top border for the row
		{
			var p clip.Path
			p.Begin(gtx.Ops)
			p.MoveTo(f32.Pt(0, 0))
			p.LineTo(f32.Pt(float32(dims.Size.X), 0))
			stack := clip.Stroke{
				Path:  p.End(),
				Width: 1.0,
			}.Op().Push(gtx.Ops)
			paint.ColorOp{Color: lineColor}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			stack.Pop()
		}
		return dims
	})
}

func (t table) buildHeaderRow(h *holder, columnWidths []float32, padding unit.Dp) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		width := float32(gtx.Constraints.Max.X)
		cols := make([]layout.FlexChild, 0, len(columnWidths))
		for i, w := range columnWidths {
			col := i
			colWd := int(width * w)
			hdr := ""
			align := text.Start
			if col < len(t.columns) {
				hdr = t.columns[col].header
				align = t.columns[col].align
			}
			cols = append(cols, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Max.X, gtx.Constraints.Min.X = colWd, colWd
				dims := layout.Inset{Top: padding, Bottom: padding, Left: padding, Right: padding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(h.theme, h.theme.TextSize, hdr)
					lbl.MaxLines = 1
					lbl.Font.Weight = font.Bold
					lbl.Alignment = align
					return lbl.Layout(gtx)
				})
				return dims
			}))
		}
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, cols...)
	})
}

func tableBorders(gtx layout.Context, columnWidths []float32, dims layout.Dimensions) {
	paint.ColorOp{Color: lineColor}.Add(gtx.Ops)
	// outer border
	{
		stack := clip.Stroke{
			Path: clip.Rect{
				Min: image.Point{},
				Max: dims.Size,
			}.Path(),
			Width: 1.0,
		}.Op().Push(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		stack.Pop()
	}
	// column separators
	x := float32(0)
	for i := 0; i < len(columnWidths)-1; i++ {
		x += columnWidths[i] * float32(dims.Size.X)

		var p clip.Path
		p.Begin(gtx.Ops)
		p.MoveTo(f32.Pt(x, 0))
		p.LineTo(f32.Pt(x, float32(dims.Size.Y)))

		stack := clip.Stroke{
			Path:  p.End(),
			Width: 1.0,
		}.Op().Push(gtx.Ops)

		paint.PaintOp{}.Add(gtx.Ops)
		stack.Pop()
	}
}
