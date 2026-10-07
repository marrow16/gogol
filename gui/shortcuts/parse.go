package shortcuts

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-andiamo/splitter"
	"strings"
)

func ParseShortcut(name string, lines []string) (Holder, []error) {
	errs := make([]error, 0)
	if len(lines) > 0 && strings.HasPrefix(lines[0], "[") {
		// smells like it might be a JSON representation...
		jr := strings.Join(lines, "")
		var arr []any
		if err := json.Unmarshal([]byte(jr), &arr); err == nil {
			a, jErrs := parseJson(arr)
			errs = append(errs, jErrs...)
			result := Holder{
				Name:    name,
				Actions: a,
			}
			return result, errs
		} else {
			errs = append(errs, fmt.Errorf("invalid JSON format: %w", err))
		}
	}
	a, aErrs := parseActions(lines)
	errs = append(errs, aErrs...)
	result := Holder{
		Name:    name,
		Actions: a,
	}
	return result, errs
}

func parseJson(arr []any) ([]ActionHolder, []error) {
	result := make([]ActionHolder, 0, len(arr))
	errs := make([]error, 0)
	for _, e := range arr {
		switch element := e.(type) {
		case string:
			// simple action (no operand)
			name := element
			isInc, isDec := false, false
			name, isInc = strings.CutSuffix(name, "++")
			name, isDec = strings.CutSuffix(name, "--")
			if isInc && isDec {
				errs = append(errs, errors.New("action has both increment (++) and decrement (--)"))
				continue
			}
			ar, ok := actionsRegistry[name]
			switch {
			case !ok:
				errs = append(errs, fmt.Errorf("unknown action name %q", name))
				continue
			case isInc && !ar.operand.incDecAllowed():
				errs = append(errs, fmt.Errorf("cannot increment action %q", name))
				continue
			case isDec && !ar.operand.incDecAllowed():
				errs = append(errs, fmt.Errorf("cannot decrement action %q", name))
				continue
			case ar.operand.required():
				errs = append(errs, fmt.Errorf("expected operand: action %q", name))
				continue
			}
			if ar.action > NoAction {
				result = append(result, ActionHolder{
					Action: ar.action,
					IsInc:  isInc,
					IsDec:  isDec,
				})
			}
		case map[string]any:
			// an object - it must have one key
			if len(element) == 1 {
				for k, v := range element {
					ar, ok := actionsRegistry[k]
					switch {
					case !ok:
						errs = append(errs, fmt.Errorf("unknown action name %q", k))
						continue
					case ar.operand == none && !ar.subActions:
						errs = append(errs, fmt.Errorf("unexpected operand: action %q", k))
						continue
					}
					switch {
					case ar.subActions:
						// expecting the value to be an object...
						if mv, ok := v.(map[string]any); ok {
							// must have properties "actions" and "for" (optionally)...
							av, ok := mv[ptyActions]
							if !ok {
								errs = append(errs, fmt.Errorf("expected %q property: action %q", ptyActions, k))
								continue
							}
							hasOperand := false
							operand := ""
							forPty := ptyFor
							if ar.action == If || ar.action == While {
								forPty = ptyCondition
							}
							forV, ok := mv[forPty]
							if ok {
								operand = fmt.Sprintf("%v", forV)
								hasOperand = len(operand) > 0
							}
							if !hasOperand && ar.operand.required() {
								errs = append(errs, fmt.Errorf("expected %q property: action %q", forPty, k))
								continue
							}
							var subV []any
							switch forVt := av.(type) {
							case []any:
								subV = forVt
							default:
								subV = []any{fmt.Sprintf("%v", av)}
							}
							subs, subErrs := parseJson(subV)
							errs = append(errs, subErrs...)
							result = append(result, ActionHolder{
								Action:     ar.action,
								HasOperand: hasOperand,
								Operand:    operand,
								Subs:       subs,
							})
						} else {
							errs = append(errs, fmt.Errorf("repeater action %q: expected JSON object", k))
						}
					default:
						// simple action with operand...
						var av string
						switch vt := v.(type) {
						case string:
							av = vt
						case bool, int, int64, float32, float64:
							av = fmt.Sprintf("%v", v)
						}
						if len(av) == 0 {
							errs = append(errs, fmt.Errorf("expected operand: action %q", k))
							continue
						}
						result = append(result, ActionHolder{
							Action:     ar.action,
							HasOperand: true,
							Operand:    av,
						})
					}
				}
			} else {
				errs = append(errs, errors.New("action objects must have one key only"))
			}
		default:
			errs = append(errs, fmt.Errorf("invalid JSON element type: %v", element))
		}
	}
	return result, errs
}

func parseActions(lines []string) ([]ActionHolder, []error) {
	result := make([]ActionHolder, 0, len(lines))
	errs := make([]error, 0)
	for i, s := range lines {
		a, pErrs := parseLine(s, i)
		errs = append(errs, pErrs...)
		switch {
		case a.Action == Unknown && len(pErrs) == 0:
			errs = append(errs, fmt.Errorf("line %d: unknown action %s", i+1, s))
		case a.Action > NoAction:
			result = append(result, a)
		}
	}
	return result, errs
}

