package telegram

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPairResult_DisplayName(t *testing.T) {
	cases := []struct {
		name     string
		result   PairResult
		expected string
	}{
		{
			name:     "full name",
			result:   PairResult{FirstName: "Leon", LastName: "Letto"},
			expected: "Leon Letto",
		},
		{
			name:     "first name only",
			result:   PairResult{FirstName: "Leon"},
			expected: "Leon",
		},
		{
			name:     "empty name",
			result:   PairResult{},
			expected: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.result.DisplayName())
		})
	}
}
