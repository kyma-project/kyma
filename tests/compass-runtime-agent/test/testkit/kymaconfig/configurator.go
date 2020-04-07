package kymaconfig

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"

	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	serviceCatalogApi "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit"

	"github.com/stretchr/testify/require"
)

const (
	serviceClassWaitTime = 30 * time.Second

	defaultCheckInterval         = 3 * time.Second
	serviceInstanceCheckInterval = 2 * time.Second
	serviceInstanceWait          = 60 * time.Second

	serviceInstanceDeletionWaitTime = 30 * time.Second
)

// KymaConfigurator configures Compass Applications to be usable from Kyma
type KymaConfigurator struct {
	namespace                string
	applicationMappingClient v1alpha1.ApplicationMappingInterface
	serviceClassClient       serviceCatalogClient.ServiceClassInterface
	serviceInstanceClient    serviceCatalogClient.ServiceInstanceInterface
	serviceBindingClient     serviceCatalogClient.ServiceBindingInterface
}

func NewKymaConfigurator(namespace string,
	appMappingClient v1alpha1.ApplicationMappingInterface,
	serviceClassClient serviceCatalogClient.ServiceClassInterface,
	serviceInstanceClient serviceCatalogClient.ServiceInstanceInterface,
	serviceBindingClient serviceCatalogClient.ServiceBindingInterface) *KymaConfigurator {

	return &KymaConfigurator{
		namespace:                namespace,
		applicationMappingClient: appMappingClient,
		serviceClassClient:       serviceClassClient,
		serviceInstanceClient:    serviceInstanceClient,
		serviceBindingClient:     serviceBindingClient,
	}
}

type SecretMapping struct {
	PackagesSecrets map[string]string
}

// ConfigureApplication provisions Service Instance and Service Binding for each API package inside the Application
// This results in Application Broker downloading credentials from Director
func (c *KymaConfigurator) ConfigureApplication(t *testing.T, log *testkit.Logger, applicationId string, packages []*graphql.PackageExt) SecretMapping {
	secretMapping := SecretMapping{
		PackagesSecrets: make(map[string]string, len(packages)),
	}

	for _, pkg := range packages {
		log := log.NewExtended(map[string]string{"APIPackageID": pkg.ID, "APIPackageName": pkg.Name})

		svcClass := c.waitForServiceClass(t, applicationId)
		log.Log(fmt.Sprintf("Service class %s created", svcClass.Name))

		svcInstance, err := c.createServiceInstance(applicationId, pkg.ID, pkg.ID)
		require.NoError(t, err)

		err = c.waitForServiceInstance(log, svcInstance.Name)
		require.NoError(t, err)
		log.Log(fmt.Sprintf("Service Instance %s created", svcInstance.Name))

		serviceBinding, err := c.createServiceBinding(pkg.ID, svcInstance)
		require.NoError(t, err)

		err = c.waitForServiceBinding(log, serviceBinding.Name)
		require.NoError(t, err)
		log.Log(fmt.Sprintf("Service Binding %s created", serviceBinding.Name))

		secretMapping.PackagesSecrets[pkg.ID] = serviceBinding.Name
	}

	return secretMapping
}

func (c *KymaConfigurator) createApplicationMapping(appName string) error {
	applicationMapping := &applicationconnectorv1alpha1.ApplicationMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApplicationMapping",
			APIVersion: applicationconnectorv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: c.namespace,
		},
		Spec: applicationconnectorv1alpha1.ApplicationMappingSpec{},
	}

	_, err := c.applicationMappingClient.Create(applicationMapping)
	return err
}

func (c *KymaConfigurator) deleteApplicationMapping(appName string) error {
	return c.applicationMappingClient.Delete(appName, &metav1.DeleteOptions{})
}

func (c *KymaConfigurator) waitForServiceInstance(log *testkit.Logger, name string) error {
	return testkit.WaitForFunction(serviceInstanceCheckInterval, serviceInstanceWait, func() bool {
		err := c.isServiceInstanceCreated(name)
		if err != nil {
			log.Log(fmt.Sprintf("Service instance not ready: %s", err.Error()))
			return false
		}

		return true
	})
}

