package broker

func NewOSBContext(originatingIdentity, apiVersion string) *osbContext {
	return &osbContext{
		OriginatingIdentity: originatingIdentity,
		APIVersion:          apiVersion,
	}
}
