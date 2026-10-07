package help

import (
	"github.com/marrow16/gogol/gui/icons"
)

var (
	iconBackward     = icon{alt: "backward", image: icons.Backward}
	iconBurger       = icon{alt: "burger", image: icons.Burger}
	iconPause        = icon{alt: "pause", image: icons.Pause}
	iconPlay         = icon{alt: "play", image: icons.Play}
	iconSkipBackward = icon{alt: "skip-backward", image: icons.SkipBackward}
	iconSkipForward  = icon{alt: "skip-forward", image: icons.SkipForward}
	iconStep         = icon{alt: "step", image: icons.Step}
	iconZoomIn       = icon{alt: "zoom-in", image: icons.ZoomIn}
	iconZoomOut      = icon{alt: "zoom-out", image: icons.ZoomOut}
	allIcons         = []icon{
		iconBackward,
		iconBurger,
		iconPause,
		iconPlay,
		iconSkipBackward,
		iconSkipForward,
		iconStep,
		iconZoomIn,
		iconZoomOut,
	}
)

var contents = map[Topic]content{
	About:                contentAbout,
	CapturedPatterns:     contentCapturedPatterns,
	CollectedRules:       contentCollectedRules,
	Colors:               contentColors,
	Editor:               contentEditor,
	FileFinder:           contentFileFinder,
	GridRecipes:          contentGridRecipes,
	GridRecipesReference: contentGridRecipesReference,
	HeatMap:              contentHeatMap,
	ImportGrid:           contentImportGrid,
	Index:                contentIndex,
	Instrumentation:      contentInstrumentation,
	Keys:                 contentKeys,
	LoadPatterns:         contentLoadPatterns,
	MainMenu:             contentMainMenu,
	MetaRules:            contentMetaRules,
	MetaRulesRef:         contentMetaRulesRef,
	Patterns:             contentPatterns,
	PlacePattern:         contentPlacePattern,
	Rules:                contentRules,
	ShortCuts:            contentShortcuts,
	ShortCutsRef:         contentShortCutsRef,
	SizingWrapping:       contentSizingWrapping,
	StatusBar:            contentStatusBar,
	Stepping:             contentStepping,
}
