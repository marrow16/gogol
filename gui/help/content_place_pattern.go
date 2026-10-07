package help

var contentPlacePattern = content{
	"Place pattern mode is instigated from the ", Patterns.link("Patterns"), " popout from the ", MainMenu.link("main menu"),
	" or, if a pattern is already selected, by pressing ", keys{altMac, "P"}, ".\n\n",
	"When in pattern place mode, the selected pattern will appear on the grid as a highlighted overlay.\n",
	"Move the pattern to the desired position using keys ", keys{keyLeft}, ", ", keys{keyRight}, ", ", keys{keyUp}, " and ", keys{keyDown}, ".",
	" Use key ", keys{altMac, "R"}, " to alter the rotation.\n\n",
	"Once the desired position and rotation are found, press ", keys{keyEnter}, " to place the pattern on the grid.\n\n",
	italic("Press "), keys{"Esc"}, italic(" to abandon the pattern placement and exit placement mode."),
}
