package widgets

import (
	"errors"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/patterns"
	"image"
	"image/color"
	"image/draw"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
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
	btnRemove     *button
	btnClear      *button
	error         error
	name          *input
	filename      *input
	origin        *input
	comment       *input
	// pattern identification...
	btnIdentify        *button
	chkWithPhases      *checkbox
	identifyStatus     atomic.Int32
	identifying        *patterns.Pattern
	identLibLen        int
	identifyError      error
	identifyHashCount  int
	identifyFound      string
	identifyFoundCount int
	identifyFoundClick widget.Clickable
	index              map[[32]byte][]string
	cachedPattern      *patterns.Pattern
	cachedImage        *image.NRGBA
}

func newCapturedPatternsPopout(p *menuPopup, c *Core) *capturedPatternsPopout {
	result := &capturedPatternsPopout{
		parent:        p,
		core:          c,
		previewMode:   &widget.Enum{Value: previewMetadata},
		btnSave:       newButton("Save"),
		btnRemove:     newButton("Remove"),
		btnClear:      newButton("Clear"),
		chkAddLibrary: newCheckBox("Add to library", true),
		btnIdentify:   newButton("Identify"),
		chkWithPhases: newCheckBox("With phases", false),
	}
	result.radioPreview = newRadioButton(result.previewMode, previewImage, "Preview")
	result.radioMetadata = newRadioButton(result.previewMode, previewMetadata, "Metadata")
	result.name = newInput("", 0, result.updateName)
	result.filename = newInput("", 0, result.updateFilename)
	result.origin = newInput("", 0, result.updateOrigin)
	result.comment = newInput("", 0, result.updateComment)
	result.comment.editor.SingleLine = false
	result.comment.editor.Submit = false
	result.chooser = newChooser[*patterns.Pattern](38,
		c.settings.CapturedPatterns,
		result.patternSelected,
		func(pattern *patterns.Pattern) string {
			return pattern.String()
		},
	)
	return result
}

func (p *capturedPatternsPopout) addCapturedPattern(pattern patterns.Pattern) {
	name := strconv.Itoa(len(p.core.settings.CapturedPatterns)+1) + " (" + time.Now().Format("2006-01-02 15-04-05") + ")"
	origin := p.core.settings.Originator
	if len(origin) == 0 {
		origin = "(your name)"
	}
	p.core.settings.CapturedPatterns = append(p.core.settings.CapturedPatterns,
		&patterns.Pattern{
			Name:        name,
			Width:       pattern.Width,
			Height:      pattern.Height,
			Cells:       slices.Clone(pattern.Cells),
			Comments:    []string{"Captured by GoGoL"},
			Origination: origin,
			Rule:        p.core.gridHolder.grid.Rule,
			Filename:    name + ".rle"})
	p.chooser.resetItems(p.core.settings.CapturedPatterns)
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

func (p *capturedPatternsPopout) layoutNoPatterns(gtx layout.Context) layout.Dimensions {
	const note = "Note: Multiple patterns can be captured in one edit session"
	minX := measureText(gtx, note).Size.X
	return popoutLayout(gtx, flexVertical(30,
		rigidFixedWidth(label("No patterns captured yet."), minX, layout.Center),
		rigidFixedWidth(flexVertical(0,
			rigidLabel("To capture patterns:", 0, 0, 0),
			rigid(inset(0, 0, 8, 0, flexVertical(0,
				rigid(textLabel("1. Start edit mode ("+modKeyName+"E)")),
				rigid(textLabel("2. Mark the pattern area\n\u2007  Hold down shift+arrow keys and then hit Return")),
				rigid(textLabel("3. Come back here to edit/save patterns")),
			))),
		), minX, 0),
		rigidFixedWidth(label(note), minX, layout.Center),
	))
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
			// remove from list...
			for i, v := range p.core.settings.CapturedPatterns {
				if v == pattern {
					p.core.settings.CapturedPatterns = append(p.core.settings.CapturedPatterns[:i], p.core.settings.CapturedPatterns[i+1:]...)
					break
				}
			}
			p.identLibLen = 0
			p.chooser.resetItems(p.core.settings.CapturedPatterns)
		}
	} else {
		p.error = err
	}
}

