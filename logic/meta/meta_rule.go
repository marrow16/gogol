package meta

import (
	"fmt"
	"github.com/marrow16/gogol/logic"
	"iter"
	"strings"
)

const MaxPermutation = (1 << 18) - 1

type Evaluator interface {
	Matches(permutation uint32) bool
	MatchesRule(rule logic.Rule) bool
	MatchingPermutations() iter.Seq[uint32]
	MatchingRules() iter.Seq[logic.Rule]
	Next(currentPermutation int) int
	Previous(currentPermutation int) int
	String() string
}

type Rule struct {
	Birth        Conditions
	Survival     Conditions
	And          Conditions
	Or           Conditions
	Xor          Conditions
	Permutations Ranges
}

var _ Evaluator = (*Rule)(nil)

func (r Rule) Matches(permutation uint32) bool {
	const nineBits = uint32(1<<9) - 1
	survival := permutation & nineBits
	birth := (permutation >> 9) & nineBits
	andBits := birth & survival
	orBits := birth | survival
	xorBits := birth ^ survival
	return r.Birth.Matches(birth) &&
		r.Survival.Matches(survival) &&
		r.And.Matches(andBits) && r.Or.Matches(orBits) && r.Xor.Matches(xorBits) &&
		r.Permutations.Matches(permutation)
}

func (r Rule) MatchesRule(rule logic.Rule) bool {
	return r.Matches(uint32(rule.Permutation()))
}

func (r Rule) MatchingPermutations() iter.Seq[uint32] {
	return func(yield func(uint32) bool) {
		for b := 0; b < 512; b++ {
			for s := 0; s < 512; s++ {
				perm := uint32((b * 512) + s)
				if r.Matches(perm) {
					if !yield(perm) {
						return
					}
				}
			}
		}
	}
}

func (r Rule) MatchingRules() iter.Seq[logic.Rule] {
	return func(yield func(logic.Rule) bool) {
		for b := 0; b < 512; b++ {
			for s := 0; s < 512; s++ {
				perm := uint32((b * 512) + s)
				if r.Matches(perm) {
					rule, _ := logic.NewRuleFromPermutation(int(perm))
					if !yield(rule) {
						return
					}
				}
			}
		}
	}
}

func (r Rule) Next(currentPermutation int) int {
	start := currentPermutation + 1
	if currentPermutation < 0 {
		start = 0
	}
	for i := uint32(start); i < 1<<18; i++ {
		if r.Matches(i) {
			return int(i)
		}
	}
	return currentPermutation
}

func (r Rule) Previous(currentPermutation int) int {
	start := currentPermutation - 1
	if currentPermutation < 0 {
		start = MaxPermutation
	}
	for i := start; i >= 0; i-- {
		if r.Matches(uint32(i)) {
			return i
		}
	}
	return currentPermutation
}

func (r Rule) String() string {
	var b strings.Builder
	if len(r.Birth) > 0 {
		b.WriteString("B")
		b.WriteString(r.Birth.String())
	}
	if len(r.Survival) > 0 {
		if b.Len() > 0 {
			b.WriteString(" / ")
		}
		b.WriteString("S")
		b.WriteString(r.Survival.String())
	}
	if len(r.And) > 0 {
		if b.Len() > 0 {
			b.WriteString(" / ")
		}
		b.WriteString("&")
		b.WriteString(r.And.String())
	}
	if len(r.Or) > 0 {
		if b.Len() > 0 {
			b.WriteString(" / ")
		}
		b.WriteString("|")
		b.WriteString(r.Or.String())
	}
	if len(r.Xor) > 0 {
		if b.Len() > 0 {
			b.WriteString(" / ")
		}
		b.WriteString("^")
		b.WriteString(r.Xor.String())
	}
	if len(r.Permutations) > 0 {
		if b.Len() > 0 {
			b.WriteString(" / ")
		}
		b.WriteString("P")
		b.WriteString(r.Permutations.String())
	}
	return b.String()
}

type CompositeMode int

const (
	AllOfMode CompositeMode = iota
	AnyOfMode
	NoneOfMode
	OneOfMode
)

