package meta

import (
	"github.com/marrow16/gogol/logic"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCommonalityFromPermutations(t *testing.T) {
	mr, err := ParseRule(`AllOf(
	S(+47) / B(!2345),
	AnyOf(
		B(+0,!1) / S(!0123,-568),
		B(+0,!16) / S(+568,!0123),
		B(+07,!1) / S(+568,!0123),
		B(+17,!06) / S(+5,!3,-12),
		B(+017,!6) / S(+5,!3,-12,-68),
		B(+1,!067) / S(+5,!03,-12),
		B(+1,!067) / S(+5,!13,-12),
		B(+1,!067) / S(+56,!3,-12),
		B(+01,!678) / S(+5,!03,-12,-68),
		B(+01,!678) / S(+5,!13,-12,-68),
		B(+018,!67) / S(+5,!03,-12,-68),
		B(+018,!67) / S(+5,!13,-12,-68),
		B(+018,!67) / S(+56,!38,-12),
		B(+018,!67) / S(+58,!36,-12)
	)
)`)
	require.NoError(t, err)
	perms := make([]int, 0)
	permsMap := make(map[int]struct{})
	for perm := range mr.MatchingPermutations() {
		perms = append(perms, int(perm))
		permsMap[int(perm)] = struct{}{}
	}
	ev, err := CommonalityFromPermutations(perms...)
	require.NoError(t, err)
	require.Equal(t, `B(!2345) / S(+47,!3)`, ev.String())
}

func TestCommonalityFromPermutations_Errors(t *testing.T) {
	_, err := CommonalityFromPermutations()
	require.Error(t, err)
	require.Equal(t, ErrNoRules, err)

	_, err = CommonalityFromPermutations(-1)
	require.Error(t, err)
	_, err = CommonalityFromPermutations(1 << 18)
	require.Error(t, err)
}

func TestCommonalityFromRules(t *testing.T) {
	r, err := ParseRule(`AllOf(
	S(+47) / B(!2345),
	AnyOf(
		B(+0,!1) / S(!0123,-568),
		B(+0,!16) / S(+568,!0123),
		B(+07,!1) / S(+568,!0123),
		B(+17,!06) / S(+5,!3,-12),
		B(+017,!6) / S(+5,!3,-12,-68),
		B(+1,!067) / S(+5,!03,-12),
		B(+1,!067) / S(+5,!13,-12),
		B(+1,!067) / S(+56,!3,-12),
		B(+01,!678) / S(+5,!03,-12,-68),
		B(+01,!678) / S(+5,!13,-12,-68),
		B(+018,!67) / S(+5,!03,-12,-68),
		B(+018,!67) / S(+5,!13,-12,-68),
		B(+018,!67) / S(+56,!38,-12),
		B(+018,!67) / S(+58,!36,-12)
	)
)`)
	require.NoError(t, err)
	rules := make([]logic.Rule, 0)
	for rule := range r.MatchingRules() {
		rules = append(rules, rule)
	}
	ev, err := CommonalityFromRules(rules...)
	require.NoError(t, err)
	require.Equal(t, `B(!2345) / S(+47,!3)`, ev.String())
}

func TestNoCommonality(t *testing.T) {
	r1, err := logic.NewRuleRle("", "B/S")
	require.NoError(t, err)
	r2, err := logic.NewRuleRle("", "B012345678/S012345678")
	require.NoError(t, err)
	ev, err := CommonalityFromRules(r1, r2)
	require.NoError(t, err)
	require.Equal(t, "", ev.String())
}
