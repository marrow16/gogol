package help

var contentRules = content{
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
	button("Add Collected"), " / ", button("Remove Collected"),
	indent{indent: 20, spaceAfter: 8, content: content{
		"Use these buttons to add or removed the current rule to/from ", CollectedRules.link("collected rules"), ".\n",
		"Which button displayed indicates whether the current rule is currently collected.",
	}},
	h4("Changing rule with keys"),
	"The current rule can be altered at any point (without opening the rule popup) by using the following shortcut keys:",
	indent{indent: 20, content: content{
		keys{"Ctrl+", "0"}, " - ", keys{"Ctrl+", "8"}, " toggles the corresponding digit in the born (", bold("B"), ").",
	}},
	indent{indent: 20, content: content{
		keys{altMac, "0"}, " - ", keys{altMac, "8"}, " toggles the corresponding digit in the survives (", bold("S"), ").",
	}},
}
