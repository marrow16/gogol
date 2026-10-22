package gui

import (
	"encoding/json"
	"errors"
	"fmt"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/dop251/goja"
	"github.com/marrow16/gogol/animator"
	"github.com/marrow16/gogol/gui/shortcuts"
	"github.com/marrow16/gogol/imaging"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/logic/meta"
	"github.com/marrow16/gogol/patterns"
	"github.com/marrow16/gogol/recipes"
	"image/color"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

func (c *Core) shortcutJavascript(ah shortcuts.ActionHolder) {
	var err error
	var filename string
	if filename, err = resolveSavePath(c.shortcutVarString(ah)); err == nil {
		var src []byte
		if src, err = os.ReadFile(filename); err == nil {
			rt := c.createJavascriptRuntime(ah)
			_, err = rt.RunString(string(src))
		}
	}
	if err != nil {
		if _, ok := errors.AsType[*goja.InterruptedError](err); !ok {
			if ge, ok := errors.AsType[*goja.Exception](err); ok {
				if ge.Value() != nil {
					err = fmt.Errorf("uncaught exception: %s", ge.Value().ToString().String())
				}
			}
			c.shortcutErrorLogger.Error(err, &ah)
		} else {
			c.shortcutErrorLogger.Warn("javascript interrupted", &ah)
		}
	}
}

func (c *Core) createJavascriptRuntime(ah shortcuts.ActionHolder) *goja.Runtime {
	rt := goja.New()
	_ = rt.Set("exceptions", false)
	c.registerJavascriptFunctions(rt, ah)
	return rt
}

var jsFunctions = map[string]func(*Core, *goja.Runtime, []goja.Value) (any, error){
	"borders": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		if len(args) > 0 {
			val := args[0].ToBoolean()
			c.setCellBorders(val)
		}
		return c.settings.CellBorders, nil
	},
	"boundaryMode": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].ToString().String()
			if bm := logic.BoundaryModeFromString(val, -1); bm != -1 {
				c.setBoundaryMode(bm)
			} else {
				err = errors.New("invalid/unkown boundary mode")
			}
		}
		return c.gridHolder.grid.BoundaryMode().String(), err
	},
	"cellAliveColor": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].ToString().String()
			if clr, ok := parseColor(val); ok {
				c.stop()
				c.settings.CellAliveColor = clr
				c.gridHolder.grid.Draw()
			} else {
				err = fmt.Errorf("invalid color %q", val)
			}
		}
		clr := c.settings.CellAliveColor
		return fmt.Sprintf("#%02X%02X%02X", clr.R, clr.G, clr.B), err
	},
	"cellDeadColor": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].ToString().String()
			if clr, ok := parseColor(val); ok {
				c.stop()
				c.settings.CellDeadColor = clr
				c.gridHolder.grid.Draw()
			} else {
				err = fmt.Errorf("invalid color %q", val)
			}
		}
		clr := c.settings.CellDeadColor
		return fmt.Sprintf("#%02X%02X%02X", clr.R, clr.G, clr.B), err
	},
	"cellBorderColor": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].ToString().String()
			if clr, ok := parseColor(val); ok {
				c.stop()
				c.settings.CellBorderColor = clr
				c.gridHolder.grid.Draw()
			} else {
				err = fmt.Errorf("invalid color %q", val)
			}
		}
		clr := c.settings.CellBorderColor
		return fmt.Sprintf("#%02X%02X%02X", clr.R, clr.G, clr.B), err
	},
	"cellSize": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val >= 1 && val <= 100 {
				c.setCellSize(int(val))
			} else {
				err = fmt.Errorf("value %d is out of range (%d-%d)", val, 1, 100)
			}
		}
		return c.settings.CellSize, err
	},
	"gridWidth": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val >= 3 && val <= 99999 {
				c.gridResize(c.settings.Height, int(val))
			} else {
				err = fmt.Errorf("value %d is out of range (%d-%d)", val, 3, 99999)
			}
		}
		return c.settings.Width, err
	},
	"gridHeight": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val >= 3 && val <= 99999 {
				c.gridResize(int(val), c.settings.Width)
			} else {
				err = fmt.Errorf("value %d is out of range (%d-%d)", val, 3, 99999)
			}
		}
		return c.settings.Height, err
	},
	"gridSize": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 1 {
			w := args[0].ToInteger()
			h := args[1].ToInteger()
			if w >= 3 && w <= 99999 {
				if h >= 3 && h <= 99999 {
					c.gridResize(int(h), int(w))
				} else {
					err = fmt.Errorf("height %d is out of range (%d-%d)", h, 3, 99999)
				}
			} else {
				err = fmt.Errorf("width %d is out of range (%d-%d)", w, 3, 99999)
			}
		}
		return strconv.Itoa(c.settings.Width) + "X" + strconv.Itoa(c.settings.Height), err
	},
	"randomization": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val >= 0 && val <= 100 {
				c.settings.Randomization = int(val)
			} else {
				err = fmt.Errorf("value %d is out of range (%d-%d)", val, 0, 100)
			}
		}
		return c.settings.Randomization, err
	},
	"stepAheadSize": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val > 0 {
				c.settings.StepAheadBy = int(val)
			} else {
				err = fmt.Errorf("value %d is out of range (min %d)", val, 1)
			}
		}
		return c.settings.StepAheadBy, err
	},
	"stepBackSize": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val > 0 {
				c.settings.SkipBackBy = int(val)
			} else {
				err = fmt.Errorf("value %d is out of range (min %d)", val, 1)
			}
		}
		return c.settings.SkipBackBy, err
	},
	"stepDelay": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val >= 0 && val <= 2000 {
				c.settings.StepDelay = int(val)
			} else {
				err = fmt.Errorf("value %d is out of range (%d-%d)", val, 0, 2000)
			}
		}
		return c.settings.StepDelay, err
	},
	"wrapMode": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].ToString().String()
			if wm := logic.WrapModeFromString(val, -1); wm != -1 {
				c.setWrapMode(wm)
			} else {
				err = errors.New("invalid/unkown wrap mode")
			}
		}
		return c.gridHolder.grid.WrapMode().String(), err
	},
	"zoom": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val >= 10 && val <= 1000 {
				c.zoom(int(val))
			} else {
				err = fmt.Errorf("value %d is out of range (%d-%d)", val, 10, 1000)
			}
		}
		zf := int(c.gridHolder.zoom * 100.0)
		return zf, err
	},
	"rule": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			arg := args[0].Export()
			switch v := arg.(type) {
			case string:
				if rule, ok := logic.Rules[v]; ok {
					c.setRule(rule)
				} else if rule, err = logic.NewRuleRle("", v); err == nil {
					c.setRule(rule)
				}
			case int64:
				var rule logic.Rule
				if rule, err = logic.NewRuleFromPermutation(int(v)); err == nil {
					c.setRule(rule)
				}
			case logic.Rule:
				c.setRule(v)
			default:
				err = errors.New("invalid arg type for rule (expected string, integer or Rule)")
			}
		}
		return c.gridHolder.grid.Rule(), err
	},
	"newRule": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		var rule logic.Rule
		if len(args) > 0 {
			arg := args[0].Export()
			switch v := arg.(type) {
			case string:
				if r, ok := logic.Rules[v]; ok {
					rule = r
				} else if rule, err = logic.NewRuleRle("", v); err == nil {
					c.setRule(rule)
				}
			case int64:
				if rule, err = logic.NewRuleFromPermutation(int(v)); err == nil {
					c.setRule(rule)
				}
			}
		}
		return rule, err
	},
	"record": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		if len(args) > 0 {
			c.setInstrumentationRecord(args[0].ToBoolean())
		}
		return c.instrumentRecord != nil, nil
	},
	"recordInstrument": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		return c.instrumentRecord, nil
	},
	"animationFormat": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		if len(args) > 0 {
			val := args[0].ToString().String()
			if strings.ToLower(val) == "mp4" && animator.Mp4Available() {
				c.settings.AnimationFormat = "mp4"
			} else {
				c.settings.AnimationFormat = "gif"
			}
		}
		return c.settings.AnimationFormat, nil
	},
	"animationSave": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		c.stop()
		recorder := c.instrumentRecord
		if recorder != nil && recorder.FramesCount() > 1 {
			var filename string
			if len(args) > 0 {
				filename, err = resolveSavePath(args[0].ToString().String())
			} else if c.settings.AnimationFormat != "mp4" {
				filename, err = resolveSavePath(c.nowFilename("Grid", ".gif"))
			} else {
				filename, err = resolveSavePath(c.nowFilename("Grid", ".mp4"))
			}
			if err == nil {
				ani := animator.NewAnimator(c.settings.CellSize, c.settings.CellAliveColor, c.settings.CellDeadColor, c.settings.CellBorderColor, c.settings.CellBorders, c.settings.AnimationFormat)
				err = ani.Animate(filename, recorder)
				if err != nil {
					err = fmt.Errorf("animation save failed: %w", err)
				}
			}
		}
		return nil, err
	},
	"heatMap": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		if len(args) > 0 {
			val := args[0].ToString().String()
			hmt := logic.HeatMapperTypeFrom(val)
			c.setInstrumentationHeatMapper(hmt)
		}
		if c.instrumentHeatMap == nil {
			return nil, nil
		}
		return c.instrumentHeatMap.Type().String(), nil
	},
	"heatMapSave": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		if c.instrumentHeatMap != nil {
			var filename string
			hmt := c.instrumentHeatMap.Type()
			var metadata [][2]string
			if len(args) > 0 {
				filename = args[0].ToString().String()
				var metadataArg map[string]any
				for a := 1; a < len(args) && a < 3; a++ {
					val := args[a].Export()
					switch v := val.(type) {
					case map[string]any:
						metadataArg = v
					case string:
						if hmt == logic.AllHeatMapper {
							hmt = logic.HeatMapperTypeFrom(v)
						}
					}
				}
				for k, v := range metadataArg {
					metadata = append(metadata, [2]string{
						k, fmt.Sprintf("%v", v),
					})
				}
				slices.SortFunc(metadata, func(a, b [2]string) int {
					return strings.Compare(a[0], b[0])
				})
			} else {
				filename = c.nowFilename("Heat Map "+hmt.String(), ".png")
			}
			if hmt == logic.AllHeatMapper || hmt == logic.NoHeatMapper {
				hmt = logic.ActivityHeatMapper
			}
			return nil, c.saveNamedHeatMapImage(filename, hmt, metadata)
		}
		return nil, nil
	},
	"heatMapColors": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) >= 2 {
			clrs := make([]color.NRGBA, len(args))
			for i, arg := range args {
				s := arg.ToString().String()
				if clr, ok := parseColor(s); ok {
					clrs[i] = clr
				} else {
					err = fmt.Errorf("invalid color %q", s)
					break
				}
			}
			if err == nil {
				c.settings.HeatMapColors = clrs
			}
		} else if len(args) > 0 {
			c.settings.HeatMapColors = nil
		}
		result := make([]string, len(c.settings.HeatMapColors))
		for i, c := range c.settings.HeatMapColors {
			result[i] = fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
		}
		return result, err
	},
	"heatMapReveal": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		c.stop()
		if c.instrumentHeatMap != nil {
			hm := c.instrumentHeatMap
			if hm.Type() == logic.AllHeatMapper {
				if all, ok := hm.(*logic.AllHeatMapInstrument); ok {
					hmt := logic.ActivityHeatMapper
					if len(args) > 0 {
						hmt = logic.HeatMapperTypeFrom(args[0].ToString().String())
						if hmt == logic.AllHeatMapper || hmt == logic.NoHeatMapper {
							hmt = logic.ActivityHeatMapper
						}
					}
					c.showHeatMapperType = hmt
					hm = all.Specific(hmt)
				}
			}
			c.gridHolder.buildHeatMap(hm)
			c.mode = heatMapMode
		}
		return nil, nil
	},
	"repeatDetect": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		if len(args) > 0 {
			c.setInstrumentationRepeat(args[0].ToBoolean())
		}
		if c.instrumentRepeat == nil {
			return nil, nil
		}
		return struct {
			Step       uint64
			Examined   uint64
			FirstStep  uint64
			RepeatStep uint64
			Period     uint64
			Found      bool
		}{
			Step:       c.instrumentRepeat.Step,
			Examined:   c.instrumentRepeat.Steps,
			FirstStep:  c.instrumentRepeat.FirstStep,
			RepeatStep: c.instrumentRepeat.RepeatStep,
			Period:     c.instrumentRepeat.Period,
			Found:      c.instrumentRepeat.Found,
		}, nil
	},
	"variables": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		return c.settings.ShortcutVariables.Clone(), nil
	},
	"variable": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		switch len(args) {
		case 1:
			name := args[0].ToString().String()
			if v, ok := c.settings.ShortcutVariables.Get(name); ok {
				return v, nil
			}
		case 2:
			name := args[0].ToString().String()
			if validVariableName(name) {
				val := args[1].Export()
				var vs string
				switch v := val.(type) {
				case string:
					vs = v
				default:
					vs = fmt.Sprintf("%v", v)
				}
				c.settings.ShortcutVariables.Set(name, vs)
				return vs, nil
			} else {
				return nil, fmt.Errorf("invalid variable name %q", name)
			}
		}
		return nil, nil
	},
	"clearVariables": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		if len(args) > 0 {
			for _, arg := range args {
				name := arg.ToString().String()
				c.settings.ShortcutVariables.Delete(name)
			}
		} else {
			c.settings.ShortcutVariables.DeleteAll()
		}
		return nil, nil
	},
	"sleep": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		ms := 1 * time.Second
		if len(args) > 0 {
			val := args[0].ToInteger()
			ms = time.Duration(val) * time.Millisecond
		}
		time.Sleep(ms)
		return nil, nil
	},
	"clear": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		c.clear()
		return nil, nil
	},
	"maxAdjacents": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val >= 0 && val <= 8 {
				c.maximumAdjacents(int(val))
			} else {
				err = fmt.Errorf("value %d is out of range (%d-%d)", val, 0, 8)
			}
		}
		return nil, err
	},
	"randomize": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		n := c.settings.Randomization
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val >= 0 && val <= 100 {
				n = int(val)
			} else {
				err = fmt.Errorf("value %d is out of range (%d-%d)", val, 0, 100)
			}
		}
		if err == nil {
			c.randomize(n)
		}
		return nil, err
	},
	"randomizePopulation": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		n := c.settings.Randomization
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val >= 0 && val <= 100 {
				n = int(val)
			} else {
				err = fmt.Errorf("value %d is out of range (%d-%d)", val, 0, 100)
			}
		}
		if err == nil {
			c.randomizePopulation(n)
		}
		return nil, err
	},
	"randomAdditions": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		n := c.settings.Randomization
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val >= 0 && val <= 100 {
				n = int(val)
			} else {
				err = fmt.Errorf("value %d is out of range (%d-%d)", val, 0, 100)
			}
		}
		if err == nil {
			c.randomAdditions(n)
		}
		return nil, err
	},
	"randomChanges": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		n := c.settings.Randomization
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val >= 0 && val <= 100 {
				n = int(val)
			} else {
				err = fmt.Errorf("value %d is out of range (%d-%d)", val, 0, 100)
			}
		}
		if err == nil {
			c.randomChanges(n)
		}
		return nil, err
	},
	"randomCull": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		n := c.settings.Randomization
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val >= 0 && val <= 100 {
				n = int(val)
			} else {
				err = fmt.Errorf("value %d is out of range (%d-%d)", val, 0, 100)
			}
		}
		if err == nil {
			c.randomCull(n)
		}
		return nil, err
	},
	"snapshot": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		c.snapshot()
		return nil, nil
	},
	"undoToSnapshot": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		c.undoToSnapshot()
		return nil, nil
	},
	"replaySnapshot": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		c.replaySnapshot()
		return nil, nil
	},
	"run": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		c.start()
		return nil, nil
	},
	"stop": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		c.stop()
		return nil, nil
	},
	"step": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		return c.step(), nil
	},
	"stepCount": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		return c.gridHolder.grid.StepCount.Load(), nil
	},
	"stepAhead": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		n := c.settings.StepAheadBy
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val > 0 {
				n = int(val)
			} else {
				err = fmt.Errorf("value %d is out of range (min %d)", val, 1)
			}
		}
		if err == nil {
			c.stepAheadBy(n)
		}
		return nil, err
	},
	"stepBack": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		n := c.settings.SkipBackBy
		if len(args) > 0 {
			val := args[0].ToInteger()
			if val > 0 {
				n = int(val)
			} else {
				err = fmt.Errorf("value %d is out of range (min %d)", val, 1)
			}
		}
		if err == nil {
			c.skipBackBy(n)
		}
		return nil, err
	},
	"runRecipe": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			arg := args[0].Export()
			switch v := arg.(type) {
			case string:
				var filename string
				if filename, err = resolvePath(v); err == nil {
					err = c.runRecipeFile(filename)
				}
			case map[string]any:
				var data []byte
				if data, err = json.Marshal(v); err == nil {
					recipe := &recipes.Recipe{}
					if err = json.Unmarshal(data, recipe); err == nil {
						if err = recipe.Validate(""); err == nil {
							err = c.runRecipe(recipe)
						}
					}
				}
			}
		} else if c.gridRecipes != nil {
			c.stop()
			err = c.gridRecipes.runRecipe()
		}
		return err == nil, err
	},
	"status": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		if len(args) > 0 {
			val := args[0].ToString().String()
			c.setShortcutsStatus(val, nil, true)
		} else {
			c.setShortcutsStatus("", nil, true)
		}
		return nil, nil
	},
	"population": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		return c.population(), nil
	},
	"exportImage": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var filename string
		var metadata [][2]string
		if len(args) > 0 {
			filename = args[0].ToString().String()
			if len(args) > 1 {
				val := args[1].Export()
				if metadataArg, ok := val.(map[string]any); ok {
					for k, v := range metadataArg {
						metadata = append(metadata, [2]string{
							k, fmt.Sprintf("%v", v),
						})
					}
					slices.SortFunc(metadata, func(a, b [2]string) int {
						return strings.Compare(a[0], b[0])
					})
				}
			}
		} else {
			filename = c.nowFilename("Grid", ".png")
		}
		return nil, c.exportImageName(filename, metadata)
	},
	"export": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var filename string
		if len(args) > 0 {
			filename = args[0].ToString().String()
		} else {
			filename = c.nowFilename("Grid Export", ".rle")
		}
		return nil, c.exportName(filename)
	},
	"newMetaRule": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		var mr meta.Evaluator
		if len(args) > 0 {
			mrs := args[0].ToString().String()
			if named, ok := c.settings.MetaRules[mrs]; ok {
				mrs = named
			}
			mr, err = meta.ParseRule(mrs)
		}
		return mr, err
	},
	"collectedRules": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		rules := make([]logic.Rule, 0, len(c.settings.CollectedRules))
		for p := range c.settings.CollectedRules {
			if rule, err := logic.NewRuleFromPermutation(p); err == nil {
				rules = append(rules, rule)
			}
		}
		slices.SortFunc(rules, func(a, b logic.Rule) int {
			if a.Permutation() < b.Permutation() {
				return -1
			} else if a.Permutation() > b.Permutation() {
				return 1
			}
			return 0
		})
		return rules, nil
	},
	"addCollectedRule": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].Export()
			switch v := val.(type) {
			case string:
				if rule, ok := logic.Rules[v]; ok {
					c.settings.CollectedRules[rule.Permutation()] = true
				} else if rule, err = logic.NewRuleRle("", v); err == nil {
					c.settings.CollectedRules[rule.Permutation()] = true
				}
			case int64:
				var rule logic.Rule
				if rule, err = logic.NewRuleFromPermutation(int(v)); err == nil {
					c.settings.CollectedRules[rule.Permutation()] = true
				}
			case logic.Rule:
				c.settings.CollectedRules[v.Permutation()] = true
			}
		} else {
			c.settings.CollectedRules[c.gridHolder.grid.Rule().Permutation()] = true
		}
		return nil, err
	},
	"removeCollectedRule": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			val := args[0].Export()
			switch v := val.(type) {
			case string:
				if rule, ok := logic.Rules[v]; ok {
					delete(c.settings.CollectedRules, rule.Permutation())
				} else if rule, err = logic.NewRuleRle("", v); err == nil {
					delete(c.settings.CollectedRules, rule.Permutation())
				}
			case int64:
				var rule logic.Rule
				if rule, err = logic.NewRuleFromPermutation(int(v)); err == nil {
					delete(c.settings.CollectedRules, rule.Permutation())
				}
			case logic.Rule:
				delete(c.settings.CollectedRules, v.Permutation())
			}
		} else {
			delete(c.settings.CollectedRules, c.gridHolder.grid.Rule().Permutation())
		}
		return nil, err
	},
	"ruleCollected": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		result := false
		if len(args) > 0 {
			val := args[0].Export()
			switch v := val.(type) {
			case logic.Rule:
				result = c.settings.CollectedRules[v.Permutation()]
			case string:
				var rule logic.Rule
				if rule, err = logic.NewRuleRle("", v); err == nil {
					result = c.settings.CollectedRules[rule.Permutation()]
				}
			case int64:
				result = c.settings.CollectedRules[int(v)]
			}
		} else {
			result = c.settings.CollectedRules[c.gridHolder.grid.Rule().Permutation()]
		}
		return result, err
	},
	"clearCollectedRules": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		c.settings.CollectedRules = make(map[int]bool)
		return nil, nil
	},
	"patterns": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		items := make([]patterns.Pattern, 0)
		if len(args) > 0 {
			name := strings.ToLower(args[0].ToString().String())
			for _, pattern := range patterns.Library.Filtered([]func(p patterns.Pattern) bool{
				func(p patterns.Pattern) bool {
					return strings.Contains(strings.ToLower(p.Name), name)
				}}) {
				items = append(items, pattern)
			}
		} else {
			for _, pattern := range patterns.Library.All() {
				items = append(items, pattern)
			}
		}
		return items, nil
	},
	"patternFromGrid": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		c.stop()
		var result patterns.Pattern
		var err error
		if len(args) >= 4 {
			col := int(args[0].ToInteger())
			row := int(args[1].ToInteger())
			wd := int(args[2].ToInteger())
			ht := int(args[3].ToInteger())
			result = patterns.NewPatternFromGridPortion(c.gridHolder.grid, row, col, ht, wd)
		} else {
			result, err = patterns.NewPatternFromGrid(c.gridHolder.grid)
		}
		return result, err
	},
	"savePatternImage": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			p := args[0].Export()
			if pattern, ok := p.(patterns.Pattern); ok {
				var filename string
				var metadata [][2]string
				if len(args) > 1 {
					filename = args[1].ToString().String()
					if len(args) > 2 {
						val := args[2].Export()
						if metadataArg, ok := val.(map[string]any); ok {
							for k, v := range metadataArg {
								metadata = append(metadata, [2]string{
									k, fmt.Sprintf("%v", v),
								})
							}
							slices.SortFunc(metadata, func(a, b [2]string) int {
								return strings.Compare(a[0], b[0])
							})
						}
					}
				} else {
					filename = c.nowFilename("Pattern", ".png")
				}
				var f *os.File
				if f, err = saveFile(filename, false); err == nil {
					defer func() {
						_ = f.Close()
					}()
					img := imaging.PatternImagePaletted(pattern, imaging.Config{
						CellSize:    c.settings.CellSize,
						Borders:     c.settings.CellBorders,
						AliveColor:  c.settings.CellAliveColor,
						DeadColor:   c.settings.CellDeadColor,
						BorderColor: c.settings.CellBorderColor,
					})
					err = imaging.PngEncode(f, img, metadata)
				}
			} else {
				err = errors.New("first arg must be a pattern")
			}
		}
		return nil, err
	},
	"addCapturedPattern": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) >= 1 {
			p := args[0].Export()
			if pattern, ok := p.(patterns.Pattern); ok {
				core.statusBar.menuPopup.capturedPatternsPopout.addCapturedPattern(pattern, false)
			} else {
				err = errors.New("arg must be a pattern")
			}
		}
		return nil, err
	},
	"newPattern": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		switch len(args) {
		case 1:
			//rle
			rle := args[0].ToString().String()
			if pattern, err := patterns.PatternRleDecoder(strings.NewReader(rle)); err == nil {
				return pattern, nil
			} else {
				return nil, err
			}
		case 3:
			//width,height,cells
			width := int(args[0].ToInteger())
			if width < 0 {
				return nil, errors.New("width must be greater than zero")
			}
			height := int(args[1].ToInteger())
			if height < 0 {
				return nil, errors.New("height must be greater than zero")
			}
			cellsArg := args[2].Export()
			cells := make([]bool, width*height)
			switch cat := cellsArg.(type) {
			case []uint8:
				if len(cat) != width*height {
					return nil, errors.New("cells data length must be width X height")
				}
				for i := range width * height {
					cells[i] = cat[i] != 0
				}
			case []any:
				if len(cat) != width*height {
					return nil, errors.New("cells data length must be width X height")
				}
				for i := range width * height {
					switch ct := cat[i].(type) {
					case bool:
						cells[i] = ct
					case int64, int:
						cells[i] = ct != 0
					case float64, float32:
						cells[i] = ct != 0
					case string:
						if b, err := strconv.ParseBool(ct); err == nil {
							cells[i] = b
						}
					}
				}
			case string:
				if len(cat) != width*height {
					return nil, errors.New("cells data length must be width X height")
				}
				for i, ch := range []byte(cat) {
					cells[i] = ch == '1' || ch == 'X'
				}
			default:
				return nil, errors.New("invalid cells argument type")
			}
			if pattern, err := patterns.NewPattern("", width, cells); err == nil {
				return pattern, nil
			} else {
				return nil, err
			}
		default:
			return nil, errors.New("expected 1 or 3 arguments")
		}
	},
	"drawPattern": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		if len(args) > 2 {
			x := int(args[0].ToInteger())
			y := int(args[1].ToInteger())
			val := args[2].Export()
			var pattern patterns.Pattern
			switch v := val.(type) {
			case string:
				if p, err := patterns.PatternRleDecoder(strings.NewReader(v)); err == nil {
					pattern = p
				} else {
					return nil, err
				}
			case patterns.Pattern:
				pattern = v
			}
			rot := patterns.Rotate0
			if len(args) > 3 {
				val := args[3].ToInteger()
				rot = patterns.Rotation((val / 90) % 4)
			}
			interlaced := false
			if len(args) > 4 {
				interlaced = args[4].ToBoolean()
			}
			c.stop()
			c.resetInstrumentation()
			c.gridHolder.grid.StepCount.Store(0)
			pattern.Draw(c.gridHolder.grid, y, x, rot, interlaced)
			window.Invalidate()
		}
		return nil, nil
	},
	"cell": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		if len(args) > 1 {
			x := int(args[0].ToInteger())
			y := int(args[1].ToInteger())
			if len(args) > 2 {
				alive := args[2].ToBoolean()
				c.stop()
				c.resetInstrumentation()
				return c.gridHolder.grid.SetCell(y, x, alive), nil
			} else {
				return c.gridHolder.grid.GetCell(y, x), nil
			}
		}
		return nil, nil
	},
	"prompt": func(c *Core, rt *goja.Runtime, args []goja.Value) (any, error) {
		prompt := ""
		replies := make([]string, 0)
		if len(args) > 0 {
			prompt = args[0].ToString().String()
			for i := 1; i < len(args); i++ {
				replies = append(replies, args[i].ToString().String())
			}
		}
		c.shortcutPrompt = newShortcutPrompt(c, prompt, replies...)
		window.Invalidate()
		reply, ok := c.shortcutPrompt.wait()
		if !ok {
			rt.Interrupt(nil)
			return nil, nil
		}
		return reply, nil
	},
	// console...
	"console.log": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		switch {
		case len(args) == 1:
			msg := fmt.Sprintf("%v\n", args[0].Export())
			_, _ = c.shortcutLogger.Write([]byte(msg))
		case len(args) > 1:
			var sb strings.Builder
			for _, arg := range args {
				if sb.Len() > 0 {
					sb.WriteString(" ")
				}
				sb.WriteString(fmt.Sprintf("%v", arg.Export()))
			}
			sb.WriteString("\n")
			_, _ = c.shortcutLogger.Write([]byte(sb.String()))
		}
		return nil, nil
	},
	"console.error": func(c *Core, rt *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		switch {
		case len(args) == 1:
			err = fmt.Errorf("%v", args[0])
		case len(args) > 1:
			var sb strings.Builder
			for _, arg := range args {
				if sb.Len() > 0 {
					sb.WriteString(" ")
				}
				sb.WriteString(fmt.Sprintf("%v", arg))
			}
			err = errors.New(sb.String())
		}
		c.shortcutErrorLogger.Error(err, nil, jsCallInfo(rt, "console.error")...)
		return nil, nil
	},
	"console.warn": func(c *Core, rt *goja.Runtime, args []goja.Value) (any, error) {
		var msg string
		switch {
		case len(args) == 1:
			msg = fmt.Sprintf("%v", args[0].Export())
		case len(args) > 1:
			var sb strings.Builder
			for _, arg := range args {
				if sb.Len() > 0 {
					sb.WriteString(" ")
				}
				sb.WriteString(fmt.Sprintf("%v", arg.Export()))
			}
			msg = sb.String()
		}
		c.shortcutErrorLogger.Warn(msg, nil, jsCallInfo(rt, "console.warn")...)
		return nil, nil
	},
	// files...
	"files.read": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			path := args[0].ToString().String()
			if path, err = resolvePath(path); err == nil {
				var data []byte
				if data, err = os.ReadFile(path); err == nil {
					return string(data), nil
				}
			}
		}
		return nil, err
	},
	"files.write": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			path := args[0].ToString().String()
			var data []byte
			if len(args) > 1 {
				data = []byte(args[1].ToString().String())
			}
			if path, err = resolveSavePath(path); err == nil {
				if err = os.WriteFile(path, data, 0644); err == nil {
					return true, nil
				}
			}
		}
		return nil, err
	},
	"files.append": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 1 {
			path := args[0].ToString().String()
			data := []byte(args[1].ToString().String())
			if path, err = resolveSavePath(path); err == nil {
				var f *os.File
				if f, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
					defer func() {
						_ = f.Close()
					}()
					if _, err = f.Write(data); err == nil {
						return true, nil
					}
				}
			}
		}
		return nil, err
	},
	"files.walk": func(c *Core, rt *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 1 {
			root := args[0].ToString().String()
			if root, err = resolvePath(root); err == nil {
				fnArg := args[1]
				if fn, ok := goja.AssertFunction(fnArg); ok {
					showDirs := true
					if len(args) > 2 {
						showDirs = args[2].ToBoolean()
					}
					// check that the root exists and is a dir (which filepath.WalkDir() does not do)...
					var fi os.FileInfo
					if fi, err = os.Stat(root); err == nil && fi.IsDir() {
						firstDir := true
						err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
							isDir := d.IsDir()
							switch {
							case !c.shortcutRunning:
								rt.Interrupt(nil)
								return filepath.SkipAll
							case isDir && firstDir:
								firstDir = false
								return nil
							case isDir && !showDirs:
								return filepath.SkipDir
							}
							info := map[string]any{
								"path":  filepath.Dir(path),
								"name":  d.Name(),
								"isDir": isDir,
							}
							retVal, cErr := fn(goja.Undefined(), rt.ToValue(info))
							if retVal.ToBoolean() {
								if isDir {
									return filepath.SkipDir
								}
								return filepath.SkipAll
							}
							return cErr
						})
					} else if err == nil {
						err = fmt.Errorf("path %q is not a directory", root)
					}
				} else {
					err = errors.New("expected function argument")
				}
			}
		}
		return nil, err
	},
	"files.info": func(c *Core, _ *goja.Runtime, args []goja.Value) (any, error) {
		var err error
		if len(args) > 0 {
			path := args[0].ToString().String()
			if path, err = resolvePath(path); err == nil {
				var fi os.FileInfo
				if fi, err = os.Stat(path); err == nil {
					return map[string]any{
						"name":    fi.Name(),
						"path":    filepath.Dir(path),
						"size":    fi.Size(),
						"isDir":   fi.IsDir(),
						"modTime": fi.ModTime().Format(time.RFC3339),
					}, nil
				} else if os.IsNotExist(err) {
					err = nil
				}
			}
		}
		return nil, err
	},
}
var jsConstructors = map[string]string{
	"newMetaRule": "MetaRule",
	"newPattern":  "Pattern",
	"newRule":     "Rule",
}