func (c *KymaConfigurator) waitForServiceBinding(log *testkit.Logger, name string) error {
	return testkit.WaitForFunction(serviceInstanceCheckInterval, serviceInstanceWait, func() bool {
		err := c.isServiceBindingReady(name)
		if err != nil {
			log.Log(fmt.Sprintf("Service binding not ready: %s", err.Error()))
			return false
		}

		return true
	})
}

func (c *KymaConfigurator) isServiceInstanceCreated(name string) error {
	svcInstance, err := c.serviceInstanceClient.Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if svcInstance.Status.ProvisionStatus != "Provisioned" {
		return errors.Errorf("unexpected provision status: %s", svcInstance.Status.ProvisionStatus)
	}
	return nil
}

func (c *KymaConfigurator) isServiceBindingReady(name string) error {
	sb, err := c.serviceBindingClient.Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	for _, condition := range sb.Status.Conditions {
		if condition.Type == serviceCatalogApi.ServiceBindingConditionReady {
			if condition.Status != serviceCatalogApi.ConditionTrue {
				return errors.New("ServiceBinding is not ready")
			}
			break
		}
	}
	return nil
}

func (c *KymaConfigurator) CleanupConfiguration(t *testing.T, packages ...*graphql.PackageExt) {
	for _, pkg := range packages {
		err := c.serviceBindingClient.Delete(pkg.ID, &metav1.DeleteOptions{})
		assert.NoError(t, err)

		err = c.serviceInstanceClient.Delete(pkg.ID, &metav1.DeleteOptions{})
		assert.NoError(t, err)
	}

	for _, pkg := range packages {
		err := testkit.WaitForFunction(defaultCheckInterval, serviceInstanceDeletionWaitTime, func() bool {
			_, err := c.serviceInstanceClient.Get(pkg.ID, metav1.GetOptions{})
			if err == nil {
				return false
			} else if k8sErrors.IsNotFound(err) {
				return true
			}

			return false
		})
		assert.NoError(t, err)
	}
}

func (c *KymaConfigurator) waitForServiceClass(t *testing.T, serviceClassName string) *serviceCatalogApi.ServiceClass {
	t.Logf("Waiting for %s Service Class...", serviceClassName)

	var svcClass *serviceCatalogApi.ServiceClass
	var err error

	err = testkit.WaitForFunction(defaultCheckInterval, serviceClassWaitTime, func() bool {
		svcClass, err = c.serviceClassClient.Get(serviceClassName, metav1.GetOptions{})
		if err != nil {
			t.Logf("Service Class %s not ready: %s. Retrying until timeout is reached...", serviceClassName, err.Error())
			return false
		}

		return true
	})
	require.NoError(t, err)

	return svcClass
}

func (c *KymaConfigurator) createServiceInstance(appId, apiPackageId, apiPackageName string) (*v1beta1.ServiceInstance, error) {
	serviceInstance := &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:       apiPackageName,
			Finalizers: []string{"kubernetes-incubator/service-catalog"},
		},
		Spec: v1beta1.ServiceInstanceSpec{
			Parameters: &runtime.RawExtension{},
			PlanReference: serviceCatalogApi.PlanReference{
				ServiceClassExternalID: appId,
				ServicePlanExternalID:  apiPackageId,
			},
			UpdateRequests: 0,
		},
	}

	return c.serviceInstanceClient.Create(serviceInstance)
}

func (c *KymaConfigurator) createServiceBinding(apiPkgName string, svcInstance *v1beta1.ServiceInstance) (*v1beta1.ServiceBinding, error) {
	serviceBinding := &v1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{Name: apiPkgName, Namespace: c.namespace},
		Spec: v1beta1.ServiceBindingSpec{
			InstanceRef: serviceCatalogApi.LocalObjectReference{Name: svcInstance.Name},
		},
	}

	return c.serviceBindingClient.Create(serviceBinding)
}
