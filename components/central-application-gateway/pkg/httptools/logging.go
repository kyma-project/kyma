package httptools

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type ContextKey string

const ContextUUID ContextKey = "context-uuid"

func LogResponse(log *zap.SugaredLogger, res *http.Response) error {
	req := res.Request
	log.WithOptions(
		zap.AddCallerSkip(1),
	).
		With(
			"method", req.Method,
			"host", req.Host,
			"url", req.URL.RequestURI(),
			"proto", req.Proto,
			"code", res.StatusCode,
			"contentLength", res.ContentLength,
			"kind", "response",
		).Debugf("%s %s %s %s", req.Method, req.Host, req.URL.RequestURI(), req.Proto)

	return nil
}

func LogRequest(log *zap.SugaredLogger, r *http.Request) {
	log.WithOptions(
		zap.AddCallerSkip(1),
	).With(
		"method", r.Method,
		"host", r.Host,
		"url", r.URL.RequestURI(),
		"proto", r.Proto,
		"kind", "request",
	).Debugf("%s %s %s %s", r.Method, r.Host, r.URL.RequestURI(), r.Proto)
}

func RequestLogger(label string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		lw := newLoggingResponseWriter(w)

		h.ServeHTTP(lw, r)

		responseCode := lw.status
		duration := time.Since(lw.start).Nanoseconds() / int64(time.Millisecond)

		log := zap.L().Sugar().
			With(
				"label", label,
				"duration", duration,
				"code", responseCode,
			)

		LogRequest(log, r)
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
