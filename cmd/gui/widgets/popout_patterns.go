package widgets

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
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
		chkFilterCurrentRule: newCheckBox(c.theme, "Filter current rule", false),
		btnPlace:             newButton(c.theme, "Place"),
		chkInterlaced:        newCheckBox(c.theme, "Interlaced", false),
	}
	result.radioPreview = newRadioButton(c.theme, result.previewMode, previewImage, "Preview")
	result.radioMetadata = newRadioButton(c.theme, result.previewMode, previewMetadata, "Metadata")
	result.chooser = newChooser[patterns.Pattern](c.theme, 38,
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

func (p *patternsPopout) sortedPatterns() []patterns.Pattern {
	result := make([]patterns.Pattern, 0, len(patterns.PatternLibrary))
	if p.chkFilterCurrentRule.Checked() {
		p.currentRule = p.core.gridHolder.grid.Rule
		perm := p.currentRule.Permutation()
		for _, pattern := range patterns.PatternLibrary {
			if pattern.Rule.Permutation() == perm {
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

func (p *patternsPopout) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
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
	chd := measureText(gtx, p.core.theme, "M")
	gtx.Constraints.Min.Y = chd.Size.Y * 20
	return layout.Inset{
		Left: unit.Dp(8), Right: unit.Dp(8),
		Top: unit.Dp(8), Bottom: unit.Dp(4),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
				return p.layoutPreview(gtx, theme, p.chooser.dims.Size.X, chd.Size.Y*15)
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

func (p *patternsPopout) layoutPreview(gtx layout.Context, theme *material.Theme, maxWd, maxHt int) layout.Dimensions {
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
		return p.layoutPreviewMetadata(*currentPattern, gtx, theme)
	default:
		return p.layoutPreviewImage(*currentPattern, gtx, theme, maxWd, maxHt)
	}
}

func (p *patternsPopout) layoutPreviewMetadata(pattern patterns.Pattern, gtx layout.Context, theme *material.Theme) layout.Dimensions {
	labelMax := measureMaxText(gtx, theme, font.Bold, "Size: ", "Filename: ", "Origin: ", "Comment: ").Size.X
	return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceEnd}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
				layout.Rigid(rightAlignedBoldLabel(theme, "Size:", labelMax)),
				layout.Flexed(1, label(theme, strconv.Itoa(pattern.Width)+"w X "+strconv.Itoa(pattern.Height)+"h")),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
				layout.Rigid(rightAlignedBoldLabel(theme, "Rule:", labelMax)),
				layout.Flexed(1, label(theme, pattern.Rule.Name())),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
				layout.Rigid(rightAlignedBoldLabel(theme, "Filename:", labelMax)),
				layout.Flexed(1, label(theme, pattern.Filename)),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
				layout.Rigid(rightAlignedBoldLabel(theme, "Origin:", labelMax)),
				layout.Flexed(1, label(theme, pattern.Origination)),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
				layout.Rigid(rightAlignedBoldLabel(theme, "Comment:", labelMax)),
				layout.Flexed(1, material.Label(theme, theme.TextSize, strings.Join(pattern.Comments, "\n")).Layout),
			)
		}),
	)
}

func (p *patternsPopout) layoutPreviewImage(pattern patterns.Pattern, gtx layout.Context, theme *material.Theme, maxWd, maxHt int) layout.Dimensions {
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
	return p.chooser.isFocused(gtx) || p.chkFilterCurrentRule.Focused(gtx) || p.chkInterlaced.Focused(gtx)
}
