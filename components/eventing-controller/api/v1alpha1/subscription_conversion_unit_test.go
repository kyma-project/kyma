package v1alpha1_test

import (
	"testing"

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
				testingv2.WithMaxInFlight("nonint"),
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
				testingv2.WithSource(eventSource),
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
				testingv2.WithSource(eventSource),
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
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//WHEN
			t.Run("Test v1 to v2 conversion", func(t *testing.T) {
				// skip the conversion if the backwards conversion cannot succeed
				if testCase.wantErrMsgV2toV1 != "" {
					return
				}
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

func v1ToV2Assertions(t *testing.T, wantSub, convertedSub *v1alpha2.Subscription) {
	assert.Equal(t, wantSub.ObjectMeta, convertedSub.ObjectMeta)

	// Spec
	assert.Equal(t, wantSub.Spec.ID, convertedSub.Spec.ID)
	assert.Equal(t, wantSub.Spec.Sink, convertedSub.Spec.Sink)
	assert.Equal(t, wantSub.Spec.TypeMatching, convertedSub.Spec.TypeMatching)
	assert.Equal(t, wantSub.Spec.Source, convertedSub.Spec.Source)
	assert.Equal(t, wantSub.Spec.Types, convertedSub.Spec.Types)
	assert.Equal(t, wantSub.Spec.Config, convertedSub.Spec.Config)

	// Status
	assert.Equal(t, wantSub.Status.Ready, convertedSub.Status.Ready)
	assert.Equal(t, wantSub.Status.Conditions, convertedSub.Status.Conditions)
	assert.Equal(t, wantSub.Status.Types, convertedSub.Status.Types)

	assert.Equal(t, wantSub.Status.Backend, convertedSub.Status.Backend)
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
