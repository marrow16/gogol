package patterns

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestPatternRleDecoder(t *testing.T) {
	const rle = `#N Canada goose
#O Jason Summers
#C A c/4 period 4 spaceship. At the time of its discovery, the Canada goose was the smallest known diagonal spaceship other than the glider, but this record has since been beaten
#C first by Orion 2, and more recently by the crab.
#C www.conwaylife.com/wiki/index.php?title=Canada_goose
#r B3/S23
#P somewhere
x = 13, y = 12, rule = B3/S23
3o10b $o9b2ob$bo6b3obo$3b2o2b2o4b$4bo8b$8bo4b$4b2o3bo3b$3bobob2o4b$3bob
o2bob2ob$2bo4b2o4b$2b2o9b$2b2o!`
	p, err := PatternRleDecoder(strings.NewReader(rle))
	require.NoError(t, err)
	assert.Equal(t, "Canada goose", p.Name)
	assert.Equal(t, "Jason Summers", p.Origination)
	assert.Equal(t, "somewhere", p.Coordinates)
	assert.Len(t, p.Comments, 3)
	assert.Equal(t, 13, p.Width)
	assert.Equal(t, 12, p.Height)
	assert.Len(t, p.Cells, 12*13)
	expectedPattern := []bool{
		true, true, true, false, false, false, false, false, false, false, false, false, false,
		true, false, false, false, false, false, false, false, false, false, true, true, false,
		false, true, false, false, false, false, false, false, true, true, true, false, true,
		false, false, false, true, true, false, false, true, true, false, false, false, false,
		false, false, false, false, true, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, true, false, false, false, false,
		false, false, false, false, true, true, false, false, false, true, false, false, false,
		false, false, false, true, false, true, false, true, true, false, false, false, false,
		false, false, false, true, false, true, false, false, true, false, true, true, false,
		false, false, true, false, false, false, false, true, true, false, false, false, false,
		false, false, true, true, false, false, false, false, false, false, false, false, false,
		false, false, true, true, false, false, false, false, false, false, false, false, false,
	}
	assert.Equal(t, expectedPattern, p.Cells)
}

func TestPatternRleDecoder_Errors(t *testing.T) {
	testCases := []struct {
		rle string
		err string
	}{
		{
			rle: "",
			err: "invalid RLE format - no data",
		},
		{
			rle: "#r invalid rule",
			err: "invalid RLE rule",
		},
		{
			rle: "x = 0 not enough args",
			err: "invalid RLE format - bad dimension line",
		},
		{
			rle: "x = not a number, y = 0",
			err: "invalid RLE format - bad dimension line",
		},
		{
			rle: "x = 0, y = not a number",
			err: "invalid RLE format - bad dimension line",
		},
		{
			rle: "x = 0, y = 0, rule = bad rule",
			err: "invalid RLE rule",
		},
		{
			rle: "x = 0, y = 0, rule = bad rule:bad",
			err: "invalid RLE rule",
		},
		{
			rle: "x = 0, y = 0, rule = 9/9",
			err: "invalid RLE rule",
		},
		{
			rle: `x = 1, y = 1
b`,
			err: "invalid RLE format - no end",
		},
		{
			rle: `x = 1, y = 1
b$b!`,
			err: "invalid RLE format - too many rows",
		},
		{
			rle: `x = 1, y = 2
b2$b!`,
			err: "invalid RLE format - too many rows",
		},
		{
			rle: `x = 1, y = 1
0b!`,
			err: "invalid RLE format - bad run length",
		},
		{
			rle: `x = 1, y = 2
0$!`,
			err: "invalid RLE format - bad run length",
		},
		{
			rle: `x = 1, y = 2
3b!`,
			err: "invalid RLE format - row exceeds width",
		},
		{
			rle: `x = 1, y = 2
3?!`,
			err: "invalid RLE format",
		},
		{
			rle: `x = 1, y = 2
b2!`,
			err: "invalid RLE format - dangling run length",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			_, err := PatternRleDecoder(strings.NewReader(tc.rle))
			require.Error(t, err)
			assert.Equal(t, tc.err, err.Error())
		})
	}
}

func TestParseRLEData_Raw(t *testing.T) {
	cells, err := ParseRLEData(2, 2, "bo$ob")
	require.NoError(t, err)
	assert.Len(t, cells, 4)
	assert.Equal(t, []bool{false, true, true, false}, cells)
}

func TestParseRLEData_Errors(t *testing.T) {
	testCases := []struct {
		width, height int
		rle           string
		err           string
	}{
		{
			width:  1,
			height: 1,
			rle:    `2b`,
			err:    "invalid RLE format - row exceeds width",
		},
		{
			width:  1,
			height: 1,
			rle:    `b$o`,
			err:    "invalid RLE format - too many rows",
		},
		{
			width:  1,
			height: 2,
			rle:    `b2$o`,
			err:    "invalid RLE format - too many rows",
		},
		{
			width:  1,
			height: 1,
			rle:    `b0`,
			err:    "invalid RLE format - dangling run length",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			_, err := ParseRLEData(tc.width, tc.height, tc.rle)
			require.Error(t, err)
			assert.Equal(t, tc.err, err.Error())
		})
	}
}

/*
const patternsPath = "./../_patterns/standard"

func TestRlePatternsWalk(t *testing.T) {
	count := 0
	_ = filepath.WalkDir(patternsPath, func(path string, de os.DirEntry, err error) error {
		if !de.IsDir() {
			testRleDecodeFile(t, de.Name())
			count++
		}
		return nil
	})
	fmt.Println("count:", count)
}

func testRleDecodeFile(t *testing.T, filename string) Pattern {
	f, err := os.Open(patternsPath + "/" + filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	p, err := PatternRleDecoder(f)
	assert.NoError(t, err, filename)
	assert.True(t, p.Height <= 100)
	assert.True(t, p.Width <= 100)
	return p
}
*/
