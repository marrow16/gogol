package help

var contentColors = content{
	"Use the ", bold("Colors"), " popout from the ", MainMenu.link("main menu"), " to adjust the colors of cells on the grid.\n\n",
	"Each color setting ", bold("Alive cells"), ", ", bold("Dead cells"), " & ", bold("Cell border"),
	" is represented as an edit with the ", code(`#rrggbb`), " color string.",
	" Use ", keys{keyUp}, " and ", keys{keyDown}, " keys to increment/decrement any part of the color string.",
	" Click the color box (or press ", keys{keyEnter}, " key) to show the color picker.",
	" Having changed the color, use key ", keys{cmdMac, "Z"}, " to revert.",
	"\n\nUse the ", bold("Show Borders"), " checkbox to determine whether the grid displays borders.\n",
	"Cell borders can also be toggled by pressing key ", keys{altMac, "B"}, ".",
}
