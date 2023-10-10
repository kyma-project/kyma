package builder

import (
	"context"
	"encoding/json"
	golog "log"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/fake"
	kymalogger "github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/stretchr/testify/require"
)

func Test_Build(t *testing.T) {
	t.Parallel()

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
			wantType:             "prefix.source1.order.created.v1",
			wantSource:           "source1",
		},
		{
			name:                 "should return cleaned source and type (without application)",
			givenSource:          "source1",
			givenType:            "o-rder.creat ed.v1",
			givenApplicationName: "appName1",
			wantType:             "prefix.source1.o-rder.created.v1",
			wantSource:           "source1",
		},
		{
			name:                 "should return application name as source",
			givenSource:          "appName1",
			givenType:            "order.created.v1",
			givenApplicationName: "appName1",
			wantType:             "prefix.appName1.order.created.v1",
			wantSource:           "appName1",
		},
		{
			name:                   "should return application label as source",
			givenSource:            "appName1",
			givenType:              "order.created.v1",
			givenApplicationName:   "appName1",
			givenApplicationLabels: map[string]string{application.TypeLabel: "t..est-apptype"},
			wantType:               "prefix.test-apptype.order.created.v1",
			wantSource:             "appName1",
		},
		{
			name:                   "should return error for providing invalid uri reference as source",
			givenSource:            "a@@p##p%%",
			givenType:              "order.created.v1",
			givenApplicationName:   "appName1",
			givenApplicationLabels: map[string]string{application.TypeLabel: "a@@p##p%%"},
			wantError:              true,
		},
		{
			name:                 "should return error if empty type",
			givenSource:          "source1",
			givenType:            "",
			givenApplicationName: "appName1",
			wantError:            true,
		},
		{
			name:                 "should return error if empty source",
			givenSource:          "",
			givenType:            "order.created.v1",
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

			genericBuilder := &GenericBuilder{
				typePrefix:        "prefix",
				applicationLister: appLister,
				logger:            logger,
				cleaner:           cleaner.NewJetStreamCleaner(logger),
			}

			// when
			buildEvent, buildErr := genericBuilder.Build(newEvent)

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

func Test_GetAppNameOrSource(t *testing.T) {
	t.Parallel()

	// init the logger
	logger, err := kymalogger.New("json", "debug")
	if err != nil {
		golog.Fatalf("Failed to initialize logger, error: %v", err)
	}

	testCases := []struct {
		name                          string
		givenApplicationName          string
		givenApplicationLabels        map[string]string
		givenSource                   string
		givenApplicationListerEnabled bool
		wantSource                    string
	}{
		{
			name:                          "should return application name instead of source name",
			givenSource:                   "appName1",
			givenApplicationName:          "appName1",
			givenApplicationListerEnabled: true,
			wantSource:                    "appName1",
		},
		{
			name:                          "should return application label instead of source name or app name",
			givenSource:                   "appName1",
			givenApplicationName:          "appName1",
			givenApplicationListerEnabled: true,
			givenApplicationLabels:        map[string]string{application.TypeLabel: "testapptype"},
			wantSource:                    "testapptype",
		},
		{
			name:                          "should return non-clean application label",
			givenSource:                   "appName1",
			givenApplicationName:          "appName1",
			givenApplicationListerEnabled: true,
			givenApplicationLabels:        map[string]string{application.TypeLabel: "t..e--s__t!!a@@p##p%%t^^y&&p**e"},
			wantSource:                    "t..e--s__t!!a@@p##p%%t^^y&&p**e",
		},
		{
			name:                          "should return source name as application does not exists",
			givenSource:                   "noapp1",
			givenApplicationName:          "appName1",
			givenApplicationListerEnabled: true,
			givenApplicationLabels:        map[string]string{application.TypeLabel: "testapptype"},
			wantSource:                    "noapp1",
		},
		{
			name:                          "should return source name when application lister is disabled",
			givenSource:                   "appName1",
			givenApplicationName:          "appName1",
			givenApplicationListerEnabled: false,
			givenApplicationLabels:        map[string]string{application.TypeLabel: "testapptype"},
			wantSource:                    "appName1",
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := applicationtest.NewApplication(tc.givenApplicationName, tc.givenApplicationLabels)

			var appLister *application.Lister
			if tc.givenApplicationListerEnabled {
				appLister = fake.NewApplicationListerOrDie(context.Background(), app)
			}

			genericBuilder := &GenericBuilder{
				applicationLister: appLister,
				logger:            logger,
			}

			namedLogger := logger.WithContext().Named(genericBuilderName).With("source", tc.givenSource)
			require.Equal(t, tc.wantSource, genericBuilder.GetAppNameOrSource(tc.givenSource, namedLogger))
		})
	}
}

func Test_getFinalSubject(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		givenTypePrefix string
		givenSource     string
		givenType       string
		wantSubject     string
	}{
		{
			name:            "should return correct subject",
			givenTypePrefix: "prefix",
			givenSource:     "test1",
			givenType:       "test2",
			wantSubject:     "prefix.test1.test2",
		},
		{
			name:            "should return correct subject",
			givenTypePrefix: "kyma",
			givenSource:     "inapp",
			givenType:       "order.created.v1",
			wantSubject:     "kyma.inapp.order.created.v1",
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			genericBuilder := &GenericBuilder{
				typePrefix: tc.givenTypePrefix,
			}

			require.Equal(t, tc.wantSubject, genericBuilder.getFinalSubject(tc.givenSource, tc.givenType))
		})
	}
}
