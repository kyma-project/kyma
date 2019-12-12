// +build acceptance

package test

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	appsTypes "k8s.io/api/apps/v1beta1"
	k8sCoreTypes "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	dataModel "github.com/kyma-project/kyma/tests/service-catalog/cmd/env-tester/dto"
	"github.com/kyma-project/kyma/tests/service-catalog/utils/retriever"

	"github.com/pkg/errors"

	scTypes "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"

	bucTypes "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	bucClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	bucInterface "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"

	appTypes "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appClient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	appInterface "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"

	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	mappingClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"

	"github.com/kyma-project/kyma/tests/service-catalog/utils/repeat"
	"github.com/kyma-project/kyma/tests/service-catalog/utils/report"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	timeoutPerStep   = time.Minute
	timeoutPerAssert = 2 * time.Minute
	baseEnvName      = "GATEWAY_URL"
)

// Config contains all configurations for Service Binding Usage Acceptance tests
type Config struct {
	StubsDockerImage string `envconfig:"STUBS_DOCKER_IMAGE"`
}

func TestServiceBindingUsagePrefixing(t *testing.T) {
	// given
	ts := NewTestSuite(t)

	ts.waitForAPIServer()

	ts.createTestNamespace()
	ts.createApplication()

	defer func() {
		if t.Failed() {
			namespaceReport := report.NewReport(t,
				ts.k8sClientCfg,
				report.WithK8s(),
				report.WithSC(),
				report.WithApp(),
				report.WithSBU())
			namespaceReport.PrintJsonReport(ts.namespace)
		}
		ts.cleanup()
	}()

	ts.enableApplicationInTestNamespace()
	ts.waitForAppServiceClasses(timeoutPerAssert)

	ts.createAndWaitForServiceInstanceA(timeoutPerStep)
	ts.createAndWaitForServiceInstanceB(timeoutPerStep)

	ts.createAndWaitForServiceBindingA(timeoutPerStep)
	ts.createAndWaitForServiceBindingB(timeoutPerStep)

	ts.createTesterDeploymentAndService()

	// when
	ts.createBindingUsageForInstanceAWithoutPrefix(timeoutPerStep)
	ts.createBindingUsageForInstanceBWithPrefix(timeoutPerStep)
	// then

	ts.assertInjectedEnvVariables([]dataModel.EnvVariable{
		{Name: ts.envPrefix + baseEnvName, Value: ts.gatewayUrl},
		{Name: baseEnvName, Value: ts.gatewayUrl},
	}, timeoutPerAssert)
}

func NewTestSuite(t *testing.T) *TestSuite {
	var cfg Config
	err := envconfig.Init(&cfg)
	require.NoError(t, err)

	k8sCfg, err := restclient.InClusterConfig()
	require.NoError(t, err)

	randID := rand.String(5)

	return &TestSuite{
		t: t,

		k8sClientCfg:     k8sCfg,
		stubsDockerImage: cfg.StubsDockerImage,

		namespace:            fmt.Sprintf("svc-test-ns-%s", randID),
		testerDeploymentName: fmt.Sprintf("acc-test-env-tester-%s", randID),
		applicationName:      fmt.Sprintf("acc-test-app-env-%s", randID),
		gatewayUrl:           fmt.Sprintf("http://some-gateway-%s.url", randID),
		envPrefix:            "SOME_DUMMY_PREFIX_",

		serviceInstanceNameA: fmt.Sprintf("acc-test-instance-a-%s", randID),
		bindingNameA:         fmt.Sprintf("acc-test-credential-a-%s", randID),
		appSvcIDA:            fmt.Sprintf("acc-test-svc-id-a-%s", randID),

		serviceInstanceNameB: fmt.Sprintf("acc-test-instance-b-%s", randID),
		bindingNameB:         fmt.Sprintf("acc-test-credential-b-%s", randID),
		appSvcIDB:            fmt.Sprintf("acc-test-svc-id-b-%s", randID),
	}
}

type TestSuite struct {
	t *testing.T

	k8sClientCfg *restclient.Config

	namespace            string
	applicationName      string
	testerDeploymentName string
	gatewayUrl           string
	envPrefix            string

	serviceInstanceNameA string
	classExternalNameA   string
	appSvcIDA            string
	bindingNameA         string

	serviceInstanceNameB string
	classExternalNameB   string
	appSvcIDB            string
	bindingNameB         string

	stubsDockerImage string
}

