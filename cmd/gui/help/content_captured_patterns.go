package help

var contentCapturedPatterns = content{
	"The ", bold("Captured Patterns"), " popout is available from the ", MainMenu.link("main menu"), ".\n\n",
	"Captured patterns are patterns that have been obtained from a grid state - see ", Editor.link("edit mode"), ".\n",
	"Briefly, in edit mode, just mark an area of the grid using ", keys{"Shift+", keyLeft}, " ", keys{"Shift+", keyRight}, " ", keys{"Shift+", keyUp}, " ", keys{"Shift+", keyDown}, " keys and then press ", keys{keyEnter}, " to capture. ",
	"Multiple patterns can be captured in a single edit session.\n\n",
	"Once a pattern has been captured from grid edit, it will appear in the dropdown.\n",
	"You can also add captured patterns from external sources by copying the rle content and pressing the ", button("Paste RLE"), " button in this popout.\n\n",
	"Having selected a captured pattern from the dropdown - use the view modes:",
	indent{spaceBefore: 8, content: content{
		h5("Preview"),
		indent{indent: 20, content: content{
			"Shows a preview image of the captured pattern.\n\n",
			"Use the ", button("Identify"), " button to search for the same pattern in currently loaded ", Patterns.link("patterns"), ". ",
			"All rotations and mirrors of the pattern will be searched - and if the ", button("With phases"), " is checked, GoGoL will run a short simulation on the pattern to detect transitional phases to search for.\n",
			"On identify, the number of hashes searched and whether the pattern was found is displayed. If the pattern is found, a link is displayed and clicking on that link will take you to the ", Patterns.link("pattern"), ".\n",
			italic("Note: On first use of identify, GoGoL must build an in-memory hash index for loaded patterns - this can take a couple of seconds for large libraries (i.e. >4000 patterns)"),
		}},
	}},
	indent{spaceBefore: 8, content: content{
		h5("Metadata"),
		indent{indent: 20, content: content{
			"Shows editable metadata for the captured pattern.\n\n",
			"Edit the metadata as required and then press ", button("Save"), " button to save the pattern to an ", code(".rle"), " file. ",
			"Check the ", button("Add to library"), " option to also add it to the ", Patterns.link("patterns"), " library.\n\n",
			"Use the ", button("Remove"), " button to remove the currently selected captured pattern or press ", button("Clear"), " button to clear all captured patterns.",
		}},
	}},
}
