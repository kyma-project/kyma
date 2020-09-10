package health

import (
	"net/http"
)

const (
	livenessURI  = "/healthz"
	readinessURI = "/readyz"
)

func CheckHealth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.RequestURI == livenessURI {
			writer.WriteHeader(http.StatusOK)
			return
		}
		if request.RequestURI == readinessURI {
			writer.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
