package help

var contentHeatMap = content{
	"Heat map display is activated by pressing ", keys{altMac, "H"},
	" or by using the ", button("Reveal"), " button in ", Instrumentation.link("Instrumentation"), "\n",
	"When ", bold("All"), " heat map type is selected - pressing ", keys{altMac, "H"}, " will cycle through the different heat map displays.\n\n",
	"Heat map display is only available when the ", Instrumentation.link("Heat Map Instrument"), " is enabled.\n\n",
	"Press ", keys{"Esc"}, " to exit the heat map display.\n",
}
