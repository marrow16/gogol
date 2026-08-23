package widgets

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/imaging"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"image"
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

func (p *patternsPopout) layoutPreviewImage(pattern *patterns.Pattern, gtx layout.Context, maxWd, maxHt int) layout.Dimensions {
	if pattern != p.cachedPattern {
		const minCellSize = 10
		p.cachedPattern = pattern
		if pattern.Width > maxWd || pattern.Height > maxHt {
			img := imaging.PatternImagePaletted(*pattern, imaging.Config{
				CellSize:   1,
				Borders:    false,
				AliveColor: p.core.settings.CellAliveColor,
				DeadColor:  p.core.settings.CellDeadColor,
			})
			scale := min(float32(maxWd)/float32(pattern.Width), float32(maxHt)/float32(pattern.Height))
			p.cachedImage = imaging.ScaleSparse(img, scale)
		} else {
			cellSize := min((maxWd-1)/pattern.Width, (maxHt-1)/pattern.Height)
			borders := cellSize > minCellSize
			p.cachedImage = imaging.PatternImage(*pattern, imaging.Config{
				CellSize:    cellSize,
				Borders:     borders,
				AliveColor:  p.core.settings.CellAliveColor,
				DeadColor:   p.core.settings.CellDeadColor,
				BorderColor: p.core.settings.CellBorderColor,
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
