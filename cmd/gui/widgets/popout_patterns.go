package widgets

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"image"
	"image/draw"
	"slices"
	"strconv"
	"strings"
)

type patternsPopout struct {
	parent               *menuPopup
	core                 *Core
	chooser              *chooser[patterns.Pattern]
	previewMode          *widget.Enum
	radioPreview         *radioButton
	radioMetadata        *radioButton
	chkFilterCurrentRule *checkbox
	currentRule          logic.Rule
	btnPlace             *button
	chkInterlaced        *checkbox
}

const (
	previewImage    = "image"
	previewMetadata = "metadata"
)

func newPatternsPopout(p *menuPopup, c *Core) *patternsPopout {
	result := &patternsPopout{
		parent:               p,
		core:                 c,
		previewMode:          &widget.Enum{Value: previewImage},
		chkFilterCurrentRule: newCheckBox("Filter current rule", false),
		btnPlace:             newButton("Place"),
		chkInterlaced:        newCheckBox("Interlaced", false),
	}
	result.radioPreview = newRadioButton(result.previewMode, previewImage, "Preview")
	result.radioMetadata = newRadioButton(result.previewMode, previewMetadata, "Metadata")
	result.chooser = newChooser[patterns.Pattern](38,
		result.sortedPatterns(),
		result.patternSelected,
		func(pattern patterns.Pattern) string {
			return pattern.String()
		},
	)
	return result
}

func (p *patternsPopout) patternSelected(pattern *patterns.Pattern) {
	//fmt.Printf("Pattern selected: %+v\n", pattern)
}

func (p *patternsPopout) setSelected(name string) {
	p.chkFilterCurrentRule.SetChecked(false)
	p.chooser.opened = false
	p.chooser.resetItems(p.sortedPatterns())
	p.chooser.editor.SetText(name)
}

func (p *patternsPopout) sortedPatterns() []patterns.Pattern {
	result := make([]patterns.Pattern, 0, len(patterns.PatternLibrary))
	if p.chkFilterCurrentRule.Checked() {
		p.currentRule = p.core.gridHolder.grid.Rule
		perm := p.currentRule.Permutation()
		for _, pattern := range patterns.PatternLibrary {
			if pattern.Rule != nil && pattern.Rule.Permutation() == perm {
				result = append(result, pattern)
			}
		}
	} else {
		p.currentRule = nil
		for _, pattern := range patterns.PatternLibrary {
			result = append(result, pattern)
		}
	}
	slices.SortStableFunc(result, func(a, b patterns.Pattern) int {
		return strings.Compare(strings.ToLower(a.String()), strings.ToLower(b.String()))
	})
	return result
}

func (p *patternsPopout) currentPattern() (patt *patterns.Pattern, interlaced bool) {
	return p.chooser.currentItem(), p.chkInterlaced.Checked()
}

func (p *patternsPopout) layout(gtx layout.Context) layout.Dimensions {
	if p.btnPlace.Clicked(gtx) {
		if pattern := p.chooser.currentItem(); pattern != nil {
			p.core.startPatternPlace(gtx, pattern, p.chkInterlaced.Checked())
		}
	}
	if p.chkFilterCurrentRule.Checked() && p.currentRule != nil && p.currentRule.Permutation() != p.core.gridHolder.grid.Rule.Permutation() {
		p.chooser.resetItems(p.sortedPatterns())
	}
	if ok := p.chkFilterCurrentRule.Update(gtx); ok {
		p.chooser.resetItems(p.sortedPatterns())
	}
	chd := measureText(gtx, "M")
	gtx.Constraints.Min.Y = chd.Size.Y * 20
	return layout.Inset{
		Left: 8, Right: 8, Top: 8, Bottom: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		var chooserDims layout.Dimensions
		dims := layout.Flex{Axis: layout.Vertical, Gap: 10}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				chooserDims = p.chooser.layout(gtx)
				return chooserDims
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Gap: 10}.Layout(gtx,
					layout.Rigid(p.radioPreview.Layout),
					layout.Rigid(p.radioMetadata.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{
							Size:     image.Point{X: gtx.Dp(100)},
							Baseline: 0,
						}
					}),
					layout.Rigid(p.chkFilterCurrentRule.Layout),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.Y = int(float32(chd.Size.Y) * 15.5)
				gtx.Constraints.Max.X = p.chooser.dims.Size.X
				return p.layoutPreview(gtx, p.chooser.dims.Size.X, chd.Size.Y*15)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if p.chooser.currentItem() != nil {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(p.btnPlace.Layout),
						layout.Rigid(p.chkInterlaced.Layout),
					)
				} else {
					return layout.Dimensions{}
				}
			}),
		)
		p.chooser.layoutDropdown(gtx, chooserDims)
		return dims
	})
}

