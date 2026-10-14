package icons

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
)

//go:embed backward.png
var backward []byte

//go:embed burger.png
var burger []byte

//go:embed button-a.png
var buttonA []byte

//go:embed button-b.png
var buttonB []byte

//go:embed button-c.png
var buttonC []byte

//go:embed button-d.png
var buttonD []byte

//go:embed button-e.png
var buttonE []byte

//go:embed button-f.png
var buttonF []byte

//go:embed close.png
var closeIcon []byte

//go:embed copy.png
var copyIcon []byte

//go:embed pause.png
var pause []byte

//go:embed play.png
var play []byte

//go:embed plus.png
var plus []byte

//go:embed minus.png
var minus []byte

//go:embed reverse.png
var reverse []byte

//go:embed skip-backward.png
var skipBackward []byte

//go:embed skip-forward.png
var skipForward []byte

//go:embed step.png
var step []byte

//go:embed zoomIn.png
var zoomIn []byte

//go:embed zoomOut.png
var zoomOut []byte

//go:embed search.png
var search []byte

//go:embed folder.png
var folder []byte

//go:embed fileRLE.png
var fileRLE []byte

//go:embed fileIMG.png
var fileIMG []byte

//go:embed fileJSON.png
var fileJSON []byte

var (
	Backward     = mustImage(backward)
	Burger       = mustImage(burger)
	ButtonA      = mustImage(buttonA)
	ButtonB      = mustImage(buttonB)
	ButtonC      = mustImage(buttonC)
	ButtonD      = mustImage(buttonD)
	ButtonE      = mustImage(buttonE)
	ButtonF      = mustImage(buttonF)
	Close        = mustImage(closeIcon)
	Copy         = mustImage(copyIcon)
	Pause        = mustImage(pause)
	Play         = mustImage(play)
	Plus         = mustImage(plus)
	Minus        = mustImage(minus)
	Reverse      = mustImage(reverse)
	Search       = mustImage(search)
	SkipBackward = mustImage(skipBackward)
	SkipForward  = mustImage(skipForward)
	Step         = mustImage(step)
	ZoomIn       = mustImage(zoomIn)
	ZoomOut      = mustImage(zoomOut)

	Folder   = mustImage(folder)
	FileRLE  = mustImage(fileRLE)
	FileIMG  = mustImage(fileIMG)
	FileJSON = mustImage(fileJSON)
)

func mustImage(data []byte) image.Image {
	if img, _, err := image.Decode(bytes.NewReader(data)); err == nil {
		return img
	} else {
		panic(err)
	}
}
