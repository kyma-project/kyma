package eventtype

import "testing"

func TestBuilder(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			if got := build(tc.givenPrefix, tc.givenApplicationName, tc.givenEvent, tc.givenVersion); tc.wantEventType != got {
				t.Errorf("build failed, want [%s] but got [%s]", tc.wantEventType, got)
			}
		})
	}
}
