package meta

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCount(t *testing.T) {
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
	count := Count(b1Plus, b2Plus)
	assert.Equal(t, 1280, count)

	count = Count()
	assert.Equal(t, 0, count)
}

func TestIntersection(t *testing.T) {
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
	count := 0
	for _ = range Intersection(b1Plus, b2Plus) {
		count++
	}
	assert.Equal(t, 1280, count)
	count = 0
	for _ = range Intersection(b1Plus, b2Plus) {
		count++
		break
	}
	assert.Equal(t, 1, count)
	count = 0
	for _ = range Intersection() {
		count++
	}
	assert.Equal(t, 0, count)
}

func TestDifference(t *testing.T) {
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
	count := 0
	for _ = range Difference(b1Plus, b2Plus) {
		count++
	}
	assert.Equal(t, 1280, count)
	count = 0
	for _ = range Difference(b1Plus, b2Plus) {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestUnion(t *testing.T) {
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
	count := 0
	for _ = range Union(b1Plus, b2Plus) {
		count++
	}
	assert.Equal(t, 3840, count)
	count = 0
	for _ = range Union(b1Plus, b2Plus) {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestSymmetricDifference(t *testing.T) {
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
	count := 0
	for _ = range SymmetricDifference(b1Plus, b2Plus) {
		count++
	}
	assert.Equal(t, 2560, count)
	count = 0
	for _ = range SymmetricDifference(b1Plus, b2Plus) {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestComplement(t *testing.T) {
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
	count := 0
	for _ = range Complement(b1Plus) {
		count++
	}
	assert.Equal(t, 259584, count)
	count = 0
	for _ = range Complement(b1Plus) {
		count++
		break
	}
	assert.Equal(t, 1, count)

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
	count = 0
	for _ = range Complement(b1Plus, b2Plus) {
		count++
	}
	assert.Equal(t, 258304, count)
}
