package help

var contentCollectedRules = content{
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
}