func (c *Core) registerJavascriptFunctions(rt *goja.Runtime, ah shortcuts.ActionHolder) {
	objs := make(map[string]*goja.Object)
	for name, fn := range jsFunctions {
		if objName, mName, ok := strings.Cut(name, "."); ok {
			obj, ok := objs[objName]
			if !ok {
				obj = rt.NewObject()
				objs[objName] = obj
			}
			_ = obj.Set(mName, c.wrapJavascriptFunc(name, fn, rt, ah))
		} else {
			_ = rt.Set(name, c.wrapJavascriptFunc(name, fn, rt, ah))
		}
		if cName, ok := jsConstructors[name]; ok {
			_ = rt.Set(cName, func(call goja.ConstructorCall) *goja.Object {
				if !c.shortcutRunning {
					rt.Interrupt(nil)
					return nil
				}
				v, err := fn(c, rt, call.Arguments)
				if err != nil {
					c.shortcutErrorLogger.Error(err, &ah, jsCallInfo(rt, name)...)
					if rt.Get("exceptions").ToBoolean() {
						panic(rt.ToValue(err.Error()))
					}
				}
				if v != nil {
					return rt.ToValue(v).ToObject(rt)
				}
				return nil
			})
		}
	}
	for k, v := range objs {
		_ = rt.Set(k, v)
	}
}

