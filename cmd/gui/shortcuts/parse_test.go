package shortcuts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestParseShortcut(t *testing.T) {
	testCases := []struct {
		raw           string
		expectActions []ActionHolder
		expectErrors  int
	}{
		{},
		{
			raw: `run`,
			expectActions: []ActionHolder{
				{Action: Run},
			},
		},
		{
			raw: `repeat:10,step`,
			expectActions: []ActionHolder{
				{
					Action:     Repeat,
					HasOperand: true,
					Operand:    "10",
					Subs: []ActionHolder{
						{Action: Step},
					},
				},
			},
		},
		{
			raw: `repeat:10,repeat:20,step`,
			expectActions: []ActionHolder{
				{
					Action:     Repeat,
					HasOperand: true,
					Operand:    "10",
					Subs: []ActionHolder{
						{
							Action:     Repeat,
							HasOperand: true,
							Operand:    "20",
							Subs: []ActionHolder{
								{Action: Step},
							},
						},
					},
				},
			},
		},
		{
			raw: `repeat:10,repeat:20,iterate-collected-rules:step`,
			expectActions: []ActionHolder{
				{
					Action:     Repeat,
					HasOperand: true,
					Operand:    "10",
					Subs: []ActionHolder{
						{
							Action:     Repeat,
							HasOperand: true,
							Operand:    "20",
							Subs: []ActionHolder{
								{
									Action: IterateCollectedRules,
									Subs: []ActionHolder{
										{Action: Step},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			raw:          `unknown`,
			expectErrors: 1,
		},
		{
			raw:          `repeat:`,
			expectErrors: 1,
		},
		{
			raw: `iterate-collected-rules:step`,
			expectActions: []ActionHolder{
				{
					Action: IterateCollectedRules,
					Subs: []ActionHolder{
						{Action: Step},
					},
				},
			},
		},
		{
			raw:          `repeat:10,unbalanced:"`,
			expectErrors: 1,
		},
		{
			raw: `repeat:10,unknown`,
			expectActions: []ActionHolder{
				{
					Action:     Repeat,
					HasOperand: true,
					Operand:    "10",
				},
			},
			expectErrors: 1,
		},
		{
			raw: `repeat:10,,,,step`,
			expectActions: []ActionHolder{
				{
					Action:     Repeat,
					HasOperand: true,
					Operand:    "10",
					Subs: []ActionHolder{
						{Action: Step},
					},
				},
			},
		},
		{
			raw: `repeat:10,step++`,
			expectActions: []ActionHolder{
				{
					Action:     Repeat,
					HasOperand: true,
					Operand:    "10",
				},
			},
			expectErrors: 1,
		},
		{
			raw: `if:$foo>10,repeat:10,step`,
			expectActions: []ActionHolder{
				{
					Action:     If,
					HasOperand: true,
					Operand:    "$foo>10",
					Subs: []ActionHolder{
						{
							Action:     Repeat,
							HasOperand: true,
							Operand:    "10",
							Subs:       []ActionHolder{{Action: Step}},
						},
					},
				},
			},
		},
		{
			raw: `if:$foo>10,if:$foo>20,step`,
			expectActions: []ActionHolder{
				{
					Action:     If,
					HasOperand: true,
					Operand:    "$foo>10",
					Subs: []ActionHolder{
						{
							Action:     If,
							HasOperand: true,
							Operand:    "$foo>20",
							Subs:       []ActionHolder{{Action: Step}},
						},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			lines := strings.Split(tc.raw, "\n")
			result, errs := ParseShortcut("", lines)
			require.NotNil(t, result)
			require.Equal(t, len(tc.expectActions), len(result.Actions))
			for a, action := range result.Actions {
				assert.Equal(t, tc.expectActions[a], action)
			}
			require.Equal(t, tc.expectErrors, len(errs))

			if len(errs) == 0 {
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				enc.SetIndent("", "  ")
				err := enc.Encode(result)
				require.NoError(t, err)

				lines := []string{buf.String()}
				result2, errs2 := ParseShortcut("", lines)
				require.Empty(t, errs2)
				assert.Equal(t, result, result2)
				//fmt.Println(buf.String())
			}
		})
	}
}

func TestParseShortcut_Json(t *testing.T) {
	testCases := []struct {
		raw           string
		expectActions []ActionHolder
		expectErrors  int
	}{
		{
			raw:          `[`,
			expectErrors: 2,
		},
		{
			raw:           `["step"]`,
			expectActions: []ActionHolder{{Action: Step}},
		},
		{
			raw:           `["step-ahead"]`,
			expectActions: []ActionHolder{{Action: StepAhead}},
		},
		{
			raw:           `["step-ahead++"]`,
			expectActions: []ActionHolder{{Action: StepAhead, IsInc: true}},
		},
		{
			raw:           `["step-ahead--"]`,
			expectActions: []ActionHolder{{Action: StepAhead, IsDec: true}},
		},
		{
			raw:          `["step-ahead++--"]`,
			expectErrors: 1,
		},
		{
			raw:          `["step-ahead--++"]`,
			expectErrors: 1,
		},
		{
			raw:          `["randomization"]`,
			expectErrors: 1,
		},
		{
			raw:          `["unknown"]`,
			expectErrors: 1,
		},
		{
			raw:          `["step++"]`,
			expectErrors: 1,
		},
		{
			raw:          `["step--"]`,
			expectErrors: 1,
		},
		{
			raw:          `[null]`,
			expectErrors: 1,
		},
		{
			raw:           `[{"step-ahead-by": "10"}]`,
			expectActions: []ActionHolder{{Action: StepAheadBy, HasOperand: true, Operand: "10"}},
		},
		{
			raw:          `[{"unknown": "foo"}]`,
			expectErrors: 1,
		},
		{
			raw:          `[{"step": "foo"}]`,
			expectErrors: 1,
		},
		{
			raw:           `[{"step-ahead": 10}]`,
			expectActions: []ActionHolder{{Action: StepAhead, HasOperand: true, Operand: "10"}},
		},
		{
			raw:          `[{"step-ahead": ""}]`,
			expectErrors: 1,
		},
		{
			raw:          `[{"step-ahead": "10", "step-ahead-by": "10"}]`,
			expectErrors: 1,
		},
		{
			raw:          `[{"repeat": "10"}]`,
			expectErrors: 1,
		},
		{
			raw:          `[{"repeat": {}}]`,
			expectErrors: 1,
		},
		{
			raw:          `[{"repeat": {"actions":[]}}]`,
			expectErrors: 1,
		},
		{
			raw: `[{"repeat": {"for":10, "actions":[]}}]`,
			expectActions: []ActionHolder{{
				Action:     Repeat,
				HasOperand: true,
				Operand:    "10",
				Subs:       []ActionHolder{},
			}},
		},
		{
			raw: `[{"repeat": {"for":10, "actions":"step"}}]`,
			expectActions: []ActionHolder{{
				Action:     Repeat,
				HasOperand: true,
				Operand:    "10",
				Subs: []ActionHolder{
					{Action: Step},
				},
			}},
		},
		{
			raw:          `[{"step": "10"}]`,
			expectErrors: 1,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			lines := strings.Split(tc.raw, "\n")
			result, errs := ParseShortcut("", lines)
			require.NotNil(t, result)
			require.Equal(t, len(tc.expectActions), len(result.Actions))
			for a, action := range result.Actions {
				assert.Equal(t, tc.expectActions[a], action)
			}
			require.Equal(t, tc.expectErrors, len(errs))
		})
	}
}

func TestParseActionItem(t *testing.T) {
	testCases := []struct {
		raw                string
		expectRegistryItem registryItem
		expectIsInc        bool
		expectIsDec        bool
		expectHasOperand   bool
		expectOperand      string
		expectErr          string
	}{
		{
			expectRegistryItem: registryItem{action: NoAction},
		},
		{
			raw:                `run`,
			expectRegistryItem: registryItem{action: Run},
		},
		{
			raw:       `run:`,
			expectErr: `unexpected operand: action "run"`,
		},
		{
			raw:                `step-delay:10`,
			expectRegistryItem: registryItem{action: StepDelay},
			expectHasOperand:   true,
			expectOperand:      "10",
		},
		{
			raw:                `step-delay++`,
			expectRegistryItem: registryItem{action: StepDelay},
			expectIsInc:        true,
		},
		{
			raw:                `step-delay--`,
			expectRegistryItem: registryItem{action: StepDelay},
			expectIsDec:        true,
		},
		{
			raw:       `repeat`,
			expectErr: `expected operand: action "repeat"`,
		},
		{
			raw:       `repeat:`,
			expectErr: `expected operand: action "repeat"`,
		},
		{
			raw:                `heat-map`,
			expectRegistryItem: registryItem{action: HeatMap},
		},
		{
			raw:                `heat-map:true`,
			expectRegistryItem: registryItem{action: HeatMap},
			expectHasOperand:   true,
			expectOperand:      "true",
		},
		{
			raw:       `heat-map++`,
			expectErr: `cannot increment action "heat-map"`,
		},
		{
			raw:                `heat-map:foo++`,
			expectRegistryItem: registryItem{action: HeatMap},
			expectHasOperand:   true,
			expectOperand:      "foo++",
		},
		{
			raw:                `heat-map:foo--`,
			expectRegistryItem: registryItem{action: HeatMap},
			expectHasOperand:   true,
			expectOperand:      "foo--",
		},
		{
			raw:                `unknown`,
			expectRegistryItem: registryItem{action: Unknown},
		},
		{
			raw:       `cell-color-alive++`,
			expectErr: `cannot increment action "cell-color-alive"`,
		},
		{
			raw:       `cell-color-alive--`,
			expectErr: `cannot decrement action "cell-color-alive"`,
		},
		{
			raw:       `if`,
			expectErr: `expected operand: action "if"`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			ri, isInc, isDec, hasOperand, operand, err := parseActionItem(tc.raw)
			if tc.expectErr == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.expectRegistryItem.action, ri.action)
				assert.Equal(t, tc.expectIsInc, isInc)
				assert.Equal(t, tc.expectIsDec, isDec)
				assert.Equal(t, tc.expectHasOperand, hasOperand)
				assert.Equal(t, tc.expectOperand, operand)
			} else {
				require.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			}
		})
	}
}

func TestCutActionName(t *testing.T) {
	testCases := []struct {
		raw  string
		name string
		tail string
	}{
		{},
		{
			raw:  `run`,
			name: "run",
		},
		{
			raw:  `repeat:10,step`,
			name: "repeat",
			tail: ":10,step",
		},
		{
			raw:  `foo++`,
			name: "foo",
			tail: "++",
		},
		{
			raw:  `foo--`,
			name: "foo",
			tail: "--",
		},
		{
			raw:  `foo-bar--`,
			name: "foo-bar",
			tail: "--",
		},
		{
			raw:  `repeat:10,foo++`,
			name: "repeat",
			tail: ":10,foo++",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			name, tail := cutActionName(tc.raw)
			assert.Equal(t, tc.name, name)
			assert.Equal(t, tc.tail, tail)
		})
	}
}
