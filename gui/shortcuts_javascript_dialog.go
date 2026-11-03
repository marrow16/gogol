package gui

import (
	"fmt"
	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/dop251/goja"
	"github.com/marrow16/gogol/gui/controls"
	"image"
	"image/color"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync"
)

func newShortcutPromptDialog(c *Core, rt *goja.Runtime, prompt string, def []any) *shortcutPromptDialog {
	if len(prompt) == 0 {
		prompt = "Prompt"
	}
	result := &shortcutPromptDialog{
		core:      c,
		rt:        rt,
		reply:     make(chan map[string]any),
		prompt:    prompt,
		items:     make([]dialogItem, 0, len(def)),
		labels:    make([]string, 0, len(def)),
		setValues: make(map[string]any),
		hidden:    make(map[string]bool),
	}
	okLabel := "Ok"
	cancelLabel := "Cancel"
	ctrls := 0
	var initFunc func(goja.FunctionCall) goja.Value
	for _, i := range def {
		switch it := i.(type) {
		case string:
			result.items = append(result.items, dialogItemLabel(it))
		case map[string]any:
			itm := jsItemDecoder(it)
			if typeStr, ok := itm.isString("type"); ok {
				lbl := ""
				if s, ok := itm.isString("label"); ok {
					lbl = s
				}
				switch strings.ToLower(typeStr) {
				case "ok":
					if lbl != "" {
						okLabel = lbl
					}
					result.onOkClick, _ = itm.jsFunc("onClick")
				case "cancel":
					if lbl != "" {
						cancelLabel = lbl
					}
					result.onCancelClick, _ = itm.jsFunc("onClick")
				case "init", "initialise", "initialize":
					for k, v := range itm {
						if k != "type" {
							switch k {
							case "func":
								initFunc, _ = itm.jsFunc("func")
							case "onChange":
								result.onChangeCallback, _ = itm.jsFunc("onChange")
							case "hide":
								switch vt := v.(type) {
								case string:
									result.hidden[vt] = true
								case []any:
									for _, name := range vt {
										result.hidden[fmt.Sprintf("%v", name)] = true
									}
								}
							default:
								result.setValues[k] = v
							}
						}
					}
				default:
					if item, addLabel := buildPromptDialogRow(result, rt, typeStr, lbl, it); item != nil {
						ctrls++
						result.items = append(result.items, item)
						if addLabel {
							result.labels = append(result.labels, lbl)
						}
					}
				}
			}
		}
	}
	if ctrls == 0 {
		return nil
	}
	result.btnOk = controls.NewButton(theme, okLabel)
	result.btnCancel = controls.NewButton(theme, cancelLabel)
	if initFunc != nil {
		response := make(map[string]any)
		maps.Copy(response, result.setValues)
		for _, item := range result.items {
			item.response(response)
		}
		result.jsCall(initFunc, response)
	}
	return result
}

type shortcutPromptDialog struct {
	core             *Core
	rt               *goja.Runtime
	onChangeCallback func(goja.FunctionCall) goja.Value
	reply            chan map[string]any
	prompt           string
	btnOk            *controls.Button
	onOkClick        func(goja.FunctionCall) goja.Value
	btnCancel        *controls.Button
	onCancelClick    func(goja.FunctionCall) goja.Value
	items            []dialogItem
	labels           []string
	hidden           map[string]bool
	setValues        map[string]any
	mutex            sync.RWMutex
}

func (p *shortcutPromptDialog) copySetValues() map[string]any {
	p.mutex.RLock()
	result := make(map[string]any, len(p.setValues))
	maps.Copy(result, p.setValues)
	p.mutex.RUnlock()
	return result
}

func (p *shortcutPromptDialog) wait() (map[string]any, bool) {
	for {
		select {
		case reply, ok := <-p.reply:
			p.core.shortcutPromptDialog = nil
			return reply, ok
		}
	}
}

func (p *shortcutPromptDialog) Layout(gtx layout.Context) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := inset(8, 8, 8, 8, p.layoutDialog)(gtx)
	call := macro.Stop()
	paint.FillShape(gtx.Ops, backgroundColor, clip.Rect{Max: dims.Size}.Op())
	borders(gtx, dims, borderColor, true, true, false, true)
	call.Add(gtx.Ops)
	return dims
}

type PromptDialogWrapper struct {
	dlg *shortcutPromptDialog
}

func (w *PromptDialogWrapper) Show(names ...string) {
	w.dlg.mutex.Lock()
	for _, name := range names {
		delete(w.dlg.hidden, name)
	}
	w.dlg.mutex.Unlock()
	window.Invalidate()
}

func (w *PromptDialogWrapper) Hide(names ...string) {
	w.dlg.mutex.Lock()
	for _, name := range names {
		w.dlg.hidden[name] = true
	}
	w.dlg.mutex.Unlock()
	window.Invalidate()
}

func (w *PromptDialogWrapper) IsHidden(name string) bool {
	w.dlg.mutex.RLock()
	result := w.dlg.hidden[name]
	w.dlg.mutex.RUnlock()
	return result
}

func (w *PromptDialogWrapper) Get(name string) any {
	var result any
	got := false
	for _, i := range w.dlg.items {
		if v, ok := i.get(name); ok {
			got = true
			result = v
			break
		}
	}
	if !got {
		w.dlg.mutex.RLock()
		result = w.dlg.setValues[name]
		w.dlg.mutex.RUnlock()
	}
	return result
}

func (w *PromptDialogWrapper) Set(name string, value any) {
	set := false
	for _, i := range w.dlg.items {
		set = set || i.set(name, value)
	}
	if !set {
		// not a known control - just set as named value
		w.dlg.setValues[name] = value
	}
	window.Invalidate()
}

func (w *PromptDialogWrapper) Clear(name string) {
	w.dlg.mutex.Lock()
	delete(w.dlg.setValues, name)
	w.dlg.mutex.Unlock()
	for _, i := range w.dlg.items {
		i.set(name, nil)
	}
	window.Invalidate()
}

func (w *PromptDialogWrapper) Ok() {
	response := w.dlg.copySetValues()
	for _, item := range w.dlg.items {
		item.response(response)
	}
	w.dlg.submit(response, "")
}

func (w *PromptDialogWrapper) Cancel() {
	if w.dlg.onCancelClick != nil {
		response := w.dlg.copySetValues()
		for _, item := range w.dlg.items {
			item.response(response)
		}
		w.dlg.jsCall(w.dlg.onCancelClick, response)
	}
	w.dlg.reply <- nil
}

func (p *shortcutPromptDialog) jsCall(fn func(goja.FunctionCall) goja.Value, response map[string]any) bool {
	defer func() {
		if r := recover(); r != nil {
			if p.core.shortcutErrorLogger != nil {
				p.core.shortcutErrorLogger.Error(fmt.Errorf("call js: %v", r), nil)
			}
		}
	}()
	val := fn(goja.FunctionCall{This: goja.Undefined(), Arguments: []goja.Value{p.rt.ToValue(response), p.rt.ToValue(&PromptDialogWrapper{dlg: p})}})
	return val.ToBoolean()
}

func (p *shortcutPromptDialog) submit(response map[string]any, button string) {
	if button != "" {
		response["button"] = button
	}
	p.mutex.RLock()
	maps.Copy(response, p.setValues)
	p.mutex.RUnlock()
	p.reply <- response
}