func (p *capturedPatternsPopout) layout(gtx layout.Context) layout.Dimensions {
	if len(p.core.settings.CapturedPatterns) == 0 {
		// no patterns captured yet
		return p.layoutNoPatterns(gtx)
	}
	m := measureText(gtx, "M")
	currentPattern := p.chooser.currentItem()
	if p.btnSave.Clicked(gtx) {
		if currentPattern != nil {
			p.savePattern(*currentPattern)
			currentPattern = p.chooser.currentItem()
		}
	}
	if p.btnRemove.Clicked(gtx) {
		if currentPattern != nil {
			for i, v := range p.core.settings.CapturedPatterns {
				if v == *currentPattern {
					p.core.settings.CapturedPatterns = append(p.core.settings.CapturedPatterns[:i], p.core.settings.CapturedPatterns[i+1:]...)
					break
				}
			}
			p.chooser.resetItems(p.core.settings.CapturedPatterns)
		}
	}
	if p.btnClear.Clicked(gtx) {
		p.core.settings.CapturedPatterns = nil
	}
	return popoutLayout(gtx, func(gtx layout.Context) layout.Dimensions {
		dims := flexVertical(10,
			rigid(p.chooser.layout),
			rigid(flexHorizontal(10,
				rigid(p.radioMetadata.Layout),
				rigid(p.radioPreview.Layout),
			)),
			rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.Y = int(float32(m.Size.Y) * 15.5)
				gtx.Constraints.Max.X = p.chooser.dims.Size.X
				return p.layoutPreview(gtx, p.chooser.dims.Size.X, m.Size.Y*15)
			}),
			rigid(func(gtx layout.Context) layout.Dimensions {
				switch {
				case currentPattern == nil:
					return layout.Dimensions{}
				case p.previewMode.Value == previewMetadata:
					return flexHorizontal(20,
						rigid(p.btnSave.Layout),
						rigid(p.chkAddLibrary.Layout),
						rigid(insetErrorLabel(p.error)),
						rigid(p.btnRemove.Layout),
						rigid(p.btnClear.Layout))(gtx)
				default:
					return p.layoutIdentifyControls(gtx, *currentPattern)
				}
			}),
		)(gtx)
		p.chooser.layoutDropdown(gtx)
		return dims
	})
}

