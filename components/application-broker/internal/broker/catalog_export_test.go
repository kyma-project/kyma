package broker

//noinspection GoExportedFuncWithUnexportedType
func NewConverter() appToServiceConverter {
	return appToServiceConverter{}
}

func NewCatalogService(finder applicationFinder, appEnabledChecker appEnabledChecker, conv converter) *catalogService {
	return &catalogService{
		finder:            finder,
		appEnabledChecker: appEnabledChecker,
		conv:              conv,
	}
}
