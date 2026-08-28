package widgets

import (
	"image"
	"slices"
	"strconv"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/imaging"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
)

type patternsPopout struct {
	parent        *menuPopup
	core          *Core
	chooser       *chooser[patterns.Pattern]
	previewMode   *widget.Enum
	radioPreview  *radioButton
	radioMetadata *radioButton
	radioSearch   *radioButton
	ruleClick     widget.Clickable
	btnPlace      *button
	chkInterlaced *checkbox
	cachedPattern *patterns.Pattern
	cachedImage   *image.NRGBA
	// search/filter controls...
	patternsCount        int
	currentRule          logic.Rule
	chkFilterCurrentRule *checkbox
	filterName           *input
	filterRule           *input
	filterRuleCurrent    *input
	filterWidthMin       *numberInput[int]
	filterWidthMax       *numberInput[int]
	filterHeightMin      *numberInput[int]
	filterHeightMax      *numberInput[int]
	filterFilename       *input
	filterOrigin         *input
	filterComment        *input
	btnFilterApply       *button
	btnFilterClear       *button
	filterApplied        bool
}

const (
	searchFilter    = "search"
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
		filterName:           newInput(nil, 256, nil).maximumWidth(30),
		filterRule:           newInput("sbSB/012345678", 21, nil).maximumWidth(30),
		filterRuleCurrent:    newInput(nil, 21, nil).maximumWidth(30),
		filterWidthMin:       newNumberInput[int](5, 0, 99999, 100, nil),
		filterWidthMax:       newNumberInput[int](5, 0, 99999, 100, nil),
		filterHeightMin:      newNumberInput[int](5, 0, 99999, 100, nil),
		filterHeightMax:      newNumberInput[int](5, 0, 99999, 100, nil),
		filterFilename:       newInput(nil, 256, nil).maximumWidth(30),
		filterOrigin:         newInput(nil, 256, nil).maximumWidth(30),
		filterComment:        newInput(nil, 256, nil).maximumWidth(30),
		btnFilterApply:       newButton("Apply Filter"),
		btnFilterClear:       newButton("Clear Filter"),
	}
	result.filterRuleCurrent.editor.ReadOnly = true
	result.radioPreview = newRadioButton(result.previewMode, previewImage, "Preview")
	result.radioMetadata = newRadioButton(result.previewMode, previewMetadata, "Metadata")
	result.radioSearch = newRadioButton(result.previewMode, searchFilter, "Search/filter")
	result.chooser = newChooser[patterns.Pattern](38,
		nil,
		result.patternSelected,
		func(pattern patterns.Pattern) string {
			return pattern.String()
		},
	)
	result.resetPatterns(true)
	return result
}

func (p *patternsPopout) patternSelected(pattern *patterns.Pattern) {
	//fmt.Printf("Pattern selected: %+v\n", pattern)
}

func (p *patternsPopout) setSelected(name string) {
	p.chooser.opened = false
	p.filterApplied = false
	if p.previewMode.Value != previewImage && p.previewMode.Value != previewMetadata {
		p.previewMode.Value = previewImage
	}
	p.resetPatterns(true)
	p.chooser.editor.SetText(name)
}

func (p *patternsPopout) resetPatterns(force bool) {
	do := force || p.patternsCount != len(patterns.PatternLibrary)
	if !do && p.filterApplied && p.chkFilterCurrentRule.Checked() {
		if p.currentRule == nil || p.currentRule.Permutation() != p.core.gridHolder.grid.Rule.Permutation() {
			do = true
			p.currentRule = p.core.gridHolder.grid.Rule
		}
	}
	if do {
		p.patternsCount = len(patterns.PatternLibrary)
		items := make([]patterns.Pattern, 0, len(patterns.PatternLibrary))
		if filter := p.buildPatternsFilter(); filter != nil {
			for _, pattern := range patterns.PatternLibrary {
				if filter.matches(pattern) {
					items = append(items, pattern)
				}
			}
		} else {
			for _, pattern := range patterns.PatternLibrary {
				items = append(items, pattern)
			}
		}
		slices.SortStableFunc(items, func(a, b patterns.Pattern) int {
			return strings.Compare(strings.ToLower(a.String()), strings.ToLower(b.String()))
		})
		p.chooser.resetItems(items)
	}
}

type patternFilter struct {
	fns []func(pattern patterns.Pattern) bool
}

