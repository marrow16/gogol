package help

import "github.com/marrow16/gogol/cmd/gui/icons"

var contentEditor = content{
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
}
