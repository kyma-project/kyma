package tracing

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
