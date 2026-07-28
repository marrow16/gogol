package meta

import (
	"fmt"
	"github.com/go-andiamo/splitter"
	"strconv"
	"strings"
)

/*
meta-rule     = predicates

predicates    = predicate ["," predicate]...
predicate     = rule | all-of | any-of | none-of | one-of

all-of        = "AllOf(" predicates ")"
any-of        = "AnyOf(" predicates ")"
none-of       = "NoneOf(" predicates ")"
one-of        = "OneOf(" predicates ")"

rule          = rule-part ["/" rule-part]...
rule-part     = birth
              | survival
              | permutation
birth         = "B(" conditions ")"
survival      = "S(" conditions ")"
permutation   = "P(" ranges ")"

conditions    = condition ["," condition]...
condition     = required
              | forbidden
              | excluded-combination
              | cardinality

required             = "+" digits
forbidden            = "!" digits
excluded-combination = "-" digits
cardinality          = "#" integer ".." integer ":" digits

digits               = digit...
digit                = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8"

ranges               = range ["," range]...
range                = integer | integer "-" integer
*/

/*
Notation Rules
* Tokens are case-insensitive.
* Whitespace may appear freely between tokens.
* Within a rule, B(...), S(...) and P(...) may each appear at most once.
* Digits within a condition are unordered.
* Duplicate digits are not ignored - and will cause errors if repeated.
*/

func MustParseRule(value string) Evaluator {
	if rule, err := ParseRule(value); err == nil {
		return rule
	} else {
		panic(err)
	}
}

func ParseRule(value string) (Evaluator, error) {
	p := ruleParser{
		input: strings.Trim(strings.ToUpper(value), " \t\n\r"),
	}
	rule, err := p.parse()
	if err != nil {
		return Rule{}, err
	}
	return rule, nil
}

type ruleParser struct {
	input  string
	pos    int
	offset int
}

func (p *ruleParser) parse() (result Evaluator, err error) {
	if cmr, ok, err := p.parseComposites(); ok || err != nil {
		return cmr, err
	}
	var birth Conditions
	var survival Conditions
	var andC Conditions
	var orC Conditions
	var xorC Conditions
	var perm Ranges
	seen := map[byte]bool{}
	for !p.atEnd() {
		p.skipSpace()
		b := p.peek()
		if seen[b] {
			return nil, p.errorf("repeat '%s' not allowed", string(b))
		}
		switch b {
		case 'B':
			p.pos++
			seen[b] = true
			if birth, err = p.parseConditions(); err != nil {
				return nil, err
			}
		case 'S':
			p.pos++
			seen[b] = true
			if survival, err = p.parseConditions(); err != nil {
				return nil, err
			}
		case 'A', '&':
			p.pos++
			seen['A'] = true
			if andC, err = p.parseConditions(); err != nil {
				return nil, err
			}
		case 'O', '|':
			p.pos++
			seen['O'] = true
			if orC, err = p.parseConditions(); err != nil {
				return nil, err
			}
		case 'X', '^':
			p.pos++
			seen['X'] = true
			if xorC, err = p.parseConditions(); err != nil {
				return nil, err
			}
		case 'P':
			p.pos++
			seen[b] = true
			if perm, err = p.parsePermutation(); err != nil {
				return nil, err
			}
		case '/':
			if len(seen) == 0 {
				return nil, p.errorf("unexpected character %q", string(b))
			}
			p.pos++
		default:
			return nil, p.errorf("unexpected character %q", string(b))
		}
	}
	return Rule{
		Birth:        birth,
		Survival:     survival,
		And:          andC,
		Or:           orC,
		Xor:          xorC,
		Permutations: perm,
	}, nil
}

var commaSplitter = splitter.MustCreateSplitter(',', splitter.Parenthesis)

