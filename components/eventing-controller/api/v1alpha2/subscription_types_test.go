package v1alpha2 //nolint:nolintlint,testpackage

import (
	"strconv"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/stretchr/testify/assert"
)

func TestGetMaxInFlightMessages(t *testing.T) {
	testCases := []struct {
		name              string
		givenSubscription *Subscription
		wantResult        int
		wantErr           error
	}{
		{
			name: "function should give the default MaxInFlight if the Subscription config is missing",
			givenSubscription: &Subscription{
				Spec: SubscriptionSpec{
					Config: nil,
				},
			},
			wantResult: env.DefaultMaxInFlight,
			wantErr:    nil,
		},
		{
			name: "function should give the default MaxInFlight if the it is missing in the Subscription config",
			givenSubscription: &Subscription{
				Spec: SubscriptionSpec{
					Config: map[string]string{
						"otherConfigKey": "20"},
				},
			},
			wantResult: env.DefaultMaxInFlight,
			wantErr:    nil,
		},
		{
			name: "function should give the expectedConfig",
			givenSubscription: &Subscription{
				Spec: SubscriptionSpec{
					Config: map[string]string{
						MaxInFlightMessages: "20"},
				},
			},
			wantResult: 20,
			wantErr:    nil,
		},
		{
			name: "function should result into an error",
			givenSubscription: &Subscription{
				Spec: SubscriptionSpec{
					Config: map[string]string{
						MaxInFlightMessages: "nonInt"},
				},
			},
			wantResult: -1,
			wantErr:    &strconv.NumError{Func: "Atoi", Num: "nonInt", Err: strconv.ErrSyntax},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.givenSubscription.GetMaxInFlightMessages()

			assert.Equal(t, tc.wantResult, result)
			assert.Equal(t, err, tc.wantErr)
		})
	}
}
