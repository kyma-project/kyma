package authn

import (
	"net/http"
	"testing"

	"k8s.io/apiserver/pkg/authentication/authenticator"
)

type fakeAuthenticator struct {
	authenticated bool
	err           error
}

func (f *fakeAuthenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	return nil, f.authenticated, f.err
}

func TestProxyAuthenticator_AuthenticateRequest(t *testing.T) {
	var testCases = []struct {
		authenticated  bool
		authenticators []authenticator.Request
	}{
		{true,
			[]authenticator.Request{&fakeAuthenticator{
				authenticated: true,
				err:           nil,
			},
				&fakeAuthenticator{
					authenticated: false,
					err:           nil,
				},
			},
		},
		{true,
			[]authenticator.Request{&fakeAuthenticator{
				authenticated: false,
				err:           nil,
			},
				&fakeAuthenticator{
					authenticated: true,
					err:           nil,
				},
			},
		},
		{
			authenticated: false,
			authenticators: []authenticator.Request{&fakeAuthenticator{
				authenticated: false,
				err:           nil,
			},
				&fakeAuthenticator{
					authenticated: false,
					err:           nil,
				},
			},
		},
	}

	for _, tc := range testCases {
		req, _ := http.NewRequest("GET", "http://whatever.com", nil)
		proxyAuthenticator := New(tc.authenticators...)

		_, authenticated, err := proxyAuthenticator.AuthenticateRequest(req)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if tc.authenticated != authenticated {
			t.Errorf("Unexpected authentication result. Request authenticated : %v but expected to be : %v", authenticated, tc.authenticated)
		}
	}
}
