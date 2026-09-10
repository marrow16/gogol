package widgets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/cmd/gui/icons"
	"github.com/marrow16/gogol/imaging"
	"github.com/marrow16/gogol/patterns"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	_ "image/png"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
)

func newFileFinder(core *Core) *fileFinder {
	result := &fileFinder{
		core:              core,
		btnOpen:           newButton("Open"),
		btnOpenDisabled:   newButton("Open"),
		btnCancel:         newButton("Cancel"),
		chkOnlySelectable: newCheckBox("Only show selectable files", false),
	}
	result.btnOpenDisabled.style.Background = color.NRGBA{R: 160, G: 160, B: 160, A: 255}
	result.list = newListControl[*dirEntry](nil, false).
		rowRenderer(result.layoutEntry).
		onItemSelect(result.selectEntry).
		onItemNavigate(result.navigateEntry).
		onIsSelected(func(index int, entry *dirEntry) bool {
			return entry == result.result
		})
	result.list.selectedBg = color.NRGBA{R: 220, G: 220, B: 220, A: 255}
	result.list.focusedBg = color.NRGBA{R: 220, G: 220, B: 220, A: 255}
	result.keyFilters = []event.Filter{
		key.Filter{Name: key.NameEscape},
		key.Filter{Name: key.NameF1},
		key.Filter{Name: key.NameDeleteBackward},
		key.Filter{Name: key.NameLeftArrow},
		key.Filter{Name: key.NameRightArrow},
	}
	for k := 'A'; k <= 'Z'; k++ {
		result.keyFilters = append(result.keyFilters, key.Filter{Name: key.Name(k)})
	}
	for k := '0'; k <= '9'; k++ {
		result.keyFilters = append(result.keyFilters, key.Filter{Name: key.Name(k)})
	}
	return result
}

func isHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}

type fileFinder struct {
	core              *Core
	showing           bool
	allowExts         map[string]struct{}
	allowDir          bool
	result            *dirEntry
	onSelect          func(path string)
	title             string
	btnOpen           *button
	btnOpenDisabled   *button
	btnCancel         *button
	chkOnlySelectable *checkbox
	keyFilters        []event.Filter

	checkCurrent bool
	currentDir   *dirEntry
	list         *listControl[*dirEntry]
	scannedDir   *dirEntry
	currentPath  []*dirEntry
	error        error
}

func (f *fileFinder) selectEntry(entry *dirEntry, keyboard bool) {
	if !entry.isFile && (keyboard || f.result == entry) {
		f.result = nil
		f.currentPath = append(f.currentPath, entry)
		f.currentDir = entry
		window.Invalidate()
		return
	}
	if f.result != entry {
		window.Invalidate()
	}
	f.result = entry
}

func (f *fileFinder) navigateEntry(entry *dirEntry) {
	if f.result != entry {
		window.Invalidate()
	}
	f.result = entry
}

func (f *fileFinder) layoutEntry(gtx layout.Context, i int, entry *dirEntry) layout.Dimensions {
	allowed := !entry.isFile || f.canSelect(entry)
	return layout.Flex{Axis: layout.Horizontal, Gap: 6}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				ht := gtx.Sp(theme.TextSize) + gtx.Dp(6)
				var img image.Image
				if !entry.isFile {
					img = icons.Folder
				} else {
					ext := strings.ToLower(filepath.Ext(entry.path))
					switch ext {
					case ".rle":
						img = icons.FileRLE
					case ".jpg", ".jpeg", ".png", ".gif":
						img = icons.FileIMG
					case ".json":
						img = icons.FileJSON
					}
				}
				if img != nil {
					r := img.Bounds()
					stack := op.Affine(
						f32.Affine2D{}.Scale(
							f32.Point{},
							f32.Point{X: float32(ht) / float32(r.Dx()), Y: float32(ht) / float32(r.Dy())},
						),
					).Push(gtx.Ops)
					paint.NewImageOp(img).Add(gtx.Ops)
					paint.PaintOp{}.Add(gtx.Ops)
					stack.Pop()
					if !allowed {
						defer clip.Rect{Max: image.Point{X: ht, Y: ht}}.Push(gtx.Ops).Pop()
						paint.ColorOp{Color: color.NRGBA{R: 255, G: 255, B: 255, A: 160}}.Add(gtx.Ops)
						paint.PaintOp{}.Add(gtx.Ops)
					}
				}
				return layout.Dimensions{Size: image.Point{X: ht, Y: ht}}
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(theme, theme.TextSize, entry.name)
			lbl.MaxLines = 1
			if !allowed {
				lbl.Color = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
			}
			return lbl.Layout(gtx)
		}),
	)
}

