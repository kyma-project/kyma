package trace

import (
	"log"
	"net/http"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

const (
	DefaultTraceDebug         = false
	DefaultTraceAPIURL        = "http://localhost:9411/api/v1/spans"
	DefaultTraceHostPort      = "0.0.0.0:0"
	DefaultTraceServiceName   = "trace-service"
	DefaultTraceOperationName = "trace-operation"
)

type Options struct {
	Debug         bool
	APIURL        string
	HostPort      string
	ServiceName   string
	OperationName string
}

type Tracer interface {
	Started() bool
	Options() *Options
	Stop()
}

type OpenTracer struct {
	opts      Options
	collector zipkin.Collector
}

func GetDefaultTraceOptions() *Options {
	options := Options{
		DefaultTraceDebug,
		DefaultTraceAPIURL,
		DefaultTraceHostPort,
		DefaultTraceServiceName,
		DefaultTraceOperationName,
	}
	return &options
}

func StartNewTracer(opts *Options) Tracer {
	tracer := new(OpenTracer)
	tracer.opts = *opts
	tracer.Start()
	return tracer
}

func (zk *OpenTracer) Start() {
	if collector, err := zipkin.NewHTTPCollector(zk.opts.APIURL); err != nil {
		log.Printf("Tracer :: Start :: Error creating Zipkin collector :: Error: %v", err)
	} else {
		recorder := zipkin.NewRecorder(collector, zk.opts.Debug, zk.opts.HostPort, zk.opts.ServiceName)
		if tracer, err := zipkin.NewTracer(recorder, zipkin.TraceID128Bit(false)); err != nil {
			log.Printf("Tracer :: Start :: Error creating Zipkin tracer :: Error: %v", err)
		} else {
			zk.collector = collector
			opentracing.SetGlobalTracer(tracer)
		}
	}
}

func (zk *OpenTracer) Started() bool {
	return zk.collector != nil
}

func (zk *OpenTracer) Options() *Options {
	return &zk.opts
}

func (zk *OpenTracer) Stop() {
	if zk.collector != nil {
		zk.collector.Close()
	}
}

func ReadTraceHeaders(header *http.Header) *opentracing.SpanContext {
	if header == nil {
		return nil
	}
	if spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(*header)); err != nil {
		return nil
	} else {
		return &spanContext
	}
}

func ReadTraceContext(traceContext *api.TraceContext) *opentracing.SpanContext {
	if spanContext, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, (*opentracing.TextMapCarrier)(traceContext)); err != nil {
		return nil
	} else {
		return &spanContext
	}
}

func StartSpan(spanContext *opentracing.SpanContext, operationName *string, opts ...opentracing.StartSpanOption) *opentracing.Span {
	if spanContext != nil {
		if opts == nil || len(opts) == 0 {
			opts = make([]opentracing.StartSpanOption, 0)
		}
		opts = append(opts, opentracing.ChildOf(*spanContext))
	}
	span := opentracing.StartSpan(*operationName, opts...)
	return &span
}

func WriteSpan(span *opentracing.Span) *api.TraceContext {
	if span == nil {
		log.Printf("Tracer :: WriteSpan :: Error writing trace span nil")
		return nil
	}
	traceContext := make(api.TraceContext)
	carrier := opentracing.TextMapCarrier(traceContext)
	opentracing.GlobalTracer().Inject((*span).Context(), opentracing.TextMap, carrier)
	return &traceContext
}

func TagSpanAsError(span *opentracing.Span, errorMessage, errorStack string) {
	if span != nil {
		ext.Error.Set(*span, true)

		// log more details about the error
		var fields []otlog.Field

		if len(errorMessage) != 0 {
			fields = append(fields, otlog.String("message", errorMessage)) // human readable error message
		}

		if len(errorStack) != 0 {
			fields = append(fields, otlog.String("stack", errorStack)) // error stacktrace
		}

		(*span).LogFields(fields...)
	}
}

func TagSpanWithHttpStatusCode(span *opentracing.Span, statusCode uint16) {
	if span != nil {
		ext.HTTPStatusCode.Set(*span, statusCode)
	}
}

func FinishSpan(span *opentracing.Span) {
	if span != nil {
		(*span).Finish()
	}
}

func SetSpanTags(span *opentracing.Span, tags *map[string]string) {
	if span != nil && tags != nil {
		for key, value := range *tags {
			(*span).SetTag(key, value)
		}
	}
}
