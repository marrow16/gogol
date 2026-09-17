package help

var contentGridRecipes = content{
	"The ", bold("Grid Recipes"), " popout is available from the ", MainMenu.link("main menu"), ".\n\n",
	"Select ", "an existing grid recipe from the dropdown, or enter the path to a recipe .json file. Alternatively, use the ", button("..."), " button to open the ", FileFinder.link("file finder"), ".\n\n",
	"Selecting a recipe file registers it for use by GoGoL. The recipe is not loaded, parsed or validated at this point - the file is only read when the recipe is run.\n",
	"Because recipes are loaded on demand, the recipe file can be edited and re-run without restarting GoGoL.\n\n",
	"Press the ", button("Run"), " button or press ", keys{altMac, "G"}, " to execute the selected recipe. If the recipe cannot be loaded or contains an error, the error will be displayed.\n\n",
	"Use ", button("Save as RLE"), " button to execute the recipe and save the resulting grid as an RLE file.\n\n",
	"For details of the Grid Recipe JSON format and available operations, see the ", GridRecipesReference.link(), ".",
}

var contentGridRecipesReference = content{
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
					"The optional ", code(`grid`), " property configures the overall grid (if this property is not specified, then the current grid settings are used)",
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
							"Note: the ", code(`width`), " and ", code(`height`), " properties are mandatory when ", code(`rle`), " property is specified.\n",
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
					"All properties within a ", code(`do`), " object are optional.",
					indent{content: content{
						"• ", code("do.place"), bold(" property"),
						indent{indent: 20, content: content{
							"The name of the variable or pattern to place - referenced to a name in ", code(`vars`), "/", code(`patterns`), " within in the recipe.\n",
							"If the ", code(`place`), " is omitted or null - nothing happens, but nested ", code(`do`), " instructions are still carried.\n",
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
							"Both the ", code(`x`), " and ", code(`y`), " properties are optional - if not specified, the current position is used.",
							hanging{prefix: "• ", content: content{code("x"), " can also be specified as ", code(`"gw"`), ", ", code(`"gridwidth"`), " or ", code(`"grid-width"`), " to denote right-most position."}},
							hanging{prefix: "• ", content: content{code("y"), " can also be specified as ", code(`"gh"`), ", ", code(`"gridheight"`), " or ", code(`"grid-height"`), " to denote bottom-most position."}},
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
							"If the ", code(`x`), "/", code(`y`), " properties are omitted or null then the current position for that part is unaffected.\n",
							"The ", code(`x`), " and ", code(`y`), " properties are relative to the current position, so can be positive or negative.  They can also be specified as a string - one of the following:",
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
	indent{indent: 20, content: content{
		h4("Draw line around grid"),
		codeBlock{code: `{
	"name": "Line around grid",
	"vars": {
		"h-pattern": "fill-width:0b1",
		"v-pattern": "fill-height:0b1"
	},
	"do": [
		{
			"at": {"x": 0, "y": 0},
			"place": "h-pattern"
		},
		{
			"at": {"x": 0, "y": "gh"},
			"place": "h-pattern"
		},
		{
			"at": {"x": 0, "y": 0},
			"place": "v-pattern"
		},
		{
			"at": {"x": "gw", "y": 0},
			"place": "v-pattern"
		}
	]
}`},
	}},
}
