package meta

import (
	"github.com/marrow16/gogol/logic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRule(t *testing.T) {
	b1Plus := Rule{
		Birth: Conditions{
			Require(BitMask(1)),
		},
		Survival: Conditions{
			Require(BitMask(4, 5, 7)),
			Forbid(BitMask(3)),
			ExcludeCombination(BitMask(0, 1)),
			ExcludeCombination(BitMask(1, 2)),
		},
		Permutations: Ranges{{0, (512 * 256) - 1}},
	}
	assert.Equal(t, "B(+1) / S(+457,!3,-01,-12) / P(0-131071)", b1Plus.String())
	count := 0
	for b := 0; b < 512; b++ {
		for s := 0; s < 512; s++ {
			perm := (b * 512) + s
			if b1Plus.Matches(uint32(perm)) {
				count++
			}
		}
	}
	assert.Equal(t, 2560, count)
}

func TestRule_MatchesRule(t *testing.T) {
	mr := MustParseRule("S(+23) / B(+3)")
	r, err := logic.NewRuleRle("", "S23/B3")
	require.NoError(t, err)
	assert.True(t, mr.MatchesRule(r))

	r, err = logic.NewRuleRle("", "S23/B22")
	require.NoError(t, err)
	assert.False(t, mr.MatchesRule(r))
}

