package broker

//noinspection GoExportedFuncWithUnexportedType
// Deprecated
func NewConverter() *appToServiceConverter {
	return &appToServiceConverter{}
}

func NewConverterV2() *appToServiceConverterV2 {
	return &appToServiceConverterV2{}
}

func NewCatalogService(finder applicationFinder, serviceCheckerFactory serviceCheckerFactory, conv converter) *catalogService {
	return &catalogService{
		finder:            finder,
		appEnabledChecker: serviceCheckerFactory,
		conv:              conv,
	}
}
