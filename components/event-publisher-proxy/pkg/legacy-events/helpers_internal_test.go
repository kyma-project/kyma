//go:build unit
// +build unit

package legacy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		tc := tc
		eventType := formatEventType(tc.givenPrefix, tc.givenApplication, tc.givenEventType, tc.givenVersion)
		assert.Equal(t, tc.wantEventType, eventType)
	}
}
