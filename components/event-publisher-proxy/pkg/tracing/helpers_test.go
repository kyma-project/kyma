package tracing

import (
	"net/http"
	"testing"

	cev2event "github.com/cloudevents/sdk-go/v2/event"

	. "github.com/onsi/gomega"
)

func TestAddTracingContextToCEExtensions(t *testing.T) {
	g := NewGomegaWithT(t)
	testCases := []struct {
		name               string
		headers            http.Header
		expectedExtensions map[string]interface{}
	}{
		{
			name: "headers with w3c tracing headers",
			headers: func() http.Header {
				headers := http.Header{}
				headers.Add(traceParentKey, "traceparent")
				return headers
			}(),
			expectedExtensions: map[string]interface{}{
				traceParentKey: "traceparent",
			},
		}, {
			name: "headers with b3 tracing headers",
			headers: func() http.Header {
				headers := http.Header{}
				headers.Add(b3TraceIDKey, "traceID")
				headers.Add(b3ParentSpanIDKey, "parentspanID")
				headers.Add(b3SpanIDKey, "spanID")
				headers.Add(b3SampledKey, "1")
				headers.Add(b3FlagsKey, "1")

				return headers
			}(),
			expectedExtensions: map[string]interface{}{
				b3TraceIDCEExtensionsKey:      "traceID",
				b3ParentSpanIDCEExtensionsKey: "parentspanID",
				b3SpanIDCEExtensionsKey:       "spanID",
				b3SampledCEExtensionsKey:      "1",
				b3FlagsCEExtensionsKey:        "1",
			},
		}, {
			name: "headers without tracing headers",
			headers: func() http.Header {
				headers := http.Header{}
				headers.Add("foo", "bar")
				return headers
			}(),
			expectedExtensions: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := cev2event.New()
			AddTracingContextToCEExtensions(tc.headers, &event)
			g.Expect(event.Extensions()).To(Equal(tc.expectedExtensions))
		})
	}
}
