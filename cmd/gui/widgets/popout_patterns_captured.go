package widgets

import (
	"errors"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/patterns"
	"image"
	"image/draw"
	"slices"
	"strconv"
	"strings"
	"time"
)

type capturedPatternsPopout struct {
	parent        *menuPopup
	core          *Core
	chooser       *chooser[*patterns.Pattern]
	previewMode   *widget.Enum
	radioPreview  *radioButton
	radioMetadata *radioButton
	btnSave       *button
	chkAddLibrary *checkbox
	error         error
	patterns      []*patterns.Pattern
	name          *input
	filename      *input
	origin        *input
	comment       *input
}

func newCapturedPatternsPopout(p *menuPopup, c *Core) *capturedPatternsPopout {
	result := &capturedPatternsPopout{
		parent:        p,
		core:          c,
		previewMode:   &widget.Enum{Value: previewMetadata},
		btnSave:       newButton(c.theme, "Save"),
		chkAddLibrary: newCheckBox(c.theme, "Add to library", true),
	}
	result.radioPreview = newRadioButton(c.theme, result.previewMode, previewImage, "Preview")
	result.radioMetadata = newRadioButton(c.theme, result.previewMode, previewMetadata, "Metadata")
	result.name = newInput(c.theme, "", 0, result.updateName)
	result.filename = newInput(c.theme, "", 0, result.updateFilename)
	result.origin = newInput(c.theme, "", 0, result.updateOrigin)
	result.comment = newInput(c.theme, "", 0, result.updateComment)
	result.comment.editor.SingleLine = false
	result.comment.editor.Submit = false
	result.chooser = newChooser[*patterns.Pattern](c.theme, 38,
		result.patterns,
		result.patternSelected,
		func(pattern *patterns.Pattern) string {
			return pattern.String()
		},
	)
	return result
}

func (p *capturedPatternsPopout) addCapturedPattern(pattern patterns.Pattern) {
	name := time.Now().Format("2006-01-02 15:04:05")
	origin := p.core.settings.Originator
	if len(origin) == 0 {
		origin = "(your name)"
	}
	p.patterns = append(p.patterns, &patterns.Pattern{
		Name:        name,
		Width:       pattern.Width,
		Height:      pattern.Height,
		Cells:       slices.Clone(pattern.Cells),
		Comments:    []string{"Captured by GoGoL"},
		Origination: origin,
		Rule:        p.core.gridHolder.grid.Rule,
		Filename:    name + ".rle",
	})
	p.chooser.resetItems(p.patterns)
	p.chooser.setText(name)
}

func (p *capturedPatternsPopout) updateName(text string) {
	if c := p.chooser.currentItem(); c != nil {
		patt := *c
		if strings.EqualFold(patt.Name+".rle", patt.Filename) {
			patt.Filename = text + ".rle"
			p.filename.setText(patt.Filename)
		}
		patt.Name = text
		if p.chooser.editor.Text() != text {
			if text == "" {
				p.chooser.setText(patt.Filename)
			} else {
				p.chooser.setText(patt.Name)
			}
		}
	}
}

func (p *capturedPatternsPopout) updateFilename(text string) {
	p.error = nil
	if c := p.chooser.currentItem(); c != nil {
		patt := *c
		patt.Filename = text
	}
}

func (p *capturedPatternsPopout) updateOrigin(text string) {
	if c := p.chooser.currentItem(); c != nil {
		patt := *c
		patt.Origination = text
	}
}

func (p *capturedPatternsPopout) updateComment(text string) {
	if c := p.chooser.currentItem(); c != nil {
		patt := *c
		patt.Comments = strings.Split(text, "\n")
	}
}

func (p *capturedPatternsPopout) patternSelected(pattern **patterns.Pattern) {
	p.error = nil
	if pattern != nil {
		patt := *pattern
		p.name.setText(patt.Name)
		p.filename.setText(patt.Filename)
		p.origin.setText(patt.Origination)
		p.comment.setText(strings.Join(patt.Comments, "\n"))
	}
}

