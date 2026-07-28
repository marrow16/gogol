package meta

import (
	"github.com/marrow16/gogol/logic"
	"iter"
	"strconv"
	"strings"
)

type Ranges []Range
type Range [2]uint32 // 0 is min, 1 is max (but order isn't really important - the min/max is resolved when used)

func Single(perm uint32) Range {
	return Range{perm, perm}
}

var _ Evaluator = (Ranges)(nil)
var _ Condition = (Ranges)(nil)

func (rn Ranges) Matches(permutation uint32) bool {
	if len(rn) == 0 {
		return true
	}
	for _, r := range rn {
		if permutation >= min(r[0], r[1]) && permutation <= max(r[0], r[1]) {
			return true
		}
	}
	return false
}

func (rn Ranges) MatchesRule(rule logic.Rule) bool {
	return rn.Matches(uint32(rule.Permutation()))
}

func (rn Ranges) MatchingPermutations() iter.Seq[uint32] {
	return func(yield func(uint32) bool) {
		mn, mx := rn.Minimum(), rn.Maximum()
		for perm := mn; perm <= mx; perm++ {
			if rn.Matches(perm) {
				if !yield(perm) {
					return
				}
			}
		}
	}
}

func (rn Ranges) MatchingRules() iter.Seq[logic.Rule] {
	return func(yield func(logic.Rule) bool) {
		mn, mx := rn.Minimum(), rn.Maximum()
		for perm := mn; perm <= mx; perm++ {
			if rn.Matches(perm) {
				rule, _ := logic.NewRuleFromPermutation(int(perm))
				if !yield(rule) {
					return
				}
			}
		}
	}
}

func (rn Ranges) Next(currentPermutation int) int {
	start := currentPermutation + 1
	if currentPermutation < 0 {
		start = 0
	}
	for i := uint32(start); i < 1<<18; i++ {
		if rn.Matches(i) {
			return int(i)
		}
	}
	return currentPermutation
}

func (rn Ranges) Previous(currentPermutation int) int {
	start := currentPermutation - 1
	if currentPermutation < 0 {
		start = MaxPermutation
	}
	for i := start; i >= 0; i-- {
		if rn.Matches(uint32(i)) {
			return i
		}
	}
	return currentPermutation
}

func (rn Ranges) String() string {
	parts := make([]string, 0, len(rn))
	for _, r := range rn {
		part := strconv.FormatUint(uint64(min(r[0], r[1])), 10)
		if r[0] != r[1] {
			part += "-" + strconv.FormatUint(uint64(max(r[0], r[1])), 10)
		}
		parts = append(parts, part)
	}
	return "(" + strings.Join(parts, ",") + ")"
}

func (rn Ranges) Minimum() uint32 {
	const maxPermutation = uint32(512 * 512)
	result := maxPermutation
	for _, r := range rn {
		if m := min(r[0], r[1]); m < result {
			result = m
		}
	}
	return result
}

func (rn Ranges) Maximum() uint32 {
	const maxPermutation = uint32(512 * 512)
	result := uint32(0)
	for _, r := range rn {
		if m := max(r[0], r[1]); m > result && m < maxPermutation {
			result = m
		}
	}
	return result
}
