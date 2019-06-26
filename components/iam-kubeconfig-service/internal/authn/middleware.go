package authn

import (
	"net/http"

	log "github.com/sirupsen/logrus"
	"k8s.io/apiserver/pkg/authentication/authenticator"
)

func AuthMiddleware(a authenticator.Request) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			restoreAuthHeaderValue := r.Header.Get("Authorization") //Preserve original value

			_, ok, err := a.AuthenticateRequest(r) //Strips "Authorization" Header value on auth success!
			if err != nil {
				log.Errorf("Unable to authenticate the request due to an error: %v", err)
			}
			if !ok || err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			r.Header.Set("Authorization", restoreAuthHeaderValue)

			next.ServeHTTP(w, r)
		})
	}
}
