package widgets

import (
	"fmt"
	"github.com/marrow16/gogol/logic"
	"slices"
	"strconv"
	"strings"
	"time"
)

func (c *Core) shortcutToken(token shortcutToken, repeats []int, rIndex int, now time.Time) string {
	switch token {
	case tokenRule:
		return c.gridHolder.grid.Rule().Rle()
	case tokenBorn:
		return "B" + c.gridHolder.grid.Rule().BornWith()
	case tokenSurvives:
		return "S" + c.gridHolder.grid.Rule().SurvivesWith()
	case tokenPerm:
		return strconv.Itoa(c.gridHolder.grid.Rule().Permutation())
	case tokenInteger:
		return strconv.Itoa(c.gridHolder.grid.Rule().Integer())
	case tokenRand:
		return strconv.Itoa(c.settings.Randomization)
	case tokenStep:
		return strconv.FormatUint(c.gridHolder.grid.StepCount.Load(), 10)
	case tokenRepeatDetection:
		if c.instrumentRepeat != nil {
			return "1"
		}
		return "0"
	case tokenRepeatFound:
		if c.instrumentRepeat != nil {
			if c.instrumentRepeat.Found {
				return "1"
			}
			return "0"
		}
	case tokenRepeatFirst:
		if c.instrumentRepeat != nil {
			return strconv.FormatUint(c.instrumentRepeat.FirstStep, 10)
		}
	case tokenRepeatAt:
		if c.instrumentRepeat != nil {
			return strconv.FormatUint(c.instrumentRepeat.RepeatStep, 10)
		}
	case tokenRepeatPeriod:
		if c.instrumentRepeat != nil {
			return strconv.FormatUint(c.instrumentRepeat.Period, 10)
		}
	case tokenPopulation:
		return strconv.Itoa(c.population())
	case tokenIteration:
		if rIndex < len(repeats) {
			return strconv.Itoa(repeats[rIndex])
		} else if len(repeats) > 0 {
			return ""
		}
	case tokenIterationCurrent:
		if len(repeats) > 0 {
			return strconv.Itoa(repeats[len(repeats)-1])
		}
	case tokenStepAhead:
		return strconv.Itoa(c.settings.StepAheadBy)
	case tokenStepBackBy:
		return strconv.Itoa(c.settings.SkipBackBy)
	case tokenWrapMode:
		return c.gridHolder.grid.WrapMode().String()
	case tokenBoundaryMode:
		return c.gridHolder.grid.BoundaryMode().String()
	case tokenStepDelay:
		return strconv.Itoa(c.settings.StepDelay)
	case tokenGridWidth:
		return strconv.Itoa(c.gridHolder.grid.Width())
	case tokenGridHeight:
		return strconv.Itoa(c.gridHolder.grid.Height())
	case tokenGridSize:
		return strconv.Itoa(c.gridHolder.grid.Width()) + "X" + strconv.Itoa(c.gridHolder.grid.Height())
	case tokenRecord:
		if c.instrumentRecord != nil {
			return "1"
		}
		return "0"
	case tokenRecordSteps:
		if c.instrumentRecord != nil {
			return strconv.Itoa(c.instrumentRecord.FramesCount())
		}
		return "0"
	case tokenHeatMap:
		if c.instrumentHeatMap != nil {
			return "1"
		}
		return "0"
	case tokenHeatMapType:
		if c.instrumentHeatMap != nil {
			return c.instrumentHeatMap.Type().String()
		}
		return logic.NoHeatMapper.String()
	case tokenBorders:
		if c.settings.CellBorders {
			return "1"
		}
		return "0"
	case tokenCellSize:
		return strconv.Itoa(c.settings.CellSize)
	case tokenCellColorAlive:
		clr := c.settings.CellAliveColor
		return fmt.Sprintf("#%02X%02X%02X", clr.R, clr.G, clr.B)
	case tokenCellColorDead:
		clr := c.settings.CellDeadColor
		return fmt.Sprintf("#%02X%02X%02X", clr.R, clr.G, clr.B)
	case tokenCellColorBorder:
		clr := c.settings.CellBorderColor
		return fmt.Sprintf("#%02X%02X%02X", clr.R, clr.G, clr.B)
	case tokenNow:
		return now.Format("2006-01-02 15-04-05") + fmt.Sprintf("-%03d", now.Nanosecond()/1e6)
	case tokenTimestamp:
		return now.Format(time.RFC3339Nano)
	case tokenDateYY:
		return now.Format("06")
	case tokenDateYYYY:
		return now.Format("2006")
	case tokenDateM:
		return now.Format("1")
	case tokenDateMM:
		return now.Format("01")
	case tokenDateD:
		return now.Format("2")
	case tokenDateDD:
		return now.Format("02")
	case tokenDateDth:
		n := now.Day()
		suffix := "th"
		if n < 11 || n > 13 {
			switch n % 10 {
			case 1:
				suffix = "st"
			case 2:
				suffix = "nd"
			case 3:
				suffix = "rd"
			}
		}
		return strconv.Itoa(n) + suffix
	case tokenDateMmm:
		return now.Format("Jan")
	case tokenDateMmmm:
		return now.Format("January")
	case tokenDateDdd:
		return now.Format("Mon")
	case tokenDateDddd:
		return now.Format("Monday")
	case tokenDateWw:
		_, w := now.ISOWeek()
		return fmt.Sprintf("%02d", w)
	case tokenDateW:
		_, w := now.ISOWeek()
		return strconv.Itoa(w)
	case tokenDateWy:
		y, _ := now.ISOWeek()
		return strconv.Itoa(y)
	case tokenTimeHH:
		return now.Format("15")
	case tokenTimeH:
		return now.Format("3")
	case tokenTimePm:
		return strings.ToLower(now.Format("PM"))
	case tokenTimePmUpper:
		return strings.ToUpper(now.Format("PM"))
	case tokenTimeMm:
		return now.Format("04")
	case tokenTimeSs:
		return now.Format("05")
	case tokenTimeNnn:
		return fmt.Sprintf("%03d", now.Nanosecond()/1e6)
	case tokenTimeNnnnnn:
		return fmt.Sprintf("%06d", now.Nanosecond()/1e3)
	case tokenTimeTz:
		return now.Format("-0700")
	case tokenTimeTzhm:
		return now.Format("-07:00")
	case tokenTimeTZz:
		return now.Format("Z0700")
	case tokenTimeTZzhm:
		return now.Format("Z07:00")
	}
	return "_"
}

