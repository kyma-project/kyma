package broker

func NewCatalogService(finder bundleFinder, conv converter) *catalogService {
	return &catalogService{finder: finder, conv: conv}
}

//noinspection GoExportedFuncWithUnexportedType
func NewConverter() bundleToServiceConverter {
	return bundleToServiceConverter{}
}