func (m CompositeMode) String() string {
	switch m {
	case AllOfMode:
		return "AllOf"
	case AnyOfMode:
		return "AnyOf"
	case NoneOfMode:
		return "NoneOf"
	case OneOfMode:
		return "OneOf"
	}
	return "Unknown"
}

func AllOf(rules ...Evaluator) Evaluator {
	return CompositeRule{
		Mode:  AllOfMode,
		Rules: rules,
	}
}

func AnyOf(rules ...Evaluator) Evaluator {
	return CompositeRule{
		Mode:  AnyOfMode,
		Rules: rules,
	}
}

func NoneOf(rules ...Evaluator) Evaluator {
	return CompositeRule{
		Mode:  NoneOfMode,
		Rules: rules,
	}
}

func OneOf(rules ...Evaluator) Evaluator {
	return CompositeRule{
		Mode:  OneOfMode,
		Rules: rules,
	}
}

type CompositeRule struct {
	Mode  CompositeMode
	Rules []Evaluator
}

var _ Evaluator = (*CompositeRule)(nil)

func (cr CompositeRule) Matches(permutation uint32) bool {
	if len(cr.Rules) == 0 {
		return false
	}
	switch cr.Mode {
	case AllOfMode:
		for _, rule := range cr.Rules {
			if rule != nil && !rule.Matches(permutation) {
				return false
			}
		}
		return true
	case AnyOfMode:
		for _, rule := range cr.Rules {
			if rule != nil && rule.Matches(permutation) {
				return true
			}
		}
		return false
	case NoneOfMode:
		for _, rule := range cr.Rules {
			if rule != nil && rule.Matches(permutation) {
				return false
			}
		}
		return true
	case OneOfMode:
		count := 0
		for _, rule := range cr.Rules {
			if rule != nil && rule.Matches(permutation) {
				count++
			}
		}
		return count == 1
	default:
		return false
	}
}

func (cr CompositeRule) MatchesRule(rule logic.Rule) bool {
	return cr.Matches(uint32(rule.Permutation()))
}

func (cr CompositeRule) MatchingPermutations() iter.Seq[uint32] {
	return func(yield func(uint32) bool) {
		for b := 0; b < 512; b++ {
			for s := 0; s < 512; s++ {
				perm := uint32((b * 512) + s)
				if cr.Matches(perm) {
					if !yield(perm) {
						return
					}
				}
			}
		}
	}
}

func (cr CompositeRule) MatchingRules() iter.Seq[logic.Rule] {
	return func(yield func(logic.Rule) bool) {
		for b := 0; b < 512; b++ {
			for s := 0; s < 512; s++ {
				perm := uint32((b * 512) + s)
				if cr.Matches(perm) {
					rule, _ := logic.NewRuleFromPermutation(int(perm))
					if !yield(rule) {
						return
					}
				}
			}
		}
	}
}

func (cr CompositeRule) Next(currentPermutation int) int {
	start := currentPermutation + 1
	if currentPermutation < 0 {
		start = 0
	}
	for i := uint32(start); i < 1<<18; i++ {
		if cr.Matches(i) {
			return int(i)
		}
	}
	return currentPermutation
}

func (cr CompositeRule) Previous(currentPermutation int) int {
	start := currentPermutation - 1
	if currentPermutation < 0 {
		start = MaxPermutation
	}
	for i := start; i >= 0; i-- {
		if cr.Matches(uint32(i)) {
			return i
		}
	}
	return currentPermutation
}

func (cr CompositeRule) String() string {
	parts := make([]string, 0, len(cr.Rules))
	for _, rule := range cr.Rules {
		parts = append(parts, rule.String())
	}
	return cr.Mode.String() + "(" + strings.Join(parts, ", ") + ")"
}

func BitMask(values ...uint8) uint16 {
	var mask uint16
	for _, value := range values {
		if value > 8 {
			panic(fmt.Sprintf("invalid digit bit: %d", value))
		}
		mask |= 1 << value
	}
	return mask
}

func DigitMask(digits ...byte) uint16 {
	values := make([]uint8, 0, len(digits))
	for _, d := range digits {
		if d < 9 {
			values = append(values, d)
		} else if d >= '0' && d < '9' {
			values = append(values, d-'0')
		}
	}
	return BitMask(values...)
}
