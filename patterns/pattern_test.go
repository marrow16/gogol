package patterns

import (
	"github.com/marrow16/gogol/logic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestNewPattern(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		p, err := NewPattern("Test", 3, []bool{true, false, true})
		require.NoError(t, err)
		assert.Equal(t, "Test", p.Name)
	})
	t.Run("errors", func(t *testing.T) {
		_, err := NewPattern("Test", 3, []bool{true, false})
		require.Error(t, err)
	})
	t.Run("must", func(t *testing.T) {
		require.NotPanics(t, func() {
			_ = MustNewPattern("Test", 3, []bool{true, false, true})
		})
	})
	t.Run("must panics", func(t *testing.T) {
		require.Panics(t, func() {
			_ = MustNewPattern("Test", 3, []bool{true, false})
		})
	})
}

func TestNewPatternFromRle(t *testing.T) {
	const data = `#N Glider
x = 3, y = 3, rule = B3/S23
bob$2bo$3o!`
	t.Run("ok", func(t *testing.T) {
		p, err := NewPatternFromRle(strings.NewReader(data))
		require.NoError(t, err)
		assert.Equal(t, "Glider", p.Name)
		assert.Equal(t, 3, p.Height)
		assert.Equal(t, 3, p.Width)
		require.NotNil(t, p.Rule)
		assert.Equal(t, "B3/S23", p.Rule.Rle())
	})
	t.Run("errors", func(t *testing.T) {
		_, err := NewPatternFromRle(strings.NewReader(""))
		require.Error(t, err)
	})
	t.Run("must", func(t *testing.T) {
		require.NotPanics(t, func() {
			_ = MustNewPatternFromRle(strings.NewReader(data))
		})
	})
	t.Run("must panics", func(t *testing.T) {
		require.Panics(t, func() {
			_ = MustNewPatternFromRle(strings.NewReader(""))
		})
	})
}

func TestPattern_Draw(t *testing.T) {
	p, err := NewPattern("Glider", 5, []bool{
		false, false, false, false, false,
		false, false, false, true, false,
		false, true, false, true, false,
		false, false, true, true, false,
		false, false, false, false, false})
	require.NoError(t, err)

	g, err := logic.NewGrid(5, 5, 0, 0)
	require.NoError(t, err)

	p.Draw(g, 0, 0, Rotate0)
	assert.True(t, g.GetCell(1, 3).Alive)
	assert.True(t, g.GetCell(2, 1).Alive)
	assert.True(t, g.GetCell(2, 3).Alive)
	assert.True(t, g.GetCell(3, 2).Alive)
	assert.True(t, g.GetCell(3, 3).Alive)
}
