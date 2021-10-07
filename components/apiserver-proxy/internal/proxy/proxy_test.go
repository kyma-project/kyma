package proxy

import (
	"context"
	"fmt"
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

	maliciousGroup := "malicious-group"
	maliciousUser := "malicious-user"
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
				if strings.Contains(v.req.Header.Get("Impersonate-Group"), maliciousGroup) {
					t.Errorf("Groups should not contain %s injected in the request", maliciousGroup)
				}
				if strings.Contains(v.req.Header.Get("Impersonate-User"), maliciousUser) {
					t.Errorf("User should not contain %s injected in the request", maliciousUser)
				}
				if strings.Contains(v.req.Header.Get("Connection"), "Impersonate-Group") {
					t.Errorf("Connection header should not contain Impersonate-Group key")
				}
				if strings.Contains(v.req.Header.Get("Connection"), "Impersonate-User") {
					t.Errorf("Connection header should not contain Impersonate-User key")
				}

				if groups != strings.Join(fakeUser.GetGroups(), cfg.Authentication.Header.GroupSeparator) {
					t.Errorf("Groups in the response header does not match authenticated user groups. Expected : %s, received : %s ", fakeUser.GetGroups(), groups)
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
	header := http.Header{}
	const maliciousGroup = "malicious-group"
	const maliciousUser = "malicious-user"
	testScenario := []testCase{

		{
			description: "Request with invalid Token should be authenticated and rejected with 401",
			given: given{
				req:        fakeJWTRequest("GET", "/accounts", "Bearer INVALID", header),
				authorizer: denier{},
			},
			expected: expected{
				status: http.StatusUnauthorized,
			},
		},
		{
			description: "Request with valid Token should be authenticated and accepted with 200",
			given: given{
				req:        fakeJWTRequest("GET", "/accounts", "Bearer VALID", header),
				authorizer: denier{},
			},
			expected: expected{
				status:     http.StatusOK,
				verifyUser: true,
			},
		},
		{
			description: "Request with valid token, and malicious user/group in impersonate headers should return 200 due to sufficient permissions",
			given: given{
				req:        fakeJWTRequest("GET", "/accounts", "Bearer VALID", addHeaderImpersonateGroupAndUser(header, maliciousUser, maliciousGroup)),
				authorizer: approver{},
			},
			expected: expected{
				status:     http.StatusOK,
				verifyUser: true,
			},
		},
		{
			description: "Request with valid token, and two malicious groups in Connection header should return 200 due to sufficient permissions",
			given: given{
				req:        fakeJWTRequest("GET", "/accounts", "Bearer VALID", addConnectionHeaderImpersonateGroupValueTwice(header, maliciousGroup, maliciousGroup)),
				authorizer: approver{},
			},
			expected: expected{
				status:     http.StatusOK,
				verifyUser: true,
			},
		},
		{
			description: "Request with valid token, and impersonate group/user key in Connection header should return 200 due to sufficient permissions",
			given: given{
				req:        fakeJWTRequest("GET", "/accounts", "Bearer VALID", addConnectionHeaderImpersonateGroupImpersonateUser(header)),
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

func addHeaderImpersonateGroupAndUser(header http.Header, maliciousUser string, maliciousGroup string) http.Header {
	return addHeaderImpersonateGroup(addHeaderImpersonateUser(header, maliciousUser), maliciousGroup)
}

func fakeJWTRequest(method, path, token string, extraHeaders http.Header) *http.Request {
	req := requestFor(method, path)
	req.Header.Add("Authorization", token)
	for key, values := range extraHeaders {
		for _, v := range values {
			req.Header.Add(key, v)
		}
	}
	return req
}

func addConnectionHeaderImpersonateGroupValueTwice(header http.Header, group1 string, group2 string) http.Header {
	header.Add("Connection", fmt.Sprintf("Impersonate-Group=%s,Impersonate-Group=%s", group1, group2))
	return header
}

func addConnectionHeaderImpersonateGroupImpersonateUser(header http.Header) http.Header {
	header.Add("Connection", "Impersonate-Group,Impersonate-User")
	return header
}

func addHeaderImpersonateGroup(h http.Header, group string) http.Header {
	h.Add("Impersonate-Group", group)
	return h
}

func addHeaderImpersonateUser(h http.Header, user string) http.Header {
	h.Add("Impersonate-Group", user)
	return h
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

func (d denier) Authorize(ctx context.Context, auth authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {
	return authorizer.DecisionDeny, "user not allowed", nil
}

type approver struct{}

func (a approver) Authorize(ctx context.Context, auth authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {
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