func (f *patternFilter) matches(pattern patterns.Pattern) bool {
	for _, fn := range f.fns {
		if !fn(pattern) {
			return false
		}
	}
	return true
}

func (p *patternsPopout) buildPatternsFilter() *patternFilter {
	if !p.filterApplied {
		return nil
	}
	result := &patternFilter{}
	if p.chkFilterCurrentRule.Checked() {
		result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
			return pattern.Rule != nil && pattern.Rule.Permutation() == p.core.gridHolder.grid.Rule.Permutation()
		})
	} else if s := p.filterRule.editor.Text(); len(s) != 0 {
		if r, err := logic.NewRuleRle("", s); err == nil {
			result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
				return pattern.Rule != nil && pattern.Rule.Permutation() == r.Permutation()
			})
		} else {
			p.filterRule.setText("")
		}
	}
	if s := p.filterName.editor.Text(); len(s) != 0 {
		s = strings.ToLower(s)
		result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
			return strings.Contains(strings.ToLower(pattern.Name), s)
		})
	}
	if s := p.filterFilename.editor.Text(); len(s) != 0 {
		s = strings.ToLower(s)
		result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
			return strings.Contains(strings.ToLower(pattern.Filename), s)
		})
	}
	if s := p.filterOrigin.editor.Text(); len(s) != 0 {
		s = strings.ToLower(s)
		result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
			return strings.Contains(strings.ToLower(pattern.Origination), s)
		})
	}
	if s := p.filterComment.editor.Text(); len(s) != 0 {
		s = strings.ToLower(s)
		result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
			return strings.Contains(strings.ToLower(strings.Join(pattern.Comments, "\n")), s)
		})
	}
	wMin, wMax := -1, -1
	if p.filterWidthMin.input.editor.Text() != "" {
		wMin = p.filterWidthMin.current()
	}
	if p.filterWidthMax.input.editor.Text() != "" {
		wMax = p.filterWidthMax.current()
	}
	switch {
	case wMin != -1 && wMax != -1:
		wMin, wMax = min(wMin, wMax), max(wMin, wMax)
		result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
			return pattern.Width >= wMin && pattern.Width <= wMax
		})
	case wMin != -1:
		result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
			return pattern.Width >= wMin
		})
	case wMax != -1:
		result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
			return pattern.Width <= wMax
		})
	}
	hMin, hMax := -1, -1
	if p.filterHeightMin.input.editor.Text() != "" {
		hMin = p.filterHeightMin.current()
	}
	if p.filterHeightMax.input.editor.Text() != "" {
		hMax = p.filterHeightMax.current()
	}
	switch {
	case hMin != -1 && hMax != -1:
		hMin, hMax = min(hMin, hMax), max(hMin, hMax)
		result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
			return pattern.Height >= hMin && pattern.Height <= hMax
		})
	case hMin != -1:
		result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
			return pattern.Height >= hMin
		})
	case hMax != -1:
		result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
			return pattern.Height <= hMax
		})
	}
	if len(result.fns) == 0 {
		p.filterApplied = false
		return nil
	}
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
	chd := measureText(gtx, "M")
	gtx.Constraints.Min.Y = chd.Size.Y * 20
	p.previewMode.Update(gtx)
	p.resetPatterns(false)
	return popoutLayout(gtx, func(gtx layout.Context) layout.Dimensions {
		dims := flexVertical(10,
			rigid(p.chooser.layout),
			rigid(flexHorizontal(10,
				rigid(p.radioPreview.Layout),
				rigid(p.radioMetadata.Layout),
				rigid(p.radioSearch.Layout),
			)),
			rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.Y = int(float32(chd.Size.Y) * 15.5)
				gtx.Constraints.Max.X = p.chooser.dims.Size.X
				return p.layoutPreview(gtx, p.chooser.dims.Size.X, chd.Size.Y*15)
			}),
			conditionalRigid(p.previewMode.Value != searchFilter && p.chooser.currentItem() != nil, flexHorizontal(20,
				rigid(p.btnPlace.Layout),
				rigid(p.chkInterlaced.Layout)), nil),
			conditionalRigid(p.previewMode.Value == searchFilter, flexHorizontal(20,
				rigid(p.btnFilterApply.Layout),
				conditionalRigid(p.filterApplied, p.btnFilterClear.Layout, nil),
				conditionalRigid(p.filterApplied, label("("+strconv.Itoa(len(p.chooser.items))+"/"+strconv.Itoa(len(patterns.PatternLibrary))+")"), nil),
			), nil),
		)(gtx)
		p.chooser.layoutDropdown(gtx)
		return dims
	})
}

