package eventtype

import (
	"context"
	"testing"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/fake"
)

func TestCleaner(t *testing.T) {
	testCases := []struct {
		givenEventTypePrefix   string
		givenApplicationName   string
		givenApplicationLabels map[string]string
		givenEventType         string
		wantEventType          string
		wantError              bool
	}{
		// valid even-types
		{
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "testapp",
			givenEventType:       "sap.kyma.testapp.order.created.v1",
			wantEventType:        "sap.kyma.testapp.order.created.v1",
			wantError:            false,
		},
		{
			givenEventTypePrefix:   "sap.kyma",
			givenApplicationName:   "testapp",
			givenApplicationLabels: map[string]string{application.TypeLabel: "testapptype"},
			givenEventType:         "sap.kyma.testapp.order.created.v1",
			wantEventType:          "sap.kyma.testapptype.order.created.v1",
			wantError:              false,
		},
		{
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "t..e--s__t!!a@@p##p%%",
			givenEventType:       "sap.kyma.t..e--s__t!!a@@p##p%%.order.created.v1",
			wantEventType:        "sap.kyma.testapp.order.created.v1",
			wantError:            false,
		},
		{
			givenEventTypePrefix:   "sap.kyma",
			givenApplicationName:   "t..e--s__t!!a@@p##p%%",
			givenApplicationLabels: map[string]string{application.TypeLabel: "t..e--s__t!!a@@p##p%%t^^y&&p**e"},
			givenEventType:         "sap.kyma.t..e--s__t!!a@@p##p%%.order.created.v1",
			wantEventType:          "sap.kyma.testapptype.order.created.v1",
			wantError:              false,
		},
		// valid even-types but have not existing applications
		{
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "test-app",
			givenEventType:       "sap.kyma.testapp.order.created.v1",
			wantEventType:        "sap.kyma.testapp.order.created.v1",
			wantError:            false,
		},
		{
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "testapp",
			givenEventType:       "sap.kyma.test-app.order.created.v1",
			wantEventType:        "sap.kyma.testapp.order.created.v1",
			wantError:            false,
		},
		// invalid even-types
		{
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "testapp",
			givenEventType:       "sap.kyma.testapp.order.created", // missing the version
			wantError:            true,
		},
		{
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "testapp",
			givenEventType:       "sap.testapp.order.created.v1", // prefix not matched
			wantError:            true,
		},
	}

	for _, tc := range testCases {
		app := applicationtest.NewApplication(tc.givenApplicationName, tc.givenApplicationLabels)
		appLister := fake.NewApplicationListerOrDie(context.Background(), app)
		cleaner := NewCleaner(tc.givenEventTypePrefix, appLister, ctrl.Log.WithName("cleaner"))

		gotEventType, err := cleaner.Clean(tc.givenEventType)

		if tc.wantError == true && err == nil {
			t.Errorf("clean should have failed with an error")
			continue
		}
		if tc.wantError != true && err != nil {
			t.Errorf("clean should have succeeded without an error")
			continue
		}
		if tc.wantError == true && err != nil {
			// error occurred as expected
			continue
		}

		if tc.wantEventType != gotEventType {
			t.Errorf("clean failed event-type[%s] prefix[%s], want event-type [%s] but got [%s]", tc.givenEventType, tc.givenEventTypePrefix, tc.wantEventType, gotEventType)
		}
	}
}
