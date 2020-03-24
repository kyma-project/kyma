package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pkg/errors"

	servicecatalogv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogclientset "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"

	appconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	kymaeventingclientset "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset"
)

const (
	appBrokerServiceClass  = "application-broker"
	serviceInstanceKind    = "ServiceInstance"
	eventsServiceEntryType = "Events"
)

type serviceInstancesList []servicecatalogv1beta1.ServiceInstance

type eventActivationsByServiceInstance map[string][]string
type eventActivationsByServiceInstanceAndNamespace map[string]eventActivationsByServiceInstance

// serviceInstanceManager performs operations on ServiceInstances.
type serviceInstanceManager struct {
	svcCatalogClient servicecatalogclientset.Interface
	kymaClient       kymaeventingclientset.Interface
	dynClient        dynamic.Interface

	serviceInstances      serviceInstancesList
	eventActivationsIndex eventActivationsByServiceInstanceAndNamespace
}

// newServiceInstanceManager creates and initializes a serviceInstanceManager.
func newServiceInstanceManager(svcCatalogClient servicecatalogclientset.Interface,
	kymaClient kymaeventingclientset.Interface, dynClient dynamic.Interface,
	namespaces []string) (*serviceInstanceManager, error) {

	m := &serviceInstanceManager{
		svcCatalogClient: svcCatalogClient,
		kymaClient:       kymaClient,
		dynClient:        dynClient,
	}

	if err := m.populateServiceInstances(namespaces); err != nil {
		return nil, errors.Wrap(err, "populating ServiceInstances cache")
	}

	if err := m.populateEventActivationIndex(namespaces); err != nil {
		return nil, errors.Wrap(err, "populating EventActivation index")
	}

	return m, nil
}

// populateServiceInstances populates the local list of ServiceInstances.
// In order to be marked as candidate for re-creation, a ServiceInstance must relate to a ServiceClass which:
//  * is used for events
//  * is handled by the Kyma Application broker
func (m *serviceInstanceManager) populateServiceInstances(namespaces []string) error {
	eventsSvcClasses, err := m.buildServiceClassIndex()
	if err != nil {
		return errors.Wrap(err, "building index of ServiceClasses")
	}

	svcBrokerIndex, err := m.buildServiceBrokerIndex(namespaces)
	if err != nil {
		return errors.Wrap(err, "building index of ServiceBrokers")
	}

	for _, ns := range namespaces {
		svcis, err := m.svcCatalogClient.ServicecatalogV1beta1().ServiceInstances(ns).List(metav1.ListOptions{})
		switch {
		case apierrors.IsNotFound(err):
			return NewTypeNotFoundError(err.(*apierrors.StatusError).ErrStatus.Details.Kind)
		case err != nil:
			return errors.Wrapf(err, "listing ServiceInstances in namespace %s", ns)
		}

		for _, svci := range svcis.Items {
			// Eventing ServiceInstance will only have ServiceClassRef hence rest are ignored
			if svci.Spec.ServiceClassRef == nil {
				continue
			}
			scName := svci.Spec.ServiceClassRef.Name

			_, isEventsClass := eventsSvcClasses[scName]
			_, isAppBrokerClass := svcBrokerIndex[ns][scName]

			if isEventsClass && isAppBrokerClass {
				m.serviceInstances = append(m.serviceInstances, svci)
			}
		}
	}

	return nil
}

type appServiceClasses map[string]struct{}
type appServiceClassesByNamespaces map[string]appServiceClasses

// buildServiceBrokerIndex returns a map of all ServiceClasses associated to the Kyma Application broker, indexed by
// namespaces.
func (m *serviceInstanceManager) buildServiceBrokerIndex(namespaces []string) (appServiceClassesByNamespaces, error) {
	appSvcClassesByNsIndex := make(appServiceClassesByNamespaces)

	for _, ns := range namespaces {
		appSvcClassesIndex := make(appServiceClasses)

		scs, err := m.svcCatalogClient.ServicecatalogV1beta1().ServiceClasses(ns).List(metav1.ListOptions{})
		switch {
		case apierrors.IsNotFound(err):
			return nil, NewTypeNotFoundError(err.(*apierrors.StatusError).ErrStatus.Details.Kind)
		case err != nil:
			return nil, errors.Wrapf(err, "listing ServiceClasses in namespace %s", ns)
		}

		for _, sc := range scs.Items {
			if sc.Spec.ServiceBrokerName == appBrokerServiceClass {
				appSvcClassesIndex[sc.Name] = struct{}{}
			}
		}

		appSvcClassesByNsIndex[ns] = appSvcClassesIndex
	}

	return appSvcClassesByNsIndex, nil
}

type eventsServiceClasses map[string]struct{}

// buildServiceClassIndex returns a map of all ServiceClass UUIDs that are used by Applications for events.
func (m *serviceInstanceManager) buildServiceClassIndex() (eventsServiceClasses, error) {
	eventsSvcClasses := make(eventsServiceClasses)

	appGVR := appconnectorv1alpha1.SchemeGroupVersion.WithResource("applications")
	appCli := m.dynClient.Resource(appGVR)

	apps, err := appCli.List(metav1.ListOptions{})
	switch {
	case apierrors.IsNotFound(err):
		return nil, NewTypeNotFoundError(err.(*apierrors.StatusError).ErrStatus.Details.Kind)
	case err != nil:
		return nil, errors.Wrapf(err, "listing Applications")
	}

	for _, app := range apps.Items {
		appObj, err := ApplicationFromUnstructured(&app)
		if err != nil {
			return nil, err
		}

		for _, svc := range appObj.Spec.Services {
			for _, entry := range svc.Entries {
				if entry.Type == eventsServiceEntryType {
					eventsSvcClasses[svc.ID] = struct{}{}
					break
				}
			}
		}

	}

	return eventsSvcClasses, nil
}

