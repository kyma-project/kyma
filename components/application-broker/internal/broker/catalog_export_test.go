package broker

//noinspection GoExportedFuncWithUnexportedType
func NewConverter() reToServiceConverter {
	return reToServiceConverter{}
}

func NewCatalogService(finder applicationFinder, reEnabledChecker reEnabledChecker, conv converter) *catalogService{
	return &catalogService{
		finder:           finder,
		reEnabledChecker: reEnabledChecker,
		conv:             conv,
	}
}
