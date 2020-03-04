package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"

	servicecatalogv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogclientset "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const appBrokerServiceClass = "application-broker"

type serviceClassByNamespaceMap map[string][]servicecatalogv1beta1.ServiceClass
type serviceInstanceList []servicecatalogv1beta1.ServiceInstance

// serviceInstanceManager performs operations on ServiceInstances.
type serviceInstanceManager struct {
	cli servicecatalogclientset.Interface

	serviceClassByNamespace serviceClassByNamespaceMap
	serviceInstances        serviceInstanceList
}

// newServiceInstanceManager creates and initializes a serviceInstanceManager.
func newServiceInstanceManager(cli servicecatalogclientset.Interface, namespaces []string) (*serviceInstanceManager, error) {
	m := &serviceInstanceManager{
		cli: cli,
	}

	if err := m.populateServiceClasses(namespaces); err != nil {
		return nil, err
	}

	if err := m.populateServiceInstances(namespaces); err != nil {
		return nil, err
	}

	return m, nil
}

// populateServiceInstances populates the local list of ServiceInstances. Only ServiceInstances related to the
// Application broker are taken into account.
func (m *serviceInstanceManager) populateServiceInstances(namespaces []string) error {
	var serviceInstances []servicecatalogv1beta1.ServiceInstance

	for _, ns := range namespaces {
		svcis, err := m.cli.ServicecatalogV1beta1().ServiceInstances(ns).List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrapf(err, "listing ServiceInstances in namespace %s", ns)
		}
		for _, svci := range svcis.Items {
			svcClassName := svci.Spec.ServiceClassRef.Name
			for _, sc := range m.serviceClassByNamespace[ns] {
				if sc.Name == svcClassName && sc.Spec.ServiceBrokerName == appBrokerServiceClass {
					serviceInstances = append(serviceInstances, svci)
				}
			}
		}
	}

	m.serviceInstances = serviceInstances

	return nil
}

// populateServiceClasses populates the local serviceClassByNamespace map.
func (m *serviceInstanceManager) populateServiceClasses(namespaces []string) error {
	serviceClassByNamespace := make(serviceClassByNamespaceMap)

	for _, ns := range namespaces {
		serviceClass, err := m.cli.ServicecatalogV1beta1().ServiceClasses(ns).List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrapf(err, "listing ServiceClasses in namespace %s", ns)
		}
		serviceClassByNamespace[ns] = serviceClass.Items
	}

	m.serviceClassByNamespace = serviceClassByNamespace

	return nil
}

// recreateAll re-creates all ServiceInstance objects listed in the serviceInstanceManager. This ensures the Kyma
// Application broker re-triggers the provisioning of Kyma Applications.
func (m *serviceInstanceManager) recreateAll() error {
	log.Printf("Starting re-creation of %d ServiceInstances", len(m.serviceInstances))

	for _, svci := range m.serviceInstances {
		if err := m.recreateServiceInstance(svci); err != nil {
			return errors.Wrapf(err, "re-creating ServiceInstance %s/%s", svci.Namespace, svci.Name)
		}
	}

	return nil
}

// recreateServiceInstance re-creates a single ServiceInstance object.
func (m *serviceInstanceManager) recreateServiceInstance(svci servicecatalogv1beta1.ServiceInstance) error {
	objKey := fmt.Sprintf("%s/%s", svci.Namespace, svci.Name)

	log.Printf("Deleting ServiceInstance %q", objKey)

	if err := m.cli.ServicecatalogV1beta1().ServiceInstances(svci.Namespace).
		Delete(svci.Name, &metav1.DeleteOptions{}); err != nil {

		return errors.Wrapf(err, "deleting ServiceInstance %q", objKey)
	}

	if err := m.waitForServiceInstanceDeletion(svci.Namespace, svci.Name); err != nil {
		return errors.Wrapf(err, "waiting for deletion of ServiceInstance %q", objKey)
	}

	// Sanitize the ServiceInstance to avoid the following error from the webhook
	//
	//   Error creating service instance: xyz-1234 in namespace: xyz with error: admission webhook
	//   "validating.serviceinstances.servicecatalog.k8s.io" denied the request: [spec.serviceClassRef:
	//   Forbidden: serviceClassRef must not be present on create, spec.servicePlanRef: Forbidden:
	//   servicePlanRef must not be present on create]
	//
	svci.Spec.ServiceClassRef = nil
	svci.Spec.ServicePlanRef = nil
	svci.ResourceVersion = ""

	log.Printf("Re-creating ServiceInstance %q", objKey)

	if err := m.createServiceInstanceWithRetry(svci); err != nil {
		return errors.Wrapf(err, "creating ServiceInstance %q", objKey)
	}

	return nil
}

// waitForServiceInstanceDeletion waits for the deletion of a ServiceInstance.
func (m *serviceInstanceManager) waitForServiceInstanceDeletion(ns, name string) error {
	var expectNoServiceInstance wait.ConditionFunc = func() (bool, error) {
		_, err := m.cli.ServicecatalogV1beta1().ServiceInstances(ns).Get(name, metav1.GetOptions{})
		switch {
		case apierrors.IsNotFound(err):
			return true, nil
		case err != nil:
			return false, err
		}
		return false, nil
	}

	return wait.PollImmediateUntil(time.Second, expectNoServiceInstance, make(<-chan struct{}))
}

// createServiceInstanceWithRetry creates a ServiceInstance and retries in case of failure.
func (m *serviceInstanceManager) createServiceInstanceWithRetry(svci servicecatalogv1beta1.ServiceInstance) error {
	var expectSuccessfulServiceInstanceCreation wait.ConditionFunc = func() (bool, error) {
		_, err := m.cli.ServicecatalogV1beta1().ServiceInstances(svci.Namespace).Create(&svci)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return false, nil
		}
		return true, nil
	}

	return wait.PollImmediateUntil(5*time.Second, expectSuccessfulServiceInstanceCreation, make(<-chan struct{}))
}
