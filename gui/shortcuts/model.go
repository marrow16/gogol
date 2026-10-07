package shortcuts

import (
	"encoding/json"
	"fmt"
)

type Holder struct {
	Name    string
	Actions []ActionHolder
}

func (h Holder) MarshalJSON() ([]byte, error) {
	if len(h.Actions) == 0 {
		return json.Marshal([]ActionHolder{})
	}
	return json.Marshal(h.Actions)
}

type ActionHolder struct {
	Action     Action
	HasOperand bool
	Operand    string
	IsInc      bool
	IsDec      bool
	Subs       []ActionHolder
	ElseSubs   []ActionHolder // "if" only
}

const (
	ptyActions   = "actions"
	ptyFor       = "for"
	ptyCondition = "condition"
)

func (h ActionHolder) MarshalJSON() ([]byte, error) {
	ar, ok := actionsRegistry[h.Action.String()]
	if !ok {
		return nil, fmt.Errorf("unknown action: %s", h.Action.String())
	}
	switch {
	case h.IsInc:
		return json.Marshal(h.Action.String() + "++")
	case h.IsDec:
		return json.Marshal(h.Action.String() + "--")
	case ar.subActions:
		m := map[string]any{
			ptyActions: append([]ActionHolder{}, h.Subs...),
		}
		if ar.operand != none {
			if ar.action == If {
				m[ptyCondition] = h.Operand
			} else {
				m[ptyFor] = h.Operand
			}
		}
		return json.Marshal(map[string]any{h.Action.String(): m})
	case h.HasOperand:
		return json.Marshal(map[string]any{h.Action.String(): h.Operand})
	default:
		return json.Marshal(h.Action.String())
	}
}