func (c *Core) wrapJavascriptFunc(name string, fn func(*Core, *goja.Runtime, []goja.Value) (any, error), rt *goja.Runtime, ah shortcuts.ActionHolder) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if !c.shortcutRunning {
			rt.Interrupt(nil)
			return nil
		}
		v, err := fn(c, rt, call.Arguments)
		if err != nil {
			c.shortcutErrorLogger.Error(err, &ah, jsCallInfo(rt, name)...)
			if rt.Get("exceptions").ToBoolean() {
				panic(rt.ToValue(err.Error()))
			}
		}
		if v != nil {
			return rt.ToValue(v)
		}
		return goja.Undefined()
	}
}

func jsCallInfo(rt *goja.Runtime, fn string) []any {
	stack := rt.CaptureCallStack(2, nil)
	if len(stack) > 1 {
		caller := stack[1]
		pos := caller.Position()
		return []any{"function", fn, "line", pos.Line, "column", pos.Column}
	}
	return nil
}

type shortcutPrompt struct {
	core    *Core
	reply   chan string
	prompt  string
	replies []string
	input   widget.Editor
	style   material.EditorStyle
}

func newShortcutPrompt(c *Core, prompt string, replies ...string) *shortcutPrompt {
	if len(prompt) == 0 {
		prompt = "Prompt"
	}
	maxLen := 0
	for _, reply := range replies {
		if len(reply) > maxLen {
			maxLen = len(reply)
		}
	}
	result := &shortcutPrompt{
		core:    c,
		reply:   make(chan string),
		prompt:  strings.TrimSuffix(prompt, ":") + ":",
		replies: replies,
		input: widget.Editor{
			MaxLen:     maxLen,
			SingleLine: true,
			Submit:     true,
		},
	}
	result.style = material.Editor(theme, &result.input, "")
	return result
}