// Application helpers
func (ts *TestSuite) createApplication() {
	reCli := ts.applicationClient()

	err := wait.Poll(time.Second, timeoutPerStep, func() (bool, error) {
		app, err := reCli.Create(ts.fixApplication())
		if err != nil {
			ts.t.Logf("while creating application %s: %v", ts.applicationName, err)
			return false, nil
		}
		ts.t.Logf("Application created [%s]", app.Name)
		return true, nil
	})
	require.NoError(ts.t, err)
}

func (ts *TestSuite) deleteApplication() {
	reCli := ts.applicationClient()

	err := wait.Poll(time.Second, timeoutPerStep, func() (bool, error) {
		if err := reCli.Delete(ts.applicationName, &metav1.DeleteOptions{}); err != nil {
			ts.t.Logf("while deleting application %s: %v", ts.applicationName, err)
			return false, nil
		}
		return true, nil
	})
	require.NoError(ts.t, err)
}

func (ts *TestSuite) enableApplicationInTestNamespace() {
	client, err := mappingClient.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)
	emCli := client.ApplicationconnectorV1alpha1().ApplicationMappings(ts.namespace)

	err = wait.Poll(time.Second, timeoutPerStep, func() (bool, error) {
		mapping, err := emCli.Create(ts.fixApplicationMapping())
		if err != nil {
			ts.t.Logf("while creating application mapping %s: %v", ts.applicationName, err)
			return false, nil
		}
		ts.t.Logf("Mapping created, name [%s], namespace: [%s]", mapping.Name, mapping.Namespace)
		return true, nil
	})
	require.NoError(ts.t, err)
}

func (ts *TestSuite) applicationClient() appInterface.ApplicationInterface {
	client, err := appClient.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)

	return client.ApplicationconnectorV1alpha1().Applications()
}

func (ts *TestSuite) fixApplicationMapping() *mappingTypes.ApplicationMapping {
	return &mappingTypes.ApplicationMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApplicationMapping",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ts.applicationName,
		},
	}
}

func (ts *TestSuite) fixApplication() *appTypes.Application {
	return &appTypes.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ts.applicationName,
		},
		Spec: appTypes.ApplicationSpec{
			AccessLabel:      "app-access-label",
			Description:      "Application used by acceptance test",
			SkipInstallation: true,
			Services: []appTypes.Service{
				{
					ID:   ts.appSvcIDA,
					Name: ts.appSvcIDA,
					Labels: map[string]string{
						"connected-app": ts.applicationName,
					},
					ProviderDisplayName: "Hakuna Matata",
					DisplayName:         "Some testable Application service",
					Description:         "Application Service Class used by application acceptance test",
					Tags:                []string{},
					Entries: []appTypes.Entry{
						{
							Type:        "API",
							AccessLabel: "some-access-label-A",
							GatewayUrl:  ts.gatewayUrl,
						},
					},
				},
				{
					ID:   ts.appSvcIDB,
					Name: ts.appSvcIDB,
					Labels: map[string]string{
						"connected-app": ts.applicationName,
					},
					ProviderDisplayName: "Hakuna Matata",
					DisplayName:         "Some testable Application service",
					Description:         "Application Service Class used by application acceptance test",
					Tags:                []string{},
					Entries: []appTypes.Entry{
						{
							Type:        "API",
							AccessLabel: "some-access-label-B",
							GatewayUrl:  ts.gatewayUrl,
						},
					},
				},
			},
		},
	}
}

// K8s namespace helpers
func (ts *TestSuite) createTestNamespace() {
	k8sCli, err := kubernetes.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)
	nsClient := k8sCli.CoreV1().Namespaces()

	err = wait.Poll(time.Second, timeoutPerStep, func() (bool, error) {
		if _, err := nsClient.Create(&k8sCoreTypes.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ts.namespace,
				Labels: map[string]string{
					"env": "true",
				},
			},
		}); err != nil {
			ts.t.Logf("while creating a namespace %s: %v", ts.namespace, err)
			return false, nil
		}
		return true, nil
	})
	require.NoError(ts.t, err)
	ts.t.Logf("Test namespace created [%s]", ts.namespace)
}

