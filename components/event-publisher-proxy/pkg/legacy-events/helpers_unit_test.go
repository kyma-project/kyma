//go:build unit
// +build unit

package legacy_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	legacy "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
)

func TestParseApplicationNameFromPath(t *testing.T) {
	testCases := []struct {
		name           string
		givenInputPath string
		wantAppName    string
	}{
		{
			name:           "should return application when correct path is used",
			givenInputPath: "/application/v1/events",
			wantAppName:    "application",
		}, {
			name:           "should return application when extra slash is in the path",
			givenInputPath: "//application/v1/events",
			wantAppName:    "application",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotAppName := legacy.ParseApplicationNameFromPath(tc.givenInputPath)
			assert.Equal(t, tc.wantAppName, gotAppName)
		})
	}
}
