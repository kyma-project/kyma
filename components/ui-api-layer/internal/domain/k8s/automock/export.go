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