func (ts *TestSuite) deleteTestNamespace() {
	k8sCli, err := kubernetes.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)
	nsClient := k8sCli.CoreV1().Namespaces()

	err = wait.Poll(time.Second, timeoutPerStep, func() (bool, error) {
		if err := nsClient.Delete(ts.namespace, &metav1.DeleteOptions{}); err != nil {
			ts.t.Logf("while deleting a namespace %s: %v", ts.namespace, err)
			return false, nil
		}
		return true, nil
	})
	require.NoError(ts.t, err)
}

// Binding helpers
func (ts *TestSuite) createAndWaitForServiceBindingA(timeout time.Duration) {
	ts.createAndWaitForServiceBinding(ts.bindingNameA, ts.serviceInstanceNameA, timeout)
	ts.t.Logf("ServiceBinding [%s] for namespace [%s] is ready", ts.bindingNameA, ts.serviceInstanceNameA)
}

func (ts *TestSuite) createAndWaitForServiceBindingB(timeout time.Duration) {
	ts.createAndWaitForServiceBinding(ts.bindingNameB, ts.serviceInstanceNameB, timeout)
	ts.t.Logf("ServiceBinding [%s] for namespace [%s] is ready", ts.bindingNameB, ts.serviceInstanceNameB)

}

func (ts *TestSuite) deleteServiceBindingA(timeout time.Duration) {
	ts.deleteServiceBinding(ts.bindingNameA, timeout)
}

func (ts *TestSuite) deleteServiceBindingB(timeout time.Duration) {
	ts.deleteServiceBinding(ts.bindingNameB, timeout)
}

func (ts *TestSuite) deleteServiceBinding(bindingName string, timeout time.Duration) {
	clientSet, err := scClient.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)
	siClient := clientSet.ServicecatalogV1beta1().ServiceBindings(ts.namespace)

	err = wait.Poll(time.Second, timeout, func() (bool, error) {
		if err := siClient.Delete(bindingName, &metav1.DeleteOptions{}); err != nil {
			ts.t.Logf("while deleting binding %s: %v", bindingName, err)
			return false, nil
		}
		return true, nil
	})
	require.NoError(ts.t, err)

	repeat.AssertFuncAtMost(ts.t, func() error {
		binding, err := siClient.Get(bindingName, metav1.GetOptions{})
		switch {
		case err == nil:
			return fmt.Errorf(
				"serviceBiding %q still exists. [%s]",
				bindingName,
				prettyBindingResourceStatus(binding.Status.Conditions))
		case apiErrors.IsNotFound(err):
			return nil
		default:
			return errors.Wrap(err, "while getting ServiceBiding")
		}
	}, timeout)
}

func (ts *TestSuite) createAndWaitForServiceBinding(bindingName, instanceName string, timeout time.Duration) {
	scCli, err := scClient.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)

	bindingClient := scCli.ServicecatalogV1beta1().ServiceBindings(ts.namespace)
	err = wait.Poll(time.Second, timeoutPerStep, func() (bool, error) {
		if _, err := bindingClient.Create(&scTypes.ServiceBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bindingName,
				Namespace: ts.namespace,
			},
			Spec: scTypes.ServiceBindingSpec{
				ServiceInstanceRef: scTypes.LocalObjectReference{
					Name: instanceName,
				},
			},
		}); err != nil {
			ts.t.Logf("while creating Service Binding %s: %v", bindingName, err)
			return false, nil
		}
		return true, nil
	})
	require.NoError(ts.t, err)

	repeat.AssertFuncAtMost(ts.t, func() error {
		b, err := bindingClient.Get(bindingName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		isNotReady := func(instance *scTypes.ServiceBinding) bool {
			for _, cond := range instance.Status.Conditions {
				if cond.Type == scTypes.ServiceBindingConditionReady {
					return cond.Status != scTypes.ConditionTrue
				}
			}
			return true
		}

		if isNotReady(b) {
			return fmt.Errorf("ServiceBinding %s/%s is not in ready state. Status: %+v", b.Namespace, b.Name, b.Status)
		}

		return nil
	}, timeout)
}

