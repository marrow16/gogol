package help

var contentFileFinder = content{
	"File finder is the built-in file (and path) selector for GoGoL - replacing native OS file finders and file open dialogs.\n",
	"It specifically allows previewing ", code(".rle"), " files (including meta data). ",
	"It also allows limited previewing of ", code(".png"), ", ", code(".jpg"), ", ", code(".jpeg"), " and ", code(".gif"), " files.\n\n",
	"The file finder can be reached by pressing the ", button("..."), " button in ", LoadPatterns.link("Load Patterns"), ", ", GridRecipes.link("Grid Recipes"), " or ", ImportGrid.link("Import Grid"), ".",
	h4("Usage", 8, 4),
	bold("Path"), " shows the current path:",
	indent{indent: 20, spaceAfter: 8, content: content{
		hanging{prefix: "• ", content: "Click on any of the path items to navigate up the folder path."},
		hanging{prefix: "• ", content: content{
			"Press keys ", keys{keyLeft}, " or ", keys{keyBack}, " to navigate up to parent folder.",
		}},
	}},
	"The left pane shows the list of files and folders in the current path:",
	indent{indent: 20, spaceAfter: 8, content: content{
		hanging{prefix: "• ", content: "Clicking on any file or folder selects it."},
		hanging{prefix: "• ", content: content{
			"Use keys ", keys{keyUp}, " ", keys{keyDown}, " ", keys{"PgUp"}, " ", keys{"PgDn"}, " ", keys{"Home"}, " and ", keys{"End"}, " to navigate.",
		}},
		hanging{prefix: "• ", content: content{
			"When on a folder, click again or press ", keys{keyRight}, " to navigate into that folder.",
		}},
		hanging{prefix: "• ", content: "Press any character key to navigate to the next file/folder whose name starts with that character."},
	}},
	"The right pane shows the details of the currently selected file or folder:",
	indent{indent: 20, spaceAfter: 8, content: content{
		hanging{prefix: "• ", content: content{
			"If the currently selected file is a ", code(".rle"), ", ", code(".png"), ", ", code(".jpg"), ", ", code(".jpeg"), " or ", code(".gif"),
			" - a preview is displayed.",
		}},
	}},
	"The bottom pane shows controls:",
	indent{indent: 20, spaceAfter: 8, content: content{
		hanging{prefix: "• ", content: content{
			"Use the ", button("Only show selectable files"), " checkbox to limit files shown to only those that can be selected.",
		}},
		hanging{prefix: "• ", content: content{
			"Click ", button("Open"), " button to select the current file/folder and return to the calling popout.",
			"\n", italic("This button will be disabled if the current file type is not allowed or selecting a folder when folder selection is not permissible."),
		}},
		hanging{prefix: "• ", content: content{
			"Click on ", button("Cancel"), " button or press ", keys{"Esc"}, " to exit file finder",
		}},
	}},
	italic("Note: Clicking on the window close will "), boldItalic("not"), " close the file finder - it will close the entire application!",
}
