package v1alpha1

import (
	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	testing2 "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"testing"
)

func Test_ConvertTo(t *testing.T) {
	testSink := "https://svc2.test.local"
	eventSource := "commerceMock"
	testCases := []struct {
		name     string
		givenSub Subscription
		wantSub  v1alpha2.Subscription
	}{
		{
			name: "test1",
			givenSub: *testing2.NewSubscription("sub1", "test",
				testing2.WithSink(testSink),
				testing2.WithFilter(eventSource, testing2.NewOrderCreatedEventType),
				testing2.WithFilter(eventSource, testing2.NewOrderCreatedEventType),
			),
			wantSub: v1alpha2.Subscription{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//WHEN
			//var dst Subscription
			//testCase.givenSub.ConvertTo(dst)
		})
	}

}
func Test_ConvertFrom(t *testing.T) {

	testCases := []struct {
		name     string
		givenSub Subscription
		wantSub  Subscription
	}{
		{
			name:     "test1",
			givenSub: Subscription{},
			wantSub:  Subscription{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//WHEN
			//var dst Subscription
			//testCase.givenSub.ConvertTo(dst)
		})
	}
}
