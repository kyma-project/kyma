package authn

import (
	"context"
	"net/http"
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

			wsProtocolHeader := r.Header.Get("sec-websocket-protocol")
			if wsProtocolHeader != "" {
				wsProtocolParts := strings.Split(wsProtocolHeader, ",")
				wsProtocol, wsToken := strings.TrimSpace(wsProtocolParts[0]), strings.TrimSpace(wsProtocolParts[1])
				r.Header.Set("Authorization", "Bearer "+wsToken)
				r.Header.Set("sec-websocket-protocol", wsProtocol)
			}

			u, ok, err := a.AuthenticateRequest(r)
			if err != nil {
				glog.Errorf("Unable to authenticate the request due to an error: %v", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			}
			if ok {
				ctx := context.WithValue(r.Context(), userInfoCtxKey, u)

				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			}
		})
	}
}

func UserInfoForContext(ctx context.Context) user.Info {
	raw, _ := ctx.Value(userInfoCtxKey).(user.Info)
	return raw
}
