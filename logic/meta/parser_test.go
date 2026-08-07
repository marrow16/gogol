package meta

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestMustParseRule_Panics(t *testing.T) {
	require.Panics(t, func() {
		_ = MustParseRule("XXX")
	})
}

func TestParseRule(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "B( +1 )/S(+457, !3, -01, -12 )",
			expected: "B(+1) / S(+457,!3,-01,-12)",
		},
		{
			input:    "B ( +1 )",
			expected: "B(+1)",
		},
		{
			input:    "S( !3 )",
			expected: "S(!3)",
		},
		{
			input:    "S(+457, !3, -01, -12 )  /  B( +1 )",
			expected: "B(+1) / S(+457,!3,-01,-12)",
		},
		{
			input:    "S( #0..1 : 0123)",
			expected: "S(#0..1:0123)",
		},
		{
			input:    "O( +1 ) / X(-1) / A(+457, !3, -01, -12 )",
			expected: "&(+457,!3,-01,-12) / |(+1) / ^(-1)",
		},
		{
			input:    "S(+457, !3, -01, -12 )  /  B( +1 ) O( +1 ) / X(-1) / A(+457, !3, -01, -12 )",
			expected: "B(+1) / S(+457,!3,-01,-12) / &(+457,!3,-01,-12) / |(+1) / ^(-1)",
		},
		{
			input:    "P ( 123 )",
			expected: "P(123)",
		},
		{
			input:    "P(0xFF)",
			expected: "P(255)",
		},
		{
			input:    "P(0b11111111)",
			expected: "P(255)",
		},
		{
			input:    "P(0o377)",
			expected: "P(255)",
		},
		{
			input:    "P ( 123, 456 - 200 )",
			expected: "P(123,200-456)",
		},
		{
			input:    "AllOf(S(+457, !3, -01, -12 )  /  B( +1 ), S(+4, !3, -01, -12 )  /  B( +1 ))",
			expected: "AllOf(B(+1) / S(+457,!3,-01,-12), B(+1) / S(+4,!3,-01,-12))",
		},
		{
			input:    "AnyOf(AllOf(S(+457, !3, -01, -12 )  /  B( +1 ), S(+4, !3, -01, -12 )  /  B( +1 )), AllOf(S(+7, !3, -01, -12 )  /  B( +1 )))",
			expected: "AnyOf(AllOf(B(+1) / S(+457,!3,-01,-12), B(+1) / S(+4,!3,-01,-12)), AllOf(B(+1) / S(+7,!3,-01,-12)))",
		},
		{
			input: `B( +1 ) // this is a comment
	/S(+457, !3, -01, -12 )`,
			expected: "B(+1) / S(+457,!3,-01,-12)",
		},
		{
			input: `AllOf(
    B(!2345) / S(+47,!3),
	//B(!2345) / S(+47,!3),
    AnyOf(
        //B(+0,!1) / S(!012,-568),
        B(+0,!16) / S(+568,!012),
    )
)`,
			expected: "AllOf(B(!2345) / S(+47,!3), AnyOf(B(+0,!16) / S(+568,!012)))",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			mr, err := ParseRule(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, mr.String())
		})
	}
}

func TestParseRule_Errors(t *testing.T) {
	testCases := []struct {
		input    string
		position int
	}{
		{
			input: " /",
		},
		{
			input:    "S() / s()",
			position: 6,
		},
		{
			input:    "S(+9)",
			position: 3,
		},
		{
			input:    "S(!9)",
			position: 3,
		},
		{
			input:    "S(-9)",
			position: 3,
		},
		{
			input:    "S",
			position: 1,
		},
		{
			input:    "S(+1,)",
			position: 5,
		},
		{
			input:    "S(+1,-1",
			position: 7,
		},
		{
			input:    "S(+1,?1)",
			position: 5,
		},
		{
			input:    "S(#X)",
			position: 3,
		},
		{
			input:    "S(#0.)",
			position: 4,
		},
		{
			input:    "S(#0..X)",
			position: 6,
		},
		{
			input:    "S(#0..1)",
			position: 7,
		},
		{
			input:    "S(#0..10:012345678)",
			position: 6,
		},
		{
			input:    "S(#0..1:)",
			position: 8,
		},
		{
			input:    "S(#0..1:9)",
			position: 8,
		},
		{
			input:    "S(#1..0:8)",
			position: 7,
		},
		{
			input:    "S(#0..8:8)",
			position: 9,
		},
		{
			input:    "S(+44)",
			position: 4,
		},
		{
			input:    "S(",
			position: 2,
		},
		{
			input:    "B(+9)",
			position: 3,
		},
		{
			input:    "B",
			position: 1,
		},
		{
			input:    "P(-1)",
			position: 2,
		},
		{
			input:    "P(300000)",
			position: 2,
		},
		{
			input:    "P(not-a-number)",
			position: 2,
		},
		{
			input:    "P(0-not-a-number)",
			position: 4,
		},
		{
			input:    "P",
			position: 1,
		},
		{
			input:    "P()",
			position: 2,
		},
		{
			input:    "P(0,)",
			position: 4,
		},
		{
			input:    "P(0,1",
			position: 5,
		},
		{
			input:    "A(+9)",
			position: 3,
		},
		{
			input:    "O(+9)",
			position: 3,
		},
		{
			input:    "X(+9)",
			position: 3,
		},
		{
			input:    `AnyOf(AllOf(S(+457, !3, -01, -12 )  /  B( +1 ), S(+4, !3, -01, -12 )  /  B( +1 )), AllOf(S(+7, !3, -01, -12 ) / ?  B( +1 )))`,
			position: 112,
		},
		{
			input:    `AnyOf(S(+457, !3, -01, -12)`,
			position: 5,
		},
		{
			input:    `AnyOf(`,
			position: 5,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			_, err := ParseRule(tc.input)
			require.Error(t, err)
			if tc.position != -1 {
				errPrefix := fmt.Sprintf("parse meta-rule at position %d:", tc.position)
				assert.True(t, strings.HasPrefix(err.Error(), errPrefix), err.Error())
			}
		})
	}
}
