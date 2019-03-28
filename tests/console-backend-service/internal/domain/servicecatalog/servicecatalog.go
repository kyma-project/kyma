package servicecatalog

const (
	ModuleName               string = "servicecatalog"
	TestNamespace            string = "console-backend-service-sc"
	ClusterBrokerReleaseName        = "helm-broker"

	ClusterServiceBrokerKind = "ClusterServiceBroker"
	ServiceBrokerKind        = "ServiceBroker"
)

var noGracefulPeriod int64 = 0
