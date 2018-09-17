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
