package patterns

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPatternRleDecoder(t *testing.T) {
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
	assert.Equal(t, "Canada goose", p.Name)
	assert.Equal(t, "Jason Summers", p.Origination)
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

const patternsPath = "./../_patterns"

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

func testRleDecodeFile(t *testing.T, filename string) {
	f, err := os.Open(patternsPath + "/" + filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	p, err := PatternRleDecoder(f)
	assert.NoError(t, err, filename)
	assert.True(t, p.Height <= 100)
	assert.True(t, p.Width <= 100)
}
