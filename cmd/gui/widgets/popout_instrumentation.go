package widgets

import (
	"errors"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/animator"
	"image"
	"path/filepath"
	"strconv"
	"time"
)

type instrumentationPopout struct {
	parent           *menuPopup
	core             *Core
	repeatDetect     *widget.Bool
	btnRepeatReset   widget.Clickable
	record           *widget.Bool
	skipBackBy       *numberInput[int]
	btnSaveAnimation widget.Clickable
	animationSaving  bool
	animationResult  *animationResult
	linkAnimation    widget.Clickable
}

type animationResult struct {
	filename string
	err      error
}

func newInstrumentationPopout(p *menuPopup, c *Core) *instrumentationPopout {
	result := &instrumentationPopout{
		parent:       p,
		core:         c,
		repeatDetect: &widget.Bool{Value: c.instrumentRepeat != nil},
		record:       &widget.Bool{Value: c.instrumentRecord != nil},
	}
	result.skipBackBy = newNumberInput[int](c.theme, 4, 1, 9999, 100, result.skipBackByChanged)
	return result
}

func (p *instrumentationPopout) skipBackByChanged(n int) {
	if n > 0 {
		p.core.settings.SkipBackBy = n
	}
}

func (p *instrumentationPopout) reset() {
	p.repeatDetect.Value = p.core.instrumentRepeat != nil
	p.record.Value = p.core.instrumentRecord != nil
	p.skipBackBy.setValue(p.core.settings.SkipBackBy)
}

func (p *instrumentationPopout) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	if p.repeatDetect.Update(gtx) {
		p.core.setInstrumentationRepeat(p.repeatDetect.Value)
	}
	if p.record.Update(gtx) {
		p.core.setInstrumentationRecord(p.record.Value)
	}
	if p.btnRepeatReset.Clicked(gtx) {
		p.core.setInstrumentationRepeat(true)
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
	width := measureText(gtx, theme, "When enabled, will stop stepping (and step ahead)").Size.X
	return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8), Top: unit.Dp(8), Bottom: unit.Dp(4)}.
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
					chkBox := material.CheckBox(p.core.theme, p.repeatDetect, "Repeat detect")
					chkBox.TextSize = unit.Sp(16)
					chkBox.Size = 18
					return chkBox.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !p.repeatDetect.Value {
						gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
						desc := material.Body1(theme, "When enabled, will stop stepping (and step ahead) when the grid repeats.\nInformation about repeats will be shown here.")
						desc.TextSize = desc.TextSize - 2
						desc.Font.Style = font.Italic
						return desc.Layout(gtx)
					} else {
						return p.layoutRepeat(gtx, theme)
					}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
					chkBox := material.CheckBox(p.core.theme, p.record, "Record")
					chkBox.TextSize = unit.Sp(16)
					chkBox.Size = 18
					return chkBox.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !p.record.Value {
						gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
						desc := material.Body1(theme, "Records every generation, enabling backward stepping and animation export.")
						desc.TextSize = desc.TextSize - 2
						desc.Font.Style = font.Italic
						return desc.Layout(gtx)
					} else if p.core.instrumentRecord != nil {
						return p.layoutRecord(gtx, theme)
					} else {
						return layout.Dimensions{}
					}
				}),
			)
		})
}

func (p *instrumentationPopout) layoutRepeat(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	labelMax := measureMaxText(gtx, theme, font.Bold, "Found: ", "First: ", "Repeat: ", "Period: ").Size.X
	return layout.Inset{Left: 28, Bottom: 4}.
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = labelMax
							lbl := rightLabel(p.core.theme, "Found:")
							lbl.Font.Weight = font.Bold
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body1(theme, "No")
							if p.core.instrumentRepeat.Found {
								lbl.Text = "Yes"
							}
							return lbl.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = labelMax
							lbl := rightLabel(p.core.theme, "First:")
							lbl.Font.Weight = font.Bold
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body1(theme, "--")
							if p.core.instrumentRepeat.Found {
								lbl.Text = commas(strconv.FormatUint(p.core.instrumentRepeat.FirstStep, 10))
							}
							return lbl.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = labelMax
							lbl := rightLabel(p.core.theme, "Repeat:")
							lbl.Font.Weight = font.Bold
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body1(theme, "--")
							if p.core.instrumentRepeat.Found {
								lbl.Text = commas(strconv.FormatUint(p.core.instrumentRepeat.RepeatStep, 10))
							}
							return lbl.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = labelMax
							lbl := rightLabel(p.core.theme, "Period:")
							lbl.Font.Weight = font.Bold
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body1(theme, "--")
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
							gtx.Constraints.Min.X = labelMax
							return layout.Dimensions{Size: image.Point{X: gtx.Constraints.Min.X}}
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(p.core.theme, &p.btnRepeatReset, "Reset")
							btn.Inset = layout.Inset{Bottom: 2, Left: 3, Right: 3}
							btn.TextSize = unit.Sp(16)
							return btn.Layout(gtx)
						}),
					)
				}),
			)
		})
}

func (p *instrumentationPopout) layoutRecord(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	labelMax := measureMaxText(gtx, theme, font.Bold, "Steps recorded: ", "Skip back by: ").Size.X
	return layout.Inset{Left: 28, Bottom: 4}.
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = labelMax
							lbl := rightLabel(p.core.theme, "Steps recorded:")
							lbl.Font.Weight = font.Bold
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body1(theme, commas(strconv.Itoa(p.core.instrumentRecord.StepsCount())))
							return lbl.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = labelMax
							lbl := rightLabel(p.core.theme, "Skip back by:")
							lbl.Font.Weight = font.Bold
							return lbl.Layout(gtx)
						}),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return p.skipBackBy.layout(gtx)
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !p.animationSaving {
						return layout.Inset{Top: 4, Bottom: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(p.core.theme, &p.btnSaveAnimation, "Save Animation")
							btn.Inset = layout.Inset{Bottom: 2, Left: 3, Right: 3}
							btn.TextSize = unit.Sp(16)
							return btn.Layout(gtx)
						})
					} else {
						return layout.Dimensions{}
					}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if p.animationSaving {
						lbl := material.Body1(theme, "Saving animation - please wait...")
						lbl.MaxLines = 1
						return lbl.Layout(gtx)
					} else if p.animationResult != nil {
						if p.animationResult.err == nil {
							return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return material.Body1(theme, "Saved to:").Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return material.Clickable(gtx, &p.linkAnimation, func(gtx layout.Context) layout.Dimensions {
										lbl := material.Body1(theme, filepath.Base(p.animationResult.filename))
										lbl.MaxLines = 1
										return lbl.Layout(gtx)
									})
								}),
							)
						} else {
							lbl := material.Body1(theme, p.animationResult.err.Error())
							lbl.MaxLines = 1
							lbl.Color = errorColor
							return lbl.Layout(gtx)
						}
					} else {
						return layout.Dimensions{}
					}
				}),
			)
		})
}

func (p *instrumentationPopout) saveAnimation() {
	filename, err := resolveSavePath("Grid " + time.Now().Format("2006-01-02 15:04:05") + ".mp4")
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

func (p *instrumentationPopout) hasFocus(gtx layout.Context) bool {
	return p.skipBackBy.isFocused(gtx) || gtx.Focused(p.record) || gtx.Focused(p.repeatDetect)
}