// BindingUsage helpers
func (ts *TestSuite) createBindingUsageForInstanceAWithoutPrefix(timeout time.Duration) {
	sbuName := "binding-usage-tester"
	ts.createAndWaitBindingUsage(ts.bindingNameA, sbuName, "", timeout)
	ts.t.Logf("Binding usage [%s] is ready", sbuName)
}

func (ts *TestSuite) createBindingUsageForInstanceBWithPrefix(timeout time.Duration) {
	sbuName := "binding-usage-tester-with-prefix"
	ts.createAndWaitBindingUsage(ts.bindingNameB, sbuName, ts.envPrefix, timeout)
	ts.t.Logf("Binding usage [%s] is ready", sbuName)
}

func (ts *TestSuite) createAndWaitBindingUsage(bindingName, sbuName, envPrefix string, timeout time.Duration) {
	usageCli := ts.bindingUsageClient()
	sbu := &bucTypes.ServiceBindingUsage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBindingUsage",
			APIVersion: "servicecatalog.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: sbuName,
		},
		Spec: bucTypes.ServiceBindingUsageSpec{
			ServiceBindingRef: bucTypes.LocalReferenceByName{
				Name: bindingName,
			},
			UsedBy: bucTypes.LocalReferenceByKindAndName{
				Kind: "deployment",
				Name: ts.testerDeploymentName,
			},
		},
	}

	if envPrefix != "" {
		sbu.Spec.Parameters = &bucTypes.Parameters{
			EnvPrefix: &bucTypes.EnvPrefix{
				Name: envPrefix,
			},
		}
	}

	err := wait.Poll(time.Second, time.Minute, func() (bool, error) {
		if _, err := usageCli.Create(sbu); err != nil {
			ts.t.Logf("while creating ServiceBindingUsage %s: %v", sbuName, err)
			return false, nil
		}
		return true, nil
	})
	require.NoError(ts.t, err)

	repeat.AssertFuncAtMost(ts.t, func() error {
		usage, err := usageCli.Get(sbuName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		isNotReady := func(usage *bucTypes.ServiceBindingUsage) bool {
			for _, cond := range usage.Status.Conditions {
				if cond.Type == bucTypes.ServiceBindingUsageReady {
					return cond.Status != bucTypes.ConditionTrue
				}
			}
			return true
		}

		if isNotReady(usage) {
			return fmt.Errorf("ServiceBindingUsage %s/%s is not in ready state. Status: %+v", usage.Namespace, usage.Name, usage.Status)
		}

		return nil
	}, timeout)

}

func (ts *TestSuite) bindingUsageClient() bucInterface.ServiceBindingUsageInterface {
	client, err := bucClient.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)
	return client.ServicecatalogV1alpha1().ServiceBindingUsages(ts.namespace)
}

// ServiceInstance helpers
func (ts *TestSuite) createAndWaitForServiceInstanceA(timeout time.Duration) {
	ts.createAndWaitForServiceInstance(ts.serviceInstanceNameA, ts.classExternalNameA, timeout)
	ts.t.Logf("Service instance [%s] is ready", ts.serviceInstanceNameA)
}

func (ts *TestSuite) createAndWaitForServiceInstanceB(timeout time.Duration) {
	ts.createAndWaitForServiceInstance(ts.serviceInstanceNameB, ts.classExternalNameB, timeout)
	ts.t.Logf("Service instance [%s] is ready", ts.serviceInstanceNameA)

}

func (ts *TestSuite) deleteServiceInstanceA(timeout time.Duration) {
	ts.deleteServiceInstance(ts.serviceInstanceNameA, timeout)
}

func (ts *TestSuite) deleteServiceInstanceB(timeout time.Duration) {
	ts.deleteServiceInstance(ts.serviceInstanceNameB, timeout)
}

func (ts *TestSuite) deleteServiceInstance(instanceName string, timeout time.Duration) {
	clientSet, err := scClient.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)
	siClient := clientSet.ServicecatalogV1beta1().ServiceInstances(ts.namespace)

	err = wait.Poll(time.Second, timeout, func() (bool, error) {
		if err := siClient.Delete(instanceName, &metav1.DeleteOptions{}); err != nil {
			ts.t.Logf("while deleting instance %s: %v", instanceName, err)
			return false, nil
		}
		return true, nil
	})
	require.NoError(ts.t, err)

	repeat.AssertFuncAtMost(ts.t, func() error {
		instance, err := siClient.Get(instanceName, metav1.GetOptions{})
		switch {
		case err == nil:
			return fmt.Errorf(
				"serviceInstance %q still exists [%s]",
				instanceName,
				prettyInstanceResourceStatus(instance.Status.Conditions))
		case apiErrors.IsNotFound(err):
			return nil
		default:
			return errors.Wrap(err, "while getting ServiceInstance")
		}
	}, timeout)
}

