package patterns

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestPatternRleEncoderFull(t *testing.T) {
	const rle = `#N Canada goose
#O Jason Summers
#C A c/4 period 4 spaceship. At the time of its discovery, the Canada goose was the smallest known diagonal spaceship other than the glider, but this record has since been beaten
#C first by Orion 2, and more recently by the crab.
#C www.conwaylife.com/wiki/index.php?title=Canada_goose
x = 13, y = 12, rule = B3/S23
3o10b$o9b2ob$bo6b3obo$3b2o2b2o4b$4bo8b$8bo4b$4b2o3bo3b$3bobob2o4b$3bob
o2bob2ob$2bo4b2o4b$2b2o9b$2b2o!`
	p, err := PatternRleDecoder(strings.NewReader(rle))
	require.NoError(t, err)

	var b bytes.Buffer
	err = PatternRleEncode(p, &b)
	require.NoError(t, err)

	p2, err := NewPatternFromRle(strings.NewReader(b.String()))
	require.NoError(t, err)
	require.Equal(t, p.Width, p2.Width)
	require.Equal(t, p.Height, p2.Height)
	require.Equal(t, p.Cells, p2.Cells)
	require.Equal(t, p.Name, p2.Name)
	require.Equal(t, p.Origination, p2.Origination)
	require.Equal(t, p.Comments, p2.Comments)
}

func TestPatternRleEncoder(t *testing.T) {
	testCases := []struct {
		pattern   Pattern
		expect    string
		expectErr bool
	}{
		{
			expectErr: true,
		},
		{
			pattern: Pattern{
				Width:  5,
				Height: 5,
				Cells: []bool{
					false, false, false, false, false,
					false, false, false, true, false,
					false, true, false, true, false,
					false, false, true, true, false,
					false, false, false, false, false},
			},
			expect: `x = 5, y = 5
$3bo$bobo$2b2o!`,
		},
		{
			pattern: Pattern{
				Width:  5,
				Height: 5,
				Cells: []bool{
					false, false, false, false, false,
					false, false, false, false, false,
					false, false, false, false, false,
					false, false, false, false, false,
					false, false, false, false, false},
			},
			expect: `x = 5, y = 5
!`,
		},
		{
			pattern: Pattern{
				Width:  5,
				Height: 5,
				Cells: []bool{
					false, false, false, false, false,
					false, false, false, false, false,
					false, false, false, false, false,
					false, false, false, false, false,
					false, false, false, false, true},
			},
			expect: `x = 5, y = 5
4$4bo!`,
		},
		{
			pattern: Pattern{
				Width:  5,
				Height: 5,
				Cells: []bool{
					true, false, false, false, false,
					false, false, false, false, false,
					false, false, false, false, false,
					false, false, false, false, false,
					false, false, false, false, false},
			},
			expect: `x = 5, y = 5
o!`,
		},
		{
			pattern: Pattern{
				Width:  5,
				Height: 5,
				Cells: []bool{
					false, false, true, false, false,
					false, false, false, false, false,
					false, false, false, false, false,
					false, false, false, false, false,
					false, false, false, false, true},
			},
			expect: `x = 5, y = 5
2bo4$4bo!`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			var b bytes.Buffer
			err := PatternRleEncode(tc.pattern, &b)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expect, b.String())
				p2, err := NewPatternFromRle(strings.NewReader(b.String()))
				require.NoError(t, err)
				assert.Equal(t, tc.pattern.Width, p2.Width)
				assert.Equal(t, tc.pattern.Height, p2.Height)
				assert.Equal(t, tc.pattern.Cells, p2.Cells)
			}
		})
	}
}

func TestRuns(t *testing.T) {
	testCases := []struct {
		row    []bool
		expect string
	}{
		{
			row:    []bool{false, false, true, true},
			expect: "2b2o",
		},
		{
			row:    []bool{false, false, false, false},
			expect: "4b",
		},
		{
			row:    []bool{true, true, true, true},
			expect: "4o",
		},
		{
			row:    []bool{true, false, true, false},
			expect: "obob",
		},
		{
			row:    []bool{false, true, false, true},
			expect: "bobo",
		},
		{
			row:    []bool{true, true, true, false, false, false, false, false, false, false, false, false, false},
			expect: "3o10b",
		},
		{
			row:    []bool{true, false, false, false, false, false, false, false, false, false, true, true, false},
			expect: "o9b2ob",
		},
		{
			row:    []bool{false, true, false, false, false, false, false, false, true, true, true, false, true},
			expect: "bo6b3obo",
		},
		{
			row:    []bool{false, false, false, true, true, false, false, true, true, false, false, false, false},
			expect: "3b2o2b2o4b",
		},
		{
			row:    []bool{false, false, false, false, true, false, false, false, false, false, false, false, false},
			expect: "4bo8b",
		},
		{
			row:    []bool{false, false, false, false, false, false, false, false, true, false, false, false, false},
			expect: "8bo4b",
		},
		{
			row:    []bool{false, false, false, false, true, true, false, false, false, true, false, false, false},
			expect: "4b2o3bo3b",
		},
		{
			row:    []bool{false, false, false, true, false, true, false, true, true, false, false, false, false},
			expect: "3bobob2o4b",
		},
		{
			row:    []bool{false, false, false, true, false, true, false, false, true, false, true, true, false},
			expect: "3bobo2bob2ob",
		},
		{
			row:    []bool{false, false, true, false, false, false, false, true, true, false, false, false, false},
			expect: "2bo4b2o4b",
		},
		{
			row:    []bool{false, false, true, true, false, false, false, false, false, false, false, false, false},
			expect: "2b2o9b",
		},
		{
			row:    []bool{false, false, true, true, false, false, false, false, false, false, false, false, false},
			expect: "2b2o9b",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			s := strings.Join(runs(tc.row), "")
			assert.Equal(t, tc.expect, s)
		})
	}
}

/*
func TestRleEncodingAgainstPatternsLibrary(t *testing.T) {
	_ = filepath.WalkDir(patternsPath, func(path string, de os.DirEntry, err error) error {
		if !de.IsDir() {
			p := testRleDecodeFile(t, de.Name())

			var b bytes.Buffer
			w := &rleWriter{w: &b}
			w.writeData(p.Width, p.Cells)
			require.NoError(t, w.err)
			if err == nil {
				p2, err := NewPatternFromRle(strings.NewReader(fmt.Sprintf("x = %d, y = %d\n%s!", p.Width, p.Height, b.String())))
				require.NoError(t, err)
				require.Equal(t, p.Width, p2.Width)
				require.Equal(t, p.Height, p2.Height)
				require.Equal(t, p.Cells, p2.Cells)
			}
		}
		return nil
	})
}
*/
