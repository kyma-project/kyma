package servicecatalog_test

import (
	"fmt"
	"strings"
	"testing"

	"time"

	catalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func getCatalogForBroker(t *testing.T, url string) []osb.Service {
	config := osb.DefaultClientConfiguration()
	config.URL = url

	client, err := osb.NewClient(config)
	require.NoError(t, err)

	response, err := client.GetCatalog()
	require.NoError(t, err)

	return response.Services
}

// awaitCatalogContainsServiceClasses asserts that service catalog contains all OSB services mapped to Cluster Service Class.
func awaitCatalogContainsServiceClasses(t *testing.T, timeout time.Duration, services []osb.Service) {
	timeoutCh := time.After(timeout)
	for {
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
			return
		}

		select {
		case <-timeoutCh:
			assert.Fail(t, fmt.Sprintf("Service Catalog must contains ClusterServiceClasses for every broker service. Missing: %s", serviceNames(serviceMap)))
			return
		default:
			time.Sleep(2 * time.Second)
		}
	}
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

func deleteClusterServiceClassesForRemoteEnvironment(t *testing.T, re *v1alpha1.RemoteEnvironment) {
	cs, err := clientset.NewForConfig(kubeConfig(t))
	require.NoError(t, err)

	planClient := cs.ServicecatalogV1beta1().ClusterServicePlans()
	plans, err := planClient.List(v1.ListOptions{})
	require.NoError(t, err)

	// remove all plans related to Remote Environment services
	for _, plan := range plans.Items {
		for _, svc := range re.Spec.Services {
			if plan.Spec.ClusterServiceClassRef.Name == svc.ID {
				err := planClient.Delete(plan.Name, &v1.DeleteOptions{})
				assert.NoError(t, err)
			}
		}
	}

	classClient := cs.ServicecatalogV1beta1().ClusterServiceClasses()
	for _, svc := range re.Spec.Services {
		err := classClient.Delete(svc.ID, &v1.DeleteOptions{})
		assert.NoError(t, err)
	}
}

func kubeConfig(t *testing.T) *rest.Config {
	cfg, err := rest.InClusterConfig()
	require.NoError(t, err)
	return cfg
}