func (p *shortcutPromptDialog) layoutDialog(gtx layout.Context) layout.Dimensions {
	labelWidth := measureMaxText(gtx, font.Bold, p.labels...)
	response := p.copySetValues()
	for _, item := range p.items {
		item.response(response)
	}
	if p.btnOk.Clicked(gtx) {
		ok := true
		if p.onOkClick != nil {
			ok = !p.jsCall(p.onOkClick, response)
		}
		if ok {
			p.submit(response, "OK")
		}
	}
	if p.btnCancel.Clicked(gtx) {
		if p.onCancelClick != nil {
			p.jsCall(p.onCancelClick, response)
		}
		p.reply <- nil
	}
	rows := make([]layout.FlexChild, 0, len(p.items)+1)
	for _, i := range p.items {
		item := i
		if item.when(response) && !p.isHidden(item) {
			rows = append(rows, rigid(func(gtx layout.Context) layout.Dimensions {
				return item.layout(gtx, p, response, labelWidth)
			}))
		}
	}
	rows = append(rows, rigid(inset(8, 0, 0, 0, flexHorizontal(10, rigid(p.btnOk.Layout), rigid(p.btnCancel.Layout)))))
	return flexVertical(10, rows...)(gtx)
}

func (p *shortcutPromptDialog) isHidden(item dialogItem) bool {
	name, ok := item.getName()
	if !ok {
		return false
	}
	p.mutex.RLock()
	result := p.hidden[name]
	p.mutex.RUnlock()
	return result
}

func buildPromptDialogRow(parent *shortcutPromptDialog, rt *goja.Runtime, t string, lbl string, item jsItemDecoder) (result dialogItem, addLabel bool) {
	name := item.string("name")
	addLabel = true
	switch strings.ToLower(t) {
	case "str", "string":
		maxLen := max(item.integer("maxLen", 10), 0)
		input := controls.NewInput(theme, item.string("filter"), maxLen, func(s string) {
			if parent.onChangeCallback != nil && name != "" {
				parent.jsCall(parent.onChangeCallback, map[string]any{"value": s, "name": name})
			}
			window.Invalidate()
		}).MaximumWidth(max(item.integer("maxWidth", 0), 0))
		input.SetText(item.string("initial"))
		result = &dialogItemString{
			parent: parent,
			whenFn: item.boolFunc("when"),
			name:   name,
			label:  lbl,
			input:  input,
		}
	case "int", "integer":
		maxLen := max(item.integer("maxLen", 10), 0)
		input := controls.NewNumberInput[int](theme, maxLen, item.integer("min", 0), item.integer("max", 0), 0, func(i int) {
			if parent.onChangeCallback != nil && name != "" {
				parent.jsCall(parent.onChangeCallback, map[string]any{"value": i, "name": name})
			}
			window.Invalidate()
		}).WithSpinner(item.bool("spinner", false))
		input.SetValue(item.integer("initial", 0))
		result = &dialogItemInt{
			parent: parent,
			whenFn: item.boolFunc("when"),
			name:   name,
			label:  lbl,
			input:  input,
		}
	case "label":
		addLabel = false
		result = &dialogItemLabelLine{
			name:   name,
			whenFn: item.boolFunc("when"),
			label:  lbl,
			small:  item.bool("small", false),
			bold:   item.bool("bold", false),
			italic: item.bool("italic", false),
			color:  item.color("color"),
		}
	case "radio":
		if options, ok := item.arrayString("options"); ok {
			initial := ""
			if v := item.integer("initial", -1); v >= 0 && v < len(options) {
				initial = options[v]
			}
			r := &dialogItemRadio{
				whenFn:     item.boolFunc("when"),
				parent:     parent,
				name:       name,
				label:      lbl,
				enum:       &widget.Enum{Value: initial},
				horizontal: item.bool("horizontal", false),
			}
			for _, o := range options {
				r.radios = append(r.radios, controls.NewRadioButton(theme, r.enum, o, o))
			}
			result = r
		}
	case "check":
		if options, ok := item.arrayString("options"); ok {
			r := &dialogItemCheck{
				parent:     parent,
				whenFn:     item.boolFunc("when"),
				name:       name,
				label:      lbl,
				horizontal: item.bool("horizontal", false),
			}
			initial := item.integer("initial", 0)
			for i, o := range options {
				idx := 1 << i
				r.checks = append(r.checks, controls.NewCheckBox(theme, o, idx&initial != 0))
			}
			result = r
		}
	case "list":
		if options, ok := item.arrayString("options"); ok {
			result = &dialogItemList{
				parent: parent,
				whenFn: item.boolFunc("when"),
				name:   name,
				label:  lbl,
				list: controls.NewList[string](theme, options, true).RowRenderer(func(gtx layout.Context, index int, item string) layout.Dimensions {
					return inset(0, 0, 4, 4, label(item))(gtx)
				}).OnItemSelect(func(item string, _ bool) {
					if parent.onChangeCallback != nil && name != "" {
						parent.jsCall(parent.onChangeCallback, map[string]any{"value": item, "name": name})
						window.Invalidate()
					}
				}),
				maxLen:   item.integer("maxLen", 0),
				maxWidth: item.integer("maxWidth", 0),
			}
		}
	case "dropdown":
		if options, ok := item.arrayString("options"); ok {
			chooser := controls.NewChooser[string](theme, max(item.integer("maxWidth", 10), 10), options, func(s *string) {
				if parent.onChangeCallback != nil && name != "" {
					parent.jsCall(parent.onChangeCallback, map[string]any{"value": *s, "name": name})
				}
				window.Invalidate()
			}, nil).MaxRows(item.integer("maxLen", 10)).Direction(!item.bool("down", false))
			chooser.SetText(item.string("initial"))
			return &dialogItemDropdown{
				parent:  parent,
				whenFn:  item.boolFunc("when"),
				name:    name,
				label:   lbl,
				chooser: chooser,
			}, true
		}
	case "text":
		r := &dialogItemText{
			whenFn: item.boolFunc("when"),
			name:   name,
			label:  lbl,
			input: controls.NewEditor(theme, func(text string, gtx layout.Context) {
				if parent.onChangeCallback != nil && name != "" {
					parent.jsCall(parent.onChangeCallback, map[string]any{"value": text, "name": name})
				}
				gtx.Execute(op.InvalidateCmd{})
			}),
			maxLines: item.integer("maxLines", 0),
			maxWidth: item.integer("maxWidth", 0),
		}
		if initial := item.string("initial"); initial != "" {
			r.input.SetText(initial)
		}
		result = r
	case "button":
		addLabel = false
		bLabel := item.string("label")
		if bLabel == "" {
			bLabel = "Button"
		}
		r := &dialogItemButton{
			whenFn: item.boolFunc("when"),
			name:   name,
			rt:     rt,
			label:  bLabel,
		}
		r.onClick, _ = item.jsFunc("onClick")
		if item.bool("flat", false) {
			r.flatButton = controls.NewFlatButton(theme, bLabel).Inset(layout.Inset{Top: 2, Bottom: 3, Left: 4, Right: 4})
		} else {
			r.button = controls.NewButton(theme, bLabel)
		}
		result = r
	case "buttons":
		addLabel = false
		if options, ok := item.arrayString("options"); ok {
			r := &dialogItemButtons{
				whenFn: item.boolFunc("when"),
				rt:     rt,
				name:   name,
				label:  lbl,
				gap:    item.integer("gap", 10),
			}
			flat := item.bool("flat", false)
			for _, o := range options {
				if len(o) > 0 {
					if flat {
						r.flatButtons = append(r.flatButtons, controls.NewFlatButton(theme, o).Inset(layout.Inset{Top: 2, Bottom: 3, Left: 4, Right: 4}))
					} else {
						r.buttons = append(r.buttons, controls.NewButton(theme, o))
					}
				}
			}
			r.onClick, _ = item.jsFunc("onClick")
			result = r
		}
	case "file":
		input := controls.NewInput(theme, nil, 0, func(s string) {
			if parent.onChangeCallback != nil && name != "" {
				parent.jsCall(parent.onChangeCallback, map[string]any{"value": s, "name": name})
			}
			window.Invalidate()
		})
		allowExts := make([]string, 0)
		if rv, ok := item["extensions"]; ok {
			if lv, ok := rv.([]any); ok {
				for _, v := range lv {
					allowExts = append(allowExts, fmt.Sprintf("%v", v))
				}
			}
		}
		r := &dialogItemFile{
			parent:    parent,
			whenFn:    item.boolFunc("when"),
			name:      name,
			label:     lbl,
			input:     input,
			allowExts: allowExts,
			allowDirs: item.bool("allowDirs", false),
		}
		r.input.MaximumWidth(max(item.integer("maxWidth", 16), 16)).WithEllipsisButton(r.ellipsisClicked)
		if initial := item.string("initial"); initial != "" {
			r.input.SetText(initial)
		}
		result = r
	case "progress", "progressbar":
		initial := min(max(item.integer("initial", 0), 0), 100)
		initF := float32(initial) / 100
		ht := unit.Dp(max(item.integer("height", 8), 4))
		var radius unit.Dp
		if _, ok := item["radius"]; ok {
			radius = unit.Dp(max(item.integer("radius", 0), 0))
		} else {
			radius = ht / 2
		}
		r := &dialogItemProgress{
			parent: parent,
			whenFn: item.boolFunc("when"),
			name:   name,
			label:  lbl,
			control: controls.ProgressBar{
				Color:      backgroundColor,
				Height:     ht,
				Radius:     radius,
				TrackColor: selectedFocusedColor,
				Progress:   initF,
				Reversed:   item.bool("reversed", false),
			},
			width: max(item.integer("width", 16), 6),
		}
		if clr := item.color("color"); clr != nil {
			r.control.Color = *clr
		}
		if clr := item.color("progressColor"); clr != nil {
			r.control.TrackColor = *clr
		}
		result = r
	case "slider":
		float := &widget.Float{}
		float.Value = min(max(item.float("initial", 0), 0), 1)
		r := &dialogItemSlider{
			parent:  parent,
			whenFn:  item.boolFunc("when"),
			name:    name,
			label:   lbl,
			control: material.Slider(theme, float),
			width:   max(item.integer("width", 16), 6),
			height:  unit.Dp(max(item.integer("height", 0), 20)),
		}
		if !item.bool("horizontal", true) {
			r.control.Axis = layout.Vertical
		}
		result = r
	case "switch":
		r := &dialogItemSwitch{
			parent: parent,
			whenFn: item.boolFunc("when"),
			name:   name,
			label:  lbl,
		}
		r.control = material.Switch(theme, &r.value, "")
		r.value.Value = item.bool("initial", false)
		result = r
	case "group":
		if children, ok := item.arrayObject("items", true); ok {
			items := make([]dialogItem, 0, len(children))
			labels := make([]string, 0, len(children))
			gLabel := lbl
			for _, itm := range children {
				if typeStr, ok := itm.isString("type"); ok {
					sLbl := itm.string("label")
					if sItem, sAddLabel := buildPromptDialogRow(parent, rt, typeStr, sLbl, itm); sItem != nil {
						items = append(items, sItem)
						if sAddLabel {
							labels = append(labels, sLbl)
						}
					}
				}
			}
			if len(items) > 0 {
				result = &dialogItemGroup{
					whenFn:     item.boolFunc("when"),
					name:       name,
					label:      gLabel,
					items:      items,
					labels:     labels,
					horizontal: item.bool("horizontal", false),
					gap:        unit.Dp(item.integer("gap", 10)),
					border:     item.bool("border", false),
					padding:    unit.Dp(item.integer("padding", 0)),
				}
			}
		}
	case "tabbed":
		addLabel = false
		if rawTabs, ok := item.arrayObject("tabs", true); ok && len(rawTabs) > 0 {
			buttons := make([]*controls.FlatButton, 0, len(rawTabs))
			tabs := make([]*dialogItemGroup, 0, len(rawTabs))
			for tabNo, itm := range rawTabs {
				if rawItems, ok := itm.arrayObject("items", true); ok {
					actualItems := make([]dialogItem, 0, len(rawItems))
					labels := make([]string, 0, len(rawItems))
					for _, ai := range rawItems {
						aiLbl := ai.string("label")
						if r, ad := buildPromptDialogRow(parent, rt, ai.string("type"), ai.string("label"), ai); r != nil {
							actualItems = append(actualItems, r)
							if ad {
								labels = append(labels, aiLbl)
							}
						}
					}
					tabLabel := itm.string("label")
					if tabLabel == "" {
						tabLabel = "Tab " + strconv.Itoa(tabNo+1)
					}
					buttons = append(buttons, controls.NewFlatButton(theme, tabLabel))
					tabs = append(tabs, &dialogItemGroup{
						whenFn: itm.boolFunc("when"),
						parent: parent,
						name:   itm.string("name"),
						items:  actualItems,
						labels: labels,
					})
				}
			}
			if len(tabs) > 0 {
				r := &dialogItemTabbed{
					whenFn:   item.boolFunc("when"),
					parent:   parent,
					name:     name,
					selected: item.integer("initial", 0),
					width:    max(item.integer("width", 0), 0),
					height:   max(item.integer("height", 0), 0),
					buttons:  buttons,
					tabs:     tabs,
				}
				result = r
			}
		}
	}
	return result, addLabel
}

