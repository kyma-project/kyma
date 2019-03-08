package middlewares

import (
	"net/http"
	"net/http/httputil"

	log "github.com/sirupsen/logrus"
)

type middleware struct {
}

func NewRequestLoggingMiddleware() *middleware {
	return &middleware{}
}

func (cc *middleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dumpRequestToLog(r)

		handler.ServeHTTP(w, r)
	})
}

func dumpRequestToLog(r *http.Request) {
	b, err := httputil.DumpRequest(r, false)
	if err != nil {
		log.Errorf("Failed to log request")
		return
	}
	log.Infof("%s", b)
}
