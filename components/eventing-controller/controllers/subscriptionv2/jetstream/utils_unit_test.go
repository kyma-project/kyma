package jetstream

import (
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	subtesting "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/stretchr/testify/require"
)

func Test_isInDeletion(t *testing.T) {
	testCases := []struct {
		name       string
		givenSub   *v1alpha2.Subscription
		wantResult bool
	}{
		{
			name:       "subscription with no deletion timestamp",
			givenSub:   subtesting.NewSubscription("test", "test"),
			wantResult: false,
		},
		{
			name: "subscription with deletion timestamp",
			givenSub: subtesting.NewSubscription("test", "test",
				subtesting.WithNonZeroDeletionTimestamp()),
			wantResult: true,
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// when
			result := isInDeletion(tc.givenSub)

			// then
			require.Equal(t, tc.wantResult, result)
		})
	}
}

func Test_containsFinalizer(t *testing.T) {
	testCases := []struct {
		name       string
		givenSub   *v1alpha2.Subscription
		wantResult bool
	}{
		{
			name: "subscription containing finalizer",
			givenSub: subtesting.NewSubscription("test", "test",
				subtesting.WithFinalizers([]string{v1alpha2.Finalizer})),
			wantResult: true,
		},
		{
			name: "subscription containing finalizer",
			givenSub: subtesting.NewSubscription("test", "test",
				subtesting.WithFinalizers([]string{"invalid"})),
			wantResult: false,
		},
		{
			name:       "subscription not containing finalizer",
			givenSub:   subtesting.NewSubscription("test", "test"),
			wantResult: false,
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// when
			result := containsFinalizer(tc.givenSub)

			// then
			require.Equal(t, tc.wantResult, result)
		})
	}
}
