package authn

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"errors"

	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
)

func TestAuthMiddleware(t *testing.T) {

	userInfo := user.DefaultInfo{Name: "Test User", UID: "deadbeef", Groups: []string{"admins", "testers"}}

	t.Run("When HTTP request is unauthorised", func(t *testing.T) {
		reject := &mockAuthenticator{Authorised: false}
		middleware := AuthMiddleware(reject)
		next := &mockHandler{}
		response := httptest.NewRecorder()
		middleware(next).ServeHTTP(response, newHttpRequest())

		t.Run("Then authorizer is called with token", func(t *testing.T) {
			assert.True(t, reject.Called)
			assert.Equal(t, "Bearer token", reject.LastReq.Header.Get("authorization"))
		})
		t.Run("Then next handler is not called", func(t *testing.T) {
			assert.False(t, next.Called)
		})
		t.Run("Then request is rejected with status code unauthorised", func(t *testing.T) {
			assert.Equal(t, http.StatusUnauthorized, response.Code)
		})
	})

	t.Run("When authentication error occurs on HTTP request", func(t *testing.T) {
		erroneous := &mockAuthenticator{Err: errors.New("failure")}
		middleware := AuthMiddleware(erroneous)
		next := &mockHandler{}
		response := httptest.NewRecorder()
		middleware(next).ServeHTTP(response, newHttpRequest())

		t.Run("Then authorizer is called with token", func(t *testing.T) {
			assert.True(t, erroneous.Called)
			assert.Equal(t, "Bearer token", erroneous.LastReq.Header.Get("authorization"))
		})
		t.Run("Then next handler is not called", func(t *testing.T) {
			assert.False(t, next.Called)
		})
		t.Run("Then request is rejected with status code unauthorised", func(t *testing.T) {
			assert.Equal(t, http.StatusUnauthorized, response.Code)
		})
	})

	t.Run("When HTTP request is authenticated", func(t *testing.T) {
		authenticated := &mockAuthenticator{Authorised: true, UserInfo: &userInfo}
		middleware := AuthMiddleware(authenticated)
		next := &mockHandler{}
		response := httptest.NewRecorder()
		response.Code = 0
		middleware(next).ServeHTTP(response, newHttpRequest())

		t.Run("Then authorizer is called with token", func(t *testing.T) {
			assert.True(t, authenticated.Called)
			assert.Equal(t, "Bearer token", authenticated.LastReq.Header.Get("authorization"))
		})
		t.Run("Then next handler is called", func(t *testing.T) {
			assert.True(t, next.Called)
		})
		t.Run("Then status code is not set", func(t *testing.T) {
			assert.Equal(t, 0, response.Code)
		})
		t.Run("Then authorization header is present in the request", func(t *testing.T) {
			assert.Equal(t, "Bearer token", next.r.Header.Get("Authorization"))
		})
	})
}

type mockAuthenticator struct {
	UserInfo   user.Info
	Authorised bool
	Err        error
	LastReq    *http.Request
	Called     bool
}

func (a *mockAuthenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	a.Called = true
	a.LastReq = req

	//Mimic behaviour of k8s.io/apiserver/pkg/authentication/request/bearertoken.Authenticator.AuthenticateRequest
	if a.Authorised {
		req.Header.Del("Authorization")
	}

	return &authenticator.Response{User: a.UserInfo}, a.Authorised, a.Err
}

type mockHandler struct {
	Called bool
	w      http.ResponseWriter
	r      *http.Request
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.w = w
	h.r = r
	h.Called = true
}

func newHttpRequest() *http.Request {
	req := httptest.NewRequest("POST", "/kube-config", strings.NewReader(""))
	req.Header.Set("authorization", "Bearer token")
	return req
}
