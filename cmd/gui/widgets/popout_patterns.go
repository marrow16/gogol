package widgets

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"image"
	"image/color"
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
	cachedPattern        *patterns.Pattern
	cachedImage          *image.NRGBA
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
	return popoutLayout(gtx, func(gtx layout.Context) layout.Dimensions {
		dims := flexVertical(10,
			rigid(p.chooser.layout),
			rigid(flexHorizontal(10,
				rigid(p.radioPreview.Layout),
				rigid(p.radioMetadata.Layout),
				rigidFixedWidth(nil, 100, 0),
				rigid(p.chkFilterCurrentRule.Layout),
			)),
			rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.Y = int(float32(chd.Size.Y) * 15.5)
				gtx.Constraints.Max.X = p.chooser.dims.Size.X
				return p.layoutPreview(gtx, p.chooser.dims.Size.X, chd.Size.Y*15)
			}),
			conditionalRigid(p.chooser.currentItem() != nil, flexHorizontal(20,
				rigid(p.btnPlace.Layout),
				rigid(p.chkInterlaced.Layout)), nil),
		)(gtx)
		p.chooser.layoutDropdown(gtx)
		return dims
	})
}

func (p *patternsPopout) layoutPreview(gtx layout.Context, maxWd, maxHt int) layout.Dimensions {
	currentPattern := p.chooser.currentItem()
	switch {
	case currentPattern == nil:
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceAround}.Layout(gtx,
			rigidFixedWidth(label("(select a pattern)"), gtx.Constraints.Max.X, layout.Center),
		)
	case p.previewMode.Value == previewMetadata:
		return p.layoutPreviewMetadata(*currentPattern, gtx)
	default:
		return p.layoutPreviewImage(currentPattern, gtx, maxWd, maxHt)
	}
}

func (p *patternsPopout) layoutPreviewMetadata(pattern patterns.Pattern, gtx layout.Context) layout.Dimensions {
	labelMax := measureMaxText(gtx, font.Bold, "Size: ", "Filename: ", "Origin: ", "Comment: ").Size.X
	return flexVertical(0,
		rigid(flexHorizontal(20,
			rigidLabel("Size:", text.End, font.Bold, labelMax),
			flexed(label(strconv.Itoa(pattern.Width)+"w X "+strconv.Itoa(pattern.Height)+"h")),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Rule:", text.End, font.Bold, labelMax),
			flexed(label(pattern.Rule.Name())),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Filename:", text.End, font.Bold, labelMax),
			flexed(label(pattern.Filename)),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Origin:", text.End, font.Bold, labelMax),
			flexed(label(pattern.Origination)),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Comment:", text.End, font.Bold, labelMax),
			flexed(material.Label(theme, theme.TextSize, strings.Join(pattern.Comments, "\n")).Layout),
		)),
	)(gtx)
}

func scaleSparse(src *image.Paletted, scale float32) *image.NRGBA {
	sb := src.Bounds()
	w := max(1, int(float32(sb.Dx())*scale))
	h := max(1, int(float32(sb.Dy())*scale))
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	const background = uint8(0)
	bg := src.Palette[background]
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)
	for y := sb.Min.Y; y < sb.Max.Y; y++ {
		for x := sb.Min.X; x < sb.Max.X; x++ {
			i := src.ColorIndexAt(x, y)
			if i == background {
				continue
			}
			dx := int(float32(x-sb.Min.X) * scale)
			dy := int(float32(y-sb.Min.Y) * scale)
			if dx >= w {
				dx = w - 1
			}
			if dy >= h {
				dy = h - 1
			}
			dst.Set(dx, dy, src.Palette[i])
		}
	}
	return dst
}

func (p *patternsPopout) layoutPreviewImage(pattern *patterns.Pattern, gtx layout.Context, maxWd, maxHt int) layout.Dimensions {
	if pattern != p.cachedPattern {
		const minCellSize = 10
		p.cachedPattern = pattern
		if pattern.Width > maxWd || pattern.Height > maxHt {
			scale := min(float32(maxWd)/float32(pattern.Width), float32(maxHt)/float32(pattern.Height))
			rect := image.Rect(0, 0, pattern.Width, pattern.Height)
			img := image.NewPaletted(rect, color.Palette{
				0: p.core.settings.CellDeadColor,
				1: p.core.settings.CellAliveColor})
			pattern.DrawTo(patterns.Rotate0, func(row, col int, alive bool) {
				if alive {
					img.Pix[img.PixOffset(col, row)] = 1
				}
			})
			p.cachedImage = scaleSparse(img, scale)
		} else {
			offset := 0
			cellSize := min((maxWd-1)/pattern.Width, (maxHt-1)/pattern.Height)
			if cellSize > minCellSize {
				offset = 1
			}
			rect := image.Rect(0, 0, (cellSize*pattern.Width)+offset, (cellSize*pattern.Height)+offset)
			p.cachedImage = image.NewNRGBA(rect)
			draw.Draw(p.cachedImage, rect, &image.Uniform{p.core.settings.CellDeadColor}, image.Point{}, draw.Src)
			if cellSize > minCellSize {
				for y := 0; y <= pattern.Height; y++ {
					yy := y * cellSize
					draw.Draw(
						p.cachedImage,
						image.Rect(0, yy, pattern.Width*cellSize, yy+1),
						&image.Uniform{p.core.settings.CellBorderColor},
						image.Point{},
						draw.Src,
					)
				}
				for x := 0; x <= pattern.Width; x++ {
					xx := x * cellSize
					draw.Draw(
						p.cachedImage,
						image.Rect(xx, 0, xx+1, pattern.Height*cellSize),
						&image.Uniform{p.core.settings.CellBorderColor},
						image.Point{},
						draw.Src,
					)
				}
			}
			pattern.DrawTo(patterns.Rotate0, func(row, col int, alive bool) {
				if alive {
					draw.Draw(p.cachedImage, image.Rect(
						(col*cellSize)+offset,
						(row*cellSize)+offset,
						(col+1)*cellSize,
						(row+1)*cellSize),
						&image.Uniform{p.core.settings.CellAliveColor}, image.Point{}, draw.Src)
				}
			})
		}
	}
	return flexVertical(0,
		rigidImage(p.cachedImage),
	)(gtx)
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
