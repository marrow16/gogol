package widgets

import (
	"errors"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
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
	result := &instrumentationPopout{
		parent:           p,
		core:             c,
		chkRepeatDetect:  newCheckBox(c.theme, "Repeat Detect", c.instrumentRepeat != nil),
		btnRepeatReset:   newButton(c.theme, "Reset"),
		btnRepeatSave:    newButton(c.theme, "Save Report"),
		chkRecord:        newCheckBox(c.theme, "Record", c.instrumentRecord != nil),
		btnRecordReset:   newButton(c.theme, "Reset"),
		btnSaveAnimation: newButton(c.theme, "Save Animation"),
		chkHeatMap:       newCheckBox(c.theme, "Heat Mapping", c.instrumentHeatMap != nil),
		btnHeatMapReset:  newButton(c.theme, "Reset"),
		btnHeatMapReveal: newButton(c.theme, "Reveal"),
		btnHeatMapSave:   newButton(c.theme, "Save Image"),
		heatMapType:      &widget.Enum{Value: c.heatMapperType.String()},
	}
	result.radioActivity = newRadioButton(c.theme, result.heatMapType, activityHeatMapper.String(), "Activity")
	result.radioOccupancy = newRadioButton(c.theme, result.heatMapType, occupancyHeatMapper.String(), "Occupancy")
	result.radioBirths = newRadioButton(c.theme, result.heatMapType, birthsHeatMapper.String(), "Births")
	result.radioFreshness = newRadioButton(c.theme, result.heatMapType, freshnessHeatMapper.String(), "Freshness")
	result.radioPhaseParity = newRadioButton(c.theme, result.heatMapType, phaseParityHeatMapper.String(), "Phase Parity")
	result.radioAll = newRadioButton(c.theme, result.heatMapType, allHeatMapper.String(), "All")
	result.skipBackBy = newNumberInput[int](c.theme, 4, 1, 9999, 100, result.skipBackByChanged)
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
}

func (p *instrumentationPopout) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	p.update(gtx)
	width := measureText(gtx, theme, "When enabled, will stop stepping (and step ahead) extra").Size.X
	return layout.Inset{Left: 4, Right: 4, Top: 4, Bottom: 4}.
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
					return p.chkRepeatDetect.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !p.chkRepeatDetect.Checked() {
						gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
						desc := material.Label(theme, theme.TextSize, "When enabled, will stop stepping (and step ahead) when the grid repeats.\nInformation about repeats will be shown here.")
						desc.TextSize = desc.TextSize - 2
						desc.Font.Style = font.Italic
						return desc.Layout(gtx)
					} else {
						return p.layoutRepeat(gtx, theme)
					}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
					paint.FillShape(gtx.Ops, popupBorder,
						clip.Rect(image.Rect(0, 0, width, 1)).Op(),
					)
					return layout.Dimensions{Size: image.Point{X: width, Y: 1}}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
					return p.chkRecord.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !p.chkRecord.Checked() {
						gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
						desc := material.Label(theme, theme.TextSize, "Records every generation, enabling backward stepping and animation export.")
						desc.TextSize = desc.TextSize - 2
						desc.Font.Style = font.Italic
						return desc.Layout(gtx)
					} else if p.core.instrumentRecord != nil {
						return p.layoutRecord(gtx, theme)
					} else {
						return layout.Dimensions{}
					}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
					paint.FillShape(gtx.Ops, popupBorder,
						clip.Rect(image.Rect(0, 0, width, 1)).Op(),
					)
					return layout.Dimensions{Size: image.Point{X: width, Y: 1}}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
					return p.chkHeatMap.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !p.chkHeatMap.Checked() {
						gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
						desc := material.Label(theme, theme.TextSize, "Accumulates various types of grid activity for later display as a visual heat map.")
						desc.TextSize = desc.TextSize - 2
						desc.Font.Style = font.Italic
						return desc.Layout(gtx)
					} else if p.core.instrumentHeatMap != nil {
						return p.layoutHeatMap(gtx, theme)
					} else {
						return layout.Dimensions{}
					}
				}),
			)
		})
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
			if p.core.instrumentRecord.StepsCount() > 0 {
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

func (p *instrumentationPopout) layoutRepeat(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	labelMax := measureMaxText(gtx, theme, font.Bold, "Examined: ", "Found: ", "First: ", "Repeat: ", "Period: ").Size.X
	return layout.Inset{Left: 16, Bottom: 4}.
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(rightAlignedLabel(theme, "Examined:", labelMax)),
						layout.Rigid(label(theme, commas(strconv.FormatUint(p.core.instrumentRepeat.Steps, 10)))),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(rightAlignedLabel(theme, "Found:", labelMax)),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if p.core.instrumentRepeat.Found {
								return material.Label(theme, theme.TextSize, "Yes").Layout(gtx)
							}
							return material.Label(theme, theme.TextSize, "No").Layout(gtx)
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(rightAlignedLabel(theme, "First:", labelMax)),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(theme, theme.TextSize, "--")
							if p.core.instrumentRepeat.Found {
								lbl.Text = commas(strconv.FormatUint(p.core.instrumentRepeat.FirstStep, 10))
							}
							return lbl.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(rightAlignedLabel(theme, "Repeat:", labelMax)),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(theme, theme.TextSize, "--")
							if p.core.instrumentRepeat.Found {
								lbl.Text = commas(strconv.FormatUint(p.core.instrumentRepeat.RepeatStep, 10))
							}
							return lbl.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(rightAlignedLabel(theme, "Period:", labelMax)),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(theme, theme.TextSize, "--")
							if p.core.instrumentRepeat.Found {
								lbl.Text = commas(strconv.FormatUint(p.core.instrumentRepeat.Period, 10))
							}
							return lbl.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: 4, Bottom: 4}.Layout(gtx, p.btnRepeatReset.Layout)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: 4, Bottom: 4}.Layout(gtx, p.btnRepeatSave.Layout)
						}),
					)
				}),
			)
		})
}

