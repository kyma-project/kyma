package tracing

import (
	"context"
	"net/http"
	"strings"
)

const (
	SPAN_HEADER_KEY  = "X-B3-Spanid"
	TRACE_HEADER_KEY = "X-B3-Traceid"
	TRACE_KEY        = "traceid"
	SPAN_KEY         = "spanid"
	UNKNOWN_VALUE    = "unknown"
)

type tracingMiddleware struct {
	handler func(w http.ResponseWriter, r *http.Request)
}

func NewTracingMiddleware(handler func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return &tracingMiddleware{
		handler: handler,
	}

}

func (m *tracingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	newCtx := addHeaderToCtx(r.Context(), r.Header, TRACE_HEADER_KEY)
	newCtx = addHeaderToCtx(newCtx, r.Header, SPAN_HEADER_KEY)

	m.handler(w, r.WithContext(newCtx))
	return
}

func addHeaderToCtx(ctx context.Context, headers http.Header, key string) context.Context {
	header, ok := headers[key]
	if !ok {
		return ctx
	}
	value := strings.Join(header, "")
	return context.WithValue(ctx, key, value)
}