// populateEventActivationIndex populates the local index of EventActivations.
func (m *serviceInstanceManager) populateEventActivationIndex(namespaces []string) error {
	m.eventActivationsIndex = make(eventActivationsByServiceInstanceAndNamespace)

	for _, ns := range namespaces {
		eventActivationsBySvcInstance := make(eventActivationsByServiceInstance)

		eas, err := m.kymaClient.ApplicationconnectorV1alpha1().EventActivations(ns).List(metav1.ListOptions{})
		switch {
		case apierrors.IsNotFound(err):
			return NewTypeNotFoundError(err.(*apierrors.StatusError).ErrStatus.Details.Kind)
		case err != nil:
			return errors.Wrapf(err, "listing EventActivations in namespace %s", ns)
		}

		for _, ea := range eas.Items {
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
			m.eventActivationsIndex[ns] = eventActivationsBySvcInstance
		}
	}

	return nil
}

// isServiceInstanceOwnerReference returns whether the given OwnerReference matches a ServiceInstance.
func isServiceInstanceOwnerReference(ownRef metav1.OwnerReference) bool {
	grp, err := apiGroup(ownRef.APIVersion)
	if err != nil {
		log.Printf("Failed to parse API group: %s", err)
		return false
	}

	return ownRef.Kind == serviceInstanceKind && grp == servicecatalogv1beta1.GroupName
}

// apiGroup returns the API group of a SchemeGroupVersion string.
func apiGroup(groupVersion string) (string, error) {
	elements := strings.Split(groupVersion, "/")
	if len(elements) != 2 {
		return "", errors.Errorf("expected 2 elements in groupVersion %q", groupVersion)
	}
	return elements[0], nil
}

// recreateAll re-creates all ServiceInstance objects listed in the serviceInstanceManager. This ensures the Kyma
// Application broker re-triggers the provisioning of Kyma Applications.
func (m *serviceInstanceManager) recreateAll() error {
	log.Printf("Starting re-creation of %d ServiceInstance(s)", len(m.serviceInstances))

	for _, svci := range m.serviceInstances {
		if err := m.recreateServiceInstance(svci); err != nil {
			return errors.Wrapf(err, "re-creating ServiceInstance %s/%s", svci.Namespace, svci.Name)
		}
	}

	return nil
}

// recreateServiceInstance re-creates a single ServiceInstance object.
func (m *serviceInstanceManager) recreateServiceInstance(svci servicecatalogv1beta1.ServiceInstance) error {
	svciKey := fmt.Sprintf("%s/%s", svci.Namespace, svci.Name)

	log.Printf("+ Deleting ServiceInstance %q", svciKey)

	if err := m.svcCatalogClient.ServicecatalogV1beta1().ServiceInstances(svci.Namespace).
		Delete(svci.Name, &metav1.DeleteOptions{}); err != nil {

		return errors.Wrapf(err, "deleting ServiceInstance %q", svciKey)
	}

	if err := m.waitForServiceInstanceDeletion(svci.Namespace, svci.Name); err != nil {
		return errors.Wrapf(err, "waiting for deletion of ServiceInstance %q", svciKey)
	}

	eventActivationsForServiceInstance := m.eventActivationsIndex[svci.Namespace][svci.Name]
	for _, eaName := range eventActivationsForServiceInstance {
		eaKey := fmt.Sprintf("%s/%s", svci.Namespace, eaName)

		log.Printf("++ Waiting for deletion of EventActivation %q", eaKey)
		if err := m.waitForEventActivationDeletion(svci.Namespace, eaName); err != nil {
			return errors.Wrapf(err, "waiting for deletion of EventActivation %q", eaKey)
		}
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

	log.Printf("+ Re-creating ServiceInstance %q", svciKey)

	if err := m.createServiceInstanceWithRetry(svci); err != nil {
		return errors.Wrapf(err, "creating ServiceInstance %q", svciKey)
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

	return wait.PollImmediateUntil(time.Second, expectNoServiceInstance, newTimeoutChannel())
}

// waitForEventActivationDeletion waits for the deletion of an EventActivation.
func (m *serviceInstanceManager) waitForEventActivationDeletion(ns, name string) error {
	var expectNoEventActivation wait.ConditionFunc = func() (bool, error) {
		_, err := m.kymaClient.ApplicationconnectorV1alpha1().EventActivations(ns).Get(name, metav1.GetOptions{})
		switch {
		case apierrors.IsNotFound(err):
			return true, nil
		case err != nil:
			return false, err
		}
		return false, nil
	}

	return wait.PollImmediateUntil(time.Second, expectNoEventActivation, newTimeoutChannel())
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

	return wait.PollImmediateUntil(5*time.Second, expectSuccessfulServiceInstanceCreation, newTimeoutChannel())
}

// ApplicationFromUnstructured converts an instance of Unstructured to an Application object.
func ApplicationFromUnstructured(app *unstructured.Unstructured) (*appconnectorv1alpha1.Application, error) {
	appObj := &appconnectorv1alpha1.Application{}

	convertCtx := runtime.NewMultiGroupVersioner(appconnectorv1alpha1.SchemeGroupVersion)
	if err := scheme.Scheme.Convert(app, appObj, convertCtx); err != nil {
		return nil, errors.Wrap(err, "converting Unstructured to Application")
	}

	return appObj, nil
}
