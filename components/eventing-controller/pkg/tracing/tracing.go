package tracing

import (
	"context"
	"fmt"
	"net/http"

	cev2protocolhttp "github.com/cloudevents/sdk-go/v2/protocol/http"

	cev2 "github.com/cloudevents/sdk-go/v2/event"
)

const (
	traceParentCEExtensionsKey = "traceparent"
	traceParentKey             = "traceparent"

	b3TraceIDCEExtensionsKey      = "b3traceid"
	b3ParentSpanIDCEExtensionsKey = "b3parentspanid"
	b3SpanIDCEExtensionsKey       = "b3spanid"
	b3SampledCEExtensionsKey      = "b3sampled"
	b3FlagsCEExtensionsKey        = "b3flags"

	b3TraceIDKey      = "X-B3-TraceId"
	b3ParentSpanIDKey = "X-B3-ParentSpanId"
	b3SpanIDKey       = "X-B3-SpanId"
	b3SampledKey      = "X-B3-Sampled"
	b3FlagsKey        = "X-B3-Flags"
)

func AddTracingHeadersToContext(ctx context.Context, ce *cev2.Event) context.Context {
	traceHeader := http.Header{}
	if traceParent, ok := ce.Extensions()[traceParentCEExtensionsKey]; ok {
		traceHeader.Add(traceParentKey, fmt.Sprintf("%v", traceParent))
		// CE extension, "traceparent" was added in publisher proxy to continue the trace from here. Hence it needs to be deleted here.
		removeCEExtension(ce, traceParentKey)
	}

	if b3TraceID, ok := ce.Extensions()[b3TraceIDCEExtensionsKey]; ok {
		traceHeader.Add(b3TraceIDKey, fmt.Sprintf("%v", b3TraceID))
		// CE extensions were added in publisher proxy to continue the trace from here. Hence it needs to be deleted here.
		removeCEExtension(ce, b3TraceIDCEExtensionsKey)
	}
	if b3ParentSpanID, ok := ce.Extensions()[b3ParentSpanIDCEExtensionsKey]; ok {
		traceHeader.Add(b3ParentSpanIDKey, fmt.Sprintf("%v", b3ParentSpanID))
		// CE extensions were added in publisher proxy to continue the trace from here. Hence it needs to be deleted here.
		removeCEExtension(ce, b3ParentSpanIDCEExtensionsKey)
	}
	if b3SpanID, ok := ce.Extensions()[b3SpanIDCEExtensionsKey]; ok {
		traceHeader.Add(b3SpanIDKey, fmt.Sprintf("%v", b3SpanID))
		// CE extensions were added in publisher proxy to continue the trace from here. Hence it needs to be deleted here.
		removeCEExtension(ce, b3SpanIDCEExtensionsKey)
	}
	if b3Sampled, ok := ce.Extensions()[b3SampledCEExtensionsKey]; ok {
		traceHeader.Add(b3SampledKey, fmt.Sprintf("%v", b3Sampled))
		// CE extensions were added in publisher proxy to continue the trace from here. Hence it needs to be deleted here.
		removeCEExtension(ce, b3SampledCEExtensionsKey)
	}
	if b3Flags, ok := ce.Extensions()[b3FlagsCEExtensionsKey]; ok {
		traceHeader.Add(b3FlagsKey, fmt.Sprintf("%v", b3Flags))
		// CE extensions were added in publisher proxy to continue the trace from here. Hence it needs to be deleted here.
		removeCEExtension(ce, b3FlagsCEExtensionsKey)
	}
	if len(traceHeader) > 0 {
		ctx = cev2protocolhttp.WithCustomHeader(ctx, traceHeader)
	}
	return ctx
}

func removeCEExtension(e *cev2.Event, key string) {
	v1Context := e.Context.AsV1()
	delete(v1Context.Extensions, key)
}
