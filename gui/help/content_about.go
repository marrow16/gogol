package help

var contentAbout = content{
	"GoGoL is an interactive cellular automata explorer, built around Conway's Game of Life (",
	url("https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life", "wikipedia"), ") and the wider family of Life-like rules.\n\n",
	"It provides tools for creating, running and analysing cellular automata; experimenting with rules and patterns; recording and identifying patterns; and visualising behaviour through instrumentation and heat maps.\n\n",
	"GoGoL is intended as much for ", bold("experimentation and discovery"), " as for simply watching cells live and die.\n\n",
	"Author: Martin \"Marrow\" Rowlinson\n",
	"Repository: ", url("https://github.com/marrow16/gogol"), "\n",
	"Implemented using ", url("https://go.dev/", "Go"), italic(" (for performance and maintainability)"),
	" and ", url("https://gioui.org/", "Gio UI"), ".\n",
}
