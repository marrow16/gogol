package logic

import "strings"

type WrapMode int

const (
	WrapNone WrapMode = iota
	WrapHorizontal
	WrapVertical
	WrapAll
)

func (m WrapMode) String() string {
	switch m {
	case WrapNone:
		return "None"
	case WrapHorizontal:
		return "Horizontal"
	case WrapVertical:
		return "Vertical"
	case WrapAll:
		return "Toroidal"
	}
	return "Unknown"
}

func WrapModeFromString(s string, def WrapMode) WrapMode {
	switch strings.ToLower(s) {
	case "none":
		return WrapNone
	case "horizontal":
		return WrapHorizontal
	case "vertical":
		return WrapVertical
	case "all", "toroidal":
		return WrapAll
	}
	return def
}

type BoundaryMode int

const (
	DeadBoundary BoundaryMode = iota
	AliveBoundary
)

func (m BoundaryMode) String() string {
	switch m {
	case DeadBoundary:
		return "Dead"
	case AliveBoundary:
		return "Alive"
	}
	return "Unknown"
}

func BoundaryModeFromString(s string, def BoundaryMode) BoundaryMode {
	switch strings.ToLower(s) {
	case "dead":
		return DeadBoundary
	case "alive":
		return AliveBoundary
	}
	return def
}
