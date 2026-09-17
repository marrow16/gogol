package help

var contentImportGrid = content{
	"Use the ", bold("Import Grid"), " popout from the ", MainMenu.link("main menu"), " to load the current grid from a previously exported grid.\n\n",
	"Enter the ", bold("Path"), " for the grid ", code(".rle"), " file to import or use the ", button("..."), " button to open the ", FileFinder.link("file finder"), ".\n",
	"Use the ", bold("Resize grid"), " to determine whether the current grid can be resized to accommodate the imported grid",
	" and press the ", button("Import"), " button\n\n",
	"If the grid is successfully imported, the current grid will be updated - otherwise, an error is displayed.",
}
