package authn

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"k8s.io/apiserver/pkg/authentication/authenticator"

	"github.com/stretchr/testify/assert"

	"github.com/pkg/errors"
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
		t.Run("Then user.Info is added to the context", func(t *testing.T) {
			userInfoFromCtx, err := UserInfoForContext(next.r.Context())
			assert.Equal(t, &userInfo, userInfoFromCtx)
			assert.Nil(t, err)
		})
	})

	t.Run("When websocket request has malformed protocol header", func(t *testing.T) {
		authenticated := &mockAuthenticator{Authorised: true, UserInfo: &userInfo}
		middleware := AuthMiddleware(authenticated)
		next := &mockHandler{}
		response := httptest.NewRecorder()
		response.Code = 0
		middleware(next).ServeHTTP(response, newMalformedWebsocketRequest())

		t.Run("Must not call authorizer", func(t *testing.T) {
			assert.False(t, authenticated.Called)
		})
		t.Run("Then next handler is not called", func(t *testing.T) {
			assert.False(t, next.Called)
		})
		t.Run("Then request is rejected with status bad request", func(t *testing.T) {
			assert.Equal(t, http.StatusBadRequest, response.Code)
		})
	})

	t.Run("When Websocket request is unauthorised", func(t *testing.T) {
		reject := &mockAuthenticator{Authorised: false}
		middleware := AuthMiddleware(reject)
		next := &mockHandler{}
		response := httptest.NewRecorder()
		middleware(next).ServeHTTP(response, newWebsocketRequest())

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

	t.Run("When authentication error occurs on Websocket request", func(t *testing.T) {
		erroneous := &mockAuthenticator{Err: errors.New("failure")}
		middleware := AuthMiddleware(erroneous)
		next := &mockHandler{}
		response := httptest.NewRecorder()
		middleware(next).ServeHTTP(response, newWebsocketRequest())

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

	t.Run("When Websocket request is authenticated", func(t *testing.T) {
		authenticated := &mockAuthenticator{Authorised: true, UserInfo: &userInfo}
		middleware := AuthMiddleware(authenticated)
		next := &mockHandler{}
		response := httptest.NewRecorder()
		response.Code = 0
		middleware(next).ServeHTTP(response, newWebsocketRequest())

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
		t.Run("Then user.Info is added to the context", func(t *testing.T) {
			userInfoFromCtx, err := UserInfoForContext(next.r.Context())
			assert.Equal(t, &userInfo, userInfoFromCtx)
			assert.Nil(t, err)
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
	req := httptest.NewRequest("POST", "/graphql", strings.NewReader(""))
	req.Header.Set("authorization", "Bearer token")
	return req
}

func newMalformedWebsocketRequest() *http.Request {
	req := httptest.NewRequest("GET", "/graphql", strings.NewReader(""))
	req.Header.Set("sec-websocket-protocol", "graphql token")
	return req
}

func newWebsocketRequest() *http.Request {
	req := httptest.NewRequest("GET", "/graphql", strings.NewReader(""))
	req.Header.Set("sec-websocket-protocol", "graphql, token")
	return req
}
