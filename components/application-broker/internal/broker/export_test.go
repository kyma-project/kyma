package broker

func NewOSBContext(originatingIdentity, apiVersion, namespace string) *osbContext {
	return &osbContext{
		OriginatingIdentity: originatingIdentity,
		APIVersion:          apiVersion,
		BrokerNamespace:     namespace,
	}
}
