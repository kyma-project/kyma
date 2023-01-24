package v1alpha1_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"

	testingv1 "github.com/kyma-project/kyma/components/eventing-controller/testing"
	testingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Conversion(t *testing.T) {
	type TestCase struct {
		name             string
		alpha1Sub        *v1alpha1.Subscription
		alpha2Sub        *v1alpha2.Subscription
		wantErrMsgV1toV2 string
		wantErrMsgV2toV1 string
	}

	testCases := []TestCase{
		{
			name: "Converting NATS Subscription with empty Status",
			alpha1Sub: newDefaultSubscription(
				testingv1.WithEmptyFilter(),
				testingv1.WithEmptyConfig(),
				testingv1.WithEmptyStatus(),
			),
			alpha2Sub: newV2DefaultSubscription(
				testingv2.WithEmptyStatus(),
				testingv2.WithEmptyConfig(),
			),
		},
		{
			name: "Converting NATS Subscription with empty Filters",
			alpha1Sub: newDefaultSubscription(
				testingv1.WithEmptyFilter(),
				testingv1.WithStatusCleanEventTypes(nil),
			),
			alpha2Sub: newV2DefaultSubscription(),
		},
		{
			name: "Converting NATS Subscription with multiple source which should result in a conversion error",
			alpha1Sub: newDefaultSubscription(
				testingv1.WithFilter("app", orderUpdatedEventType),
				testingv1.WithFilter("", orderDeletedEventTypeNonClean),
			),
			alpha2Sub:        newV2DefaultSubscription(),
			wantErrMsgV1toV2: v1alpha1.ErrorMultipleSourceMsg,
		},
		{
			name: "Converting NATS Subscription with non-convertable maxInFlight in the config which should result in a conversion error",
			alpha1Sub: newDefaultSubscription(
				testingv1.WithFilter("", orderUpdatedEventType),
			),
			alpha2Sub: newV2DefaultSubscription(
				testingv2.WithMaxInFlightMessages("nonint"),
			),
			wantErrMsgV2toV1: "strconv.Atoi: parsing \"nonint\": invalid syntax",
		},
		{
			name: "Converting NATS Subscription with Filters",
			alpha1Sub: newDefaultSubscription(
				testingv1.WithFilter(eventSource, orderCreatedEventType),
				testingv1.WithFilter(eventSource, orderUpdatedEventType),
				testingv1.WithFilter(eventSource, orderDeletedEventTypeNonClean),
				testingv1.WithStatusCleanEventTypes([]string{
					orderCreatedEventType,
					orderUpdatedEventType,
					orderDeletedEventType,
				}),
			),
			alpha2Sub: newV2DefaultSubscription(
				testingv2.WithEventSource(eventSource),
				testingv2.WithTypes([]string{
					orderCreatedEventType,
					orderUpdatedEventType,
					orderDeletedEventTypeNonClean,
				}),
				v2WithStatusTypes([]v1alpha2.EventType{
					{
						OriginalType: orderCreatedEventType,
						CleanType:    orderCreatedEventType,
					},
					{
						OriginalType: orderUpdatedEventType,
						CleanType:    orderUpdatedEventType,
					},
					{
						OriginalType: orderDeletedEventTypeNonClean,
						CleanType:    orderDeletedEventType,
					},
				}),
				v2WithStatusJetStreamTypes([]v1alpha2.JetStreamTypes{
					{
						OriginalType: orderCreatedEventType,
						ConsumerName: "",
					},
					{
						OriginalType: orderUpdatedEventType,
						ConsumerName: "",
					},
					{
						OriginalType: orderDeletedEventTypeNonClean,
						ConsumerName: "",
					},
				}),
			),
		},
		{
			name: "Converting BEB Subscription",
			alpha1Sub: newDefaultSubscription(
				testingv1.WithProtocolBEB(),
				v1WithWebhookAuthForBEB(),
				testingv1.WithFilter(eventSource, orderCreatedEventType),
				testingv1.WithFilter(eventSource, orderUpdatedEventType),
				testingv1.WithFilter(eventSource, orderDeletedEventTypeNonClean),
				testingv1.WithStatusCleanEventTypes([]string{
					orderCreatedEventType,
					orderUpdatedEventType,
					orderDeletedEventType,
				}),
				v1WithBEBStatusFields(),
			),
			alpha2Sub: newV2DefaultSubscription(
				testingv2.WithEventSource(eventSource),
				testingv2.WithTypes([]string{
					orderCreatedEventType,
					orderUpdatedEventType,
					orderDeletedEventTypeNonClean,
				}),
				testingv2.WithProtocolBEB(),
				testingv2.WithWebhookAuthForBEB(),
				v2WithStatusTypes([]v1alpha2.EventType{
					{
						OriginalType: orderCreatedEventType,
						CleanType:    orderCreatedEventType,
					},
					{
						OriginalType: orderUpdatedEventType,
						CleanType:    orderUpdatedEventType,
					},
					{
						OriginalType: orderDeletedEventTypeNonClean,
						CleanType:    orderDeletedEventType,
					},
				}),
				v2WithBEBStatusFields(),
			),
		},
		{
			name: "Converting Subscription with Protocol, ProtocolSettings and WebhookAuth",
			alpha1Sub: newDefaultSubscription(
				testingv1.WithProtocolBEB(),
				testingv1.WithProtocolSettings(
					testingv1.NewProtocolSettings(
						testingv1.WithAtLeastOnceQOS(),
						testingv1.WithRequiredWebhookAuth())),
				testingv1.WithFilter(eventSource, orderCreatedEventType),
				testingv1.WithStatusCleanEventTypes([]string{
					orderCreatedEventType,
				}),
			),
			alpha2Sub: newV2DefaultSubscription(
				testingv2.WithEventSource(eventSource),
				testingv2.WithTypes([]string{
					orderCreatedEventType,
				}),
				testingv2.WithProtocolBEB(),
				testingv2.WithConfigValue(v1alpha2.ProtocolSettingsQos,
					string(types.QosAtLeastOnce)),
				testingv2.WithConfigValue(v1alpha2.WebhookAuthGrantType,
					"client_credentials"),
				testingv2.WithConfigValue(v1alpha2.WebhookAuthClientID,
					"xxx"),
				testingv2.WithConfigValue(v1alpha2.WebhookAuthClientSecret,
					"xxx"),
				testingv2.WithConfigValue(v1alpha2.WebhookAuthTokenURL,
					"https://oauth2.xxx.com/oauth2/token"),
				v2WithStatusTypes([]v1alpha2.EventType{
					{
						OriginalType: orderCreatedEventType,
						CleanType:    orderCreatedEventType,
					},
				}),
			),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//WHEN
			t.Run("Test v1 to v2 conversion", func(t *testing.T) {
				// skip the conversion if the backwards conversion cannot succeed
				if testCase.wantErrMsgV2toV1 != "" {
					return
				}

				// initialize dummy cleaner
				cleaner := eventtype.CleanerFunc(func(et string) (string, error) { return et, nil })
				v1alpha1.InitializeEventTypeCleaner(cleaner)

				convertedV1Alpha2 := &v1alpha2.Subscription{}
				err := v1alpha1.V1ToV2(testCase.alpha1Sub, convertedV1Alpha2)
				if err != nil && testCase.wantErrMsgV1toV2 != "" {
					require.Equal(t, err.Error(), testCase.wantErrMsgV1toV2)
				} else {
					require.NoError(t, err)
					v1ToV2Assertions(t, testCase.alpha2Sub, convertedV1Alpha2)
				}
			})

			// test ConvertFrom
			t.Run("Test v2 to v1 conversion", func(t *testing.T) {
				// skip the backwards conversion if the initial one cannot succeed
				if testCase.wantErrMsgV1toV2 != "" {
					return
				}
				convertedV1Alpha1 := &v1alpha1.Subscription{}
				err := v1alpha1.V2ToV1(convertedV1Alpha1, testCase.alpha2Sub)
				if err != nil && testCase.wantErrMsgV2toV1 != "" {
					require.Equal(t, err.Error(), testCase.wantErrMsgV2toV1)
				} else {
					require.NoError(t, err)
					v2ToV1Assertions(t, testCase.alpha1Sub, convertedV1Alpha1)
				}

			})
		})
	}
}

