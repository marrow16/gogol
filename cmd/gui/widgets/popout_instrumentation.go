package widgets

import (
	"errors"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/animator"
	"image"
	"path/filepath"
	"strconv"
)

type instrumentationPopout struct {
	parent *menuPopup
	core   *Core
	// repeat...
	chkRepeatDetect *checkbox
	btnRepeatReset  *button
	btnRepeatSave   *button
	// record...
	chkRecord        *checkbox
	skipBackBy       *numberInput[int]
	btnRecordReset   *button
	btnSaveAnimation *button
	animationSaving  bool
	animationResult  *animationResult
	linkAnimation    widget.Clickable
	animationFormat  *widget.Enum
	radioMp4         *radioButton
	radioGif         *radioButton
	// heat map...
	chkHeatMap       *checkbox
	heatMapType      *widget.Enum
	btnHeatMapReset  *button
	btnHeatMapReveal *button
	btnHeatMapSave   *button
	radioActivity    *radioButton
	radioOccupancy   *radioButton
	radioBirths      *radioButton
	radioFreshness   *radioButton
	radioPhaseParity *radioButton
	radioAll         *radioButton
}

type animationResult struct {
	filename string
	err      error
}

func newInstrumentationPopout(p *menuPopup, c *Core) *instrumentationPopout {
	animFormat := c.settings.AnimationFormat
	if animFormat != "gif" && animFormat != "mp4" {
		animFormat = "gif"
	}
	result := &instrumentationPopout{
		parent:           p,
		core:             c,
		chkRepeatDetect:  newCheckBox("Repeat Detect", c.instrumentRepeat != nil),
		btnRepeatReset:   newButton("Reset"),
		btnRepeatSave:    newButton("Save Report"),
		chkRecord:        newCheckBox("Record", c.instrumentRecord != nil),
		btnRecordReset:   newButton("Reset"),
		btnSaveAnimation: newButton("Save Animation"),
		animationFormat:  &widget.Enum{Value: animFormat},
		chkHeatMap:       newCheckBox("Heat Mapping", c.instrumentHeatMap != nil),
		btnHeatMapReset:  newButton("Reset"),
		btnHeatMapReveal: newButton("Reveal"),
		btnHeatMapSave:   newButton("Save Image"),
		heatMapType:      &widget.Enum{Value: c.heatMapperType.String()},
	}
	result.radioActivity = newRadioButton(result.heatMapType, activityHeatMapper.String(), "Activity")
	result.radioOccupancy = newRadioButton(result.heatMapType, occupancyHeatMapper.String(), "Occupancy")
	result.radioBirths = newRadioButton(result.heatMapType, birthsHeatMapper.String(), "Births")
	result.radioFreshness = newRadioButton(result.heatMapType, freshnessHeatMapper.String(), "Freshness")
	result.radioPhaseParity = newRadioButton(result.heatMapType, phaseParityHeatMapper.String(), "Phase Parity")
	result.radioAll = newRadioButton(result.heatMapType, allHeatMapper.String(), "All")
	result.radioGif = newRadioButton(result.animationFormat, "gif", "Gif")
	result.radioMp4 = newRadioButton(result.animationFormat, "mp4", "Mp4")
	result.skipBackBy = newNumberInput[int](4, 1, 9999, 100, result.skipBackByChanged)
	return result
}

func (p *instrumentationPopout) skipBackByChanged(n int) {
	if n > 0 {
		p.core.settings.SkipBackBy = n
	}
}

func (p *instrumentationPopout) reset() {
	p.chkRepeatDetect.SetChecked(p.core.instrumentRepeat != nil)
	p.chkRecord.SetChecked(p.core.instrumentRecord != nil)
	p.chkHeatMap.SetChecked(p.core.instrumentHeatMap != nil)
	p.heatMapType.Value = p.core.heatMapperType.String()
	p.skipBackBy.setValue(p.core.settings.SkipBackBy)
	af := p.core.settings.AnimationFormat
	if af != "gif" && af != "mp4" {
		af = "gif"
	}
	p.animationFormat.Value = af
}

