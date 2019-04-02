package automock

func NewGQLPodConverter() *gqlPodConverter {
	return new(gqlPodConverter)
}

func NewGQLSecretConverter() *gqlSecretConverter {
	return new(gqlSecretConverter)
}

func NewGQLServiceConverter() *gqlServiceConverter {
	return new(gqlServiceConverter)
}

func NewGQLConfigMapConverter() *gqlConfigMapConverter {
	return new(gqlConfigMapConverter)
}
