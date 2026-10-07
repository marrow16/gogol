package gui

import (
	"errors"
	"fmt"
	"github.com/marrow16/gogol/gui/shortcuts"
	"strconv"
	"strings"
)

func (c *Core) shortcutCheckVariableName(name string, ah shortcuts.ActionHolder) (ok bool) {
	if len(name) > 0 {
		ok = true
		for i, ch := range name {
			if ok = ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (i > 0 && ch >= '0' && ch <= '9'); !ok {
				break
			}
		}
	}
	if !ok {
		c.shortcutErrorLogger.Error(fmt.Errorf("invalid variable name %q", name), &ah)
	}
	return ok
}

func (c *Core) shortcutVariable(ah shortcuts.ActionHolder, repeats []int) {
	operand := strings.TrimPrefix(ah.Operand, "$")
	if kv := strings.SplitN(operand, "=", 2); len(kv) == 2 {
		switch {
		case strings.HasSuffix(kv[0], "+"):
			vn := strings.TrimRight(strings.TrimSuffix(kv[0], "+"), " ")
			if c.shortcutCheckVariableName(vn, ah) {
				vv := kv[1]
				if v, ok := c.settings.ShortcutVariables.Get(vn); ok {
					done := false
					if nv, err := strconv.Atoi(v); err == nil {
						if opv, err := strconv.Atoi(c.shortcutFormat(vv, repeats)); err == nil {
							c.settings.ShortcutVariables.Set(vn, strconv.Itoa(nv+opv))
							done = true
						}
					}
					if !done {
						// string concat...
						c.settings.ShortcutVariables.Set(vn, v+c.shortcutFormat(vv, repeats))
					}
				} else {
					c.shortcutErrorLogger.Error(fmt.Errorf("variable %q does not exist", vn), &ah)
				}
			}
		case strings.HasSuffix(kv[0], "-"):
			vn := strings.TrimRight(strings.TrimSuffix(kv[0], "-"), " ")
			if c.shortcutCheckVariableName(vn, ah) {
				vv := strings.TrimSpace(kv[1])
				if v, ok := c.settings.ShortcutVariables.Get(vn); ok {
					if nv, err := strconv.Atoi(v); err == nil {
						ops := c.shortcutFormat(vv, repeats)
						if opv, err := strconv.Atoi(ops); err == nil {
							c.settings.ShortcutVariables.Set(vn, strconv.Itoa(nv+opv))
						} else {
							c.shortcutErrorLogger.Error(fmt.Errorf("value %q is not numeric", ops), &ah)
						}
					} else {
						c.shortcutErrorLogger.Error(fmt.Errorf("variable %q is not numeric", vn), &ah)
					}
				} else {
					c.shortcutErrorLogger.Error(fmt.Errorf("variable %q does not exist", vn), &ah)
				}
			}
		case strings.HasSuffix(kv[0], "*"):
			vn := strings.TrimRight(strings.TrimSuffix(kv[0], "*"), " ")
			if c.shortcutCheckVariableName(vn, ah) {
				vv := strings.TrimSpace(kv[1])
				if v, ok := c.settings.ShortcutVariables.Get(vn); ok {
					if nv, err := strconv.Atoi(v); err == nil {
						ops := c.shortcutFormat(vv, repeats)
						if opv, err := strconv.Atoi(ops); err == nil {
							c.settings.ShortcutVariables.Set(vn, strconv.Itoa(nv*opv))
						} else {
							c.shortcutErrorLogger.Error(fmt.Errorf("value %q is not numeric", ops), &ah)
						}
					} else {
						c.shortcutErrorLogger.Error(fmt.Errorf("variable %q is not numeric", vn), &ah)
					}
				} else {
					c.shortcutErrorLogger.Error(fmt.Errorf("variable %q does not exist", vn), &ah)
				}
			}
		case strings.HasSuffix(kv[0], "/"):
			vn := strings.TrimRight(strings.TrimSuffix(kv[0], "/"), " ")
			if c.shortcutCheckVariableName(vn, ah) {
				vv := strings.TrimSpace(kv[1])
				if v, ok := c.settings.ShortcutVariables.Get(vn); ok {
					if nv, err := strconv.Atoi(v); err == nil {
						ops := c.shortcutFormat(vv, repeats)
						if opv, err := strconv.Atoi(ops); err == nil && opv != 0 {
							c.settings.ShortcutVariables.Set(vn, strconv.Itoa(nv/opv))
						} else if err != nil {
							c.shortcutErrorLogger.Error(fmt.Errorf("value %q is not numeric", ops), &ah)
						} else {
							c.shortcutErrorLogger.Error(errors.New("division by zero"), &ah)
						}
					} else {
						c.shortcutErrorLogger.Error(fmt.Errorf("variable %q is not numeric", vn), &ah)
					}
				} else {
					c.shortcutErrorLogger.Error(fmt.Errorf("variable %q does not exist", vn), &ah)
				}
			}
		case strings.HasSuffix(kv[0], "&"):
			vn := strings.TrimRight(strings.TrimSuffix(kv[0], "&"), " ")
			if c.shortcutCheckVariableName(vn, ah) {
				vv := strings.TrimSpace(kv[1])
				if v, ok := c.settings.ShortcutVariables.Get(vn); ok {
					if nv, err := strconv.Atoi(v); err == nil {
						ops := c.shortcutFormat(vv, repeats)
						if opv, err := strconv.Atoi(ops); err == nil {
							c.settings.ShortcutVariables.Set(vn, strconv.Itoa(nv&opv))
						} else {
							c.shortcutErrorLogger.Error(fmt.Errorf("value %q is not numeric", ops), &ah)
						}
					} else {
						c.shortcutErrorLogger.Error(fmt.Errorf("variable %q is not numeric", vn), &ah)
					}
				} else {
					c.shortcutErrorLogger.Error(fmt.Errorf("variable %q does not exist", vn), &ah)
				}
			}
		case strings.HasSuffix(kv[0], "|"):
			vn := strings.TrimRight(strings.TrimSuffix(kv[0], "|"), " ")
			if c.shortcutCheckVariableName(vn, ah) {
				vv := strings.TrimSpace(kv[1])
				if v, ok := c.settings.ShortcutVariables.Get(vn); ok {
					if nv, err := strconv.Atoi(v); err == nil {
						ops := c.shortcutFormat(vv, repeats)
						if opv, err := strconv.Atoi(ops); err == nil {
							c.settings.ShortcutVariables.Set(vn, strconv.Itoa(nv|opv))
						} else {
							c.shortcutErrorLogger.Error(fmt.Errorf("value %q is not numeric", ops), &ah)
						}
					} else {
						c.shortcutErrorLogger.Error(fmt.Errorf("variable %q is not numeric", vn), &ah)
					}
				} else {
					c.shortcutErrorLogger.Error(fmt.Errorf("variable %q does not exist", vn), &ah)
				}
			}
		case strings.HasSuffix(kv[0], "^"):
			vn := strings.TrimRight(strings.TrimSuffix(kv[0], "^"), " ")
			if c.shortcutCheckVariableName(vn, ah) {
				vv := strings.TrimSpace(kv[1])
				if v, ok := c.settings.ShortcutVariables.Get(vn); ok {
					if nv, err := strconv.Atoi(v); err == nil {
						ops := c.shortcutFormat(vv, repeats)
						if opv, err := strconv.Atoi(ops); err == nil {
							c.settings.ShortcutVariables.Set(vn, strconv.Itoa(nv^opv))
						} else {
							c.shortcutErrorLogger.Error(fmt.Errorf("value %q is not numeric", ops), &ah)
						}
					} else {
						c.shortcutErrorLogger.Error(fmt.Errorf("variable %q is not numeric", vn), &ah)
					}
				} else {
					c.shortcutErrorLogger.Error(fmt.Errorf("variable %q does not exist", vn), &ah)
				}
			}
		default:
			vn := strings.TrimRight(kv[0], " ")
			if c.shortcutCheckVariableName(vn, ah) {
				vv := kv[1]
				c.settings.ShortcutVariables.Set(vn, c.shortcutFormat(vv, repeats))
			}
		}
	} else if vn, ok := strings.CutSuffix(operand, "++"); ok {
		if c.shortcutCheckVariableName(vn, ah) {
			if v, ok := c.settings.ShortcutVariables.Get(vn); ok {
				if nv, err := strconv.Atoi(v); err == nil {
					c.settings.ShortcutVariables.Set(vn, strconv.Itoa(nv+1))
				} else {
					c.shortcutErrorLogger.Error(fmt.Errorf("variable %q is not numeric", vn), &ah)
				}
			} else {
				c.shortcutErrorLogger.Error(fmt.Errorf("variable %q does not exist", vn), &ah)
			}
		}
	} else if vn, ok = strings.CutSuffix(operand, "--"); ok {
		if c.shortcutCheckVariableName(vn, ah) {
			if v, ok := c.settings.ShortcutVariables.Get(vn); ok {
				if nv, err := strconv.Atoi(v); err == nil {
					c.settings.ShortcutVariables.Set(vn, strconv.Itoa(nv-1))
				} else {
					c.shortcutErrorLogger.Error(fmt.Errorf("variable %q is not numeric", vn), &ah)
				}
			} else {
				c.shortcutErrorLogger.Error(fmt.Errorf("variable %q does not exist", vn), &ah)
			}
		}
	} else {
		c.shortcutErrorLogger.Error(fmt.Errorf("variable %q unknown operand", vn), &ah)
	}
}