type dialogItem interface {
	getName() (string, bool)
	set(name string, value any) bool
	get(name string) (any, bool)
	when(data map[string]any) bool
	response(res map[string]any)
	layout(gtx layout.Context, dlg *shortcutPromptDialog, response map[string]any, labelWidth int) layout.Dimensions
}

type dialogItemLabel string

func (i dialogItemLabel) getName() (string, bool) {
	return "", false
}

func (i dialogItemLabel) set(_ string, _ any) bool {
	// do nothing
	return false
}

func (i dialogItemLabel) get(_ string) (any, bool) {
	return nil, false
}

func (i dialogItemLabel) when(_ map[string]any) bool {
	return true
}

func (i dialogItemLabel) response(_ map[string]any) {
	// nothing to set
}

func (i dialogItemLabel) layout(gtx layout.Context, _ *shortcutPromptDialog, _ map[string]any, _ int) layout.Dimensions {
	lbl := material.Label(theme, theme.TextSize, string(i))
	lbl.MaxLines = 1
	lbl.Font.Weight = font.Bold
	return lbl.Layout(gtx)
}

type dialogItemLabelLine struct {
	dialogItemCommon
	name   string
	label  string
	small  bool
	bold   bool
	italic bool
	color  *color.NRGBA
}

func (i *dialogItemLabelLine) getName() (string, bool) {
	return i.name, true
}

func (i *dialogItemLabelLine) get(name string) (any, bool) {
	if name == i.name {
		return i.label, true
	}
	return nil, false
}

func (i *dialogItemLabelLine) set(name string, value any) bool {
	if name == i.name {
		switch vt := value.(type) {
		case nil:
			i.label = ""
		case string:
			i.label = vt
		default:
			i.label = fmt.Sprintf("%v", value)
		}
		return true
	}
	return false
}

func (i *dialogItemLabelLine) response(_ map[string]any) {
	// nothing to set
}

func (i *dialogItemLabelLine) layout(gtx layout.Context, _ *shortcutPromptDialog, _ map[string]any, _ int) layout.Dimensions {
	lbl := material.Label(theme, theme.TextSize, i.label)
	if i.small {
		lbl.TextSize = (theme.TextSize * 4) / 5
	}
	if i.italic {
		lbl.Font.Style = font.Italic
	}
	if i.bold {
		lbl.Font.Weight = font.Bold
	}
	if i.color != nil {
		lbl.Color = *i.color
	}
	return lbl.Layout(gtx)
}

type dialogItemCommon struct {
	parent *shortcutPromptDialog
	whenFn func(data map[string]any) bool
}

func (i dialogItemCommon) when(data map[string]any) bool {
	if i.whenFn != nil {
		return i.whenFn(data)
	}
	return true
}

