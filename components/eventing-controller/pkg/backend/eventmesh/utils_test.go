package eventmesh

import (
	"fmt"
	"testing"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	eventingtestingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/stretchr/testify/require"
)

func Test_IsEventTypeSegmentsOverLimit(t *testing.T) {
	// cases
	testCases := []struct {
		name           string
		givenEventType string
		wantResult     bool
	}{
		{
			name:           "success if the given event type does not exceeds limit",
			givenEventType: "one.two.three.four.five.six.seven",
			wantResult:     false,
		},
		{
			name:           "fail if the given event type exceeds limit",
			givenEventType: "one.two.three.four.five.six.seven.eight",
			wantResult:     true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.wantResult, IsEventTypeSegmentsOverLimit(tc.givenEventType))
		})
	}

}

func Test_GetEventMeshSubject(t *testing.T) {
	// cases
	testCases := []struct {
		name                 string
		givenSource          string
		givenSubject         string
		givenEventMeshPrefix string
		wantEventMeshSubject string
	}{
		{
			name:                 "success if the segments are in correct order",
			givenSource:          "test1",
			givenSubject:         "one.two.three",
			givenEventMeshPrefix: eventingtestingv2.EventMeshPrefix,
			wantEventMeshSubject: fmt.Sprintf("%s.test1.one.two.three", eventingtestingv2.EventMeshPrefix),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.wantEventMeshSubject, GetEventMeshSubject(tc.givenSource, tc.givenSubject, tc.givenEventMeshPrefix))
		})
	}
}

func Test_statusCleanEventTypes(t *testing.T) {
	// cases
	testCases := []struct {
		name           string
		givenTypeInfos []backendutils.EventTypeInfo
		wantEventTypes []eventingv1alpha2.EventType
	}{
		{
			name: "success if the EventTypes are correct",
			givenTypeInfos: []backendutils.EventTypeInfo{
				{
					OriginalType:  "test1",
					CleanType:     "test2",
					ProcessedType: "test3",
				},
				{
					OriginalType:  "order1",
					CleanType:     "order2",
					ProcessedType: "order3",
				},
			},
			wantEventTypes: []eventingv1alpha2.EventType{
				{
					OriginalType: "test1",
					CleanType:    "test2",
				},
				{
					OriginalType: "order1",
					CleanType:    "order2",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.wantEventTypes, statusCleanEventTypes(tc.givenTypeInfos))
		})
	}
}

func Test_setEventMeshServerSubHashInStatus(t *testing.T) {
	t.Parallel()

	// given
	eventMeshSubscription := eventingtestingv2.NewSampleEventMeshSubscription()
	kymaSubscription := eventingtestingv2.NewSubscription("test", "test")
	wantHash, err := backendutils.GetHash(eventMeshSubscription)
	require.NoError(t, err)

	// when
	err = setEventMeshServerSubHashInStatus(kymaSubscription, eventMeshSubscription)

	// then
	require.NoError(t, err)
	require.Equal(t, kymaSubscription.Status.Backend.Emshash, wantHash)
}

func Test_setEventMeshLocalSubHashInStatus(t *testing.T) {
	t.Parallel()

	// given
	eventMeshSubscription := eventingtestingv2.NewSampleEventMeshSubscription()
	kymaSubscription := eventingtestingv2.NewSubscription("test", "test")
	wantHash, err := backendutils.GetHash(eventMeshSubscription)
	require.NoError(t, err)

	// when
	err = setEventMeshLocalSubHashInStatus(kymaSubscription, eventMeshSubscription)

	// then
	require.NoError(t, err)
	require.Equal(t, kymaSubscription.Status.Backend.Ev2hash, wantHash)
}

func Test_updateHashesInStatus(t *testing.T) {
	t.Parallel()

	// given
	eventMeshSubscription := eventingtestingv2.NewSampleEventMeshSubscription()
	kymaSubscription := eventingtestingv2.NewSubscription("test", "test")
	wantHash, err := backendutils.GetHash(eventMeshSubscription)
	require.NoError(t, err)

	// when
	err = updateHashesInStatus(kymaSubscription, eventMeshSubscription, eventMeshSubscription)

	// then
	require.NoError(t, err)
	require.Equal(t, kymaSubscription.Status.Backend.Ev2hash, wantHash)
	require.Equal(t, kymaSubscription.Status.Backend.Emshash, wantHash)
}

func Test_setEmsSubscriptionStatus(t *testing.T) {
	t.Parallel()

	// given
	eventMeshSubscription := eventingtestingv2.NewSampleEventMeshSubscription()
	eventMeshSubscription.SubscriptionStatus = "ready"
	eventMeshSubscription.SubscriptionStatusReason = "unknown"
	eventMeshSubscription.LastSuccessfulDelivery = "09:00"
	eventMeshSubscription.LastFailedDelivery = "00:00"
	eventMeshSubscription.LastFailedDeliveryReason = "failed"

	kymaSubscription := eventingtestingv2.NewSubscription("test", "test")

	// when
	isChanged := setEmsSubscriptionStatus(kymaSubscription, eventMeshSubscription)

	// then
	require.True(t, isChanged)
	require.NotNil(t, kymaSubscription.Status.Backend.EmsSubscriptionStatus)
	require.Equal(t, kymaSubscription.Status.Backend.EmsSubscriptionStatus.Status, string(eventMeshSubscription.SubscriptionStatus))
	require.Equal(t, kymaSubscription.Status.Backend.EmsSubscriptionStatus.StatusReason, eventMeshSubscription.SubscriptionStatusReason)
	require.Equal(t, kymaSubscription.Status.Backend.EmsSubscriptionStatus.LastSuccessfulDelivery, eventMeshSubscription.LastSuccessfulDelivery)
	require.Equal(t, kymaSubscription.Status.Backend.EmsSubscriptionStatus.LastFailedDelivery, eventMeshSubscription.LastFailedDelivery)
	require.Equal(t, kymaSubscription.Status.Backend.EmsSubscriptionStatus.LastFailedDeliveryReason, eventMeshSubscription.LastFailedDeliveryReason)
}
