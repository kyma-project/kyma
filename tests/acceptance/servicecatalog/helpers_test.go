package servicecatalog_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	catalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	scClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/acceptance/pkg/repeat"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func isCatalogForbidden(url string) (bool, int, error) {
	config := osb.DefaultClientConfiguration()
	config.URL = fmt.Sprintf("%s/cluster", url)

	client, err := osb.NewClient(config)
	if err != nil {
		return false, 0, errors.Wrapf(err, "while creating osb client for broker with URL: %s", url)
	}
	var statusCode int
	isForbiddenError := func(err error) bool {
		statusCodeError, ok := err.(osb.HTTPStatusCodeError)
		if !ok {
			return false
		}
		statusCode = statusCodeError.StatusCode
		return statusCodeError.StatusCode == http.StatusForbidden
	}

	_, err = client.GetCatalog()
	switch {
	case err == nil:
		return false, http.StatusOK, nil
	case isForbiddenError(err):
		return true, statusCode, nil
	default:
		return false, statusCode, errors.Wrapf(err, "while getting catalog from broker with URL: %s", url)
	}
}
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

func awaitCatalogContainsServiceClasses(t *testing.T, namespace string, timeout time.Duration, services []osb.Service) {
	repeat.AssertFuncAtMost(t, func() error {
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

func testDetailsReport(t *testing.T, services []osb.Service, ns string) {
	t.Log("##### Start test report #####")

	cs, err := clientset.NewForConfig(kubeConfig(t))
	if err != nil {
		t.Errorf("Cannot get client during creating a report: %s", err)
	}

	scs, err := cs.ServicecatalogV1beta1().ServiceClasses(ns).List(v1.ListOptions{})
	if err != nil {
		t.Errorf("Cannot fetch ClusterServiceClasses list during creating a report: %s", err)
	}

	scCnt, err := scClient.NewForConfig(kubeConfig(t))
	if err != nil {
		t.Errorf("Cannot get ServiceCatalog client during creating a report: %s", err)
	}

	t.Logf("Available Classes from Service broker (amount: %d)", len(services))
	for _, service := range services {
		t.Logf(" - ServiceId: %q - Service name: %q \n", service.ID, service.Name)
	}

	t.Logf("Status of ClusterServiceClasses (amount: %d)", len(scs.Items))
	for _, sc := range scs.Items {
		t.Logf(" - Name: %q (ExternalName: %s, ExternalId: %q) \n",
			sc.Name,
			sc.GetExternalName(),
			sc.Spec.ExternalID)
		t.Logf("   Is removed from catalog: %t", sc.Status.CommonServiceClassStatus.RemovedFromBrokerCatalog)
	}

	sbs, err := scCnt.ServicecatalogV1beta1().ServiceBrokers(ns).List(metav1.ListOptions{})
	if err != nil {
		t.Errorf("Cannot fetch ServiceBrokers list during creating a report: %s", err)
	}

	t.Logf("Status Conditions of ServiceBrokers (amount: %d)", len(sbs.Items))
	for _, sb := range sbs.Items {
		t.Logf(" - ServiceBroker %q:", sb.Name)
		for _, cond := range sb.Status.Conditions {
			t.Logf("   StatusType: %s", cond.Type)
			t.Logf("   Status: %s", cond.Status)
			t.Logf("   StatusReason: %s", cond.Reason)
			t.Logf("   StatusMessage: %s", cond.Message)
			t.Log("    ---")
		}
	}

	t.Log("#####  End test report  #####")
}
