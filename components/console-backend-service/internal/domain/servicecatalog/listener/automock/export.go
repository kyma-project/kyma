package automock

func NewGQLBindingConverter() *gqlBindingConverter {
	return new(gqlBindingConverter)
}

func NewGQLInstanceConverter() *gqlInstanceConverter {
	return new(gqlInstanceConverter)
}

func NewGQLServiceBrokerConverter() *gqlServiceBrokerConverter {
	return new(gqlServiceBrokerConverter)
}

func NewGQLClusterServiceBrokerConverter() *gqlClusterServiceBrokerConverter {
	return new(gqlClusterServiceBrokerConverter)
}
