package automock

func NewGQLBindingUsageConverter() *gqlBindingUsageConverter {
	return new(gqlBindingUsageConverter)
}

func NewGQLClusterAddonsConfigurationConverter() *gqlClusterAddonsConfigurationConverter {
	return new(gqlClusterAddonsConfigurationConverter)
}

func NewGQLAddonsConfigurationConverter() *gqlAddonsConfigurationConverter {
	return new(gqlAddonsConfigurationConverter)
}
