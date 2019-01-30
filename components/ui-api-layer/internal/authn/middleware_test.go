package authn

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/apiserver/pkg/authentication/user"
)

func TestAuthMiddleware(t *testing.T) {

	userInfo := user.DefaultInfo{Name: "Test User", UID: "deadbeef", Groups: []string{"admins", "testers"}}

	Convey("When HTTP request is unauthorised", t, func() {

		reject := &mockAuthenticator{Authorised: false}
		middleware := AuthMiddleware(reject)
		next := &mockHandler{}
		response := httptest.NewRecorder()
		middleware(next).ServeHTTP(response, newHttpRequest())

		Convey("Then authorizer is called with token", func() {
			So(reject.Called, ShouldBeTrue)
			So(reject.LastReq.Header.Get("authorization"), ShouldEqual, "Bearer token")
		})
		Convey("Then next handler is not called", func() {
			So(next.Called, ShouldBeFalse)
		})
		Convey("Then request is rejected with status code unauthorised", func() {
			So(response.Code, ShouldEqual, http.StatusUnauthorized)
		})
	})

	Convey("When authentication error occurs on HTTP request", t, func() {

		erroneous := &mockAuthenticator{Err: errors.New("failure")}
		middleware := AuthMiddleware(erroneous)
		next := &mockHandler{}
		response := httptest.NewRecorder()
		middleware(next).ServeHTTP(response, newHttpRequest())

		Convey("Then authorizer is called with token", func() {
			So(erroneous.Called, ShouldBeTrue)
			So(erroneous.LastReq.Header.Get("authorization"), ShouldEqual, "Bearer token")
		})
		Convey("Then next handler is not called", func() {
			So(next.Called, ShouldBeFalse)
		})
		Convey("Then request is rejected with status code unauthorised", func() {
			So(response.Code, ShouldEqual, http.StatusUnauthorized)
		})
	})

	Convey("When HTTP request is authenticated", t, func() {

		authenticated := &mockAuthenticator{Authorised: true, UserInfo: &userInfo}
		middleware := AuthMiddleware(authenticated)
		next := &mockHandler{}
		response := httptest.NewRecorder()
		response.Code = 0
		middleware(next).ServeHTTP(response, newHttpRequest())

		Convey("Then authorizer is called with token", func() {
			So(authenticated.Called, ShouldBeTrue)
			So(authenticated.LastReq.Header.Get("authorization"), ShouldEqual, "Bearer token")
		})
		Convey("Then next handler is called", func() {
			So(next.Called, ShouldBeTrue)
		})
		Convey("Then status code is not set", func() {
			So(response.Code, ShouldEqual, 0)
		})
		Convey("Then user.Info is added to the context", func() {
			userInfoFromCtx, err := UserInfoForContext(next.r.Context())
			So(userInfoFromCtx, ShouldEqual, &userInfo)
			So(err, ShouldBeNil)
		})
	})

	Convey("When websocket request has malformed protocol header", t, func() {

		authenticated := &mockAuthenticator{Authorised: true, UserInfo: &userInfo}
		middleware := AuthMiddleware(authenticated)
		next := &mockHandler{}
		response := httptest.NewRecorder()
		response.Code = 0
		middleware(next).ServeHTTP(response, newMalformedWebsocketRequest())

		Convey("Must not call authorizer", func() {
			So(authenticated.Called, ShouldBeFalse)
		})
		Convey("Then next handler is not called", func() {
			So(next.Called, ShouldBeFalse)
		})
		Convey("Then request is rejected with status bad request", func() {
			So(response.Code, ShouldEqual, http.StatusBadRequest)
		})
	})

	Convey("When Websocket request is unauthorised", t, func() {

		reject := &mockAuthenticator{Authorised: false}
		middleware := AuthMiddleware(reject)
		next := &mockHandler{}
		response := httptest.NewRecorder()
		middleware(next).ServeHTTP(response, newWebsocketRequest())

		Convey("Then authorizer is called with token", func() {
			So(reject.Called, ShouldBeTrue)
			So(reject.LastReq.Header.Get("authorization"), ShouldEqual, "Bearer token")
		})
		Convey("Then next handler is not called", func() {
			So(next.Called, ShouldBeFalse)
		})
		Convey("Then request is rejected with status code unauthorised", func() {
			So(response.Code, ShouldEqual, http.StatusUnauthorized)
		})
	})

	Convey("When authentication error occurs on Websocket request", t, func() {

		erroneous := &mockAuthenticator{Err: errors.New("failure")}
		middleware := AuthMiddleware(erroneous)
		next := &mockHandler{}
		response := httptest.NewRecorder()
		middleware(next).ServeHTTP(response, newWebsocketRequest())

		Convey("Then authorizer is called with token", func() {
			So(erroneous.Called, ShouldBeTrue)
			So(erroneous.LastReq.Header.Get("authorization"), ShouldEqual, "Bearer token")
		})
		Convey("Then next handler is not called", func() {
			So(next.Called, ShouldBeFalse)
		})
		Convey("Then request is rejected with status code unauthorised", func() {
			So(response.Code, ShouldEqual, http.StatusUnauthorized)
		})
	})

	Convey("When Websocket request is authenticated", t, func() {

		authenticated := &mockAuthenticator{Authorised: true, UserInfo: &userInfo}
		middleware := AuthMiddleware(authenticated)
		next := &mockHandler{}
		response := httptest.NewRecorder()
		response.Code = 0
		middleware(next).ServeHTTP(response, newWebsocketRequest())

		Convey("Then authorizer is called with token", func() {
			So(authenticated.Called, ShouldBeTrue)
			So(authenticated.LastReq.Header.Get("authorization"), ShouldEqual, "Bearer token")
		})
		Convey("Then next handler is called", func() {
			So(next.Called, ShouldBeTrue)
		})
		Convey("Then status code is not set", func() {
			So(response.Code, ShouldEqual, 0)
		})
		Convey("Then user.Info is added to the context", func() {
			userInfoFromCtx, err := UserInfoForContext(next.r.Context())
			So(userInfoFromCtx, ShouldEqual, &userInfo)
			So(err, ShouldBeNil)
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

func (a *mockAuthenticator) AuthenticateRequest(req *http.Request) (user.Info, bool, error) {
	a.Called = true
	a.LastReq = req
	return a.UserInfo, a.Authorised, a.Err
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