func (p *patternsPopout) layoutPreview(gtx layout.Context, maxWd, maxHt int) layout.Dimensions {
	currentPattern := p.chooser.currentItem()
	switch {
	case p.previewMode.Value == searchFilter:
		return p.layoutSearchFilter(gtx)
	case currentPattern == nil:
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceAround}.Layout(gtx,
			rigidFixedWidth(label("(select a pattern)"), gtx.Constraints.Max.X, layout.Center),
		)
	case p.previewMode.Value == previewMetadata:
		return p.layoutPreviewMetadata(currentPattern, gtx)
	default:
		return p.layoutPreviewImage(currentPattern, gtx, maxWd, maxHt)
	}
}

func (p *patternsPopout) layoutSearchFilter(gtx layout.Context) layout.Dimensions {
	if r := p.core.gridHolder.grid.Rule; r.IsCustom() {
		p.filterRuleCurrent.setText(r.Rle())
	} else {
		p.filterRuleCurrent.setText(r.Name())
	}
	if p.btnFilterApply.Clicked(gtx) {
		p.filterApplied = true
		p.resetPatterns(true)
	}
	if p.btnFilterClear.Clicked(gtx) {
		p.filterApplied = false
		p.resetPatterns(true)
	}
	p.chkFilterCurrentRule.Update(gtx)
	labelMax := measureMaxText(gtx, font.Bold, "Name: ", "Rule: ", "Width: ", "Height: ", "Filename: ", "Origin: ", "Comment: ").Size.X
	return flexVertical(10,
		rigid(flexHorizontal(20,
			rigidLabel("Rule:", text.End, font.Bold, labelMax),
			conditionalFlexed(p.chkFilterCurrentRule.Checked(), p.filterRuleCurrent.layout, p.filterRule.layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("", 0, 0, labelMax),
			rigid(p.chkFilterCurrentRule.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Width:", text.End, font.Bold, labelMax),
			flexed(p.filterWidthMin.layout),
			rigid(label("to")),
			flexed(p.filterWidthMax.layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Height:", text.End, font.Bold, labelMax),
			flexed(p.filterHeightMin.layout),
			rigid(label("to")),
			flexed(p.filterHeightMax.layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Name:", text.End, font.Bold, labelMax),
			flexed(p.filterName.layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Filename:", text.End, font.Bold, labelMax),
			flexed(p.filterFilename.layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Origin:", text.End, font.Bold, labelMax),
			flexed(p.filterOrigin.layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Comment:", text.End, font.Bold, labelMax),
			flexed(p.filterComment.layout),
		)),
	)(gtx)
}

func (p *patternsPopout) layoutPreviewMetadata(pattern *patterns.Pattern, gtx layout.Context) layout.Dimensions {
	labelMax := measureMaxText(gtx, font.Bold, "Size: ", "Name: ", "Filename: ", "Origin: ", "Comment: ").Size.X
	if p.ruleClick.Clicked(gtx) && pattern.Rule != nil {
		p.core.setRule(pattern.Rule)
	}
	return flexVertical(0,
		rigid(flexHorizontal(20,
			rigidLabel("Rule:", text.End, font.Bold, labelMax),
			conditionalRigid(pattern.Rule != nil, linkLabel(&p.ruleClick, pattern.Rule.Name()), nil),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Size:", text.End, font.Bold, labelMax),
			flexed(label(strconv.Itoa(pattern.Width)+"w X "+strconv.Itoa(pattern.Height)+"h")),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Name:", text.End, font.Bold, labelMax),
			flexed(label(pattern.Name)),
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
	p.resetPatterns(false)
}

func (p *patternsPopout) hasFocus(gtx layout.Context) bool {
	_, radios := p.previewMode.Focused()
	return radios || p.chooser.isFocused(gtx) || p.chkFilterCurrentRule.isFocused(gtx) || p.chkInterlaced.isFocused(gtx) ||
		p.btnPlace.isFocused(gtx) || gtx.Focused(&p.ruleClick)
}
