package controller

var TracingAnnotationKey = tracingAnnotationKey

func NewGenericUsageAnnotationTracer() *genericUsageAnnotationTracer {
	return &genericUsageAnnotationTracer{}
}
