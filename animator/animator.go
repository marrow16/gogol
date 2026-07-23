package animator

import (
	"errors"
	"github.com/marrow16/gogol/logic"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os/exec"
)

func NewAnimator(cellSize int, alive, dead, border color.NRGBA, borders bool) *Animator {
	return &Animator{
		cellSize: cellSize,
		alive:    alive,
		dead:     dead,
		border:   border,
		borders:  borders,
	}
}

type Animator struct {
	cellSize      int
	alive         color.NRGBA
	dead          color.NRGBA
	border        color.NRGBA
	borders       bool
	height, width int
	grid          [][]bool
	img           *image.NRGBA
	borderOffset  int
}

func (a *Animator) Animate(filename string, recorder *logic.RecordInstrument) (err error) {
	const (
		fps         = "30"
		preset      = "slow"
		crf         = "0"
		pixelFormat = "yuv420p"
	)
	a.grid = recorder.InitialGrid()
	a.height = len(a.grid)
	a.width = len(a.grid[0])
	a.drawInitialGrid()
	cmd := exec.Command("ffmpeg",
		"-y",
		"-f", "image2pipe",
		"-vcodec", "png",
		"-framerate", fps,
		"-i", "-", // pipe input
		"-c:v", "libx264",
		"-preset", preset,
		"-crf", crf,
		"-pix_fmt", pixelFormat,
		filename,
	)
	ffmpegStdin, err := cmd.StdinPipe()
	if err != nil {
		return errors.New("Pipe create failed")
	}
	// start ffmpeg process in the background
	if err := cmd.Start(); err != nil {
		return errors.New("Failed to start ffmpeg")
	}
	// draw the initial frame (staring grid)
	if err = png.Encode(ffmpegStdin, a.img); err != nil {
		return errors.New("Failed to write start frame")
	}
	// iterate over the step changes from recorder...
	for changes := range recorder.StepChangeLocations() {
		for _, change := range changes {
			r, c := change[0], change[1]
			alive := !a.grid[r][c]
			a.grid[r][c] = alive
			a.drawCell(r, c, alive)
		}
		if err = png.Encode(ffmpegStdin, a.img); err != nil {
			return errors.New("Failed to encode frame")
		}
	}
	if err = ffmpegStdin.Close(); err != nil {
		return errors.New("Failed to close ffmpeg pipe")
	}
	if err = cmd.Wait(); err != nil {
		return errors.New("Failed waiting on ffmpeg")
	}
	return nil
}

func (a *Animator) drawInitialGrid() {
	wd, ht := a.width*a.cellSize, a.height*a.cellSize
	a.img = image.NewNRGBA(image.Rect(0, 0, wd, ht))
	draw.Draw(a.img, image.Rect(0, 0, wd, ht), &image.Uniform{a.dead}, image.Point{}, draw.Src)
	a.borderOffset = 0
	if a.borders {
		a.borderOffset = 1
		for y := 0; y <= a.height; y++ {
			yy := y * a.cellSize
			draw.Draw(
				a.img,
				image.Rect(0, yy, wd, yy+1),
				&image.Uniform{a.border},
				image.Point{},
				draw.Src,
			)
		}
		for x := 0; x <= a.width; x++ {
			xx := x * a.cellSize
			draw.Draw(
				a.img,
				image.Rect(xx, 0, xx+1, ht),
				&image.Uniform{a.border},
				image.Point{},
				draw.Src,
			)
		}
	}
	for r, row := range a.grid {
		for c, alive := range row {
			if alive {
				a.drawCell(r, c, alive)
			}
		}
	}
}

func (a *Animator) drawCell(r, c int, alive bool) {
	clr := a.dead
	if alive {
		clr = a.alive
	}
	draw.Draw(a.img, image.Rect(
		(c*a.cellSize)+a.borderOffset,
		(r*a.cellSize)+a.borderOffset,
		((c+1)*a.cellSize)-a.borderOffset,
		((r+1)*a.cellSize)-a.borderOffset),
		&image.Uniform{clr}, image.Point{}, draw.Src)
}
