// +build unit

package utils_test

import (
	"net/url"
	"reflect"
	"testing"

	utils "github.com/kyma-project/kyma/components/eventing-controller/utils"
)

func Test_GetPortNumberFromURL(t *testing.T) {
	testCases := []struct {
		givenURL url.URL
		wantPort uint32
	}{
		{
			givenURL: url.URL{},
			wantPort: 80,
		},
		{
			givenURL: url.URL{
				Host: "domain.com:####",
			},
			wantPort: 80,
		},
		{
			givenURL: url.URL{
				Host: "domain.com",
			},
			wantPort: 80,
		},
		{
			givenURL: url.URL{
				Scheme: "http",
			},
			wantPort: 80,
		},
		{
			givenURL: url.URL{
				Scheme: "https",
			},
			wantPort: 443,
		},
		{
			givenURL: url.URL{
				Scheme: "http",
				Host:   "domain.com:8080",
			},
			wantPort: 8080,
		},
		{
			givenURL: url.URL{
				Scheme: "https",
				Host:   "domain.com:8081",
			},
			wantPort: 8081,
		},
	}

	for _, tc := range testCases {
		gotPort, err := utils.GetPortNumberFromURL(tc.givenURL)
		if err != nil {
			t.Errorf("test failed with error: [%v]", err)
			continue
		}
		if tc.wantPort != gotPort {
			t.Errorf("test failed with given URL: {Scheme:%s Host:%s}, want port: [%d] but got: [%d]",
				tc.givenURL.Scheme, tc.givenURL.Host, tc.wantPort, gotPort)
		}
	}
}

func Test_ContainsString(t *testing.T) {
	testCases := []struct {
		sl   []string
		s    string
		want bool
	}{
		{
			sl:   []string{"kyma", "eventing"},
			s:    "kyma",
			want: true,
		},
		{
			sl:   []string{"kyma", "eventing"},
			s:    "nats",
			want: false,
		},
		{
			sl:   []string{"kyma.eventing", "nats"},
			s:    "kyma",
			want: false,
		},
	}
	for _, tc := range testCases {
		result := utils.ContainsString(tc.sl, tc.s)
		if tc.want != result {
			t.Errorf("test failed with give slice of strings: %s and string: %s, expected: %v but got: %v",
				tc.sl, tc.s, tc.want, result)
		}
	}
}

func Test_RemoveString(t *testing.T) {
	testCases := []struct {
		sl   []string
		s    string
		want []string
	}{
		{
			sl:   []string{"kyma", "eventing"},
			s:    "kyma",
			want: []string{"eventing"},
		},
		{
			sl:   []string{"kyma", "eventing"},
			s:    "nats",
			want: []string{"kyma", "eventing"},
		},
		{
			sl:   []string{"kyma.eventing", "nats"},
			s:    "kyma",
			want: []string{"kyma.eventing", "nats"},
		},
	}
	for _, tc := range testCases {
		result := utils.RemoveString(tc.sl, tc.s)
		if !reflect.DeepEqual(tc.want, result) {
			t.Errorf("test failed with give slice of strings: %s and string: %s, expected: %s but got: %s",
				tc.sl, tc.s, tc.want, result)
		}
	}
}
