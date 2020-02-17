package automock

func NewDeploymentLister() *deploymentLister {
	return new(deploymentLister)
}

func NewResourceQuotaLister() *resourceQuotaLister {
	return new(resourceQuotaLister)
}

func NewGQLResourceQuotaConverter() *gqlResourceQuotaConverter {
	return new(gqlResourceQuotaConverter)
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

func NewSecretSvc() *secretSvc {
	return new(secretSvc)
}

func NewGQLPodConverter() *gqlPodConverter {
	return new(gqlPodConverter)
}

func NewGQLSecretConverter() *gqlSecretConverter {
	return new(gqlSecretConverter)
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

func NewSelfSubjectRulesSvc() *selfSubjectRulesSvc {
	return new(selfSubjectRulesSvc)
}

func NewSelfSubjectRulesConverter() *gqlSelfSubjectRulesConverter {
	return new(gqlSelfSubjectRulesConverter)
}

func NewNamespaceSvc() *namespaceSvc {
	return new(namespaceSvc)
}

func NewGqlVersionInfoConverter() *gqlVersionInfoConverter {
	return new(gqlVersionInfoConverter)
}
