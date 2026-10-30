package gui

import (
	"github.com/marrow16/gogol/gui/controls"
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
	chooser       *controls.Chooser[patterns.Pattern]
	previewMode   *widget.Enum
	radioPreview  *controls.RadioButton
	radioMetadata *controls.RadioButton
	radioSearch   *controls.RadioButton
	ruleClick     widget.Clickable
	btnPlace      *controls.Button
	chkInterlaced *controls.Checkbox
	cachedPattern *patterns.Pattern
	cachedImage   *image.NRGBA
	// search/filter controls...
	patternsCount        int
	currentRule          *logic.Rule
	chkFilterCurrentRule *controls.Checkbox
	filterName           *controls.Input
	filterRule           *controls.Input
	filterRuleCurrent    *controls.Input
	filterWidthMin       *controls.NumberInput[int]
	filterWidthMax       *controls.NumberInput[int]
	filterHeightMin      *controls.NumberInput[int]
	filterHeightMax      *controls.NumberInput[int]
	filterFilename       *controls.Input
	filterOrigin         *controls.Input
	filterComment        *controls.Input
	btnFilterApply       *controls.Button
	btnFilterClear       *controls.Button
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
		chkFilterCurrentRule: controls.NewCheckBox(theme, "Filter current rule", false),
		btnPlace:             controls.NewButton(theme, "Place"),
		chkInterlaced:        controls.NewCheckBox(theme, "Interlaced", false),
		filterName:           controls.NewInput(theme, nil, 256, nil).MaximumWidth(30),
		filterRule:           controls.NewInput(theme, "sbSB/012345678", 21, nil).MaximumWidth(30),
		filterRuleCurrent:    controls.NewInput(theme, nil, 21, nil).MaximumWidth(30),
		filterWidthMin:       controls.NewNumberInput[int](theme, 5, 0, 99999, 100, nil),
		filterWidthMax:       controls.NewNumberInput[int](theme, 5, 0, 99999, 100, nil),
		filterHeightMin:      controls.NewNumberInput[int](theme, 5, 0, 99999, 100, nil),
		filterHeightMax:      controls.NewNumberInput[int](theme, 5, 0, 99999, 100, nil),
		filterFilename:       controls.NewInput(theme, nil, 256, nil).MaximumWidth(30),
		filterOrigin:         controls.NewInput(theme, nil, 256, nil).MaximumWidth(30),
		filterComment:        controls.NewInput(theme, nil, 256, nil).MaximumWidth(30),
		btnFilterApply:       controls.NewButton(theme, "Apply Filter"),
		btnFilterClear:       controls.NewButton(theme, "Clear Filter"),
	}
	result.filterRuleCurrent.ReadOnly(true)
	result.radioPreview = controls.NewRadioButton(theme, result.previewMode, previewImage, "Preview")
	result.radioMetadata = controls.NewRadioButton(theme, result.previewMode, previewMetadata, "Metadata")
	result.radioSearch = controls.NewRadioButton(theme, result.previewMode, searchFilter, "Search/filter")
	result.chooser = controls.NewChooser[patterns.Pattern](theme, 38,
		nil,
		result.patternSelected,
		func(pattern patterns.Pattern) string {
			return pattern.String()
		},
	)
	result.chooser.SetText(c.settings.SelectedPattern)
	result.resetPatterns(true)
	return result
}

func (p *patternsPopout) patternSelected(pattern *patterns.Pattern) {
	if pattern != nil {
		p.core.settings.SelectedPattern = pattern.Name
	}
}

func (p *patternsPopout) setSelected(name string) {
	p.chooser.Opened(false)
	p.filterApplied = false
	if p.previewMode.Value != previewImage && p.previewMode.Value != previewMetadata {
		p.previewMode.Value = previewImage
	}
	p.resetPatterns(true)
	p.chooser.SetText(name)
}

