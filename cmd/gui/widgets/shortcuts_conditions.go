package widgets

import (
	"strconv"
	"strings"
)

func (c *Core) shortcutConditionEvaluate(condition string, repeats []int) bool {
	switch {
	case condition == "rule-collected":
		return c.settings.CollectedRules[c.gridHolder.grid.Rule().Permutation()]
	case condition == "!rule-collected":
		return !c.settings.CollectedRules[c.gridHolder.grid.Rule().Permutation()]
	case strings.HasPrefix(condition, "variable-exists:"):
		after, _ := strings.CutPrefix(condition, "variable-exists:")
		_, exists := c.settings.ShortcutVariables.Get(strings.TrimPrefix(after, "$"))
		return exists
	case strings.HasPrefix(condition, "!variable-exists:"):
		after, _ := strings.CutPrefix(condition, "!variable-exists:")
		_, exists := c.settings.ShortcutVariables.Get(strings.TrimPrefix(after, "$"))
		return !exists
	case strings.HasPrefix(condition, "variable-exists-n:"):
		after, _ := strings.CutPrefix(condition, "variable-exists-n:")
		v, exists := c.settings.ShortcutVariables.Get(strings.TrimPrefix(after, "$"))
		if exists {
			if _, err := strconv.Atoi(v); err != nil {
				exists = false
			}
		}
		return exists
	case strings.HasPrefix(condition, "!variable-exists-n:"):
		after, _ := strings.CutPrefix(condition, "!variable-exists-n:")
		v, exists := c.settings.ShortcutVariables.Get(strings.TrimPrefix(after, "$"))
		if exists {
			if _, err := strconv.Atoi(v); err != nil {
				exists = false
			}
		}
		return !exists
	case strings.Contains(condition, "=="):
		return c.shortcutCompare(equals, strings.SplitN(condition, "==", 2), repeats)
	case strings.Contains(condition, "!="):
		return c.shortcutCompare(notEquals, strings.SplitN(condition, "!=", 2), repeats)
	case strings.Contains(condition, ">="):
		return c.shortcutCompare(greaterThanOrEqual, strings.SplitN(condition, ">=", 2), repeats)
	case strings.Contains(condition, "<="):
		return c.shortcutCompare(lessThanOrEqual, strings.SplitN(condition, "<=", 2), repeats)
	case strings.Contains(condition, ">"):
		return c.shortcutCompare(greaterThan, strings.SplitN(condition, ">", 2), repeats)
	case strings.Contains(condition, "<"):
		return c.shortcutCompare(lessThan, strings.SplitN(condition, "<", 2), repeats)
	}
	return false
}

type comparator int

const (
	equals comparator = iota
	notEquals
	greaterThan
	lessThan
	greaterThanOrEqual
	lessThanOrEqual
)

func (c *Core) shortcutCompare(comp comparator, parts []string, repeats []int) bool {
	if len(parts) != 2 {
		return false
	}
	vl := c.shortcutFormat(strings.TrimSpace(parts[0]), repeats)
	vr := c.shortcutFormat(strings.TrimSpace(parts[1]), repeats)
	if nl, err := strconv.Atoi(vl); err == nil {
		if nr, err := strconv.Atoi(vr); err == nil {
			switch comp {
			case equals:
				return nl == nr
			case notEquals:
				return nl != nr
			case greaterThan:
				return nl > nr
			case lessThan:
				return nl < nr
			case greaterThanOrEqual:
				return nl >= nr
			case lessThanOrEqual:
				return nl <= nr
			}
		}
	}
	sc := strings.Compare(strings.ToLower(vl), strings.ToLower(vr))
	switch comp {
	case equals:
		return sc == 0
	case notEquals:
		return sc != 0
	case greaterThan:
		return sc > 0
	case lessThan:
		return sc < 0
	case greaterThanOrEqual:
		return sc >= 0
	case lessThanOrEqual:
		return sc <= 0
	default:
		return false
	}
}
