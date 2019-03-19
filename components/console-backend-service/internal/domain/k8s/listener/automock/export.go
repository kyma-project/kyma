package automock

func NewGQLPodConverter() *gqlPodConverter {
	return new(gqlPodConverter)
}

func NewGQLServiceConverter() *gqlServiceConverter {
	return new(gqlServiceConverter)
}