func (ts *TestSuite) createAndWaitForServiceInstance(instanceName, classExternalName string, timeout time.Duration) {
	clientSet, err := scClient.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)
	siClient := clientSet.ServicecatalogV1beta1().ServiceInstances(ts.namespace)

	err = wait.Poll(time.Second, timeoutPerStep, func() (bool, error) {
		if _, err := siClient.Create(&scTypes.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name: instanceName,
			},
			Spec: scTypes.ServiceInstanceSpec{
				PlanReference: scTypes.PlanReference{
					ServiceClassExternalName: classExternalName,
					ServicePlanExternalName:  "default",
				},
			},
		}); err != nil {
			ts.t.Logf("while creating ServiceInstance %s: %v", instanceName, err)
			return false, nil
		}
		return true, nil
	})
	require.NoError(ts.t, err)

	repeat.AssertFuncAtMost(ts.t, func() error {
		si, err := siClient.Get(instanceName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		isNotReady := func(instance *scTypes.ServiceInstance) bool {
			for _, cond := range instance.Status.Conditions {
				if cond.Type == scTypes.ServiceInstanceConditionReady {
					return cond.Status != scTypes.ConditionTrue
				}
			}
			return true
		}

		if isNotReady(si) {
			return fmt.Errorf("ServiceInstance %s/%s is not in ready state. Status: %+v", si.Namespace, si.Name, si.Status)
		}

		return nil
	}, timeout)
}

// ServiceClass helpers
func (ts *TestSuite) waitForAppServiceClasses(timeout time.Duration) {
	repeat.AssertFuncAtMost(ts.t, ts.serviceClassIsAvailableA(), timeout)
	ts.t.Logf("Service class [%s] is available in the testing namespace", ts.appSvcIDA)
	repeat.AssertFuncAtMost(ts.t, ts.serviceClassIsAvailableB(), timeout)
	ts.t.Logf("Service class [%s] is available in the testing namespace", ts.appSvcIDB)

}

func (ts *TestSuite) serviceClassIsAvailableA() func() error {
	clientSet, err := scClient.NewForConfig(ts.k8sClientCfg)
	scCli := clientSet.ServicecatalogV1beta1()
	require.NoError(ts.t, err)

	return func() error {
		sc, err := retriever.ServiceClassByExternalID(scCli, ts.namespace, ts.appSvcIDA)
		if err != nil {
			return err
		}
		ts.classExternalNameA = sc.Spec.ExternalName
		return nil
	}
}

func (ts *TestSuite) serviceClassIsAvailableB() func() error {
	clientSet, err := scClient.NewForConfig(ts.k8sClientCfg)
	scCli := clientSet.ServicecatalogV1beta1()
	require.NoError(ts.t, err)

	return func() error {
		sc, err := retriever.ServiceClassByExternalID(scCli, ts.namespace, ts.appSvcIDB)
		if err != nil {
			return err
		}
		ts.classExternalNameB = sc.Spec.ExternalName
		return nil
	}
}

func (ts *TestSuite) cleanup() {
	ts.deleteServiceBindingA(timeoutPerStep)
	ts.deleteServiceBindingB(timeoutPerStep)
	ts.deleteServiceInstanceA(timeoutPerStep)
	ts.deleteServiceInstanceB(timeoutPerStep)
	ts.deleteTestNamespace()
	ts.deleteApplication()
}

