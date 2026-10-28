package gui

import (
	"fmt"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
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
			lbl := ""
			if rawL, ok := it["label"]; ok {
				if s, ok := rawL.(string); ok {
					lbl = s
				}
			}
			if rawV, ok := it["type"]; ok {
				if typeStr, ok := rawV.(string); ok {
					switch strings.ToLower(typeStr) {
					case "ok":
						if lbl != "" {
							okLabel = lbl
						}
						if oc, ok := it["onClick"]; ok {
							if ocf, ok := oc.(func(goja.FunctionCall) goja.Value); ok {
								result.onOkClick = ocf
							}
						}
					case "cancel":
						if lbl != "" {
							cancelLabel = lbl
						}
						if oc, ok := it["onClick"]; ok {
							if ocf, ok := oc.(func(goja.FunctionCall) goja.Value); ok {
								result.onCancelClick = ocf
							}
						}
					case "init":
						for k, v := range it {
							if k != "type" {
								switch k {
								case "func":
									if ocf, ok := v.(func(goja.FunctionCall) goja.Value); ok {
										initFunc = ocf
									}
								case "onChange":
									if ocf, ok := v.(func(goja.FunctionCall) goja.Value); ok {
										result.onChangeCallback = ocf
									}
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
	for _, name := range names {
		delete(w.dlg.hidden, name)
	}
	window.Invalidate()
}

func (w *PromptDialogWrapper) Hide(names ...string) {
	for _, name := range names {
		w.dlg.hidden[name] = true
	}
	window.Invalidate()
}

func (w *PromptDialogWrapper) IsHidden(name string) bool {
	return w.dlg.hidden[name]
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
		result = w.dlg.setValues[name]
	}
	return result
}

func (w *PromptDialogWrapper) Set(name string, value any) {
	set := false
	for _, i := range w.dlg.items {
		set = set || i.set(name, value)
	}
	if !set {
		w.dlg.setValues[name] = value
	}
	window.Invalidate()
}

func (w *PromptDialogWrapper) Clear(name string) {
	delete(w.dlg.setValues, name)
	for _, i := range w.dlg.items {
		i.set(name, nil)
	}
	window.Invalidate()
}

func (w *PromptDialogWrapper) Ok() {
	response := make(map[string]any)
	maps.Copy(response, w.dlg.setValues)
	for _, item := range w.dlg.items {
		item.response(response)
	}
	w.dlg.submit(response, "")
}

func (w *PromptDialogWrapper) Cancel() {
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
	maps.Copy(response, p.setValues)
	p.reply <- response
}

func (p *shortcutPromptDialog) layoutDialog(gtx layout.Context) layout.Dimensions {
	labelWidth := measureMaxText(gtx, font.Bold, p.labels...)
	response := make(map[string]any, len(p.items))
	maps.Copy(response, p.setValues)
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
	if name, ok := item.getName(); ok {
		return p.hidden[name]
	}
	return false
}

func buildPromptDialogRow(parent *shortcutPromptDialog, rt *goja.Runtime, t string, lbl string, item map[string]any) (result dialogItem, addLabel bool) {
	itemString := func(name string) string {
		if v, ok := item[name]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}
	itemInt := func(name string, def int) int {
		if v, ok := item[name]; ok {
			switch vt := v.(type) {
			case int:
				return vt
			case int64:
				return int(vt)
			case float64:
				return int(vt)
			case float32:
				return int(vt)
			}
		}
		return def
	}
	itemBool := func(name string, def bool) bool {
		if v, ok := item[name]; ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
		return def
	}
	buildWhenFunc := func() func(data map[string]any) bool {
		if v, ok := item["when"]; ok {
			if vs, ok := v.(string); ok {
				if pts := strings.SplitN(vs, "=", 2); len(pts) == 2 {
					name := pts[0]
					not := false
					if before, ok := strings.CutSuffix(name, "!"); ok {
						not = true
						name = before
					}
					value := pts[1]
					return func(data map[string]any) bool {
						b := false
						if dv, vok := data[name]; vok {
							switch dvt := dv.(type) {
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
	maxL := itemInt("maxLen", 10)
	addLabel = true
	switch strings.ToLower(t) {
	case "str", "string":
		name := itemString("name")
		input := controls.NewInput(theme, itemString("filter"), maxL, func(s string) {
			if parent.onChangeCallback != nil {
				parent.jsCall(parent.onChangeCallback, map[string]any{"value": s, "name": name})
			}
			window.Invalidate()
		}).MaximumWidth(maxL)
		input.SetText(itemString("initial"))
		result = &dialogItemString{
			parent: parent,
			whenFn: buildWhenFunc(),
			name:   name,
			label:  lbl,
			input:  input,
		}
	case "int", "integer":
		name := itemString("name")
		input := controls.NewNumberInput[int](theme, maxL, itemInt("min", 0), itemInt("max", 0), 0, func(i int) {
			if parent.onChangeCallback != nil {
				parent.jsCall(parent.onChangeCallback, map[string]any{"value": i, "name": name})
			}
			window.Invalidate()
		})
		input.SetValue(itemInt("initial", 0))
		result = &dialogItemInt{
			parent: parent,
			whenFn: buildWhenFunc(),
			name:   name,
			label:  lbl,
			input:  input,
		}
	case "label":
		addLabel = false
		var clr *color.NRGBA
		if _, ok := item["color"]; ok {
			if c, cOk := parseColor(itemString("color")); cOk {
				clr = &c
			}
		}
		result = &dialogItemLabelLine{
			name:   itemString("name"),
			whenFn: buildWhenFunc(),
			label:  lbl,
			small:  itemBool("small", false),
			bold:   itemBool("bold", false),
			italic: itemBool("italic", false),
			color:  clr,
		}
	case "radio":
		if rv, ok := item["options"]; ok {
			if lv, ok := rv.([]any); ok && len(lv) > 0 {
				options := make([]string, len(lv))
				for i, v := range lv {
					switch iv := v.(type) {
					case string:
						options[i] = iv
					case fmt.Stringer:
						options[i] = iv.String()
					default:
						options[i] = fmt.Sprintf("%v", v)
					}
				}
				initial := ""
				if v := itemInt("initial", -1); v >= 0 && v < len(options) {
					initial = options[v]
				}
				r := &dialogItemRadio{
					whenFn:     buildWhenFunc(),
					parent:     parent,
					name:       itemString("name"),
					label:      lbl,
					enum:       &widget.Enum{Value: initial},
					horizontal: itemBool("horizontal", false),
				}
				for _, o := range options {
					r.radios = append(r.radios, controls.NewRadioButton(theme, r.enum, o, o))
				}
				result = r
			}
		}
	case "check":
		if rv, ok := item["options"]; ok {
			if lv, ok := rv.([]any); ok && len(lv) > 0 {
				options := make([]string, len(lv))
				for i, v := range lv {
					switch iv := v.(type) {
					case string:
						options[i] = iv
					case fmt.Stringer:
						options[i] = iv.String()
					default:
						options[i] = fmt.Sprintf("%v", v)
					}
				}
				r := &dialogItemCheck{
					parent:     parent,
					whenFn:     buildWhenFunc(),
					name:       itemString("name"),
					label:      lbl,
					horizontal: itemBool("horizontal", false),
				}
				initial := itemInt("initial", 0)
				for i, o := range options {
					idx := 1 << i
					r.checks = append(r.checks, controls.NewCheckBox(theme, o, idx&initial != 0))
				}
				result = r
			}
		}
	case "list":
		if rv, ok := item["options"]; ok {
			name := itemString("name")
			if lv, ok := rv.([]any); ok && len(lv) > 0 {
				options := make([]string, len(lv))
				for i, v := range lv {
					switch iv := v.(type) {
					case string:
						options[i] = iv
					case fmt.Stringer:
						options[i] = iv.String()
					default:
						options[i] = fmt.Sprintf("%v", v)
					}
				}
				result = &dialogItemList{
					parent: parent,
					whenFn: buildWhenFunc(),
					name:   name,
					label:  lbl,
					list: controls.NewList[string](theme, options, true).RowRenderer(func(gtx layout.Context, index int, item string) layout.Dimensions {
						return inset(0, 0, 4, 4, label(item))(gtx)
					}).OnItemSelect(func(item string, _ bool) {
						if parent.onChangeCallback != nil {
							parent.jsCall(parent.onChangeCallback, map[string]any{"value": item, "name": name})
							window.Invalidate()
						}
					}),
					maxLen:   itemInt("maxLen", 0),
					maxWidth: itemInt("maxWidth", 0),
				}
			}
		}
	case "text":
		name := itemString("name")
		r := &dialogItemText{
			whenFn: buildWhenFunc(),
			name:   name,
			label:  lbl,
			input: controls.NewEditor(theme, func(text string, gtx layout.Context) {
				if parent.onChangeCallback != nil {
					parent.jsCall(parent.onChangeCallback, map[string]any{"value": text, "name": name})
				}
				gtx.Execute(op.InvalidateCmd{})
			}),
			maxLines: itemInt("maxLines", 0),
			maxWidth: itemInt("maxWidth", 0),
		}
		if initial := itemString("initial"); initial != "" {
			r.input.SetText(initial)
		}
		result = r
	case "button":
		addLabel = false
		bLabel := itemString("label")
		if bLabel == "" {
			bLabel = "Button"
		}
		r := &dialogItemButton{
			whenFn: buildWhenFunc(),
			name:   itemString("name"),
			rt:     rt,
			label:  bLabel,
		}
		if oc, ok := item["onClick"]; ok {
			if ocf, ok := oc.(func(goja.FunctionCall) goja.Value); ok {
				r.onClick = ocf
			}
		}
		if itemBool("flat", false) {
			r.flatButton = controls.NewFlatButton(theme, bLabel).Inset(layout.Inset{Top: 2, Bottom: 3, Left: 4, Right: 4})
		} else {
			r.button = controls.NewButton(theme, bLabel)
		}
		result = r
	case "buttons":
		addLabel = false
		if rv, ok := item["options"]; ok {
			if lv, ok := rv.([]any); ok && len(lv) > 0 {
				options := make([]string, len(lv))
				for i, v := range lv {
					switch iv := v.(type) {
					case string:
						options[i] = iv
					case fmt.Stringer:
						options[i] = iv.String()
					default:
						options[i] = fmt.Sprintf("%v", v)
					}
				}
				r := &dialogItemButtons{
					whenFn: buildWhenFunc(),
					rt:     rt,
					name:   itemString("name"),
					label:  lbl,
				}
				flat := itemBool("flat", false)
				for _, o := range options {
					if len(o) > 0 {
						if flat {
							r.flatButtons = append(r.flatButtons, controls.NewFlatButton(theme, o).Inset(layout.Inset{Top: 2, Bottom: 3, Left: 4, Right: 4}))
						} else {
							r.buttons = append(r.buttons, controls.NewButton(theme, o))
						}
					}
				}
				if oc, ok := item["onClick"]; ok {
					if ocf, ok := oc.(func(goja.FunctionCall) goja.Value); ok {
						r.onClick = ocf
					}
				}
				result = r
			}
		}
	case "file":
		name := itemString("name")
		input := controls.NewInput(theme, nil, 0, func(s string) {
			if parent.onChangeCallback != nil {
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
			whenFn:    buildWhenFunc(),
			name:      name,
			label:     lbl,
			input:     input,
			path:      controls.NewPathButton(theme),
			allowExts: allowExts,
			allowDirs: itemBool("allowDirs", false),
			maxWidth:  max(itemInt("maxWidth", 16), 16),
		}
		if initial := itemString("initial"); initial != "" {
			r.input.SetText(initial)
		}
		result = r
	case "group":
		if raw, ok := item["items"]; ok {
			if its, ok := raw.([]any); ok {
				items := make([]dialogItem, 0, len(its))
				labels := make([]string, 0, len(its))
				gLabel := lbl
				for _, itm := range its {
					if it, ok := itm.(map[string]any); ok {
						lbl := ""
						if rawL, ok := it["label"]; ok {
							if s, ok := rawL.(string); ok {
								lbl = s
							}
						}
						if rawV, ok := it["type"]; ok {
							if typeStr, ok := rawV.(string); ok {
								if item, addLabel := buildPromptDialogRow(parent, rt, typeStr, lbl, it); item != nil {
									items = append(items, item)
									if addLabel {
										labels = append(labels, lbl)
									}
								}
							}
						}
					}
				}
				if len(items) > 0 {
					result = &dialogItemGroup{
						whenFn:     buildWhenFunc(),
						name:       itemString("name"),
						label:      gLabel,
						items:      items,
						labels:     labels,
						horizontal: itemBool("horizontal", false),
					}
				}
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
	if len(i.label) > 0 {
		return flexHorizontal(10,
			rigidLabel(i.label, text.End, font.Bold, labelWidth),
			rigid(func(gtx layout.Context) layout.Dimensions {
				if i.maxLen > 0 {
					chh := measureText(gtx, "Xy").Size.Y
					gtx.Constraints.Max.Y = (chh * i.maxLen) + 2
					gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
				}
				if i.maxWidth > 0 {
					chw := measureText(gtx, "\u2003").Size.X
					gtx.Constraints.Max.X = (chw * i.maxWidth) + 10
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
				} else {
					maxW := measureMaxText(gtx, 0, i.list.Items...) + measureText(gtx, "W").Size.X + 8
					gtx.Constraints.Max.X, gtx.Constraints.Min.X = maxW, maxW
				}
				return i.list.Layout(gtx)
			}),
		)(gtx)
	}
	if i.maxLen > 0 {
		chh := measureText(gtx, "Xy").Size.Y
		gtx.Constraints.Max.Y = (chh * i.maxLen) + 2
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	}
	if i.maxWidth > 0 {
		chw := measureText(gtx, "\u2003").Size.X
		gtx.Constraints.Max.X = (chw * i.maxWidth) + 10
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	} else {
		maxW := measureMaxText(gtx, 0, i.list.Items...) + measureText(gtx, "W").Size.X + 8
		gtx.Constraints.Max.X, gtx.Constraints.Min.X = maxW, maxW
	}
	return i.list.Layout(gtx)
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
			rigid(flexHorizontal(10, cols...)),
		)(gtx)
	}
	return flexHorizontal(10, cols...)(gtx)
}

type dialogItemGroup struct {
	dialogItemCommon
	name       string
	label      string
	items      []dialogItem
	labels     []string
	horizontal bool
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
	for _, i := range i.items {
		item := i
		if item.when(response) && !dlg.isHidden(item) {
			rows = append(rows, rigid(func(gtx layout.Context) layout.Dimensions {
				return item.layout(gtx, dlg, response, subLabelWidth)
			}))
		}
	}
	if i.horizontal {
		if len(i.label) > 0 {
			return flexHorizontal(10,
				rigidLabel(i.label, text.End, font.Bold, labelWidth),
				rigid(flexHorizontal(10, rows...)),
			)(gtx)
		}
		return flexHorizontal(10, rows...)(gtx)
	}
	if len(i.label) > 0 {
		return flexHorizontal(10,
			rigidLabel(i.label, text.End, font.Bold, labelWidth),
			rigid(flexVertical(10, rows...)),
		)(gtx)
	}
	return flexVertical(10, rows...)(gtx)
}

type dialogItemFile struct {
	dialogItemCommon
	name       string
	label      string
	input      *controls.Input
	path       *controls.PathButton
	allowExts  []string
	allowDirs  bool
	maxWidth   int
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

func (i *dialogItemFile) layout(gtx layout.Context, dlg *shortcutPromptDialog, response map[string]any, labelWidth int) layout.Dimensions {
	if i.path.Clicked(gtx) {
		dlg.core.showFileFinder("Select File", i.allowExts, i.allowDirs, func(path string) {
			i.input.SetText(path)
			i.setFocused = true
		}, func() {
			i.setFocused = true
		})
	}
	if i.setFocused {
		i.input.SetFocused(gtx)
		i.setFocused = false
	}
	wd := measureText(gtx, "\u2003").Size.X * i.maxWidth
	if len(i.label) > 0 {
		return flexHorizontal(10,
			rigidLabel(i.label, text.End, font.Bold, labelWidth),
			rigid(flexHorizontal(0,
				rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X, gtx.Constraints.Max.X = wd, wd
					return i.input.Layout(gtx)
				}),
				rigid(i.path.Layout),
			)),
		)(gtx)
	} else {
		return flexHorizontal(0,
			rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X, gtx.Constraints.Max.X = wd, wd
				return i.input.Layout(gtx)
			}),
			rigid(i.path.Layout),
		)(gtx)
	}
}