func (p *instrumentationPopout) layout(gtx layout.Context) layout.Dimensions {
	p.update(gtx)
	width := measureText(gtx, "When enabled, will stop stepping (and step ahead) extra").Size.X
	return popoutLayout(gtx, flexVertical(0,
		rigidFixedWidth(p.chkRepeatDetect.Layout, width, 0),
		rigid(func(gtx layout.Context) layout.Dimensions {
			if !p.chkRepeatDetect.Checked() {
				gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
				desc := material.Label(theme, theme.TextSize, "When enabled, will stop stepping (and step ahead) when the grid repeats.\nInformation about repeats will be shown here.")
				desc.TextSize = desc.TextSize - 2
				desc.Font.Style = font.Italic
				return desc.Layout(gtx)
			} else {
				return p.layoutRepeat(gtx)
			}
		}),
		rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
			horizontalLine(gtx, popupBorder, width, 1)
			return layout.Dimensions{Size: image.Point{X: width, Y: 1}}
		}),
		rigidFixedWidth(p.chkRecord.Layout, width, 0),
		rigid(func(gtx layout.Context) layout.Dimensions {
			if !p.chkRecord.Checked() {
				gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
				desc := material.Label(theme, theme.TextSize, "Records every generation, enabling backward stepping and animation export.")
				desc.TextSize = desc.TextSize - 2
				desc.Font.Style = font.Italic
				return desc.Layout(gtx)
			} else if p.core.instrumentRecord != nil {
				return p.layoutRecord(gtx)
			} else {
				return layout.Dimensions{}
			}
		}),
		rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
			horizontalLine(gtx, popupBorder, width, 1)
			return layout.Dimensions{Size: image.Point{X: width, Y: 1}}
		}),
		rigidFixedWidth(p.chkHeatMap.Layout, width, 0),
		rigid(func(gtx layout.Context) layout.Dimensions {
			if !p.chkHeatMap.Checked() {
				gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
				desc := material.Label(theme, theme.TextSize, "Accumulates various types of grid activity for later display as a visual heat map.")
				desc.TextSize = desc.TextSize - 2
				desc.Font.Style = font.Italic
				return desc.Layout(gtx)
			} else if p.core.instrumentHeatMap != nil {
				return p.layoutHeatMap(gtx)
			} else {
				return layout.Dimensions{}
			}
		}),
	))
}

func (p *instrumentationPopout) update(gtx layout.Context) {
	if p.chkRepeatDetect.Update(gtx) {
		p.core.setInstrumentationRepeat(p.chkRepeatDetect.Checked())
	}
	if p.chkRecord.Update(gtx) {
		p.core.setInstrumentationRecord(p.chkRecord.Checked())
		if !p.chkRecord.Checked() {
			p.animationResult = nil
		}
	}
	if p.btnRepeatReset.Clicked(gtx) {
		p.core.setInstrumentationRepeat(true)
	}
	if p.btnRepeatSave.Clicked(gtx) {
		p.core.saveRepeatDetect()
	}
	if p.btnRecordReset.Clicked(gtx) {
		if !p.animationSaving {
			p.animationResult = nil
		}
		p.core.setInstrumentationRecord(true)
	}
	if p.btnSaveAnimation.Clicked(gtx) {
		if !p.animationSaving && p.core.instrumentRecord != nil {
			p.animationSaving = true
			p.animationResult = nil
			if p.core.instrumentRecord.FramesCount() > 0 {
				p.saveAnimation()
			} else {
				p.animationSaving = false
				p.animationResult = &animationResult{
					err: errors.New("No steps to animate"),
				}
			}
		}
	}
	if p.linkAnimation.Clicked(gtx) && !p.animationSaving && p.animationResult != nil && p.animationResult.err == nil {
		openInBrowser(p.animationResult.filename)
	}
	if p.chkHeatMap.Update(gtx) {
		if p.chkHeatMap.Checked() {
			p.core.setInstrumentationHeatMapper(p.selectedHeatMapType())
		} else {
			p.core.setInstrumentationHeatMapper(noHeatMapper)
		}
	}
	if p.heatMapType.Update(gtx) {
		nt := p.selectedHeatMapType()
		if nt != p.core.heatMapperType {
			p.core.setInstrumentationHeatMapper(nt)
		}
	}
	if p.btnHeatMapReset.Clicked(gtx) {
		p.core.setInstrumentationHeatMapper(p.selectedHeatMapType())
	}
	if p.btnHeatMapReveal.Clicked(gtx) {
		p.core.showHeatMap()
	}
	if p.btnHeatMapSave.Clicked(gtx) {
		p.core.saveHeatMapImage()
	}
}

