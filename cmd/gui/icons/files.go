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

//go:embed forward.png
var forward []byte

//go:embed pause.png
var pause []byte

//go:embed play.png
var play []byte

//go:embed record.png
var record []byte

//go:embed rewind.png
var rewind []byte

//go:embed reverse.png
var reverse []byte

//go:embed skip-backward.png
var skipBackward []byte

//go:embed skip-forward.png
var skipForward []byte

//go:embed step.png
var step []byte

//go:embed stop.png
var stop []byte

//go:embed zoom.png
var zoom []byte

//go:embed zoomIn.png
var zoomIn []byte

//go:embed zoomOut.png
var zoomOut []byte

var (
	Backward     = mustImage(backward)
	Burger       = mustImage(burger)
	Forward      = mustImage(forward)
	Pause        = mustImage(pause)
	Play         = mustImage(play)
	Record       = mustImage(record)
	Reverse      = mustImage(reverse)
	Rewind       = mustImage(rewind)
	SkipBackward = mustImage(skipBackward)
	SkipForward  = mustImage(skipForward)
	Step         = mustImage(step)
	Stop         = mustImage(stop)
	Zoom         = mustImage(zoom)
	ZoomIn       = mustImage(zoomIn)
	ZoomOut      = mustImage(zoomOut)
)

func mustImage(data []byte) image.Image {
	if img, _, err := image.Decode(bytes.NewReader(data)); err == nil {
		return img
	} else {
		panic(err)
	}
}
