package help

var contentLoadPatterns = content{
	"The ", bold("Load Patterns"), " popout is available from the ", MainMenu.link("main menu"), ".\n\n",
	"Enter the ", bold("Path"), " to either a pattern file or a directory containing pattern files. Alternatively, use the ", button("..."), " button to open ", FileFinder.link("file finder"), ".\n",
	"Press the ", button("Load"), " button - the number of successfully loaded patterns will be displayed or an error.\n\n",
	"On successful load, the patterns will be available in the ", Patterns.link("Patterns"), " popout (and are also loaded next time you restart GoGoL).",
}
