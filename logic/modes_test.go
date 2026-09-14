package logic

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWrapMode(t *testing.T) {
	testCases := map[WrapMode]string{
		WrapNone:       "None",
		WrapHorizontal: "Horizontal",
		WrapVertical:   "Vertical",
		WrapAll:        "Toroidal",
	}
	for m, s := range testCases {
		t.Run(s, func(t *testing.T) {
			require.Equal(t, s, m.String())
			require.Equal(t, m, WrapModeFromString(m.String(), -1))
		})
	}
	assert.Equal(t, WrapAll, WrapModeFromString("", WrapAll))
	assert.Equal(t, "Unknown", WrapMode(-1).String())
}

func TestBoundaryMode(t *testing.T) {
	testCases := map[BoundaryMode]string{
		DeadBoundary:  "Dead",
		AliveBoundary: "Alive",
	}
	for m, s := range testCases {
		t.Run(s, func(t *testing.T) {
			require.Equal(t, s, m.String())
			require.Equal(t, m, BoundaryModeFromString(m.String(), -1))
		})
	}
	assert.Equal(t, AliveBoundary, BoundaryModeFromString("", AliveBoundary))
	assert.Equal(t, "Unknown", BoundaryMode(-1).String())

}
