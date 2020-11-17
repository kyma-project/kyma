package utils

import (
	"net/url"
	"testing"
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
		gotPort, err := GetPortNumberFromURL(tc.givenURL)
		if err != nil {
			t.Errorf("Test failed with error: [%v]", err)
			continue
		}
		if tc.wantPort != gotPort {
			t.Errorf("Test failed with given URL: {Scheme:%s Host:%s}, want port: [%d] but got: [%d]",
				tc.givenURL.Scheme, tc.givenURL.Host, tc.wantPort, gotPort)
		}
	}
}
