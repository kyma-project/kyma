package broker

func NewOSBContext(originatingIdentity, apiVersion string) *OsbContext {
	return &OsbContext{
		OriginatingIdentity: originatingIdentity,
		APIVersion:          apiVersion,
	}
}

func NewCatalogSyncerService(underlying catalogGetter) *catalogSyncerService {
	return newCatalogSyncerService(underlying)
}