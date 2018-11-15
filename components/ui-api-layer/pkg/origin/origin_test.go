package origin_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/origin"
	"github.com/stretchr/testify/assert"
)

func TestCheckFn(t *testing.T) {
	var testCases = []struct {
		allowedOrigins []string
		origin         string
		expected       bool
	}{
		{[]string{"foo.bar.local", "test.foo.com", "*.test.com"}, "foo.bar.local", true},
		{[]string{"foo.bar.local", "test.foo.com", "*.test.com"}, "bar.bar.local", false},
		{[]string{"foo.bar.local", "https://*.test.com"}, "https://sample.test.com", true},
		{[]string{"foo.bar.local", "https://*.test.com"}, "http://sample.test.com", false},
		{[]string{"foo.bar.local", "*.test.com"}, "sample.test.com", true},
	}

	for testCaseNo, testCase := range testCases {
		t.Run(fmt.Sprintf("TestCase%d", testCaseNo), func(t *testing.T) {
			check := origin.CheckFn(testCase.allowedOrigins)
			r, err := http.NewRequest("GET", "url", nil)
			require.NoError(t, err)
			r.Header.Add("Origin", testCase.origin)

			result := check(r)

			assert.Equal(t, testCase.expected, result)
		})
	}
}
