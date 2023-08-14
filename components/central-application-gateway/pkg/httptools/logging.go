package httptools

import (
	"context"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/uuid"
	"net/http"
	"time"
)

type ContextKey string

const ContextUUID ContextKey = "context-uuid"

var LoggingOn = false

func LogResponse(log *zap.SugaredLogger, res *http.Response) error {
	if !LoggingOn {
		return nil
	}

	req := res.Request
	log.With(
		"requestID", req.Context().Value(ContextUUID),
		"method", req.Method,
		"host", req.Host,
		"url", req.URL.RequestURI(),
		"proto", req.Proto,
		"code", res.StatusCode,
		"contentLength", res.ContentLength,
		"kind", "response",
	).Infof("%s %s %s %s", req.Method, req.Host, req.URL.RequestURI(), req.Proto)

	return nil
}

func LogRequest(log *zap.SugaredLogger, r *http.Request) {
	if !LoggingOn {
		return
	}

	log.With(
		"requestID", r.Context().Value(ContextUUID),
		"method", r.Method,
		"host", r.Host,
		"url", r.URL.RequestURI(),
		"proto", r.Proto,
		"kind", "request",
	).Infof("%s %s %s %s", r.Method, r.Host, r.URL.RequestURI(), r.Proto)
}

func RequestLogger(label string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := uuid.NewUUID()

		r = r.WithContext(context.WithValue(r.Context(), ContextUUID, uid))

		lw := newLoggingResponseWriter(w)

		h.ServeHTTP(lw, r)

		fullPath := r.RequestURI
		if fullPath == "" {
			fullPath = r.URL.RequestURI()
		}
		responseCode := lw.status
		duration := time.Since(lw.start).Nanoseconds() / int64(time.Millisecond)

		log := zap.L().Sugar().With(
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
