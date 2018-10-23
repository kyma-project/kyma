package servicecatalog_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	catalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/acceptance/pkg/repeat"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func getCatalogForBroker(url string) ([]osb.Service, error) {
	config := osb.DefaultClientConfiguration()
	config.URL = url

	client, err := osb.NewClient(config)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating osb client for broker with URL: %s", url)
	}
	response, err := client.GetCatalog()
	if err != nil {
		return nil, errors.Wrapf(err, "while getting catalog from broker with URL: %s", url)
	}

	return response.Services, nil
}

// awaitCatalogContainsClusterServiceClasses asserts that service catalog contains all OSB services mapped to Cluster Service Class.
func awaitCatalogContainsClusterServiceClasses(t *testing.T, timeout time.Duration, services []osb.Service) {
	repeat.FuncAtMost(t, func() error {
		serviceMap := make(map[string]osb.Service)
		for _, service := range services {
			serviceMap[service.ID] = service
		}
		// fetch service classes from service-catalog
		clusterServiceClasses := getClusterServiceClasses(t).Items

		// cluster service class name is equal to OSB service ID
		for _, csc := range clusterServiceClasses {
			if svc, ok := serviceMap[csc.Name]; ok {
				assert.Equal(t, svc.Name, csc.Spec.ExternalName, fmt.Sprintf("ClusterServiceClass (ID: %s) must have the same name as OSB Service.", svc.ID))
				delete(serviceMap, csc.Name)
			}
		}
		if len(serviceMap) == 0 {
			return nil
		}

		return fmt.Errorf("service catalog must contains ClusterServiceClasses for every broker service. Missing: %s", serviceNames(serviceMap))
	}, timeout)
}

func awaitCatalogContainsServiceClasses(t *testing.T, namespace string, timeout time.Duration, services []osb.Service) {
	repeat.FuncAtMost(t, func() error {
		serviceMap := make(map[string]osb.Service)
		for _, service := range services {
			serviceMap[service.ID] = service
		}
		// fetch service classes from service-catalog
		serviceClasses := getServiceClasses(t, namespace).Items

		// cluster service class name is equal to OSB service ID
		for _, csc := range serviceClasses {
			if svc, ok := serviceMap[csc.Name]; ok {
				assert.Equal(t, svc.Name, csc.Spec.ExternalName, fmt.Sprintf("ServiceClass (ID: %s) must have the same name as OSB Service.", svc.ID))
				delete(serviceMap, csc.Name)
			}
		}
		if len(serviceMap) == 0 {
			return nil
		}

		return fmt.Errorf("service catalog must contains ServiceClasses for every broker service. Missing: %s", serviceNames(serviceMap))
	}, timeout)
}

func serviceNames(services map[string]osb.Service) string {
	parts := make([]string, 0)
	for _, svc := range services {
		parts = append(parts, svc.Name+":"+svc.ID)
	}
	return strings.Join(parts, ", ")
}

func getClusterServiceClasses(t *testing.T) *catalog.ClusterServiceClassList {
	cs, err := clientset.NewForConfig(kubeConfig(t))
	require.NoError(t, err)

	res, err := cs.ServicecatalogV1beta1().ClusterServiceClasses().List(v1.ListOptions{})
	require.NoError(t, err)

	return res
}

func getServiceClasses(t *testing.T, namespace string) *catalog.ServiceClassList {
	cs, err := clientset.NewForConfig(kubeConfig(t))
	require.NoError(t, err)

	res, err := cs.ServicecatalogV1beta1().ServiceClasses(namespace).List(v1.ListOptions{})
	require.NoError(t, err)

	return res
}

func kubeConfig(t *testing.T) *rest.Config {
	cfg, err := rest.InClusterConfig()
	require.NoError(t, err)
	return cfg
}

func fixNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
	}
}
