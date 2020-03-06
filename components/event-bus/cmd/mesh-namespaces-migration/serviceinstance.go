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

	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/applicationconnector/v1alpha1"
	kymaeventingclientset "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset"
)

const (
	appBrokerServiceClass = "application-broker"
	serviceInstanceKind   = "ServiceInstance"
)

type serviceInstanceList []servicecatalogv1beta1.ServiceInstance

type eventActivationsByServiceInstance map[string][]string
type eventActivationsByServiceInstanceAndNamespace map[string]eventActivationsByServiceInstance

// serviceInstanceManager performs operations on ServiceInstances.
type serviceInstanceManager struct {
	svcCatalogClient servicecatalogclientset.Interface
	kymaClient       kymaeventingclientset.Interface

	serviceInstances     serviceInstanceList
	eventActivationIndex eventActivationsByServiceInstanceAndNamespace
}

// newServiceInstanceManager creates and initializes a serviceInstanceManager.
func newServiceInstanceManager(svcCatalogClient servicecatalogclientset.Interface,
	kymaClient kymaeventingclientset.Interface, namespaces []string) (*serviceInstanceManager, error) {

	m := &serviceInstanceManager{
		svcCatalogClient: svcCatalogClient,
		kymaClient:       kymaClient,
	}

	if err := m.populateServiceInstances(namespaces); err != nil {
		return nil, errors.Wrap(err, "populating ServiceInstances cache")
	}

	if err := m.populateEventActivationIndex(namespaces); err != nil {
		return nil, errors.Wrap(err, "populating EventActivation index")
	}

	return m, nil
}

// populateServiceInstances populates the local list of ServiceInstances. Only ServiceInstances related to the
// Application broker are taken into account.
func (m *serviceInstanceManager) populateServiceInstances(namespaces []string) error {
	svcBrokerIndex, err := m.buildServiceBrokerIndex(namespaces)
	if err != nil {
		return errors.Wrap(err, "building index of ServiceBrokers")
	}

	for _, ns := range namespaces {
		svcis, err := m.svcCatalogClient.ServicecatalogV1beta1().ServiceInstances(ns).List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrapf(err, "listing ServiceInstances in namespace %s", ns)
		}

		for _, svci := range svcis.Items {
			scName := svci.Spec.ServiceClassRef.Name

			if svcBrokerName := svcBrokerIndex[ns][scName]; svcBrokerName == appBrokerServiceClass {
				m.serviceInstances = append(m.serviceInstances, svci)
			}
		}
	}

	return nil
}

type serviceBrokersByServiceClass map[string]string
type serviceBrokersByServiceClassAndNamespace map[string]serviceBrokersByServiceClass

// buildServiceBrokerIndex returns an map of ServiceBrokers names indexed by ServiceClass and namespace.
func (m *serviceInstanceManager) buildServiceBrokerIndex(namespaces []string) (serviceBrokersByServiceClassAndNamespace, error) {
	svcBrokerIndex := make(serviceBrokersByServiceClassAndNamespace)

	for _, ns := range namespaces {
		svcBrokersBySvcClass := make(serviceBrokersByServiceClass)

		serviceClasses, err := m.svcCatalogClient.ServicecatalogV1beta1().ServiceClasses(ns).List(metav1.ListOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "listing ServiceClasses in namespace %s", ns)
		}

		for _, sc := range serviceClasses.Items {
			svcBrokersBySvcClass[sc.Name] = sc.Spec.ServiceBrokerName
		}

		svcBrokerIndex[ns] = svcBrokersBySvcClass
	}

	return svcBrokerIndex, nil
}

// populateEventActivationIndex populates the local index of EventActivations.
func (m *serviceInstanceManager) populateEventActivationIndex(namespaces []string) error {
	m.eventActivationIndex = make(eventActivationsByServiceInstanceAndNamespace)

	for _, ns := range namespaces {
		eventActivationsBySvcInstance := make(eventActivationsByServiceInstance)

		eventActivations, err := m.kymaClient.ApplicationconnectorV1alpha1().EventActivations(ns).List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrapf(err, "listing EventActivations in namespace %s", ns)
		}

		for _, ea := range eventActivations.Items {
			for _, ownRef := range ea.OwnerReferences {
				if isServiceInstanceOwnerReference(ownRef) {
					eventActivationsBySvcInstance[ownRef.Name] = append(
						eventActivationsBySvcInstance[ownRef.Name],
						ea.Name,
					)
				}
			}
		}

		if len(eventActivationsBySvcInstance) != 0 {
			m.eventActivationIndex[ns] = eventActivationsBySvcInstance
		}
	}

	return nil
}

// isServiceInstanceOwnerReference returns whether the given OwnerReference matches a ServiceInstance.
func isServiceInstanceOwnerReference(ownRef metav1.OwnerReference) bool {
	if ownRef.Kind == serviceInstanceKind && ownRef.APIVersion == serviceInstanceAPIVersion() {
		return true
	}
	return false
}

// serviceInstanceAPIVersion returns the group and version of the ServiceInstance type.
func serviceInstanceAPIVersion() string {
	return applicationconnectorv1alpha1.SchemeGroupVersion.String()
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

	// ensures the ServiceInstance disappears only once all its children
	// have been deleted (EventActivations)
	foregroundDelete := metav1.DeletePropagationForeground

	if err := m.svcCatalogClient.ServicecatalogV1beta1().ServiceInstances(svci.Namespace).
		Delete(svci.Name, &metav1.DeleteOptions{PropagationPolicy: &foregroundDelete}); err != nil {

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
		_, err := m.svcCatalogClient.ServicecatalogV1beta1().ServiceInstances(ns).Get(name, metav1.GetOptions{})
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
		_, err := m.svcCatalogClient.ServicecatalogV1beta1().ServiceInstances(svci.Namespace).Create(&svci)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return false, nil
		}
		return true, nil
	}

	return wait.PollImmediateUntil(5*time.Second, expectSuccessfulServiceInstanceCreation, make(<-chan struct{}))
}
