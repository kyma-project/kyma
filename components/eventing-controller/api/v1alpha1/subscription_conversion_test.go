package v1alpha1

import (
	"testing"
)

func Test_ConvertTo(t *testing.T) {

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
