package meta

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConditions_String(t *testing.T) {
	c := Conditions{
		NewRequire("012345678"),
		NewForbid("012345678"),
		NewExcludeCombination("012345678"),
		Cardinality{
			Mask: 0x1ff,
			Min:  1,
			Max:  3,
		},
	}
	assert.Equal(t, "(+012345678,!012345678,-012345678,#1..3:012345678)", c.String())
}

func TestConditions_Matches(t *testing.T) {
	testCases := []struct {
		conditions Conditions
		input      uint32
		expected   bool
	}{
		{
			conditions: Conditions{},
			input:      0,
			expected:   true,
		},
		{
			conditions: Conditions{
				NewRequire("01"),
				NewForbid("456789"),
				NewExcludeCombination("0123"),
			},
			input:    0,
			expected: false,
		},
		{
			conditions: Conditions{
				NewRequire("01"),
				NewForbid("456789"),
				NewExcludeCombination("0123"),
			},
			input:    0b11,
			expected: true,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.conditions.Matches(tc.input))
		})
	}
}

func TestNewRequire(t *testing.T) {
	c := NewRequire("")
	assert.Equal(t, uint32(0), uint32(c))
	c = NewRequire("012345678")
	assert.Equal(t, uint32(0x1ff), uint32(c))
	assert.Equal(t, "+012345678", c.String())
}

func TestRequire_Matches(t *testing.T) {
	testCases := []struct {
		s        string
		input    uint32
		expected bool
	}{
		{
			s:        "",
			input:    0,
			expected: true,
		},
		{
			s:        "0",
			input:    0,
			expected: false,
		},
		{
			s:        "0",
			input:    1,
			expected: true,
		},
		{
			s:        "012345678",
			input:    15,
			expected: false,
		},
		{
			s:        "012345678",
			input:    0x1ff,
			expected: true,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			c := NewRequire(tc.s)
			assert.Equal(t, tc.expected, c.Matches(tc.input))
		})
	}
}

func TestNewForbid(t *testing.T) {
	c := NewForbid("")
	assert.Equal(t, uint32(0), uint32(c))
	c = NewForbid("012345678")
	assert.Equal(t, uint32(0x1ff), uint32(c))
	assert.Equal(t, "!012345678", c.String())
}

func TestForbid_Matches(t *testing.T) {
	testCases := []struct {
		s        string
		input    uint32
		expected bool
	}{
		{
			s:        "",
			input:    0,
			expected: true,
		},
		{
			s:        "0",
			input:    0,
			expected: true,
		},
		{
			s:        "0",
			input:    1,
			expected: false,
		},
		{
			s:        "012345678",
			input:    15,
			expected: false,
		},
		{
			s:        "012345678",
			input:    0x1ff,
			expected: false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			c := NewForbid(tc.s)
			assert.Equal(t, tc.expected, c.Matches(tc.input))
		})
	}
}

func TestNewExcludeCombination(t *testing.T) {
	c := NewExcludeCombination("")
	assert.Equal(t, uint32(0), uint32(c))
	c = NewExcludeCombination("012345678")
	assert.Equal(t, uint32(0x1ff), uint32(c))
	assert.Equal(t, "-012345678", c.String())
}

func TestExcludeCombination_Matches(t *testing.T) {
	testCases := []struct {
		s        string
		input    uint32
		expected bool
	}{
		{
			s:        "",
			input:    0,
			expected: false,
		},
		{
			s:        "0",
			input:    0,
			expected: true,
		},
		{
			s:        "0",
			input:    1,
			expected: false,
		},
		{
			s:        "012345678",
			input:    15,
			expected: true,
		},
		{
			s:        "012345678",
			input:    0x1ff,
			expected: false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			c := NewExcludeCombination(tc.s)
			assert.Equal(t, tc.expected, c.Matches(tc.input))
		})
	}
}

func TestCardinality(t *testing.T) {
	testCases := []struct {
		c        Cardinality
		s        string
		input    uint32
		expected bool
	}{
		{
			c:        Cardinality{},
			s:        "#0..0:",
			input:    0,
			expected: true,
		},
		{
			c: Cardinality{
				Mask: 0x1ff,
				Min:  1,
				Max:  2,
			},
			s:        "#1..2:012345678",
			input:    0x1ff,
			expected: false,
		},
		{
			c: Cardinality{
				Mask: 0x1ff,
				Min:  1,
				Max:  2,
			},
			s:        "#1..2:012345678",
			input:    0xf,
			expected: false,
		},
		{
			c: Cardinality{
				Mask: 0x1ff,
				Min:  1,
				Max:  2,
			},
			s:        "#1..2:012345678",
			input:    0b11,
			expected: true,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			assert.Equal(t, tc.s, tc.c.String())
			assert.Equal(t, tc.expected, tc.c.Matches(tc.input))
		})
	}
}

func Test_maskFromString(t *testing.T) {
	testCases := []struct {
		input    string
		expected uint32
	}{
		{
			input:    "",
			expected: 0,
		},
		{
			input:    "0",
			expected: 0x01,
		},
		{
			input:    "1",
			expected: 0x02,
		},
		{
			input:    "2",
			expected: 0x04,
		},
		{
			input:    "3",
			expected: 0x08,
		},
		{
			input:    "4",
			expected: 0x10,
		},
		{
			input:    "5",
			expected: 0x20,
		},
		{
			input:    "6",
			expected: 0x40,
		},
		{
			input:    "7",
			expected: 0x80,
		},
		{
			input:    "8",
			expected: 0x100,
		},
		{
			input:    "9",
			expected: 0x0,
		},
		{
			input:    "012345678",
			expected: 0x1ff,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			n := maskFromString(tc.input)
			assert.Equal(t, tc.expected, n)
		})
	}
}
