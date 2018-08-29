package main

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scbeta "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	typedCorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	scCs "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
)

const (
	brokerName             = "remote-env-broker"
	serviceNamePattern     = "reb-ns-for-%"
	brokerLabelKey         = "namespaced-remote-env-broker"
	brokerLabelValue       = "true"
	annotationCreatedByKey = "createdBy"
	annotationCreatedByVal = "remote-environment-broker"
)

func main() {
	cfg, err := clientcmd.BuildConfigFromFlags("", "/Users/i303785/.kube/config")
	if err != nil {
		panic(err)
	}
	cliset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}

	sccliset, err := scCs.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}
	mgr := NewManager(sccliset.ServicecatalogV1beta1(), cliset.CoreV1(), "app", "core-remote-environment-broker", 8080)

	sysNs := "kyma-test"
	testNs := "test"

	if err := mgr.Create(testNs, sysNs); err != nil {
		panic(err)
	}

	if ex, err := mgr.Exist("test"); err != nil {
		panic(err)
	} else {
		fmt.Println("exist1 :", ex)
	}
	//
	//if err := mgr.Delete("test"); err != nil {
	//	panic(err)
	//}
	//
	//if  ex, err := mgr.Exist("test"); err != nil {
	//	panic(err)
	//} else {
	//	fmt.Println("exist1 :", ex)
	//}

}

type Manager struct {
	brokerGetter     scbeta.ServiceBrokersGetter
	servicesGetter   typedCorev1.ServicesGetter
	rebSelectorKey   string
	rebSelectorValue string
	rebTargetPort    int32
}

func NewManager(brokerGetter scbeta.ServiceBrokersGetter, servicesGetter typedCorev1.ServicesGetter, rebSelectorKey string, rebSelectorValue string, rebTargetPort int32) *Manager {
	return &Manager{
		brokerGetter:     brokerGetter,
		servicesGetter:   servicesGetter,
		rebSelectorKey:   rebSelectorKey,
		rebSelectorValue: rebSelectorValue,
		rebTargetPort:    rebTargetPort,
	}
}

func (m *Manager) Create(forNs, kymaSystemNs string) error {
	serviceName := fmt.Sprintf(serviceNamePattern, forNs)

	_, err := m.servicesGetter.Services(kymaSystemNs).Create(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: kymaSystemNs,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Selector: map[string]string{
				m.rebSelectorKey: m.rebSelectorValue,
			},
			Ports: []corev1.ServicePort{
				{
					Name: "broker",
					Port: 80,
					TargetPort: intstr.IntOrString{
						IntVal: m.rebTargetPort,
					},
				},
			},
		},
	})

	switch {
	case errors.IsAlreadyExists(err):
		// if already exist, we allow to create broker
	case err != nil:
		return err
	}

	if err != nil {
		return err
	}

	broker := &v1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      brokerName,
			Namespace: forNs,
			Labels: map[string]string{
				brokerLabelKey: brokerLabelValue,
			},
		},
		Spec: v1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: fmt.Sprintf("http://%s.%s.svc.cluster.local", serviceName, kymaSystemNs),
			},
		},
	}
	_, err = m.brokerGetter.ServiceBrokers(forNs).Create(broker)
	switch {
	case errors.IsAlreadyExists(err):
		return nil
	case err != nil:
		return err
	}

	return nil
}

func (m *Manager) Delete(forNs, kymaSystemNs string) error {
	err := m.brokerGetter.ServiceBrokers(forNs).Delete(brokerName, nil)
	switch {
	case errors.IsNotFound(err):
	case err != nil:
		return err
	}

	serviceName := fmt.Sprintf(serviceNamePattern, forNs)
	err = m.servicesGetter.Services(forNs).Delete(serviceName, nil);
	switch {
	case errors.IsNotFound(err):
		return nil
	case err != nil:
		return err
	}
	return nil
}

func (m *Manager) Exist(forNs string) (bool, error) {
	_, err := m.brokerGetter.ServiceBrokers(forNs).Get(brokerName, metav1.GetOptions{})
	switch {
	case errors.IsNotFound(err):
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}

}
