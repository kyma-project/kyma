package ui

import (
	"time"

	"github.com/kyma-project/kyma/components/idppreset/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/idppreset/pkg/client/informers/externalversions"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

type Resolver struct {
	*idpPresetResolver

	informerFactory externalversions.SharedInformerFactory
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration) (*Resolver, error) {
	client, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Clientset")
	}

	informerFactory := externalversions.NewSharedInformerFactory(client, informerResyncPeriod)
	idpPresetGroup := informerFactory.Ui().V1alpha1()

	svc := newIDPPresetService(client.UiV1alpha1(), idpPresetGroup.IDPPresets().Informer())

	return &Resolver{
		idpPresetResolver: newIDPPresetResolver(svc),
		informerFactory:   informerFactory,
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.informerFactory.Start(stopCh)
	r.informerFactory.WaitForCacheSync(stopCh)
}
