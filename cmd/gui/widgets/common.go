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
	"gioui.org/widget/material"
	"image"
	"image/color"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var modKeyName = func() string {
	if runtime.GOOS == "darwin" {
		return "⌥" //"⌘"
	}
	return "Ctrl+"
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
	errorColor                     = color.NRGBA{R: 200, G: 0, B: 0, A: 255}
)

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
	return filepath.Join(dir, filepath.Base(path)), nil
}

func nowFilename(prefix string, extension string) string {
	return prefix + " " + time.Now().Format("2006-01-02 15-04-05.999") + extension
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
			clip.Rect(image.Rect(
				0,
				0,
				dims.Size.X,
				1,
			)).Op(),
		)
	}
	if left {
		paint.FillShape(gtx.Ops, popupBorder,
			clip.Rect(image.Rect(
				0,
				0,
				1,
				dims.Size.Y,
			)).Op(),
		)
	}
	if bottom {
		paint.FillShape(gtx.Ops, popupBorder,
			clip.Rect(image.Rect(
				0,
				dims.Size.Y-1,
				dims.Size.X,
				dims.Size.Y,
			)).Op(),
		)
	}
	if right {
		paint.FillShape(gtx.Ops, popupBorder,
			clip.Rect(image.Rect(
				dims.Size.X-1,
				0,
				dims.Size.X,
				dims.Size.Y,
			)).Op(),
		)
	}
}

func errorLabel(theme *material.Theme, err error) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(theme, theme.TextSize, err.Error())
		lbl.MaxLines = 1
		lbl.Color = errorColor
		return lbl.Layout(gtx)
	}
}

func label(theme *material.Theme, s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(theme, theme.TextSize, s)
		lbl.MaxLines = 1
		return lbl.Layout(gtx)
	}
}

func rightAlignedLabel(theme *material.Theme, s string, width int) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min.X = width
		lbl := material.Label(theme, theme.TextSize, s)
		lbl.Alignment = text.End
		lbl.MaxLines = 1
		return lbl.Layout(gtx)
	}
}

func rightAlignedBoldLabel(theme *material.Theme, s string, width int) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min.X = width
		lbl := material.Label(theme, theme.TextSize, s)
		lbl.Alignment = text.End
		lbl.Font.Weight = font.Bold
		lbl.MaxLines = 1
		return lbl.Layout(gtx)
	}
}

func measureText(gtx layout.Context, theme *material.Theme, text string) layout.Dimensions {
	gtx.Constraints.Min = image.Point{}
	gtx.Constraints.Max = image.Pt(1e6, 1e6)
	macro := op.Record(gtx.Ops)
	dims := material.Label(theme, theme.TextSize, text).Layout(gtx)
	_ = macro.Stop()
	return dims
}

func measureMaxText(gtx layout.Context, theme *material.Theme, weight font.Weight, text ...string) layout.Dimensions {
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
