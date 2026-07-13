package widgets

import (
	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
	"github.com/marrow16/gogol/cmd/gui/settings"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"strconv"
	"sync"
)

func NewCore(s *settings.Settings) (*Core, error) {
	c := &Core{
		settings: s,
		theme:    material.NewTheme(),
	}
	var err error
	c.gridHolder, err = newGridHolder(c)
	c.statusBar = newStatusBar(c)
	if s.Recording {
		c.instrumentRecord = logic.NewRecordInstrument(c.gridHolder.grid)
	}
	if s.RepeatDetection {
		c.instrumentRepeat = logic.NewRepeatInstrument(c.gridHolder.grid)
	}
	if s.HeatMappingType != "" {
		c.heatMapperType = heatMapperTypeFrom(s.HeatMappingType)
		c.instrumentHeatMap = c.heatMapperType.newHeatMapper(c.gridHolder.grid, s.HeatMappingHalfLife)
	}
	c.updateInstrumentation()
	return c, err
}

type mode int

const (
	noMode mode = iota
	editMode
	placePatternMode
	heatMapMode
)

func (m mode) String() string {
	switch m {
	case editMode:
		return "Edit Mode (Esc exit)"
	case placePatternMode:
		return "Place (Enter or Esc exit)"
	case heatMapMode:
		return "Heat Map"
	}
	return ""
}

type heatMapperType int

const (
	noHeatMapper heatMapperType = iota
	activityHeatMapper
	occupancyHeatMapper
	freshnessHeatMapper
	phaseParityHeatMapper
	birthsHeatMapper
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
	default:
		return "None"
	}
}

func heatMapperTypeFrom(s string) heatMapperType {
	switch s {
	case "Activity":
		return activityHeatMapper
	case "Occupancy":
		return occupancyHeatMapper
	case "Freshness":
		return freshnessHeatMapper
	case "Phase Parity":
		return phaseParityHeatMapper
	case "Births":
		return birthsHeatMapper
	default:
		return noHeatMapper
	}
}

type Core struct {
	settings    *settings.Settings
	window      *app.Window
	windowRect  clip.Rect
	theme       *material.Theme
	gridHolder  *gridHolder
	statusBar   *statusBar
	gridRecipes *gridRecipesPopout

	mode              mode
	running           bool
	stopRun           chan struct{}
	status            string
	stepAheadQueued   bool
	skipBackQueued    bool
	mutex             sync.Mutex
	settingsListeners []func(*settings.Settings)

	snapshots     []patterns.Pattern
	snapshotsStep []uint64

	instrumentRepeat  *logic.RepeatInstrument
	instrumentRecord  *logic.RecordInstrument
	instrumentHeatMap logic.HeatMap
	heatMapperType    heatMapperType
	instrumentation   logic.CompositeInstrument

	runningShortcut bool
	// pattern placing...
	placePatternCol, placePatternRow int
	placePatternRotation             patterns.Rotation
}

