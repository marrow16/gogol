package widgets

import (
	"errors"
	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
	"image/color"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var modKeyName = func() string {
	if runtime.GOOS == "darwin" {
		return "⌥" //"⌘"
	}
	return "Ctrl+"
}()
var altKeyName = func() string {
	if runtime.GOOS == "darwin" {
		return "⌥" //"⌘"
	}
	return "Alt+"
}()
var modKey = func() key.Modifiers {
	if runtime.GOOS == "darwin" {
		return key.ModAlt
	}
	return key.ModCtrl
}()

var isMac = runtime.GOOS == "darwin"

var (
	backgroundColor                = color.NRGBA{R: 147, G: 147, B: 147, A: 255}
	popupForeground                = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	popupBackground                = color.NRGBA{R: 240, G: 240, B: 239, A: 255}
	popupBorder                    = color.NRGBA{R: 181, G: 181, B: 181, A: 255}
	popupBorderFocused             = color.NRGBA{R: 102, G: 128, B: 230, A: 230}
	popupBorderLight               = color.NRGBA{R: 250, G: 250, B: 250, A: 255}
	popupSelectedBackground        = color.NRGBA{R: 102, G: 128, B: 230, A: 128}
	popupSelectedFocusedBackground = color.NRGBA{R: 102, G: 128, B: 230, A: 200}
	popupHighlightColor            = popupSelectedFocusedBackground
	popupLinkColor                 = color.NRGBA{R: 102, G: 128, B: 230, A: 255}
	errorColor                     = color.NRGBA{R: 200, G: 0, B: 0, A: 255}
	borderThicknessNormal          = unit.Dp(1)
	borderThicknessFocused         = unit.Dp(2)
)

func focusedBorder(focused bool) (color.NRGBA, unit.Dp) {
	if focused {
		return popupBorderFocused, borderThicknessFocused
	}
	return popupBorder, borderThicknessNormal
}

func commas(s string) string {
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func saveFile(path string, allowOverwrite bool) (result *os.File, err error) {
	if path, err = resolveSavePath(path); err != nil {
		return
	}
	if allowOverwrite {
		if result, err = os.Create(path); err != nil {
			err = errors.New("Unable to create file")
		}
	} else if result, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		if errors.Is(err, fs.ErrExist) {
			err = errors.New("File already exists")
		} else {
			err = errors.New("Unable to create file")
		}
	}
	return result, err
}

func resolveSavePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("Invalid user home directory")
	}
	dir := filepath.Join(home, "Documents", "GoGoL")
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return "", errors.New("Invalid user home directory")
	}
	result := filepath.Join(dir, path)
	dir = filepath.Dir(result)
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return "", errors.New("Invalid path")
	}
	return result, nil
}

func filePicker(fn func(filename string)) {
	if fn != nil && isMac {
		out, err := exec.Command(
			"osascript",
			"-e",
			`POSIX path of (choose file)`,
		).Output()
		if err == nil {
			fn(string(out))
		}
	}
}

func openURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default: // Linux, BSD, etc.
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func openInBrowser(filename string) {
	fp, err := filepath.Abs(filename)
	if err != nil {
		return
	}
	u := (&url.URL{
		Scheme: "file",
		Path:   fp,
	}).String()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-a", "Google Chrome", u)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	default:
		cmd = exec.Command("xdg-open", u)
	}
	_ = cmd.Start()
}

func border(gtx layout.Context, dims layout.Dimensions, top, left, bottom, right bool) {
	if top {
		paint.FillShape(gtx.Ops, popupBorder,
			clip.Rect(image.Rect(0, 0, dims.Size.X, 1)).Op(),
		)
	}
	if left {
		paint.FillShape(gtx.Ops, popupBorder,
			clip.Rect(image.Rect(0, 0, 1, dims.Size.Y)).Op(),
		)
	}
	if bottom {
		paint.FillShape(gtx.Ops, popupBorder,
			clip.Rect(image.Rect(0, dims.Size.Y-1, dims.Size.X, dims.Size.Y)).Op(),
		)
	}
	if right {
		paint.FillShape(gtx.Ops, popupBorder,
			clip.Rect(image.Rect(dims.Size.X-1, 0, dims.Size.X, dims.Size.Y)).Op(),
		)
	}
}

func errorLabel(err error) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if err == nil {
			return layout.Dimensions{}
		}
		lbl := material.Label(theme, theme.TextSize, err.Error())
		lbl.MaxLines = 1
		lbl.Color = errorColor
		return lbl.Layout(gtx)
	}
}

func insetErrorLabel(err error) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if err == nil {
			return layout.Dimensions{}
		}
		lbl := material.Label(theme, theme.TextSize, err.Error())
		lbl.MaxLines = 1
		lbl.Color = errorColor
		return layout.Inset{Top: 3}.Layout(gtx, lbl.Layout)
	}
}

func label(s string) layout.Widget {
	lbl := material.Label(theme, theme.TextSize, s)
	lbl.MaxLines = 1
	return lbl.Layout
}

func textLabel(s string) layout.Widget {
	return material.Label(theme, theme.TextSize, s).Layout
}