func (f *fileFinder) canSelect(entry *dirEntry) bool {
	if entry == nil {
		return false
	}
	if entry.isFile {
		return f.canSelectPath(entry.path)
	}
	return f.allowDir
}

func (f *fileFinder) canSelectPath(path string) bool {
	if len(f.allowExts) > 0 {
		ext := strings.ToLower(filepath.Ext(path))
		_, ok := f.allowExts[ext]
		return ok
	}
	return true
}

func (f *fileFinder) layout(gtx layout.Context) {
	f.setCurrentDir()
	gtx.Execute(key.FocusCmd{Tag: &f.list.tag})
	layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		f.header(gtx),
		f.breadcrumbs(gtx),
		f.body(gtx),
		f.footer(gtx),
	)
}

func (f *fileFinder) header(gtx layout.Context) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		dims := layout.Inset{Top: 8, Left: 8, Bottom: 8, Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(theme, (theme.TextSize*5)/4, f.title)
			lbl.Font.Weight = font.Bold
			lbl.MaxLines = 1
			return lbl.Layout(gtx)
		})
		border(gtx, dims, false, false, true, false)
		return dims
	})
}

func (f *fileFinder) breadcrumbs(gtx layout.Context) layout.FlexChild {
	for i, d := range f.currentPath {
		if d.clickable.Clicked(gtx) {
			f.currentPath = f.currentPath[:i+1]
			f.currentDir = f.currentPath[len(f.currentPath)-1]
			f.result = nil
			window.Invalidate()
		}
	}
	items := make([]layout.FlexChild, 0, len(f.currentPath)*2)
	items = append(items, rigid(func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(theme, theme.TextSize, "Path: ")
		lbl.Font.Weight = font.Bold
		lbl.MaxLines = 1
		return lbl.Layout(gtx)
	}))
	sep := label("/")
	last := len(f.currentPath) - 1
	for i, item := range f.currentPath {
		if i > 0 {
			items = append(items, layout.Rigid(sep))
		}
		switch {
		case i == last:
			items = append(items, rigid(label(item.name)))
		case i == 0 || item.path == "/" || item.path == "\\":
			items = append(items, rigid(linkLabel(item.clickable, "[root]")))
		default:
			items = append(items, rigid(linkLabel(item.clickable, item.name)))
		}
	}
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		dims := layout.Inset{Top: 8, Left: 8, Bottom: 8, Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Gap: 6}.Layout(gtx, items...)
		})
		border(gtx, dims, false, false, true, false)
		return dims
	})
}

func (f *fileFinder) body(gtx layout.Context) layout.FlexChild {
	return layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Gap: 8}.Layout(gtx,
			f.filesList(gtx),
			f.preview(gtx),
		)
	})
}

