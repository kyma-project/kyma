package propagation

import (
	"net/http"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

// HTTPFormatSequence is a propagation.HTTPFormat that applies multiple other propagation formats.
// For incoming requests, it will use the first SpanContext it can find, checked in the order of
// HTTPFormatSequence.Ingress.
// For outgoing requests, it will apply all the formats to the outgoing request, in the order of
// HTTPFormatSequence.Egress.
type HTTPFormatSequence struct {
	Ingress []propagation.HTTPFormat
	Egress  []propagation.HTTPFormat
}

var _ propagation.HTTPFormat = (*HTTPFormatSequence)(nil)

// SpanContextFromRequest satisfies the propagation.HTTPFormat interface.
func (h *HTTPFormatSequence) SpanContextFromRequest(req *http.Request) (trace.SpanContext, bool) {
	for _, format := range h.Ingress {
		if sc, ok := format.SpanContextFromRequest(req); ok {
			return sc, true
		}
	}
	return trace.SpanContext{}, false
}

// SpanContextToRequest satisfies the propagation.HTTPFormat interface.
func (h *HTTPFormatSequence) SpanContextToRequest(sc trace.SpanContext, req *http.Request) {
	for _, format := range h.Egress {
		format.SpanContextToRequest(sc, req)
	}
}
