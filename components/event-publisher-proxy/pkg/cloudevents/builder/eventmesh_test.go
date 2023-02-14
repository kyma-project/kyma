package builder

import (
	"context"
	"encoding/json"
	"fmt"
	golog "log"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/fake"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
	kymalogger "github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/stretchr/testify/require"
)

func Test_EventMesh_Build(t *testing.T) {
	t.Parallel()

	const sampleEventMeshNamespace = "/default/sample1/kyma1"
	const eventMeshPrefix = "one.two.three"

	// init the logger
	logger, err := kymalogger.New("json", "debug")
	if err != nil {
		golog.Fatalf("Failed to initialize logger, error: %v", err)
	}

	testCases := []struct {
		name                   string
		givenSource            string
		givenType              string
		givenApplicationName   string
		givenApplicationLabels map[string]string
		wantType               string
		wantSource             string
		wantError              bool
	}{
		{
			name:                 "should return correct source and type (without application)",
			givenSource:          "source1",
			givenType:            "order.created.v1",
			givenApplicationName: "appName1",
			wantType:             fmt.Sprintf("%s.source1.order.created.v1", eventMeshPrefix),
			wantSource:           sampleEventMeshNamespace,
		},
		{
			name:                 "should return cleaned source and type (without application)",
			givenSource:          "source1",
			givenType:            "o-rder.creat ed.v1",
			givenApplicationName: "appName1",
			wantType:             fmt.Sprintf("%s.source1.order.created.v1", eventMeshPrefix),
			wantSource:           sampleEventMeshNamespace,
		},
		{
			name:                 "should return merged type segments if exceeds segments limit",
			givenSource:          "source1",
			givenType:            "haha.hehe.hmm.order.created.v1",
			givenApplicationName: "appName1",
			wantType:             fmt.Sprintf("%s.source1.hahahehehmmorder.created.v1", eventMeshPrefix),
			wantSource:           sampleEventMeshNamespace,
		},
		{
			name:                 "should return application name as source",
			givenSource:          "appName1",
			givenType:            "order.created.v1",
			givenApplicationName: "appName1",
			wantType:             fmt.Sprintf("%s.appName1.order.created.v1", eventMeshPrefix),
			wantSource:           sampleEventMeshNamespace,
		},
		{
			name:                   "should return application label as source",
			givenSource:            "appName1",
			givenType:              "order.created.v1",
			givenApplicationName:   "appName1",
			givenApplicationLabels: map[string]string{application.TypeLabel: "t..e--s__t!!a@@p##p%%t^^y&&p**e"},
			wantType:               fmt.Sprintf("%s.testapptype.order.created.v1", eventMeshPrefix),
			wantSource:             sampleEventMeshNamespace,
		},
		{
			name:                 "should return error if empty type",
			givenSource:          "source1",
			givenType:            "",
			givenApplicationName: "appName1",
			wantError:            true,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// given
			// build cloud event
			builder := testingutils.NewCloudEventBuilder()
			payload, _ := builder.BuildStructured()
			newEvent := cloudevents.NewEvent()
			err = json.Unmarshal([]byte(payload), &newEvent)
			require.NoError(t, err)
			newEvent.SetType(tc.givenType)
			newEvent.SetSource(tc.givenSource)

			appLister := fake.NewApplicationListerOrDie(
				context.Background(),
				applicationtest.NewApplication(tc.givenApplicationName, tc.givenApplicationLabels))

			eventMeshBuilder := NewEventMeshBuilder(
				eventMeshPrefix,
				sampleEventMeshNamespace,
				cleaner.NewEventMeshCleaner(logger),
				appLister,
				logger,
			)

			// when
			buildEvent, buildErr := eventMeshBuilder.Build(newEvent)

			// then
			if tc.wantError {
				require.Error(t, buildErr)
			} else {
				require.NoError(t, buildErr)
				require.Equal(t, tc.wantSource, buildEvent.Source())
				require.Equal(t, tc.wantType, buildEvent.Type())

				// check that original type header exists
				originalType, ok := buildEvent.Extensions()[OriginalTypeHeaderName]
				require.True(t, ok)
				require.Equal(t, tc.givenType, originalType)
			}
		})
	}
}
