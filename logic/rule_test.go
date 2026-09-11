package logic

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStandard(t *testing.T) {
	r := Rules["Standard"]
	require.Equal(t, "B3/S23", r.Rle())
	require.Equal(t, 4108, r.Permutation())
	perm := r.Permutation()
	r2, err := NewRuleFromPermutation(perm)
	require.NoError(t, err)
	assert.Equal(t, r.Rle(), r2.Rle())
	testCases := []struct {
		alive   bool
		adjs    int
		changed bool
	}{
		{adjs: 0},
		{adjs: 1},
		{adjs: 2},
		{adjs: 3, changed: true},
		{adjs: 4},
		{adjs: 5},
		{adjs: 6},
		{adjs: 7},
		{adjs: 8},
		{alive: true, adjs: 0, changed: true},
		{alive: true, adjs: 1, changed: true},
		{alive: true, adjs: 2},
		{alive: true, adjs: 3},
		{alive: true, adjs: 4, changed: true},
		{alive: true, adjs: 5, changed: true},
		{alive: true, adjs: 6, changed: true},
		{alive: true, adjs: 7, changed: true},
		{alive: true, adjs: 8, changed: true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("alive=%t,adjs=%d", tc.alive, tc.adjs), func(t *testing.T) {
			c := newTestCell(tc.alive, tc.adjs)
			assert.Equal(t, tc.changed, r.StateChanged(c))
		})
	}
}

func newTestCell(alive bool, adjsAlive int) *Cell {
	result := &Cell{
		Alive: alive,
	}
	for c := range 8 {
		result.Adjacents[c] = &Cell{
			Alive: c < adjsAlive,
		}
	}
	return result
}

func TestFlippedStandard(t *testing.T) {
	r, err := NewRuleRle("", "S23/B3")
	require.NoError(t, err)
	testCases := []struct {
		alive   bool
		adjs    int
		changed bool
	}{
		{adjs: 0},
		{adjs: 1},
		{adjs: 2},
		{adjs: 3, changed: true},
		{adjs: 4},
		{adjs: 5},
		{adjs: 6},
		{adjs: 7},
		{adjs: 8},
		{alive: true, adjs: 0, changed: true},
		{alive: true, adjs: 1, changed: true},
		{alive: true, adjs: 2},
		{alive: true, adjs: 3},
		{alive: true, adjs: 4, changed: true},
		{alive: true, adjs: 5, changed: true},
		{alive: true, adjs: 6, changed: true},
		{alive: true, adjs: 7, changed: true},
		{alive: true, adjs: 8, changed: true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("alive=%t,adjs=%d", tc.alive, tc.adjs), func(t *testing.T) {
			c := newTestCell(tc.alive, tc.adjs)
			assert.Equal(t, tc.changed, r.StateChanged(c))
		})
	}
}

func TestAntiLife(t *testing.T) {
	r := Rules["AntiLife"]
	testCases := []struct {
		alive   bool
		adjs    int
		changed bool
	}{
		{adjs: 0, changed: true},
		{adjs: 1, changed: true},
		{adjs: 2, changed: true},
		{adjs: 3, changed: true},
		{adjs: 4, changed: true},
		{adjs: 5},
		{adjs: 6},
		{adjs: 7, changed: true},
		{adjs: 8, changed: true},
		{alive: true, adjs: 0},
		{alive: true, adjs: 1},
		{alive: true, adjs: 2},
		{alive: true, adjs: 3},
		{alive: true, adjs: 4},
		{alive: true, adjs: 5, changed: true},
		{alive: true, adjs: 6},
		{alive: true, adjs: 7},
		{alive: true, adjs: 8},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("alive=%t,adjs=%d", tc.alive, tc.adjs), func(t *testing.T) {
			c := newTestCell(tc.alive, tc.adjs)
			assert.Equal(t, tc.changed, r.StateChanged(c))
		})
	}
}

func TestInvalidRule(t *testing.T) {
	t.Run("invalid survives", func(t *testing.T) {
		_, err := NewRuleRle("", "SX/B3")
		require.Error(t, err)
		require.ErrorIs(t, ErrInvalidRule, err)
	})
	t.Run("invalid born", func(t *testing.T) {
		_, err := NewRuleRle("", "S3/BX")
		require.Error(t, err)
		require.ErrorIs(t, ErrInvalidRule, err)
	})
	t.Run("invalid survives - panics", func(t *testing.T) {
		require.Panics(t, func() {
			_ = MustNewRuleRle("", "SX/B3")
		})
	})
	t.Run("invalid born - panics", func(t *testing.T) {
		require.Panics(t, func() {
			_ = MustNewRuleRle("", "S3/BX")
		})
	})
}

func TestRlePermutationReversible(t *testing.T) {
	for n, r := range Rules {
		t.Run(n, func(t *testing.T) {
			assert.Equal(t, n, r.Name())
			perm := r.Permutation()
			r2, err := NewRuleFromPermutation(perm)
			assert.NoError(t, err)
			assert.Equal(t, r.Rle(), r2.Rle())
			assert.Equal(t, r.Name(), r2.Name())
		})
	}
	r, err := NewRuleRle("", "S0/B0")
	require.NoError(t, err)
	assert.Equal(t, "B0/S0", r.Rle())
	assert.Equal(t, "Custom B0/S0 (513)", r.Name())
}

func TestNewRuleEvaluatorFromPermutation_ZeroPerm(t *testing.T) {
	r, err := NewRuleFromPermutation(0)
	require.NoError(t, err)
	assert.Equal(t, "B/S", r.Rle())
	assert.Equal(t, "Custom B/S (0)", r.Name())
	assert.Equal(t, 0, r.Permutation())
}

func TestNewRuleEvaluatorFromPermutation_BadPerm(t *testing.T) {
	_, err := NewRuleFromPermutation(-1)
	require.Error(t, err)
	require.Equal(t, ErrInvalidPermutation, err)

	_, err = NewRuleFromPermutation(1 << 18)
	require.Error(t, err)
	require.Equal(t, ErrInvalidPermutation, err)
}

func TestRule_Integer(t *testing.T) {
	i := StandardRule.Integer()
	require.Equal(t, 6152, i)
}

func TestPermutationToInteger(t *testing.T) {
	i := PermutationToInteger(StandardRule.Permutation())
	require.Equal(t, 6152, i)
}

func TestIntegerToPermutation(t *testing.T) {
	i := IntegerToPermutation(StandardRule.Integer())
	require.Equal(t, 4108, i)
}
