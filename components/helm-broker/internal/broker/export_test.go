package broker

func NewOSBContext(originatingIdentity, apiVersion string) *OsbContext {
	return &OsbContext{
		OriginatingIdentity: originatingIdentity,
		APIVersion:          apiVersion,
	}
}