var shortcutCommasSplitter = splitter.MustCreateSplitter(',', splitter.DoubleQuotes, splitter.SingleQuotes).
	AddDefaultOptions(splitter.IgnoreEmpties)

func parseLine(line string, index int) (result ActionHolder, errs []error) {
	ar, isInc, isDec, hasOperand, operand, aErr := parseActionItem(line)
	if aErr != nil {
		errs = append(errs, fmt.Errorf("line %d: %w", index+1, aErr))
		return
	}
	if ar.action <= NoAction {
		result.Action = ar.action
		return
	}
	if ar.subActions {
		// repeater (or "if")...
		result, errs = parseSubsLine(ar, operand)
		for i := range errs {
			errs[i] = fmt.Errorf("line %d: %w", index+1, errs[i])
		}
	} else {
		// regular action...
		result = ActionHolder{
			Action:     ar.action,
			HasOperand: hasOperand,
			Operand:    operand,
			IsInc:      isInc,
			IsDec:      isDec,
		}
	}
	return
}

func parseSubsLine(ar registryItem, operand string) (result ActionHolder, errs []error) {
	parts, spErr := shortcutCommasSplitter.Split(operand)
	if spErr != nil {
		errs = append(errs, fmt.Errorf("invalid syntax action %q: %w", ar.action.String(), spErr))
		return
	}
	result = ActionHolder{
		Action: ar.action,
	}
	if ar.operand != none {
		// the first part is operand for repeater - the rest are sub actions...
		result.HasOperand = true
		result.Operand = parts[0]
		parts = parts[1:]
	}
	result.Subs, errs = parseSubActions(parts)
	return
}

func parseSubActions(subs []string) (result []ActionHolder, errs []error) {
	for i, sub := range subs {
		if len(sub) > 0 {
			ar, isInc, isDec, hasOperand, operand, arErr := parseActionItem(sub)
			switch {
			case arErr != nil:
				errs = append(errs, arErr)
				continue
			case ar.action == Unknown:
				errs = append(errs, fmt.Errorf("sub[%d]: unknown action %s", i+1, sub))
				continue
			case ar.action == NoAction:
				continue
			case ar.subActions:
				var tail []string
				if i < len(subs)-1 {
					tail = append(tail, subs[i+1:]...)
				}
				if ar.operand == none && hasOperand {
					tail = append([]string{operand}, tail...)
					hasOperand = false
					operand = ""
				}
				subSubs, subErrs := parseSubActions(tail)
				errs = append(errs, subErrs...)
				result = append(result, ActionHolder{
					Action:     ar.action,
					HasOperand: hasOperand,
					Operand:    operand,
					Subs:       subSubs,
				})
				return
			default:
				result = append(result, ActionHolder{
					Action:     ar.action,
					HasOperand: hasOperand,
					Operand:    operand,
					IsInc:      isInc,
					IsDec:      isDec,
				})
			}
		}
	}
	return
}

func parseActionItem(name string) (ar registryItem, isInc, isDec bool, hasOperand bool, operand string, err error) {
	if name == "" {
		ar = registryItem{action: NoAction}
		return
	}
	var ok bool
	if ar, ok = actionsRegistry[name]; ok {
		if ar.operand.required() || ar.subActions {
			err = fmt.Errorf("expected operand: action %q", name)
			return
		}
		return
	}
	name, operand = cutActionName(name)
	ar, ok = actionsRegistry[name]
	if !ok {
		ar = registryItem{action: Unknown}
		return
	}
	switch {
	case operand == "++":
		operand = ""
		if ar.operand.incDecAllowed() {
			isInc = true
		} else {
			err = fmt.Errorf("cannot increment action %q", name)
		}
		return
	case operand == "--":
		operand = ""
		if ar.operand.incDecAllowed() {
			isDec = true
		} else {
			err = fmt.Errorf("cannot decrement action %q", name)
		}
		return
	}
	if after, ok := strings.CutPrefix(operand, ":"); ok {
		if ar.operand == none && !ar.subActions {
			err = fmt.Errorf("unexpected operand: action %q", name)
			return
		}
		operand = after
		hasOperand = len(operand) > 0
	}
	if !hasOperand && (ar.operand.required() || ar.subActions) {
		err = fmt.Errorf("expected operand: action %q", name)
	}
	return
}

func cutActionName(s string) (name, tail string) {
	rs := []rune(s)
	for i, r := range rs {
		switch r {
		case ':':
			return string(rs[:i]), string(rs[i:])
		case '+', '-':
			if i+1 < len(rs) && rs[i+1] == r {
				return string(rs[:i]), string(rs[i:])
			}
		}
	}
	return s, ""
}
