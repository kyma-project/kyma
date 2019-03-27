package tester

import "time"

const (
	ReleaseName                = "console-backend-service-cluster-ups-broker"
	DefaultSubscriptionTimeout = 5 * time.Second
	UPSBrokerImage             = "quay.io/kubernetes-service-catalog/user-broker:v0.1.36"
	ClusterServiceBroker       = "ClusterServiceBroker"
	ServiceBroker              = "ServiceBroker"
	ClusterBrokerReleaseName   = "cluster-ups-broker"
	BrokerReleaseName          = "ups-broker"
)