// Deployment helpers
func (ts *TestSuite) createTesterDeploymentAndService() {
	clientset, err := kubernetes.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)

	labels := map[string]string{
		"app": "acc-test-env-tester",
	}
	deploy := ts.envTesterDeployment(labels)
	svc := ts.envTesterService(labels)

	deploymentClient := clientset.AppsV1beta1().Deployments(ts.namespace)
	err = wait.Poll(time.Second, timeoutPerStep, func() (bool, error) {
		if _, err := deploymentClient.Create(deploy); err != nil {
			ts.t.Logf("while creating a deployment: %v", err)
			return false, nil
		}
		return true, nil
	})
	require.NoError(ts.t, err)

	serviceClient := clientset.CoreV1().Services(ts.namespace)
	err = wait.Poll(time.Second, timeoutPerStep, func() (bool, error) {
		if _, err := serviceClient.Create(svc); err != nil {
			ts.t.Logf("while creating service: %v", err)
			return false, nil
		}
		return true, nil
	})
	require.NoError(ts.t, err)

	ts.t.Logf("Tester deployment and service created")
}

func (*TestSuite) envTesterService(labels map[string]string) *k8sCoreTypes.Service {
	return &k8sCoreTypes.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "acc-test-env-tester",
		},
		Spec: k8sCoreTypes.ServiceSpec{
			Selector: labels,
			Ports: []k8sCoreTypes.ServicePort{
				{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
				},
			},
			Type: k8sCoreTypes.ServiceTypeNodePort,
		},
	}
}

func (ts *TestSuite) envTesterDeployment(labels map[string]string) *appsTypes.Deployment {
	var replicas int32 = 1
	return &appsTypes.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: ts.testerDeploymentName,
		},
		Spec: appsTypes.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas: &replicas,
			Template: k8sCoreTypes.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: k8sCoreTypes.PodSpec{
					Containers: []k8sCoreTypes.Container{
						{
							Name:  "app",
							Image: ts.stubsDockerImage,
							Ports: []k8sCoreTypes.ContainerPort{
								{
									Name:          "http",
									Protocol:      k8sCoreTypes.ProtocolTCP,
									ContainerPort: 8080,
								},
							},
							Command: []string{"/go/bin/env-tester.bin"},
						},
					},
				},
			},
		},
	}
}

func (ts *TestSuite) assertInjectedEnvVariables(requiredVariables []dataModel.EnvVariable, timeout time.Duration) {
	url := fmt.Sprintf("http://acc-test-env-tester.%s.svc.cluster.local/envs", ts.namespace)

	repeat.AssertFuncAtMost(ts.t, func() error {
		client := http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
				DialContext: (&net.Dialer{
					Timeout:   20 * time.Second,
					KeepAlive: 20 * time.Second,
					DualStack: true,
				}).DialContext,
				Proxy:                 http.ProxyFromEnvironment,
				MaxIdleConns:          50,
				IdleConnTimeout:       30 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}
		req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("while getting envs from [%s], received unexpected status code [got: %d, expected: %d]", url, resp.StatusCode, http.StatusOK)
		}

		decoder := json.NewDecoder(resp.Body)
		var data []dataModel.EnvVariable
		err = decoder.Decode(&data)
		if err != nil {
			return err
		}

		var missing []dataModel.EnvVariable
		for _, req := range requiredVariables {
			found := false
			for _, act := range data {
				if req.Value == act.Value && req.Name == act.Name {
					found = true
					break
				}
			}
			if !found {
				missing = append(missing, req)
			}
		}

		if len(missing) > 0 {
			return fmt.Errorf("environment variables were not injected: [%v]", missing)
		}
		err = resp.Body.Close()
		if err != nil {
			return err
		}

		return nil
	}, timeout)
	ts.t.Logf("Environment variables are injected [%v]", requiredVariables)
}

func (ts *TestSuite) waitForAPIServer() {
	k8sCli, err := kubernetes.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)

	nsClient := k8sCli.CoreV1().Namespaces()

	repeat.AssertFuncAtMost(ts.t, func() error {
		_, err := nsClient.List(metav1.ListOptions{})
		return err
	}, 20*time.Second)
}

func prettyBindingResourceStatus(conditions []scTypes.ServiceBindingCondition) string {
	var response string

	for _, condition := range conditions {
		response += fmt.Sprintf("Status: %q, Reason: %q, Message: %q", condition.Status, condition.Reason, condition.Message)
	}

	return response
}

func prettyInstanceResourceStatus(conditions []scTypes.ServiceInstanceCondition) string {
	var response string

	for _, condition := range conditions {
		response += fmt.Sprintf("Status: %q, Reason: %q, Message: %q", condition.Status, condition.Reason, condition.Message)
	}

	return response
}