func TestRule_MatchingPermutations(t *testing.T) {
	b1Plus := Rule{
		Birth: Conditions{
			Require(BitMask(1)),
		},
		Survival: Conditions{
			Require(BitMask(4, 5, 7)),
			Forbid(BitMask(3)),
			ExcludeCombination(BitMask(0, 1)),
			ExcludeCombination(BitMask(1, 2)),
		},
		Permutations: Ranges{{0, (512 * 256) - 1}},
	}
	assert.Equal(t, "B(+1) / S(+457,!3,-01,-12) / P(0-131071)", b1Plus.String())
	count := 0
	for _ = range b1Plus.MatchingPermutations() {
		count++
	}
	assert.Equal(t, 2560, count)
	count = 0
	for _ = range b1Plus.MatchingPermutations() {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestRule_MatchingRules(t *testing.T) {
	b1Plus := Rule{
		Birth: Conditions{
			Require(BitMask(1)),
		},
		Survival: Conditions{
			Require(BitMask(4, 5, 7)),
			Forbid(BitMask(3)),
			ExcludeCombination(BitMask(0, 1)),
			ExcludeCombination(BitMask(1, 2)),
		},
		Permutations: Ranges{{0, (512 * 256) - 1}},
	}
	assert.Equal(t, "B(+1) / S(+457,!3,-01,-12) / P(0-131071)", b1Plus.String())
	count := 0
	for _ = range b1Plus.MatchingRules() {
		count++
	}
	assert.Equal(t, 2560, count)
	count = 0
	for _ = range b1Plus.MatchingRules() {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestRule_Next(t *testing.T) {
	b1Plus := Rule{
		Birth: Conditions{
			Require(BitMask(1)),
		},
		Survival: Conditions{
			Require(BitMask(4, 5, 7)),
			Forbid(BitMask(3)),
			ExcludeCombination(BitMask(0, 1)),
			ExcludeCombination(BitMask(1, 2)),
		},
		Permutations: Ranges{{0, (512 * 256) - 1}},
	}
	next := b1Plus.Next(0)
	require.Equal(t, 0x4b0, next)
	next = b1Plus.Next((1 << 18) - 1)
	require.Equal(t, 262143, next)
	next = b1Plus.Next(-1)
	require.Equal(t, 0x4b0, next)
}

func TestRule_Previous(t *testing.T) {
	b1Plus := Rule{
		Birth: Conditions{
			Require(BitMask(1)),
		},
		Survival: Conditions{
			Require(BitMask(4, 5, 7)),
			Forbid(BitMask(3)),
			ExcludeCombination(BitMask(0, 1)),
			ExcludeCombination(BitMask(1, 2)),
		},
		Permutations: Ranges{{0, (512 * 256) - 1}},
	}
	prev := b1Plus.Previous(0x4b1)
	require.Equal(t, 1200, prev)
	prev = b1Plus.Previous((1 << 18) - 1)
	require.Equal(t, 131061, prev)
	prev = b1Plus.Previous(0)
	require.Equal(t, 0, prev)
	prev = b1Plus.Previous(-1)
	require.Equal(t, 131061, prev)
}

func TestCompositeMode_String(t *testing.T) {
	require.Equal(t, "AllOf", AllOfMode.String())
	require.Equal(t, "AnyOf", AnyOfMode.String())
	require.Equal(t, "NoneOf", NoneOfMode.String())
	require.Equal(t, "OneOf", OneOfMode.String())
	require.Equal(t, "Unknown", CompositeMode(-1).String())
}

func TestCompositeRule(t *testing.T) {
	b1Plus := Rule{
		Birth: Conditions{
			Require(BitMask(1)),
		},
		Survival: Conditions{
			Require(BitMask(4, 5, 7)),
			Forbid(BitMask(3)),
			ExcludeCombination(BitMask(0, 1)),
			ExcludeCombination(BitMask(1, 2)),
		},
		Permutations: Ranges{{0, (512 * 256) - 1}},
	}
	b2Plus := Rule{
		Birth: Conditions{
			Require(BitMask(2)),
		},
		Survival: Conditions{
			Require(BitMask(4, 5, 7)),
			Forbid(BitMask(3)),
			ExcludeCombination(BitMask(0, 1)),
			ExcludeCombination(BitMask(1, 2)),
		},
		Permutations: Ranges{{0, (512 * 256) - 1}},
	}
	modes := map[CompositeMode]Evaluator{
		AllOfMode:  AllOf(b1Plus, b2Plus),
		AnyOfMode:  AnyOf(b1Plus, b2Plus),
		NoneOfMode: NoneOf(b1Plus, b2Plus),
		OneOfMode:  OneOf(b1Plus, b2Plus),
	}
	expected := map[CompositeMode]int{
		AllOfMode:  1280,
		AnyOfMode:  3840,
		NoneOfMode: 258304,
		OneOfMode:  2560,
	}
	counts := map[CompositeMode]int{
		AllOfMode:  0,
		AnyOfMode:  0,
		NoneOfMode: 0,
		OneOfMode:  0,
	}
	for b := 0; b < 512; b++ {
		for s := 0; s < 512; s++ {
			perm := (b * 512) + s
			for mode, evaluator := range modes {
				if evaluator.Matches(uint32(perm)) {
					counts[mode]++
					r, err := logic.NewRuleFromPermutation(perm)
					require.NoError(t, err)
					assert.True(t, evaluator.MatchesRule(r))
				}
			}
		}
	}
	for mode, expect := range expected {
		assert.Equal(t, expect, counts[mode], mode.String())
	}
	assert.Equal(t, "AllOf(B(+1) / S(+457,!3,-01,-12) / P(0-131071), B(+2) / S(+457,!3,-01,-12) / P(0-131071))", modes[AllOfMode].String())
	assert.Equal(t, "AnyOf(B(+1) / S(+457,!3,-01,-12) / P(0-131071), B(+2) / S(+457,!3,-01,-12) / P(0-131071))", modes[AnyOfMode].String())
	assert.Equal(t, "NoneOf(B(+1) / S(+457,!3,-01,-12) / P(0-131071), B(+2) / S(+457,!3,-01,-12) / P(0-131071))", modes[NoneOfMode].String())
	assert.Equal(t, "OneOf(B(+1) / S(+457,!3,-01,-12) / P(0-131071), B(+2) / S(+457,!3,-01,-12) / P(0-131071))", modes[OneOfMode].String())
}

func TestCompositeRule_MatchingPermutations(t *testing.T) {
	b1Plus := Rule{
		Birth: Conditions{
			Require(BitMask(1)),
		},
		Survival: Conditions{
			Require(BitMask(4, 5, 7)),
			Forbid(BitMask(3)),
			ExcludeCombination(BitMask(0, 1)),
			ExcludeCombination(BitMask(1, 2)),
		},
		Permutations: Ranges{{0, (512 * 256) - 1}},
	}
	b2Plus := Rule{
		Birth: Conditions{
			Require(BitMask(2)),
		},
		Survival: Conditions{
			Require(BitMask(4, 5, 7)),
			Forbid(BitMask(3)),
			ExcludeCombination(BitMask(0, 1)),
			ExcludeCombination(BitMask(1, 2)),
		},
		Permutations: Ranges{{0, (512 * 256) - 1}},
	}
	cr := AnyOf(b1Plus, b2Plus)
	count := 0
	for _ = range cr.MatchingPermutations() {
		count++
	}
	assert.Equal(t, 3840, count)
	count = 0
	for _ = range cr.MatchingPermutations() {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestCompositeRule_MatchingRules(t *testing.T) {
	b1Plus := Rule{
		Birth: Conditions{
			Require(BitMask(1)),
		},
		Survival: Conditions{
			Require(BitMask(4, 5, 7)),
			Forbid(BitMask(3)),
			ExcludeCombination(BitMask(0, 1)),
			ExcludeCombination(BitMask(1, 2)),
		},
		Permutations: Ranges{{0, (512 * 256) - 1}},
	}
	b2Plus := Rule{
		Birth: Conditions{
			Require(BitMask(2)),
		},
		Survival: Conditions{
			Require(BitMask(4, 5, 7)),
			Forbid(BitMask(3)),
			ExcludeCombination(BitMask(0, 1)),
			ExcludeCombination(BitMask(1, 2)),
		},
		Permutations: Ranges{{0, (512 * 256) - 1}},
	}
	cr := AnyOf(b1Plus, b2Plus)
	count := 0
	for _ = range cr.MatchingRules() {
		count++
	}
	assert.Equal(t, 3840, count)
	count = 0
	for _ = range cr.MatchingRules() {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestCompositeRule_Empty(t *testing.T) {
	m := CompositeRule{}
	require.False(t, m.Matches(0))
}

func TestCompositeRule_InvalidMode(t *testing.T) {
	m := CompositeRule{
		Mode: CompositeMode(-1),
		Rules: []Evaluator{
			Rule{Birth: Conditions{Require(BitMask(2))}},
		},
	}
	require.False(t, m.Matches(0))
}

func TestCompositeRule_Next(t *testing.T) {
	b1Plus := Rule{
		Birth: Conditions{
			Require(BitMask(1)),
		},
		Survival: Conditions{
			Require(BitMask(4, 5, 7)),
			Forbid(BitMask(3)),
			ExcludeCombination(BitMask(0, 1)),
			ExcludeCombination(BitMask(1, 2)),
		},
		Permutations: Ranges{{0, (512 * 256) - 1}},
	}
	b2Plus := Rule{
		Birth: Conditions{
			Require(BitMask(2)),
		},
		Survival: Conditions{
			Require(BitMask(4, 5, 7)),
			Forbid(BitMask(3)),
			ExcludeCombination(BitMask(0, 1)),
			ExcludeCombination(BitMask(1, 2)),
		},
		Permutations: Ranges{{0, (512 * 256) - 1}},
	}
	cr := AnyOf(b1Plus, b2Plus)
	next := cr.Next(0)
	require.Equal(t, 1200, next)
	next = cr.Next((1 << 18) - 1)
	require.Equal(t, 262143, next)
	next = cr.Next(-1)
	require.Equal(t, 1200, next)
}

func TestCompositeRule_Previous(t *testing.T) {
	b1Plus := Rule{
		Birth: Conditions{
			Require(BitMask(1)),
		},
		Survival: Conditions{
			Require(BitMask(4, 5, 7)),
			Forbid(BitMask(3)),
			ExcludeCombination(BitMask(0, 1)),
			ExcludeCombination(BitMask(1, 2)),
		},
		Permutations: Ranges{{0, (512 * 256) - 1}},
	}
	b2Plus := Rule{
		Birth: Conditions{
			Require(BitMask(2)),
		},
		Survival: Conditions{
			Require(BitMask(4, 5, 7)),
			Forbid(BitMask(3)),
			ExcludeCombination(BitMask(0, 1)),
			ExcludeCombination(BitMask(1, 2)),
		},
		Permutations: Ranges{{0, (512 * 256) - 1}},
	}
	cr := AnyOf(b1Plus, b2Plus)
	prev := cr.Previous(0x4b1)
	require.Equal(t, 1200, prev)
	prev = cr.Previous((1 << 18) - 1)
	require.Equal(t, 131061, prev)
	prev = cr.Previous(0)
	require.Equal(t, 0, prev)
	prev = cr.Previous(-1)
	require.Equal(t, 131061, prev)
}

func TestBitMask(t *testing.T) {
	mask := BitMask(0, 1, 2, 3, 4, 5, 6, 7, 8)
	assert.Equal(t, uint16(0x1ff), mask)
	require.Panics(t, func() {
		_ = BitMask(9)
	})
}

func TestDigitMask(t *testing.T) {
	mask := DigitMask(0, '1', 2, '3', '4', '5', '6', '7', '8', '9', 'X')
	assert.Equal(t, uint16(0x1ff), mask)
}