func (p *capturedPatternsPopout) layoutPreview(gtx layout.Context, maxWd, maxHt int) layout.Dimensions {
	currentPattern := p.chooser.currentItem()
	switch {
	case currentPattern == nil:
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceAround}.Layout(gtx,
			rigid(func(gtx layout.Context) layout.Dimensions {
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

func (p *capturedPatternsPopout) layoutPreviewMetadata(pattern *patterns.Pattern, gtx layout.Context) layout.Dimensions {
	txtDim := measureText(gtx, "My")
	labelMax := measureMaxText(gtx, font.Bold, "Size: ", "Filename: ", "Origin: ", "Comment: ").Size.X
	return layout.Flex{Axis: layout.Vertical, Gap: 10, Spacing: layout.SpaceEnd}.Layout(gtx,
		rigid(flexHorizontal(20,
			rigidLabel("Name:", text.End, font.Bold, labelMax),
			flexed(p.name.layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Filename:", text.End, font.Bold, labelMax),
			flexed(p.filename.layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Origin:", text.End, font.Bold, labelMax),
			flexed(p.origin.layout),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Comment:", text.End, font.Bold, labelMax),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.Y = txtDim.Size.Y * 7
				gtx.Constraints.Max.Y = gtx.Constraints.Min.Y
				return p.comment.layout(gtx)
			}),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Size:", text.End, font.Bold, labelMax),
			flexed(label(strconv.Itoa(pattern.Width)+"w X "+strconv.Itoa(pattern.Height)+"h")),
		)),
		rigid(flexHorizontal(20,
			rigidLabel("Rule:", text.End, font.Bold, labelMax),
			flexed(label(pattern.Rule.Name())),
		)),
	)
}

func (p *capturedPatternsPopout) layoutPreviewImage(pattern *patterns.Pattern, gtx layout.Context, maxWd, maxHt int) layout.Dimensions {
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

const (
	identifyingNone int32 = iota
	identifyingIndexing
	identifyingPhases
	identifyingSearching
	identifyingNotFound
	identifyingFound
)

func (p *capturedPatternsPopout) layoutIdentifyControls(gtx layout.Context, pattern *patterns.Pattern) layout.Dimensions {
	status := p.identifyStatus.Load()
	p.chkWithPhases.Update(gtx)
	if status == identifyingNone || (p.identifying != pattern && (status == identifyingNotFound || status == identifyingFound)) {
		if p.btnIdentify.Clicked(gtx) {
			p.identify(pattern, p.chkWithPhases.Checked())
		}
		return flexHorizontal(20,
			rigid(p.btnIdentify.Layout),
			rigid(p.chkWithPhases.Layout))(gtx)
	} else if p.identifying != pattern {
		return insetLabel("(Identification currently busy)")(gtx)
	}
	switch status {
	case identifyingIndexing:
		return insetLabel("Indexing pattern library...")(gtx)
	case identifyingPhases:
		return insetLabel("Generating pattern phases...")(gtx)
	case identifyingSearching:
		return insetLabel("Searching pattern library...")(gtx)
	case identifyingNotFound:
		if p.btnIdentify.Clicked(gtx) {
			p.identify(pattern, p.chkWithPhases.Checked())
		}
		return flexHorizontal(20,
			rigid(p.btnIdentify.Layout),
			rigid(p.chkWithPhases.Layout),
			rigid(insetLabel("Not found (searched "+strconv.Itoa(p.identifyHashCount)+" hashes)")))(gtx)
	case identifyingFound:
		if p.btnIdentify.Clicked(gtx) {
			p.identify(pattern, p.chkWithPhases.Checked())
		}
		if p.identifyFoundClick.Clicked(gtx) {
			p.parent.showPattern(p.identifyFound)
		}
		return flexVertical(4,
			rigid(flexHorizontal(20,
				rigid(p.btnIdentify.Layout),
				rigid(p.chkWithPhases.Layout),
				rigid(insetLabel("Found "+strconv.Itoa(p.identifyFoundCount)+" (searched "+strconv.Itoa(p.identifyHashCount)+" hashes)")),
			)),
			rigid(linkLabel(&p.identifyFoundClick, p.identifyFound)),
		)(gtx)
	}
	return layout.Dimensions{}
}

func (p *capturedPatternsPopout) identify(pattern *patterns.Pattern, withPhases bool) {
	p.identifyStatus.Store(identifyingIndexing)
	p.identifyError = nil
	p.identifying = pattern
	window.Invalidate()
	go func() {
		searchPattern := pattern.Trimmed()
		if p.identLibLen != len(patterns.PatternLibrary) {
			p.index = make(map[[32]byte][]string, len(patterns.PatternLibrary))
			for k, v := range patterns.PatternLibrary {
				trimmed := v.Trimmed()
				h := trimmed.Hash()
				if _, ok := p.index[h]; ok {
					p.index[h] = append(p.index[h], k)
				} else {
					p.index[h] = []string{k}
				}
			}
			p.identLibLen = len(patterns.PatternLibrary)
		}
		p.identifyHashCount = -1
		searchHashes := make(map[[32]byte]struct{})
		for _, h := range searchPattern.AllHashes() {
			searchHashes[h] = struct{}{}
		}
		if withPhases {
			p.identifyStatus.Store(identifyingPhases)
			window.Invalidate()
			phases, _ := searchPattern.Phases(100, 0, 4)
			for _, phase := range phases {
				for _, h := range phase.AllHashes() {
					searchHashes[h] = struct{}{}
				}
			}
		}
		p.identifyHashCount = len(searchHashes)
		p.identifyStatus.Store(identifyingSearching)
		window.Invalidate()
		p.identifyFoundCount = 0
		for h := range searchHashes {
			if names, ok := p.index[h]; ok {
				for _, name := range names {
					if p.identifyFoundCount == 0 {
						p.identifyFound = name
					}
					p.identifyFoundCount++
				}
			}
		}
		if p.identifyFoundCount > 0 {
			p.identifyStatus.Store(identifyingFound)
		} else {
			p.identifyStatus.Store(identifyingNotFound)
		}
		window.Invalidate()
	}()
}

func (p *capturedPatternsPopout) hasFocus(gtx layout.Context) bool {
	_, radios := p.previewMode.Focused()
	return radios || p.chooser.isFocused(gtx) || p.btnSave.isFocused(gtx) || p.chkAddLibrary.isFocused(gtx) ||
		p.btnRemove.isFocused(gtx) || p.btnClear.isFocused(gtx) ||
		p.name.isFocused(gtx) || p.filename.isFocused(gtx) ||
		p.origin.isFocused(gtx) || p.comment.isFocused(gtx) ||
		p.btnIdentify.isFocused(gtx) || p.chkWithPhases.isFocused(gtx)
}

func (p *capturedPatternsPopout) reset() {
	// nothing to reset
}
