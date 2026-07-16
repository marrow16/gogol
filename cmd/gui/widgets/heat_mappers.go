package widgets

import (
	"github.com/marrow16/gogol/logic"
	"iter"
	"strings"
)

type heatMapperType int

const (
	noHeatMapper heatMapperType = iota
	activityHeatMapper
	occupancyHeatMapper
	birthsHeatMapper
	freshnessHeatMapper
	phaseParityHeatMapper
	allHeatMapper
)

func (hmt heatMapperType) newHeatMapper(g *logic.Grid, halfLife float32) logic.HeatMap {
	switch hmt {
	case activityHeatMapper:
		return logic.NewActivityHeatMapInstrument(g)
	case occupancyHeatMapper:
		return logic.NewOccupancyHeatMapInstrument(g)
	case freshnessHeatMapper:
		return logic.NewFreshnessHeatMapInstrument(g, halfLife)
	case phaseParityHeatMapper:
		return logic.NewPhaseHeatMapInstrument(g)
	case birthsHeatMapper:
		return logic.NewBirthsHeatMapInstrument(g, halfLife)
	case allHeatMapper:
		return newAllHeatMapInstrument(g, halfLife)
	default:
		return nil
	}
}

func (hmt heatMapperType) String() string {
	switch hmt {
	case activityHeatMapper:
		return "Activity"
	case occupancyHeatMapper:
		return "Occupancy"
	case freshnessHeatMapper:
		return "Freshness"
	case phaseParityHeatMapper:
		return "Phase Parity"
	case birthsHeatMapper:
		return "Births"
	case allHeatMapper:
		return "All"
	default:
		return "None"
	}
}

func heatMapperTypeFrom(s string) heatMapperType {
	switch strings.ToLower(s) {
	case "activity":
		return activityHeatMapper
	case "occupancy":
		return occupancyHeatMapper
	case "freshness":
		return freshnessHeatMapper
	case "phase parity":
		return phaseParityHeatMapper
	case "births":
		return birthsHeatMapper
	case "all":
		return allHeatMapper
	default:
		return noHeatMapper
	}
}

func newAllHeatMapInstrument(g *logic.Grid, decay float32) *allHeatMapInstrument {
	result := &allHeatMapInstrument{
		showType:    activityHeatMapper,
		heatMappers: make(map[heatMapperType]logic.HeatMap),
		instruments: make([]logic.DualUseInstrumentation, 0),
	}
	for _, hmt := range []heatMapperType{activityHeatMapper, occupancyHeatMapper, freshnessHeatMapper, birthsHeatMapper, phaseParityHeatMapper} {
		hm := hmt.newHeatMapper(g, decay)
		result.heatMappers[hmt] = hm
		result.instruments = append(result.instruments, hm.(logic.DualUseInstrumentation))
	}
	return result
}

type allHeatMapInstrument struct {
	showType    heatMapperType
	heatMappers map[heatMapperType]logic.HeatMap
	instruments []logic.DualUseInstrumentation
}

var _ logic.HeatMap = (*allHeatMapInstrument)(nil)
var _ logic.StepInstrumentation = (*allHeatMapInstrument)(nil)
var _ logic.StepStopInstrumentation = (*allHeatMapInstrument)(nil)
var _ logic.DualUseInstrumentation = (*allHeatMapInstrument)(nil)

func (a *allHeatMapInstrument) cycleShowType() {
	switch a.showType {
	case activityHeatMapper:
		a.showType = occupancyHeatMapper
	case occupancyHeatMapper:
		a.showType = birthsHeatMapper
	case birthsHeatMapper:
		a.showType = freshnessHeatMapper
	case freshnessHeatMapper:
		a.showType = phaseParityHeatMapper
	case phaseParityHeatMapper:
		a.showType = activityHeatMapper
	}
}

func (a *allHeatMapInstrument) HeatMap() iter.Seq[logic.HeatLocation] {
	return a.heatMappers[a.showType].HeatMap()
}

func (a *allHeatMapInstrument) Maximum() uint64 {
	return a.heatMappers[a.showType].Maximum()
}

func (a *allHeatMapInstrument) StepsCount() uint64 {
	return a.heatMappers[a.showType].StepsCount()
}

func (a *allHeatMapInstrument) InstrumentStop(step uint64, changes []*logic.Cell, locations [][2]int) bool {
	for _, i := range a.instruments {
		i.InstrumentStop(step, changes, locations)
	}
	return false
}

func (a *allHeatMapInstrument) Instrument(step uint64, changes []*logic.Cell, locations [][2]int) {
	for _, i := range a.instruments {
		i.Instrument(step, changes, locations)
	}
}