type dialogItemList struct {
	dialogItemCommon
	name     string
	label    string
	list     *controls.List[string]
	maxLen   int
	maxWidth int
}

func (i *dialogItemList) getName() (string, bool) {
	return i.name, true
}

func (i *dialogItemList) get(name string) (any, bool) {
	if name == i.name {
		return i.list.Selected(), true
	}
	return nil, false
}

func (i *dialogItemList) set(name string, value any) bool {
	if name == i.name {
		set := false
		if value != nil {
			switch vt := value.(type) {
			case int:
				if vt >= 0 && vt < len(i.list.Items) {
					set = true
					i.list.SetSelected(vt)
					i.list.EnsureVisible(vt)
				}
			case int64:
				if vt >= 0 && int(vt) < len(i.list.Items) {
					set = true
					i.list.SetSelected(int(vt))
					i.list.EnsureVisible(int(vt))
				}
			case string:
				for n, v := range i.list.Items {
					if v == vt {
						set = true
						i.list.SetSelected(n)
						i.list.EnsureVisible(n)
						break
					}
				}
			}
		}
		if !set {
			i.list.SetSelected(0)
			i.list.EnsureVisible(0)
		}
		return true
	}
	return false
}

func (i *dialogItemList) response(response map[string]any) {
	if len(i.name) > 0 {
		response[i.name] = i.list.Selected()
	}
}

func (i *dialogItemList) layout(gtx layout.Context, _ *shortcutPromptDialog, _ map[string]any, labelWidth int) layout.Dimensions {
	ht := 0
	wd := 0
	if i.maxLen > 2 {
		ht = (measureText(gtx, "Xy").Size.Y * i.maxLen) + 2
	} else {
		ht = (measureText(gtx, "Xy").Size.Y * 3) + 2
	}
	if i.maxWidth > 0 {
		wd = (measureText(gtx, "\u2003").Size.X * i.maxWidth) + 10
	} else {
		wd = measureMaxText(gtx, 0, i.list.Items...) + measureText(gtx, "W").Size.X + 8
	}
	if len(i.label) > 0 {
		return flexHorizontal(10,
			rigidLabel(i.label, text.End, font.Bold, labelWidth),
			rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X, gtx.Constraints.Max.X = wd, wd
				gtx.Constraints.Min.Y, gtx.Constraints.Max.Y = ht, ht
				return i.list.Layout(gtx)
			}),
		)(gtx)
	}
	gtx.Constraints.Min.X, gtx.Constraints.Max.X = wd, wd
	gtx.Constraints.Min.Y, gtx.Constraints.Max.Y = ht, ht
	return i.list.Layout(gtx)
}

type dialogItemDropdown struct {
	dialogItemCommon
	name    string
	label   string
	chooser *controls.Chooser[string]
}

func (i *dialogItemDropdown) getName() (string, bool) {
	return i.name, true
}

func (i *dialogItemDropdown) set(name string, value any) bool {
	if name == i.name {
		switch vt := value.(type) {
		case string:
			i.chooser.SetText(vt)
		default:
			i.chooser.SetText(fmt.Sprintf("%v", value))
		}
		return true
	}
	return false
}

func (i *dialogItemDropdown) get(name string) (any, bool) {
	if name == i.name {
		return i.chooser.Text(), true
	}
	return nil, false
}

func (i *dialogItemDropdown) response(response map[string]any) {
	if len(i.name) > 0 {
		response[i.name] = i.chooser.Text()
	}
}

func (i *dialogItemDropdown) layout(gtx layout.Context, dlg *shortcutPromptDialog, response map[string]any, labelWidth int) layout.Dimensions {
	if len(i.label) > 0 {
		return flexHorizontal(10,
			rigidLabel(i.label, text.End, font.Bold, labelWidth),
			rigid(func(gtx layout.Context) layout.Dimensions {
				dims := i.chooser.Layout(gtx)
				i.chooser.LayoutDropdown(gtx)
				return dims
			}),
		)(gtx)
	}
	dims := i.chooser.Layout(gtx)
	i.chooser.LayoutDropdown(gtx)
	return dims
}

type dialogItemRadio struct {
	dialogItemCommon
	name       string
	label      string
	enum       *widget.Enum
	radios     []*controls.RadioButton
	horizontal bool
}

func (i *dialogItemRadio) getName() (string, bool) {
	return i.name, true
}

func (i *dialogItemRadio) get(name string) (any, bool) {
	if name == i.name {
		return i.enum.Value, true
	}
	return nil, false
}

func (i *dialogItemRadio) set(name string, value any) bool {
	if name == i.name {
		switch vt := value.(type) {
		case string:
			i.enum.Value = vt
		default:
			i.enum.Value = ""
		}
		return true
	}
	return false
}

func (i *dialogItemRadio) response(response map[string]any) {
	if len(i.name) > 0 {
		response[i.name] = i.enum.Value
	}
}

func (i *dialogItemRadio) layout(gtx layout.Context, _ *shortcutPromptDialog, _ map[string]any, labelWidth int) layout.Dimensions {
	if i.enum.Update(gtx) {
		if i.parent.onChangeCallback != nil {
			i.parent.jsCall(i.parent.onChangeCallback, map[string]any{"value": i.enum.Value, "name": i.name})
		}
		gtx.Execute(op.InvalidateCmd{})
	}
	items := make([]layout.FlexChild, len(i.radios))
	for n, r := range i.radios {
		items[n] = rigid(r.Layout)
	}
	switch {
	case len(i.label) > 0 && i.horizontal:
		return flexHorizontal(10,
			append([]layout.FlexChild{rigidLabel(i.label, text.End, font.Bold, labelWidth)}, items...)...,
		)(gtx)
	case i.horizontal:
		return flexHorizontal(10, items...)(gtx)
	case len(i.label) > 0 && len(items) > 1:
		return flexVertical(4,
			rigid(flexHorizontal(10,
				rigidLabel(i.label, text.End, font.Bold, labelWidth),
				items[0])),
			rigid(flexHorizontal(10,
				rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: image.Point{X: labelWidth}}
				}),
				rigid(flexVertical(4, items[1:]...)),
			)),
		)(gtx)
	default:
		return flexVertical(4, items...)(gtx)
	}
}

type dialogItemCheck struct {
	dialogItemCommon
	name       string
	label      string
	checks     []*controls.Checkbox
	horizontal bool
}

func (i *dialogItemCheck) getName() (string, bool) {
	return i.name, true
}

func (i *dialogItemCheck) get(name string) (any, bool) {
	if name == i.name {
		if len(i.checks) == 1 {
			return i.checks[0].Checked(), true
		}
		result := make([]string, 0, len(i.checks))
		for _, check := range i.checks {
			if check.Checked() {
				result = append(result, check.Label)
			}
		}
		return result, true
	}
	return nil, false
}

func (i *dialogItemCheck) set(name string, value any) bool {
	if name == i.name {
		isInt := false
		intV := 0
		switch vt := value.(type) {
		case bool:
			if len(i.checks) == 1 {
				i.checks[0].SetChecked(vt)
				return true
			}
		case string:
			for _, check := range i.checks {
				check.SetChecked(check.Label == vt)
			}
		case int:
			isInt = true
			intV = vt
		case int64:
			isInt = true
			intV = int(vt)
		case float32:
			isInt = true
			intV = int(vt)
		case float64:
			isInt = true
			intV = int(vt)
		case []any:
			for _, check := range i.checks {
				check.SetChecked(false)
			}
			for _, x := range vt {
				xs := fmt.Sprintf("%v", x)
				for _, check := range i.checks {
					if check.Label == xs {
						check.SetChecked(true)
					}
				}
			}
		default:
			for _, check := range i.checks {
				check.SetChecked(false)
			}
		}
		if isInt {
			for n, check := range i.checks {
				check.SetChecked(intV&(1<<n) != 0)
			}
		}
		return true
	}
	return false
}

