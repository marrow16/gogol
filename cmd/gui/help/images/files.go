package images

import (
	"bytes"
	"embed"
	"image"
	_ "image/png"
)

//go:embed *.png
var files embed.FS

func LoadImage(name string) (image.Image, error) {
	if data, err := files.ReadFile(name); err == nil {
		img, _, err := image.Decode(bytes.NewReader(data))
		return img, err
	} else {
		return nil, err
	}
}
