package servicecatalog_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/pkg/errors"

	scTypes "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"

	bucTypes "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	bucClient "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned"
	bucInterface "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"

	reTypes "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	reClient "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
	reInterface "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/tests/acceptance/pkg/repeat"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
	appsTypes "k8s.io/api/apps/v1beta1"
	k8sCoreTypes "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
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

	ts.createTestNamespace()
	ts.createRemoteEnvironment()

	defer func() {
		if t.Failed() {
			ts.dumpTestNamespace()
		}
		ts.cleanup()
	}()

	ts.enableRemoteEnvironmentInTestNamespace()
	ts.waitForREServiceClasses(time.Second * 90)

	ts.createAndWaitForServiceInstanceA(timeoutPerStep)
	ts.createAndWaitForServiceInstanceB(timeoutPerStep)

	ts.createAndWaitForServiceBindingA(timeoutPerStep)
	ts.createAndWaitForServiceBindingB(timeoutPerStep)

	ts.createTesterDeploymentAndService()

	// when
	ts.createBindingUsageForInstanceAWithoutPrefix()
	ts.createBindingUsageForInstanceBWithPrefix()

	// then
	ts.assertInjectedEnvVariable(baseEnvName, ts.gatewayUrl, timeoutPerAssert)
	ts.assertInjectedEnvVariable(ts.envPrefix+baseEnvName, ts.gatewayUrl, timeoutPerAssert)
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

		namespace:             fmt.Sprintf("svc-test-ns-%s", randID),
		testerDeploymentName:  fmt.Sprintf("acc-test-env-tester-%s", randID),
		remoteEnvironmentName: fmt.Sprintf("acc-test-re-env-%s", randID),
		gatewayUrl:            fmt.Sprintf("http://some-gateway-%s.url", randID),
		envPrefix:             "SOME_DUMMY_PREFIX_",

		serviceInstanceNameA: fmt.Sprintf("acc-test-instance-a-%s", randID),
		bindingNameA:         fmt.Sprintf("acc-test-credential-a-%s", randID),
		reSvcNameA:           fmt.Sprintf("acc-test-svc-id-a-%s", randID),

		serviceInstanceNameB: fmt.Sprintf("acc-test-instance-b-%s", randID),
		bindingNameB:         fmt.Sprintf("acc-test-credential-b-%s", randID),
		reSvcNameB:           fmt.Sprintf("acc-test-svc-id-b-%s", randID),
	}
}

type TestSuite struct {
	t *testing.T

	k8sClientCfg *restclient.Config

	namespace             string
	remoteEnvironmentName string
	testerDeploymentName  string
	gatewayUrl            string
	envPrefix             string

	serviceInstanceNameA string
	classExternalNameA   string
	reSvcNameA           string
	bindingNameA         string

	serviceInstanceNameB string
	classExternalNameB   string
	reSvcNameB           string
	bindingNameB         string

	stubsDockerImage string
}

// Remote Environment helpers
func (ts *TestSuite) createRemoteEnvironment() {
	reCli := ts.remoteEnvironmentClient()

	_, err := reCli.Create(ts.fixRemoteEnvironment())
	require.NoError(ts.t, err)
}

func (ts *TestSuite) deleteRemoteEnvironment() {
	reCli := ts.remoteEnvironmentClient()

	err := reCli.Delete(ts.remoteEnvironmentName, &metav1.DeleteOptions{})
	require.NoError(ts.t, err)
}

func (ts *TestSuite) enableRemoteEnvironmentInTestNamespace() {
	reCli, err := reClient.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)

	emCli := reCli.ApplicationconnectorV1alpha1().EnvironmentMappings(ts.namespace)
	_, err = emCli.Create(ts.fixEnvironmentMapping())
	require.NoError(ts.t, err)
}

func (ts *TestSuite) remoteEnvironmentClient() reInterface.RemoteEnvironmentInterface {
	client, err := reClient.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)

	return client.ApplicationconnectorV1alpha1().RemoteEnvironments()
}

func (ts *TestSuite) fixEnvironmentMapping() *reTypes.EnvironmentMapping {
	return &reTypes.EnvironmentMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EnvironmentMapping",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ts.remoteEnvironmentName,
		},
	}
}