func (f *fileFinder) filesList(gtx layout.Context) layout.FlexChild {
	var backTo *dirEntry
	for {
		ev, ok := gtx.Event(f.keyFilters...)
		if !ok {
			break
		}
		if kev, ok := ev.(key.Event); ok && kev.State == key.Press {
			switch kev.Name {
			case key.NameEscape:
				f.showing = false
				window.Invalidate()
			case key.NameF1:
				f.core.showHelp(-1)
			case key.NameDeleteBackward, key.NameLeftArrow:
				if len(f.currentPath) > 1 {
					backTo = f.currentDir
					f.result = nil
					f.currentPath = f.currentPath[:len(f.currentPath)-1]
					f.currentDir = f.currentPath[len(f.currentPath)-1]
					window.Invalidate()
				}
			case key.NameRightArrow:
				if f.list.selectedIndex >= 0 && f.list.selectedIndex < len(f.list.items) && !f.list.items[f.list.selectedIndex].isFile {
					f.result = nil
					f.currentPath = append(f.currentPath, f.list.items[f.list.selectedIndex])
					f.currentDir = f.list.items[f.list.selectedIndex]
					window.Invalidate()
				}
			default:
				f.handleNavKeys(strings.ToUpper(string(kev.Name)))
			}
		}
	}
	if f.currentDir != f.scannedDir {
		f.scannedDir = f.currentDir
		dirs := make([]*dirEntry, 0)
		files := make([]*dirEntry, 0)
		entries, _ := os.ReadDir(f.scannedDir.path)
		onlySelectable := f.chkOnlySelectable.Checked()
		for _, entry := range entries {
			name := entry.Name()
			path := filepath.Join(f.scannedDir.path, name)
			if !isHidden(name) {
				if entry.IsDir() {
					dirs = append(dirs, &dirEntry{
						path:      path,
						name:      name,
						clickable: new(widget.Clickable),
					})
				} else if !onlySelectable || f.canSelectPath(path) {
					files = append(files, &dirEntry{
						path:   path,
						name:   name,
						isFile: true,
					})
				}
			}
		}
		slices.SortStableFunc(dirs, func(a, b *dirEntry) int {
			return strings.Compare(strings.ToLower(a.name), strings.ToLower(b.name))
		})
		slices.SortStableFunc(files, func(a, b *dirEntry) int {
			return strings.Compare(strings.ToLower(a.name), strings.ToLower(b.name))
		})
		items := append(dirs, files...)
		f.list.resetItems(items)
		if backTo != nil {
			for i, e := range items {
				if e.isFile {
					break
				}
				if e.path == backTo.path {
					f.list.selectedIndex = i
					f.result = e
					f.list.list.ScrollTo(f.list.selectedIndex)
					break
				}
			}
		}
		if f.result == nil && len(items) > 0 {
			f.result = items[0]
			f.list.list.ScrollTo(f.list.selectedIndex)
		}
	}
	return layout.Flexed(1.5, func(gtx layout.Context) layout.Dimensions {
		dims := layout.Inset{Top: 4, Left: 4, Bottom: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return f.list.Layout(gtx)
		})
		border(gtx, dims, false, false, false, true)
		return dims
	})
}

func (f *fileFinder) handleNavKeys(k string) {
	idx := f.list.selectedIndex
	if idx >= 0 && idx < len(f.list.items) {
		curr := f.list.items[idx]
		if !curr.isFile {
			if strings.HasPrefix(strings.ToUpper(curr.name), k) && idx+1 < len(f.list.items) &&
				strings.HasPrefix(strings.ToUpper(f.list.items[idx+1].name), k) {
				f.list.selectedIndex++
				f.list.list.ScrollTo(f.list.selectedIndex)
				f.navigateEntry(f.list.items[f.list.selectedIndex])
			} else {
				for i, e := range f.list.items {
					if strings.HasPrefix(strings.ToUpper(e.name), k) {
						f.list.selectedIndex = i
						f.list.list.ScrollTo(f.list.selectedIndex)
						f.navigateEntry(f.list.items[f.list.selectedIndex])
						break
					}
				}
			}
		} else if strings.HasPrefix(strings.ToUpper(curr.name), k) && idx+1 < len(f.list.items) &&
			strings.HasPrefix(strings.ToUpper(f.list.items[idx+1].name), k) {
			f.list.selectedIndex++
			f.list.list.ScrollTo(f.list.selectedIndex)
			f.navigateEntry(f.list.items[f.list.selectedIndex])
		} else {
			for i, e := range f.list.items {
				if strings.HasPrefix(strings.ToUpper(e.name), k) {
					f.list.selectedIndex = i
					f.list.list.ScrollTo(f.list.selectedIndex)
					f.navigateEntry(f.list.items[f.list.selectedIndex])
					break
				}
			}
		}
	}
}