func (i *dialogItemCheck) response(response map[string]any) {
	if len(i.name) > 0 {
		if len(i.checks) == 1 {
			response[i.name] = i.checks[0].Checked()
			return
		}
		result := make([]string, 0, len(i.checks))
		for _, check := range i.checks {
			if check.Checked() {
				result = append(result, check.Label)
			}
		}
		response[i.name] = result
	}
}

func (i *dialogItemCheck) layout(gtx layout.Context, _ *shortcutPromptDialog, _ map[string]any, labelWidth int) layout.Dimensions {
	items := make([]layout.FlexChild, len(i.checks))
	changed := false
	for n, c := range i.checks {
		if c.Update(gtx) {
			changed = true
		}
		items[n] = rigid(c.Layout)
	}
	if changed {
		if i.name != "" && i.parent.onChangeCallback != nil {
			var value any
			if len(i.checks) == 1 {
				value = i.checks[0].Checked()
			} else {
				values := make([]string, 0, len(i.checks))
				for _, check := range i.checks {
					if check.Checked() {
						values = append(values, check.Label)
					}
				}
				value = values
			}
			i.parent.jsCall(i.parent.onChangeCallback, map[string]any{"value": value, "name": i.name})
		}
		gtx.Execute(op.InvalidateCmd{})
	}
	switch {
	case len(i.label) > 0 && i.horizontal:
		return flexHorizontal(10,
			append([]layout.FlexChild{rigidLabel(i.label, text.End, font.Bold, labelWidth)}, items...)...,
		)(gtx)
	case i.horizontal:
		return flexHorizontal(10, items...)(gtx)
	case len(i.label) > 0:
		return flexVertical(4,
			rigid(flexHorizontal(10,
				rigidLabel(i.label, text.End, font.Bold, labelWidth),
				items[0])),
			rigid(flexHorizontal(10,
				rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: image.Point{X: labelWidth}}
				}),
				rigid(flexVertical(4, items[1:]...)),
			)),
		)(gtx)
	default:
		return flexVertical(4, items...)(gtx)
	}
}

type dialogItemString struct {
	dialogItemCommon
	name  string
	label string
	input *controls.Input
}

func (i *dialogItemString) getName() (string, bool) {
	return i.name, true
}

func (i *dialogItemString) get(name string) (any, bool) {
	if name == i.name {
		return i.input.Text(), true
	}
	return nil, false
}

func (i *dialogItemString) set(name string, value any) bool {
	if name == i.name {
		switch vt := value.(type) {
		case nil:
			i.input.SetText("")
		case string:
			i.input.SetText(vt)
		default:
			i.input.SetText(fmt.Sprintf("%v", value))
		}
		return true
	}
	return false
}

func (i *dialogItemString) response(response map[string]any) {
	if len(i.name) > 0 {
		response[i.name] = i.input.Text()
	}
}

func (i *dialogItemString) layout(gtx layout.Context, _ *shortcutPromptDialog, _ map[string]any, labelWidth int) layout.Dimensions {
	if len(i.label) > 0 {
		return flexHorizontal(10,
			rigidLabel(i.label, text.End, font.Bold, labelWidth),
			flexed(i.input.Layout),
		)(gtx)
	} else {
		return flexHorizontal(0, flexed(i.input.Layout))(gtx)
	}
}

type dialogItemInt struct {
	dialogItemCommon
	name  string
	label string
	input *controls.NumberInput[int]
}

func (i *dialogItemInt) getName() (string, bool) {
	return i.name, true
}

func (i *dialogItemInt) get(name string) (any, bool) {
	if name == i.name {
		return i.input.Current(), true
	}
	return nil, false
}

func (i *dialogItemInt) set(name string, value any) bool {
	if name == i.name {
		switch vt := value.(type) {
		case string:
			if n, err := strconv.Atoi(vt); err == nil {
				i.input.SetValue(n)
			} else {
				i.input.SetText("")
			}
		case int:
			i.input.SetValue(vt)
		case int64:
			i.input.SetValue(int(vt))
		case float32:
			i.input.SetValue(int(vt))
		case float64:
			i.input.SetValue(int(vt))
		default:
			i.input.SetText("")
		}
		return true
	}
	return false
}

func (i *dialogItemInt) response(response map[string]any) {
	if len(i.name) > 0 {
		response[i.name] = i.input.Current()
	}
}

func (i *dialogItemInt) layout(gtx layout.Context, _ *shortcutPromptDialog, _ map[string]any, labelWidth int) layout.Dimensions {
	if len(i.label) > 0 {
		return flexHorizontal(10,
			rigidLabel(i.label, text.End, font.Bold, labelWidth),
			flexed(i.input.Layout),
		)(gtx)
	} else {
		return flexHorizontal(0, flexed(i.input.Layout))(gtx)
	}
}

type dialogItemText struct {
	dialogItemCommon
	name     string
	label    string
	input    *controls.Editor
	maxLines int
	maxWidth int
}

func (i *dialogItemText) getName() (string, bool) {
	return i.name, true
}

func (i *dialogItemText) get(name string) (any, bool) {
	if name == i.name {
		return i.input.Text(), true
	}
	return nil, false
}

func (i *dialogItemText) set(name string, value any) bool {
	if name == i.name {
		switch vt := value.(type) {
		case nil:
			i.input.SetText("")
		case string:
			i.input.SetText(vt)
		default:
			i.input.SetText(fmt.Sprintf("%v", value))
		}
		return true
	}
	return false
}

func (i *dialogItemText) response(response map[string]any) {
	if len(i.name) > 0 {
		response[i.name] = i.input.Text()
	}
}

func (i *dialogItemText) layout(gtx layout.Context, _ *shortcutPromptDialog, _ map[string]any, labelWidth int) layout.Dimensions {
	if len(i.label) > 0 {
		return flexHorizontal(10,
			rigidLabel(i.label, text.End, font.Bold, labelWidth),
			flexed(func(gtx layout.Context) layout.Dimensions {
				if i.maxLines > 0 {
					chh := measureText(gtx, "Xy").Size.Y
					gtx.Constraints.Max.Y = (chh * i.maxLines) + 6
					gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
				}
				if i.maxWidth > 0 {
					chw := measureText(gtx, "\u2003").Size.X
					gtx.Constraints.Max.X = (chw * i.maxWidth) + 10
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
				}
				return i.input.Layout(gtx)
			}),
		)(gtx)
	}
	return flexHorizontal(0, flexed(func(gtx layout.Context) layout.Dimensions {
		if i.maxLines > 0 {
			chh := measureText(gtx, "Xy").Size.Y
			gtx.Constraints.Max.Y = (chh * i.maxLines) + 6
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
		}
		if i.maxWidth > 0 {
			chw := measureText(gtx, "\u2003").Size.X
			gtx.Constraints.Max.X = (chw * i.maxWidth) + 10
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
		}
		return i.input.Layout(gtx)
	}))(gtx)
}

type dialogItemButton struct {
	dialogItemCommon
	rt         *goja.Runtime
	name       string
	label      string
	onClick    func(goja.FunctionCall) goja.Value
	button     *controls.Button
	flatButton *controls.FlatButton
}

func (i *dialogItemButton) getName() (string, bool) {
	return i.name, true
}

func (i *dialogItemButton) get(_ string) (any, bool) {
	return nil, i.name == i.name
}

func (i *dialogItemButton) set(name string, _ any) bool {
	return i.name == i.name
}

func (i *dialogItemButton) response(_ map[string]any) {
	// nothing to set
}