func (p *ruleParser) parseComposites() (Evaluator, bool, error) {
	for _, cm := range []CompositeMode{AllOfMode, AnyOfMode, NoneOfMode, OneOfMode} {
		pfx := strings.ToUpper(cm.String()) + "("
		if strings.HasPrefix(p.input, pfx) {
			if !strings.HasSuffix(p.input, ")") {
				p.pos += len(pfx) - 1
				return nil, false, p.errorf("unclosed parenthesis")
			}
			m := p.input[6 : len(p.input)-1]
			if parts, err := commaSplitter.Split(m); err != nil {
				p.pos += len(pfx) - 1
				return nil, false, p.errorf("unclosed parenthesis")
			} else {
				cmr := CompositeRule{Mode: cm}
				offset := len(pfx)
				for _, part := range parts {
					sub := ruleParser{input: strings.Trim(part, " \r\n\t"), offset: p.pos + p.offset + offset}
					offset += len(part) + 2
					var sr Evaluator
					if sr, err = sub.parse(); err == nil {
						cmr.Rules = append(cmr.Rules, sr)
					} else {
						return nil, false, err
					}
				}
				return cmr, true, nil
			}
		}
	}
	return nil, false, nil
}

func (p *ruleParser) parseConditions() (Conditions, error) {
	p.skipSpace()
	if err := p.expect('('); err != nil {
		return nil, err
	}
	p.skipSpace()
	var conditions Conditions
	if p.consume(')') {
		return conditions, nil
	}
	for {
		condition, err := p.parseCondition()
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, condition)
		p.skipSpace()
		switch {
		case p.consume(','):
			p.skipSpace()
			if p.peek() == ')' {
				return nil, p.errorf("expected condition after comma")
			}
		case p.consume(')'):
			return conditions, nil
		default:
			return nil, p.errorf("expected ',' or ')'")
		}
	}
}

func (p *ruleParser) parseCondition() (Condition, error) {
	if p.atEnd() {
		return nil, p.errorf("expected condition")
	}
	switch p.peek() {
	case '+':
		p.pos++
		mask, err := p.parseMask()
		if err != nil {
			return nil, err
		}
		return Require(mask), nil
	case '!':
		p.pos++
		mask, err := p.parseMask()
		if err != nil {
			return nil, err
		}
		return Forbid(mask), nil
	case '-':
		p.pos++
		mask, err := p.parseMask()
		if err != nil {
			return nil, err
		}
		return ExcludeCombination(mask), nil
	case '#':
		p.consume('#')
		return p.parseCardinality()
	default:
		return nil, p.errorf("expected '+', '!', '-' or '#'")
	}
}

func (p *ruleParser) parseCardinality() (Condition, error) {
	p.skipSpace()
	pos := p.pos
	minimum, err := p.parseInteger(9)
	if err != nil {
		return nil, p.errorPosf(pos, "parse cardinality minimum")
	}
	p.skipSpace()
	if err := p.expectString(".."); err != nil {
		return nil, err
	}
	p.skipSpace()
	pos = p.pos
	maximum, err := p.parseInteger(9)
	if err != nil {
		return nil, p.errorPosf(pos, "parse cardinality maximum")
	}
	if minimum > maximum {
		return nil, p.errorf("cardinality minimum %d exceeds maximum %d", minimum, maximum)
	}
	p.skipSpace()
	if err = p.expect(':'); err != nil {
		return nil, err
	}
	p.skipSpace()
	mask, err := p.parseMask()
	if err != nil {
		return nil, err
	}
	bitCount := maskBitCount(mask)
	if int(maximum) > bitCount {
		return nil, p.errorf("cardinality maximum %d exceeds mask size %d", maximum, bitCount)
	}
	return Cardinality{
		Mask: mask,
		Min:  minimum,
		Max:  maximum,
	}, nil
}

