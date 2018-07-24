package controller

func (m *KubelessFunctionSupervisor) WithUsageAnnotationTracer(tracer usageBindingAnnotationTracer) *KubelessFunctionSupervisor {
	m.tracer = tracer
	return m
}
