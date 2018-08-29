package broker

//noinspection GoExportedFuncWithUnexportedType
func NewConverter() reToServiceConverter {
	return reToServiceConverter{}
}

func NewCatalogService(finder remoteEnvironmentFinder, reEnabledChecker reEnabledChecker, conv converter) *catalogService{
	return &catalogService{
		finder:           finder,
		reEnabledChecker: reEnabledChecker,
		conv:             conv,
	}
}
