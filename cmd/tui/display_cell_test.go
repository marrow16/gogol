package main

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestQuadRune_Update(t *testing.T) {
	testCases := []struct {
		qr       quadRune
		xOff     int
		yOff     int
		active   bool
		expected int
	}{
		{
			qr:       quadCell[0],
			xOff:     0,
			yOff:     0,
			active:   true,
			expected: 1,
		},
		{
			qr:       quadCell[0b0001],
			xOff:     0,
			yOff:     0,
			active:   false,
			expected: 0,
		},
		{
			qr:       quadCell[0b1111],
			xOff:     0,
			yOff:     0,
			active:   false,
			expected: 0b1110,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.qr.update(tc.xOff, tc.yOff, tc.active, true).toInt())
		})
	}
}
