package controller

func (m *GenericSupervisor) WithUsageAnnotationTracer(tracer genericUsageBindingAnnotationTracer) *GenericSupervisor {
	m.tracer = tracer
	return m
}
