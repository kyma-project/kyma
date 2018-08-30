package nsbroker

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	//scCs "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	scbeta "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	//"k8s.io/client-go/kubernetes"
	typedCorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	//"k8s.io/client-go/tools/clientcmd"
)

const (
	brokerName       = "remote-env-broker"
	brokerLabelKey   = "namespaced-remote-env-broker"
	brokerLabelValue = "true"
)

func main() {
	//cfg, err := clientcmd.BuildConfigFromFlags("", "/Users/i303785/.kube/config")
	//if err != nil {
	//	panic(err)
	//}
	//cliset, err := kubernetes.NewForConfig(cfg)
	//if err != nil {
	//	panic(err)
	//}
	//
	//sccliset, err := scCs.NewForConfig(cfg)
	//if err != nil {
	//	panic(err)
	//}
	////mgr := NewFacade(sccliset.ServicecatalogV1beta1(), cliset.CoreV1(), "app", "core-remote-environment-broker", 8080)
	//
	//sysNs := "kyma-system"
	//testNs := "test"
	//
	//err = mgr.Create(testNs, sysNs)
	//fmt.Println("Create error", err, err == nil)
	//err = filterOutMultiError(err, IgnoreAlreadyExist)
	//fmt.Println("Create error without already exist", err)
	//ex, err := mgr.Exist("test")
	//fmt.Println("Exist: ", ex)
	//fmt.Println("Exist err", err)
	//err = mgr.Delete("test", sysNs)
	//fmt.Println("Delete err", err)
	//err = filterOutMultiError(err, IgnoreIsNotFound)
	//fmt.Println("Delete err without not found", err)
	//time.Sleep(time.Second*5)
	//ex, err = mgr.Exist("test")
	//fmt.Println("Exist", ex)
	//fmt.Println("Exist err", err)

}

type Facade struct {
	brokerGetter     scbeta.ServiceBrokersGetter
	servicesGetter   typedCorev1.ServicesGetter
	rebSelectorKey   string
	rebSelectorValue string
	rebTargetPort    int32
	log              logrus.FieldLogger
}

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

func (f *Facade) Create(destinationNs, systemNs string) error {
	var resultErr error

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
	return resultErr
}

func (f *Facade) Delete(destinationNs, systemNs string) error {
	var resultErr error
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
	return resultErr
}

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

func (f *Facade) filterOutMultiError(maybeMultiError error, predicate func(err error) bool) error {
	if merr, ok := maybeMultiError.(*multierror.Error); ok {
		var out *multierror.Error
		for _, wrapped := range merr.Errors {
			if predicate(wrapped) {
				out = multierror.Append(out, wrapped)
			}
		}
		return out
	}
	return maybeMultiError

}

func (f *Facade) ignoreAlreadyExist(err error) bool {
	return !errors.IsAlreadyExists(err)
}

func (f *Facade) ignoreIsNotFound(err error) bool {
	return !errors.IsNotFound(err)
}
