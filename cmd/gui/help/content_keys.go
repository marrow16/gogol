package help

import "github.com/marrow16/gogol/cmd/gui/icons"

var contentKeys = content{
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
}
