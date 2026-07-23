package logic

import (
	"errors"
	"strconv"
	"strings"
)

type Rule interface {
	NextState(c *Cell) (nextState bool)
	StateChanged(c *Cell) (changed bool)
	Rle() string
	BornWith() string
	SurvivesWith() string
	Permutation() int
	Name() string
	IsCustom() bool
}

type rule struct {
	name         string
	bornWith     [9]bool
	survivesWith [9]bool
}

func (r rule) NextState(c *Cell) (nextState bool) {
	adjsAlive := c.AdjacentsAlive()
	if c.Alive {
		nextState = r.survivesWith[adjsAlive]
	} else {
		nextState = r.bornWith[adjsAlive]
	}
	return nextState
}

func (r rule) StateChanged(c *Cell) (changed bool) {
	nextState := r.NextState(c)
	return nextState != c.Alive
}

func (r rule) Rle() string {
	var sb strings.Builder
	sb.Grow(18 + 3)
	sb.WriteString("B")
	sb.WriteString(r.BornWith())
	sb.WriteString("/S")
	sb.WriteString(r.SurvivesWith())
	return sb.String()
}

func (r rule) BornWith() string {
	var sb strings.Builder
	sb.Grow(9)
	for i := 0; i < 9; i++ {
		if r.bornWith[i] {
			sb.WriteString(strconv.Itoa(i))
		}
	}
	return sb.String()
}

func (r rule) SurvivesWith() string {
	var sb strings.Builder
	sb.Grow(9)
	for i := 0; i < 9; i++ {
		if r.survivesWith[i] {
			sb.WriteString(strconv.Itoa(i))
		}
	}
	return sb.String()
}

func (r rule) Permutation() int {
	result := 0
	for i := 0; i < 9; i++ {
		if r.bornWith[i] {
			result |= 1 << (i + 9)
		}
		if r.survivesWith[i] {
			result |= 1 << i
		}
	}
	return result
}

func (r rule) Name() string {
	if r.name != "" {
		return r.name
	}
	if rn, ok := rleToName[r.Rle()]; ok {
		return rn
	}
	return "Custom " + r.Rle() + " (" + strconv.Itoa(r.Permutation()) + ")"
}

func (r rule) IsCustom() bool {
	_, known := rleToName[r.Rle()]
	return !known
}

func NewRuleFromPermutation(permutation int) (Rule, error) {
	if permutation < 0 || permutation >= 1<<18 {
		return nil, ErrInvalidPermutation
	}
	r := &rule{}
	for i := 0; i < 9; i++ {
		r.bornWith[i] = permutation&(1<<(i+9)) != 0
		r.survivesWith[i] = permutation&(1<<i) != 0
	}
	return r, nil
}

func MustNewRuleRle(name string, rle string) Rule {
	if r, err := NewRuleRle(name, rle); err != nil {
		panic(err)
	} else {
		return r
	}
}

func NewRuleRle(name string, rle string) (Rule, error) {
	result := &rule{
		bornWith:     [9]bool{},
		survivesWith: [9]bool{},
		name:         name,
	}
	parts := strings.Split(strings.ToUpper(rle), "/")
	b := ""
	s := ""
	foundB, foundS := false, false
	if len(parts) > 0 {
		if strings.HasPrefix(parts[0], "B") {
			foundB = true
			b = parts[0][1:]
		} else if strings.HasPrefix(parts[0], "S") {
			foundS = true
			s = parts[0][1:]
		}
	}
	if len(parts) > 1 {
		if strings.HasPrefix(parts[1], "B") {
			if foundB {
				return nil, ErrInvalidRule
			}
			foundB = true
			b = parts[1][1:]
		} else if strings.HasPrefix(parts[1], "S") {
			if foundS {
				return nil, ErrInvalidRule
			}
			foundS = true
			s = parts[1][1:]
		}
	}
	if !foundB || !foundS {
		return nil, ErrInvalidRule
	}
	for _, ch := range b {
		idx := ch - '0'
		if idx >= 0 && idx <= 8 {
			result.bornWith[idx] = true
		} else {
			return nil, ErrInvalidRule
		}
	}
	for _, ch := range s {
		idx := ch - '0'
		if idx >= 0 && idx <= 8 {
			result.survivesWith[idx] = true
		} else {
			return nil, ErrInvalidRule
		}
	}
	return result, nil
}

var (
	ErrInvalidRule        = errors.New("invalid RLE rule")
	ErrInvalidPermutation = errors.New("invalid permutation")
)
