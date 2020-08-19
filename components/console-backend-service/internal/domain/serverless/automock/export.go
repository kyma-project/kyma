package automock

func NewFunctionService() *functionSvc {
	return new(functionSvc)
}

func NewGQLFunctionConverter() *gqlFunctionConverter {
	return new(gqlFunctionConverter)
}
