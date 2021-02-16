package eventtype

import (
	"testing"
)

func TestParser(t *testing.T) {
	testCases := []struct {
		givenEventType      string
		givenPrefix         string
		wantApplicationName string
		wantEvent           string
		wantVersion         string
		wantError           bool
	}{
		// missing prefix
		{
			givenEventType: "sap.kyma.test.app.name.123.order.created.v1",
			givenPrefix:    "missing.prefix",
			wantError:      true,
		},
		// invalid event-types
		{
			givenEventType: "sap.kyma.order.created.v1",
			givenPrefix:    "sap.kyma",
			wantError:      true,
		},
		{
			givenEventType: "sap.kyma.created.v1",
			givenPrefix:    "sap.kyma",
			wantError:      true,
		},
		{
			givenEventType: "sap.kyma.v1",
			givenPrefix:    "sap.kyma",
			wantError:      true,
		},
		{
			givenEventType: "sap.kyma",
			givenPrefix:    "sap.kyma",
			wantError:      true,
		},
		// valid event-types
		{
			givenEventType:      "sap.kyma.test_app.order.created.v1",
			givenPrefix:         "sap.kyma",
			wantApplicationName: "test_app",
			wantEvent:           "order.created",
			wantVersion:         "v1",
		},
		{
			givenEventType:      "sap.kyma.test-app.order.created.v1",
			givenPrefix:         "sap.kyma",
			wantApplicationName: "test-app",
			wantEvent:           "order.created",
			wantVersion:         "v1",
		},
		{
			givenEventType:      "sap.kyma.test.app.order.created.v1",
			givenPrefix:         "sap.kyma",
			wantApplicationName: "test.app",
			wantEvent:           "order.created",
			wantVersion:         "v1",
		},
		{
			givenEventType:      "sap.kyma.test_app.name.123.order.created.v1",
			givenPrefix:         "sap.kyma",
			wantApplicationName: "test_app.name.123",
			wantEvent:           "order.created",
			wantVersion:         "v1",
		},
		{
			givenEventType:      "sap.kyma.test_app-name.123.order.created.v1",
			givenPrefix:         "sap.kyma",
			wantApplicationName: "test_app-name.123",
			wantEvent:           "order.created",
			wantVersion:         "v1",
		},
		{
			givenEventType:      "sap.kyma.test...app.order.created.v1",
			givenPrefix:         "sap.kyma",
			wantApplicationName: "test...app",
			wantEvent:           "order.created",
			wantVersion:         "v1",
		},
	}

	for _, tc := range testCases {
		applicationName, event, version, err := parse(tc.givenEventType, tc.givenPrefix)

		if tc.wantError == true && err == nil {
			t.Errorf("parse should have failed with an error")
			continue
		}
		if tc.wantError != true && err != nil {
			t.Errorf("parse should have succeeded without an error")
			continue
		}

		if tc.wantApplicationName != applicationName {
			t.Errorf("parse failed event-type[%s] prefix[%s], invalid application name, want [%s] but got [%s]", tc.givenEventType, tc.givenPrefix, tc.wantApplicationName, applicationName)
		}
		if tc.wantEvent != event {
			t.Errorf("parse failed event-type[%s] prefix[%s], invalid event, want [%s] but got [%s]", tc.givenEventType, tc.givenPrefix, tc.wantEvent, event)
		}
		if tc.wantVersion != version {
			t.Errorf("parse failed event-type[%s] prefix[%s], invalid version, want [%s] but got [%s]", tc.givenEventType, tc.givenPrefix, tc.wantVersion, version)
		}
	}
}
