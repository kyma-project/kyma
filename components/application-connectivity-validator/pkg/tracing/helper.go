package tracing

import "context"

type Trace struct {
	traceID string
	spanID  string
}

//
//func GetTracingFromHeaders(headers map[string][]string) Trace {
//	traceID, ok := headers[TRACE_KEY]
//	if ok {
//		traceID = traceID
//	}
//
//	spanID, ok := headers[TRACE_KEY]
//	if !ok {
//		spanID = ""
//	}
//	//return Trace{traceID: traceID, spanID: spanID}
//}

func GetMetadata(ctx context.Context) map[string]string {
	m := map[string]string{
		TRACE_KEY: UNKNOWN_VALUE,
		SPAN_KEY:  UNKNOWN_VALUE,
	}
	if val, ok := ctx.Value(TRACE_KEY).(string); ok {
		m[TRACE_KEY] = val
	}
	if val, ok := ctx.Value(SPAN_KEY).(string); ok {
		m[SPAN_KEY] = val
	}
	return m
}
