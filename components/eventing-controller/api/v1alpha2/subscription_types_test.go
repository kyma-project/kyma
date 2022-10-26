package v1alpha2_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/stretchr/testify/assert"
)

func TestGetMaxInFlightMessages(t *testing.T) {
	defaultSubConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	testCases := []struct {
		name              string
		givenSubscription *v1alpha2.Subscription
		wantResult        int
	}{
		{
			name: "function should give the default MaxInFlight if the Subscription config is missing",
			givenSubscription: &v1alpha2.Subscription{
				Spec: v1alpha2.SubscriptionSpec{
					Config: nil,
				},
			},
			wantResult: defaultSubConfig.MaxInFlightMessages,
		},
		{
			name: "function should give the default MaxInFlight if it is missing in the Subscription config",
			givenSubscription: &v1alpha2.Subscription{
				Spec: v1alpha2.SubscriptionSpec{
					Config: map[string]string{
						"otherConfigKey": "20"},
				},
			},
			wantResult: defaultSubConfig.MaxInFlightMessages,
		},
		{
			name: "function should give the expectedConfig",
			givenSubscription: &v1alpha2.Subscription{
				Spec: v1alpha2.SubscriptionSpec{
					Config: map[string]string{
						v1alpha2.MaxInFlightMessages: "20"},
				},
			},
			wantResult: 20,
		},
		{
			name: "function should result into an error",
			givenSubscription: &v1alpha2.Subscription{
				Spec: v1alpha2.SubscriptionSpec{
					Config: map[string]string{
						v1alpha2.MaxInFlightMessages: "nonInt"},
				},
			},
			wantResult: defaultSubConfig.MaxInFlightMessages,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			result := tc.givenSubscription.GetMaxInFlightMessages(&defaultSubConfig)

			assert.Equal(t, tc.wantResult, result)
		})
	}
}
