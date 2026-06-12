package recipes

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRecipeUnmarshal(t *testing.T) {
	const j = `{
  "name": "my new script",
  "grid": {
    "width": 100,
    "height": 100,
    "rule": "B3/S23",
    "wrap": "toroidal",
    "boundary": "dead"
  },
  "vars": {
    "pattern": 1,
    "row": "0b000101"
  },
  "patterns": {
    "glider": "Fungus Glider",
    "glider2": {
      "filename": "fungus_bar_blinker.rle"
    },
    "my-pattern": {
      "width": 4,
      "height": 6,
      "rle": "4b$4b$4o$4o$4b$4b!"
    }
  },
  "do": [
    {
      "at": {"x":  0, "y":  0},
      "place": "pattern",
      "repeat": 5,
      "do": [
        {
          "place": "row",
          "move": {"y": 1, "when": "after"},
          "ops": [
            {
              "target": "row",
              "shift-left": 1,
              "when": "after"
            }
          ]
        }
      ]
    },
    {
      "at": {"x": 0, "y": 10},
      "place": "glider",
      "rotate": 1
    },
    {
      "at": {"x": 0, "y": 20},
      "place": "my-pattern"
    },
    {
      "at": {"x": 0, "y": 30},
      "place": "glider2"
    }
  ]
}`
	r := &Recipe{}
	err := json.Unmarshal([]byte(j), r)
	require.NoError(t, err)
	fmt.Printf("%+v\n", r)
}

func TestNumberToPattern(t *testing.T) {
	testCases := []struct {
		str         string
		expectWidth int
		expectCells []bool
	}{
		{
			str:         "0",
			expectWidth: 1,
			expectCells: []bool{false},
		},
		{
			str:         "1",
			expectWidth: 1,
			expectCells: []bool{true},
		},
		{
			str:         `"0b101"`,
			expectWidth: 3,
			expectCells: []bool{true, false, true},
		},
		{
			str:         `"0b0101"`,
			expectWidth: 4,
			expectCells: []bool{false, true, false, true},
		},
		{
			str:         `"0xF"`,
			expectWidth: 4,
			expectCells: []bool{true, true, true, true},
		},
		{
			str:         `"0x0F"`,
			expectWidth: 8,
			expectCells: []bool{false, false, false, false, true, true, true, true},
		},
		{
			str:         `"0o7"`,
			expectWidth: 3,
			expectCells: []bool{true, true, true},
		},
		{
			str:         `"0o4"`,
			expectWidth: 3,
			expectCells: []bool{true, false, false},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			n := Number(tc.str)
			p := n.toPattern(0, 0, 0, 0)
			assert.Equal(t, tc.expectWidth, p.Width)
			assert.Equal(t, tc.expectCells, p.Cells)
		})
	}
}

func TestOpPerformOnVar(t *testing.T) {
	s := "x:0b001"
	num := Number(s)
	nump := &num
	v := 1
	op := Op{
		ShiftLeft: &v,
	}
	op.performOnVar(nump)
	assert.Equal(t, "x:0b010", string(*nump))
}
