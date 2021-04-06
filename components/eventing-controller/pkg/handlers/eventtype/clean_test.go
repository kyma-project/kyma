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
		name                   string
		givenEventTypePrefix   string
		givenApplicationName   string
		givenApplicationLabels map[string]string
		givenEventType         string
		wantEventType          string
		wantError              bool
	}{
		// valid even-types for existing applications
		{
			name:                 "success if the given application name is clean",
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "testapp",
			givenEventType:       "sap.kyma.testapp.Segment1.Segment2.v1",
			wantEventType:        "sap.kyma.testapp.Segment1.Segment2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the given application name is clean and event has more than two segments",
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "testapp",
			givenEventType:       "sap.kyma.testapp.Segment1.Segment2.Segment3.Segment4.Segment5.v1",
			wantEventType:        "sap.kyma.testapp.Segment1Segment2Segment3Segment4.Segment5.v1",
			wantError:            false,
		},
		{
			name:                   "success if the given application type label is clean",
			givenEventTypePrefix:   "sap.kyma",
			givenApplicationName:   "testapp",
			givenApplicationLabels: map[string]string{application.TypeLabel: "testapptype"},
			givenEventType:         "sap.kyma.testapp.Segment1.Segment2.v1",
			wantEventType:          "sap.kyma.testapptype.Segment1.Segment2.v1",
			wantError:              false,
		},
		{
			name:                   "success if the given application type label is clean and event has more than two segments",
			givenEventTypePrefix:   "sap.kyma",
			givenApplicationName:   "testapp",
			givenApplicationLabels: map[string]string{application.TypeLabel: "testapptype"},
			givenEventType:         "sap.kyma.testapp.Segment1.Segment2.Segment3.Segment4.Segment5.v1",
			wantEventType:          "sap.kyma.testapptype.Segment1Segment2Segment3Segment4.Segment5.v1",
			wantError:              false,
		},
		{
			name:                 "success if the given application name needs to be cleaned",
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "te--s__t!!a@@p##p%%",
			givenEventType:       "sap.kyma.te--s__t!!a@@p##p%%.Segment1.Segment2.v1",
			wantEventType:        "sap.kyma.testapp.Segment1.Segment2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the given application name needs to be cleaned and event has more than two segments",
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "te--s__t!!a@@p##p%%",
			givenEventType:       "sap.kyma.te--s__t!!a@@p##p%%.Segment1.Segment2.Segment3.Segment4.Segment5.v1",
			wantEventType:        "sap.kyma.testapp.Segment1Segment2Segment3Segment4.Segment5.v1",
			wantError:            false,
		},
		{
			name:                   "success if the given application type label needs to be cleaned",
			givenEventTypePrefix:   "sap.kyma",
			givenApplicationName:   "te--s__t!!a@@p##p%%",
			givenApplicationLabels: map[string]string{application.TypeLabel: "t..e--s__t!!a@@p##p%%t^^y&&p**e"},
			givenEventType:         "sap.kyma.te--s__t!!a@@p##p%%.Segment1.Segment2.v1",
			wantEventType:          "sap.kyma.testapptype.Segment1.Segment2.v1",
			wantError:              false,
		},
		{
			name:                   "success if the given application type label needs to be cleaned and event has more than two segments",
			givenEventTypePrefix:   "sap.kyma",
			givenApplicationName:   "te--s__t!!a@@p##p%%",
			givenApplicationLabels: map[string]string{application.TypeLabel: "t..e--s__t!!a@@p##p%%t^^y&&p**e"},
			givenEventType:         "sap.kyma.te--s__t!!a@@p##p%%.Segment1.Segment2.Segment3.Segment4.Segment5.v1",
			wantEventType:          "sap.kyma.testapptype.Segment1Segment2Segment3Segment4.Segment5.v1",
			wantError:              false,
		},
		// valid even-types for non-existing applications
		{
			name:                 "success if the given application name is clean for non-existing application",
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "",
			givenEventType:       "sap.kyma.test-app.Segment1.Segment2.v1",
			wantEventType:        "sap.kyma.testapp.Segment1.Segment2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the given application name is clean for non-existing application and event has more than two segments",
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "",
			givenEventType:       "sap.kyma.test-app.Segment1.Segment2.Segment3.Segment4.Segment5.v1",
			wantEventType:        "sap.kyma.testapp.Segment1Segment2Segment3Segment4.Segment5.v1",
			wantError:            false,
		},
		{
			name:                 "success if the given application name is not clean for non-existing application",
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "",
			givenEventType:       "sap.kyma.testapp.Segment1.Segment2.v1",
			wantEventType:        "sap.kyma.testapp.Segment1.Segment2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the given application name is not clean for non-existing application and event has more than two segments",
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "",
			givenEventType:       "sap.kyma.testapp.Segment1.Segment2.Segment3.Segment4.Segment5.v1",
			wantEventType:        "sap.kyma.testapp.Segment1Segment2Segment3Segment4.Segment5.v1",
			wantError:            false,
		},
		// invalid even-types
		{
			name:                 "fail if the prefix is invalid",
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "testapp",
			givenEventType:       "sapkyma.testapp.Segment1.Segment2.v1",
			wantError:            true,
		},
		{
			name:                 "fail if the prefix is missing",
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "testapp",
			givenEventType:       "testapp.Segment1.Segment2.v1",
			wantError:            true,
		},
		{
			name:                 "fail if the event-type is incomplete",
			givenEventTypePrefix: "sap.kyma",
			givenApplicationName: "testapp",
			givenEventType:       "sap.kyma.testapp.Segment1.v1",
			wantError:            true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := applicationtest.NewApplication(tc.givenApplicationName, tc.givenApplicationLabels)
			appLister := fake.NewApplicationListerOrDie(context.Background(), app)
			cleaner := NewCleaner(tc.givenEventTypePrefix, appLister, ctrl.Log.WithName("cleaner"))
			gotEventType, err := cleaner.Clean(tc.givenEventType)

			if tc.wantError == true && err == nil {
				t.Fatalf("%s: should have failed with an error", tc.name)
			}
			if tc.wantError != true && err != nil {
				t.Fatalf("%s: should have succeeded without an error", tc.name)
			}
			if tc.wantError == true && err != nil {
				// error occurred as expected
				return
			}
			if tc.wantEventType != gotEventType {
				t.Fatalf("%s: failed event-type[%s] prefix[%s], want event-type [%s] but got [%s]",
					tc.name, tc.givenEventType, tc.givenEventTypePrefix, tc.wantEventType, gotEventType)
			}
		})
	}
}