func (ts *TestSuite) fixRemoteEnvironment() *reTypes.RemoteEnvironment {
	return &reTypes.RemoteEnvironment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RemoteEnvironment",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ts.remoteEnvironmentName,
		},
		Spec: reTypes.RemoteEnvironmentSpec{
			AccessLabel: "re-access-label",
			Description: "Remote Environment used by acceptance test",
			Services: []reTypes.Service{
				{
					ID:   ts.reSvcNameA,
					Name: ts.reSvcNameA,
					Labels: map[string]string{
						"connected-app": ts.remoteEnvironmentName,
					},
					ProviderDisplayName: "SAP Hybris",
					DisplayName:         "Some testable RE service",
					Description:         "Remote Environment Service Class used by remote-environment acceptance test",
					Tags:                []string{},
					Entries: []reTypes.Entry{
						{
							Type:        "API",
							AccessLabel: "some-access-label-A",
							GatewayUrl:  ts.gatewayUrl,
						},
					},
				},
				{
					ID:   ts.reSvcNameB,
					Name: ts.reSvcNameB,
					Labels: map[string]string{
						"connected-app": ts.remoteEnvironmentName,
					},
					ProviderDisplayName: "SAP Hybris",
					DisplayName:         "Some testable RE service",
					Description:         "Remote Environment Service Class used by remote-environment acceptance test",
					Tags:                []string{},
					Entries: []reTypes.Entry{
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
	_, err = nsClient.Create(&k8sCoreTypes.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ts.namespace,
			Labels: map[string]string{
				"env": "true",
			},
		},
	})
	require.NoError(ts.t, err)
}

func (ts *TestSuite) deleteTestNamespace() {
	k8sCli, err := kubernetes.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)

	nsClient := k8sCli.CoreV1().Namespaces()
	err = nsClient.Delete(ts.namespace, &metav1.DeleteOptions{})
	require.NoError(ts.t, err)
}

// Binding helpers
func (ts *TestSuite) createAndWaitForServiceBindingA(timeout time.Duration) {
	ts.createAndWaitForServiceBinding(ts.bindingNameA, ts.serviceInstanceNameA, timeout)
}

