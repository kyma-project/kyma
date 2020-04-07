package broker

//noinspection GoExportedFuncWithUnexportedType
// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
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