func (p *patternsPopout) resetPatterns(force bool) {
	l := patterns.Library.Len()
	do := force || p.patternsCount != l
	if !do && p.filterApplied && p.chkFilterCurrentRule.Checked() {
		if p.currentRule == nil || p.currentRule.Permutation() != p.core.gridHolder.grid.Rule().Permutation() {
			do = true
			rule := p.core.gridHolder.grid.Rule()
			p.currentRule = &rule
		}
	}
	if do {
		p.patternsCount = l
		items := make([]patterns.Pattern, 0, l)
		if filter := p.buildPatternsFilter(); filter != nil {
			for _, pattern := range patterns.Library.Filtered(filter.fns) {
				items = append(items, pattern)
			}
		} else {
			for _, pattern := range patterns.Library.All() {
				items = append(items, pattern)
			}
		}
		slices.SortStableFunc(items, func(a, b patterns.Pattern) int {
			return strings.Compare(strings.ToLower(a.String()), strings.ToLower(b.String()))
		})
		p.chooser.ResetItems(items)
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
			return pattern.Rule != nil && pattern.Rule.Permutation() == p.core.gridHolder.grid.Rule().Permutation()
		})
	} else if s := p.filterRule.Text(); len(s) != 0 {
		if r, err := logic.NewRuleRle("", s); err == nil {
			perm := r.Permutation()
			result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
				return pattern.Rule != nil && pattern.Rule.Permutation() == perm
			})
		} else {
			p.filterRule.SetText("")
		}
	}
	if s := p.filterName.Text(); len(s) != 0 {
		s = strings.ToLower(s)
		result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
			return strings.Contains(strings.ToLower(pattern.Name), s)
		})
	}
	if s := p.filterFilename.Text(); len(s) != 0 {
		s = strings.ToLower(s)
		result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
			return strings.Contains(strings.ToLower(pattern.Filename), s)
		})
	}
	if s := p.filterOrigin.Text(); len(s) != 0 {
		s = strings.ToLower(s)
		result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
			return strings.Contains(strings.ToLower(pattern.Origination), s)
		})
	}
	if s := p.filterComment.Text(); len(s) != 0 {
		s = strings.ToLower(s)
		result.fns = append(result.fns, func(pattern patterns.Pattern) bool {
			return strings.Contains(strings.ToLower(strings.Join(pattern.Comments, "\n")), s)
		})
	}
	wMin, wMax := -1, -1
	if p.filterWidthMin.Text() != "" {
		wMin = p.filterWidthMin.Current()
	}
	if p.filterWidthMax.Text() != "" {
		wMax = p.filterWidthMax.Current()
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
	if p.filterHeightMin.Text() != "" {
		hMin = p.filterHeightMin.Current()
	}
	if p.filterHeightMax.Text() != "" {
		hMax = p.filterHeightMax.Current()
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
	curr := p.chooser.Text()
	if ptn, found := patterns.Library.Get(curr); found {
		return &ptn, p.chkInterlaced.Checked()
	}
	return nil, false
}

func (p *patternsPopout) layout(gtx layout.Context) layout.Dimensions {
	if p.btnPlace.Clicked(gtx) {
		if pattern := p.chooser.CurrentItem(); pattern != nil {
			p.core.startPatternPlace(gtx, pattern, p.chkInterlaced.Checked())
		}
	}
	chd := measureText(gtx, "M")
	gtx.Constraints.Min.Y = chd.Size.Y * 20
	p.previewMode.Update(gtx)
	p.resetPatterns(false)
	return popoutLayout(gtx, func(gtx layout.Context) layout.Dimensions {
		dims := flexVertical(10,
			rigid(p.chooser.Layout),
			rigid(flexHorizontal(10,
				rigid(p.radioPreview.Layout),
				rigid(p.radioMetadata.Layout),
				rigid(p.radioSearch.Layout),
			)),
			rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.Y = int(float32(chd.Size.Y) * 15.5)
				gtx.Constraints.Max.X = p.chooser.Width()
				return p.layoutPreview(gtx, p.chooser.Width(), chd.Size.Y*15)
			}),
			conditionalRigid(p.previewMode.Value != searchFilter && p.chooser.CurrentItem() != nil, flexHorizontal(20,
				rigid(p.btnPlace.Layout),
				rigid(p.chkInterlaced.Layout)), nil),
			conditionalRigid(p.previewMode.Value == searchFilter, flexHorizontal(20,
				rigid(p.btnFilterApply.Layout),
				conditionalRigid(p.filterApplied, p.btnFilterClear.Layout, nil),
				conditionalRigid(p.filterApplied, label("("+strconv.Itoa(p.chooser.Len())+"/"+strconv.Itoa(patterns.Library.Len())+")"), nil),
			), nil),
		)(gtx)
		p.chooser.LayoutDropdown(gtx)
		return dims
	})
}

