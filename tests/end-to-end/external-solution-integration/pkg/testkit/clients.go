package testkit

import (
	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	kubelessClient "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	gatewayClient "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	appBrokerClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appOperatorClient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	eventingClient "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	serviceBindingUsageClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	coreClient "k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/rest"
)

type KymaClients struct {
	AppOperatorClientset         *appOperatorClient.Clientset
	AppBrokerClientset           *appBrokerClient.Clientset
	KubelessClientset            *kubeless.Clientset
	CoreClientset                *coreClient.Clientset
	Pods                         v1.PodInterface
	EventingClientset            *eventingClient.Clientset
	ServiceCatalogClientset      *serviceCatalogClient.Clientset
	ServiceBindingUsageClientset *serviceBindingUsageClient.Clientset
	GatewayClientset             *gatewayClient.Clientset
}

func InitKymaClients(config *rest.Config, testID string) KymaClients {
	coreClientset := coreClient.NewForConfigOrDie(config)

	return KymaClients{
		AppOperatorClientset:         appOperatorClient.NewForConfigOrDie(config),
		AppBrokerClientset:           appBrokerClient.NewForConfigOrDie(config),
		KubelessClientset:            kubelessClient.NewForConfigOrDie(config),
		CoreClientset:                coreClientset,
		Pods:                         coreClientset.CoreV1().Pods(testID),
		EventingClientset:            eventingClient.NewForConfigOrDie(config),
		ServiceCatalogClientset:      serviceCatalogClient.NewForConfigOrDie(config),
		ServiceBindingUsageClientset: serviceBindingUsageClient.NewForConfigOrDie(config),
		GatewayClientset:             gatewayClient.NewForConfigOrDie(config),
	}
}

type CompassClients struct {
	DirectorClient  *CompassDirectorClient
	ConnectorClient *CompassConnectorClient
}

func InitCompassClients(kymaClients KymaClients, state CompassDirectorClientState, domain string, skipSSLVerify bool) CompassClients {
	director := NewCompassDirectorClientOrDie(kymaClients.CoreClientset, state, domain)
	compassConnector := NewCompassConnectorClient(skipSSLVerify)

	return CompassClients{
		ConnectorClient: compassConnector,
		DirectorClient:  director,
	}
}
