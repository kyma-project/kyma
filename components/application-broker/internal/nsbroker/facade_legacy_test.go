package nsbroker_test

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scbeta "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	typedCorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	serviceBrokerAPIVersion = "apiextensions.k8s.io/v1beta1"
)

//go:generate mockery -name=serviceNameProvider -output=automock -outpkg=automock -case=underscore
type serviceNameProvider interface {
	GetServiceNameForNsBroker(ns string) string
}

// legacyFacade is responsible for creation k8s objects for namespaced broker
// like it was done in production code before a switch to one k8s service for the application broker.
// The implementation is a copy of old Facade with removed unnecessary code (broker sync, logging)
// The legacyFacade is used to do a setup in the MigrationService unit tests.
type legacyFacade struct {
	brokerGetter        scbeta.ServiceBrokersGetter
	servicesGetter      typedCorev1.ServicesGetter
	serviceNameProvider serviceNameProvider
	workingNamespace    string
	abSelectorKey       string
	abSelectorValue     string
	abTargetPort        int32
}

// newLegacyFacade returns facade
func newLegacyFacade(brokerGetter scbeta.ServiceBrokersGetter,
	servicesGetter typedCorev1.ServicesGetter,
	serviceNameProvider serviceNameProvider,
	workingNamespace, abSelectorKey, abSelectorValue string,
	abTargetPort int32) *legacyFacade {
	return &legacyFacade{
		brokerGetter:        brokerGetter,
		servicesGetter:      servicesGetter,
		serviceNameProvider: serviceNameProvider,
		abSelectorKey:       abSelectorKey,
		abSelectorValue:     abSelectorValue,
		abTargetPort:        abTargetPort,
		workingNamespace:    workingNamespace,
	}
}

// Create creates k8s service and ServiceBroker. Errors don't stop execution of method. AlreadyExist errors are ignored.
func (f *legacyFacade) Create(destinationNs string) error {
	var resultErr *multierror.Error

	svcURL := fmt.Sprintf("http://%s.%s.svc.cluster.local", f.serviceNameProvider.GetServiceNameForNsBroker(destinationNs), f.workingNamespace)
	createdBroker, err := f.createServiceBroker(svcURL, destinationNs)
	if err != nil {
		resultErr = multierror.Append(resultErr, err)
		return err
	}

	if _, err := f.servicesGetter.Services(f.workingNamespace).Create(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.serviceNameProvider.GetServiceNameForNsBroker(destinationNs),
			Namespace: f.workingNamespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: serviceBrokerAPIVersion,
					Kind:       "ServiceBroker",
					Name:       createdBroker.Name,
					UID:        createdBroker.UID,
				},
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Selector: map[string]string{
				f.abSelectorKey: f.abSelectorValue,
			},
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 80,
					TargetPort: intstr.IntOrString{
						IntVal: f.abTargetPort,
					},
				},
			},
		},
	}); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}

	resultErr = f.filterOutMultiError(resultErr, f.ignoreAlreadyExist)

	if resultErr == nil {
		return nil
	}
	return resultErr
}

// createServiceBroker returns just created or existing ServiceBroker
func (f *legacyFacade) createServiceBroker(svcURL, namespace string) (*v1beta1.ServiceBroker, error) {
	broker := &v1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsbroker.NamespacedBrokerName,
			Namespace: namespace,
			Labels: map[string]string{
				nsbroker.BrokerLabelKey: nsbroker.BrokerLabelValue,
			},
		},
		Spec: v1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: svcURL,
			},
		},
	}

	createdBroker, err := f.brokerGetter.ServiceBrokers(namespace).Create(broker)
	if k8serrors.IsAlreadyExists(err) {
		createdBroker, err = f.brokerGetter.ServiceBrokers(namespace).Get(nsbroker.NamespacedBrokerName, metav1.GetOptions{})
		return createdBroker, err
	}

	return createdBroker, err
}

// Delete removes ServiceBroker and legacyFacade. Errors don't stop execution of method. NotFound errors are ignored.
func (f *legacyFacade) Delete(destinationNs string) error {
	var resultErr *multierror.Error
	if err := f.brokerGetter.ServiceBrokers(destinationNs).Delete(nsbroker.NamespacedBrokerName, nil); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}

	if err := f.servicesGetter.Services(f.workingNamespace).Delete(f.serviceNameProvider.GetServiceNameForNsBroker(destinationNs), nil); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}

	resultErr = f.filterOutMultiError(resultErr, f.ignoreIsNotFound)
	if resultErr == nil {
		return nil
	}
	return resultErr
}

// Exist check if ServiceBroker and Service exist.
func (f *legacyFacade) Exist(destinationNs string) (bool, error) {
	_, err := f.brokerGetter.ServiceBrokers(destinationNs).Get(nsbroker.NamespacedBrokerName, metav1.GetOptions{})
	switch {
	case k8serrors.IsNotFound(err):
		return false, nil
	case err != nil:
		return false, errors.Wrapf(err, "while checking if ServiceBroker [%s] exists in the namespace [%s]", nsbroker.NamespacedBrokerName, destinationNs)
	}

	svcName := f.serviceNameProvider.GetServiceNameForNsBroker(destinationNs)
	_, err = f.servicesGetter.Services(f.workingNamespace).Get(svcName, metav1.GetOptions{})
	switch {
	case k8serrors.IsNotFound(err):
		return false, nil
	case err != nil:
		return false, errors.Wrapf(err, "while checking if Service [%s] for ServiceBroker exists in the namespace [%s]", svcName, f.workingNamespace)

	}
	return true, nil
}

func (f *legacyFacade) filterOutMultiError(merr *multierror.Error, predicate func(err error) bool) *multierror.Error {
	if merr == nil {
		return nil
	}
	var out *multierror.Error
	for _, wrapped := range merr.Errors {
		if predicate(wrapped) {
			out = multierror.Append(out, wrapped)
		}
	}
	return out

}

func (f *legacyFacade) ignoreAlreadyExist(err error) bool {
	return !k8serrors.IsAlreadyExists(err)
}

func (f *legacyFacade) ignoreIsNotFound(err error) bool {
	return !k8serrors.IsNotFound(err)
}
