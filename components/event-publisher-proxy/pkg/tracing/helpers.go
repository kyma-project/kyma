package tracing

import (
	"fmt"
	"net/http"

	cev2 "github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/extensions"
)

const (
	traceParentKey    = "traceparent"
	b3TraceIDKey      = "X-B3-TraceId"
	b3ParentSpanIDKey = "X-B3-ParentSpanId"
	b3SpanIDKey       = "X-B3-SpanId"
	b3SampledKey      = "X-B3-Sampled"
	b3FlagsKey        = "X-B3-Flags"

	b3TraceIDCEExtensionsKey      = "b3traceid"
	b3ParentSpanIDCEExtensionsKey = "b3parentspanid"
	b3SpanIDCEExtensionsKey       = "b3spanid"
	b3SampledCEExtensionsKey      = "b3sampled"
	b3FlagsCEExtensionsKey        = "b3flags"
)

func AddTracingContextToCEExtensions(reqHeaders http.Header, event *cev2.Event) {
	traceParent := reqHeaders.Get(traceParentKey)
	if len(traceParent) > 0 {
		st := extensions.DistributedTracingExtension{
			TraceParent: fmt.Sprintf("%v", traceParent),
		}
		st.AddTracingAttributes(event)
	}

	b3TraceID := reqHeaders.Get(b3TraceIDKey)
	if len(b3TraceID) > 0 {
		event.SetExtension(b3TraceIDCEExtensionsKey, b3TraceID)
	}

	b3ParentSpanID := reqHeaders.Get(b3ParentSpanIDKey)
	if len(b3ParentSpanID) > 0 {
		event.SetExtension(b3ParentSpanIDCEExtensionsKey, b3ParentSpanID)
	}

	b3SpanID := reqHeaders.Get(b3SpanIDKey)
	if len(b3SpanID) > 0 {
		event.SetExtension(b3SpanIDCEExtensionsKey, b3SpanID)
	}

	b3Sampled := reqHeaders.Get(b3SampledKey)
	if len(b3Sampled) > 0 {
		event.SetExtension(b3SampledCEExtensionsKey, b3Sampled)
	}

	b3Flags := reqHeaders.Get(b3FlagsKey)
	if len(b3Flags) > 0 {
		event.SetExtension(b3FlagsCEExtensionsKey, b3Flags)
	}
}
