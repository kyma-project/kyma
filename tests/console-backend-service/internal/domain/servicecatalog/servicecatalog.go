package servicecatalog

const (
	ModuleName               string = "servicecatalog"
	TestNamespace            string = "console-backend-service-sc"
	ClusterBrokerReleaseName        = "cluster-test-broker"
	BrokerReleaseName               = "test-broker"

	ClusterServiceBrokerKind = "ClusterServiceBroker"
	ServiceBrokerKind        = "ServiceBroker"
	CommonBrokerURL = "example.com"
)

var brokerDeletionGracefulPeriod int64 = 0
