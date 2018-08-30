package nsbroker

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scbeta "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	typedCorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	brokerName       = "remote-env-broker"
	brokerLabelKey   = "namespaced-remote-env-broker"
	brokerLabelValue = "true"
)

// Facade is responsible for creation k8s objects for namespaced broker
type Facade struct {
	brokerGetter     scbeta.ServiceBrokersGetter
	servicesGetter   typedCorev1.ServicesGetter
	rebSelectorKey   string
	rebSelectorValue string
	rebTargetPort    int32
	log              logrus.FieldLogger
}

// NewFacade returns facade
func NewFacade(brokerGetter scbeta.ServiceBrokersGetter, servicesGetter typedCorev1.ServicesGetter, rebSelectorKey string, rebSelectorValue string, rebTargetPort int32, log logrus.FieldLogger) *Facade {
	return &Facade{
		brokerGetter:     brokerGetter,
		servicesGetter:   servicesGetter,
		rebSelectorKey:   rebSelectorKey,
		rebSelectorValue: rebSelectorValue,
		rebTargetPort:    rebTargetPort,
		log:              log.WithField("service", "nsbroker-facade"),
	}
}

// Create creates k8s service and ServiceBroker. Errors don't stop execution of method. AlreadyExist errors are ignored.
func (f *Facade) Create(destinationNs, systemNs string) error {
	var resultErr *multierror.Error

	if _, err := f.servicesGetter.Services(systemNs).Create(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.getServiceName(destinationNs),
			Namespace: systemNs,
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
				URL: fmt.Sprintf("http://%s.%s.svc.cluster.local", f.getServiceName(destinationNs), systemNs),
			},
		},
	}
	if _, err := f.brokerGetter.ServiceBrokers(destinationNs).Create(broker); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}

	if resultErr != nil {
		f.log.Warnf("Creation of namespaced-broker [%s] and service for it results in error: [%s]. AlreadyExist errors will be ignored.", destinationNs, resultErr.Error())
	}
	resultErr = f.filterOutMultiError(resultErr, f.ignoreAlreadyExist)

	if resultErr == nil {
		return nil
	}
	return resultErr
}

// Delete removes ServiceBroker and Facade. Errors don't stop execution of method. NotFound errors are ignored.
func (f *Facade) Delete(destinationNs, systemNs string) error {
	var resultErr *multierror.Error
	if err := f.brokerGetter.ServiceBrokers(destinationNs).Delete(brokerName, nil); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}

	if err := f.servicesGetter.Services(systemNs).Delete(f.getServiceName(destinationNs), nil); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}

	if resultErr != nil {
		f.log.Warnf("Deleteion of namespaced-broker [%s] and service for it reults in error: [%s]. NotFound errors will be ignored. ", destinationNs, resultErr.Error())
	}
	resultErr = f.filterOutMultiError(resultErr, f.ignoreIsNotFound)
	if resultErr == nil {
		return nil
	}
	return resultErr
}

// Exist check if ServiceBroker exist.
func (f *Facade) Exist(destinationNs string) (bool, error) {
	_, err := f.brokerGetter.ServiceBrokers(destinationNs).Get(brokerName, metav1.GetOptions{})
	switch {
	case errors.IsNotFound(err):
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}

}

func (f *Facade) getServiceName(ns string) string {
	const serviceNamePattern = "reb-ns-for-%s"
	return fmt.Sprintf(serviceNamePattern, ns)
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
	return !errors.IsAlreadyExists(err)
}

func (f *Facade) ignoreIsNotFound(err error) bool {
	return !errors.IsNotFound(err)
}
