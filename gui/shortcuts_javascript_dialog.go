package gui

import (
	"errors"
	"fmt"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dop251/goja"
	"github.com/marrow16/gogol/gui/controls"
	"maps"
	"strings"
	"sync"
)

func newShortcutPromptDialog(c *Core, rt *goja.Runtime, prompt string, def []any) (*shortcutPromptDialog, error) {
	if c.javascriptPromptsActive() {
		return nil, errors.New("prompt already active")
	}
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
			result.items = append(result.items, dialogItemHeader(it))
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
							case "focused":
								if s, ok := v.(string); ok {
									result.focused = s
								}
							default:
								result.setValues[k] = v
							}
						}
					}
				default:
					if item, addLabel := buildPromptDialogItem(result, rt, typeStr, lbl, it); item != nil {
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
		return nil, errors.New("no valid control in dialog")
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
	return result, nil
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
	focused          string
	mutex            sync.RWMutex
}

func (p *shortcutPromptDialog) copySetValues() map[string]any {
	p.mutex.RLock()
	result := make(map[string]any, len(p.setValues))
	maps.Copy(result, p.setValues)
	p.mutex.RUnlock()
	return result
}

func (p *shortcutPromptDialog) wait() map[string]any {
	for {
		select {
		case reply, ok := <-p.reply:
			p.core.shortcutPromptDialog = nil
			if ok {
				return reply
			}
			return nil
		}
	}
}

func (p *shortcutPromptDialog) Escape() {
	p.reply <- nil
}

func (p *shortcutPromptDialog) Layout(gtx layout.Context) layout.Dimensions {
	if p.focused != "" {
		for _, item := range p.items {
			if ok := item.setFocused(gtx, p.focused); ok {
				break
			}
		}
		p.focused = ""
	}
	macro := op.Record(gtx.Ops)
	dims := inset(8, 8, 8, 8, p.layoutDialog)(gtx)
	call := macro.Stop()
	paint.FillShape(gtx.Ops, backgroundColor, clip.Rect{Max: dims.Size}.Op())
	borders(gtx, dims, borderColor, true, true, false, true)
	call.Add(gtx.Ops)
	return dims
}

type PromptDialogWrapper struct {
	core *Core
	dlg  *shortcutPromptDialog
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
	switch name {
	case "ok":
		switch vt := value.(type) {
		case nil:
			w.dlg.btnOk.Label("Ok")
		case string:
			w.dlg.btnOk.Label(vt)
		default:
			w.dlg.btnOk.Label(fmt.Sprintf("%v", value))
		}
		set = true
	case "cancel":
		switch vt := value.(type) {
		case nil:
			w.dlg.btnCancel.Label("Cancel")
		case string:
			w.dlg.btnCancel.Label(vt)
		default:
			w.dlg.btnCancel.Label(fmt.Sprintf("%v", value))
		}
		set = true
	default:
		for _, i := range w.dlg.items {
			set = set || i.set(name, value)
		}
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

func (w *PromptDialogWrapper) SetFocused(name string) {
	w.dlg.mutex.Lock()
	w.dlg.focused = name
	w.dlg.mutex.Unlock()
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

func (w *PromptDialogWrapper) OnChange(args ...any) *PromptDialogWrapper {
	if len(args) > 0 {
		switch vt := args[0].(type) {
		case nil:
			w.dlg.onChangeCallback = nil
		case func(goja.FunctionCall) goja.Value:
			w.dlg.onChangeCallback = vt
		}
	}
	return w
}

func (w *PromptDialogWrapper) Activate() any {
	if !w.core.javascriptPromptsActive() {
		if !core.shortcutRunning {
			w.dlg.rt.Interrupt(nil)
			return nil
		}
		w.core.shortcutPromptDialog = w.dlg
		window.Invalidate()
		reply := w.dlg.wait()
		return reply
	} else if w.core.shortcutErrorLogger != nil {
		core.shortcutErrorLogger.Error(errors.New("prompt already active"), nil, jsCallInfo(w.dlg.rt, ".Activate()")...)
	}
	return nil
}

func (p *shortcutPromptDialog) changed(name string, value any) {
	if p.onChangeCallback != nil && name != "" {
		p.jsCall(p.onChangeCallback, map[string]any{"value": value, "name": name})
		window.Invalidate()
	}
}

func (p *shortcutPromptDialog) jsCall(fn func(goja.FunctionCall) goja.Value, response map[string]any) bool {
	defer func() {
		if r := recover(); r != nil {
			if p.core.shortcutErrorLogger != nil {
				p.core.shortcutErrorLogger.Error(fmt.Errorf("call js: %v", r), nil)
			}
		}
	}()
	wrapper := p.rt.ToValue(&PromptDialogWrapper{core: p.core, dlg: p})
	val := fn(goja.FunctionCall{This: goja.Undefined(), Arguments: []goja.Value{p.rt.ToValue(response), wrapper}})
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
	txtDims := measureText(gtx, "\u2003").Size
	for _, i := range p.items {
		item := i
		if item.when(response) && !p.isHidden(item) {
			rows = append(rows, rigid(func(gtx layout.Context) layout.Dimensions {
				return item.layout(gtx, p, response, txtDims, labelWidth)
			}))
		}
	}
	hideOk, hideCancel := p.isButtonsHidden()
	switch {
	case !hideOk && !hideCancel:
		rows = append(rows, rigid(inset(8, 0, 0, 0, flexHorizontal(10, rigid(p.btnOk.Layout), rigid(p.btnCancel.Layout)))))
	case !hideOk:
		rows = append(rows, rigid(inset(8, 0, 0, 0, p.btnOk.Layout)))
	case !hideCancel:
		rows = append(rows, rigid(inset(8, 0, 0, 0, p.btnCancel.Layout)))
	}
	return flexVertical(10, rows...)(gtx)
}

func (p *shortcutPromptDialog) isButtonsHidden() (hideOk, hideCancel bool) {
	p.mutex.RLock()
	hideOk, hideCancel = p.hidden["ok"], p.hidden["cancel"]
	p.mutex.RUnlock()
	return hideOk, hideCancel
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