// Test_CleanupInV1ToV2Conversion test the cleaning from non-alphanumeric characters
// and also merging of segments in event types if they exceed the limit.
func Test_CleanupInV1ToV2Conversion(t *testing.T) {
	type TestCase struct {
		name           string
		givenAlpha1Sub *v1alpha1.Subscription
		givenPrefix    string
		wantTypes      []string
		wantError      bool
	}

	testCases := []TestCase{
		{
			name: "success if prefix is empty",
			givenAlpha1Sub: newDefaultSubscription(
				testingv1.WithFilter(eventSource, "testapp.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1"),
			),
			givenPrefix: "",
			wantTypes: []string{
				"testapp.Segment1Part1Part2.Segment2Part1Part2.v1",
			},
		},
		{
			name:        "success if the given event has more than two segments",
			givenPrefix: "prefix",
			givenAlpha1Sub: newDefaultSubscription(
				testingv1.WithFilter(eventSource, "prefix.testapp.Segment1.Segment2.Segment3."+
					"Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1"),
			),
			wantTypes: []string{
				"prefix.testapp.Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
			},
			wantError: false,
		},
		{
			name:        "success if the application name needs to be cleaned",
			givenPrefix: "prefix",
			givenAlpha1Sub: newDefaultSubscription(
				testingv1.WithFilter(eventSource, "prefix.te--s__t!!a@@p##p%%.Segment1-Part1-Part2-Ä."+
					"Segment2-Part1-Part2-Ä.v1"),
			),
			wantTypes: []string{
				"prefix.testapp.Segment1Part1Part2.Segment2Part1Part2.v1",
			},
			wantError: false,
		},
		{
			name:        "success if the application name needs to be cleaned and event has more than two segments",
			givenPrefix: "prefix",
			givenAlpha1Sub: newDefaultSubscription(
				testingv1.WithFilter(eventSource, "prefix.te--s__t!!a@@p##p%%.Segment1.Segment2.Segment3."+
					"Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1"),
			),
			wantTypes: []string{
				"prefix.testapp.Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
			},
			wantError: false,
		},
		{
			name:        "success if there are multiple filters",
			givenPrefix: "prefix",
			givenAlpha1Sub: newDefaultSubscription(
				testingv1.WithFilter(eventSource, "prefix.test-app.Segme@@nt1.Segment2.Segment3."+
					"Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1"),
				testingv1.WithFilter(eventSource, "prefix.testapp.Segment1.Segment2.Segment3."+
					"Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1"),
			),
			wantTypes: []string{
				"prefix.testapp.Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
				"prefix.testapp.Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
			},
			wantError: false,
		},
		// invalid even-types
		{
			name:        "fail if the prefix is invalid",
			givenPrefix: "prefix",
			givenAlpha1Sub: newDefaultSubscription(
				testingv1.WithFilter(eventSource, "invalid.test-app.Segme@@nt1.Segment2.Segment3."+
					"Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1"),
			),
			wantError: true,
		},
		{
			name:        "fail if the prefix is missing",
			givenPrefix: "prefix",
			givenAlpha1Sub: newDefaultSubscription(
				testingv1.WithFilter(eventSource, "test-app.Segme@@nt1.Segment2.Segment3."+
					"Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1"),
			),
			wantError: true,
		},
		{
			name:        "fail if the event-type is incomplete",
			givenPrefix: "prefix",
			givenAlpha1Sub: newDefaultSubscription(
				testingv1.WithFilter(eventSource, "prefix.testapp.Segment1-Part1-Part2-Ä.v1"),
			),
			wantError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// given
			testLogger, err := logger.New("json", "info")
			require.NoError(t, err)

			// initialize dummy cleaner
			cleaner := eventtype.NewSimpleCleaner(tc.givenPrefix, testLogger)
			v1alpha1.InitializeEventTypeCleaner(cleaner)

			// initialize v1alpha2 Subscription instance
			convertedV1Alpha2 := &v1alpha2.Subscription{}

			// when
			err = v1alpha1.V1ToV2(tc.givenAlpha1Sub, convertedV1Alpha2)

			// then
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantTypes, convertedV1Alpha2.Spec.Types)
			}
		})
	}
}

