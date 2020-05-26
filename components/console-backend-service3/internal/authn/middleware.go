package authn

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"strings"

	"github.com/golang/glog"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
)

var userInfoCtxKey = &contextKey{"userInfo"}

type contextKey struct {
	name string
}

func AuthMiddleware(a authenticator.Request) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := r.Body.Close(); err != nil {
					glog.Errorf("Closing body failed due to an error: %s", err)
				}
			}()
			wsProtocolHeader := r.Header.Get("sec-websocket-protocol")
			if wsProtocolHeader != "" {
				wsProtocolParts := strings.Split(wsProtocolHeader, ",")
				if len(wsProtocolParts) != 2 {
					http.Error(w, "sec-websocket-protocol malformed", http.StatusBadRequest)
					return
				}
				wsProtocol, wsToken := strings.TrimSpace(wsProtocolParts[0]), strings.TrimSpace(wsProtocolParts[1])
				r.Header.Set("Authorization", "Bearer "+wsToken)
				r.Header.Set("sec-websocket-protocol", wsProtocol)
			}

			u, ok, err := a.AuthenticateRequest(r)
			if err != nil {
				glog.Errorf("Unable to authenticate the request due to an error: %v", err)
			}
			if !ok || err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := WithUserInfoContext(r.Context(), u.User)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func UserInfoForContext(ctx context.Context) (user.Info, error) {
	raw := ctx.Value(userInfoCtxKey)
	if raw == nil {
		return &user.DefaultInfo{}, errors.New("Unable to find user info in request context")
	}
	userInfo, ok := raw.(user.Info)
	if !ok {
		return &user.DefaultInfo{}, errors.New("User info from request context does not comply with user.Info interface")
	}
	return userInfo, nil
}

func WithUserInfoContext(ctx context.Context, userInfo user.Info) context.Context {
	return context.WithValue(ctx, userInfoCtxKey, userInfo)
}
