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

			//r.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjIwOTUxZDAyNTFjMjE0NzY1MmQ0MDQ0Y2E3ODBiZjM5MWFhODFhNjMifQ.eyJpc3MiOiJodHRwczovL2RleC5reW1hLmxvY2FsIiwic3ViIjoiQ2lCcU5tWnFZWHBpTUdNMGEzSm9hSEU1WlhsbmJqZGhlWFZvTW1VNGVXUjBZaElGYkc5allXdyIsImF1ZCI6WyJreW1hLWNsaWVudCIsImNvbnNvbGUiXSwiZXhwIjoxNTUzNzk2NjU4LCJpYXQiOjE1NTM3Njc4NTgsImF6cCI6ImNvbnNvbGUiLCJub25jZSI6ImVlNWRlOWY4ZmE4NTQ4MmNiMjg3ZWU1YTAyNWU4NWJmIiwiYXRfaGFzaCI6InZraU1hdHhjeGp5OEFKUGltYjVNd1EiLCJlbWFpbCI6ImFkbWluQGt5bWEuY3giLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwibmFtZSI6ImFkbWluIn0.kSsFn_9TYb3HK49YJqNRUR8cJY3zBJmKIcLFJxmgTjIeylhsgvhlNRGP3_xm6n39x5GzcDm94PM1QnKdMbdAsgxY6yfXMcJBKNXPpSN7497ocicwoXKPXR6P8H8d-34PumQ6iuWBoYdupkxlnlVbXXbBD7EvJ8KhBRvud7QtTv9W7j0wNtxfBOtz672SikvdTbsGVpFunXs3qyTzVfWaD3-I0539uCGR43XVoSkqoubMNxxXFxe95IzkccDH1URzSdxufDR75wnW7DOONM6FPZcF0XyM7V4L8pf6svOj_VHSZnIPBh68cHGrHw9JHEVM3FgwGaUQhswR49yVbNzm4A")
			//r.Header.Set("sec-websocket-protocol", "graphql-ws")

			u, ok, err := a.AuthenticateRequest(r)
			if err != nil {
				glog.Errorf("Unable to authenticate the request due to an error: %v", err)
			}
			if !ok || err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := WithUserInfoContext(r.Context(), u)
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
