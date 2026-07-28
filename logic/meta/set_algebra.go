package meta

import "iter"

func Count(evaluators ...Evaluator) int {
	if len(evaluators) == 0 {
		return 0
	}
	count := 0
	source := evaluators[0]
	for permutation := range source.MatchingPermutations() {
		matches := true
		for _, evaluator := range evaluators[1:] {
			if !evaluator.Matches(permutation) {
				matches = false
				break
			}
		}
		if matches {
			count++
		}
	}
	return count
}

func Intersection(evaluators ...Evaluator) iter.Seq[uint32] {
	return func(yield func(uint32) bool) {
		if len(evaluators) == 0 {
			return
		}
		source := evaluators[0]
		for permutation := range source.MatchingPermutations() {
			matches := true
			for _, evaluator := range evaluators[1:] {
				if !evaluator.Matches(permutation) {
					matches = false
					break
				}
			}
			if matches && !yield(permutation) {
				return
			}
		}
	}
}

func Difference(left, right Evaluator) iter.Seq[uint32] {
	return func(yield func(uint32) bool) {
		for permutation := range left.MatchingPermutations() {
			if !right.Matches(permutation) {
				if !yield(permutation) {
					return
				}
			}
		}
	}
}

func Union(evaluators ...Evaluator) iter.Seq[uint32] {
	return func(yield func(uint32) bool) {
		seen := make(map[uint32]struct{})
		for _, evaluator := range evaluators {
			for permutation := range evaluator.MatchingPermutations() {
				if _, exists := seen[permutation]; exists {
					continue
				}
				seen[permutation] = struct{}{}
				if !yield(permutation) {
					return
				}
			}
		}
	}
}

func SymmetricDifference(evaluators ...Evaluator) iter.Seq[uint32] {
	return func(yield func(uint32) bool) {
		seen := make(map[uint32]struct{})
		for _, evaluator := range evaluators {
			for permutation := range evaluator.MatchingPermutations() {
				if _, exists := seen[permutation]; exists {
					continue
				}
				seen[permutation] = struct{}{}
				matches := 0
				for _, other := range evaluators {
					if other.Matches(permutation) {
						matches++
					}
				}
				if matches%2 != 0 && !yield(permutation) {
					return
				}
			}
		}
	}
}

func Complement(evaluators ...Evaluator) iter.Seq[uint32] {
	return func(yield func(uint32) bool) {
		for permutation := uint32(0); permutation < 1<<18; permutation++ {
			matched := false
			for _, evaluator := range evaluators {
				if evaluator.Matches(permutation) {
					matched = true
					break
				}
			}
			if !matched && !yield(permutation) {
				return
			}
		}
	}
}
