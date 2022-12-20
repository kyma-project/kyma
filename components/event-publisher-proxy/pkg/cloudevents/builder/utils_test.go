package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_checkForEmptySegments(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name          string
		givenSegments []string
		wantResult    bool
	}{
		{
			name:          "should pass if all segments are non-empty",
			givenSegments: []string{"one", "two", "three"},
			wantResult:    false,
		},
		{
			name:          "should fail if any segment is empty",
			givenSegments: []string{"one", "", "three"},
			wantResult:    true,
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.wantResult, CheckForEmptySegments(tc.givenSegments))
		})
	}
}