func (i *dialogItemButton) layout(gtx layout.Context, dlg *shortcutPromptDialog, response map[string]any, _ int) layout.Dimensions {
	if i.button != nil {
		if i.button.Clicked(gtx) && i.onClick != nil {
			if dlg.jsCall(i.onClick, response) {
				dlg.submit(response, i.label)
			}
		}
		return flexHorizontal(0, rigid(i.button.Layout))(gtx)
	}
	if i.flatButton.Clicked(gtx) && i.onClick != nil {
		if dlg.jsCall(i.onClick, response) {
			dlg.submit(response, i.label)
		}
	}
	return flexHorizontal(0, rigid(i.flatButton.Layout))(gtx)
}

type dialogItemButtons struct {
	dialogItemCommon
	rt          *goja.Runtime
	name        string
	label       string
	onClick     func(goja.FunctionCall) goja.Value
	buttons     []*controls.Button
	flatButtons []*controls.FlatButton
	gap         int
}

func (i *dialogItemButtons) getName() (string, bool) {
	return i.name, true
}

func (i *dialogItemButtons) get(name string) (any, bool) {
	return nil, name == i.name
}

func (i *dialogItemButtons) set(name string, value any) bool {
	return name == i.name
}

func (i *dialogItemButtons) response(_ map[string]any) {
	// nothing to set
}

func (i *dialogItemButtons) layout(gtx layout.Context, dlg *shortcutPromptDialog, response map[string]any, labelWidth int) layout.Dimensions {
	cols := make([]layout.FlexChild, 0)
	for n, b := range i.buttons {
		cols = append(cols, rigid(b.Layout))
		if b.Clicked(gtx) && i.onClick != nil {
			if len(i.name) > 0 {
				response[i.name] = n
			} else {
				response["button"] = n
			}
			if dlg.jsCall(i.onClick, response) {
				if len(i.name) > 0 {
					dlg.submit(response, i.name+":"+strconv.Itoa(n))
				} else {
					dlg.submit(response, strconv.Itoa(n))
				}
			}
		}
	}
	for n, b := range i.flatButtons {
		cols = append(cols, rigid(b.Layout))
		if b.Clicked(gtx) && i.onClick != nil {
			if len(i.name) > 0 {
				response[i.name] = n
			} else {
				response["button"] = n
			}
			if dlg.jsCall(i.onClick, response) {
				if len(i.name) > 0 {
					dlg.submit(response, i.name+":"+strconv.Itoa(n))
				} else {
					dlg.submit(response, strconv.Itoa(n))
				}
			}
		}
	}
	if len(i.label) > 0 {
		return flexHorizontal(10,
			rigidLabel(i.label, text.End, font.Bold, labelWidth),
			rigid(flexHorizontal(i.gap, cols...)),
		)(gtx)
	}
	return flexHorizontal(i.gap, cols...)(gtx)
}

type dialogItemGroup struct {
	dialogItemCommon
	name       string
	label      string
	items      []dialogItem
	labels     []string
	horizontal bool
	gap        unit.Dp
	border     bool
	padding    unit.Dp
}

func (i *dialogItemGroup) getName() (string, bool) {
	return i.name, true
}

func (i *dialogItemGroup) get(name string) (any, bool) {
	for _, item := range i.items {
		if v, ok := item.get(name); ok {
			return v, ok
		}
	}
	return nil, false
}

func (i *dialogItemGroup) set(name string, value any) bool {
	result := false
	for _, item := range i.items {
		result = result || item.set(name, value)
	}
	return result
}

func (i *dialogItemGroup) response(response map[string]any) {
	for _, itm := range i.items {
		itm.response(response)
	}
}

func (i *dialogItemGroup) layout(gtx layout.Context, dlg *shortcutPromptDialog, response map[string]any, labelWidth int) layout.Dimensions {
	for _, item := range i.items {
		item.response(response)
	}
	rows := make([]layout.FlexChild, 0, len(i.items))
	subLabelWidth := measureMaxText(gtx, font.Bold, i.labels...)
	for n := range len(i.items) {
		item := i.items[n]
		if item.when(response) && !dlg.isHidden(item) {
			rows = append(rows, rigid(func(gtx layout.Context) layout.Dimensions {
				return item.layout(gtx, dlg, response, subLabelWidth)
			}))
		}
	}
	bw := unit.Dp(0)
	if i.border {
		bw = 1
	}
	gap := gtx.Dp(i.gap)
	switch {
	case i.horizontal && len(i.label) > 0:
		return widget.Border{Color: borderColor, CornerRadius: 4, Width: bw}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: i.padding, Bottom: i.padding, Left: i.padding, Right: i.padding}.Layout(gtx, flexHorizontal(10,
				rigidLabel(i.label, text.End, font.Bold, labelWidth),
				rigid(flexHorizontal(gap, rows...)),
			))
		})
	case i.horizontal:
		return widget.Border{Color: borderColor, CornerRadius: 4, Width: bw}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: i.padding, Bottom: i.padding, Left: i.padding, Right: i.padding}.Layout(gtx, flexHorizontal(gap, rows...))
		})
	case len(i.label) > 0:
		return widget.Border{Color: borderColor, CornerRadius: 4, Width: bw}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: i.padding, Bottom: i.padding, Left: i.padding, Right: i.padding}.Layout(gtx, flexHorizontal(10,
				rigidLabel(i.label, text.End, font.Bold, labelWidth),
				rigid(flexVertical(gap, rows...)),
			))
		})
	default:
		return widget.Border{Color: borderColor, CornerRadius: 4, Width: bw}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: i.padding, Bottom: i.padding, Left: i.padding, Right: i.padding}.Layout(gtx, flexVertical(gap, rows...))
		})
	}
}

type dialogItemFile struct {
	dialogItemCommon
	name       string
	label      string
	input      *controls.Input
	allowExts  []string
	allowDirs  bool
	setFocused bool
}

func (i *dialogItemFile) getName() (string, bool) {
	return i.name, true
}

func (i *dialogItemFile) get(name string) (any, bool) {
	if i.name == name {
		return i.input.Text(), true
	}
	return nil, false
}

func (i *dialogItemFile) set(name string, value any) bool {
	if name == i.name {
		switch vt := value.(type) {
		case nil:
			i.input.SetText("")
		case string:
			i.input.SetText(vt)
		default:
			i.input.SetText(fmt.Sprintf("%v", value))
		}
		return true
	}
	return false
}

func (i *dialogItemFile) response(res map[string]any) {
	if len(i.name) > 0 {
		res[i.name] = i.input.Text()
	}
}

func (i *dialogItemFile) ellipsisClicked() {
	i.parent.core.showFileFinder("Select File", i.allowExts, i.allowDirs, func(path string) {
		i.input.SetText(path)
		i.setFocused = true
	}, func() {
		i.setFocused = true
	})
}

func (i *dialogItemFile) layout(gtx layout.Context, dlg *shortcutPromptDialog, response map[string]any, labelWidth int) layout.Dimensions {
	if i.setFocused {
		i.input.SetFocused(gtx)
		i.setFocused = false
	}
	if len(i.label) > 0 {
		return flexHorizontal(10,
			rigidLabel(i.label, text.End, font.Bold, labelWidth),
			rigid(i.input.Layout),
		)(gtx)
	} else {
		return flexHorizontal(0,
			rigid(func(gtx layout.Context) layout.Dimensions {
				return i.input.Layout(gtx)
			}),
		)(gtx)
	}
}

type dialogItemProgress struct {
	dialogItemCommon
	name    string
	label   string
	control controls.ProgressBar
	width   int
}

func (i *dialogItemProgress) getName() (string, bool) {
	if i.name != "" {
		return i.name, true
	}
	return "", false
}