func (c *Core) Run(window *app.Window) error {
	c.window = window
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			c.stop()
			c.settings.Recording = c.instrumentRecord != nil
			c.settings.RepeatDetection = c.instrumentRepeat != nil
			c.settings.HeatMappingType = c.heatMapperType.String()
			c.settings.Save(c.gridHolder.grid, c.gridHolder.zoom)
			return e.Err
		case app.FrameEvent:
			var ops op.Ops
			gtx := app.NewContext(&ops, e)

			c.handleKeys(gtx)

			c.windowRect = clip.Rect{Max: gtx.Constraints.Max}
			c.settings.ScreenWidth = int(float32(c.windowRect.Max.X) / gtx.Metric.PxPerDp)
			c.settings.ScreenHeight = int(float32(c.windowRect.Max.Y) / gtx.Metric.PxPerDp)
			paint.FillShape(gtx.Ops, backgroundColor, c.windowRect.Op())
			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Flexed(1, c.gridHolder.layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return c.statusBar.layout(gtx, c.windowRect)
				}),
			)
			c.statusBar.showPopups(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (c *Core) settingsListen(fn func(*settings.Settings)) {
	c.settingsListeners = append(c.settingsListeners, fn)
}

func (c *Core) settingsChanged() {
	for _, fn := range c.settingsListeners {
		fn(c.settings)
	}
}

func (c *Core) modeDisplay() string {
	if c.mode != noMode {
		s := c.mode.String()
		switch c.mode {
		case editMode:
			if c.gridHolder.editor.patternRotation != patterns.Rotate0 {
				s = strconv.Itoa(c.gridHolder.editor.row) + "x" + strconv.Itoa(c.gridHolder.editor.col) + " " + c.gridHolder.editor.patternRotation.String() + " " + s
			} else {
				s = strconv.Itoa(c.gridHolder.editor.row) + "x" + strconv.Itoa(c.gridHolder.editor.col) + " " + s
			}
		case heatMapMode:
			s = s + " - " + c.heatMapperType.String() + " (Esc exit)"
		}
		return s
	}
	return ""
}

func (c *Core) clearMode() {
	makeDirty := c.mode == heatMapMode
	c.mode = noMode
	c.gridHolder.overlay = nil
	c.gridHolder.stopEditing()
	if makeDirty {
		c.gridHolder.dirty = true
		c.gridHolder.grid.Draw()
	}
}

// B0135/S1 snow flakes in coal mine
// B0135/S03 has gliders
// B36/S237 like standard life but "never" properly stabilises?
// B023/S1234 mazes with flashing patches

// B13/S1236 4088

var keyFilters = []event.Filter{
	key.Filter{Required: key.ModAlt, Name: ""},
	key.Filter{Name: key.NameEscape},
	key.Filter{Required: key.ModCtrl, Name: "0"},
	key.Filter{Required: key.ModCtrl, Name: "1"},
	key.Filter{Required: key.ModCtrl, Name: "2"},
	key.Filter{Required: key.ModCtrl, Name: "3"},
	key.Filter{Required: key.ModCtrl, Name: "4"},
	key.Filter{Required: key.ModCtrl, Name: "5"},
	key.Filter{Required: key.ModCtrl, Name: "6"},
	key.Filter{Required: key.ModCtrl, Name: "7"},
	key.Filter{Required: key.ModCtrl, Name: "8"},
}

func (c *Core) handleKeys(gtx layout.Context) {
	for {
		ev, ok := gtx.Event(keyFilters...)
		if !ok {
			break
		}
		switch evt := ev.(type) {
		case key.Event:
			if evt.State == key.Press {
				if c.mode == noMode && evt.Modifiers == key.ModAlt {
					if c.userShortcutKeys(evt.Name) {
						return
					} else if fn, ok := altCommands[evt.Name]; ok {
						fn(gtx, c)
						return
					}
				}
				switch evt.Name {
				case key.NameEscape:
					c.statusBar.showHidePopup(popupNone)
					c.clearMode()
				case "0", "1", "2", "3", "4", "5", "6", "7", "8":
					if evt.Modifiers == key.ModCtrl {
						c.bornChange(string(evt.Name))
					} else if evt.Modifiers == key.ModAlt {
						c.survivesChange(string(evt.Name))
					}
				default:
					if gtx.Focused(&c.gridHolder.clickable) && (c.gridHolder.editor.active || c.gridHolder.overlay != nil) {
						// in edit mode or place pattern but we swallowed the key - pass it to grid...
						c.gridHolder.handleKeys(gtx, evt)
					}
				}
			}
		}
	}
}

func (c *Core) startPatternPlace(gtx layout.Context, pattern *patterns.Pattern, interlaced bool) {
	c.statusBar.showHidePopup(popupNone)
	c.mode = placePatternMode
	pr, pc := c.placePatternRow, c.placePatternCol
	if pr >= c.gridHolder.grid.Height-1 {
		pr = 0
	}
	if pc >= c.gridHolder.grid.Width-1 {
		pc = 0
	}
	c.gridHolder.overlay = &overlay{
		pattern:    *pattern,
		row:        pr,
		col:        pc,
		rotation:   c.placePatternRotation,
		interlaced: interlaced,
		moved:      true,
	}
	gtx.Execute(key.FocusCmd{Tag: &c.gridHolder.clickable})
}

var altCommands = map[key.Name]func(gtx layout.Context, c *Core){
	key.NameReturn: func(gtx layout.Context, c *Core) {
		if c.running {
			c.stop()
		} else {
			c.start()
		}
	},
	key.NameEnter: func(gtx layout.Context, c *Core) {
		if c.running {
			c.stop()
		} else {
			c.start()
		}
	},
	key.NameEscape: func(gtx layout.Context, c *Core) {
		c.stop()
	},
	key.NameSpace: func(gtx layout.Context, c *Core) {
		c.step()
	},
	key.NameRightArrow: func(gtx layout.Context, c *Core) {
		c.step()
	},
	key.NameTab: func(gtx layout.Context, c *Core) {
		c.stepAhead()
	},
	key.NameDeleteBackward: func(gtx layout.Context, c *Core) {
		c.skipBack()
	},
	key.NameLeftArrow: func(gtx layout.Context, c *Core) {
		c.stepBack()
	},
	"=": func(gtx layout.Context, c *Core) {
		c.zoomIn()
	},
	"-": func(gtx layout.Context, c *Core) {
		c.zoomOut()
	},
	",": func(gtx layout.Context, c *Core) {
		c.permutationDecrement()
	},
	".": func(gtx layout.Context, c *Core) {
		c.permutationIncrement()
	},
	"[": func(gtx layout.Context, c *Core) {
		c.decreaseGridWidth()
	},
	"]": func(gtx layout.Context, c *Core) {
		c.increaseGridWidth()
	},
	";": func(gtx layout.Context, c *Core) {
		c.decreaseGridHeight()
	},
	"'": func(gtx layout.Context, c *Core) {
		c.increaseGridHeight()
	},
	"C": func(gtx layout.Context, c *Core) {
		c.clear()
	},
	"B": func(gtx layout.Context, c *Core) {
		c.toggleCellBorders()
	},
	"R": func(gtx layout.Context, c *Core) {
		c.randomize()
	},
	"N": func(gtx layout.Context, c *Core) {
		c.randomChanges()
	},
	"M": func(gtx layout.Context, c *Core) {
		c.showMenu()
	},
	"L": func(gtx layout.Context, c *Core) {
		c.showLifeRules()
	},
	"P": func(gtx layout.Context, c *Core) {
		if pattern, interlaced := c.statusBar.menuPopup.patternsPopout.currentPattern(); pattern != nil {
			c.startPatternPlace(gtx, pattern, interlaced)
		}
	},
	"S": func(gtx layout.Context, c *Core) {
		c.snapshot()
	},
	"Z": func(gtx layout.Context, c *Core) {
		c.undoToSnapshot()
	},
	"X": func(gtx layout.Context, c *Core) {
		_ = c.export()
	},
	"E": func(gtx layout.Context, c *Core) {
		c.startEditMode()
		gtx.Execute(key.FocusCmd{Tag: &c.gridHolder.clickable})
	},
	"G": func(gtx layout.Context, c *Core) {
		if c.gridRecipes != nil {
			c.stop()
			c.gridRecipes.runRecipe()
		}
	},
	"H": func(gtx layout.Context, c *Core) {
		c.showHeatMap()
	},
}