func (p *patternsPopout) layoutPreview(gtx layout.Context, maxWd, maxHt int) layout.Dimensions {
	currentPattern := p.chooser.CurrentItem()
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
	if r := p.core.gridHolder.grid.Rule(); r.IsCustom() {
		p.filterRuleCurrent.SetText(r.Rle())
	} else {
		p.filterRuleCurrent.SetText(r.Name())
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
	labelMax := measureMaxText(gtx, font.Bold, "Name: ", "Rule: ", "Width: ", "Height: ", "Filename: ", "Origin: ", "Comment: ")
	return flexVertical(10,
		rigid(flexHorizontal(20,
			rigidLabel("Rule:", text.End, font.Bold, labelMax),
			conditionalFlexed(p.chkFilterCurrentRule.Checked(), p.filterRuleCurrent.Layout, p.filterRule.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("", 0, 0, labelMax),
			rigid(p.chkFilterCurrentRule.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Width:", text.End, font.Bold, labelMax),
			flexed(p.filterWidthMin.Layout),
			rigid(label("to")),
			flexed(p.filterWidthMax.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Height:", text.End, font.Bold, labelMax),
			flexed(p.filterHeightMin.Layout),
			rigid(label("to")),
			flexed(p.filterHeightMax.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Name:", text.End, font.Bold, labelMax),
			flexed(p.filterName.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Filename:", text.End, font.Bold, labelMax),
			flexed(p.filterFilename.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Origin:", text.End, font.Bold, labelMax),
			flexed(p.filterOrigin.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Comment:", text.End, font.Bold, labelMax),
			flexed(p.filterComment.Layout),
		)),
	)(gtx)
}

func (p *patternsPopout) layoutPreviewMetadata(pattern *patterns.Pattern, gtx layout.Context) layout.Dimensions {
	labelMax := measureMaxText(gtx, font.Bold, "Size: ", "Name: ", "Filename: ", "Origin: ", "Comment: ")
	if p.ruleClick.Clicked(gtx) && pattern.Rule != nil {
		p.core.setRule(*pattern.Rule)
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
			flexed(func(gtx layout.Context) layout.Dimensions {
				if pattern.Filename == "" {
					return label("[built-in]", font.Italic)(gtx)
				}
				return label(pattern.Filename)(gtx)
			}),
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
	p.chooser.Opened(false)
	p.resetPatterns(false)
}

func (p *patternsPopout) hasFocus(gtx layout.Context) bool {
	_, radios := p.previewMode.Focused()
	return radios || p.chooser.IsFocused(gtx) || p.chkFilterCurrentRule.IsFocused(gtx) || p.chkInterlaced.IsFocused(gtx) ||
		p.btnPlace.IsFocused(gtx) || gtx.Focused(&p.ruleClick) ||
		p.chkFilterCurrentRule.IsFocused(gtx) || p.filterName.IsFocused(gtx) || p.filterRule.IsFocused(gtx) ||
		p.filterRuleCurrent.IsFocused(gtx) || p.filterWidthMin.IsFocused(gtx) || p.filterWidthMax.IsFocused(gtx) ||
		p.filterHeightMin.IsFocused(gtx) || p.filterHeightMax.IsFocused(gtx) ||
		p.filterFilename.IsFocused(gtx) || p.filterOrigin.IsFocused(gtx) || p.filterComment.IsFocused(gtx) ||
		p.btnFilterApply.IsFocused(gtx) || p.btnFilterClear.IsFocused(gtx)

}
