package helpers

import (
	"k8s.io/client-go/kubernetes"

	ebClientSet "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset"
	"github.com/sirupsen/logrus"
)

type setupFlow struct {
	k8sInterface kubernetes.Interface
	ebInterface  ebClientSet.Interface

	namespace string
	log       logrus.FieldLogger
	stop      <-chan struct{}
}

func newSetupFlow(k8sInterface kubernetes.Interface, ebInterface ebClientSet.Interface,
	stop <-chan struct{}, log logrus.FieldLogger, namespace string) *setupFlow {
	return &setupFlow{k8sInterface: k8sInterface, ebInterface: ebInterface, stop: stop, log: log, namespace: namespace}
}

//
//func (f *setupFlow) createSubscriber() error {
//	return nil
//}