func labelRight(s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(theme, theme.TextSize, s)
		lbl.Alignment = text.End
		lbl.MaxLines = 1
		return lbl.Layout(gtx)
	}
}

func insetLabel(s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(theme, theme.TextSize, s)
		lbl.MaxLines = 1
		return layout.Inset{Top: 3}.Layout(gtx, lbl.Layout)
	}
}

func linkLabel(btn *widget.Clickable, s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return material.Clickable(gtx, btn, func(gtx layout.Context) layout.Dimensions {
			return layout.Stack{}.Layout(gtx,
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(theme, theme.TextSize, s)
					lbl.MaxLines = 1
					lbl.Color = popupLinkColor
					return lbl.Layout(gtx)
				}),
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					thickness := gtx.Dp(1)
					rect := image.Rect(
						0,
						gtx.Constraints.Min.Y-thickness,
						gtx.Constraints.Min.X,
						gtx.Constraints.Min.Y,
					)
					defer clip.Rect(rect).Push(gtx.Ops).Pop()
					paint.ColorOp{Color: popupLinkColor}.Add(gtx.Ops)
					paint.PaintOp{}.Add(gtx.Ops)
					return layout.Dimensions{Size: gtx.Constraints.Min}
				}),
			)
		})
	}
}

func inset(top, bottom, left, right unit.Dp, w layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: top, Bottom: bottom, Left: left, Right: right}.Layout(gtx, w)
	}
}

func borderedInset(top, bottom, left, right unit.Dp, w layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return widget.Border{
			Color:        popupBorder,
			CornerRadius: 3,
			Width:        borderThicknessNormal,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: top, Bottom: bottom, Left: left, Right: right}.Layout(gtx, w)
		})
	}
}

func flexed(w layout.Widget, weighted ...float32) layout.FlexChild {
	weight := float32(1.0)
	if len(weighted) > 0 {
		weight = weighted[0]
	}
	return layout.Flexed(weight, func(gtx layout.Context) layout.Dimensions {
		return w(gtx)
	})
}

func conditionalFlexed(cond bool, w1, w2 layout.Widget) layout.FlexChild {
	return layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		if cond {
			return w1(gtx)
		} else if w2 != nil {
			return w2(gtx)
		}
		return layout.Dimensions{}
	})
}

func conditionalRigid(cond bool, w1, w2 layout.Widget) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		if cond {
			return w1(gtx)
		} else if w2 != nil {
			return w2(gtx)
		}
		return layout.Dimensions{}
	})
}

func rigid(w layout.Widget) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return w(gtx)
	})
}

func rigidImage(img *image.NRGBA) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		size := img.Bounds().Size()
		stack := clip.Rect{Max: size}.Push(gtx.Ops)
		defer stack.Pop()
		paint.NewImageOp(img).Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		return layout.Dimensions{Size: size}
	})
}

func rigidLabel(label string, align text.Alignment, weight font.Weight, fixedWidth int) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		if fixedWidth > 0 {
			gtx.Constraints.Min.X = fixedWidth
			gtx.Constraints.Max.X = fixedWidth
		}
		lbl := material.Label(theme, theme.TextSize, label)
		lbl.MaxLines = 1
		lbl.Alignment = align
		lbl.Font.Weight = weight
		return lbl.Layout(gtx)
	})
}

func rigidFixedWidth(w layout.Widget, width int, direction layout.Direction) layout.FlexChild {
	if w == nil {
		return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = width
			return layout.Dimensions{Size: gtx.Constraints.Min}
		})
	}
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		if width > 0 {
			gtx.Constraints.Min.X = width
			gtx.Constraints.Max.X = width
		}
		return direction.Layout(gtx, w)
	})
}

func rigidSpacerVertical(ht int) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Point{Y: ht}}
	})
}

func flexVertical(gap int, children ...layout.FlexChild) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Gap: gap}.Layout(gtx, children...)
	}
}

func flexHorizontal(gap int, children ...layout.FlexChild) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Gap: gap}.Layout(gtx, children...)
	}
}

func popoutLayout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	return layout.Inset{Left: 8, Right: 8, Top: 8, Bottom: 8}.Layout(gtx, w)
}

func measureText(gtx layout.Context, text string) layout.Dimensions {
	gtx.Constraints.Min = image.Point{}
	gtx.Constraints.Max = image.Pt(1e6, 1e6)
	macro := op.Record(gtx.Ops)
	dims := material.Label(theme, theme.TextSize, text).Layout(gtx)
	_ = macro.Stop()
	return dims
}

func measureMaxText(gtx layout.Context, weight font.Weight, text ...string) layout.Dimensions {
	mx := layout.Dimensions{}
	macro := op.Record(gtx.Ops)
	for _, t := range text {
		lbl := material.Label(theme, theme.TextSize, t)
		lbl.Font.Weight = weight
		dims := lbl.Layout(gtx)
		if dims.Size.X > mx.Size.X {
			mx.Size.X = dims.Size.X
		}
		if dims.Size.Y > mx.Size.Y {
			mx.Size.Y = dims.Size.Y
		}
	}
	_ = macro.Stop()
	return mx
}