type shortcutToken int

const (
	tokenRule             shortcutToken = iota // %rule
	tokenBorn                                  // %born
	tokenSurvives                              // %survives
	tokenPerm                                  // %perm
	tokenInteger                               // %integer
	tokenRand                                  // %rand
	tokenNow                                   // %now
	tokenTimestamp                             // %timestamp
	tokenStep                                  // %step
	tokenRepeatDetection                       // %repeat-detecting
	tokenRepeatFound                           // %repeat-found
	tokenRepeatFirst                           // %repeat-first
	tokenRepeatAt                              // %repeat-at
	tokenRepeatPeriod                          // %repeat-period
	tokenPopulation                            // %population
	tokenIteration                             // %r
	tokenIterationCurrent                      // %R
	tokenStepAhead                             // %step-ahead
	tokenStepBackBy                            // %step-back
	tokenWrapMode                              // %wrap-mode
	tokenBoundaryMode                          // %boundary-mode
	tokenStepDelay                             // %step-delay
	tokenGridWidth                             // %grid-width
	tokenGridHeight                            // %grid-height
	tokenGridSize                              // %grid-size - widthXheight
	tokenRecord                                // %record
	tokenRecordSteps                           // %recorded-steps
	tokenHeatMap                               // %heat-map
	tokenHeatMapType                           // %heat-map-type
	tokenBorders                               // %borders
	tokenCellSize                              // %cell-size
	tokenCellColorAlive                        // %cell-color-alive
	tokenCellColorDead                         // %cell-color-dead
	tokenCellColorBorder                       // %cell-color-border
	tokenDateYY
	tokenDateYYYY
	tokenDateM
	tokenDateMM
	tokenDateD
	tokenDateDD
	tokenDateDth
	tokenDateMmm
	tokenDateMmmm
	tokenDateDdd
	tokenDateDddd
	tokenDateWw
	tokenDateW
	tokenDateWy
	tokenTimeHH
	tokenTimeH
	tokenTimePm
	tokenTimePmUpper
	tokenTimeMm
	tokenTimeSs
	tokenTimeNnn
	tokenTimeNnnnnn
	tokenTimeTz
	tokenTimeTzhm
	tokenTimeTZz
	tokenTimeTZzhm
)

