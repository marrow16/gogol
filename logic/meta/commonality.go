package meta

import (
	"errors"
	"fmt"

	"github.com/marrow16/gogol/logic"
)

const halfPermutationMask = uint32(1<<9) - 1

var ErrNoRules = errors.New("cannot determine commonality of an empty rule set")

// CommonalityFromRules returns the B/S conditions that are true for every supplied rule
//
//   - B/S digits present in every rule become required conditions
//   - B/S digits absent from every rule become forbidden conditions
//   - B/S digits which vary between rules remain unconstrained
func CommonalityFromRules(rules ...logic.Rule) (Rule, error) {
	permutations := make([]int, len(rules))
	for i, rule := range rules {
		permutations[i] = rule.Permutation()
	}
	return CommonalityFromPermutations(permutations...)
}

// CommonalityFromPermutations returns the B/S conditions that are true for every supplied permutation
func CommonalityFromPermutations(permutations ...int) (Rule, error) {
	if len(permutations) == 0 {
		return Rule{}, ErrNoRules
	}
	allBirth := int(halfPermutationMask)
	anyBirth := 0
	allSurvival := int(halfPermutationMask)
	anySurvival := 0
	for _, permutation := range permutations {
		if permutation < 0 || permutation > MaxPermutation {
			return Rule{}, fmt.Errorf("invalid permutation %d: range is 0 - %d", permutation, MaxPermutation)
		}
		birth := (permutation >> 9) & int(halfPermutationMask)
		survival := permutation & int(halfPermutationMask)
		allBirth &= birth
		anyBirth |= birth
		allSurvival &= survival
		anySurvival |= survival
	}
	forbiddenBirth := ^anyBirth & int(halfPermutationMask)
	forbiddenSurvival := ^anySurvival & int(halfPermutationMask)
	return Rule{
		Birth:    commonConditions(allBirth, forbiddenBirth),
		Survival: commonConditions(allSurvival, forbiddenSurvival),
	}, nil
}

func commonConditions(required, forbidden int) Conditions {
	conditions := make(Conditions, 0, 2)
	if required != 0 {
		conditions = append(conditions, Require(required))
	}
	if forbidden != 0 {
		conditions = append(conditions, Forbid(forbidden))
	}
	return conditions
}
