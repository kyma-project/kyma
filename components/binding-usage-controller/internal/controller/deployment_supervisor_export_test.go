package controller

func (m *DeploymentSupervisor) WithUsageAnnotationTracer(tracer usageBindingAnnotationTracer) *DeploymentSupervisor {
	m.tracer = tracer
	return m
}