func (p *ruleParser) parsePermutation() (Ranges, error) {
	const maxPermutation = uint32(512*512) - 1
	p.skipSpace()
	if err := p.expect('('); err != nil {
		return nil, err
	}
	result := make(Ranges, 0)
	for {
		p.skipSpace()
		pos := p.pos
		minimum, err := p.parseInteger(maxPermutation)
		if err != nil {
			return nil, p.errorPosf(pos, "parse range invalid")
		}
		maximum := minimum
		p.skipSpace()
		if p.consume('-') {
			p.skipSpace()
			pos = p.pos
			maximum, err = p.parseInteger(maxPermutation)
			if err != nil {
				return nil, p.errorPosf(pos, "parse range invalid")
			}
			p.skipSpace()
		}
		minimum, maximum = min(minimum, maximum), max(minimum, maximum)
		result = append(result, Range{minimum, maximum})
		switch {
		case p.consume(','):
			p.skipSpace()
		case p.consume(')'):
			return result, nil
		default:
			return nil, p.errorf("expected ',' or ')'")
		}
	}
}

func (p *ruleParser) parseMask() (uint32, error) {
	start := p.pos
	var mask uint32
	for !p.atEnd() {
		ch := p.peek()
		if ch < '0' || ch > '8' {
			break
		}
		bit := uint32(1) << (ch - '0')
		if mask&bit != 0 {
			return 0, p.errorf("duplicate neighbour count %q", ch)
		}
		mask |= bit
		p.pos++
	}
	if p.pos == start {
		return 0, p.errorf("expected neighbour counts 0 through 8")
	}
	return mask, nil
}

func (p *ruleParser) parseInteger(maximum uint32) (uint32, error) {
	start := p.pos
	s, err := p.collectInt()
	if err != nil {
		return 0, err
	}
	value, err := p.parseUint(s)
	if err != nil || uint32(value) > maximum {
		return 0, p.errorPosf(start, "invalid integer")
	}
	return uint32(value), nil
}

func (p *ruleParser) parseUint(s string) (uint64, error) {
	switch {
	case strings.HasPrefix(s, "0B"):
		return strconv.ParseUint(s[2:], 2, 32)
	case strings.HasPrefix(s, "0O"):
		return strconv.ParseUint(s[2:], 8, 32)
	case strings.HasPrefix(s, "0X"):
		return strconv.ParseUint(s[2:], 16, 32)
	default:
		return strconv.ParseUint(s, 10, 32)
	}
}

func (p *ruleParser) collectInt() (string, error) {
	start := p.pos
	for !p.atEnd() {
		if ch := p.peek(); ch == 'O' || ch == 'X' || (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F') {
			p.pos++
		} else {
			break
		}
	}
	if p.pos == start {
		return "", p.errorf("expected integer")
	}
	return p.input[start:p.pos], nil
}

func (p *ruleParser) expect(expected byte) error {
	if p.atEnd() {
		return p.errorf("expected %q, reached end of input", expected)
	}
	if p.peek() != expected {
		return p.errorf("expected %q, found %q", expected, p.peek())
	}
	p.pos++
	return nil
}

func (p *ruleParser) expectString(expected string) error {
	if !strings.HasPrefix(p.input[p.pos:], expected) {
		return p.errorf("expected %q", expected)
	}
	p.pos += len(expected)
	return nil
}

func (p *ruleParser) consume(ch byte) bool {
	if p.atEnd() || p.peek() != ch {
		return false
	}
	p.pos++
	return true
}

func (p *ruleParser) skipSpace() {
	for !p.atEnd() {
		switch p.peek() {
		case ' ', '\t', '\r', '\n':
			p.pos++
		default:
			return
		}
	}
}

func (p *ruleParser) peek() (result byte) {
	if !p.atEnd() {
		result = p.input[p.pos]
	}
	return
}

func (p *ruleParser) atEnd() bool {
	return p.pos >= len(p.input)
}

func (p *ruleParser) errorf(format string, args ...any) error {
	return fmt.Errorf("parse meta-rule at position %d: %s", p.pos+p.offset, fmt.Sprintf(format, args...))
}

func (p *ruleParser) errorPosf(pos int, format string, args ...any) error {
	return fmt.Errorf("parse meta-rule at position %d: %s", pos+p.offset, fmt.Sprintf(format, args...))
}

func maskBitCount(mask uint32) int {
	count := 0
	for mask != 0 {
		count += int(mask & 1)
		mask >>= 1
	}
	return count
}
