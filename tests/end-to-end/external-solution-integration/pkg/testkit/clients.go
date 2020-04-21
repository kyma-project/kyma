package testkit

import (
	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	k8s "k8s.io/client-go/kubernetes"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	appbrokerclientset "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appoperatorclientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	sbuclientset "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
)

var (
	apiRuleRes = schema.GroupVersionResource{Group: "gateway.kyma-project.io", Version: "v1alpha1", Resource: "apirules"}
)

type KymaClients struct {
	AppOperatorClientset         *appoperatorclientset.Clientset
	AppBrokerClientset           *appbrokerclientset.Clientset
	CoreClientset                *k8s.Clientset
	Pods                         coreclient.PodInterface
	ServiceCatalogClientset      *servicecatalogclientset.Clientset
	ServiceBindingUsageClientset *sbuclientset.Clientset
	ApiRules                     dynamic.ResourceInterface
}

func InitKymaClients(config *rest.Config, testID string) KymaClients {
	coreClientset := k8s.NewForConfigOrDie(config)

	return KymaClients{
		AppOperatorClientset:         appoperatorclientset.NewForConfigOrDie(config),
		AppBrokerClientset:           appbrokerclientset.NewForConfigOrDie(config),
		CoreClientset:                coreClientset,
		Pods:                         coreClientset.CoreV1().Pods(testID),
		ServiceCatalogClientset:      servicecatalogclientset.NewForConfigOrDie(config),
		ServiceBindingUsageClientset: sbuclientset.NewForConfigOrDie(config),
		ApiRules:                     dynamic.NewForConfigOrDie(config).Resource(apiRuleRes).Namespace(testID),
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
