package main

import (
	"gioui.org/app"
	"gioui.org/unit"
	"github.com/marrow16/gogol/cmd/gui/settings"
	"github.com/marrow16/gogol/cmd/gui/widgets"
	"log"
	"os"
)

func main() {
	s := settings.NewSettings()
	core, err := widgets.NewCore(s)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		window := new(app.Window)
		window.Option(
			app.Title("GoGoL"),
			app.Size(unit.Dp(s.ScreenWidth), unit.Dp(s.ScreenHeight)),
		)

		err = core.Run(window)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
