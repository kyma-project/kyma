package kubeless

import (
	"time"

	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"github.com/kubeless/kubeless/pkg/client/informers/externalversions"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

type Resolver struct {
	*functionResolver

	informerFactory externalversions.SharedInformerFactory
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration) (*Resolver, error) {
	clientset, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while initializing Clientset")
	}

	informerFactory := externalversions.NewSharedInformerFactory(clientset, informerResyncPeriod)
	functionService := newFunctionService(informerFactory.Kubeless().V1beta1().Functions().Informer())

	return &Resolver{
		informerFactory:  informerFactory,
		functionResolver: newFunctionResolver(functionService),
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.informerFactory.Start(stopCh)
	r.informerFactory.WaitForCacheSync(stopCh)
}
