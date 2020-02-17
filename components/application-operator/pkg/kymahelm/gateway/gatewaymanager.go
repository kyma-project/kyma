package gateway

import (
	"fmt"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/application-operator/pkg/utils"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	gatewayChartDirectory = "gateway"
	gatewayNameFormat     = "%s-application-gateway"
)

//go:generate mockery -name GatewayManager
type GatewayManager interface {
	InstallGateway(namespace string) error
	DeleteGateway(namespace string) error
	GatewayExists(namespace string) (bool, error)
	UpgradeGateways() error
}

//go:generate mockery -name ApplicationMappingClient
type ApplicationMappingClient interface {
	List(opts v1.ListOptions) (*v1alpha1.ApplicationMappingList, error)
}

func NewGatewayManager(helmClient kymahelm.HelmClient, overrides OverridesData, appMappingClient ApplicationMappingClient) GatewayManager {
	return &gatewayManager{
		helmClient:       helmClient,
		overrides:        overrides,
		appMappingClient: appMappingClient,
	}
}

type gatewayManager struct {
	helmClient       kymahelm.HelmClient
	overrides        OverridesData
	appMappingClient ApplicationMappingClient
}

func (g *gatewayManager) InstallGateway(namespace string) error {
	overrides, err := kymahelm.ParseOverrides(g.overrides, overridesTemplate)
	if err != nil {
		return errors.Errorf("Error parsing overrides: %s", err)
	}

	name := getGatewayName(namespace)

	_, err = g.helmClient.InstallReleaseFromChart(gatewayChartDirectory, namespace, name, overrides)
	if err != nil {
		return errors.Errorf("Error installing Gateway: %s", err)
	}
	return nil
}

func (g *gatewayManager) DeleteGateway(namespace string) error {
	gateway := getGatewayName(namespace)
	releaseExist, err := g.gatewayExists(gateway, namespace)
	if err != nil {
		return errors.Errorf("Error deleting Gateway: %s", err)
	}

	if releaseExist {
		_, err := g.helmClient.DeleteRelease(gateway)
		if err != nil {
			return errors.Errorf("Error deleting Gateway: %s", err)
		}
	}
	return nil
}

func (g *gatewayManager) GatewayExists(namespace string) (bool, error) {
	name := getGatewayName(namespace)
	return g.gatewayExists(name, namespace)
}

func (g *gatewayManager) UpgradeGateways() error {
	namespaces, err := g.getAllNamespacesWithAppMappings()

	if err != nil {
		return errors.Errorf("Error updating Gateway: %s", err)
	}

	err = g.updateGateways(namespaces)

	if err != nil {
		return errors.Errorf("Error updating Gateway: %s", err)
	}

	return nil
}

func (g *gatewayManager) gatewayExists(name, namespace string) (bool, error) {
	listResponse, err := g.helmClient.ListReleases(namespace)
	if err != nil {
		return false, errors.Errorf("Error listing releases: %s", err)
	}

	releases := listResponse.Releases

	for _, rel := range releases {
		if rel.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (g *gatewayManager) getAllNamespacesWithAppMappings() ([]string, error) {
	list, err := g.appMappingClient.List(v1.ListOptions{})

	if err != nil {
		return nil, errors.Errorf("Error listing Application Mappings: %s", err)
	}
	var namespaces []string

	for _, appMapping := range list.Items {
		namespace := appMapping.Namespace
		namespaces = appendNamespace(namespaces, namespace)
	}
	return namespaces, nil
}

func (g *gatewayManager) updateGateways(namespaces []string) error {
	for _, namespace := range namespaces {
		gateway := getGatewayName(namespace)
		exists, err := g.gatewayExists(gateway, namespace)

		if err != nil {
			return errors.Errorf("Error checking Gateway: %s", err)
		}

		if exists {
			err = g.upgradeGateway(gateway)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *gatewayManager) upgradeGateway(gateway string) error {
	overrides, err := kymahelm.ParseOverrides(g.overrides, overridesTemplate)
	if err != nil {
		return errors.Errorf("Error parsing overrides: %s", err)
	}
	_, err = g.helmClient.UpdateReleaseFromChart(gatewayChartDirectory, gateway, overrides)

	if err != nil {
		return err
	}
	return nil
}

func appendNamespace(namespaces []string, namespace string) []string {
	if namespaceExists(namespaces, namespace) || utils.IsSystemNamespace(namespace) {
		return namespaces
	}
	return append(namespaces, namespace)
}

func namespaceExists(namespaces []string, namespace string) bool {
	for _, v := range namespaces {
		if v == namespace {
			return true
		}
	}
	return false
}

func getGatewayName(namespace string) string {
	return fmt.Sprintf(gatewayNameFormat, namespace)
}