func (f *fileFinder) preview(gtx layout.Context) layout.FlexChild {
	return layout.Flexed(3.5, func(gtx layout.Context) layout.Dimensions {
		dims := layout.Inset{Top: 4, Left: 4, Bottom: 4, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return f.entryPreview(f.result)(gtx)
		})
		border(gtx, dims, false, true, false, false)
		return dims
	})
}

func (f *fileFinder) footer(gtx layout.Context) layout.FlexChild {
	if f.btnOpen.Clicked(gtx) {
		if f.onSelect != nil && f.result != nil {
			f.onSelect(f.result.path)
		}
		f.showing = false
		window.Invalidate()
	}
	if f.btnCancel.Clicked(gtx) {
		f.showing = false
		window.Invalidate()
	}
	if f.chkOnlySelectable.Update(gtx) {
		f.scannedDir = nil
		f.result = nil
		window.Invalidate()
	}
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		dims := layout.Inset{Top: 8, Left: 8, Bottom: 8, Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
				layout.Rigid(f.chkOnlySelectable.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return flexHorizontal(20,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if f.canSelect(f.result) {
								return f.btnOpen.Layout(gtx)
							}
							return f.btnOpenDisabled.Layout(gtx)
						}),
						rigid(f.btnCancel.Layout),
					)(gtx)
				}),
			)
		})
		border(gtx, dims, true, false, false, false)
		return dims
	})
}

func (f *fileFinder) setCurrentDir() {
	if f.currentDir != nil && f.checkCurrent {
		if i, err := os.Stat(f.currentDir.path); err != nil || !i.IsDir() {
			f.currentDir = nil
		} else {
			f.checkCurrent = false
		}
	}
	if f.currentDir == nil {
		home, err := os.UserHomeDir()
		if err != nil {
			f.error = err
			return
		}
		if i, err := os.Stat(filepath.Join(home, "Documents")); err == nil && i.IsDir() {
			home = filepath.Join(home, "Documents")
			if i, err = os.Stat(filepath.Join(home, "GoGoL")); err == nil && i.IsDir() {
				home = filepath.Join(home, "GoGoL")
			}
		}
		info, err := os.Stat(home)
		if err != nil {
			f.error = err
			return
		}
		f.checkCurrent = false
		f.currentDir = &dirEntry{
			path:      home,
			name:      info.Name(),
			clickable: new(widget.Clickable),
		}
		f.currentPath = []*dirEntry{f.currentDir}
		path := home
		for {
			parent := filepath.Dir(path)
			if parent == path {
				break
			}
			if info, f.error = os.Stat(parent); f.error != nil {
				return
			}
			f.currentPath = append(f.currentPath, &dirEntry{
				path:      parent,
				name:      info.Name(),
				clickable: new(widget.Clickable),
			})
			path = parent
		}
		slices.Reverse(f.currentPath)
	}
}

func (f *fileFinder) entryPreview(entry *dirEntry) layout.Widget {
	if entry == nil {
		return func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{}
		}
	}
	entry.mutex.Lock()
	if !entry.previewLoading && entry.preview != nil {
		entry.mutex.Unlock()
		return entry.preview
	}
	entry.previewLoading = true
	entry.mutex.Unlock()
	go func(ff *fileFinder, allowExts map[string]struct{}) {
		entry.buildPreview(ff.allowExts)
		if ff.result == entry {
			window.Invalidate()
		}
	}(f, maps.Clone(f.allowExts))
	return func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(theme, theme.TextSize, "Preview loading...")
		lbl.Font.Style = font.Italic
		lbl.Color = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
		return lbl.Layout(gtx)
	}
}

type dirEntry struct {
	path      string
	name      string
	isFile    bool
	clickable *widget.Clickable

	mutex          sync.Mutex
	previewLoading bool
	preview        layout.Widget
	previewExtra   [][2]string
	previewImage   image.Image
	previewError   error
	previewScrollV *widget.List
	previewScrollH *widget.List
}

