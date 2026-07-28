package meta

import (
	"fmt"
	"math/bits"
	"strings"
)

type Condition interface {
	Matches(halfPerm uint32) bool
	String() string
}

type Conditions []Condition

var _ Condition = Conditions{}

func (c Conditions) Matches(halfPerm uint32) bool {
	for _, condition := range c {
		if !condition.Matches(halfPerm) {
			return false
		}
	}
	return true
}

func (c Conditions) String() string {
	sl := make([]string, 0, len(c))
	for _, cond := range c {
		sl = append(sl, cond.String())
	}
	return "(" + strings.Join(sl, ",") + ")"
}

func maskFromString(s string) uint32 {
	var mask uint32
	for _, ch := range s {
		if ch < '0' || ch > '8' {
			continue
		}
		bit := uint32(1) << (ch - '0')
		mask |= bit
	}
	return mask
}

type Require uint32

var _ Condition = Require(0)

func NewRequire(s string) Require {
	return Require(maskFromString(s))
}

func (r Require) Matches(halfPerm uint32) bool {
	mask := uint32(r)
	return halfPerm&mask == mask
}

func (r Require) String() string {
	return "+" + maskString(uint32(r))
}

type Forbid uint32

var _ Condition = Forbid(0)

func NewForbid(s string) Forbid {
	return Forbid(maskFromString(s))
}

func (f Forbid) Matches(halfPerm uint32) bool {
	return halfPerm&uint32(f) == 0
}

func (f Forbid) String() string {
	return "!" + maskString(uint32(f))
}

// ExcludeCombination means all bits in the mask may not coexist.
// Individual bits are permitted.
type ExcludeCombination uint32

var _ Condition = ExcludeCombination(0)

func NewExcludeCombination(s string) ExcludeCombination {
	return ExcludeCombination(maskFromString(s))
}

func (e ExcludeCombination) Matches(halfPerm uint32) bool {
	mask := uint32(e)
	return halfPerm&mask != mask
}

func (e ExcludeCombination) String() string {
	return "-" + maskString(uint32(e))
}

type Cardinality struct {
	Mask uint32
	Min  uint32
	Max  uint32
}

var _ Condition = Cardinality{}

func (c Cardinality) Matches(halfPerm uint32) bool {
	count := uint32(bits.OnesCount32(halfPerm & c.Mask))
	return count >= c.Min && count <= c.Max
}

func (c Cardinality) String() string {
	return fmt.Sprintf(
		"#%d..%d:%s",
		c.Min,
		c.Max,
		maskString(c.Mask),
	)
}

func maskString(mask uint32) string {
	var s strings.Builder
	for i := 0; i <= 8; i++ {
		if mask&(1<<i) != 0 {
			s.WriteByte(byte('0' + i))
		}
	}
	return s.String()
}
