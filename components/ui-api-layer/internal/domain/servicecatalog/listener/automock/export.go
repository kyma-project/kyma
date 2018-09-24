package automock

func NewGQLBindingConverter() *gqlBindingConverter {
	return new(gqlBindingConverter)
}

func NewGQLBindingUsageConverter() *gqlBindingUsageConverter {
	return new(gqlBindingUsageConverter)
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
