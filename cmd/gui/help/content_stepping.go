package help

import "github.com/marrow16/gogol/cmd/gui/icons"

var contentStepping = content{
	"Use the ", bold("Stepping"), " popout from the ", MainMenu.link("main menu"), " to adjust the stepping settings.\n\n",
	"The ", bold("Step delay (ms)"), " setting determines the delay between each step (in milliseconds) and can be set to a value between 0 and 2000.\n\n",
	"The ", bold("Step ahead size"), " setting determines the number of steps that are performed with a step ahead - and can be set to a value between 1 and 9999.\n",
	"Step ahead is instigated by pressing key ", keys{altMac, keyTab}, " or clicking ", iconText{image: icons.SkipForward}, " on the ", StatusBar.link("statusbar"), ".\n\n",
	"Use the ", bold("Snapshot on step ahead"), " checkbox to determine whether a snapshot of the grid is automatically taken on every step ahead.\n\n\n",
	italic("Note: The step delay is decoupled from the GUI frame rate - and setting a step delay of 0 (zero) will run the simulation as \"close to the metal\" as possible with visual updates at ~30Fps."),
}
