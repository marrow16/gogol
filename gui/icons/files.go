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

//go:embed refresh.png
var refresh []byte

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

//go:embed fileTXT.png
var fileTXT []byte

//go:embed button-0.png
var button0 []byte

//go:embed button-1.png
var button1 []byte

//go:embed button-2.png
var button2 []byte

//go:embed button-3.png
var button3 []byte

//go:embed button-4.png
var button4 []byte

//go:embed button-5.png
var button5 []byte

//go:embed button-6.png
var button6 []byte

//go:embed button-7.png
var button7 []byte

//go:embed button-8.png
var button8 []byte

//go:embed button-9.png
var button9 []byte

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

//go:embed button-g.png
var buttonG []byte

//go:embed button-h.png
var buttonH []byte

//go:embed button-i.png
var buttonI []byte

//go:embed button-j.png
var buttonJ []byte

//go:embed button-k.png
var buttonK []byte

//go:embed button-l.png
var buttonL []byte

//go:embed button-m.png
var buttonM []byte

//go:embed button-n.png
var buttonN []byte

//go:embed button-o.png
var buttonO []byte

//go:embed button-p.png
var buttonP []byte

//go:embed button-r.png
var buttonR []byte

//go:embed button-s.png
var buttonS []byte

//go:embed button-t.png
var buttonT []byte

//go:embed button-u.png
var buttonU []byte

//go:embed button-v.png
var buttonV []byte

//go:embed button-w.png
var buttonW []byte

//go:embed button-x.png
var buttonX []byte

//go:embed button-y.png
var buttonY []byte

//go:embed button-z.png
var buttonZ []byte

//go:embed button-gamma.png
var buttonGamma []byte

//go:embed button-delta.png
var buttonDelta []byte

//go:embed button-theta.png
var buttonTheta []byte

//go:embed button-lambda.png
var buttonLambda []byte

//go:embed button-xi.png
var buttonXi []byte

//go:embed button-pi.png
var buttonPi []byte

//go:embed button-sigma.png
var buttonSigma []byte

//go:embed button-phi.png
var buttonPhi []byte

//go:embed button-psi.png
var buttonPsi []byte

//go:embed button-omega.png
var buttonOmega []byte

var (
	Backward     = mustImage(backward)
	Burger       = mustImage(burger)
	Button0      = mustImage(button0)
	Button1      = mustImage(button1)
	Button2      = mustImage(button2)
	Button3      = mustImage(button3)
	Button4      = mustImage(button4)
	Button5      = mustImage(button5)
	Button6      = mustImage(button6)
	Button7      = mustImage(button7)
	Button8      = mustImage(button8)
	Button9      = mustImage(button9)
	ButtonA      = mustImage(buttonA)
	ButtonB      = mustImage(buttonB)
	ButtonC      = mustImage(buttonC)
	ButtonD      = mustImage(buttonD)
	ButtonE      = mustImage(buttonE)
	ButtonF      = mustImage(buttonF)
	ButtonG      = mustImage(buttonG)
	ButtonH      = mustImage(buttonH)
	ButtonI      = mustImage(buttonI)
	ButtonJ      = mustImage(buttonJ)
	ButtonK      = mustImage(buttonK)
	ButtonL      = mustImage(buttonL)
	ButtonM      = mustImage(buttonM)
	ButtonN      = mustImage(buttonN)
	ButtonO      = mustImage(buttonO)
	ButtonP      = mustImage(buttonP)
	ButtonR      = mustImage(buttonR)
	ButtonS      = mustImage(buttonS)
	ButtonT      = mustImage(buttonT)
	ButtonU      = mustImage(buttonU)
	ButtonV      = mustImage(buttonV)
	ButtonW      = mustImage(buttonW)
	ButtonX      = mustImage(buttonX)
	ButtonY      = mustImage(buttonY)
	ButtonZ      = mustImage(buttonZ)
	ButtonGamma  = mustImage(buttonGamma)
	ButtonDelta  = mustImage(buttonDelta)
	ButtonTheta  = mustImage(buttonTheta)
	ButtonLambda = mustImage(buttonLambda)
	ButtonXi     = mustImage(buttonXi)
	ButtonPi     = mustImage(buttonPi)
	ButtonSigma  = mustImage(buttonSigma)
	ButtonPhi    = mustImage(buttonPhi)
	ButtonPsi    = mustImage(buttonPsi)
	ButtonOmega  = mustImage(buttonOmega)
	Close        = mustImage(closeIcon)
	Copy         = mustImage(copyIcon)
	Pause        = mustImage(pause)
	Play         = mustImage(play)
	Plus         = mustImage(plus)
	Minus        = mustImage(minus)
	Refresh      = mustImage(refresh)
	Reverse      = mustImage(reverse)
	Search       = mustImage(search)
	SkipBackward = mustImage(skipBackward)
	SkipForward  = mustImage(skipForward)
	Step         = mustImage(step)
	Stop         = mustImage(stop)
	ZoomIn       = mustImage(zoomIn)
	ZoomOut      = mustImage(zoomOut)

	Folder   = mustImage(folder)
	FileRLE  = mustImage(fileRLE)
	FileIMG  = mustImage(fileIMG)
	FileJSON = mustImage(fileJSON)
	FileTXT  = mustImage(fileTXT)
)

func mustImage(data []byte) image.Image {
	if img, _, err := image.Decode(bytes.NewReader(data)); err == nil {
		return img
	} else {
		panic(err)
	}
}