func (e *dirEntry) buildPreview(allowExts map[string]struct{}) {
	var preview layout.Widget
	var img image.Image
	var extra [][2]string
	var err error
	if !e.isFile {
		preview, img, err = e.buildPreviewDir(allowExts)
	} else {
		ext := strings.ToLower(filepath.Ext(e.path))
		switch ext {
		case ".rle":
			preview, extra, img, err = e.buildPreviewRle()
		case ".json":
			preview, extra, err = e.buildPreviewJson()
		case ".png", ".jpg", ".jpeg", ".gif":
			preview, extra, img, err = e.buildPreviewImg()
		default:
			preview, err = e.buildPreviewCommon()
		}
	}
	if err != nil && preview == nil {
		preview = func(gtx layout.Context) layout.Dimensions {
			return errorLabel(err)(gtx)
		}
	}
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.preview = preview
	e.previewImage = img
	e.previewExtra = extra
	e.previewError = err
	e.previewLoading = false
}

func (e *dirEntry) buildPreviewRle() (layout.Widget, [][2]string, image.Image, error) {
	f, err := os.Open(e.path)
	if err != nil {
		return nil, nil, nil, err
	}
	defer f.Close()
	pattern, err := patterns.PatternRleDecoder(f)
	if err != nil {
		extra := [][2]string{
			{"RLE error:", err.Error()},
		}
		p, e2 := e.buildPreviewCommon()
		return p, extra, nil, e2
	}
	extra := [][2]string{
		{"Name:", pattern.Name},
		{"Dimensions:", strconv.Itoa(pattern.Width) + "x" + strconv.Itoa(pattern.Height)},
	}
	if pattern.Rule != nil {
		extra = append(extra, [2]string{
			"Rule", pattern.Rule.Rle(),
		})
	}
	extra = append(extra,
		[2]string{"Origin:", pattern.Origination},
		[2]string{"Comments:", strings.Join(pattern.Comments, "\n")})
	p, e2 := e.buildPreviewCommon()
	const minCellSize = 4
	return flexVertical(0,
		rigid(p),
		rigid(func(gtx layout.Context) layout.Dimensions {
			border(gtx, layout.Dimensions{Size: image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(8)}}, false, false, true, false)
			return layout.Dimensions{Size: image.Point{Y: gtx.Dp(16)}}
		}),
		flexed(func(gtx layout.Context) layout.Dimensions {
			maxWd, maxHt := gtx.Constraints.Max.X, gtx.Constraints.Max.Y
			var img image.Image
			if pattern.Width > maxWd || pattern.Height > maxHt {
				if e.previewImage != nil {
					img = e.previewImage
				} else {
					pImg := imaging.PatternImagePaletted(pattern, imaging.Config{
						CellSize:   1,
						Borders:    false,
						AliveColor: color.NRGBA{A: 255},
						DeadColor:  color.NRGBA{R: 255, G: 255, B: 255, A: 255},
					})
					scale := min(float32(maxWd)/float32(pattern.Width), float32(maxHt)/float32(pattern.Height))
					img = imaging.ScaleSparse(pImg, scale)
					e.previewImage = img
				}
			} else {
				cellSize := min((maxWd-1)/pattern.Width, (maxHt-1)/pattern.Height)
				borders := cellSize > minCellSize
				img = imaging.PatternImage(pattern, imaging.Config{
					CellSize:    cellSize,
					Borders:     borders,
					AliveColor:  color.NRGBA{A: 255},
					DeadColor:   color.NRGBA{R: 255, G: 255, B: 255, A: 255},
					BorderColor: color.NRGBA{R: 128, G: 128, B: 128, A: 128},
				})
			}

			b := img.Bounds()
			iw := float32(b.Dx())
			ih := float32(b.Dy())
			maxW := float32(gtx.Constraints.Max.X)
			maxH := float32(gtx.Constraints.Max.Y)
			scale := min(maxW/iw, maxH/ih)
			w := int(iw * scale)
			h := int(ih * scale)
			stack := op.Affine(
				f32.Affine2D{}.Scale(
					f32.Point{},
					f32.Point{X: scale, Y: scale},
				),
			).Push(gtx.Ops)
			defer stack.Pop()
			paint.NewImageOp(img).Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			return layout.Dimensions{
				Size: image.Point{X: w, Y: h},
			}
		}),
	), extra, nil, e2
}

