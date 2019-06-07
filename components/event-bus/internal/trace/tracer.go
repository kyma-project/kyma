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
	// DefaultTraceDebug trace option
	DefaultTraceDebug = false
	// DefaultTraceAPIURL trace option
	DefaultTraceAPIURL = "http://localhost:9411/api/v1/spans"
	// DefaultTraceHostPort trace option
	DefaultTraceHostPort = "0.0.0.0:0"
	// DefaultTraceServiceName trace option
	DefaultTraceServiceName = "trace-service"
	// DefaultTraceOperationName trace option
	DefaultTraceOperationName = "trace-operation"
)

// Options represents the trace options.
type Options struct {
	Debug         bool
	APIURL        string
	HostPort      string
	ServiceName   string
	OperationName string
}

// Tracer encapsulates the trace behaviours.
type Tracer interface {
	Started() bool
	Options() *Options
	Stop()
}

// OpenTracer represents an open tracer.
type OpenTracer struct {
	opts      Options
	collector zipkin.Collector
}

// GetDefaultTraceOptions returns a default trace options instance.
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

// StartNewTracer starts a new tracer.
func StartNewTracer(opts *Options) Tracer {
	tracer := new(OpenTracer)
	tracer.opts = *opts
	tracer.Start()
	return tracer
}

// Start starts a new OpenTracer instance.
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

// Started returns a boolean value indicating id the OpenTracer is started or not.
func (zk *OpenTracer) Started() bool {
	return zk.collector != nil
}

// Options returns the OpenTracer options.
func (zk *OpenTracer) Options() *Options {
	return &zk.opts
}

// Stop stops the OpenTracer.
func (zk *OpenTracer) Stop() {
	if zk.collector != nil {
		_ = zk.collector.Close()
	}
}

// ReadTraceHeaders returns an opentracing.SpanContext instance.
func ReadTraceHeaders(header *http.Header) *opentracing.SpanContext {
	if header == nil {
		return nil
	}
	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(*header))
	if err != nil {
		return nil
	}
	return &spanContext
}

// StartSpan starts and returns an opentracing.Span instance.
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

// WriteSpan writes the given opentracing.Span and returns an api.TraceContext instance.
func WriteSpan(span *opentracing.Span) *api.TraceContext {
	if span == nil {
		log.Printf("Tracer :: WriteSpan :: Error writing trace span nil")
		return nil
	}
	traceContext := make(api.TraceContext)
	carrier := opentracing.TextMapCarrier(traceContext)
	_ = opentracing.GlobalTracer().Inject((*span).Context(), opentracing.TextMap, carrier)
	return &traceContext
}

// TagSpanAsError tags the opentracing.Span as an error with the error details.
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

// FinishSpan finishes the opentracing.Span.
func FinishSpan(span *opentracing.Span) {
	if span != nil {
		(*span).Finish()
	}
}

// SetSpanTags sets the opentracing.Span tags.
func SetSpanTags(span *opentracing.Span, tags *map[string]string) {
	if span != nil && tags != nil {
		for key, value := range *tags {
			(*span).SetTag(key, value)
		}
	}
}