func (p *instrumentationPopout) selectedHeatMapType() heatMapperType {
	hmt := heatMapperTypeFrom(p.heatMapType.Value)
	if hmt == noHeatMapper {
		hmt = activityHeatMapper
		p.heatMapType.Value = hmt.String()
	}
	return hmt
}

func (p *instrumentationPopout) foundLabel(lblFound, lblNotFound string) string {
	if p.core.instrumentRepeat.Found {
		return lblFound
	}
	return lblNotFound
}

func (p *instrumentationPopout) layoutRepeat(gtx layout.Context) layout.Dimensions {
	labelMax := measureMaxText(gtx, font.Bold, "Examined: ", "Found: ", "First: ", "Repeat: ", "Period: ").Size.X
	return layout.Inset{Left: 16, Bottom: 4}.Layout(gtx, flexVertical(0,
		rigid(flexHorizontal(20,
			rigidLabel("Examined:", text.End, 0, labelMax),
			rigidLabel(commas(strconv.FormatUint(p.core.instrumentRepeat.Steps, 10)), 0, 0, 0),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Found:", text.End, 0, labelMax),
			rigidLabel(p.foundLabel("Yes", "No"), 0, 0, 0),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("First:", text.End, 0, labelMax),
			rigidLabel(p.foundLabel(commas(strconv.FormatUint(p.core.instrumentRepeat.FirstStep, 10)), "--"), 0, 0, 0),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Repeat:", text.End, 0, labelMax),
			rigidLabel(p.foundLabel(commas(strconv.FormatUint(p.core.instrumentRepeat.RepeatStep, 10)), "--"), 0, 0, 0),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Period:", text.End, 0, labelMax),
			rigidLabel(p.foundLabel(commas(strconv.FormatUint(p.core.instrumentRepeat.Period, 10)), "--"), 0, 0, 0),
		)),
		rigid(flexHorizontal(20,
			rigid(inset(4, 4, 0, 0, p.btnRepeatReset.Layout)),
			rigid(inset(4, 4, 0, 0, p.btnRepeatSave.Layout)),
		)),
	))
}

func (p *instrumentationPopout) layoutRecord(gtx layout.Context) layout.Dimensions {
	labelMax := measureMaxText(gtx, font.Bold, "Steps recorded: ", "Skip back by: ").Size.X
	canSave := !p.animationSaving && p.core.instrumentRecord.FramesCount() > 0
	if p.animationFormat.Update(gtx) {
		p.core.settings.AnimationFormat = p.animationFormat.Value
	}
	return layout.Inset{Left: 16, Bottom: 4}.Layout(gtx, flexVertical(0,
		rigid(flexHorizontal(20,
			rigidLabel("Steps recorded:", text.End, 0, labelMax),
			rigidLabel(commas(strconv.Itoa(p.core.instrumentRecord.FramesCount())), 0, 0, 0),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Skip back by:", text.End, 0, labelMax),
			flexed(p.skipBackBy.layout),
		)),
		rigid(flexHorizontal(20,
			rigid(inset(4, 4, 0, 0, p.btnRecordReset.Layout)),
			conditionalRigid(canSave, inset(4, 4, 0, 0, p.btnSaveAnimation.Layout), nil),
			conditionalRigid(canSave, inset(4, 4, 0, 0, p.radioGif.Layout), nil),
			conditionalRigid(canSave, inset(4, 4, 0, 0, p.radioMp4.Layout), nil),
		)),
		rigid(func(gtx layout.Context) layout.Dimensions {
			if p.animationSaving {
				return label("Saving animation - please wait...")(gtx)
			} else if p.animationResult != nil {
				if p.animationResult.err == nil {
					return flexHorizontal(20,
						rigidLabel("Saved to:", 0, 0, 0),
						rigid(linkLabel(&p.linkAnimation, filepath.Base(p.animationResult.filename))))(gtx)
				} else {
					return errorLabel(p.animationResult.err)(gtx)
				}
			} else {
				return layout.Dimensions{}
			}
		}),
	))
}

