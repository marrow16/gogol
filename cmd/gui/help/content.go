package help

import (
	"github.com/marrow16/gogol/cmd/gui/icons"
)

var contents = map[Topic]content{
	Index: {
		About, "\n",
		CapturedPatterns, "\n",
		CollectedRules, "\n",
		Colors, "\n",
		Editor, "\n",
		FileFinder, "\n",
		GridRecipes, indent{indent: 20, content: content{GridRecipesReference}},
		HeatMap, "\n",
		ImportGrid, "\n",
		Instrumentation, "\n",
		Keys, "\n",
		LoadPatterns, "\n",
		MainMenu, "\n",
		MetaRules, indent{indent: 20, content: content{MetaRulesRef}},
		Patterns, "\n",
		PlacePattern, "\n",
		Rules, "\n",
		ShortCuts, indent{indent: 20, content: content{ShortCutsRef}},
		SizingWrapping, "\n",
		StatusBar, "\n",
		Stepping, "\n",
	},
	StatusBar: {
		"The status bar is split into three sections:\n",
		hanging{
			prefix: []any{bold("Step/Mode"), " - "},
			content: []any{"shows the current step or the current mode.\n",
				"When simulating, the ", bold("%"), " is the percent of cells that changed",
				" and ", bold("Hz"), " is generations per second."},
		},
		"\n",
		hanging{
			prefix:  []any{bold("Rule"), " - "},
			content: []any{"shows the current rule (click or press ", keys{altMac, "L"}, " to change ", Rules.link("Rule"), ")"},
		},
		"\n", bold("Control buttons:"),
		hanging{indent: 20,
			prefix:  content{iconText{image: icons.Play}, " - "},
			content: content{"start the simulation ", keys{altMac, keyEnter}},
		},
		hanging{indent: 20,
			prefix:  content{iconText{image: icons.Pause}, " - "},
			content: content{"pause the simulation ", keys{altMac, keyEnter}},
		},
		hanging{indent: 20,
			prefix:  content{iconText{image: icons.Step}, " - "},
			content: content{"step the simulation ", keys{altMac, keyRight}},
		},
		hanging{indent: 20,
			prefix:  content{iconText{image: icons.SkipForward}, " - "},
			content: content{"step ahead the simulation ", keys{altMac, keyTab}, " (see ", Stepping, ")"},
		},
		hanging{indent: 20,
			prefix:  content{iconText{image: icons.ZoomIn}, " - "},
			content: content{"zoom in."},
		},
		hanging{indent: 20,
			prefix:  content{iconText{image: icons.ZoomOut}, " - "},
			content: content{"zoom out."},
		},
		hanging{indent: 20,
			prefix:  content{iconText{image: icons.Burger}, " - "},
			content: content{"show the main menu ", keys{altMac, "M"}, " (see ", MainMenu, ")"},
		},
		hanging{indent: 10,
			prefix: content{"Additionally (if ", Instrumentation.link("Record instrument"), " is enabled):"},
		},
		hanging{indent: 20,
			prefix:  content{iconText{image: icons.Backward}, " - "},
			content: content{"step back ", keys{altMac, keyLeft}},
		},
		hanging{indent: 20,
			prefix:  content{iconText{image: icons.SkipBackward}, " - "},
			content: content{"skip backward the simulation ", keys{altMac, keyBack}},
		},
	},
	Editor: {
		"Edit mode allows you to edit the current grid. Start edit mode by pressing ",
		keys{altMac, "E"}, " or selecting ", bold("Edit mode"), " from ", MainMenu.link("main menu"), ".\n\n",
		"To exit edit mode, press ", keys{"Esc"}, " or ", keys{altMac, "E"}, " again.  ",
		"Edit mode is also terminated by opening ", MainMenu.link("main menu"), "; clicking on ", StatusBar.link("statusbar rule"),
		"; or clicking any of the simulation start/step buttons.  The ", iconText{icons.ZoomIn}, " and ", iconText{icons.ZoomOut}, " buttons can be used during edit mode.\n\n",
		"During edit mode, the left panel of ", StatusBar.link("statusbar"), " will show the current edit position - which is also indicated by a blinking cursor in the grid.\n",
		"The following keys can be used to edit the grid:",
		table{
			spaceBefore: 4,
			spaceAfter:  4,
			borders:     true,
			padding:     4,
			columns: []tableColumn{
				{
					width:  30,
					header: "Key(s)",
				},
				{
					header: "Description",
				},
			},
			rows: []tableRow{
				{{keys{"Space"}}, {"Clear cell (dead)"}},
				{{keys{"Shift+Space"}}, {"Set cell (alive)"}},
				{{keys{keyLeft}, " ", keys{keyRight}, " ", keys{keyUp}, " ", keys{keyDown}}, {"Move cell cursor"}},
				{{keys{"Home"}}, {"Move cell cursor to beginning of row"}},
				{{keys{"End"}}, {"Move cell cursor to end of row"}},
				{{keys{"PgUp"}}, {"Move cell cursor to top of grid"}},
				{{keys{"PgDn"}}, {"Move cell cursor to bottom of grid"}},
				{{keys{altMac, keyLeft}, " ", keys{altMac, keyRight}, " ", keys{altMac, keyUp}, " ", keys{altMac, keyDown}}, {"Draw lines"}},
				{{keys{"Shift+", altMac, keyLeft}, " ", keys{"Shift+", altMac, keyRight}, "\n", keys{"Shift+", altMac, keyUp}, " ", keys{"Shift+", altMac, keyDown}}, {"Clear lines"}},
				{{keys{"Shift+", keyLeft}, " ", keys{"Shift+", keyRight}, "\n", keys{"Shift+", keyUp}, " ", keys{"Shift+", keyDown}}, {"Mark area"}},
				{{keys{keyEnter}}, {"Capture marked area as pattern ", italic("(see "), CapturedPatterns.linkItalic("captured patterns"), italic(")")}},
				{{keys{altMac, "C"}}, {"Clear entire grid"}},
				{{keys{altMac, "F"}}, {"Fill marked area with alive cells"}},
				{{keys{"Shift+", altMac, "F"}}, {"Fill marked area with dead cells"}},
				{{keys{altMac, "I"}}, {"Toggle pattern placement interlaced", italic(" (see note below)")}},
				{{keys{altMac, "P"}}, {"Place pattern", italic(" (when "), Patterns.linkItalic("pattern"), italic(" selected)")}},
				{{keys{altMac, "R"}}, {"Pattern placement rotation"}},
				{{keys{altMac, "U"}}, {"Shift entire grid up"}},
				{{keys{altMac, "D"}}, {"Shift entire grid down"}},
				{{keys{altMac, ","}}, {"Shift entire grid left"}},
				{{keys{altMac, "."}}, {"Shift entire grid right"}},
				{{keys{cmdMac, "A"}}, {"Mark entire grid"}},
				{{keys{cmdMac, "C"}}, {"Copy marked area as pattern RLE"}},
				{{keys{cmdMac, "V"}}, {"Paste pattern RLE"}},
				{{keys{cmdMac, "X"}}, {"Cut marked area as pattern RLE"}},
				{{keys{cmdMac, "Z"}}, {"Undo last edit"}},
				{{italic("character keys")}, {"Draw character"}},
			},
		},
		italic("Note: Interlaced means that when placing a pattern, dead cells in the pattern do not overwrite alive cells in the grid."),
	},
	HeatMap: {
		"Heat map display is activated by pressing ", keys{altMac, "H"},
		" or by using the ", button("Reveal"), " button in ", Instrumentation.link("Instrumentation"), "\n",
		"When ", bold("All"), " heat map type is selected - pressing ", keys{altMac, "H"}, " will cycle through the different heat map displays.\n\n",
		"Heat map display is only available when the ", Instrumentation.link("Heat Map Instrument"), " is enabled.\n\n",
		"Press ", keys{"Esc"}, " to exit the heat map display.\n",
	},
	PlacePattern: {
		"Place pattern mode is instigated from the ", Patterns.link("Patterns"), " popout from the ", MainMenu.link("main menu"),
		" or, if a pattern is already selected, by pressing ", keys{altMac, "P"}, ".\n\n",
		"When in pattern place mode, the selected pattern will appear on the grid as a highlighted overlay.\n",
		"Move the pattern to the desired position using keys ", keys{keyLeft}, ", ", keys{keyRight}, ", ", keys{keyUp}, " and ", keys{keyDown}, ".",
		" Use key ", keys{altMac, "R"}, " to alter the rotation.\n\n",
		"Once the desired position and rotation are found, press ", keys{keyEnter}, " to place the pattern on the grid.\n\n",
		italic("Press "), keys{"Esc"}, italic(" to abandon the pattern placement and exit placement mode."),
	},
	Rules: {
		"Use the rules popup to select the currently active rule.\n\n",
		"The rules popup is available by clicking the rule area in the ", StatusBar.link("statusbar"),
		" or by pressing ", keys{altMac, "L"}, ".\n\n",
		"The list at the top shows the named rules - click on an item in that list selects that rule.",
		" You can also navigate around that list using keys ", keys{keyUpTriangle}, " ", keys{keyDownTriangle}, " or by pressing any letter key.\n\n",
		"The bottom of the popup shows details about the currently active rule - which can also be interacted with to change the rule:\n\n",
		bold("Name"),
		indent{indent: 20, spaceAfter: 4, content: content{
			"Shows the name of the current rule.\n",
			"If the current rule is already named this cannot be changed - however, if the current rule is custom, the name can be changed and a ", button("Save"), " button allows the rule name to be saved for future use.",
		}},
		bold("Rule"),
		indent{indent: 20, spaceAfter: 4, content: content{
			"Shows the actual rule string - born (", bold("B"), ") and survives (", bold("S"), ")\n",
			"Edit the rule string to change the rule.", italic(" (see also below for changing the rule by keys)"),
		}},
		bold("Perm."),
		indent{indent: 20, spaceAfter: 4, content: content{
			"Shows the permutation for the rule (0 to 262143) - an 18-bit number where born (", bold("B"), ") is the high order 9-bits and survives (", bold("S"), ") is the low order 9-bits.\n",
			"Change the value to change the rule, or use ", keys{keyUp}, " ", keys{keyDown}, " to increment/decrement. ",
			"Use ", keys{"PgUp"}, " ", keys{"PgDn"}, " to increment/decrement the high order 9-bits.",
		}},
		bold("Integer"),
		indent{indent: 20, spaceAfter: 4, content: content{
			"Is an alternative rule numbering system defined by ", url("https://conwaylife.com/wiki/Rule_integer", "conwaylife.com/wiki"), ".",
			"\n", italic("(GoGoL prefers to use the permutation numbering system because it keeps behaviours together by numeric proximity)"),
		}},
		h4("Changing rule with keys"),
		"The current rule can be altered at any point (without opening the rule popup) by using the following shortcut keys:",
		indent{indent: 20, content: content{
			keys{"Ctrl+", "0"}, " - ", keys{"Ctrl+", "8"}, " toggles the corresponding digit in the born (", bold("B"), ").",
		}},
		indent{indent: 20, content: content{
			keys{altMac, "0"}, " - ", keys{altMac, "8"}, " toggles the corresponding digit in the survives (", bold("S"), ").",
		}},
	},
	MainMenu: {
		"The main menu can be shown by clicking the ", iconText{image: icons.Burger},
		" button on the status bar or by pressing key\u00a0", keys{altMac, "M"}, "\n\n",
		"The menu contains the following options:\n\n",
		bold("Help"), " or press key ", keys{"F1"},
		indent{indent: 20, spaceAfter: 8, content: content{"Show this help screen."}},
		bold("About"),
		indent{indent: 20, spaceAfter: 8, content: content{"Shows application version and information."}},
		bold("Patterns"),
		indent{indent: 20, spaceAfter: 8, content: content{"Shows the ", Patterns.link("Patterns"), " popout."}},
		bold("Captured Patterns"),
		indent{indent: 20, spaceAfter: 8, content: content{"Shows the ", CapturedPatterns.link("Captured Patterns"), " popout."}},
		bold("Load Patterns"),
		indent{indent: 20, spaceAfter: 8, content: content{"Shows the ", LoadPatterns.link("Load Patterns"), " popout."}},
		bold("Instrumentation"),
		indent{indent: 20, spaceAfter: 8, content: content{"Shows the ", Instrumentation.link("Instrumentation"), " popout."}},
		bold("Shortcuts"),
		indent{indent: 20, spaceAfter: 8, content: content{"Shows the ", ShortCuts.link("Shortcuts"), " popout."}},
		bold("Meta Rules"),
		indent{indent: 20, spaceAfter: 8, content: content{"Shows the ", MetaRules.link("Meta Rules"), " popout."}},
		bold("Collected Rules"),
		indent{indent: 20, spaceAfter: 8, content: content{"Shows the ", CollectedRules.link("Collected Rules"), " popout."}},
		bold("Grid Recipes"),
		indent{indent: 20, spaceAfter: 8, content: content{"Shows the ", GridRecipes.link("Grid Recipes"), " popout."}},
		bold("Size/Wrapping/Boundaries"),
		indent{indent: 20, spaceAfter: 8, content: content{"Shows the ", SizingWrapping.link("Size/Wrapping/Boundaries"), " popout."}},
		bold("Stepping"),
		indent{indent: 20, spaceAfter: 8, content: content{"Shows the ", Stepping.link("Stepping"), " popout."}},
		bold("Colors"),
		indent{indent: 20, spaceAfter: 8, content: content{"Shows the ", Colors.link("Colors"), " popout."}},
		bold("Snapshot"), " or press key ", keys{altMac, "S"},
		indent{indent: 20, spaceAfter: 8, content: content{"Takes a snapshot of the current grid."}},
		bold("Undo to snapshot"), " or press key ", keys{altMac, "Z"},
		indent{indent: 20, spaceAfter: 8, content: content{"Restores the grid to the last snapshot."}},
		bold("Export Grid"), " or press key ", keys{altMac, "X"},
		indent{indent: 20, spaceAfter: 8, content: content{"Exports the current grid as a ", code(".rle"), " file in the documents/GoGoL folder."}},
		bold("Import Grid"),
		indent{indent: 20, spaceAfter: 8, content: content{"Shows the ", ImportGrid.link("Import Grid"), " popout."}},
		bold("Edit mode"), " or press key ", keys{altMac, "E"},
		indent{indent: 20, spaceAfter: 8, content: content{"   Places the grid into edit mode.  See ", topicLink{Topic: Editor, Italic: true}}},
		bold("Randomize"), " or press key ", keys{altMac, "R"},
		indent{indent: 20, spaceAfter: 8, content: content{"Clears the current grid and adds random alive cells."}},
		bold("Random Noise"), " or press key ", keys{altMac, "N"},
		indent{indent: 20, spaceAfter: 8, content: content{"Adds random alive cells to the current grid."}},
		bold("Clear"), " or press key ", keys{altMac, "C"},
		indent{indent: 20, spaceAfter: 8, content: content{"Clears all cells in the current grid."}},
	},
	Colors: {
		"Use the ", bold("Colors"), " popout from the ", MainMenu.link("main menu"), " to adjust the colors of cells on the grid.\n\n",
		"Each color setting, ", bold("Alive cells"), ", ", bold("Dead cells"), " & ", bold("Cell border"), ", ",
		"is represented by a ", bold("R"), "(red), ", bold("G"), "(green) & ", bold("B"), "(blue) component - ",
		"enter a value between 0 and 255.\n",
		"Within each color component input, use keys ", keys{keyUpTriangle}, " and ", keys{keyDownTriangle}, " to increment/decrement the value.\n\n",
		"Use the ", bold("Show Borders"), " checkbox to determine whether the grid displays borders.\n",
		"Cell borders can also be toggled by pressing key ", keys{altMac, "B"}, ".",
	},
	SizingWrapping: {
		"Use the ", bold("Sizing/Wrapping/Boundaries"), " popout from the ", MainMenu.link("main menu"), " to adjust various settings on grid sizing and behaviour.\n\n",
		"Use the ", bold("Grid size Width"), " and ", bold("Height"), " inputs to set the desired size for the grid - the grid size is only updated by pressing the ", button("Resize"), " button.\n",
		"Press the ", button("Fit screen"), " button to resize the grid to fill the available screen size (according to current cell size and zoom rate).\n",
		"The ", bold("Keep cells"), " checkbox is used to determine whether the grid alive/dead cells are preserved during resizing.\n\n",
		bold("Wrapping mode"), " determines how cells at the edges of the grid interact with cells on the opposite edges:",
		hanging{indent: 20,
			prefix:  content{bold("• None"), " - "},
			content: content{"grid edges do not wrap."},
		},
		hanging{indent: 20,
			prefix:  content{bold("• Horizontal"), " - "},
			content: content{"left and right edges wrap around to each other."},
		},
		hanging{indent: 20,
			prefix:  content{bold("• Vertical"), " - "},
			content: content{"top and bottom edges wrap around to each other."},
		},
		hanging{indent: 20,
			prefix:  content{bold("• Toroidal"), " - "},
			content: content{"all grid edges connect to their opposite edge."},
		},
		"When grid edges are not wrapped, the ", bold("Boundary mode"), " determines how cells beyond those edges are treated. They may be considered either ", bold("Dead cells"), " or ", bold("Alive cells"), " when calculating the neighbours of cells at the grid edge.\n\n",
		"The ", bold("Cell size"), " input determines the size of each grid cell in pixels (1–32). Updating this value will take immediate effect.\n\n",
		bold("Randomize %"), " input is used to determine the percentage of alive cells when randomizing the grid (from the ", MainMenu.link("main menu"), " or pressing ", keys{altMac, "R"}, ").\n",
		"Note: The randomize % setting only determines probability of alive cell - not overall grid density. ",
		"To specify an overall random density, create a ", ShortCuts.link("shortcut"), " using action ", code("randomize-population"), ".\n",
	},
	Stepping: {
		"Use the ", bold("Stepping"), " popout from the ", MainMenu.link("main menu"), " to adjust the stepping settings.\n\n",
		"The ", bold("Step delay (ms)"), " setting determines the delay between each step (in milliseconds) and can be set to a value between 0 and 2000.\n\n",
		"The ", bold("Step ahead size"), " setting determines the number of steps that are performed with a step ahead - and can be set to a value between 1 and 9999.\n",
		"Step ahead is instigated by pressing key ", keys{altMac, keyTab}, " or clicking ", iconText{image: icons.SkipForward}, " on the ", StatusBar.link("statusbar"), ".\n\n",
		"Use the ", bold("Snapshot on step ahead"), " checkbox to determine whether a snapshot of the grid is automatically taken on every step ahead.\n\n\n",
		italic("Note: The step delay is decoupled from the GUI frame rate - and setting a step delay of 0 (zero) will run the simulation as \"close to the metal\" as possible with visual updates at ~30Fps."),
	},
	CapturedPatterns: {
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
	},
	Patterns: {
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
	},
	LoadPatterns: {
		"The ", bold("Load Patterns"), " popout is available from the ", MainMenu.link("main menu"), ".\n\n",
		"Enter the ", bold("Path"), " to either a pattern file or a directory containing pattern files. Alternatively, use the ", button("..."), " button to open ", FileFinder.link("file finder"), ".\n",
		"Press the ", button("Load"), " button - the number of successfully loaded patterns will be displayed or an error.\n\n",
		"On successful load, the patterns will be available in the ", Patterns.link("Patterns"), " popout (and are also loaded next time you restart GoGoL).",
	},
	ImportGrid: {
		"Use the ", bold("Import Grid"), " popout from the ", MainMenu.link("main menu"), " to load the current grid from a previously exported grid.\n\n",
		"Enter the ", bold("Path"), " for the grid ", code(".rle"), " file to import or use the ", button("..."), " button to open the ", FileFinder.link("file finder"), ".\n",
		"Use the ", bold("Resize grid"), " to determine whether the current grid can be resized to accommodate the imported grid",
		" and press the ", button("Import"), " button\n\n",
		"If the grid is successfully imported, the current grid will be updated - otherwise, an error is displayed.",
	},
	GridRecipes: {
		"The ", bold("Grid Recipes"), " popout is available from the ", MainMenu.link("main menu"), ".\n\n",
		"Select ", "an existing grid recipe from the dropdown, or enter the path to a recipe .json file. Alternatively, use the ", button("..."), " button to open the ", FileFinder.link("file finder"), ".\n\n",
		"Selecting a recipe file registers it for use by GoGoL. The recipe is not loaded, parsed or validated at this point - the file is only read when the recipe is run.\n",
		"Because recipes are loaded on demand, the recipe file can be edited and re-run without restarting GoGoL.\n\n",
		"Press the ", button("Run"), " button or press ", keys{altMac, "G"}, " to execute the selected recipe. If the recipe cannot be loaded or contains an error, the error will be displayed.\n\n",
		"Use ", button("Save as RLE"), " button to execute the recipe and save the resulting grid as an RLE file.\n\n",
		"For details of the Grid Recipe JSON format and available operations, see the ", GridRecipesReference.link(), ".",
	},
	GridRecipesReference: {
		"A Grid Recipe is a JSON document that describes how to construct an initial Game of Life grid.\n\n",
		"Recipes are executed on demand and are not preloaded or cached. A recipe can be edited in any text editor and re-run immediately without restarting GoGoL.\n\n",
		"Grid recipes can be loaded from the ", GridRecipes.link("Grid Recipes"), " popout in the ", MainMenu.link("main menu"), ".\n",
		"Once a recipe is selected in the popout, it can be executed by pressing ", keys{altMac, "G"}, " or by clicking the ", button("Run"), " button in the popout.\n",
		h3("JSON Structure", 8),
		"The overall structure is defined as:",
		codeBlock{code: `{
	"name": "Example",
	"grid": { ... },
	"vars": { ... },
	"patterns": { ... },
	"do": [ ... ]
}`}, italic("Note: Any unknown JSON properties are ignored."),
		separator{},
		indent{
			content: content{
				"• ", code("grid"), bold(" property"),
				indent{indent: 20,
					content: content{
						"The optional `grid` property configures the overall grid (if this property is not specified, then the current grid settings are used)",
						codeBlock{code: `{
	"grid": {
		"height": 100,
		"width": 100,
		"rule": "B3/S23",
		"wrap_mode": "toroidal",
		"boundary_mode": "dead"
	}
}`}, italic("All properties are optional - properties omitted (or null) default to the current grid setting."),
					},
				},
			},
		},
		separator{},
		indent{
			content: content{
				"• ", code("vars"), bold(" property"),
				indent{indent: 20,
					content: content{
						"Variables store numeric values or binary patterns.",
						codeBlock{code: `{
	"vars": {
		"my-pattern": "0b000101",
		"my-num": 10
	}
}`},
						"The property name, e.g. `my-pattern`, denotes the var name for later use.\n",
						"When using a string to denote the initial value:",
						hanging{prefix: "• ", content: content{code(`"0b101"`), " (binary notation) denotes 5 and a specific width (when used as a pattern) of 3 cells."}},

						hanging{prefix: "• ", content: content{code(`"0b000101"`), " (binary notation) denotes 5 and a specific width (when used as a pattern) of 7 cells."}},
						hanging{prefix: "• ", content: content{code(`"0o07"`), " (octal notation) denotes 7 and a specific width (when used as a pattern) of 6 cells."}},
						hanging{prefix: "• ", content: content{code(`"0x0E"`), " (hexadecimal notation) denotes 14 and a specific width (when used as a pattern) of 8 cells."}},
						hanging{prefix: "• ", content: content{code(`"0xE"`), " (hexadecimal notation) denotes 14 and a specific width (when used as a pattern) of 4 cells."}},
						hanging{prefix: "• ", content: content{code(`"127"`), " (decimal notation) denotes 127 and the width (when used as a pattern) will be the minimum number of bits to represent that value."}},
						"Also, the string can be prefixed with various flags:",
						hanging{prefix: "• ", content: content{
							code(`"alt:"`), "|", code(`"alternate:"`), "\n",
							"alternate cells on each successive placement.",
						}},
						hanging{prefix: "• ", content: content{
							code(`"rot:"`), "|", code(`"rotate:"`), "\n",
							"rotate cells on each successive placement according to Y position.",
						}},
						hanging{prefix: "• ", content: content{
							code(`"fw:"`), "|", code(`"fullwidth:"`), "|", code(`"full-width:"`), "\n",
							"the pattern is built to fill the grid width.",
						}},
						hanging{prefix: "• ", content: content{
							code(`"fh:"`), "|", code(`"fullheight:"`), "|", code(`"full-height:"`), "\n",
							"the pattern is built to fill the grid height.",
						}},
						"A variable can be placed on the grid (as the derived binary pattern) by referencing its name in a `place` property, e.g.",
						codeBlock{code: `{
	"do": [
		{
			"place": "my-pattern"
		}
	]
}`},
					},
				},
			},
		},
		separator{},
		indent{
			content: content{
				"• ", code("patterns"), bold(" property"),
				indent{indent: 20,
					content: content{
						"Patterns can be:",
						hanging{prefix: "• ", content: content{boldItalic("External rle files"),
							indent{indent: 20, content: content{
								codeBlock{code: `{
	"patterns": {
		"beacon": {
			"filename": "beacon.rle"
		}
	}
}`},
								italic("Note: relative filenames are relative from the recipe json file.\n"),
								"The recipe will error if the file cannot be found or cannot be decoded.",
							}}}},
						hanging{prefix: "• ", content: content{boldItalic("Inline rle definitions"),
							indent{indent: 20, content: content{
								codeBlock{code: `{
	"patterns": {
		"block": {
			"rle": "2o$2o!",
			"width": 2,
			"height": 2
		}
	}
}`},
								italic("Note: the `width` and `height` properties are mandatory when `rle` property is specified.\n"),
								"The recipe will error if the `rle` cannot be decoded.",
							}}}},
						hanging{prefix: "• ", content: content{boldItalic("Reference to a name in the current pattern library"),
							indent{indent: 20, content: content{
								codeBlock{code: `{
	"patterns": {
		"glider": {
			"name": "Glider"
		}
	}
}`},
								"Or abbreviated form:",
								codeBlock{code: `{
	"patterns": {
		"glider": "Glider"
	}
}`},
								"The recipe will error if the name cannot be found in the current pattern library.",
							}},
						}},
					},
				},
			},
		},
		separator{},
		indent{
			content: content{
				"• ", code("do"), bold(" property"),
				indent{indent: 20,
					content: content{
						"Contains an array of the actual instructions for placing items on the grid.",
						codeBlock{code: `{
	"do": [
		{
			"place": "name",
			"at": { ... },
			"move": { ... },
			"repeat": 1,
			"rotate": 1,
			"do": [ ... ],
			"var_operations": [ ... ]
		},
		...
	]
}`},
						"All properties within a `do` object are optional.",
						indent{content: content{
							"• ", code("do.place"), bold(" property"),
							indent{indent: 20, content: content{
								"The name of the variable or pattern to place - referenced to a name in `vars`/`patterns` within in the recipe).\n",
								"If the `place` is omitted or null - nothing happens, but nested `do` instructions are still carried.\n",
								"If the name does not resolve to a variable or pattern, the recipe will error.",
							}},
						}},
						separator{},
						indent{content: content{
							"• ", code("do.at"), bold(" property"),
							indent{indent: 20, content: content{
								"Defines an absolute placement position on the grid, e.g.",
								codeBlock{code: `"at": {
	"x:": 0,
	"y": 0
}`},
								"Both the `x` and `y` properties are optional - if not specified, the current position is used.",
							}},
						}},
						separator{},
						indent{content: content{
							"• ", code("do.move"), bold(" property"),
							indent{indent: 20, content: content{
								"Defines a relative position to the current position, e.g.",
								codeBlock{code: `"move": {
	"x:": 0,
	"y": 0,
	"when": "before|after"
}`},
								"All properties are optional.\n",
								"If the `x`/`y` properties are omitted or null then the current position for that part is unaffected.\n",
								"The `x`/`y` properties are relative to the current position, so can be positive or negative.  They can also be specified as a string - one of the following:",
								hanging{prefix: "• ", content: content{code(`"lw"`), " is the width of the last placed pattern/var"}},
								hanging{prefix: "• ", content: content{code(`"lh"`), " is the height of the last placed pattern/var"}},
								hanging{prefix: "• ", content: content{code(`"-lw"`), " is the negative width of the last placed pattern/var"}},
								hanging{prefix: "• ", content: content{code(`"-lh"`), " is the negative height of the last placed pattern/var"}},
								"Using the above tokens, the dimension can also be adjusted e.g.",
								hanging{prefix: "• ", content: content{code(`"lw+5"`), " is the last width plus 5"}},
								hanging{prefix: "• ", content: content{code(`"lh++"`), " is the last height plus 1"}},
								"The `when` property controls when the move happens - i.e. before or after the placement.  If omitted or null it assumes before.",
							}},
						}},
						separator{},
						indent{content: content{
							"• ", code("do.repeat"), bold(" property"),
							indent{indent: 20, content: content{
								"Is the number of repeats (after the initial placement).\n",
								"This can also be specified as a string - one of the following:",
								hanging{prefix: "• ", content: content{
									code(`"gh"`), "|", code(`"gridheight"`), "|", code(`"grid-height"`), "\n",
									"repeats to fill the remaining grid height.",
								}},
								hanging{prefix: "• ", content: content{
									code(`"gw"`), "|", code(`"gridwidth"`), "|", code(`"grid-width"`), "\n",
									"repeats to fill the remaining grid width.",
								}},
							}},
						}},
						separator{},
						indent{content: content{
							"• ", code("do.rotate"), bold(" property"),
							indent{indent: 20, content: content{
								"Is the number of 90 degree rotations for the placed pattern.  Values greater than 3 are taken as modulus 4.",
							}},
						}},
						separator{},
						indent{content: content{
							"• ", code("do.do"), bold(" property"),
							indent{indent: 20, content: content{
								"Contains nested instructions.",
							}},
						}},
						separator{},
						indent{content: content{
							"• ", code("do.var_operations"), bold(" property"),
							indent{indent: 20, content: content{
								"Are operations to perform on the currently placed var (and only relevant when placing a var).",
								codeBlock{code: `{
	"place": "my-var",
	...
	"var_operations": [
		{
			"when": "before|after", // defaults to after
			"shift_left": 1,
			"shift_right": 1,
			"rotate-left": 1,
			"rotate-right": 1,
			"increment": 1,
			"decrement": 1
		},
		...
	]
}`},
							}},
						}},
					},
				},
			},
		},

		/*


			## Examples

			### Fill grid with chequered cells
			```json
			{
			  "name": "Chequered",
			  "grid": {
			    "wrap_mode": "toroidal",
			    "boundary_mode": "dead"
			  },
			  "vars": {
			    "pattern": "fill-width:rotate:0b01"
			  },
			  "do": [
			    {
			      "at": {"x": 0, "y": 0},
			      "do": [
			        {
			          "place": "pattern",
			          "repeat": "gh",
			          "move": {"y": 1, "when": "after"}
			        }
			      ]
			    }
			  ]
			}
			```
		*/
		separator{},
		h3("Examples", 8),
		indent{indent: 20, content: content{
			h4("Fill grid with chequered cells"),
			codeBlock{code: `{
	"name": "Chequered",
	"vars": {
		"pattern": "fill-width:rotate:0b01"
	},
	"do": [
		{
			"at": {"x": 0, "y": 0},
			"do": [
				{
					"place": "pattern",
					"repeat": "gh",
					"move": {"y": 1, "when": "after"}
				}
			]
		}
	]
}`},
		}},
		indent{indent: 20, content: content{
			h4("Fill grid with horizontal stripes"),
			codeBlock{code: `{
	"name": "Stripes",
	"vars": {
		"pattern": "fill-width:0b1"
	},
	"do": [
		{
			"at": {"x": 0, "y": 0},
			"do": [
				{
					"place": "pattern",
					"repeat": "gh",
					"move": {"y": 2, "when": "after"}
				}
			]
		}
	]
}`},
		}},
		indent{indent: 20, content: content{
			h4("Fill grid with vertical stripes"),
			codeBlock{code: `{
	"name": "Vertical Stripes",
	"vars": {
		"pattern": "fill-height:0b1"
	},
	"do": [
		{
			"at": {"x": 0, "y": 0},
			"do": [
				{
					"place": "pattern",
					"repeat": "gw",
					"move": {"x": 2, "when": "after"}
				}
			]
		}
	]
}`},
		}},
		indent{indent: 20, content: content{
			h4("Fill grid with lattice"),
			codeBlock{code: `{
	"name": "Lattice",
	"vars": {
		"h-pattern": "fill-width:0b1",
		"v-pattern": "fill-height:0b1"
	},
	"do": [
		{
			"at": {"x": 0, "y": 0},
			"do": [
				{
					"place": "h-pattern",
					"repeat": "gh",
					"move": {"y": 2, "when": "after"}
				}
			]
		},
		{
			"at": {"x": 0, "y": 0},
			"do": [
				{
					"place": "v-pattern",
					"repeat": "gw",
					"move": {"x": 2, "when": "after"}
				}
			]
		}
	]
}`},
		}},
	},
	Instrumentation: {
		"The ", bold("Instrumentation"), " popout is accessed from the ", MainMenu.link("main menu"), ".\n",
		"Instrumentation provides tools for analysing and recording the behaviour of the grid. Each instrument can be independently enabled or disabled using its checkbox.\n\n",
		"The following tools are available:\n\n",
		bold("Repeat Detect"), "\n",
		"When enabled, monitors successive generations looking for a grid state that has occurred previously. When a repeat is found, ", bold("First"), " and ", bold("Repeat"), " show the generations containing the matching states, and ", bold("Period"), " shows the number of generations between them.\n",
		"Also, when a repeat is detected for the first time, the running simulation will stop and the repeat found is notified on the ", StatusBar.link("statusbar"),
		hanging{indent: 10, gap: 10,
			prefix:  content{button("Reset")},
			content: content{"clears the current repeat detection history."},
		},
		hanging{indent: 10, gap: 10,
			prefix:  content{button("Save Report")},
			content: content{"saves a report of the detected repeat as a ", code(".json"), " file in the documents/GoGoL folder."},
		},
		separator{spaceBefore: 8, spaceAfter: 8},
		bold("Record"), "\n",
		"When enabled, retains previous generations, allowing the simulation to be stepped backwards and for animations to be saved.",
		hanging{indent: 10, gap: 10,
			prefix:  content{bold("Steps recorded")},
			content: content{"shows the number of generations currently retained."},
		},
		hanging{indent: 10, gap: 10,
			prefix: content{bold("Skip back by")},
			content: content{"determines how many generations are skipped when using the skip-back control.\n",
				"Press ", keys{altMac, keyBack}, " or use ", iconText{image: icons.SkipBackward}, " button on ", StatusBar.link("statusbar")},
		},
		hanging{indent: 10, gap: 10,
			prefix:  content{button("Reset")},
			content: content{"to discard the current recorded history."},
		},
		hanging{indent: 10, gap: 10,
			prefix:  content{button("Save Animation")},
			content: content{" to save the recorded steps as an animation - ", bold("Gif"), " or ", bold("Mp4"), " file."},
		},
		italic("Note: The Mp4 option is only available when "), code("ffmpeg"), italic(" is installed."), " [see ", url("https://ffmpeg.org/"), "]\n",
		italic("WARNING: Saving animations above 1000 or so steps can produce very large files!"),
		separator{spaceBefore: 8, spaceAfter: 8},
		bold("Heat Mapping"), "\n",
		"Accumulates cell activity over successive generations that can be visualised as a heat map image.\n",
		bold("Type"), " determines the information accumulated:",
		hanging{indent: 20,
			prefix:  content{bold("• Activity"), " - "},
			content: content{"how frequently each cell changes state."},
		},
		hanging{indent: 20,
			prefix:  content{bold("• Occupancy"), " - "},
			content: content{"how many successive generations a cell has survived."},
		},
		hanging{indent: 20,
			prefix:  content{bold("• Births"), " - "},
			content: content{"how frequently a cell becomes alive."},
		},
		hanging{indent: 20,
			prefix:  content{bold("• Freshness"), " - "},
			content: content{"how recently a cell became alive."},
		},
		hanging{indent: 20,
			prefix:  content{bold("• Phase Parity"), " - "},
			content: content{"distinguishes activity occurring on odd and even generations."},
		},
		hanging{indent: 20,
			prefix:  content{bold("• All"), " - "},
			content: content{"collects all the above heat map measurements."},
		},
		hanging{indent: 10, gap: 10,
			prefix:  content{button("Reset")},
			content: content{"to clear the accumulated heat map data."},
		},
		hanging{indent: 10, gap: 10,
			prefix: content{button("Reveal")},
			content: content{"to display the heat map image.\n",
				"Or press ", keys{altMac, "H"}, ". See also ", HeatMap.link("heat map display")},
		},
		hanging{indent: 10, gap: 10,
			prefix:  content{button("Save Image")},
			content: content{" to save the heat map image as a \", code(\".png\"), \" file in the documents/GoGoL folder."},
		},
	},
	ShortCuts: {
		"Use the ", bold("Shortcuts"), " popout from the ", MainMenu.link("main menu"), " to create or edit keyboard shortcuts.\n\n",
		"A shortcut can be assigned to any ", bold("Key"), " and is activated using ", keys{altMac}, " combined with your chosen key.\n\n",
		"Each key has ", bold("Actions"), " assigned to it - see ", ShortCutsRef.link("reference"), " for available actions.\n\n",
		italic("Note: Shortcuts "), boldItalic("do not"), italic(" prohibit you assigning keys already assigned by GoGoL!"),
	},
	ShortCutsRef: {
		"This page shows describes all of the shortcut actions that can be used to create ", ShortCuts.link("shortcuts"),
		table{
			borders: true,
			padding: 4,
			columns: []tableColumn{
				{
					width:  30,
					header: "Action",
				},
				{
					header: "Description",
				},
			},
			rows: []tableRow{
				{{codeCopyable("borders:bool")}, {"Sets the current grid cell borders on/off (", code("true"), "|", code("false"), ")"}},
				{{codeCopyable("boundary-mode:mode")}, {"Sets the current grid boundary ", italic("mode"), " (", code(`Dead`), "|", code(`Alive`), ")"}},
				{{codeCopyable("cell-color-alive:color")}, {"Sets the current grid cell alive ", italic("color"), " (as HTML hex color, ", code(`#rrggbb`), ")"}},
				{{codeCopyable("cell-color-dead:color")}, {"Sets the current grid cell dead ", italic("color"), " (as HTML hex color, ", code(`#rrggbb`), ")"}},
				{{codeCopyable("cell-color-border:color")}, {"Sets the current grid cell border ", italic("color"), " (as HTML hex color, ", code(`#rrggbb`), ")"}},
				{{codeCopyable("cell-size:n")}, {"Sets the current grid cell size to ", italic("n"), " (3-100)"}},
				{{codeCopyable("clear")}, {"Clears the current grid"}},
				{{codeCopyable("export")}, {"Exports the current grid as RLE"}},
				{{codeCopyable("export-image")}, {"Exports the current grid as PNG"}},
				{{codeCopyable("grid-height:n")}, {"Sets the current grid height to ", italic("n"), " (2-1000)"}},
				{{codeCopyable("grid-size:wXh")}, {"Sets the current grid size to ", italic("w"), " (width) X ", italic("h"), " (height)"}},
				{{codeCopyable("grid-width")}, {"Sets the current grid width to ", italic("n"), " (2-1000)"}},
				{{codeCopyable("heat-map")}, {"Turns on the heat map instrument (", code(`Activity`), ")"}},
				{{codeCopyable("heat-map:type")}, {"Turns on the heat map instrument with ", italic("type"), " (", code(`Activity`), "|", code(`Occupancy`), "|", code(`Births`), "|", code(`Freshness`), "|", code(`Phase Parity`), "|", code(`All`), ")\n", italic("(other values will turn heat mapper off)")}},
				{{codeCopyable("heat-map-reveal")}, {"Reveals (shows) the current heat map"}},
				{{codeCopyable("heat-map-save")}, {"Saves the current heat map as a .png"}},
				{{codeCopyable("log:format")}, {"Logs a message to ", code(`output.log`), " file - `format` is the same as `name:format`"}},
				{{codeCopyable("max-adjacents:n")}, {"Alters the current grid so that no cell can have more than ", italic("n"), " alive neighbours"}},
				{{codeCopyable("name:format")}, {"Sets the name for output files (e.g. exports, heat map images)\nThe name can be any string that is a valid filename/filepath - plus any of the following format tokens:",
					hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%rule"), " e.g. transposed to `B2/S23`"}},
					hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%born"), " e.g. transposed to `B2`"}},
					hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%survives"), " e.g. transposed to `S23`"}},
					hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%perm"), " e.g. transposed to `4108`"}},
					hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%rand"), " transposed to current randomization setting"}},
					hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%now"), " transposed to a date/time stamp"}},
					hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%r"), " transposed to repeat iteration(s)"}},
					hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%R"), " transposed to last (current) repeat iteration"}},
					hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%step"), " transposed to the current step of the grid"}},
					hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%population"), " transposed to the current population of the grid"}},
					hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%repeat-found"), " transposed to whether a repeat was found (`true`|`false`)"}},
					hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%repeat-first"), " ", codeCopyable("%repeat-at"), " and ", codeCopyable("%repeat-period"), " transposed to repeat information"}}},
				},
				{{codeCopyable("random-changes")}, {"Applies random changes (noise) to the current grid"}},
				{{codeCopyable("randomization++")}, {"Increments the randomization setting"}},
				{{codeCopyable("randomization--")}, {"Decrements the randomization setting"}},
				{{codeCopyable("randomization:n")}, {"Sets the randomization setting to ", italic("n"), " (0-100)"}},
				{{codeCopyable("randomize")}, {"Randomize the current grid (according to current randomization)"}},
				{{codeCopyable("randomize-population")}, {"Randomize the population of the current grid (according to current randomization)\n", italic("Note: This differs from `randomize` because that randomizes each cell in isolation - therefore it is not at all guaranteed that the final grid will have the specified % density. This action ensures that the overall grid density matches the specified randomization setting.")}},
				{{codeCopyable("record")}, {"Turns the record instrument on"}},
				{{codeCopyable("record:1|0")}, {"Turns the record instrument on or off"}},
				{{codeCopyable("record-animation-save")}, {"Saves the current recording as an animation ", italic("(only works when recording is on - e.g. `record:1`)")}},
				{{codeCopyable("record-animation-format:gif|mp4")}, {"Sets the current recording animation format"}},
				{{codeCopyable("repeat-detect")}, {"Turns the repeat detect instrument on"}},
				{{codeCopyable("repeat-detect:1|0")}, {"Turns the repeat detect instrument on or off"}},
				{{codeCopyable("repeat-detect-save")}, {"Saves the current repeat detection (if enabled) as a report json"}},
				{{codeCopyable("repeat:n,...,...")}, {"Repeats the following actions (comma separated) ", italic("n"), " times\nExample:",
					codeBlock{code: "repeat:100,randomize,step-ahead-by:500,heat-map-save"}, "Repeats can also be 'nested', example:",
					codeBlock{code: "repeat:5,rule-perm++,randomize,heat-map:Activity,repeat:10,step-ahead-by:10,heat-map-save`"}}},
				{{codeCopyable("replay-snapshot")}, {"Replays a snapshot\nResets the grid to the last snapshot (without removing the snapshot from the stack)\n", italic("Note: If there is no current snapshot - the grid is randomized and a snapshot added")}},
				{{codeCopyable("rule-born-with:x")}, {"Sets or adjusts the current life rule born with - where ", italic("x"), " can be:",
					hanging{indent: 10, gap: 10, prefix: "•", content: "`23` sets to `B23`"},
					hanging{indent: 10, gap: 10, prefix: "•", content: "`|23` ORs with 2 and 3 (effectively turning just those on)"},
					hanging{indent: 10, gap: 10, prefix: "•", content: "`&23` ANDs with 2 and 3"},
					hanging{indent: 10, gap: 10, prefix: "•", content: "`!23` flips 2 and 3"},
					hanging{indent: 10, gap: 10, prefix: "•", content: "`!` flips all - e.g. `B23` becomes `B0145678`"},
					italic("Note: the current survives with, `S`, is unchanged")}},
				{{codeCopyable("rule-born-with++")}, {"Increments the current life rule born with permutation (high order 9 bits)"}},
				{{codeCopyable("rule-born-with--")}, {"Decrements the current life rule born with permutation (high order 9 bits)"}},
				{{codeCopyable("rule-perm++")}, {"Increments the current life rule permutation"}},
				{{codeCopyable("rule-perm--")}, {"Decrements the current life rule permutation"}},
				{{codeCopyable("rule-perm:n")}, {"Sets the current life rule permutation to ", italic("n"), " (0-262143)"}},
				{{codeCopyable("rule-int++")}, {"Increments the current life rule integer"}},
				{{codeCopyable("rule-int--")}, {"Decrements the current life rule integer"}},
				{{codeCopyable("rule-int:n")}, {"Sets the current life rule integer to ", italic("n"), " (0-262143)"}},
				{{codeCopyable("rule-survives-with:x")}, {"Sets or adjusts the current life rule survives with - where ", italic("x"), " can be:",
					hanging{indent: 10, gap: 10, prefix: "•", content: "`23` sets to `S23`"},
					hanging{indent: 10, gap: 10, prefix: "•", content: "`|23` ORs with 2 and 3 (effectively turning just those on)"},
					hanging{indent: 10, gap: 10, prefix: "•", content: "`&23` ANDs with 2 and 3"},
					hanging{indent: 10, gap: 10, prefix: "•", content: "`!23` flips 2 and 3"},
					hanging{indent: 10, gap: 10, prefix: "•", content: "`!` flips all - e.g. `S23` becomes `S0145678`"},
					italic("Note: the current born with, `B`, is unchanged")}},
				{{codeCopyable("rule-survives-with++")}, {"Increments the current life rule survives with permutation (low order 9 bits)"}},
				{{codeCopyable("rule-survives-with--")}, {"Decrements the current life rule survives with permutation (low order 9 bits)"}},
				{{codeCopyable("rule:name|rle")}, {"Sets the current life rule (by ", italic("name"), " or specified ", italic("rle"), ", e.g. `B2/S23`)"}},
				{{codeCopyable("run-recipe")}, {"Runs the currently selected grid recipe"}},
				{{codeCopyable("run-recipe:filename")}, {"Runs the grid recipe for the specified ", italic("filename")}},
				{{codeCopyable("run")}, {"Runs simulation on the current grid"}},
				{{codeCopyable("sleep:ms")}, {"Sleeps for a given ", italic("ms"), " (millisecond) period"}},
				{{codeCopyable("snapshot")}, {"Snapshots the current grid"}},
				{{codeCopyable("step-ahead++")}, {"Increments the current step ahead setting"}},
				{{codeCopyable("step-ahead--")}, {"Decrements the current step ahead setting"}},
				{{codeCopyable("step-ahead")}, {"Performs a step ahead (by current step ahead setting)"}},
				{{codeCopyable("step-ahead:n")}, {"Sets the current step ahead setting to ", italic("n"), " (1-9999)"}},
				{{codeCopyable("step-ahead-by:n")}, {"Performs a step ahead by ", italic("n"), " (>0)"}},
				{{codeCopyable("step-back-by:n")}, {"Performs a step back by ", italic("n"), " (>0) ", italic("(only works when recording is on - e.g. `record:1`)")}},
				{{codeCopyable("step-delay++")}, {"Increments the current step delay setting"}},
				{{codeCopyable("step-delay--")}, {"Decrements the current step delay setting"}},
				{{codeCopyable("step-delay:ms")}, {"Sets the current step delay to ", italic("ms"), " (0-2000)"}},
				{{codeCopyable("step")}, {"Performs a single simulation step on the current grid"}},
				{{codeCopyable("stop")}, {"Stops simulation on the current grid"}},
				{{codeCopyable("undo-to-snapshot")}, {"Restores grid to last snapshot"}},
				{{codeCopyable("wrap-mode:mode")}, {"Sets the current grid wrap ", italic("mode"), " (", code(`none`), "|", code(`horizontal`), "|", code(`vertical`), "|", code(`all`), "|", code(`toroidal`), ")"}},
				{{codeCopyable("next-meta-rule:meta")}, {"Skips to the next rule in `meta` and sets the current rule to that (see ", MetaRules.link("meta rules"), ")"}},
				{{codeCopyable("previous-meta-rule:meta")}, {"Skips to the previous rule in `meta` and sets the current rule to that (see ", MetaRules.link("meta rules"), ")"}},
				{{codeCopyable("iterate-meta-rule:meta")}, {"Iterates over a meta rule (similar to `repeat:n`)\n", italic("Note: If the meta rule contains commas, it should enclosed in single or double quotes.")}},
				{{codeCopyable("add-collected-rule")}, {"Adds the current grid rule to the collected rules (see ", CollectedRules.link("collected rules"), ")"}},
				{{codeCopyable("remove-collected-rule")}, {"Removes the current grid rule from the collected rules"}},
				{{codeCopyable("next-collected-rule")}, {"Skips to the next rule in collected rules"}},
				{{codeCopyable("previous-collected-rule")}, {"Skips to the previous rule in collected rules"}},
				{{codeCopyable("iterate-collected-rules")}, {"Iterates over collected rules"}},
			},
			spaceBefore: 8,
		},
	},
	MetaRules: {
		"The ", bold("Meta Rules"), " popout is available from the ", MainMenu.link("main menu"), ".\n\n",
		"Select an existing ", "meta rule from the dropdown, or enter a new name to create a new meta rule.\n\n",
		"When a meta rule is selected, choose the ", button("Edit"), " radio button to view or edit the meta rule definition. Changes to the definition are reflected in the number of ", bold("Matched rules"), " shown at the bottom of the popout.\n",
		"If there is an error in the meta rule definition - it is displayed below the edit area.\n\n",
		"Select the ", button("Matching Rules"), " radio button to display the rules matched by the current meta rule. Clicking a rule in the list sets it as the current rule.\n\n",
		"Use the ", button("Delete"), " button to remove the selected meta rule.\n\n",
		"For details of the meta rule language and syntax, see the ", MetaRulesRef.link(), ".",
	},
	MetaRulesRef: {
		"The following provides the normative reference for Meta Rules syntax:",
		indent{indent: 20, content: content{
			codeBlock{code: `meta-rule     = predicates

predicates    = predicate ["," predicate]...
predicate     = rule | all-of | any-of | none-of | one-of

all-of        = "AllOf(" predicates ")"
any-of        = "AnyOf(" predicates ")"
none-of       = "NoneOf(" predicates ")"
one-of        = "OneOf(" predicates ")"

rule          = rule-part ["/" rule-part]...
rule-part     = birth
              | survival
              | and-rule
              | or-rule
              | xor-rule
              | permutation
birth         = "B(" conditions ")"
survival      = "S(" conditions ")"
and-rule      = ("A(" | "&(") conditions ")"
or-rule       = ("O(" | "|(") conditions ")"
xor-rule      = ("X(" | "^(") conditions ")"
permutation   = "P(" ranges ")"

conditions    = condition ["," condition]...
condition     = required
              | forbidden
              | excluded-combination
              | cardinality

required             = "+" digits
forbidden            = "!" digits
excluded-combination = "-" digits
cardinality          = "#" integer ".." integer ":" digits

digits               = digit...
digit                = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8"

ranges               = range ["," range]...
range                = integer | integer "-" integer`},
		}},
		indent{indent: 20, content: content{
			"Notes:",
			hanging{gap: 10, indent: 10,
				prefix:  "•",
				content: content{"Tokens are case-insensitive."},
			},
			hanging{gap: 10, indent: 10,
				prefix:  "•",
				content: content{"Whitespace may appear freely between tokens."},
			},
			hanging{gap: 10, indent: 10,
				prefix:  "•",
				content: content{"Within a rule, ", code("B(...)"), ", ", code("S(...)"), ", ", code("A(...)"), ", ", code("O(...)"), ", ", code("X(...)"), " and ", code("P(...)"), " may each appear at most once."},
			},
			hanging{gap: 10, indent: 10,
				prefix:  "•",
				content: content{"Digits within a condition are unordered."},
			},
			hanging{gap: 10, indent: 10,
				prefix:  "•",
				content: content{"Duplicate digits are not ignored - and will cause errors if repeated."},
			},
			hanging{gap: 10, indent: 10,
				prefix:  "•",
				content: content{"Integers can be specified in base 10, binary (e.g. ", code("0b11111111"), ") octal (e.g. ", code("0o377"), ") or hexadecimal (e.g. ", code("0xFF"), ")"},
			},
			hanging{gap: 10, indent: 10,
				prefix:  "•",
				content: content{"Line comments ", code(`//`), " are allowed - block comments ", code(`/*`), " are not allowed"},
			},
			separator{},
			h4("Conditions", 8),
			header{level: 5, text: "required", mono: true},
			indent{indent: 20, content: content{
				"Example:",
				codeBlock{code: `B(+23)`},
				"Means that both ", code(`2`), " and ", code(`3`), " must be present in ", bold("B"), ".",
			}},
			header{level: 5, text: "forbidden", mono: true, spaceBefore: 8},
			indent{indent: 20, content: content{
				"Example:",
				codeBlock{code: `B(!45)`},
				"Means that both ", code(`4`), " and ", code(`5`), " must not be present in ", bold("B"), ".",
			}},
			header{level: 5, text: "excluded-combination", mono: true, spaceBefore: 8},
			indent{indent: 20, content: content{
				"Example:",
				codeBlock{code: `S(-37)`},
				"Means that ", code(`3`), " and ", code(`7`), " must not appear together in ", bold("S"), ".",
			}},
			header{level: 5, text: "cardinality", mono: true, spaceBefore: 8},
			indent{indent: 20, content: content{
				"Example:",
				codeBlock{code: `S(#2..3:0123)`},
				"Means that, within ", bold("S"), ", of digits ", code(`0`), " ", code(`1`), ",", code(`2`), " and ", code(`3`), " only 2 to 3 of them can be present.\n",
				"Cardinality can also be used to ensure that ", bold("B"), " or ", bold("S"), " are empty, example:",
				codeBlock{code: `B(#0..0:012345678)`},
			}},
			header{level: 5, text: "A(conditions)", mono: true, spaceBefore: 8},
			indent{indent: 20, content: content{
				"ANDs the high-order 9 bits (birth) with the low order 9 bits (survives) and evaluates conditions on the resultant.\n",
				"Example:",
				codeBlock{code: `A(+2)`},
				"Means that ", code(`2`), " must be present in both ", bold("B"), " and ", bold("S"), ".\n",
				"Example:",
				codeBlock{code: `A(!4)`},
				"Means that ", code(`4`), " may be present in ", bold("B"), " or ", bold("S"), " - but not both.",
			}},
			header{level: 5, text: "O(conditions)", mono: true, spaceBefore: 8},
			indent{indent: 20, content: content{
				"ORs the high-order 9 bits (birth) with the low order 9 bits (survives) and evaluates conditions on the resultant.\n",
				"Example:",
				codeBlock{code: `O(+3)`},
				"Means that ", code(`3`), " must be present in either ", bold("B"), " or ", bold("S"), ".",
			}},
			header{level: 5, text: "X(conditions)", mono: true, spaceBefore: 8},
			indent{indent: 20, content: content{
				"XORs the high-order 9 bits (birth) with the low order 9 bits (survives) and evaluates conditions on the resultant.\n",
				"Example:",
				codeBlock{code: `X(!4)`},
				"Means that ", code(`4`), " is either present in both ", bold("B"), " and ", bold("S"), " - or neither.",
			}},
			header{level: 5, text: "P(ranges)", mono: true, spaceBefore: 8},
			indent{indent: 20, content: content{
				"Example:",
				codeBlock{code: `P(512-1024,1234,5678)`},
				"Means that only permutations 512 to 1024 (inclusive); 1234 and 5678 are permitted.",
			}},
			h4("Example", 16),
			codeBlock{code: `AllOf(
    B(!2345) / S(+47,!3),
    AnyOf(
        B(+0,!1) / S(!012,-568),
        B(+0,!16) / S(+568,!012),
        B(+07,!1) / S(+568,!012),
        B(+17,!06) / S(+5,-12),
        B(+017,!6) / S(+5,-12,-68),
        B(+1,!067) / S(+5,!0,-12),
        B(+1,!067) / S(+5,!1,-12),
        B(+1,!067) / S(+56,-12),
        B(+01,!678) / S(+5,!0,-12,-68),
        B(+01,!678) / S(+5,!1,-12,-68),
        B(+018,!67) / S(+5,!0,-12,-68),
        B(+018,!67) / S(+5,!1,-12,-68),
        B(+018,!67) / S(+56,!8,-12),
        B(+018,!67) / S(+58,!6,-12)
    )
)`}, "Is a meta-rule that describes one family of Life-like rules (222) that produce worlds dominated by stable thick plus patterns...",
			Image{name: "plus-pattern.png"},
			codeBlock{code: `#C Copied from GoGoL
x = 6, y = 6, rule = B0/S467
6b$2b2o$b4o$b4o$2b2o$6b!`},
		}},
	},
	CollectedRules: {
		"The ", bold("Collected Rules"), " popout is available from the ", MainMenu.link("main menu"), ".\n\n",
		"Collected Rules are a list of curated rules that are of interest - the list shows the currently collected rules. ",
		"Clicking on an item in the list sets the current rule.\n\n",
		"If the current rule is already in the list it will be highlighted - and you can use the ", button("Remove Current"), " button to remove it from the list.",
		" Conversely, if the current rule is not in the list, use the ", button("Add Current"), " button to add it to the list.\n\n",
		"Use the ", button("Clear"), " button to clear the entire current list.\n\n",
		"The ", bold("Commonality"), " shows the meta rule representing the common properties of the collected rules ", italic("(see "), MetaRulesRef.linkItalic("meta rules"), italic(")"),
		" - along with a count of current collected rules against a count of rules matched by the commonality.\n",
		bold("Commonality"), " is updated every time a rule is added or removed from the list.\n\n",
		"You can edit the ", bold("Commonality"), " - and, after editing, hitting ", keys{keyEnter}, " will update the list with rules that match that meta rule.",
	},
	About: {
		"GoGoL is an interactive cellular automata explorer, built around Conway's Game of Life (",
		url("https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life", "wikipedia"), ") and the wider family of Life-like rules.\n\n",
		"It provides tools for creating, running and analysing cellular automata; experimenting with rules and patterns; recording and identifying patterns; and visualising behaviour through instrumentation and heat maps.\n\n",
		"GoGoL is intended as much for ", bold("experimentation and discovery"), " as for simply watching cells live and die.\n\n",
		"Author: Martin \"Marrow\" Rowlinson\n",
		"Repository: ", url("https://github.com/marrow16/gogol"), "\n",
		"Implemented using ", url("https://go.dev/", "Go"), italic(" (for performance and maintainability)"),
		" and ", url("https://gioui.org/", "Gio UI"), ".\n",
	},
	Keys: {
		"This help provides reference for key shortcuts:\n",
		italic("(see also "), Editor.linkItalic("edit mode"), italic(" for editing keys)"),
		table{
			borders:     true,
			indent:      0,
			padding:     4,
			spaceBefore: 4,
			columns: []tableColumn{
				{
					width:  20,
					header: "Key(s)",
				},
				{
					width:  65,
					header: "Description",
				},
				{
					header: "Alternative",
				},
			},
			rows: []tableRow{
				{{keys{altMac, keyEnter}}, {"Start/stop the simulation"}, {iconText{image: icons.Play}, " ", iconText{image: icons.Pause}}},
				{{keys{altMac, keyRight}, " ", keys{altMac, "Space"}}, {"Step the simulation"}, {iconText{image: icons.Step}}},
				{{keys{altMac, keyTab}}, {"Step ahead the simulation"}, {iconText{image: icons.SkipForward}}},
				{{keys{altMac, keyLeft}}, {"Step back the simulation", italic(" (if "), Instrumentation.linkItalic("record"), italic(" enabled)")}, {iconText{image: icons.Backward}}},
				{{keys{altMac, keyBack}}, {"Skip back the simulation", italic(" (if "), Instrumentation.linkItalic("record"), italic(" enabled)")}, {iconText{image: icons.SkipBackward}}},
				{{keys{altMac, "="}}, {"Zoom in"}, {iconText{image: icons.ZoomIn}}},
				{{keys{altMac, "-"}}, {"Zoom out"}, {iconText{image: icons.ZoomOut}}},
				{{keys{altMac, "B"}}, {"Toggle cell borders on/off"}},
				{{keys{altMac, "C"}}, {"Clear grid"}},
				{{keys{altMac, "E"}}, {"Edit mode"}},
				{{keys{altMac, "G"}}, {"Run grid recipe", italic(" (when "), GridRecipes.linkItalic("recipe"), italic(" selected)")}},
				{{keys{altMac, "H"}}, {"Show heat map", italic(" (when "), Instrumentation.linkItalic("heat mapping"), italic(" enabled)")}},
				{{keys{altMac, "L"}}, {"Life rule editor"}, {"click ", StatusBar.link("statusbar rule")}},
				{{keys{altMac, "M"}}, {"Menu"}, {iconText{image: icons.Burger}}},
				{{keys{altMac, "N"}}, {"Random noise on grid"}},
				{{keys{altMac, "P"}}, {"Place pattern mode", italic(" (when "), Patterns.linkItalic("pattern"), italic(" selected)")}},
				{{keys{altMac, "R"}}, {"Randomize grid"}},
				{{keys{altMac, "S"}}, {"Snapshot grid"}},
				{{keys{altMac, "X"}}, {"Export grid"}},
				{{keys{altMac, "Z"}}, {"Undo to snapshot"}},
				{{keys{altMac, ","}}, {"Decrement life rule permutation"}},
				{{keys{altMac, "."}}, {"Increment life rule permutation"}},
				{{keys{altMac, "["}}, {"Decrease grid width"}},
				{{keys{altMac, "]"}}, {"Increase grid width"}},
				{{keys{altMac, ";"}}, {"Decrease grid height"}},
				{{keys{altMac, "'"}}, {"Increase grid height"}},
				{{keys{"Ctrl+0"}, " - ", keys{"Ctrl+8"}}, {"Toggle rule born with"}},
				{{keys{altMac, "0"}, " - ", keys{altMac, "8"}}, {"Toggle rule survives with"}},
				{{keys{"F1"}}, {"To show this help screen."}},
			},
		},
	},
	FileFinder: {
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
	},
}