func (i *dialogItemProgress) set(name string, value any) bool {
	if i.name == name {
		switch vt := value.(type) {
		case nil:
			i.control.Progress = 0
		case int:
			i.control.Progress = float32(vt) / 100
		case int64:
			i.control.Progress = float32(vt) / 100
		case float64:
			i.control.Progress = float32(vt) / 100
		case float32:
			i.control.Progress = vt / 100
		case string:
			if n, err := strconv.ParseFloat(vt, 32); err == nil {
				i.control.Progress = float32(n) / 100
			}
		}
		return true
	}
	return false
}

func (i *dialogItemProgress) get(name string) (any, bool) {
	if i.name == name {
		return int(i.control.Progress * 100), true
	}
	return nil, false
}

func (i *dialogItemProgress) response(res map[string]any) {
	if len(i.name) > 0 {
		res[i.name] = int(i.control.Progress * 100)
	}
}

func (i *dialogItemProgress) layout(gtx layout.Context, dlg *shortcutPromptDialog, response map[string]any, labelWidth int) layout.Dimensions {
	var wd int
	if i.width > 0 {
		wd = measureText(gtx, "\u2003").Size.X * i.width
	} else {
		wd = measureText(gtx, "\u2003").Size.X * 8
	}
	if len(i.label) > 0 {
		return flexHorizontal(10,
			rigidLabel(i.label, text.End, font.Bold, labelWidth),
			rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X, gtx.Constraints.Max.X = wd, wd
				return i.control.Layout(gtx)
			}),
		)(gtx)
	}
	gtx.Constraints.Min.X, gtx.Constraints.Max.X = wd, wd
	return i.control.Layout(gtx)
}

type dialogItemSlider struct {
	dialogItemCommon
	name    string
	label   string
	control material.SliderStyle
	width   int
	height  unit.Dp
}

func (i *dialogItemSlider) getName() (string, bool) {
	if i.name != "" {
		return i.name, true
	}
	return "", false
}

func (i *dialogItemSlider) set(name string, value any) bool {
	if i.name == name {
		switch vt := value.(type) {
		case nil:
			i.control.Float.Value = 0
		case int:
			i.control.Float.Value = float32(vt)
		case int64:
			i.control.Float.Value = float32(vt)
		case float64:
			i.control.Float.Value = float32(vt)
		case float32:
			i.control.Float.Value = vt
		case string:
			if n, err := strconv.ParseFloat(vt, 32); err == nil {
				i.control.Float.Value = float32(n)
			}
		}
		return true
	}
	return false
}

func (i *dialogItemSlider) get(name string) (any, bool) {
	if i.name == name {
		return i.control.Float.Value, true
	}
	return nil, false
}

func (i *dialogItemSlider) response(res map[string]any) {
	if len(i.name) > 0 {
		res[i.name] = i.control.Float.Value
	}
}

func (i *dialogItemSlider) layout(gtx layout.Context, dlg *shortcutPromptDialog, response map[string]any, labelWidth int) layout.Dimensions {
	var wd int
	var ht int
	if i.control.Axis == layout.Horizontal {
		if i.width > 0 {
			wd = measureText(gtx, "\u2003").Size.X * i.width
		} else {
			wd = measureText(gtx, "\u2003").Size.X * 8
		}
	} else {
		ht = gtx.Dp(i.height)
	}
	if i.control.Float.Update(gtx) && i.name != "" && i.parent.onChangeCallback != nil {
		i.parent.jsCall(i.parent.onChangeCallback, map[string]any{
			"name":  i.name,
			"value": i.control.Float.Value,
		})
	}
	if len(i.label) > 0 {
		return flexHorizontal(10,
			rigidLabel(i.label, text.End, font.Bold, labelWidth),
			rigid(func(gtx layout.Context) layout.Dimensions {
				if i.control.Axis == layout.Horizontal {
					gtx.Constraints.Min.X, gtx.Constraints.Max.X = wd, wd
					ht = measureText(gtx, "Xy").Size.Y
					gtx.Constraints.Min.Y, gtx.Constraints.Max.Y = ht, ht
					return verticalCenter(i.control.Layout)(gtx)
				} else {
					return i.control.Layout(gtx)
				}
			}),
		)(gtx)
	}
	if i.control.Axis == layout.Horizontal {
		gtx.Constraints.Min.X, gtx.Constraints.Max.X = wd, wd
	} else {
		gtx.Constraints.Min.Y, gtx.Constraints.Max.Y = ht, ht
	}
	return i.control.Layout(gtx)
}

type dialogItemSwitch struct {
	dialogItemCommon
	name    string
	label   string
	value   widget.Bool
	control material.SwitchStyle
}

func (i *dialogItemSwitch) getName() (string, bool) {
	if i.name != "" {
		return i.name, true
	}
	return "", false
}

func (i *dialogItemSwitch) set(name string, value any) bool {
	if i.name == name {
		switch vt := value.(type) {
		case nil:
			i.value.Value = false
		case int:
			i.value.Value = vt != 0
		case int64:
			i.value.Value = vt != 0
		case string:
			if b, err := strconv.ParseBool(vt); err == nil {
				i.value.Value = b
			}
		}
		return true
	}
	return false
}

func (i *dialogItemSwitch) get(name string) (any, bool) {
	if i.name == name {
		return i.value.Value, true
	}
	return nil, false
}

func (i *dialogItemSwitch) response(res map[string]any) {
	if len(i.name) > 0 {
		res[i.name] = i.value.Value
	}
}

func (i *dialogItemSwitch) layout(gtx layout.Context, dlg *shortcutPromptDialog, response map[string]any, labelWidth int) layout.Dimensions {
	if i.value.Update(gtx) && i.name != "" && i.parent.onChangeCallback != nil {
		i.parent.jsCall(i.parent.onChangeCallback, map[string]any{
			"name":  i.name,
			"value": i.value.Value,
		})
	}
	if len(i.label) > 0 {
		return flexHorizontal(10,
			rigidLabel(i.label, text.End, font.Bold, labelWidth),
			rigid(i.control.Layout),
		)(gtx)
	}
	return i.control.Layout(gtx)
}

type dialogItemTabbed struct {
	dialogItemCommon
	name     string
	selected int
	width    int
	height   int
	buttons  []*controls.FlatButton
	tabs     []*dialogItemGroup
}

func (i *dialogItemTabbed) getName() (string, bool) {
	return i.name, true
}

func (i *dialogItemTabbed) set(name string, value any) bool {
	result := false
	if name == i.name {
		result = true
		switch vt := value.(type) {
		case nil:
			i.selected = 0
		case int:
			i.selected = vt
		case int64:
			i.selected = int(vt)
		}
	}
	for _, item := range i.tabs {
		result = result || item.set(name, value)
	}
	return result
}

func (i *dialogItemTabbed) get(name string) (any, bool) {
	for _, item := range i.tabs {
		if v, ok := item.get(name); ok {
			return v, ok
		}
	}
	return nil, false
}

func (i *dialogItemTabbed) response(response map[string]any) {
	if i.name != "" {
		response[i.name] = i.selected
	}
	for _, itm := range i.tabs {
		itm.response(response)
	}
}