func (p *instrumentationPopout) saveAnimation() {
	filename, err := resolveSavePath(p.core.nowFilename("Grid", "."+p.animationFormat.Value))
	if err != nil {
		p.animationResult = &animationResult{
			filename: filename,
			err:      err,
		}
		window.Invalidate()
		return
	}
	recorder := p.core.instrumentRecord
	go func() {
		ani := animator.NewAnimator(p.core.settings.CellSize, p.core.settings.CellAliveColor, p.core.settings.CellDeadColor, p.core.settings.CellBorderColor, p.core.settings.CellBorders, p.core.settings.AnimationFormat)
		err := ani.Animate(filename, recorder)
		p.animationSaving = false
		p.animationResult = &animationResult{
			filename: filename,
			err:      err,
		}
		window.Invalidate()
	}()
}

func (p *instrumentationPopout) layoutHeatMap(gtx layout.Context) layout.Dimensions {
	labelMax := measureMaxText(gtx, font.Bold, "Steps: ", "Type: ", "Maximum: ").Size.X
	return layout.Inset{Left: 16, Bottom: 4}.Layout(gtx, flexVertical(0,
		rigid(flexHorizontal(20,
			rigidLabel("Type:", text.End, 0, labelMax),
			rigid(p.radioActivity.Layout),
			rigid(p.radioOccupancy.Layout),
			rigid(p.radioBirths.Layout),
		)),
		rigid(flexHorizontal(20,
			rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = labelMax
				return layout.Dimensions{Size: image.Point{X: labelMax}}
			}),
			rigid(p.radioFreshness.Layout),
			rigid(p.radioPhaseParity.Layout),
			rigid(p.radioAll.Layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Maximum:", text.End, 0, labelMax),
			rigidLabel(commas(strconv.FormatUint(p.core.instrumentHeatMap.Maximum(), 10)), 0, 0, 0),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Steps:", text.End, 0, labelMax),
			rigidLabel(commas(strconv.FormatUint(p.core.instrumentHeatMap.StepsCount(), 10)), 0, 0, 0),
		)),
		rigid(flexHorizontal(20,
			rigid(inset(4, 4, 0, 0, p.btnHeatMapReset.Layout)),
			rigid(inset(4, 4, 0, 0, p.btnHeatMapReveal.Layout)),
			rigid(inset(4, 4, 0, 0, p.btnHeatMapSave.Layout)),
		)),
	))
}

func (p *instrumentationPopout) hasFocus(gtx layout.Context) bool {
	_, heatRadios := p.heatMapType.Focused()
	_, animRadios := p.animationFormat.Focused()
	return heatRadios || animRadios || p.skipBackBy.isFocused(gtx) ||
		p.chkRecord.isFocused(gtx) || p.chkRepeatDetect.isFocused(gtx) || p.chkHeatMap.isFocused(gtx) ||
		p.btnRepeatReset.isFocused(gtx) || p.btnRepeatSave.isFocused(gtx) || p.btnRecordReset.isFocused(gtx) || p.btnSaveAnimation.isFocused(gtx) ||
		p.btnHeatMapReset.isFocused(gtx) || p.btnHeatMapReveal.isFocused(gtx) || p.btnHeatMapSave.isFocused(gtx)
}