func (ts *TestSuite) createAndWaitForServiceBindingB(timeout time.Duration) {
	ts.createAndWaitForServiceBinding(ts.bindingNameB, ts.serviceInstanceNameB, timeout)
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

	err = siClient.Delete(bindingName, &metav1.DeleteOptions{})
	require.NoError(ts.t, err)

	repeat.FuncAtMost(ts.t, func() error {
		_, err := siClient.Get(bindingName, metav1.GetOptions{})
		switch {
		case err == nil:
			return fmt.Errorf("ServiceBiding %q still exists", bindingName)
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
	_, err = bindingClient.Create(&scTypes.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingName,
			Namespace: ts.namespace,
		},
		Spec: scTypes.ServiceBindingSpec{
			ServiceInstanceRef: scTypes.LocalObjectReference{
				Name: instanceName,
			},
		},
	})
	require.NoError(ts.t, err)

	repeat.FuncAtMost(ts.t, func() error {
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
func (ts *TestSuite) createBindingUsageForInstanceAWithoutPrefix() {
	ts.bindingUsage(ts.bindingNameA, "binding-usage-tester", "")
}

func (ts *TestSuite) createBindingUsageForInstanceBWithPrefix() {
	ts.bindingUsage(ts.bindingNameB, "binding-usage-tester-with-prefix", ts.envPrefix)
}

func (ts *TestSuite) bindingUsage(bindingName, sbuName, envPrefix string) {
	usageCli := ts.bindingUsageClient()
	sbu := &bucTypes.ServiceBindingUsage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBindingUsage",
			APIVersion: "servicecatalog.kyma.cx/v1alpha1",
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

	_, err := usageCli.Create(sbu)
	require.NoError(ts.t, err)
}

func (ts *TestSuite) bindingUsageClient() bucInterface.ServiceBindingUsageInterface {
	client, err := bucClient.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)
	return client.ServicecatalogV1alpha1().ServiceBindingUsages(ts.namespace)
}

// ServiceInstance helpers
func (ts *TestSuite) createAndWaitForServiceInstanceA(timeout time.Duration) {
	ts.createAndWaitForServiceInstance(ts.serviceInstanceNameA, ts.classExternalNameA, timeout)
}

func (ts *TestSuite) createAndWaitForServiceInstanceB(timeout time.Duration) {
	ts.createAndWaitForServiceInstance(ts.serviceInstanceNameB, ts.classExternalNameB, timeout)
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

	err = siClient.Delete(instanceName, &metav1.DeleteOptions{})
	require.NoError(ts.t, err)

	repeat.FuncAtMost(ts.t, func() error {
		_, err := siClient.Get(instanceName, metav1.GetOptions{})
		switch {
		case err == nil:
			return fmt.Errorf("ServiceInstance %q still exists", instanceName)
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

	_, err = siClient.Create(&scTypes.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: instanceName,
		},
		Spec: scTypes.ServiceInstanceSpec{
			PlanReference: scTypes.PlanReference{
				ServiceClassExternalName: classExternalName,
				ServicePlanExternalName:  "default",
			},
		},
	})
	require.NoError(ts.t, err)

	repeat.FuncAtMost(ts.t, func() error {
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
func (ts *TestSuite) waitForREServiceClasses(timeout time.Duration) {
	repeat.FuncAtMost(ts.t, ts.serviceClassIsAvailableA(), timeout)
	repeat.FuncAtMost(ts.t, ts.serviceClassIsAvailableB(), timeout)
}

func (ts *TestSuite) serviceClassIsAvailableA() func() error {
	clientSet, err := scClient.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)

	return func() error {
		class, err := clientSet.ServicecatalogV1beta1().ServiceClasses(ts.namespace).Get(ts.reSvcNameA, metav1.GetOptions{})
		if err != nil {
			return err
		}
		ts.classExternalNameA = class.Spec.ExternalName
		return nil
	}
}

func (ts *TestSuite) serviceClassIsAvailableB() func() error {
	clientSet, err := scClient.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)

	return func() error {
		class, err := clientSet.ServicecatalogV1beta1().ServiceClasses(ts.namespace).Get(ts.reSvcNameB, metav1.GetOptions{})
		if err != nil {
			return err
		}
		ts.classExternalNameB = class.Spec.ExternalName
		return nil
	}
}

func (ts *TestSuite) cleanup() {
	ts.deleteServiceBindingA(timeoutPerStep)
	ts.deleteServiceBindingB(timeoutPerStep)
	ts.deleteServiceInstanceA(timeoutPerStep)
	ts.deleteServiceInstanceB(timeoutPerStep)
	ts.deleteTestNamespace()
	ts.deleteRemoteEnvironment()
}

func (ts *TestSuite) dumpTestNamespace() {
	clientSet, err := scClient.NewForConfig(ts.k8sClientCfg)
	require.NoError(ts.t, err)

	// AC dump
	re, err := ts.remoteEnvironmentClient().Get(ts.remoteEnvironmentName, metav1.GetOptions{})
	if err != nil {
		ts.t.Logf("Error: %v\n", err)
	}
	ts.t.Logf("RemoteEnvironment: %v\n", re)

	// SC dump
	sb, err := clientSet.ServicecatalogV1beta1().ServiceBindings(ts.namespace).List(metav1.ListOptions{})
	if err != nil {
		ts.t.Logf("Error: %v\n", err)
	}
	ts.t.Logf("ServiceBindings: %v\n", sb.Items)

	si, err := clientSet.ServicecatalogV1beta1().ServiceInstances(ts.namespace).List(metav1.ListOptions{})
	if err != nil {
		ts.t.Logf("Error: %v\n", err)
	}
	ts.t.Logf("ServiceInstances: %v\n", si.Items)

	sc, err := clientSet.ServicecatalogV1beta1().ServiceClasses(ts.namespace).List(metav1.ListOptions{})
	if err != nil {
		ts.t.Logf("Error: %v\n", err)
	}
	ts.t.Logf("ServiceClasses: %v\n", sc.Items)

	sbr, err := clientSet.ServicecatalogV1beta1().ServiceBrokers(ts.namespace).List(metav1.ListOptions{})
	if err != nil {
		ts.t.Logf("Error: %v\n", err)
	}
	ts.t.Logf("ServiceBrokers: %v\n", sbr.Items)

	// SBU dump
	sbu, err := ts.bindingUsageClient().List(metav1.ListOptions{})
	if err != nil {
		ts.t.Logf("Error: %v\n", err)
	}
	ts.t.Logf("ServiceBindingUsages: %v\n", sbu.Items)

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
	_, err = deploymentClient.Create(deploy)
	require.NoError(ts.t, err)

	serviceClient := clientset.CoreV1().Services(ts.namespace)
	_, err = serviceClient.Create(svc)
	require.NoError(ts.t, err)
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

func (ts *TestSuite) assertInjectedEnvVariable(envName string, envValue string, timeout time.Duration) {
	req := fmt.Sprintf("http://acc-test-env-tester.%s.svc.cluster.local/envs?name=%s&value=%s", ts.namespace, envName, envValue)

	repeat.FuncAtMost(ts.t, func() error {
		resp, err := http.Get(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("while checking if proper env is injected, received unexpected status code [got: %d, expected: %d]", resp.StatusCode, http.StatusOK)
		}
		return nil
	}, timeout)
}
