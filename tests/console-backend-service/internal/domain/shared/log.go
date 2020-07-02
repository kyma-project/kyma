package shared

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions/printers"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	serviceCatalogLabelSelector = "app=service-catalog-catalog-controller-manager"
	helmBrokerLabelSelector     = "app=helm-broker"
)

func LogReport(instanceName, namespace string, svcatCli *clientset.Clientset, k8sClient *v1.CoreV1Client) {
	serviceInstance, err := svcatCli.ServicecatalogV1beta1().ServiceInstances(namespace).Get(instanceName, metav1.GetOptions{})
	Report("ServiceInstances", serviceInstance, err)

	clusterServiceBrokers, err := svcatCli.ServicecatalogV1beta1().ClusterServiceBrokers().List(metav1.ListOptions{})
	Report("ClusterServiceBrokers", clusterServiceBrokers, err)

	serviceBrokers, err := svcatCli.ServicecatalogV1beta1().ServiceBrokers(namespace).List(metav1.ListOptions{})
	Report("ServiceBrokers", serviceBrokers, err)

	helm, err := k8sClient.Pods("kyma-system").List(metav1.ListOptions{LabelSelector: helmBrokerLabelSelector})
	Report("Helm Broker Pod", helm, err)

	catalog, err := k8sClient.Pods("kyma-system").List(metav1.ListOptions{LabelSelector: serviceCatalogLabelSelector})
	Report("Service Catalog Pod", catalog, err)
}

func Report(kind string, obj runtime.Object, err error) {
	printer := &printers.JSONPrinter{}
	logs := logrus.New()
	logger := &logWriter{log: logs}
	obj.(printerObject).SetGroupVersionKind(schema.GroupVersionKind{Kind: kind})
	if err != nil {
		logs.Errorf("Could not fetch resources: %v", err)
		return
	}
	err = printer.PrintObj(obj, logger)
	if err != nil {
		logs.Errorf("Could not print objects: %v", err)
	}
}

type logWriter struct {
	log logrus.FieldLogger
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.log.Infof(string(p))
	return len(p), nil
}

type printerObject interface {
	SetGroupVersionKind(gvk schema.GroupVersionKind)
}
