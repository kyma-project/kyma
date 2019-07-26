package broker

func NewCatalogService(finder addonFinder, conv converter) *catalogService {
	return &catalogService{finder: finder, conv: conv}
}

//noinspection GoExportedFuncWithUnexportedType
func NewConverter() addonToServiceConverter {
	return addonToServiceConverter{}
}
