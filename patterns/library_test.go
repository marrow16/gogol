package patterns

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestResetLibrary(t *testing.T) {
	require.Len(t, PatternLibrary, 18)

	PatternLibrary["foo"] = Pattern{}
	require.Len(t, PatternLibrary, 19)

	ResetLibrary()
	require.Len(t, PatternLibrary, 18)
}