func (e *dirEntry) buildPreviewImg() (layout.Widget, [][2]string, image.Image, error) {
	p, err := e.buildPreviewCommon()
	if err != nil {
		return nil, nil, nil, err
	}
	f, err := os.Open(e.path)
	if err != nil {
		return nil, nil, nil, err
	}
	defer f.Close()
	var img image.Image
	switch filepath.Ext(e.name) {
	case ".png":
		img, err = png.Decode(f)
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(f)
	case ".gif":
		img, err = gif.Decode(f)
	}
	if err != nil {
		extra := [][2]string{{"PNG error:", err.Error()}}
		return p, extra, nil, nil
	}
	return flexVertical(0,
		rigid(p),
		rigid(func(gtx layout.Context) layout.Dimensions {
			border(gtx, layout.Dimensions{Size: image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(8)}}, false, false, true, false)
			return layout.Dimensions{Size: image.Point{Y: gtx.Dp(16)}}
		}),
		flexed(func(gtx layout.Context) layout.Dimensions {
			img := e.previewImage
			b := img.Bounds()
			if b.Dx() > gtx.Constraints.Max.X || b.Dy() > gtx.Constraints.Max.Y {
				iw := float32(b.Dx())
				ih := float32(b.Dy())
				maxW := float32(gtx.Constraints.Max.X)
				maxH := float32(gtx.Constraints.Max.Y)
				scale := min(maxW/iw, maxH/ih)
				w := int(iw * scale)
				h := int(ih * scale)
				stack := op.Affine(
					f32.Affine2D{}.Scale(
						f32.Point{},
						f32.Point{X: scale, Y: scale},
					),
				).Push(gtx.Ops)
				defer stack.Pop()
				paint.NewImageOp(img).Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				return layout.Dimensions{
					Size: image.Point{X: w, Y: h},
				}
			} else {
				size := img.Bounds().Size()
				stack := clip.Rect{Max: size}.Push(gtx.Ops)
				defer stack.Pop()
				paint.NewImageOp(img).Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				return layout.Dimensions{Size: size}
			}
		}),
	), nil, img, nil
}

func (e *dirEntry) buildPreviewJson() (layout.Widget, [][2]string, error) {
	f, err := os.Open(e.path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	var js any
	if err := json.NewDecoder(f).Decode(&js); err != nil {
		extra := [][2]string{{"JSON error:", err.Error()}}
		p, e2 := e.buildPreviewCommon()
		return p, extra, e2
	}
	p, err := e.buildPreviewCommon()
	if err != nil {
		return nil, nil, err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	_ = enc.Encode(js)
	jtxt := string(buf.Bytes())
	if e.previewScrollV == nil {
		e.previewScrollV = &widget.List{}
		e.previewScrollV.List.Axis = layout.Vertical
		e.previewScrollH = &widget.List{}
		e.previewScrollH.List.Axis = layout.Horizontal
	}
	return flexVertical(0,
		rigid(p),
		rigid(func(gtx layout.Context) layout.Dimensions {
			border(gtx, layout.Dimensions{Size: image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(8)}}, false, false, true, false)
			return layout.Dimensions{Size: image.Point{Y: gtx.Dp(16)}}
		}),
		flexed(func(gtx layout.Context) layout.Dimensions {
			dims := layout.Dimensions{Size: gtx.Constraints.Max}
			border(gtx, dims, true, true, true, true)
			return material.List(theme, e.previewScrollV).Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
				return material.List(theme, e.previewScrollH).Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
					gtx.Constraints.Min = image.Point{}
					gtx.Constraints.Max = image.Point{X: 1 << 30, Y: 1 << 30}
					lbl := material.Label(theme, (theme.TextSize*4)/5, jtxt)
					lbl.MaxLines = 0
					return lbl.Layout(gtx)
				})
			})
		}),
	), nil, nil
}