func (i *dialogItemTabbed) layout(gtx layout.Context, dlg *shortcutPromptDialog, response map[string]any, labelWidth int) layout.Dimensions {
	visibleTabs := make([]int, 0, len(i.tabs))
	showButtons := make([]layout.FlexChild, 0, len(i.buttons))
	// work out which tabs are actually visible...
	for n, t := range i.tabs {
		if t.when(response) && !dlg.isHidden(t) {
			visibleTabs = append(visibleTabs, n)
			showButtons = append(showButtons, rigid(i.buttons[n].Layout))
		}
	}
	if len(visibleTabs) == 0 {
		// no visible tabs - show notghing...
		return layout.Dimensions{}
	}
	// handle tab button clicks...
	for _, tabNo := range visibleTabs {
		if i.buttons[tabNo].Clicked(gtx) {
			changed := tabNo != i.selected
			i.selected = tabNo
			if changed && dlg.onChangeCallback != nil && i.name != "" {
				dlg.jsCall(dlg.onChangeCallback, map[string]any{"name": i.name, "value": i.selected})
			}
			i.buttons[tabNo].SetFocused(gtx)
			gtx.Execute(op.InvalidateCmd{})
		}
	}
	// handle tab button keys...
	if len(visibleTabs) > 1 {
		for n := range len(visibleTabs) {
			btnNo := visibleTabs[n]
			btn := i.buttons[btnNo]
			if btn.IsFocused(gtx) {
				btn.HandleKeys(gtx, func(k key.Name) {
					switch k {
					case key.NameLeftArrow:
						if n > 0 {
							i.selected = visibleTabs[n-1]
						} else {
							i.selected = visibleTabs[len(visibleTabs)-1]
						}
					case key.NameRightArrow:
						if n < len(visibleTabs)-1 {
							i.selected = visibleTabs[n+1]
						} else {
							i.selected = visibleTabs[0]
						}
					}
					if dlg.onChangeCallback != nil && i.name != "" {
						dlg.jsCall(dlg.onChangeCallback, map[string]any{"name": i.name, "value": i.selected})
					}
					i.buttons[i.selected].SetFocused(gtx)
					gtx.Execute(op.InvalidateCmd{})
				}, key.NameLeftArrow, key.NameRightArrow)
			}
		}
	}
	// check that selected tab is actually visible...
	selectedVisible := false
	for _, n := range visibleTabs {
		btn := i.buttons[n]
		btn.Inset(layout.Inset{Left: 6, Right: 6, Top: 3, Bottom: 8})
		btn.Colors(foregroundColor, nil)
		if n == i.selected {
			selectedVisible = true
		}
	}
	if !selectedVisible {
		i.selected = visibleTabs[0]
	}
	i.buttons[i.selected].Colors(foregroundColor, &selectedFocusedColor)
	// build the controls for the selected tab...
	showControls := make([]layout.FlexChild, 0)
	tab := i.tabs[i.selected]
	subLabelWidth := measureMaxText(gtx, font.Bold, tab.labels...)
	for n := range len(tab.items) {
		item := tab.items[n]
		if item.when(response) && !dlg.isHidden(item) {
			showControls = append(showControls, rigid(func(gtx layout.Context) layout.Dimensions {
				return item.layout(gtx, dlg, response, subLabelWidth)
			}))
		}
	}
	// work out width & height...
	wd := 0
	ht := 0
	if i.height > 0 {
		ht = measureText(gtx, "Xy").Size.Y * i.height
	} else {
		maxItems := 0
		for _, t := range i.tabs {
			if l := len(t.items); l > maxItems {
				maxItems = l
			}
		}
		ht = measureText(gtx, "Xy").Size.Y * maxItems
	}
	if i.width > 0 {
		wd = (measureText(gtx, "\u2003").Size.X * i.width) + 10
	}
	return flexVertical(-7,
		rigid(inset(0, 0, 4, 0, flexHorizontal(-1, showButtons...))),
		rigid(func(gtx layout.Context) layout.Dimensions {
			if ht > 0 {
				gtx.Constraints.Min.Y = ht
			}
			if wd > 0 {
				gtx.Constraints.Min.X = wd
			}
			macro := op.Record(gtx.Ops)
			dims := widget.Border{Color: borderColor, CornerRadius: 4, Width: 1}.Layout(gtx, inset(4, 4, 4, 4, func(gtx layout.Context) layout.Dimensions {
				return flexVertical(10, showControls...)(gtx)
			}))
			call := macro.Stop()
			paint.FillShape(gtx.Ops, backgroundColor, clip.UniformRRect(image.Rectangle{Max: dims.Size}, 4).Op(gtx.Ops))
			call.Add(gtx.Ops)
			return dims
		}),
	)(gtx)
}

type jsItemDecoder map[string]any

func (j jsItemDecoder) string(name string) string {
	if rv, ok := j[name]; ok {
		switch v := rv.(type) {
		case string:
			return v
		case int, int64, float32, float64:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

func (j jsItemDecoder) isString(name string) (string, bool) {
	if rv, ok := j[name]; ok {
		if s, ok := rv.(string); ok {
			return s, true
		}
	}
	return "", false
}

func (j jsItemDecoder) integer(name string, def int) int {
	if rv, ok := j[name]; ok {
		switch v := rv.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case float32:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return def
}

func (j jsItemDecoder) float(name string, def float32) float32 {
	if rv, ok := j[name]; ok {
		switch v := rv.(type) {
		case int:
			return float32(v)
		case int64:
			return float32(v)
		case float64:
			return float32(v)
		case float32:
			return v
		case string:
			if i, err := strconv.ParseFloat(v, 32); err == nil {
				return float32(i)
			}
		}
	}
	return def
}

func (j jsItemDecoder) bool(name string, def bool) bool {
	if rv, ok := j[name]; ok {
		switch v := rv.(type) {
		case bool:
			return v
		case int64:
			return v != 0
		case float64:
			return v != 0
		case float32:
			return v != 0
		case string:
			if b, err := strconv.ParseBool(v); err == nil {
				return b
			}
		}
	}
	return def
}

func (j jsItemDecoder) object(name string) (jsItemDecoder, bool) {
	if rv, ok := j[name]; ok {
		if v, ok := rv.(map[string]any); ok {
			return v, true
		}
	}
	return nil, false
}

func (j jsItemDecoder) arrayString(name string) ([]string, bool) {
	if rv, ok := j[name]; ok {
		switch v := rv.(type) {
		case []any:
			result := make([]string, len(v))
			for i, it := range v {
				switch item := it.(type) {
				case nil:
					result[i] = "null"
				case string:
					result[i] = item
				case fmt.Stringer:
					result[i] = item.String()
				default:
					result[i] = fmt.Sprintf("%v", item)
				}
			}
			return result, true
		case []string:
			return v, true
		}
	}
	return nil, false
}

func (j jsItemDecoder) color(name string) *color.NRGBA {
	if s, ok := j.isString(name); ok {
		if c, ok := parseColor(s); ok {
			return &c
		}
	}
	return nil
}

func (j jsItemDecoder) arrayObject(name string, ignoreBadItems bool) ([]jsItemDecoder, bool) {
	if rv, ok := j[name]; ok {
		if v, ok := rv.([]any); ok {
			result := make([]jsItemDecoder, 0, len(v))
			for _, it := range v {
				if item, ok := it.(map[string]any); ok {
					result = append(result, jsItemDecoder(item))
				} else if !ignoreBadItems {
					return nil, false
				}
			}
			return result, true
		}
	}
	return nil, false
}

func (j jsItemDecoder) jsFunc(name string) (func(goja.FunctionCall) goja.Value, bool) {
	if rv, ok := j[name]; ok {
		if jsfunc, ok := rv.(func(goja.FunctionCall) goja.Value); ok {
			return jsfunc, jsfunc != nil
		}
	}
	return nil, false
}

func (j jsItemDecoder) boolFunc(name string) func(data map[string]any) bool {
	if rv, ok := j[name]; ok {
		if vs, ok := rv.(string); ok {
			if pts := strings.SplitN(vs, "=", 2); len(pts) == 2 {
				nm := pts[0]
				not := false
				if before, ok := strings.CutSuffix(nm, "!"); ok {
					not = true
					nm = before
				}
				value := pts[1]
				return func(data map[string]any) bool {
					b := false
					if dv, vok := data[nm]; vok {
						switch dvt := dv.(type) {
						case bool:
							b = dvt
						case string:
							b = value == dvt
						case []string:
							b = slices.Contains(dvt, value)
						default:
							b = value == fmt.Sprintf("%v", dv)
						}
					}
					if not {
						b = !b
					}
					return b
				}
			}
		}
	}
	return nil
}
