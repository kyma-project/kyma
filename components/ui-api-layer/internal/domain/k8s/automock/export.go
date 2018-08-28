package automock

func NewDeploymentLister() *deploymentLister {
	return new(deploymentLister)
}

func NewEnvLister() *envLister {
	return new(envLister)
}

func NewResourceQuotaLister() *resourceQuotaLister {
	return new(resourceQuotaLister)
}

func NewStatefulSetLister() *statefulSetLister {
	return new(statefulSetLister)
}

func NewReplicaSetLister() *replicaSetLister {
	return new(replicaSetLister)
}

func NewPodsLister() *podsLister {
	return new(podsLister)
}

func NewDeploymentGetter() *deploymentGetter {
	return new(deploymentGetter)
}