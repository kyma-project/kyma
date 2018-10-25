package httptools

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func RequestLogger(label string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := newLoggingResponseWriter(w)

		h.ServeHTTP(lw, r)

		method := r.Method
		fullPath := r.RequestURI
		if fullPath == "" {
			fullPath = r.URL.RequestURI()
		}
		proto := r.Proto
		responseCode := lw.status
		duration := time.Since(lw.start).Nanoseconds() / int64(time.Millisecond)

		log.Infof("%s: %s %s %s %d %d", label, method, fullPath, proto, responseCode, duration)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	start  time.Time
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{ResponseWriter: w, start: time.Now()}
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.status = statusCode
}
