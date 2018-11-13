package apicontroller

import (
	"time"

	"github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	"github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/informers/externalversions"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

type Resolver struct {
	*apiResolver

	informerFactory externalversions.SharedInformerFactory
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration) (*Resolver, error) {
	client, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing clientset")
	}

	informerFactory := externalversions.NewSharedInformerFactory(client, informerResyncPeriod)
	service := newApiService(informerFactory.Gateway().V1alpha2().Apis().Informer())

	return &Resolver{

		apiResolver:     newApiResolver(service),
		informerFactory: informerFactory,
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.informerFactory.Start(stopCh)
	r.informerFactory.WaitForCacheSync(stopCh)
}
