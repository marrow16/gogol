package help

var contentPatterns = content{
	"The ", bold("Patterns"), " popout is available from the ", MainMenu.link("main menu"), ".\n\n",
	"This popout shows all patterns currently loaded - see ", LoadPatterns.link("Load Patterns"), " for adding more patterns.\n\n",
	"Select ", "a pattern from the dropdown - you can also type a name into that dropdown to search for pattern(s) by name.\n\n",
	"Having selected a pattern, choose the ", button("Preview"), " option to view a preview of the pattern ",
	"or choose ", button("Metadata"), " option to see the pattern metadata (from the .rle file - and including the filename).\n\n",
	"Press the ", button("Place"), " button to place the pattern onto the grid ", italic("(see "), PlacePattern.linkItalic("Place pattern mode"), italic(")"), ".\n",
	"Check the ", button("Interlaced"), " option to place the pattern interlaced - where dead cells in the pattern do not overwrite alive cells in the grid.\n\n",
	"Choose the ", button("Search/filter"), " option to create filters for the patterns shown in the dropdown:",
	indent{indent: 20, content: content{
		hanging{prefix: "• ", content: content{italic("All filter fields are optional - leaving them empty has no effect.")}},
		hanging{prefix: "• ", content: content{"Enter a ", bold("Rule"), italic(" (e.g. \"B3/S23\")"), " to filter patterns to that rule - or check the ", button("Filter current rule"), " option to filter patterns to the currently active rule."}},
		hanging{prefix: "• ", content: content{"Use ", bold("Width"), " and ", bold("Height"), " entries to filter the patterns to ranges of specific sizes."}},
		hanging{prefix: "• ", content: content{"Use ", bold("Name"), ", ", bold("Filename"), ", ", bold("Origin"), " and ", bold("Comment"), " entries to search/filter pattern metadata."}},
		"Having entered your filter criteria, press the ", button("Apply Filter"), " button to apply - the number of found patterns will be displayed.\n",
		"Use the ", button("Clear Filter"), " button to unapply the filtering.",
	}},
}
