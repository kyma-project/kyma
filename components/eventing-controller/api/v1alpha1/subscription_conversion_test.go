package v1alpha1

import (
	"log"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Conversion(t *testing.T) {
	type TestCase struct {
		name       string
		alpha1Sub  *Subscription
		alpha2Sub  *v1alpha2.Subscription
		wantErrMsg string
	}

	testCases := []TestCase{
		{
			name: "Converting NATS Subscription with empty Filters",
			alpha1Sub: NewDefaultSubscription(
				WithEmptyFilter(),
			),
			alpha2Sub: v1alpha2.NewDefaultSubscription(),
		},
		{
			name: "Converting NATS Subscription with multiple source which should result in a conversion error",
			alpha1Sub: NewDefaultSubscription(
				WithFilter("app", OrderUpdatedEventType),
				WithFilter("", OrderDeletedEventTypeNonClean),
			),
			alpha2Sub:  v1alpha2.NewDefaultSubscription(),
			wantErrMsg: multipleSourceErrMsg,
		},
		{
			name: "Converting NATS Subscription with Filters",
			alpha1Sub: NewDefaultSubscription(
				WithFilter(EventSource, OrderCreatedEventType),
				WithFilter(EventSource, OrderUpdatedEventType),
				WithFilter(EventSource, OrderDeletedEventTypeNonClean),
				WithStatusCleanEventTypes([]string{
					OrderCreatedEventType,
					OrderUpdatedEventType,
					OrderDeletedEventType,
				}),
			),
			alpha2Sub: v1alpha2.NewDefaultSubscription(
				v1alpha2.WithSource(EventSource),
				v1alpha2.WithTypes([]string{
					OrderCreatedEventType,
					OrderUpdatedEventType,
					OrderDeletedEventTypeNonClean,
				}),
				v1alpha2.WithStatusTypes([]v1alpha2.EventType{
					{
						OriginalType: OrderCreatedEventType,
						CleanType:    OrderCreatedEventType,
					},
					{
						OriginalType: OrderUpdatedEventType,
						CleanType:    OrderUpdatedEventType,
					},
					{
						OriginalType: OrderDeletedEventTypeNonClean,
						CleanType:    OrderDeletedEventType,
					},
				}),
				v1alpha2.WithStatusJetStreamTypes([]v1alpha2.JetStreamTypes{
					{
						OriginalType: OrderCreatedEventType,
						ConsumerName: "",
					},
					{
						OriginalType: OrderUpdatedEventType,
						ConsumerName: "",
					},
					{
						OriginalType: OrderDeletedEventTypeNonClean,
						ConsumerName: "",
					},
				}),
			),
		},
		{
			name: "Converting BEB Subscription",
			alpha1Sub: NewDefaultSubscription(
				WithProtocolBEB(),
				WithWebhookAuthForBEB(),
				WithFilter(EventSource, OrderCreatedEventType),
				WithFilter(EventSource, OrderUpdatedEventType),
				WithFilter(EventSource, OrderDeletedEventTypeNonClean),
				WithStatusCleanEventTypes([]string{
					OrderCreatedEventType,
					OrderUpdatedEventType,
					OrderDeletedEventType,
				}),
				WithBEBStatusFields(),
			),
			alpha2Sub: v1alpha2.NewDefaultSubscription(
				v1alpha2.WithSource(EventSource),
				v1alpha2.WithTypes([]string{
					OrderCreatedEventType,
					OrderUpdatedEventType,
					OrderDeletedEventTypeNonClean,
				}),
				v1alpha2.WithProtocolBEB(),
				v1alpha2.WithWebhookAuthForBEB(),
				v1alpha2.WithStatusTypes([]v1alpha2.EventType{
					{
						OriginalType: OrderCreatedEventType,
						CleanType:    OrderCreatedEventType,
					},
					{
						OriginalType: OrderUpdatedEventType,
						CleanType:    OrderUpdatedEventType,
					},
					{
						OriginalType: OrderDeletedEventTypeNonClean,
						CleanType:    OrderDeletedEventType,
					},
				}),
				v1alpha2.WithBEBStatusFields(),
			),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//WHEN

			// convert the v1 to v2
			convertedV1Alpha2 := &v1alpha2.Subscription{}
			err := v1ToV2(testCase.alpha1Sub, convertedV1Alpha2)
			if err != nil && testCase.wantErrMsg != "" {
				require.Equal(t, err.Error(), testCase.wantErrMsg)
			} else {
				require.NoError(t, err)
				v1ToV2Assertions(t, testCase.alpha2Sub, convertedV1Alpha2)
			}

			// test ConvertFrom
			log.Print("Test v2 to v1 conversion")
			convertedV1Alpha1 := &Subscription{}
			require.NoError(t, v2ToV1(convertedV1Alpha1, testCase.alpha2Sub))
			if err != nil && testCase.wantErrMsg != "" {
				require.Equal(t, err.Error(), testCase.wantErrMsg)
			} else {
				require.NoError(t, err)
				v2ToV1Assertions(t, testCase.alpha1Sub, convertedV1Alpha1)
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

	// Status
	assert.Equal(t, wantSub.Status.Ready, convertedSub.Status.Ready)
	assert.Equal(t, wantSub.Status.Conditions, convertedSub.Status.Conditions)
	assert.Equal(t, wantSub.Status.Types, convertedSub.Status.Types)

	assert.Equal(t, wantSub.Status.Backend, convertedSub.Status.Backend)
}

func v2ToV1Assertions(t *testing.T, wantSub, convertedSub *Subscription) {
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
