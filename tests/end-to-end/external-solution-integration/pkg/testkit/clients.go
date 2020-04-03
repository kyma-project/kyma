package testkit

import (
	kubelessclientset "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	k8s "k8s.io/client-go/kubernetes"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	gatewayclientset "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	appbrokerclientset "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appoperatorclientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	sbuclientset "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
)

type KymaClients struct {
	AppOperatorClientset         *appoperatorclientset.Clientset
	AppBrokerClientset           *appbrokerclientset.Clientset
	KubelessClientset            *kubelessclientset.Clientset
	CoreClientset                *k8s.Clientset
	Pods                         coreclient.PodInterface
	ServiceCatalogClientset      *servicecatalogclientset.Clientset
	ServiceBindingUsageClientset *sbuclientset.Clientset
	GatewayClientset             *gatewayclientset.Clientset
}

func InitKymaClients(config *rest.Config, testID string) KymaClients {
	coreClientset := k8s.NewForConfigOrDie(config)

	return KymaClients{
		AppOperatorClientset:         appoperatorclientset.NewForConfigOrDie(config),
		AppBrokerClientset:           appbrokerclientset.NewForConfigOrDie(config),
		KubelessClientset:            kubelessclientset.NewForConfigOrDie(config),
		CoreClientset:                coreClientset,
		Pods:                         coreClientset.CoreV1().Pods(testID),
		ServiceCatalogClientset:      servicecatalogclientset.NewForConfigOrDie(config),
		ServiceBindingUsageClientset: sbuclientset.NewForConfigOrDie(config),
		GatewayClientset:             gatewayclientset.NewForConfigOrDie(config),
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