func (e *dirEntry) buildPreviewCommon() (layout.Widget, error) {
	stat, err := os.Stat(e.path)
	if err != nil {
		return nil, err
	}
	modifiedLabel := stat.ModTime().Format("2006-01-02 15:04:05")
	sizeLabel := fileSize(stat.Size())
	return flexVertical(8,
		rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(theme, (theme.TextSize*5)/4, e.name)
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		}),
		rigid(func(gtx layout.Context) layout.Dimensions {
			labelMax := measureMaxText(gtx, font.Bold, "Modified: ", "Size:: ", "Dimensions: ").Size.X
			rows := []layout.FlexChild{
				rigid(flexHorizontal(20,
					rigidLabel("Size:", text.End, font.Bold, labelMax),
					rigid(label(sizeLabel)),
				)),
				rigid(flexHorizontal(20,
					rigidLabel("Modified:", text.End, font.Bold, labelMax),
					rigid(label(modifiedLabel)),
				)),
			}
			if e.previewImage != nil && len(e.previewExtra) == 0 {
				dims := e.previewImage.Bounds()
				rows = append(rows,
					rigid(flexHorizontal(20,
						rigidLabel("Dimensions:", text.End, font.Bold, labelMax),
						rigid(label(strconv.Itoa(dims.Max.X)+" x "+strconv.Itoa(dims.Max.Y))),
					)),
				)
			}
			for _, x := range e.previewExtra {
				extra := x
				rows = append(rows, rigid(flexHorizontal(20,
					rigidLabel(extra[0], text.End, font.Bold, labelMax),
					rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Label(theme, theme.TextSize, extra[1]).Layout(gtx)
					}),
				)))
			}
			return flexVertical(4, rows...)(gtx)
		}),
	), nil
}

func (e *dirEntry) buildPreviewDir(allowExts map[string]struct{}) (layout.Widget, image.Image, error) {
	stat, err := os.Stat(e.path)
	if err != nil {
		return nil, nil, err
	} else if !stat.IsDir() {
		return nil, nil, fmt.Errorf("%q is not a directory", e.path)
	}
	modifiedLabel := stat.ModTime().Format("2006-01-02 15:04:05")
	files := 0
	selectable := 0
	subDirs := 0
	entries, err := os.ReadDir(e.path)
	if err != nil {
		return nil, nil, err
	}
	for _, entry := range entries {
		name := entry.Name()
		if !isHidden(name) {
			if entry.IsDir() {
				subDirs++
			} else {
				files++
				if len(allowExts) > 0 {
					ext := filepath.Ext(name)
					if _, ok := allowExts[ext]; ok {
						selectable++
					}
				} else {
					selectable++
				}
			}
		}
	}
	filesLabel := strconv.Itoa(files)
	if len(allowExts) > 0 {
		filesLabel += " (" + strconv.Itoa(selectable) + " selectable)"
	}
	dirsLabel := strconv.Itoa(subDirs)
	preview := flexVertical(8,
		rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(theme, (theme.TextSize*5)/4, e.name)
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		}),
		rigid(func(gtx layout.Context) layout.Dimensions {
			labelMax := measureMaxText(gtx, font.Bold, "Files: ", "Modified: ", "Sub-dirs: ").Size.X
			return flexVertical(4,
				rigid(flexHorizontal(20,
					rigidLabel("Modified:", text.End, font.Bold, labelMax),
					rigid(label(modifiedLabel)),
				)),
				rigid(flexHorizontal(20,
					rigidLabel("Files:", text.End, font.Bold, labelMax),
					rigid(label(filesLabel)),
				)),
				rigid(flexHorizontal(20,
					rigidLabel("Sub-dirs:", text.End, font.Bold, labelMax),
					rigid(label(dirsLabel)),
				)),
			)(gtx)
		}),
	)
	return preview, nil, nil
}

func fileSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case size >= GB:
		return strconv.FormatFloat(float64(size)/GB, 'f', 1, 64) + " GB"
	case size >= MB:
		return strconv.FormatFloat(float64(size)/MB, 'f', 1, 64) + " MB"
	case size >= KB:
		return strconv.FormatFloat(float64(size)/KB, 'f', 1, 64) + " KB"
	default:
		return strconv.Itoa(int(size)) + " bytes"
	}
}