func v1ToV2Assertions(t *testing.T, wantSub, convertedSub *v1alpha2.Subscription) {
	assert.Equal(t, wantSub.ObjectMeta, convertedSub.ObjectMeta)

	// Spec
	assert.Equal(t, wantSub.Spec.ID, convertedSub.Spec.ID)
	assert.Equal(t, wantSub.Spec.Sink, convertedSub.Spec.Sink)
	assert.Equal(t, wantSub.Spec.TypeMatching, convertedSub.Spec.TypeMatching)
	assert.Equal(t, wantSub.Spec.Source, convertedSub.Spec.Source)
	assert.Equal(t, wantSub.Spec.Types, convertedSub.Spec.Types)
	assert.Equal(t, wantSub.Spec.Config, convertedSub.Spec.Config)
}

func v2ToV1Assertions(t *testing.T, wantSub, convertedSub *v1alpha1.Subscription) {
	assert.Equal(t, wantSub.ObjectMeta, convertedSub.ObjectMeta)

	// Spec
	assert.Equal(t, wantSub.Spec.ID, convertedSub.Spec.ID)
	assert.Equal(t, wantSub.Spec.Sink, convertedSub.Spec.Sink)
	assert.Equal(t, wantSub.Spec.Protocol, convertedSub.Spec.Protocol)
	assert.Equal(t, wantSub.Spec.ProtocolSettings, convertedSub.Spec.ProtocolSettings)

	assert.Equal(t, wantSub.Spec.Filter, convertedSub.Spec.Filter)
	assert.Equal(t, wantSub.Spec.Config, convertedSub.Spec.Config)

	// Status
	assert.Equal(t, wantSub.Status.Ready, convertedSub.Status.Ready)
	assert.Equal(t, wantSub.Status.Conditions, convertedSub.Status.Conditions)
	assert.Equal(t, wantSub.Status.CleanEventTypes, convertedSub.Status.CleanEventTypes)

	// BEB fields
	assert.Equal(t, wantSub.Status.Ev2hash, convertedSub.Status.Ev2hash)
	assert.Equal(t, wantSub.Status.Emshash, convertedSub.Status.Emshash)
	assert.Equal(t, wantSub.Status.ExternalSink, convertedSub.Status.ExternalSink)
	assert.Equal(t, wantSub.Status.FailedActivation, convertedSub.Status.FailedActivation)
	assert.Equal(t, wantSub.Status.APIRuleName, convertedSub.Status.APIRuleName)
	assert.Equal(t, wantSub.Status.EmsSubscriptionStatus, convertedSub.Status.EmsSubscriptionStatus)

	assert.Equal(t, wantSub.Status.Config, convertedSub.Status.Config)
}
