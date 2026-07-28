package meta

import (
	"github.com/marrow16/gogol/logic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRanges(t *testing.T) {
	r := Ranges{
		Single(0),
		{200, 100},
	}
	assert.Equal(t, "(0,100-200)", r.String())
	count := 0
	for perm := uint32(0); perm < 300; perm++ {
		if r.Matches(perm) {
			count++
			rule, err := logic.NewRuleFromPermutation(int(perm))
			require.NoError(t, err)
			assert.True(t, r.MatchesRule(rule))
		}
	}
	assert.Equal(t, 102, count)
}

func TestRanges_MatchingPermutations(t *testing.T) {
	r := Ranges{
		Single(0),
		{200, 100},
	}
	count := 0
	for _ = range r.MatchingPermutations() {
		count++
	}
	assert.Equal(t, 102, count)
	count = 0
	for _ = range r.MatchingPermutations() {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestRanges_MatchingRules(t *testing.T) {
	r := Ranges{
		Single(0),
		{200, 100},
	}
	count := 0
	for _ = range r.MatchingRules() {
		count++
	}
	assert.Equal(t, 102, count)
	count = 0
	for _ = range r.MatchingRules() {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestRanges_Next(t *testing.T) {
	r := Ranges{
		Single(0),
		{200, 100},
	}
	next := r.Next(0)
	assert.Equal(t, 100, next)
	next = r.Next(200)
	assert.Equal(t, 200, next)
	next = r.Next(-1)
	assert.Equal(t, 0, next)
}

func TestRanges_Previous(t *testing.T) {
	r := Ranges{
		Single(0),
		{200, 100},
	}
	prev := r.Previous((1 << 18) - 1)
	assert.Equal(t, 200, prev)
	prev = r.Previous(0)
	assert.Equal(t, 0, prev)
	prev = r.Previous(-1)
	assert.Equal(t, 200, prev)
}
