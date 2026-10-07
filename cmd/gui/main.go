package main

import (
	"gioui.org/app"
	"gioui.org/unit"
	"github.com/marrow16/gogol/gui"
	"github.com/marrow16/gogol/gui/settings"
	"github.com/marrow16/gogol/logic"
	"log"
	"os"
)

func main() {
	s := settings.NewSettings()
	for rn, rle := range s.Rules {
		if r, err := logic.NewRuleRle(rn, rle); err == nil {
			logic.AddRule(rn, r)
		}
	}
	core, err := gui.NewCore(s)
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
