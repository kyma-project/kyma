package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/components/apiserver-proxy/internal/authn"
	"github.com/kyma-project/kyma/components/apiserver-proxy/internal/authz"
	"github.com/kyma-project/kyma/components/apiserver-proxy/internal/monitoring"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/request/bearertoken"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

func TestProxyWithOIDCSupport(t *testing.T) {
	cfg := Config{
		Authentication: &authn.AuthnConfig{
			OIDC: &authn.OIDCConfig{},
			Header: &authn.AuthnHeaderConfig{
				Enabled:         true,
				UserFieldName:   "user",
				GroupsFieldName: "groups",
			},
		},
		Authorization: &authz.Config{},
	}

	fakeUser := user.DefaultInfo{Name: "Foo Bar", Groups: []string{"foo-bars"}}
	authenticator := fakeOIDCAuthenticator(t, &fakeUser)
	metrics, _ := monitoring.NewProxyMetrics()

	scenario := setupTestScenario()
	for _, v := range scenario {

		t.Run(v.description, func(t *testing.T) {

			w := httptest.NewRecorder()
			proxy := New(cfg, v.authorizer, authenticator, metrics)

			proxy.Handle(w, v.req)

			resp := w.Result()

			if resp.StatusCode != v.status {
				t.Errorf("Expected response: %d received : %d", v.status, resp.StatusCode)
			}

			if v.verifyUser {
				user := v.req.Header.Get(cfg.Authentication.Header.UserFieldName)
				groups := v.req.Header.Get(cfg.Authentication.Header.GroupsFieldName)
				if user != fakeUser.GetName() {
					t.Errorf("User in the response header does not match authenticated user. Expected : %s, received : %s ", fakeUser.GetName(), user)
				}
				if groups != strings.Join(fakeUser.GetGroups(), cfg.Authentication.Header.GroupSeparator) {
					t.Errorf("Groupsr in the response header does not match authenticated user groups. Expected : %s, received : %s ", fakeUser.GetName(), groups)
				}
			}
		})
	}
}

func TestRequestResolver(t *testing.T) {

	rir := newRequestInfoResolver()
	cases := []struct {
		method            string
		path              string
		isResourceRequest bool
		resource          string
		apiVersion        string
		namespace         string
	}{
		{"GET", "/api/v1/namespaces", true, "namespaces", "v1", ""},
		{"GET", "/apis/authentication/v1/namespaces/default/groups", true, "groups", "v1", "default"},
		{"GET", "/openapi/v2/", false, "", "", ""},
	}

	for _, c := range cases {

		t.Run(c.path, func(t *testing.T) {
			req := requestFor(c.method, c.path)

			info, err := rir.NewRequestInfo(req)

			if err != nil {
				t.Fatalf("Resolving request info failed : %s", err.Error())
			}

			if info == nil {
				t.Fatalf("Fore request path %s request info object is nil", c.path)
			}

			if info.Resource != c.resource {
				t.Fatalf("Resource does not match. Expected %s actual %s ", c.resource, info.Resource)
			}

			if info.APIVersion != c.apiVersion {
				t.Fatalf("APIVersion does not match. Expected %s actual %s ", c.apiVersion, info.APIVersion)
			}

			if info.IsResourceRequest != c.isResourceRequest {
				t.Fatalf("IsResourceRequest does not match. Expected %t actual %t ", c.isResourceRequest, info.IsResourceRequest)
			}
		})
	}
}

func setupTestScenario() []testCase {
	testScenario := []testCase{
		{
			description: "Request with invalid Token should be authenticated and rejected with 401",
			given: given{
				req:        fakeJWTRequest("GET", "/accounts", "Bearer INVALID"),
				authorizer: denier{},
			},
			expected: expected{
				status: http.StatusUnauthorized,
			},
		},
		{
			description: "Request with valid token, should return 200 due to sufficient permissions",
			given: given{
				req:        fakeJWTRequest("GET", "/accounts", "Bearer VALID"),
				authorizer: approver{},
			},
			expected: expected{
				status:     http.StatusOK,
				verifyUser: true,
			},
		},
	}
	return testScenario
}

func fakeJWTRequest(method, path, token string) *http.Request {
	req := requestFor(method, path)
	req.Header.Add("Authorization", token)

	return req
}

func requestFor(method, path string) *http.Request {
	req := httptest.NewRequest(method, path, nil)

	return req
}

func fakeOIDCAuthenticator(t *testing.T, fakeUser *user.DefaultInfo) authenticator.Request {

	auth := bearertoken.New(authenticator.TokenFunc(func(ctx context.Context, token string) (*authenticator.Response, bool, error) {
		if token != "VALID" {
			return nil, false, nil
		}
		return &authenticator.Response{User: fakeUser}, true, nil
	}))
	return auth
}

type denier struct{}

func (d denier) Authorize(auth authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {
	return authorizer.DecisionDeny, "user not allowed", nil
}

type approver struct{}

func (a approver) Authorize(auth authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {
	return authorizer.DecisionAllow, "user allowed", nil
}

type given struct {
	req        *http.Request
	authorizer authorizer.Authorizer
}

type expected struct {
	status     int
	verifyUser bool
}

type testCase struct {
	given
	expected
	description string
}
