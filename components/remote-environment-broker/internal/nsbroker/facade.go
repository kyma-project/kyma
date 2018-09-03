package nsbroker

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scbeta "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	typedCorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	brokerName       = "remote-env-broker"
	brokerLabelKey   = "namespaced-remote-env-broker"
	brokerLabelValue = "true"
)

//go:generate mockery -name=serviceNameProvider -output=automock -outpkg=automock -case=underscore
type serviceNameProvider interface {
	GetServiceNameForNsBroker(ns string) string
}

// Facade is responsible for creation k8s objects for namespaced broker
type Facade struct {
	brokerGetter        scbeta.ServiceBrokersGetter
	servicesGetter      typedCorev1.ServicesGetter
	serviceNameProvider serviceNameProvider
	workingNamespace    string
	rebSelectorKey      string
	rebSelectorValue    string
	rebTargetPort       int32
	log                 logrus.FieldLogger
}

// NewFacade returns facade
func NewFacade(brokerGetter scbeta.ServiceBrokersGetter, servicesGetter typedCorev1.ServicesGetter, serviceNameProvider serviceNameProvider, workingNamespace, rebSelectorKey, rebSelectorValue string, rebTargetPort int32, log logrus.FieldLogger) *Facade {
	return &Facade{
		brokerGetter:        brokerGetter,
		servicesGetter:      servicesGetter,
		serviceNameProvider: serviceNameProvider,
		rebSelectorKey:      rebSelectorKey,
		rebSelectorValue:    rebSelectorValue,
		rebTargetPort:       rebTargetPort,
		workingNamespace:    workingNamespace,
		log:                 log.WithField("service", "nsbroker-facade"),
	}
}

// Create creates k8s service and ServiceBroker. Errors don't stop execution of method. AlreadyExist errors are ignored.
func (f *Facade) Create(destinationNs string) error {
	var resultErr *multierror.Error

	if _, err := f.servicesGetter.Services(f.workingNamespace).Create(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.serviceNameProvider.GetServiceNameForNsBroker(destinationNs),
			Namespace: f.workingNamespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Selector: map[string]string{
				f.rebSelectorKey: f.rebSelectorValue,
			},
			Ports: []corev1.ServicePort{
				{
					Name: "broker",
					Port: 80,
					TargetPort: intstr.IntOrString{
						IntVal: f.rebTargetPort,
					},
				},
			},
		},
	}); err != nil {
		f.log.Warnf("Creation of service for namespaced-broker for namespace [%s] results in error: [%s]. AlreadyExist error will be ignored.", destinationNs, err)
		resultErr = multierror.Append(resultErr, err)
	}

	broker := &v1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      brokerName,
			Namespace: destinationNs,
			Labels: map[string]string{
				brokerLabelKey: brokerLabelValue,
			},
		},
		Spec: v1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: fmt.Sprintf("http://%s.%s.svc.cluster.local", f.serviceNameProvider.GetServiceNameForNsBroker(destinationNs), f.workingNamespace),
			},
		},
	}
	if _, err := f.brokerGetter.ServiceBrokers(destinationNs).Create(broker); err != nil {
		resultErr = multierror.Append(resultErr, err)
		f.log.Warnf("Creation of namespaced-broker for namespace [%s] results in error: [%s]. AlreadyExist errors will be ignored.", destinationNs, err)
	}

	resultErr = f.filterOutMultiError(resultErr, f.ignoreAlreadyExist)

	if resultErr == nil {
		return nil
	}
	return resultErr
}

// Delete removes ServiceBroker and Facade. Errors don't stop execution of method. NotFound errors are ignored.
func (f *Facade) Delete(destinationNs string) error {
	var resultErr *multierror.Error
	if err := f.brokerGetter.ServiceBrokers(destinationNs).Delete(brokerName, nil); err != nil {
		f.log.Warnf("Deletion of namespaced-broker for namespace [%s] results in error: [%s]. NotFound errors will be ignored. ", destinationNs, err)
		resultErr = multierror.Append(resultErr, err)
	}

	if err := f.servicesGetter.Services(f.workingNamespace).Delete(f.serviceNameProvider.GetServiceNameForNsBroker(destinationNs), nil); err != nil {
		f.log.Warnf("Deletion of service for namespaced-broker for namespace [%s] results in error: [%s]. NotFound errors will be ignored. ", destinationNs, err)
		resultErr = multierror.Append(resultErr, err)
	}

	resultErr = f.filterOutMultiError(resultErr, f.ignoreIsNotFound)
	if resultErr == nil {
		return nil
	}
	return resultErr
}

// Exist check if ServiceBroker and Service exist.
func (f *Facade) Exist(destinationNs string) (bool, error) {
	_, err := f.brokerGetter.ServiceBrokers(destinationNs).Get(brokerName, metav1.GetOptions{})
	switch {
	case k8serrors.IsNotFound(err):
		return false, nil
	case err != nil:
		return false, errors.Wrapf(err, "while checking if ServiceBroker [%s] exists in the namespace [%s]", brokerName, destinationNs)
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

func (f *Facade) filterOutMultiError(merr *multierror.Error, predicate func(err error) bool) *multierror.Error {
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

func (f *Facade) ignoreAlreadyExist(err error) bool {
	return !k8serrors.IsAlreadyExists(err)
}

func (f *Facade) ignoreIsNotFound(err error) bool {
	return !k8serrors.IsNotFound(err)
}
