package help

var contentColors = content{
	"Use the ", bold("Colors"), " popout from the ", MainMenu.link("main menu"), " to adjust the colors of cells on the grid.\n\n",
	"Each color setting, ", bold("Alive cells"), ", ", bold("Dead cells"), " & ", bold("Cell border"), ", ",
	"is represented by a ", bold("R"), "(red), ", bold("G"), "(green) & ", bold("B"), "(blue) component - ",
	"enter a value between 0 and 255.\n",
	"Within each color component input, use keys ", keys{keyUpTriangle}, " and ", keys{keyDownTriangle}, " to increment/decrement the value.\n\n",
	"Use the ", bold("Show Borders"), " checkbox to determine whether the grid displays borders.\n",
	"Cell borders can also be toggled by pressing key ", keys{altMac, "B"}, ".",
}