func (p *instrumentationPopout) layoutRecord(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	labelMax := measureMaxText(gtx, theme, font.Bold, "Steps recorded: ", "Skip back by: ").Size.X
	return layout.Inset{Left: 16, Bottom: 4}.
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(rightAlignedLabel(theme, "Steps recorded:", labelMax)),
						layout.Rigid(label(theme, commas(strconv.Itoa(p.core.instrumentRecord.StepsCount())))),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(rightAlignedLabel(theme, "Skip back by:", labelMax)),
						layout.Flexed(1, p.skipBackBy.layout),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: 4, Bottom: 4}.Layout(gtx, p.btnRecordReset.Layout)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if !p.animationSaving {
								return layout.Inset{Top: 4, Bottom: 4}.Layout(gtx, p.btnSaveAnimation.Layout)
							} else {
								return layout.Dimensions{}
							}
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if p.animationSaving {
						return label(theme, "Saving animation - please wait...")(gtx)
					} else if p.animationResult != nil {
						if p.animationResult.err == nil {
							return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
								layout.Rigid(label(theme, "Saved to:")),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return material.Clickable(gtx, &p.linkAnimation, label(theme, filepath.Base(p.animationResult.filename)))
								}),
							)
						} else {
							return errorLabel(theme, p.animationResult.err)(gtx)
						}
					} else {
						return layout.Dimensions{}
					}
				}),
			)
		})
}

func (p *instrumentationPopout) saveAnimation() {
	filename, err := resolveSavePath(p.core.nowFilename("Grid", ".mp4"))
	if err != nil {
		p.animationResult = &animationResult{
			filename: filename,
			err:      err,
		}
		p.core.window.Invalidate()
		return
	}
	recorder := p.core.instrumentRecord
	go func() {
		ani := animator.NewAnimator(p.core.settings.CellSize, p.core.settings.CellAliveColor, p.core.settings.CellDeadColor, p.core.settings.CellBorderColor, p.core.settings.CellBorders)
		err := ani.Animate(filename, recorder)
		p.animationSaving = false
		p.animationResult = &animationResult{
			filename: filename,
			err:      err,
		}
		p.core.window.Invalidate()
	}()
}

func (p *instrumentationPopout) layoutHeatMap(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	labelMax := measureMaxText(gtx, theme, font.Bold, "Steps: ", "Type: ", "Maximum: ").Size.X
	return layout.Inset{Left: 16, Bottom: 4}.
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(rightAlignedLabel(theme, "Type:", labelMax)),
						layout.Rigid(p.radioActivity.Layout),
						layout.Rigid(p.radioOccupancy.Layout),
						layout.Rigid(p.radioBirths.Layout),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = labelMax
							return layout.Dimensions{Size: image.Point{X: labelMax}}
						}),
						layout.Rigid(p.radioFreshness.Layout),
						layout.Rigid(p.radioPhaseParity.Layout),
						layout.Rigid(p.radioAll.Layout),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(rightAlignedLabel(theme, "Maximum:", labelMax)),
						layout.Rigid(label(theme, commas(strconv.FormatUint(p.core.instrumentHeatMap.Maximum(), 10)))),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(rightAlignedLabel(theme, "Steps:", labelMax)),
						layout.Rigid(label(theme, commas(strconv.FormatUint(p.core.instrumentHeatMap.StepsCount(), 10)))),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: 4, Bottom: 4}.Layout(gtx, p.btnHeatMapReset.Layout)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: 4, Bottom: 4}.Layout(gtx, p.btnHeatMapReveal.Layout)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: 4, Bottom: 4}.Layout(gtx, p.btnHeatMapSave.Layout)
						}),
					)
				}),
			)
		})
}

func (p *instrumentationPopout) hasFocus(gtx layout.Context) bool {
	_, radios := p.heatMapType.Focused()
	return radios || p.skipBackBy.isFocused(gtx) ||
		p.chkRecord.isFocused(gtx) || p.chkRepeatDetect.isFocused(gtx) || p.chkHeatMap.isFocused(gtx) ||
		p.btnRepeatReset.isFocused(gtx) || p.btnRepeatSave.isFocused(gtx) || p.btnRecordReset.isFocused(gtx) || p.btnSaveAnimation.isFocused(gtx) ||
		p.btnHeatMapReset.isFocused(gtx) || p.btnHeatMapReveal.isFocused(gtx) || p.btnHeatMapSave.isFocused(gtx)
}
