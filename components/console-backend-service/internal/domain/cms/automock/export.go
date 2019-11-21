package automock

// ClusterDocsTopic

func NewClusterDocsTopicService() *clusterDocsTopicSvc {
	return new(clusterDocsTopicSvc)
}

func NewGQLClusterDocsTopicConverter() *gqlClusterDocsTopicConverter {
	return new(gqlClusterDocsTopicConverter)
}

// DocsTopic

func NewDocsTopicService() *docsTopicSvc {
	return new(docsTopicSvc)
}

func NewGQLDocsTopicConverter() *gqlDocsTopicConverter {
	return new(gqlDocsTopicConverter)
}