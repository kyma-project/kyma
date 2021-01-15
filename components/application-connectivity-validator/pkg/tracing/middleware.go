package tracing

import (
	"context"
	"net/http"
	"strings"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/pkg/logger"
)

const (
	SPAN_HEADER_KEY  = "Spanid"
	TRACE_HEADER_KEY = "Traceid"
)

type tracingMiddleware struct {
	log     *logger.Logger
	handler func(w http.ResponseWriter, r *http.Request)
}

func NewTracingMiddleware(log *logger.Logger, handler func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return &tracingMiddleware{
		log:     log,
		handler: handler,
	}

}

func (m *tracingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	headers := r.Header
	m.log.Infof("My favourite header: %+v", headers)

	newCtx := addHeaderToCtx(r.Context(), r.Header, TRACE_HEADER_KEY)
	newCtx = addHeaderToCtx(newCtx, r.Header, SPAN_HEADER_KEY)

	m.handler(w, r.WithContext(newCtx))
	return
}

//TODO: think where to enhance logger in tracing or in logger, maybe pass function
// log.Enhance(func()).Info()
// log.Enhance(tracing.AddMetadata(r.ctx)).Info()
// I think that such API looks not so bad
func AddTracingInfo(log *logger.Logger, ctx context.Context) *logger.Logger {
	enhancedLogger := log.With("abc", "123")

	return &logger.Logger{enhancedLogger}
}

func addHeaderToCtx(ctx context.Context, headers http.Header, key string) context.Context {
	header, ok := headers[key]
	if !ok {
		return ctx
	}
	value := strings.Join(header, "")
	return context.WithValue(ctx, strings.ToLower(key), value)
}