var shortcutFormatTokens = sortShortcutFormatTokens()

// sort the token pairs by longest string first
func sortShortcutFormatTokens() []shortcutTokenPair {
	slices.SortFunc(shortcutTokenPairs, func(a, b shortcutTokenPair) int {
		return len(b.string) - len(a.string)
	})
	return shortcutTokenPairs
}

type shortcutTokenPair struct {
	token  shortcutToken
	string string
}

var shortcutTokenPairs = []shortcutTokenPair{
	{tokenRule, "rule"},
	{tokenBorn, "born"},
	{tokenBorn, "rule-born-with"},
	{tokenSurvives, "survives"},
	{tokenSurvives, "rule-survives-with"},
	{tokenPerm, "perm"},
	{tokenPerm, "rule-perm"},
	{tokenInteger, "integer"},
	{tokenInteger, "rule-int"},
	{tokenRand, "rand"},
	{tokenRand, "randomization"},
	{tokenStep, "step"},
	{tokenRepeatDetection, "repeat-detecting"},
	{tokenRepeatFound, "repeat-found"},
	{tokenRepeatFirst, "repeat-first"},
	{tokenRepeatAt, "repeat-at"},
	{tokenRepeatPeriod, "repeat-period"},
	{tokenPopulation, "population"},
	{tokenIteration, "r"},
	{tokenIterationCurrent, "R"},
	{tokenStepAhead, "step-ahead"},
	{tokenStepBackBy, "step-back"},
	{tokenWrapMode, "wrap-mode"},
	{tokenBoundaryMode, "boundary-mode"},
	{tokenStepDelay, "step-delay"},
	{tokenGridWidth, "grid-width"},
	{tokenGridHeight, "grid-height"},
	{tokenGridSize, "grid-size"}, // widthXheight
	{tokenRecord, "record"},
	{tokenRecordSteps, "recorded-steps"},
	{tokenHeatMap, "heat-map"},
	{tokenHeatMapType, "heat-map-type"},
	{tokenBorders, "borders"},
	{tokenCellSize, "cell-size"},
	{tokenCellColorAlive, "cell-color-alive"},
	{tokenCellColorDead, "cell-color-dead"},
	{tokenCellColorBorder, "cell-color-border"},
	{tokenNow, "now"},
	{tokenTimestamp, "timestamp"},
	{tokenDateYY, "YY"},
	{tokenDateYYYY, "YYYY"},
	{tokenDateM, "M"},
	{tokenDateMM, "MM"},
	{tokenDateD, "D"},
	{tokenDateDD, "DD"},
	{tokenDateDth, "Dth"},
	{tokenDateMmm, "Mmm"},
	{tokenDateMmmm, "Mmmm"},
	{tokenDateDdd, "Ddd"},
	{tokenDateDddd, "Dddd"},
	{tokenDateWw, "ww"},
	{tokenDateW, "w"},
	{tokenDateWy, "wy"},
	{tokenTimeHH, "hh"},
	{tokenTimeH, "h"},
	{tokenTimePm, "pm"},
	{tokenTimePmUpper, "PM"},
	{tokenTimeMm, "mm"},
	{tokenTimeSs, "ss"},
	{tokenTimeNnn, "nnn"},
	{tokenTimeNnnnnn, "nnnnnn"},
	{tokenTimeTz, "tz"},
	{tokenTimeTzhm, "tzhm"},
	{tokenTimeTZz, "tzz"},
	{tokenTimeTZzhm, "tzzhm"},
}
