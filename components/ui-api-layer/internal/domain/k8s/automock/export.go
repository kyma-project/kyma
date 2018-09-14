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

func NewDeploymentGetter() *deploymentGetter {
	return new(deploymentGetter)
}

func NewLimitRangeLister() *limitRangeLister {
	return new(limitRangeLister)
}

func NewResourceQuotaStatusChecker() *resourceQuotaStatusChecker {
	return new(resourceQuotaStatusChecker)
}