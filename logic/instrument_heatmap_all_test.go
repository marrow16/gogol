package logic

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHeatMapperType_String(t *testing.T) {
	testCases := map[HeatMapperType]string{
		NoHeatMapper:          "None",
		AllHeatMapper:         "All",
		ActivityHeatMapper:    "Activity",
		AgeHeatMapper:         "Age",
		BirthsHeatMapper:      "Births",
		FreshnessHeatMapper:   "Freshness",
		LongevityHeatMapper:   "Longevity",
		OccupancyHeatMapper:   "Occupancy",
		PhaseParityHeatMapper: "Phase Parity",
		-1:                    "None",
	}
	for k, v := range testCases {
		assert.Equal(t, v, k.String())
	}
}

func TestHeatMapperType_Token(t *testing.T) {
	testCases := map[HeatMapperType]string{
		NoHeatMapper:          "X",
		AllHeatMapper:         "AGBFLOP",
		ActivityHeatMapper:    "A",
		AgeHeatMapper:         "G",
		BirthsHeatMapper:      "B",
		FreshnessHeatMapper:   "F",
		LongevityHeatMapper:   "L",
		OccupancyHeatMapper:   "O",
		PhaseParityHeatMapper: "P",
		-1:                    "U",
	}
	for k, v := range testCases {
		assert.Equal(t, v, k.Token())
	}
}

func TestHeatMapperType_New(t *testing.T) {
	testCases := map[HeatMapperType]any{
		NoHeatMapper:          nil,
		AllHeatMapper:         &AllHeatMapInstrument{},
		ActivityHeatMapper:    &ActivityHeatMapInstrument{},
		AgeHeatMapper:         &AgeHeatMapInstrument{},
		BirthsHeatMapper:      &BirthsHeatMapInstrument{},
		FreshnessHeatMapper:   &FreshnessHeatMapInstrument{},
		LongevityHeatMapper:   &LongevityHeatMapInstrument{},
		OccupancyHeatMapper:   &OccupancyHeatMapInstrument{},
		PhaseParityHeatMapper: &PhaseHeatMapInstrument{},
		-1:                    nil,
	}
	g, err := NewGrid(5, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	for k, v := range testCases {
		require.IsType(t, v, k.New(g, 0.5))
	}
}

func TestHeatMapperTypeFrom(t *testing.T) {
	testCases := map[string]HeatMapperType{
		"activity":     ActivityHeatMapper,
		"Activity":     ActivityHeatMapper,
		"age":          AgeHeatMapper,
		"births":       BirthsHeatMapper,
		"freshness":    FreshnessHeatMapper,
		"longevity":    LongevityHeatMapper,
		"occupancy":    OccupancyHeatMapper,
		"phase parity": PhaseParityHeatMapper,
		"all":          AllHeatMapper,
		"":             NoHeatMapper,
		"?":            NoHeatMapper,
	}
	for k, v := range testCases {
		require.Equal(t, v, HeatMapperTypeFrom(k))
	}
}

func TestAllHeatMapInstrument_Step(t *testing.T) {
	g, err := NewGrid(5, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	g.SetCell(0, 2, true)
	g.SetCell(1, 0, true)
	g.SetCell(1, 2, true)
	g.SetCell(2, 1, true)
	g.SetCell(2, 2, true)
	i := NewAllHeatMapInstrument(g, 0.95)
	assert.Equal(t, AllHeatMapper, i.Type())
	for range 100 {
		g.StepWithInstrumentation(i)
	}
	assert.Equal(t, uint64(100), i.StepsCount())
	count := 0
	for hl := range i.HeatMap() {
		count++
		assert.True(t, hl.Value >= 0 && hl.Value <= 1, hl.Value)
	}
	assert.Equal(t, 50, count)
}

func TestAllHeatMapInstrument_StepAhead(t *testing.T) {
	g, err := NewGrid(5, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	g.SetCell(0, 2, true)
	g.SetCell(1, 0, true)
	g.SetCell(1, 2, true)
	g.SetCell(2, 1, true)
	g.SetCell(2, 2, true)
	i := NewAllHeatMapInstrument(g, 0.95)
	g.StepAheadWithInstrumentation(100, i)
	assert.Equal(t, uint64(100), i.StepsCount())
	count := 0
	for hl := range i.HeatMap() {
		count++
		assert.True(t, hl.Value >= 0 && hl.Value <= 1, hl.Value)
	}
	assert.Equal(t, 50, count)
	// coverage on break...
	count = 0
	for range i.HeatMap() {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestAllHeatMapInstrument_Specific(t *testing.T) {
	g, err := NewGrid(5, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	i := NewAllHeatMapInstrument(g, 0.95)
	for _, hmt := range []HeatMapperType{
		ActivityHeatMapper,
		AgeHeatMapper,
		BirthsHeatMapper,
		FreshnessHeatMapper,
		LongevityHeatMapper,
		OccupancyHeatMapper,
		PhaseParityHeatMapper} {
		hm := i.Specific(hmt)
		require.NotNil(t, hm)
		assert.Equal(t, hmt, hm.Type())
	}
	require.Nil(t, i.Specific(AllHeatMapper))
	require.Nil(t, i.Specific(NoHeatMapper))
	require.Nil(t, i.Specific(-1))
}

func TestAllHeatMapInstrument_List(t *testing.T) {
	g, err := NewGrid(5, 10, StandardRule, WrapAll, DeadBoundary)
	require.NoError(t, err)
	i := NewAllHeatMapInstrument(g, 0.95)
	seen := make(map[HeatMapperType]struct{})
	for hm := range i.List() {
		seen[hm.Type()] = struct{}{}
	}
	require.Equal(t, 7, len(seen))
	for _, hmt := range []HeatMapperType{
		ActivityHeatMapper,
		AgeHeatMapper,
		BirthsHeatMapper,
		FreshnessHeatMapper,
		LongevityHeatMapper,
		OccupancyHeatMapper,
		PhaseParityHeatMapper} {
		_, ok := seen[hmt]
		require.True(t, ok)
	}
	// break coverage...
	count := 0
	for range i.List() {
		count++
		break
	}
	assert.Equal(t, 1, count)
}
