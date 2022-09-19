//go:build unit
// +build unit

package eventtype

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name                                                        string
		givenPrefix, givenApplicationName, givenEvent, givenVersion string
		wantEventType                                               string
	}{
		{
			name:        "prefix is empty",
			givenPrefix: "", givenApplicationName: "test.app-1", givenEvent: "order.created", givenVersion: "v1",
			wantEventType: "test.app-1.order.created.v1",
		},
		{
			name:        "prefix is not empty",
			givenPrefix: "prefix", givenApplicationName: "test.app-1", givenEvent: "order.created", givenVersion: "v1",
			wantEventType: "prefix.test.app-1.order.created.v1",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// When
			eventType := build(tc.givenPrefix, tc.givenApplicationName, tc.givenEvent, tc.givenVersion)

			// Then
			assert.Equal(t, tc.wantEventType, eventType)
		})
	}
}
