package animator

import (
	"errors"
	"github.com/marrow16/gogol/imaging"
	"github.com/marrow16/gogol/logic"
	"image/color"
	"image/png"
	"os"
	"os/exec"
)

var ffmpegChecked = false
var ffmpegAvailable = false

const (
	// palette colors...
	deadColor   = 0
	aliveColor  = 1
	borderColor = 2
)

func Mp4Available() bool {
	if !ffmpegChecked {
		ffmpegChecked = true
		ffmpegAvailable = false
		if _, err := exec.LookPath("ffmpeg"); err == nil {
			ffmpegAvailable = exec.Command("ffmpeg", "-version").Run() == nil
		}
	}
	return ffmpegAvailable
}

func NewAnimator(cellSize int, alive, dead, border color.NRGBA, borders bool, format string) *Animator {
	return &Animator{
		cellSize:        cellSize,
		alive:           alive,
		dead:            dead,
		border:          border,
		borders:         borders,
		animationFormat: format,
	}
}

type Animator struct {
	cellSize        int
	alive           color.NRGBA
	dead            color.NRGBA
	border          color.NRGBA
	borders         bool
	animationFormat string // "gif" or "mp4" - defaults to "gif"
}

func (a *Animator) Animate(filename string, recorder *logic.RecordInstrument) (err error) {
	if a.animationFormat != "mp4" {
		return a.animateGif(filename, recorder)
	}
	if !Mp4Available() {
		return errors.New("mp4 (ffmpeg) not available")
	}
	const (
		fps         = "30"
		preset      = "slow"
		crf         = "0"
		pixelFormat = "yuv420p"
	)
	grid := recorder.InitialGrid()
	img := imaging.GridSliceImage(grid, imaging.Config{
		CellSize:    a.cellSize,
		Borders:     a.borders,
		AliveColor:  a.alive,
		DeadColor:   a.dead,
		BorderColor: a.border,
	})
	cmd := exec.Command("ffmpeg",
		"-y",
		"-f", "image2pipe",
		"-vcodec", "png",
		"-framerate", fps,
		"-i", "-", // pipe input
		"-vf", "pad=ceil(iw/2)*2:ceil(ih/2)*2",
		"-c:v", "libx264",
		"-preset", preset,
		"-crf", crf,
		"-pix_fmt", pixelFormat,
		filename,
	)
	//var stderr bytes.Buffer
	//cmd.Stderr = &stderr
	ffmpegStdin, err := cmd.StdinPipe()
	if err != nil {
		return errors.New("Pipe create failed")
	}
	// start ffmpeg process in the background
	if err := cmd.Start(); err != nil {
		return errors.New("Failed to start ffmpeg")
	}
	// draw the initial frame (staring grid)
	if err = png.Encode(ffmpegStdin, img); err != nil {
		return errors.New("Failed to write start frame")
	}
	// iterate over the step changes from recorder...
	offset := 0
	if a.borders && a.cellSize > 2 {
		offset = 1
	}
	cellWidth := a.cellSize - offset
	stride := img.Stride
	for changes := range recorder.StepChangeLocations() {
		pix := img.Pix
		for _, change := range changes {
			row, col := change[0], change[1]
			alive := !grid[row][col]
			grid[row][col] = alive
			xMin := col*a.cellSize + offset
			yMin := row*a.cellSize + offset
			yMax := yMin + cellWidth
			line := yMin*stride + xMin*4
			lineBytes := cellWidth * 4
			c := a.dead
			if alive {
				c = a.alive
			}
			for y := yMin; y < yMax; y++ {
				i := line
				end := line + lineBytes
				for i < end {
					pix[i] = c.R
					pix[i+1] = c.G
					pix[i+2] = c.B
					pix[i+3] = c.A
					i += 4
				}
				line += stride
			}
		}
		if err = png.Encode(ffmpegStdin, img); err != nil {
			return errors.New("Failed to encode frame")
		}
	}
	if err = ffmpegStdin.Close(); err != nil {
		return errors.New("Failed to close ffmpeg pipe")
	}
	if err = cmd.Wait(); err != nil {
		//fmt.Printf("FFMPEG stderr: %s\n", stderr.String())
		return errors.New("Failed waiting on ffmpeg")
	}
	return nil
}

func (a *Animator) animateGif(filename string, recorder *logic.RecordInstrument) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	grid := recorder.InitialGrid()
	img := imaging.GridImageSlicePaletted(grid, imaging.Config{
		CellSize:    a.cellSize,
		Borders:     a.borders,
		AliveColor:  a.alive,
		DeadColor:   a.dead,
		BorderColor: a.border,
	})
	anim := &gifEncoder{
		w:      f,
		width:  img.Bounds().Dx(),
		height: img.Bounds().Dy(),
		palette: []color.RGBA{
			deadColor:   {a.dead.R, a.dead.G, a.dead.B, 255},
			aliveColor:  {a.alive.R, a.alive.G, a.alive.B, 255},
			borderColor: {a.border.R, a.border.G, a.border.B, 255},
		},
		loopCount: -1,
		delay:     7,
	}
	anim.writeHeader()
	// iterate over the step changes from recorder...
	offset := 0
	if a.borders && a.cellSize > 2 {
		offset = 1
	}
	stride := img.Stride
	cellWidth := a.cellSize - offset
	for changes := range recorder.StepChangeLocations() {
		pix := img.Pix
		for _, change := range changes {
			row, col := change[0], change[1]
			alive := !grid[row][col]
			grid[row][col] = alive
			xMin := col*a.cellSize + offset
			yMin := row*a.cellSize + offset
			yMax := yMin + cellWidth
			line := yMin*stride + xMin
			c := uint8(0)
			if alive {
				c = 1
			}
			for y := yMin; y < yMax; y++ {
				end := line + cellWidth
				for i := line; i < end; i++ {
					pix[i] = c
				}
				line += stride
			}
		}
		anim.writeImageBlock(img)
	}
	return anim.end()
}