func (p *patternsPopout) layoutPreview(gtx layout.Context, maxWd, maxHt int) layout.Dimensions {
	currentPattern := p.chooser.currentItem()
	switch {
	case currentPattern == nil:
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceAround}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				lbl := material.Label(theme, theme.TextSize, "(select a pattern)")
				lbl.MaxLines = 1
				lbl.Alignment = text.Middle
				return lbl.Layout(gtx)
			}),
		)
	case p.previewMode.Value == previewMetadata:
		return p.layoutPreviewMetadata(*currentPattern, gtx)
	default:
		return p.layoutPreviewImage(*currentPattern, gtx, maxWd, maxHt)
	}
}

func (p *patternsPopout) layoutPreviewMetadata(pattern patterns.Pattern, gtx layout.Context) layout.Dimensions {
	labelMax := measureMaxText(gtx, font.Bold, "Size: ", "Filename: ", "Origin: ", "Comment: ").Size.X
	return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceEnd}.Layout(gtx,
		layout.Rigid(flexHorizontal(20,
			rigidLabel("Size:", text.End, font.Bold, labelMax),
			layout.Flexed(1, label(strconv.Itoa(pattern.Width)+"w X "+strconv.Itoa(pattern.Height)+"h")),
		)),
		layout.Rigid(flexHorizontal(20,
			rigidLabel("Rule:", text.End, font.Bold, labelMax),
			layout.Flexed(1, label(pattern.Rule.Name())),
		)),
		layout.Rigid(flexHorizontal(20,
			rigidLabel("Filename:", text.End, font.Bold, labelMax),
			layout.Flexed(1, label(pattern.Filename)),
		)),
		layout.Rigid(flexHorizontal(20,
			rigidLabel("Origin:", text.End, font.Bold, labelMax),
			layout.Flexed(1, label(pattern.Origination)),
		)),
		layout.Rigid(flexHorizontal(20,
			rigidLabel("Comment:", text.End, font.Bold, labelMax),
			layout.Flexed(1, material.Label(theme, theme.TextSize, strings.Join(pattern.Comments, "\n")).Layout),
		)),
	)
}

func (p *patternsPopout) layoutPreviewImage(pattern patterns.Pattern, gtx layout.Context, maxWd, maxHt int) layout.Dimensions {
	cellSize := min(maxWd/pattern.Width, maxHt/pattern.Height)
	rect := image.Rect(0, 0, cellSize*pattern.Width, cellSize*pattern.Height)
	canvas := image.NewNRGBA(rect)
	draw.Draw(canvas, rect, &image.Uniform{p.core.settings.CellDeadColor}, image.Point{}, draw.Src)
	offset := 0
	if cellSize > 3 {
		offset = 1
		for y := 0; y <= pattern.Height; y++ {
			yy := y * cellSize
			draw.Draw(
				canvas,
				image.Rect(0, yy, pattern.Width*cellSize, yy+1),
				&image.Uniform{p.core.settings.CellBorderColor},
				image.Point{},
				draw.Src,
			)
		}
		for x := 0; x <= pattern.Width; x++ {
			xx := x * cellSize
			draw.Draw(
				canvas,
				image.Rect(xx, 0, xx+1, pattern.Height*cellSize),
				&image.Uniform{p.core.settings.CellBorderColor},
				image.Point{},
				draw.Src,
			)
		}
	}
	pattern.DrawTo(patterns.Rotate0, func(row, col int, alive bool) {
		if alive {
			draw.Draw(canvas, image.Rect(
				(col*cellSize)+offset,
				(row*cellSize)+offset,
				(col+1)*cellSize,
				(row+1)*cellSize),
				&image.Uniform{p.core.settings.CellAliveColor}, image.Point{}, draw.Src)
		}
	})
	return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceEnd}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := canvas.Bounds().Size()
			stack := clip.Rect{Max: size}.Push(gtx.Ops)
			defer stack.Pop()
			paint.NewImageOp(canvas).Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			return layout.Dimensions{Size: size}
		}),
	)
}

func (p *patternsPopout) reset() {
	p.chooser.opened = false
	p.chooser.resetItems(p.sortedPatterns())
}

func (p *patternsPopout) hasFocus(gtx layout.Context) bool {
	_, radios := p.previewMode.Focused()
	return radios || p.chooser.isFocused(gtx) || p.chkFilterCurrentRule.isFocused(gtx) || p.chkInterlaced.isFocused(gtx) ||
		p.btnPlace.isFocused(gtx)
}
