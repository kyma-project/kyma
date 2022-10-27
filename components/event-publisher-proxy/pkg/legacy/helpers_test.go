package legacy

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			gotAppName := ParseApplicationNameFromPath(tc.givenInputPath)
			assert.Equal(t, tc.wantAppName, gotAppName)
		})
	}
}

func TestFormatEventType(t *testing.T) {
	testCases := []struct {
		givenPrefix      string
		givenApplication string
		givenEventType   string
		givenVersion     string
		wantEventType    string
	}{
		{
			givenPrefix:      "prefix",
			givenApplication: "app",
			givenEventType:   "order.foo",
			givenVersion:     "v1",
			wantEventType:    "prefix.app.order.foo.v1",
		},
		{
			givenPrefix:      "prefix",
			givenApplication: "app",
			givenEventType:   "order-foo",
			givenVersion:     "v1",
			wantEventType:    "prefix.app.order-foo.v1",
		},
		{
			givenPrefix:      "prefix",
			givenApplication: "app",
			givenEventType:   "order-foo.bar@123",
			givenVersion:     "v1",
			wantEventType:    "prefix.app.order-foo.bar@123.v1",
		},
		{
			givenPrefix:      "",
			givenApplication: "app",
			givenEventType:   "order-foo",
			givenVersion:     "v1",
			wantEventType:    "app.order-foo.v1",
		},
		{
			givenPrefix:      "",
			givenApplication: "app",
			givenEventType:   "order-foo.bar@123",
			givenVersion:     "v1",
			wantEventType:    "app.order-foo.bar@123.v1",
		},
	}
	for _, tc := range testCases {
		eventType := formatEventType(tc.givenPrefix, tc.givenApplication, tc.givenEventType, tc.givenVersion)
		assert.Equal(t, tc.wantEventType, eventType)
	}
}