func (c *Core) shortcutResolveVar(s string, ah shortcuts.ActionHolder) string {
	if after, ok := strings.CutPrefix(s, "$"); ok {
		if strings.HasPrefix(after, "$") {
			return after
		}
		if v, ok := c.settings.ShortcutVariables.Get(after); ok {
			return v
		} else {
			c.shortcutErrorLogger.Warn(fmt.Sprintf("unknown variable %q", after), &ah)
		}
	}
	return s
}

func (c *Core) shortcutVarString(ah shortcuts.ActionHolder) string {
	return c.shortcutResolveVar(ah.Operand, ah)
}

func (c *Core) shortcutVarInt(ah shortcuts.ActionHolder) (int, error) {
	i, err := strconv.Atoi(c.shortcutResolveVar(ah.Operand, ah))
	if err != nil {
		c.shortcutErrorLogger.Error(err, &ah)
	}
	return i, err
}

func (c *Core) shortcutVarIntRange(ah shortcuts.ActionHolder, minimum, maximum int) (int, error) {
	i, err := strconv.Atoi(c.shortcutResolveVar(ah.Operand, ah))
	if err == nil && (i < minimum || i > maximum) {
		err = fmt.Errorf("value %d is out of range (%d-%d)", i, minimum, maximum)
	}
	if err != nil {
		c.shortcutErrorLogger.Error(err, &ah)
	}
	return i, err
}

func (c *Core) shortcutVarIntMin(ah shortcuts.ActionHolder, minimumExc int) (int, error) {
	i, err := strconv.Atoi(c.shortcutResolveVar(ah.Operand, ah))
	if err == nil && i <= minimumExc {
		err = fmt.Errorf("value %d is out of range (must be greater than %d)", i, minimumExc)
	}
	if err != nil {
		c.shortcutErrorLogger.Error(err, &ah)
	}
	return i, err
}

func (c *Core) shortcutVarBool(ah shortcuts.ActionHolder) (bool, error) {
	b, err := strconv.ParseBool(c.shortcutResolveVar(ah.Operand, ah))
	if err != nil {
		c.shortcutErrorLogger.Error(err, &ah)
	}
	return b, err
}
