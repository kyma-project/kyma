package tracecontextb3

import (
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	ocpropagation "go.opencensus.io/trace/propagation"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/tracing/propagation"
)

// TraceContextEgress is a propagation.HTTPFormat that reads both TraceContext and B3 tracing
// formats, preferring TraceContext. It always writes TraceContext format exclusively.
var TraceContextEgress = &propagation.HTTPFormatSequence{
	Ingress: []ocpropagation.HTTPFormat{
		&tracecontext.HTTPFormat{},
		&b3.HTTPFormat{},
	},
	Egress: []ocpropagation.HTTPFormat{
		&tracecontext.HTTPFormat{},
	},
}
