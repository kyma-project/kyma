package controller

const (
	TracingAnnotationKey = tracingAnnotationKey
)

func NewUsageAnnotationTracer() *usageAnnotationTracer {
	return &usageAnnotationTracer{}
}
