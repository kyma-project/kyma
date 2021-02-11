package eventtype

import "testing"

func TestBuilder(t *testing.T) {
	testCases := []struct {
		givenPrefix, givenApplicationName, givenEvent, givenVersion string
		wantEventType                                               string
	}{
		{
			givenPrefix: "sap.kyma", givenApplicationName: "testapp", givenEvent: "order.created", givenVersion: "v1",
			wantEventType: "sap.kyma.testapp.order.created.v1",
		},
		{
			givenPrefix: "sap.kyma", givenApplicationName: "test.app", givenEvent: "order.created", givenVersion: "v1",
			wantEventType: "sap.kyma.test.app.order.created.v1",
		},
		{
			givenPrefix: "sap.kyma", givenApplicationName: "test-app", givenEvent: "order.created", givenVersion: "v1",
			wantEventType: "sap.kyma.test-app.order.created.v1",
		},
	}

	for _, tc := range testCases {
		if got := build(tc.givenPrefix, tc.givenApplicationName, tc.givenEvent, tc.givenVersion); tc.wantEventType != got {
			t.Errorf("build failed, want [%s] but got [%s]", tc.wantEventType, got)
		}
	}
}
