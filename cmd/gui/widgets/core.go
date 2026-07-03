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
	"image/color"
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
		c.instrumentRecord = &logic.RecordInstrument{Grid: c.gridHolder.grid}
	}
	if s.RepeatDetection {
		c.instrumentRepeat = logic.NewRepeatInstrument(c.gridHolder.grid)
	}
	c.updateInstrumentation()
	return c, err
}

type mode int

const (
	noMode mode = iota
	editMode
	placePatternMode
)

func (m mode) String() string {
	switch m {
	case editMode:
		return "Edit Mode (Esc exit)"
	case placePatternMode:
		return "Place (Enter or Esc exit)"
	}
	return ""
}

type Core struct {
	settings    *settings.Settings
	window      *app.Window
	windowRect  clip.Rect
	theme       *material.Theme
	gridHolder  *gridHolder
	statusBar   *statusBar
	gridRecipes *gridRecipesPopout

	mode            mode
	running         bool
	stopRun         chan struct{}
	status          string
	stepAheadQueued bool
	skipBackQueued  bool
	mutex           sync.Mutex

	snapshots     []patterns.Pattern
	snapshotsStep []uint64

	instrumentRepeat *logic.RepeatInstrument
	instrumentRecord *logic.RecordInstrument
	instrumentation  logic.CompositeInstrument

	// pattern placing...
	placePatternCol, placePatternRow int
	placePatternRotation             patterns.Rotation
}

var (
	bgColor = color.NRGBA{R: 147, G: 147, B: 147, A: 255}
)

func (c *Core) Run(window *app.Window) error {
	c.window = window
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			c.stop()
			c.settings.Recording = c.instrumentRecord != nil
			c.settings.RepeatDetection = c.instrumentRepeat != nil
			c.settings.Save(c.gridHolder.grid, c.gridHolder.zoom)
			return e.Err
		case app.FrameEvent:
			var ops op.Ops
			gtx := app.NewContext(&ops, e)

			c.handleKeys(gtx)

			c.windowRect = clip.Rect{Max: gtx.Constraints.Max}
			c.settings.ScreenWidth = int(float32(c.windowRect.Max.X) / gtx.Metric.PxPerDp)
			c.settings.ScreenHeight = int(float32(c.windowRect.Max.Y) / gtx.Metric.PxPerDp)
			paint.FillShape(gtx.Ops, bgColor, c.windowRect.Op())
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

func (c *Core) modeDisplay() string {
	if c.mode != noMode {
		s := c.mode.String()
		if c.mode == editMode {
			if c.gridHolder.editor.patternRotation != patterns.Rotate0 {
				s = strconv.Itoa(c.gridHolder.editor.row) + "x" + strconv.Itoa(c.gridHolder.editor.col) + " " + c.gridHolder.editor.patternRotation.String() + " " + s
			} else {
				s = strconv.Itoa(c.gridHolder.editor.row) + "x" + strconv.Itoa(c.gridHolder.editor.col) + " " + s
			}
		}
		return s
	}
	return ""
}

func (c *Core) clearMode() {
	c.mode = noMode
	c.gridHolder.overlay = nil
	c.gridHolder.stopEditing()
}

// B0135/S1 snow flakes in coal mine
// B0135/S03 has gliders
// B36/S237 like standard life but "never" properly stabilises?
// B023/S1234 mazes with flashing patches

var keyFilters = []event.Filter{
	key.Filter{
		Required: key.ModAlt,
		Name:     "",
	},
	key.Filter{Name: key.NameEscape},
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
				if c.mode == noMode {
					if fn, ok := altCommands[evt.Name]; ok && evt.Modifiers == key.ModAlt {
						fn(gtx, c)
						return
					}
				}
				switch evt.Name {
				case key.NameEscape:
					c.statusBar.showHidePopup(popupNone)
					c.clearMode()
				case "0", "1", "2", "3", "4", "5", "6", "7", "8":
					if evt.Modifiers == key.ModCtrl || evt.Modifiers == key.ModCommand {
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
		c.editMode()
		gtx.Execute(key.FocusCmd{Tag: &c.gridHolder.clickable})
	},
	"G": func(gtx layout.Context, c *Core) {
		if c.gridRecipes != nil {
			c.stop()
			c.gridRecipes.runRecipe()
		}
	},
}
