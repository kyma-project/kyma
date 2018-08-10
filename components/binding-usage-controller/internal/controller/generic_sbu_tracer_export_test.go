package controller

const (
	TracingAnnotationKey = tracingAnnotationKey
)

func NewGenericUsageAnnotationTracer() *genericUsageAnnotationTracer {
	return &genericUsageAnnotationTracer{}
}
