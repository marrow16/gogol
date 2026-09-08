package help

type Topic int

const (
	Index Topic = iota
	StatusBar
	Editor
	HeatMap
	PlacePattern
	Rules
	MainMenu
	Colors
	SizingWrapping
	Stepping
	CapturedPatterns
	Patterns
	LoadPatterns
	ImportGrid
	GridRecipes
	GridRecipesReference
	Instrumentation
	ShortCuts
	ShortCutsRef
	MetaRules
	MetaRulesRef
	CollectedRules
	Keys
	About
)

func (t Topic) link(s ...string) topicLink {
	if len(s) > 0 && s[0] != "" {
		return topicLink{Topic: t, Text: s[0]}
	}
	return topicLink{Topic: t}
}

func (t Topic) linkItalic(s ...string) topicLink {
	if len(s) > 0 && s[0] != "" {
		return topicLink{Topic: t, Text: s[0], Italic: true}
	}
	return topicLink{Topic: t, Italic: true}
}

func (t Topic) String() string {
	switch t {
	case Index:
		return "Index"
	case StatusBar:
		return "StatusBar Help"
	case Editor:
		return "Edit Mode Help"
	case HeatMap:
		return "Heat Map Help"
	case PlacePattern:
		return "Place Pattern Mode Help"
	case Rules:
		return "Rules Help"
	case MainMenu:
		return "Main Menu Help"
	case Colors:
		return "Colors Help"
	case SizingWrapping:
		return "Sizing/Wrapping/Boundaries Help"
	case Stepping:
		return "Stepping Help"
	case CapturedPatterns:
		return "Captured Patterns Help"
	case Patterns:
		return "Patterns Help"
	case LoadPatterns:
		return "Load Patterns Help"
	case ImportGrid:
		return "Import Grid Help"
	case GridRecipes:
		return "Grid Recipes Help"
	case GridRecipesReference:
		return "Grid Recipes Reference"
	case Instrumentation:
		return "Instrumentation Help"
	case ShortCuts:
		return "Shortcuts Help"
	case ShortCutsRef:
		return "Shortcuts Actions Reference"
	case MetaRules:
		return "Meta Rules Help"
	case MetaRulesRef:
		return "Meta Rules Reference"
	case CollectedRules:
		return "Collected Rules Help"
	case Keys:
		return "Keys Help"
	case About:
		return "About GoGoL"
	}
	return "Help"
}
