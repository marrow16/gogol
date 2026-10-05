package shortcuts

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHolder_MarshalJSON(t *testing.T) {
	h := Holder{}
	data, err := json.Marshal(h)
	require.NoError(t, err)
	require.Equal(t, `[]`, string(data))

	h = Holder{
		Actions: []ActionHolder{},
	}
	data, err = json.Marshal(h)
	require.NoError(t, err)
	require.Equal(t, `[]`, string(data))

	h = Holder{
		Actions: []ActionHolder{{Action: Step}},
	}
	data, err = json.Marshal(h)
	require.NoError(t, err)
	require.Equal(t, `["step"]`, string(data))
}

func TestActionHolder_MarshalJSON(t *testing.T) {
	testCases := []struct {
		ah        ActionHolder
		expected  string
		expectErr bool
	}{
		{
			expectErr: true,
		},
		{
			ah:       ActionHolder{Action: StepAhead},
			expected: `"step-ahead"`,
		},
		{
			ah:       ActionHolder{Action: StepAhead, IsInc: true},
			expected: `"step-ahead++"`,
		},
		{
			ah:       ActionHolder{Action: StepAhead, IsDec: true},
			expected: `"step-ahead--"`,
		},
		{
			ah:       ActionHolder{Action: StepAhead, IsInc: true, IsDec: true},
			expected: `"step-ahead++"`,
		},
		{
			ah:       ActionHolder{Action: StepAhead, HasOperand: true, Operand: "foo"},
			expected: `{"step-ahead":"foo"}`,
		},
		{
			ah:       ActionHolder{Action: IterateCollectedRules},
			expected: `{"iterate-collected-rules":{"actions":[]}}`,
		},
		{
			ah:       ActionHolder{Action: Repeat, Operand: "10"},
			expected: `{"repeat":{"actions":[],"for":"10"}}`,
		},
		{
			ah:       ActionHolder{Action: Repeat},
			expected: `{"repeat":{"actions":[],"for":""}}`,
		},
		{
			ah:       ActionHolder{Action: If, Operand: "$foo>10"},
			expected: `{"if":{"actions":[],"condition":"$foo\u003e10"}}`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			data, err := json.Marshal(tc.ah)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, string(data))
			}
		})
	}
}
