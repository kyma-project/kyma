package broker

//noinspection GoExportedFuncWithUnexportedType
func NewConverter() appToServiceConverter {
	return appToServiceConverter{}
}

func NewCatalogService(finder applicationFinder, serviceCheckerFactory serviceCheckerFactory, conv converter) *catalogService {
	return &catalogService{
		finder:            finder,
		appEnabledChecker: serviceCheckerFactory,
		conv:              conv,
	}
}