func (p *shortcutPrompt) wait() (string, bool) {
	for {
		select {
		case reply, ok := <-p.reply:
			p.core.shortcutPrompt = nil
			return reply, ok
		}
	}
}

func (p *shortcutPrompt) layout(gtx layout.Context) layout.Dimensions {
	if !gtx.Focused(&p.input) {
		gtx.Execute(key.FocusCmd{Tag: &p.input})
	}
	for {
		ev, ok := gtx.Event(key.Filter{Focus: &p.input, Name: key.NameEscape})
		if !ok {
			break
		}
		if kev, ok := ev.(*key.Event); ok && kev.State == key.Press && kev.Name == key.NameEscape {
			p.core.shortcutRunning = false
			close(p.reply)
		}
	}
	for {
		ev, ok := p.input.Update(gtx)
		if !ok {
			break
		}
		switch ev.(type) {
		case widget.ChangeEvent:
			s := strings.ToLower(p.input.Text())
			for _, reply := range p.replies {
				if s == strings.ToLower(reply) {
					p.reply <- p.input.Text()
				}
			}
		case widget.SubmitEvent:
			if len(p.replies) == 0 {
				p.reply <- p.input.Text()
			}
		}
	}
	pw := gtx.Constraints.Max.X - (measureText(gtx, p.prompt).Size.X - 4)
	return flexHorizontal(8,
		rigid(label(p.prompt)),
		rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.X, gtx.Constraints.Min.X = pw, pw
			return p.style.Layout(gtx)
		}),
	)(gtx)
}