func (p *capturedPatternsPopout) layoutNoPatterns(gtx layout.Context, theme *material.Theme, minX int) layout.Dimensions {
	return layout.Inset{
		Left: unit.Dp(8), Right: unit.Dp(8),
		Top: unit.Dp(8), Bottom: unit.Dp(4),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Gap: 30, Spacing: layout.SpaceSides}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = minX
				gtx.Constraints.Max.X = minX
				lbl := material.Label(theme, theme.TextSize, "No patterns captured yet.")
				lbl.MaxLines = 1
				lbl.Alignment = text.Middle
				return lbl.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: 8, Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = minX
							gtx.Constraints.Max.X = minX
							lbl := material.Label(theme, theme.TextSize, "To capture patterns:")
							lbl.Alignment = text.Middle
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = minX
							return material.Label(theme, theme.TextSize, "1. Start edit mode ("+modKeyName+"E)").Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = minX
							return material.Label(theme, theme.TextSize, "2. Mark the pattern area\n\u2007  Hold down shift+arrow keys and then hit Return").Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = minX
							return material.Label(theme, theme.TextSize, "3. Come back here to edit/save patterns").Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = minX
								return material.Label(theme, theme.TextSize, "Note: Multiple patterns can be captured in one edit session").Layout(gtx)
							})
						}),
					)
				})
			}),
		)
	})
}

func (p *capturedPatternsPopout) savePattern(pattern *patterns.Pattern) {
	if len(pattern.Filename) == 0 {
		p.error = errors.New("No filename specified")
		return
	}
	p.error = nil
	if f, err := saveFile(pattern.Filename, false); err == nil {
		defer func() {
			_ = f.Close()
		}()
		p.core.settings.Originator = pattern.Origination
		if p.error = patterns.PatternRleEncode(*pattern, f); p.error == nil && p.chkAddLibrary.Checked() {
			patterns.PatternLibrary[pattern.Name] = *pattern
			p.core.settings.AddPattern(pattern.Filename)
		}
	} else {
		p.error = err
	}
}

func (p *capturedPatternsPopout) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	chd := measureText(gtx, p.core.theme, "M")
	gtx.Constraints.Min.Y = chd.Size.Y * 20
	if len(p.patterns) == 0 {
		// no patterns captured yet
		gtx.Constraints.Min.Y = chd.Size.Y * 10
		gtx.Constraints.Min.X = chd.Size.X * 30
		return p.layoutNoPatterns(gtx, theme, gtx.Constraints.Min.X)
	}
	if p.btnSave.Clicked(gtx) {
		if sp := p.chooser.currentItem(); sp != nil {
			p.savePattern(*sp)
		}
	}
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
					layout.Rigid(p.radioMetadata.Layout),
					layout.Rigid(p.radioPreview.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{
							Size:     image.Point{X: gtx.Dp(100)},
							Baseline: 0,
						}
					}),
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
						layout.Rigid(p.btnSave.Layout),
						layout.Rigid(p.chkAddLibrary.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if p.error != nil {
								return errorLabel(theme, p.error)(gtx)
							}
							return layout.Dimensions{}
						}),
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

func (p *capturedPatternsPopout) layoutPreview(gtx layout.Context, theme *material.Theme, maxWd, maxHt int) layout.Dimensions {
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

func (p *capturedPatternsPopout) layoutPreviewMetadata(pattern *patterns.Pattern, gtx layout.Context, theme *material.Theme) layout.Dimensions {
	txtDim := measureText(gtx, theme, "My")
	labelMax := measureMaxText(gtx, theme, font.Bold, "Size: ", "Filename: ", "Origin: ", "Comment: ").Size.X
	return layout.Flex{Axis: layout.Vertical, Gap: 10, Spacing: layout.SpaceEnd}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
				layout.Rigid(rightAlignedBoldLabel(theme, "Name:", labelMax)),
				layout.Flexed(1, p.name.layout),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
				layout.Rigid(rightAlignedBoldLabel(theme, "Filename:", labelMax)),
				layout.Flexed(1, p.filename.layout),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
				layout.Rigid(rightAlignedBoldLabel(theme, "Origin:", labelMax)),
				layout.Flexed(1, p.origin.layout),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Gap: 20}.Layout(gtx,
				layout.Rigid(rightAlignedBoldLabel(theme, "Comment:", labelMax)),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.Y = txtDim.Size.Y * 7
					gtx.Constraints.Max.Y = gtx.Constraints.Min.Y
					return p.comment.layout(gtx)
				}),
			)
		}),
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
	)
}

func (p *capturedPatternsPopout) layoutPreviewImage(pattern *patterns.Pattern, gtx layout.Context, theme *material.Theme, maxWd, maxHt int) layout.Dimensions {
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

func (p *capturedPatternsPopout) hasFocus(gtx layout.Context) bool {
	return p.chooser.isFocused(gtx) ||
		p.name.isFocused(gtx) || p.filename.isFocused(gtx) ||
		p.origin.isFocused(gtx) || p.comment.isFocused(gtx)
}

func (p *capturedPatternsPopout) reset() {
	// nothing to reset
}
