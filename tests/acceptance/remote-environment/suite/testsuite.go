package suite

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
	"time"

	catalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	bindingusage "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
)

type TestSuite struct {
	// TestID is a short id used as suffix in resource names
	TestID string

	namespace             string
	remoteEnvironmentName string
	dockerImage           string

	config *restclient.Config

	osbServiceId              string
	gatewayUrl                string
	gatewaySvcName            string
	bindingName               string
	appSvcDeploymentName      string
	gwClientSvcDeploymentName string
	accessLabel               string

	serviceInstance    *catalog.ServiceInstance
	serviceClass       *catalog.ClusterServiceClass
	testerBindingUsage *bindingusage.ServiceBindingUsage

	t *testing.T
}

func NewTestSuite(t *testing.T, image, namespace string) *TestSuite {
	config, err := restclient.InClusterConfig()
	require.NoError(t, err)

	id := rand.String(4)
	gwSvcName := fmt.Sprintf("acc-test-gw-%s", id)

	return &TestSuite{
		TestID:                id,
		namespace:             namespace,
		dockerImage:           image,
		remoteEnvironmentName: fmt.Sprintf("acc-test-re-%s", id),

		config: config,

		appSvcDeploymentName:      fmt.Sprintf("acc-test-app-%s", id),
		gwClientSvcDeploymentName: fmt.Sprintf("acc-test-client-%s", id),
		gatewaySvcName:            gwSvcName,
		osbServiceId:              fmt.Sprintf("acc-test-osb-serviceid-%s", id),
		bindingName:               fmt.Sprintf("acc-test-credential-%s", id),
		gatewayUrl:                fmt.Sprintf("http://%s", gwSvcName),
		accessLabel:               fmt.Sprintf("acc-test-access-label-%s", id),

		t: t,
	}
}

func (ts *TestSuite) Setup() {
	ts.t.Log("Creating deployments and services")
	ts.createKubernetesResources()
	ts.t.Log("Creating Istio resources")
	ts.createIstioResources()
	ts.t.Log("Creating RemoteEnvironment")
	ts.createRemoteEnvironmentResources()
}

func (ts *TestSuite) WaitForServiceClassWithTimeout(timeout time.Duration) {
	clientSet, err := clientset.NewForConfig(ts.config)
	done := time.After(timeout)
	require.NoError(ts.t, err)
	for {
		ts.serviceClass, err = clientSet.ServicecatalogV1beta1().ClusterServiceClasses().Get(ts.osbServiceId, metav1.GetOptions{})
		if err == nil {
			return
		}

		if !apierrors.IsNotFound(err) {
			require.Fail(ts.t, err.Error())
		}

		select {
		case <-done:
			require.Fail(ts.t, fmt.Sprintf("timeout while waiting for service class %s", ts.osbServiceId))
		default:
			time.Sleep(time.Second)
		}
	}
}

func (ts *TestSuite) ProvisionServiceInstance(timeout time.Duration) {
	clientSet, err := clientset.NewForConfig(ts.config)
	require.NoError(ts.t, err)
	siClient := clientSet.ServicecatalogV1beta1().ServiceInstances(ts.namespace)
	ts.serviceInstance, err = siClient.Create(&catalog.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("acc-test-remote-env-%s", ts.TestID),
		},
		Spec: catalog.ServiceInstanceSpec{
			PlanReference: catalog.PlanReference{
				ClusterServiceClassExternalName: ts.serviceClass.Spec.ExternalName,
				ClusterServicePlanExternalName:  "default",
			},
		},
	})
	require.NoError(ts.t, err)

	done := time.After(timeout)
	for {
		si, err := siClient.Get(ts.serviceInstance.Name, metav1.GetOptions{})

		if err == nil {
			for _, cnd := range si.Status.Conditions {
				if cnd.Type == catalog.ServiceInstanceConditionReady && cnd.Status == catalog.ConditionTrue {
					return
				}
			}
		} else {
			ts.t.Logf("error while getting service instance: %s", err.Error())
		}

		select {
		case <-done:
			if si != nil {
				require.Fail(ts.t, fmt.Sprintf("timeout while waiting for service instance %s to be ready. Status: %v", si.Name, si.Status))
			} else {
				require.Fail(ts.t, "timeout while waiting for service instance to be ready")
			}
		default:
			ts.t.Log("waiting for service instance to be ready")
			time.Sleep(time.Second)
		}
	}
}

func (ts *TestSuite) Bind(timeout time.Duration) {
	clientSet, err := clientset.NewForConfig(ts.config)
	require.NoError(ts.t, err)
	bindingClient := clientSet.ServicecatalogV1beta1().ServiceBindings(ts.namespace)
	binding, err := bindingClient.Create(&catalog.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: ts.bindingName,
		},
		Spec: catalog.ServiceBindingSpec{
			ServiceInstanceRef: catalog.LocalObjectReference{
				Name: ts.serviceInstance.Name,
			},
		},
	})
	require.NoError(ts.t, err)

	done := time.After(timeout)
	for {
		b, err := bindingClient.Get(binding.Name, metav1.GetOptions{})

		if err == nil {
			for _, cnd := range b.Status.Conditions {
				if cnd.Type == catalog.ServiceBindingConditionReady && cnd.Status == catalog.ConditionTrue {
					return
				}
			}
		} else {
			ts.t.Logf("error while getting binding: %s", err.Error())
		}

		select {
		case <-done:
			if b != nil {
				require.Fail(ts.t, fmt.Sprintf("timeout while waiting for service binding %s to be ready. Status: %v", b.Name, b.Status))
			} else {
				require.Fail(ts.t, "timeout while waiting for service binding to be ready")
			}
		default:
			time.Sleep(time.Second)
		}
	}

}

func (ts *TestSuite) TearDown() {
	defer ts.deleteNamespace()
	defer ts.deleteRemoteEnvironment()
}

func (ts *TestSuite) executeCall(url string) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	return client.Get(url)
}

func (ts *TestSuite) WaitForCallSucceededAndEnvInjected(t *testing.T, timeout time.Duration) {
	ts.waitForResultCondition(t, timeout, func(d map[string]string) bool {
		return d[envInjectedKey] == "true" && d[callSucceededKey] == "true"
	})
}

func (ts *TestSuite) WaitForCallForbiddenAndEnvNotInjected(t *testing.T, timeout time.Duration) {
	ts.waitForResultCondition(t, timeout, func(d map[string]string) bool {
		return d[envInjectedKey] == "false" && d[callForbiddenKey] == "true"
	})
}

func (ts *TestSuite) waitForResultCondition(t *testing.T, timeout time.Duration, conditionFn func(map[string]string) bool) {
	clientSet, err := kubernetes.NewForConfig(ts.config)
	require.NoError(t, err)

	done := time.After(timeout)
	for {
		cfgMap, err := clientSet.CoreV1().ConfigMaps(ts.namespace).Get("test-output", metav1.GetOptions{})
		require.NoError(t, err)

		if conditionFn(cfgMap.Data) {
			return
		}

		select {
		case <-done:
			ts.printGatewayClientLogs()
			require.Fail(t, fmt.Sprintf("timeout for tester results (%s) exceeded", timeout.String()))
			return
		default:
			time.Sleep(2 * time.Second)
		}
	}
}

func int32Ptr(i int32) *int32 { return &i }
