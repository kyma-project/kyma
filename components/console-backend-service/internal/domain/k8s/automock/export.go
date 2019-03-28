package automock

func NewDeploymentLister() *deploymentLister {
	return new(deploymentLister)
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

func NewPodSvc() *podSvc {
	return new(podSvc)
}

func NewGQLPodConverter() *gqlPodConverter {
	return new(gqlPodConverter)
}

func NewResourceSvc() *resourceSvc {
	return new(resourceSvc)
}

func NewGQLResourceConverter() *gqlResourceConverter {
	return new(gqlResourceConverter)
}

func NewReplicaSetSvc() *replicaSetSvc {
	return new(replicaSetSvc)
}

func NewGqlReplicaSetConverter() *gqlReplicaSetConverter {
	return new(gqlReplicaSetConverter)
}

func NewGqlServiceConverter() *gqlServiceConverter {
	return new(gqlServiceConverter)
}

func NewServiceSvc() *serviceSvc {
	return new(serviceSvc)
}

func NewGqlConfigMapConverter() *gqlConfigMapConverter {
	return new(gqlConfigMapConverter)
}

func NewConfigMapSvc() *configMapSvc {
	return new(configMapSvc)
}
