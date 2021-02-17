package gateway

import (
	"context"
	"encoding/json"
	"fmt"

	v1beta12 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	log "github.com/sirupsen/logrus"

	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

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
	GatewayExists(namespace string) (bool, release.Status, error)
	UpgradeGateways() error
}

//go:generate mockery -name ServiceInstanceClient
type ServiceInstanceClient interface {
	List(ctx context.Context, opts metav1.ListOptions) (*v1beta12.ServiceInstanceList, error)
}

func NewGatewayManager(helmClient kymahelm.HelmClient, overrides OverridesData, serviceInstanceClient ServiceInstanceClient) GatewayManager {
	return &gatewayManager{
		helmClient:            helmClient,
		overrides:             overrides,
		serviceInstanceClient: serviceInstanceClient,
	}
}

type gatewayManager struct {
	helmClient            kymahelm.HelmClient
	overrides             OverridesData
	serviceInstanceClient ServiceInstanceClient
	namespaces            v1.NamespaceInterface
}

func (g *gatewayManager) InstallGateway(namespace string) error {
	overrides, err := g.getOverrides()
	if err != nil {
		return errors.Errorf("Error parsing overrides: %s", err.Error())
	}

	name := getGatewayReleaseName(namespace)

	_, err = g.helmClient.InstallReleaseFromChart(gatewayChartDirectory, name, namespace, overrides)
	if err != nil {
		return errors.Errorf("Error installing Gateway: %s", err.Error())
	}
	return nil
}

func (g *gatewayManager) DeleteGateway(namespace string) error {
	gateway := getGatewayReleaseName(namespace)
	releaseExist, _, err := g.gatewayExists(gateway, namespace)
	if err != nil {
		return errors.Errorf("Error deleting Gateway: %s", err.Error())
	}
	if releaseExist {
		return g.deleteGateway(gateway, namespace)
	}
	return nil
}

func (g *gatewayManager) deleteGateway(gateway string, namespace string) error {
	_, err := g.helmClient.DeleteRelease(gateway, namespace)
	if err != nil {
		return errors.Errorf("Error deleting Gateway: %s", err.Error())
	}

	return nil
}

func (g *gatewayManager) GatewayExists(namespace string) (bool, release.Status, error) {
	name := getGatewayReleaseName(namespace)
	exists, status, err := g.gatewayExists(name, namespace)
	return exists, status, err
}

func (g *gatewayManager) UpgradeGateways() error {
	namespaces, err := g.getUniqueServiceInstanceNamespaces()

	if err != nil {
		return err
	}

	if namespaces == nil {
		return nil
	}

	g.updateGateways(namespaces)

	return nil
}

func (g *gatewayManager) getUniqueServiceInstanceNamespaces() ([]string, error) {
	list, err := g.serviceInstanceClient.List(context.Background(), metav1.ListOptions{})

	if err != nil {
		return nil, errors.Errorf("Error listing Service Instances: %s", err.Error())
	}

	var namespaces []string

	for _, instance := range list.Items {
		namespaces = appendNamespace(namespaces, instance.Namespace)
	}

	return namespaces, nil
}

func (g *gatewayManager) updateGateways(namespaces []string) {
	for _, namespace := range namespaces {
		gateway := getGatewayReleaseName(namespace)
		exists, status, err := g.gatewayExists(gateway, namespace)

		if err != nil {
			log.Errorf("Error checking Gateway %s: %s", namespace, err.Error())
			continue
		}

		if exists {
			if status == release.StatusFailed {
				log.Infof("Deleting Gateway %s in failed status", namespace)
				err := g.deleteGateway(gateway, namespace)
				if err != nil {
					log.Errorf("Error deleting Gateway %s: %s", namespace, err.Error())
				}
				continue
			}
			err = g.upgradeGateway(gateway, namespace)
			if err != nil {
				log.Errorf("Error upgrading Gateway %s: %s", namespace, err.Error())
			}
		}
	}
}

func (g *gatewayManager) gatewayExists(name, namespace string) (bool, release.Status, error) {
	listResponse, err := g.helmClient.ListReleases(namespace)
	if err != nil {
		return false, release.StatusUnknown, errors.Errorf("Error listing releases: %s", err.Error())
	}

	if listResponse == nil {
		return false, release.StatusUnknown, nil
	}

	for _, rel := range listResponse {
		if rel.Name == name {
			return true, rel.Info.Status, nil
		}
	}
	return false, release.StatusUnknown, nil
}

func (g *gatewayManager) upgradeGateway(gateway string, namespace string) error {
	overrides, err := g.getOverrides()

	if err != nil {
		return errors.Errorf("Error parsing overrides: %s", err.Error())
	}

	_, err = g.helmClient.UpdateReleaseFromChart(gatewayChartDirectory, gateway, namespace, overrides)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to update %s Gateway", gateway))
	}

	return nil
}

func (g *gatewayManager) getOverrides() (map[string]interface{}, error) {
	overridesData := g.overrides

	var overridesMap map[string]interface{}
	bytes, err := json.Marshal(overridesData)

	if err = json.Unmarshal(bytes, &overridesMap); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"global": overridesMap,
	}, nil
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

func getGatewayReleaseName(namespace string) string {
	return fmt.Sprintf(gatewayNameFormat, namespace)
}
