package gateway

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/helm/pkg/proto/hapi/release"

	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/application-operator/pkg/utils"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	gatewayChartDirectory = "gateway"
	gatewayNameFormat     = "%s-gateway"
)

//go:generate mockery -name GatewayManager
type GatewayManager interface {
	InstallGateway(namespace string) error
	DeleteGateway(namespace string) error
	GatewayExists(namespace string) (bool, release.Status_Code, error)
	UpgradeGateways() error
}

//go:generate mockery -name ServiceCatalogueClient
type ServiceCatalogueClient interface {
	ServiceInstances(namespace string) v1beta1.ServiceInstanceInterface
}

func NewGatewayManager(helmClient kymahelm.HelmClient, overrides OverridesData, serviceCatalogueClient ServiceCatalogueClient, namespaceClient v12.NamespaceInterface) GatewayManager {
	return &gatewayManager{
		helmClient:             helmClient,
		overrides:              overrides,
		serviceCatalogueClient: serviceCatalogueClient,
		namespaces:             namespaceClient,
	}
}

type gatewayManager struct {
	helmClient             kymahelm.HelmClient
	overrides              OverridesData
	serviceCatalogueClient ServiceCatalogueClient
	namespaces             v12.NamespaceInterface
}

func (g *gatewayManager) InstallGateway(namespace string) error {
	overrides, err := kymahelm.ParseOverrides(g.overrides, overridesTemplate)
	if err != nil {
		return errors.Errorf("Error parsing overrides: %s", err.Error())
	}

	name := getGatewayName(namespace)

	_, err = g.helmClient.InstallReleaseFromChart(gatewayChartDirectory, namespace, name, overrides)
	if err != nil {
		return errors.Errorf("Error installing Gateway: %s", err.Error())
	}
	return nil
}

func (g *gatewayManager) DeleteGateway(namespace string) error {
	gateway := getGatewayName(namespace)
	releaseExist, _, err := g.gatewayExists(gateway, namespace)
	if err != nil {
		return errors.Errorf("Error deleting Gateway: %s", err.Error())
	}
	if releaseExist {
		return g.deleteGateway(gateway)
	}
	return nil
}

func (g *gatewayManager) deleteGateway(gateway string) error {
	_, err := g.helmClient.DeleteRelease(gateway)
	if err != nil {
		return errors.Errorf("Error deleting Gateway: %s", err.Error())
	}

	return nil
}

func (g *gatewayManager) GatewayExists(namespace string) (bool, release.Status_Code, error) {
	name := getGatewayName(namespace)
	exists, status, err := g.gatewayExists(name, namespace)
	return exists, status, err
}

func (g *gatewayManager) UpgradeGateways() error {
	namespaces, err := g.getAllNamespacesWithServiceInstances()

	if err != nil {
		return errors.Errorf("Error updating Gateways: %s", err.Error())
	}

	if len(namespaces) == 0 {
		return nil
	}

	g.updateGateways(namespaces)

	return nil
}

func (g *gatewayManager) gatewayExists(name, namespace string) (bool, release.Status_Code, error) {
	listResponse, err := g.helmClient.ListReleases(namespace)
	if err != nil {
		return false, release.Status_UNKNOWN, errors.Errorf("Error listing releases: %s", err.Error())
	}

	if listResponse == nil {
		return false, release.Status_UNKNOWN, nil
	}

	for _, rel := range listResponse.Releases {
		if rel.Name == name {
			return true, rel.Info.Status.Code, nil
		}
	}
	return false, release.Status_UNKNOWN, nil
}

func (g *gatewayManager) getAllNamespacesWithServiceInstances() ([]string, error) {
	namespacesList, err := g.namespaces.List(metav1.ListOptions{})

	if err != nil {
		return nil, errors.Errorf("Error listing namespaces: %s", err.Error())
	}

	var namespaces []string

	for _, namespace := range namespacesList.Items {
		instances := g.serviceCatalogueClient.ServiceInstances(namespace.Name)
		list, err := instances.List(metav1.ListOptions{})
		if err != nil {
			return nil, errors.Errorf("Error listing Service Instances for namespace %s: %s", namespace.Name, err.Error())
		}
		if len(list.Items) > 0 {
			namespaces = appendNamespace(namespaces, namespace.Name)
		}
	}
	return namespaces, nil
}

func (g *gatewayManager) updateGateways(namespaces []string) {
	for _, namespace := range namespaces {
		gateway := getGatewayName(namespace)
		exists, status, err := g.gatewayExists(gateway, namespace)

		if err != nil {
			log.Errorf("Error checking Gateway %s: %s", namespace, err.Error())
			continue
		}

		if exists {
			if status == release.Status_FAILED {
				log.Infof("Deleting Gateway %s in failed status", namespace)
				err := g.deleteGateway(gateway)
				if err != nil {
					log.Errorf("Error deleting Gateway %s: %s", namespace, err.Error())
				}
				continue
			}
			err = g.upgradeGateway(gateway)
			if err != nil {
				log.Errorf("Error upgrading Gateway %s: %s", namespace, err.Error())
			}
		}
	}
}

func (g *gatewayManager) upgradeGateway(gateway string) error {
	overrides, err := kymahelm.ParseOverrides(g.overrides, overridesTemplate)
	if err != nil {
		return errors.Errorf("Error parsing overrides: %s", err.Error())
	}
	_, err = g.helmClient.UpdateReleaseFromChart(gatewayChartDirectory, gateway, overrides)

	if err != nil {
		return err
	}
	return nil
}

func appendNamespace(namespaces []string, namespace string) []string {
	if utils.IsSystemNamespace(namespace) {
		return namespaces
	}
	return append(namespaces, namespace)
}

func getGatewayName(namespace string) string {
	return fmt.Sprintf(gatewayNameFormat, namespace)
}
