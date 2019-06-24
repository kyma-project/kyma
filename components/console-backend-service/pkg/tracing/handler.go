package tracing

import (
	"github.com/golang/glog"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"net/http"
)

type OpentracingHandler http.HandlerFunc

func NewWithParentSpan(next http.HandlerFunc) OpentracingHandler {
	return func(writer http.ResponseWriter, request *http.Request) {
		spanContext, err := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(request.Header))
		if err != nil {
			glog.Warning("opentracing parent span headers extract", err)
			next(writer, request)
		}
		span := opentracing.StartSpan("console-backend-service",
			opentracing.ChildOf(spanContext))
		defer span.Finish()
		ext.SpanKind.Set(span, "server")
		ext.Component.Set(span, "console-backend-service")
		ctx := opentracing.ContextWithSpan(request.Context(), span)
		next(writer, request.WithContext(ctx))
	}
}
